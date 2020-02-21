package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Qitmeer/qitmeer-wallet/util"

	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	corejson "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/log"
	chaincfg "github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer/qx"

	"github.com/Qitmeer/qitmeer-wallet/config"
	clijson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet/txrules"
	"github.com/Qitmeer/qitmeer-wallet/walletdb"
	"github.com/Qitmeer/qitmeer-wallet/wtxmgr"
)

const (
	InsecurePubPassphrase = "public"
	webUpdateBlockTicker  = 30
	defaultNewAddressNumber = 1
)

var UploadRun = false

type Wallet struct {
	cfg *config.Config

	// Data stores
	db      walletdb.DB
	Manager *waddrmgr.Manager
	TxStore *wtxmgr.Store

	chainParams *chaincfg.Params

	HttpClient *httpConfig

	// Channels for the manager locker.
	unlockRequests chan unlockRequest
	lockRequests   chan struct{}
	lockState      chan bool

	wg sync.WaitGroup

	started bool
	quit    chan struct{}
	quitMu  sync.Mutex

	SyncHeight int32
}

// Start wallet routine
func (wt *Wallet) Start() {
	log.Trace("wallet start")
	wt.quitMu.Lock()
	select {
	case <-wt.quit:
		wt.quit = make(chan struct{})
	default:
		if wt.started {
			wt.quitMu.Unlock()
			return
		}
		wt.started = true
	}
	wt.quitMu.Unlock()

	go wt.walletLocker()

	go func() {

		updateBlockTicker := time.NewTicker(webUpdateBlockTicker * time.Second)
		for {
			select {
			case <-updateBlockTicker.C:
				if UploadRun == false {
					log.Trace("Updateblock start")
					UploadRun = true
					err := wt.UpdateBlock(0)
					if err != nil {
						log.Error("Start.Updateblock err", "err", err.Error())
					}
					UploadRun = false
				}
			}
		}

	}()
}

type (
	unlockRequest struct {
		passphrase []byte
		lockAfter  <-chan time.Time // nil prevents the timeout.
		err        chan error
	}
)

type Balance struct {
	TotalAmount   types.Amount // 总数
	SpendAmount   types.Amount // 已花费
	UnspendAmount types.Amount //未花费
	ConfirmAmount types.Amount //待确认
}

// AccountBalanceResult is a single result for the Wallet.AccountBalances method.
type AccountBalanceResult struct {
	AccountNumber  uint32
	AccountName    string
	AccountBalance types.Amount
}
type AccountAndAddressResult struct {
	AccountNumber uint32
	AccountName   string
	AddrsOutput   []AddrAndAddrTxOutput
}
type AddrAndAddrTxOutput struct {
	Addr     string
	balance  Balance
	Txoutput []wtxmgr.AddrTxOutput
}

// Namespace bucket keys.
var (
	waddrmgrNamespaceKey = []byte("waddrmgr")
	wtxmgrNamespaceKey   = []byte("wtxmgr")

)

// ImportPrivateKey imports a private key to the wallet and writes the new
// wallet to disk.
//
// NOTE: If a block stamp is not provided, then the wallet's birthday will be
// set to the genesis block of the corresponding chain.
func (wt *Wallet) ImportPrivateKey(scope waddrmgr.KeyScope, wif *utils.WIF) (string, error) {

	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	// Attempt to import private key into wallet.
	var addr types.Address
	err = walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		addrMgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		maddr, err := manager.ImportPrivateKey(addrMgrNs, wif)
		if err != nil {
			return err
		}
		addr = maddr.Address()
		_, err = manager.AccountProperties(
			addrMgrNs, waddrmgr.ImportedAddrAccount,
		)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	addrStr := addr.Encode()
	log.Trace("ImportPrivateKey succ", "address", addrStr)

	// Return the payment address string of the imported private key.
	return addrStr, nil
}

// ChainParams returns the network parameters for the blockchain the wallet
// belongs to.
func (wt *Wallet) ChainParams() *chaincfg.Params {
	return wt.chainParams
}

// Database returns the underlying walletdb database. This method is provided
// in order to allow applications wrapping btcwallet to store app-specific data
// with the wallet's database.
func (wt *Wallet) Database() walletdb.DB {
	return wt.db
}

func Create(db walletdb.DB, pubPass, privPass, seed []byte, params *chaincfg.Params,
	birthday time.Time) error {

	// If a seed was provided, ensure that it is of valid length. Otherwise,
	// we generate a random seed for the wallet with the recommended seed
	// length.
	return walletdb.Update(db, func(tx walletdb.ReadWriteTx) error {
		addrMgrNs, err := tx.CreateTopLevelBucket(waddrmgrNamespaceKey)
		if err != nil {
			return err
		}
		txmgrNs, err := tx.CreateTopLevelBucket(wtxmgrNamespaceKey)
		if err != nil {
			return err
		}
		err = waddrmgr.Create(
			addrMgrNs, seed, pubPass, privPass, params, nil,
			birthday,
		)
		if err != nil {
			return err
		}
		return wtxmgr.Create(txmgrNs)
	})
}

// Open loads an already-created wallet from the passed database and namespaces.
func Open(db walletdb.DB, pubPass []byte, _ *waddrmgr.OpenCallbacks,
	params *chaincfg.Params, _ uint32, cfg *config.Config) (*Wallet, error) {

	var addrMgr *waddrmgr.Manager
	// Before attempting to open the wallet, we'll check if there are any
	// database upgrades for us to proceed. We'll also create our references
	// to the address and transaction managers, as they are backed by the
	// database.
	err := walletdb.Update(db, func(tx walletdb.ReadWriteTx) error {
		addrMgrBucket := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		if addrMgrBucket == nil {
			return errors.New("missing address manager namespace")
		}
		txMgrBucket := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		if txMgrBucket == nil {
			return errors.New("missing transaction manager namespace")
		}
		var err error
		addrMgr, err = waddrmgr.Open(addrMgrBucket, pubPass, params)
		if err != nil {
			return err
		}
		_, err = wtxmgr.Open(txMgrBucket, params)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Trace("Opened wallet")

	w := &Wallet{
		cfg: cfg,
		db:      db,
		Manager: addrMgr,
		unlockRequests: make(chan unlockRequest),
		lockState: make(chan bool),
		chainParams: params,

	}

	return w, nil
}

func (wt *Wallet) GetTx(txId string) (corejson.TxRawResult, error) {

	trx := corejson.TxRawResult{}
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		txNs := ns.NestedReadBucket(wtxmgr.BucketTxJson)
		k, err := hash.NewHashFromStr(txId)
		if err != nil {
			return err
		}
		v := txNs.Get(k.Bytes())
		if v != nil {
			err := json.Unmarshal(v, &trx)
			if err != nil {
				return err
			}
		} else {
			return errors.New("GetTx fail ")
		}
		return nil
	})
	if err != nil {
		return trx, err
	}

	return trx, nil
}

func (wt *Wallet) GetAccountAndAddress(scope waddrmgr.KeyScope) ([]AccountAndAddressResult, error) {
	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, err
	}
	var results []AccountAndAddressResult
	err = walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		lastAcct, err := manager.LastAccount(addrNs)
		if err != nil {
			return err
		}
		results = make([]AccountAndAddressResult, lastAcct+2)
		for i := range results[:len(results)-1] {
			accountName, err := manager.AccountName(addrNs, uint32(i))
			if err != nil {
				return err
			}
			results[i].AccountNumber = uint32(i)
			results[i].AccountName = accountName
		}
		results[len(results)-1].AccountNumber = waddrmgr.ImportedAddrAccount
		results[len(results)-1].AccountName = waddrmgr.ImportedAddrAccountName
		for k := range results {
			adds, err := wt.AccountAddresses(results[k].AccountNumber)
			if err != nil {
				return err
			}
			var addrOutputs []AddrAndAddrTxOutput
			for _, addr := range adds {
				addrOutput, err := wt.getAddrAndAddrTxOutputByAddr(addr.Encode())
				if err != nil {
					return err
				}
				addrOutputs = append(addrOutputs, *addrOutput)
			}
			results[k].AddrsOutput = addrOutputs
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, err
}


func reverse(s []wtxmgr.AddrTxOutput) []wtxmgr.AddrTxOutput {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
func (wt *Wallet) getAddrAndAddrTxOutputByAddr(addr string) (*AddrAndAddrTxOutput, error) {

	ato := AddrAndAddrTxOutput{}
	b := Balance{}
	var txOuts []wtxmgr.AddrTxOutput
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outNs := ns.NestedReadBucket(wtxmgr.BucketAddrtxout)
		hsOutNs := outNs.NestedReadBucket(hs)
		if hsOutNs != nil {
			err := hsOutNs.ForEach(func(k, v []byte) error {
				to := wtxmgr.AddrTxOutput{}
				err := wtxmgr.ReadAddrTxOutput(v, &to)
				if err != nil {
					return err
				}
				txOuts = append(txOuts, to)

				return nil
			})
			if err!=nil{
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	txOuts=reverse(txOuts)

	var spendAmount types.Amount
	var unspentAmount types.Amount
	var totalAmount types.Amount
	var confirmAmount types.Amount
	for _, txOut := range txOuts {
		if txOut.Spend == wtxmgr.SpendStatusSpend {
			spendAmount += txOut.Amount
		} else if txOut.Spend == wtxmgr.SpendStatusUnconfirmed {
				totalAmount += txOut.Amount
				confirmAmount += txOut.Amount
		} else {
				totalAmount += txOut.Amount
				unspentAmount += txOut.Amount
		}
	}

	b.UnspendAmount = unspentAmount
	b.SpendAmount = spendAmount
	b.TotalAmount = totalAmount
	b.ConfirmAmount = confirmAmount
	ato.Addr = addr
	ato.balance = b
	ato.Txoutput = txOuts
	return &ato, nil
}

const (
	defaultPage              = 1
	defaultPageSize          =10
	defaultMaxPageSize       =1000000000
	sTypeIn            int32 =0
	sTypeOut           int32 =1
	sTypeAll           int32 =2
)
/**
sType 0 Turn in 1 Turn out 2 all no page
*/
func (wt *Wallet) GetListTxByAddr(addr string, sType int32, page int32, pageSize int32) (*clijson.PageTxRawResult, error) {
	at, err := wt.getAddrAndAddrTxOutputByAddr(addr)
	result := clijson.PageTxRawResult{}
	if err != nil {
		return nil, err
	}
	if page == 0 {
		page = defaultPage
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	startIndex := (page - 1) * pageSize
	var endIndex int32
	var txHss []hash.Hash
	var txHssIn []hash.Hash
	var dataLen int32
	switch sType {
	case sTypeIn:
		dataLen = int32(len(at.Txoutput))
		if page < 0 {
			for _, txPut := range at.Txoutput {
				txHss = append(txHss, txPut.TxId)
			}
			dataLen = int32(len(txHss))
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("no data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = startIndex + pageSize
				}
				for s := startIndex; s < endIndex; s++ {
					txHss = append(txHss, at.Txoutput[s].TxId)
				}
			}
		}
	case sTypeOut:
		for _, txPut := range at.Txoutput {
			if txPut.Spend == wtxmgr.SpendStatusSpend && txPut.SpendTo != nil {
				txHssIn = append(txHssIn, txPut.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txHssIn))
		if page < 0 {
			txHss = append(txHss, txHssIn...)
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("no data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = startIndex + pageSize
				}
				for s := startIndex; s < endIndex; s++ {
					txHss = append(txHss, txHssIn[s])
				}
			}
		}
	case sTypeAll:
		for _, txPut := range at.Txoutput {
			txHss = append(txHss, txPut.TxId)
			if txPut.Spend ==wtxmgr.SpendStatusSpend && txPut.SpendTo != nil {
				txHss = append(txHss, txPut.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txHss))
		if page < 0 {
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("no data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = startIndex + pageSize
				}
				for s := startIndex; s < endIndex; s++ {
					txHss = append(txHss, txHssIn[s])
				}
			}
		}
	default:
		return nil,fmt.Errorf("err stype")
	}
	result.Page = page
	result.PageSize = pageSize
	result.Total = dataLen
	var transactions []corejson.TxRawResult
	err = walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		txNs := ns.NestedReadBucket(wtxmgr.BucketTxJson)
		for _, txHs := range txHss {
			v := txNs.Get(txHs.Bytes())
			if v == nil {
				return fmt.Errorf("db uploadblock err tx:%s non-existent", txHs.String())
			}
			var txr corejson.TxRawResult
			err := json.Unmarshal(v, &txr)
			if err != nil {
				return err
			}
			transactions = append(transactions, txr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	result.Transactions = transactions
	return &result, nil
}

func (wt *Wallet) GetBalance(addr string) (*Balance, error) {
	if addr == "" {
		return nil, errors.New("addr is nil")
	}
	res, err := wt.getAddrAndAddrTxOutputByAddr(addr)
	if err != nil {
		return nil, err
	}
	return &res.balance, nil
}
func (wt *Wallet) GetTxSpendInfo(txId string)  ([]*wtxmgr.AddrTxOutput,error){
	var atos []*wtxmgr.AddrTxOutput
	txHash,err:=hash.NewHashFromStr(txId)
	if err!=nil{
		return nil,err
	}
	err=walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		rb := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		txNrb := rb.NestedReadWriteBucket(wtxmgr.BucketTxJson)
		outNrb := rb.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		v := txNrb.Get(txHash.Bytes())
		if v == nil{
			return fmt.Errorf("txid does not exist")
		}
		var txr corejson.TxRawResult
		err := json.Unmarshal(v, &txr)
		if err != nil {
			return err
		}
		for i, vOut := range txr.Vout {
			addr := vOut.ScriptPubKey.Addresses[0]
			top := types.TxOutPoint{
				Hash:     *txHash,
				OutIndex: uint32(i),
			}
			var ato, err = wt.TxStore.GetAddrTxOut(outNrb, addr, top)
			if err != nil {
				return err
			}
			ato.Address=addr
			atos = append(atos, ato)
		}
		return err
	})
	if err!=nil{
		return nil,err
	}
	return atos,nil
}

func (wt *Wallet) insertTx(txins []wtxmgr.TxInputPoint, txouts []wtxmgr.AddrTxOutput, trrs []corejson.TxRawResult) error {
	err := walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		txNs := ns.NestedReadWriteBucket(wtxmgr.BucketTxJson)
		outNs := ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		for _, tr := range trrs {
			k, err := hash.NewHashFromStr(tr.Txid)
			if err != nil {
				return err
			}
			v, err := json.Marshal(tr)
			if err != nil {
				return err
			}
			ks := k.Bytes()
			err = txNs.Put(ks, v)
			if err != nil {
				return err
			}
		}
		for _, txo := range txouts {
			err := wt.TxStore.UpdateAddrTxOut(outNs, &txo)
			if err != nil {
				return err
			}
		}
		for _, txi := range txins {
			v := txNs.Get(txi.TxOutPoint.Hash.Bytes())
			if v == nil {
				continue
			}
			var txr corejson.TxRawResult
			err := json.Unmarshal(v, &txr)
			if err != nil {
				return err
			}
			addr := txr.Vout[txi.TxOutPoint.OutIndex].ScriptPubKey.Addresses[0]
			spendOut, err := wt.TxStore.GetAddrTxOut(outNs, addr, txi.TxOutPoint)
			if err != nil {
				return err
			}

			spendOut.Spend = wtxmgr.SpendStatusSpend
			spendOut.Address = addr
			spendOut.SpendTo = &txi.SpendTo

			err = wt.TxStore.UpdateAddrTxOut(outNs, spendOut)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (wt *Wallet) SyncTx(order int64) (clijson.BlockHttpResult, error) {
	var block clijson.BlockHttpResult
	blockByte, err := wt.HttpClient.getBlockByOrder(order)
	if err != nil {
		return block, err
	}
	if err := json.Unmarshal(blockByte, &block); err == nil {
		if !block.Txsvalid {
			log.Trace(fmt.Sprintf("block:%v err,txsvalid is false", block.Hash))
			return block, nil
		}
		isBlue ,err:=wt.HttpClient.isBlue(block.Hash)
		if err != nil {
			return block, err
		}else{
			if isBlue != "1"{
				log.Trace(fmt.Sprintf("block:%v err,is not blue", block.Hash))
				return block, nil
			}
		}
		txIns, txOuts, trRs, err := parseBlockTxs(block)
		if err != nil {
			return block, err
		}
		err = wt.insertTx(txIns, txOuts, trRs)
		if err != nil {
			return block, err
		}

	} else {
		log.Error(err.Error())
		return block, err
	}
	return block, nil
}

func parseTx(tr corejson.TxRawResult, height int32) ([]wtxmgr.TxInputPoint, []wtxmgr.AddrTxOutput, error) {
	var txins []wtxmgr.TxInputPoint
	var txouts []wtxmgr.AddrTxOutput
	blockhash, err := hash.NewHashFromStr(tr.BlockHash)
	if err != nil {
		return nil, nil, err
	}
	block := wtxmgr.Block{
		Hash:   *blockhash,
		Height: height,
	}
	txId, err := hash.NewHashFromStr(tr.Txid)
	if err != nil {
		return nil, nil, err
	}
	spend := wtxmgr.SpendStatusUnspent
	if tr.Confirmations < config.Cfg.Confirmations{
		spend=wtxmgr.SpendStatusUnconfirmed
	}
	for i, vi := range tr.Vin {
		if vi.Coinbase != "" {
			continue
		}
		if vi.Txid == "" && vi.Vout == 0 {
			continue
		} else {
			hs, err := hash.NewHashFromStr(vi.Txid)
			if err != nil {
				return nil, nil, err
			} else {
				txOutPoint := types.TxOutPoint{
					Hash:     *hs,
					OutIndex: vi.Vout,
				}
				spendTo :=wtxmgr.SpendTo{
					Index:  uint32(i),
					TxHash: *txId,
				}
				txIn :=wtxmgr.TxInputPoint{
					TxOutPoint: txOutPoint,
					SpendTo:    spendTo,
				}
				txins = append(txins, txIn)
			}
		}
	}
	for index, vo := range tr.Vout {
		if len(vo.ScriptPubKey.Addresses) == 0 {
			continue
		} else {
			txOut := wtxmgr.AddrTxOutput{
				Address: vo.ScriptPubKey.Addresses[0],
				TxId:    *txId,
				Index:   uint32(index),
				Amount:  types.Amount(vo.Amount),
				Block:   block,
				Spend:   spend,
			}
			txouts = append(txouts, txOut)
		}
	}

	return txins, txouts, nil
}

func parseBlockTxs(block clijson.BlockHttpResult) ([]wtxmgr.TxInputPoint, []wtxmgr.AddrTxOutput, []corejson.TxRawResult, error) {
	var txIns []wtxmgr.TxInputPoint
	var txOuts []wtxmgr.AddrTxOutput
	var tx []corejson.TxRawResult
	for _, tr := range block.Transactions {
		tx = append(tx, tr)
		tin, tout, err := parseTx(tr, block.Order)
		if err != nil {
			return nil, nil, nil, err
		} else {
			txIns = append(txIns, tin...)
			txOuts = append(txOuts, tout...)
		}
	}
	return txIns, txOuts, tx, nil
}

func (wt *Wallet) GetSyncBlockHeight() int32 {
	height := wt.Manager.SyncedTo().Height
	return height
}

func (wt *Wallet) SetSynceToNum(order int64) error {
	var block clijson.BlockHttpResult
	blockByte, err := wt.HttpClient.getBlockByOrder(order)
	if err != nil {
		return  err
	}
	if err := json.Unmarshal(blockByte, &block); err == nil {
		if !block.Txsvalid {
			log.Trace(fmt.Sprintf("block:%v err,txsvalid is false", block.Hash))
			return nil
		}
		hs, err := hash.NewHashFromStr(block.Hash)
		if err != nil {
			return fmt.Errorf("blockhash string to hash  err:%s", err.Error())
		}
		stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: block.Order}
		err = walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
			ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
			err := wt.Manager.SetSyncedTo(ns, stamp)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	} else {
		log.Error(err.Error())
		return  err
	}
}


func (wt *Wallet) handleBlockSynced(order int64) error {

	br, er := wt.SyncTx(order)
	if er != nil {
		return er
	}
	hs, err := hash.NewHashFromStr(br.Hash)
	if err != nil {
		return fmt.Errorf("blockhash string to hash  err:%s", err.Error())
	}
	if br.Confirmations > config.Cfg.Confirmations {
		stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: br.Order}
		err = walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
			ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
			err := wt.Manager.SetSyncedTo(ns, stamp)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (wt *Wallet) UpdateBlock(toHeight int64) error {
	var blockCount string
	var err error
	if toHeight == 0 {
		blockCount, err = wt.HttpClient.getblockCount()
		if err != nil {
			return err
		}
	} else {
		blockCount = strconv.FormatInt(toHeight, strIntBase)
	}
	blockHeight, err := strconv.ParseInt(blockCount, strIntBase, strIntBitSize32)
	if err != nil {
		return err
	}
	h := int64(wt.Manager.SyncedTo().Height)
	if h < blockHeight {
		log.Trace(fmt.Sprintf("localheight:%d,blockHeight:%d", h, blockHeight))
		for h < blockHeight {
			err := wt.handleBlockSynced(h)
			if err != nil {
				return err
			} else {
				wt.SyncHeight = int32(h)
				_, _ = fmt.Fprintf(os.Stdout, "update blcok:%s/%s\r", strconv.FormatInt(h, 10), strconv.FormatInt(blockHeight-1, 10))
				h++
			}
		}
		fmt.Print("\nsucc\n")
	} else {
		fmt.Println("Block data is up to date")
	}
	return nil
}


// NextAccount creates the next account and returns its account number.  The
// name must be unique to the account.  In order to support automatic seed
// restoring, new accounts may not be created when all of the previous 100
// accounts have no transaction history (this is a deviation from the BIP0044
// spec, which allows no unused account gaps).
func (wt *Wallet) NextAccount(scope waddrmgr.KeyScope, name string) (uint32, error) {
	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var (
		account uint32
	)
	err = walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		addrMgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.NewAccount(addrMgrNs, name)
		if err != nil {
			return err
		}
		_, err = manager.AccountProperties(addrMgrNs, account)
		return err
	})
	if err != nil {
		log.Error("Cannot fetch new account properties for notification "+
			"after account creation", "err", err)
		return account, err
	}
	return account, err
}

// AccountBalances returns all accounts in the wallet and their balances.
// Balances are determined by excluding transactions that have not met
// requiredConfs confirmations.
func (wt *Wallet) AccountBalances(scope waddrmgr.KeyScope) ([]AccountBalanceResult, error) {
	aaaRs, err := wt.GetAccountAndAddress(scope)
	if err != nil {
		return nil, err
	}
	results := make([]AccountBalanceResult, len(aaaRs))
	for index, aaa := range aaaRs {
		results[index].AccountNumber = aaa.AccountNumber
		results[index].AccountName = aaa.AccountName
		unSpendAmount := types.Amount(0)
		for _, addr := range aaa.AddrsOutput {
			unSpendAmount = unSpendAmount + addr.balance.UnspendAmount
		}
		results[index].AccountBalance = unSpendAmount
	}
	return results, nil
}

// AccountNumber returns the account number for an account name under a
// particular key scope.
func (wt *Wallet) AccountNumber(scope waddrmgr.KeyScope, accountName string) (uint32, error) {
	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var account uint32
	err = walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.LookupAccount(addrMgrNs, accountName)
		return err
	})
	return account, err
}

// NewAddress returns the next external chained address for a wallet.
func (wt *Wallet) NewAddress(
	scope waddrmgr.KeyScope, account uint32) (types.Address, error) {
	var (
		addr types.Address
	)
	err := walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		addrMgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		addr, _, err = wt.newAddress(addrMgrNs, account, scope)
		return err
	})
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func (wt *Wallet) newAddress(addrMgrNs walletdb.ReadWriteBucket, account uint32,
	scope waddrmgr.KeyScope) (types.Address, *waddrmgr.AccountProperties, error) {

	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, nil, err
	}

	// Get next address from wallet.
	addr, err := manager.NextExternalAddresses(addrMgrNs, account, defaultNewAddressNumber)
	if err != nil {
		return nil, nil, err
	}

	props, err := manager.AccountProperties(addrMgrNs, account)
	if err != nil {
		log.Error(fmt.Sprintf("Cannot fetch account properties for notification "+
			"after deriving next external address: %v", err))
		return nil, nil, err
	}

	return addr[0].Address(), props, nil
}

// DumpWIFPrivateKey returns the WIF encoded private key for a
// single wallet address.
func (wt *Wallet) DumpWIFPrivateKey(addr types.Address) (string, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		waddrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = wt.Manager.Address(waddrMgrNs, addr)
		return err
	})
	if err != nil {
		return "", err
	}
	pka, ok := maddr.(waddrmgr.ManagedPubKeyAddress)
	if !ok {
		return "", fmt.Errorf("address %s is not a key type", addr)
	}
	wif, err := pka.ExportPrivKey()
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}
func (wt *Wallet) getPrivateKey(addr types.Address) (waddrmgr.ManagedPubKeyAddress, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		waddrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = wt.Manager.Address(waddrMgrNs, addr)
		return err
	})
	if err != nil {
		return nil, err
	}
	pka, ok := maddr.(waddrmgr.ManagedPubKeyAddress)
	if !ok {
		return nil, fmt.Errorf("address %s is not a key type", addr)
	}
	return pka, nil
}

// Unlock unlocks the wallet's address manager and relocks it after timeout has
// expired.  If the wallet is already unlocked and the new passphrase is
// correct, the current timeout is replaced with the new one.  The wallet will
// be locked if the passphrase is incorrect or any other error occurs during the
// unlock.
func (wt *Wallet) Unlock(passphrase []byte, lock <-chan time.Time) error {
	log.Trace("wallet Unlock")
	err := make(chan error, 1)
	wt.unlockRequests <- unlockRequest{
		passphrase: passphrase,
		lockAfter:  lock,
		err:        err,
	}
	log.Trace("wallet Unlock end")
	return <-err
}

//// Lock locks the wallet's address manager.
func (wt *Wallet) Lock() {
	wt.lockRequests <- struct{}{}
}

//// Locked returns whether the account manager for a wallet is locked.
func (wt *Wallet) Locked() bool {
	return <-wt.lockState
}

// quitChan atomically reads the quit channel.
func (wt *Wallet) quitChan() <-chan struct{} {
	wt.quitMu.Lock()
	c := wt.quit
	wt.quitMu.Unlock()
	return c
}

func (wt *Wallet) UnLockManager(passphrase []byte) error {
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return wt.Manager.Unlock(addrMgrNs, passphrase)
	})
	if err != nil {
		return err
	}
	return nil
}

// walletLocker manages the locked/unlocked state of a wallet.
func (wt *Wallet) walletLocker() {
	log.Trace("wallet walletLocker")
	var timeout <-chan time.Time
	quit := wt.quitChan()
out:
	for {
		select {
		case req := <-wt.unlockRequests:
			log.Trace("walletLocker,unlockRequests")
			err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
				addMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
				return wt.Manager.Unlock(addMgrNs, req.passphrase)
			})
			if err != nil {
				req.err <- err
				continue
			}
			timeout = req.lockAfter
			if timeout == nil {
				log.Info("The wallet has been unlocked without a time limit")
			} else {
				log.Info("The wallet has been temporarily unlocked")
			}
			req.err <- nil
			continue



		case wt.lockState <- wt.Manager.IsLocked():
			continue

		case <-quit:
			break out

		}

	}
	wt.wg.Done()
}

// AccountAddresses returns the addresses for every created address for an
// account.
func (wt *Wallet) AccountAddresses(account uint32) (addrs []types.Address, err error) {
	err = walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return wt.Manager.ForEachAccountAddress(addrMgrNs, account, func(mAddr waddrmgr.ManagedAddress) error {
			addrs = append(addrs, mAddr.Address())
			return nil
		})
	})
	return
}

// AccountOfAddress finds the account that an address is associated with.
func (wt *Wallet) AccountOfAddress(a types.Address) (uint32, error) {
	var account uint32
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		_, account, err = wt.Manager.AddrAccount(addrMgrNs, a)
		return err
	})
	return account, err
}

// AccountName returns the name of an account.
func (wt *Wallet) AccountName(scope waddrmgr.KeyScope, accountNumber uint32) (string, error) {
	manager, err := wt.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	var accountName string
	err = walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		accountName, err = manager.AccountName(addrMgrNs, accountNumber)
		return err
	})
	return accountName, err
}

func (wt *Wallet) GetUtxo(addr string) ([]wtxmgr.UTxo, error) {
	var txouts []wtxmgr.AddrTxOutput
	var utxos []wtxmgr.UTxo
	err := walletdb.View(wt.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadBucket(wtxmgr.BucketAddrtxout)
		hsoutns := outns.NestedReadBucket(hs)
		if hsoutns != nil {
			_ = hsoutns.ForEach(func(k, v []byte) error {
				to := wtxmgr.AddrTxOutput{}
				err := wtxmgr.ReadAddrTxOutput(v, &to)
				if err != nil {
					log.Error("readAddrTxOutput err", "err", err.Error())
					return err
				}
				txouts = append(txouts, to)

				return nil
			})
		}
		return nil
	})
	if err != nil {
		log.Error("ReadAddrTxOutput err", "err", err)
		return nil, err
	}

	for _, txout := range txouts {
		uo := wtxmgr.UTxo{}
		if txout.Spend == wtxmgr.SpendStatusUnspent {
			uo.TxId = txout.TxId.String()
			uo.Index = txout.Index
			uo.Amount = txout.Amount
			utxos = append(utxos, uo)
		}
	}
	return utxos, nil
}

// Sendoutputs can only be accessed by a single thread at the same time to prevent the referenced utxo from being referenced again under the concurrency
var syncSendOutputs = new(sync.Mutex)
// SendOutputs creates and sends payment transactions. It returns the
// transaction upon success.
func (wt *Wallet) SendOutputs(outputs []*types.TxOutput, account int64,  satPerKb types.Amount) (*string, error) {


	// Ensure the outputs to be created adhere to the network's consensus
	// rules.
	syncSendOutputs.Lock()
	defer syncSendOutputs.Unlock()
	tx := types.NewTransaction()
	payAmout := types.Amount(0)
	feeAmout := int64(0)
	for _, output := range outputs {
		if err := txrules.CheckOutput(output, satPerKb); err != nil {
			return nil, err
		}
		payAmout = payAmout + types.Amount(output.Amount)
		tx.AddTxOut(output)
	}
	aaars, err := wt.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}

	var sendAddrTxOutput []wtxmgr.AddrTxOutput
	var prk string
b:
	for _, aaar := range aaars {

		if int64(aaar.AccountNumber) != account && account != waddrmgr.AccountMergePayNum{
			continue
		}

		for _, addroutput := range aaar.AddrsOutput {
			log.Trace(fmt.Sprintf("addr:%s,unspend:%v",addroutput.Addr,addroutput.balance.UnspendAmount))
			if addroutput.balance.UnspendAmount > payAmout {
				addr, err := address.DecodeAddress(addroutput.Addr)
				if err != nil {
					return nil, err
				}
				frompkscipt, err := txscript.PayToAddrScript(addr)
				if err != nil {
					return nil, err
				}
				pri, err := wt.getPrivateKey(addr)
				if err != nil {
					return nil, err
				}
				priKey, err := pri.PrivKey()
				if err != nil {
					return nil, err
				}
				prk = hex.EncodeToString(priKey.SerializeSecret())
				for _, output:= range addroutput.Txoutput {
					output.Address = addroutput.Addr
					if output.Spend == wtxmgr.SpendStatusUnspent {
						if payAmout > 0 && feeAmout==0{
							if output.Amount > payAmout {
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, nil))
								selfTxOut := types.NewTxOutput(uint64(output.Amount-payAmout), frompkscipt)
								feeAmout = util.CalcMinRequiredTxRelayFee(int64(tx.SerializeSize()+selfTxOut.SerializeSize()), types.Amount(config.Cfg.MinTxFee))
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								if (output.Amount-payAmout-types.Amount(feeAmout)) >= 0{
									selfTxOut.Amount = uint64(output.Amount-payAmout-types.Amount(feeAmout))
									if selfTxOut.Amount >0 {
										tx.AddTxOut(selfTxOut)
									}
									payAmout = 0
									feeAmout = 0
									break b
								}else{
									selfTxOut.Amount = uint64(output.Amount-payAmout)
									payAmout = 0
									tx.AddTxOut(selfTxOut)
								}

							}else{
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, nil))
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								payAmout = payAmout- output.Amount
							}
						}else if payAmout == 0 && feeAmout >0{
							if output.Amount >= types.Amount(feeAmout){
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, nil))
								selfTxOut := types.NewTxOutput(uint64(output.Amount-types.Amount(feeAmout)), frompkscipt)
								if selfTxOut.Amount >0{
									tx.AddTxOut(selfTxOut)
								}
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								feeAmout = 0
								break b
							}else{
								log.Trace("utxo < feeAmout")
							}

						}else{
							log.Trace(fmt.Sprintf("system err payAmout :%v ,feeAmout :%v\n",payAmout,feeAmout))
							return nil,fmt.Errorf("system err payAmout :%v ,feeAmout :%v\n",payAmout,feeAmout)
						}
					}
				}
			}
		}
	}
	if payAmout.ToCoin() != types.Amount(0).ToCoin() || feeAmout!=0{
		log.Trace("payAmout", "payAmout", payAmout)
		log.Trace("feeAmout", "feeAmout", feeAmout)
		return nil, fmt.Errorf("balance is not enough")
	}

	b, err := tx.Serialize()
	if err != nil {
		return nil, err
	}
	signTx, err := qx.TxSign(prk, hex.EncodeToString(b), wt.chainParams.Name)
	if err != nil {
		return nil, err
	}
	log.Trace(fmt.Sprintf("signTx size:%v", len(signTx)), "signTx", signTx)
	msg, err := wt.HttpClient.SendRawTransaction(signTx, false)
	if err != nil{
		log.Trace("SendRawTransaction txSign err ", "err", err.Error())
		return nil, err
	} else {
		msg=strings.ReplaceAll(msg,"\"","")
		log.Trace("SendRawTransaction txSign response msg", "msg", msg)
	}

	err = walletdb.Update(wt.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		for _, txoutput := range sendAddrTxOutput {
			txoutput.Spend = wtxmgr.SpendStatusSpend
			err = wt.TxStore.UpdateAddrTxOut(outns, &txoutput)
			if err != nil {
				log.Error("UpdateAddrTxOut to spend err", "err", err.Error())
				return err
			}
		}
		log.Trace("UpdateAddrTxOut to spend succ ")
		return nil
	})
	if err != nil {
		log.Error("UpdateAddrTxOut to spend err", "err", err.Error())
		return nil, err
	}

	return &msg, nil
}


//sendPairs creates and sends payment transactions.
//It returns the transaction hash in string format upon success
//All errors are returned in btcjson.RPCError format
func (wt *Wallet)  SendPairs( amounts map[string]types.Amount,
	account int64,  feeSatPerKb types.Amount) (string, error) {
	check,err := wt.HttpClient.CheckSyncUpdate(int64(wt.Manager.SyncedTo().Height))

	if check ==false{
		return "",err
	}
	outputs, err := makeOutputs(amounts)
	if err != nil {
		return "", err
	}
	tx, err := wt.SendOutputs(outputs, account, feeSatPerKb)
	if err != nil {
		if err == txrules.ErrAmountNegative {
			return "", qitmeerjson.ErrNeedPositiveAmount
		}
		if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
			return "", &qitmeerjson.ErrWalletUnlockNeeded
		}
		switch err.(type) {
		case qitmeerjson.RPCError:
			return "", err
		}

		return "", &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInternal.Code,
			Message: err.Error(),
		}
	}
	return *tx, nil
}

// makeOutputs creates a slice of transaction outputs from a pair of address
// strings to amounts.  This is used to create the outputs to include in newly
// created transactions from a JSON object describing the output destinations
// and amounts.
func makeOutputs(pairs map[string]types.Amount) ([]*types.TxOutput, error) {
	outputs := make([]*types.TxOutput, 0, len(pairs))
	for addrStr, amt := range pairs {
		addr, err := address.DecodeAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("cannot decode address: %s", err)
		}

		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("cannot create txout script: %s", err)
		}

		outputs = append(outputs, types.NewTxOutput(uint64(amt), pkScript))
	}
	return outputs, nil
}
