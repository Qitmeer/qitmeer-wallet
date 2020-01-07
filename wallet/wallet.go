package wallet

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/globalvariable"
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
	// InsecurePubPassphrase is the default outer encryption passphrase used
	// for public data (everything but private keys).  Using a non-default
	// public passphrase can prevent an attacker without the public
	// passphrase from discovering all past and future wallet addresses if
	// they gain access to the wallet database.
	//
	// NOTE: at time of writing, public encryption only applies to public
	// data in the waddrmgr namespace.  Transactions are not yet encrypted.
	InsecurePubPassphrase = "public"
)

var UploadRun = false

// Wallet qitmeer-wallet
type Wallet struct {
	cfg *config.Config

	// Data stores
	db      walletdb.DB
	Manager *waddrmgr.Manager
	TxStore *wtxmgr.Store

	chainParams *chaincfg.Params

	Httpclient *htpc

	// Channels for the manager locker.
	unlockRequests chan unlockRequest
	lockRequests   chan struct{}
	lockState      chan bool

	wg sync.WaitGroup

	started bool
	quit    chan struct{}
	quitMu  sync.Mutex

	//
	SyncHeight int32
}

// Start wallet routine
func (wt *Wallet) Start() {
	log.Trace("wallet start")
	wt.quitMu.Lock()
	select {
	case <-wt.quit:
		// Restart the wallet goroutines after shutdown finishes.
		//wt.WaitForShutdown()
		wt.quit = make(chan struct{})
	default:
		// Ignore when the wallet is still running.
		if wt.started {
			wt.quitMu.Unlock()
			return
		}
		wt.started = true
	}
	wt.quitMu.Unlock()

	go wt.walletLocker()

	go func() {

		updateBlockTicker := time.NewTicker(globalvariable.WebupdateBlockTicker * time.Second)
		for {
			select {
			case <-updateBlockTicker.C:
				if UploadRun == false {
					log.Trace("Updateblock start")
					UploadRun = true
					err := wt.Updateblock(0)
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

	changePassphraseRequest struct {
		old, new []byte
		private  bool
		err      chan error
	}

	changePassphrasesRequest struct {
		publicOld, publicNew   []byte
		privateOld, privateNew []byte
		err                    chan error
	}

	// heldUnlock is a tool to prevent the wallet from automatically
	// locking after some timeout before an operation which needed
	// the unlocked wallet has finished.  Any aquired heldUnlock
	// *must* be released (preferably with a defer) or the wallet
	// will forever remain unlocked.
	heldUnlock chan struct{}
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
func (w *Wallet) ImportPrivateKey(scope waddrmgr.KeyScope, wif *utils.WIF,
	bs *waddrmgr.BlockStamp, rescan bool) (string, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	// Attempt to import private key into wallet.
	var addr types.Address
	//var props *waddrmgr.AccountProperties
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		maddr, err := manager.ImportPrivateKey(addrmgrNs, wif, bs)
		if err != nil {
			return err
		}
		addr = maddr.Address()
		_, err = manager.AccountProperties(
			addrmgrNs, waddrmgr.ImportedAddrAccount,
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
func (w *Wallet) ChainParams() *chaincfg.Params {
	return w.chainParams
}

// Database returns the underlying walletdb database. This method is provided
// in order to allow applications wrapping btcwallet to store app-specific data
// with the wallet's database.
func (w *Wallet) Database() walletdb.DB {
	return w.db
}

func Create(db walletdb.DB, pubPass, privPass, seed []byte, params *chaincfg.Params,
	birthday time.Time) error {

	// If a seed was provided, ensure that it is of valid length. Otherwise,
	// we generate a random seed for the wallet with the recommended seed
	// length.
	return walletdb.Update(db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs, err := tx.CreateTopLevelBucket(waddrmgrNamespaceKey)
		if err != nil {
			return err
		}
		txmgrNs, err := tx.CreateTopLevelBucket(wtxmgrNamespaceKey)
		if err != nil {
			return err
		}
		err = waddrmgr.Create(
			addrmgrNs, seed, pubPass, privPass, params, nil,
			birthday,
		)
		if err != nil {
			return err
		}
		//return nil
		return wtxmgr.Create(txmgrNs)
	})
}

// Open loads an already-created wallet from the passed database and namespaces.
func Open(db walletdb.DB, pubPass []byte, cbs *waddrmgr.OpenCallbacks,
	params *chaincfg.Params, recoveryWindow uint32, cfg *config.Config) (*Wallet, error) {

	var addrMgr *waddrmgr.Manager
	//var	txMgr   *wtxmgr.Store

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
		//addrMgrUpgrader := waddrmgr.NewMigrationManager(addrMgrBucket)
		//txMgrUpgrader := wtxmgr.NewMigrationManager(txMgrBucket)
		//err := migration.Upgrade(txMgrUpgrader, addrMgrUpgrader)
		//if err != nil {
		//	return err
		//}
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

	log.Trace("Opened wallet") // TODO: log balance? last sync height?

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

func (w *Wallet) GetTx(txid string) (corejson.TxRawResult, error) {

	trx := corejson.TxRawResult{}
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		txns := ns.NestedReadBucket(wtxmgr.BucketTxJson)
		k, err := hash.NewHashFromStr(txid)
		if err != nil {
			return err
		}
		v := txns.Get(k.Bytes())
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

func (w *Wallet) GetAccountAndAddress(scope waddrmgr.KeyScope,
	requiredConfs int32) ([]AccountAndAddressResult, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, err
	}
	var results []AccountAndAddressResult
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		lastAcct, err := manager.LastAccount(addrmgrNs)
		if err != nil {
			return err
		}
		results = make([]AccountAndAddressResult, lastAcct+2)
		for i := range results[:len(results)-1] {
			accountName, err := manager.AccountName(addrmgrNs, uint32(i))
			if err != nil {
				return err
			}
			results[i].AccountNumber = uint32(i)
			results[i].AccountName = accountName
		}
		results[len(results)-1].AccountNumber = waddrmgr.ImportedAddrAccount
		results[len(results)-1].AccountName = waddrmgr.ImportedAddrAccountName
		for k, _ := range results {
			addrs, err := w.AccountAddresses(results[k].AccountNumber)
			if err != nil {
				return err
			}
			addroutputs := []AddrAndAddrTxOutput{}
			for _, addr := range addrs {
				addroutput, err := w.getAddrAndAddrTxOutputByAddr(addr.Encode(), requiredConfs)
				if err != nil {
					return err
				}
				addroutputs = append(addroutputs, *addroutput)
			}
			results[k].AddrsOutput = addroutputs
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, err
}

func (w *Wallet) getAddrAndAddrTxOutputByAddr(addr string, requiredConfs int32) (*AddrAndAddrTxOutput, error) {

	ato := AddrAndAddrTxOutput{}
	b := Balance{}
	var txouts []wtxmgr.AddrTxOutput
	//var txins []*types.TxOutPoint
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadBucket(wtxmgr.BucketAddrtxout)
		hsoutns := outns.NestedReadBucket(hs)
		if hsoutns != nil {
			hsoutns.ForEach(func(k, v []byte) error {
				to := wtxmgr.AddrTxOutput{}
				err := wtxmgr.ReadAddrTxOutput(v, &to)
				if err != nil {
					return err
				}
				txouts = append(txouts, to)

				return nil
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var spendAmount types.Amount
	var unspendAmount types.Amount
	var totalAmount types.Amount
	var confirmAmount types.Amount
	for _, txout := range txouts {
		if txout.Spend == wtxmgr.SpendF {
			spendAmount += txout.Amount
			//totalAmount += txout.Amount
		} else if txout.Spend == wtxmgr.SpendT {
				totalAmount += txout.Amount
				confirmAmount += txout.Amount
		} else {
				totalAmount += txout.Amount
				unspendAmount += txout.Amount
		}
	}

	b.UnspendAmount = unspendAmount
	b.SpendAmount = spendAmount
	b.TotalAmount = totalAmount
	b.ConfirmAmount = confirmAmount
	ato.Addr = addr
	ato.balance = b
	ato.Txoutput = txouts
	return &ato, nil
}

const (
	defaultPage = 1
	defaultPagesize=10
	defaultMaxPageSize=1000000000
	stypeZ int32=0
	stypeF int32=1
	stypeT int32=2
)
/**
stype 0 Turn in 1 Turn out 2 all no page
*/
func (w *Wallet) GetListTxByAddr(addr string, stype int32, page int32, pageSize int32) (*clijson.PageTxRawResult, error) {
	at, err := w.getAddrAndAddrTxOutputByAddr(addr, 1)
	result := clijson.PageTxRawResult{}
	if err != nil {
		return nil, err
	}
	if page == 0 {
		page = defaultPage
	}
	if pageSize == 0 {
		pageSize = defaultPagesize
	}
	startIndex := (page - 1) * pageSize
	var endIndex int32
	//var endIndex :=startIndex+pageSize
	var txhss []hash.Hash
	var txhssin []hash.Hash
	var dataLen int32
	switch stype {
	case stypeZ:
		dataLen = int32(len(at.Txoutput))
		if page < 0 {
			for _, txput := range at.Txoutput {
				txhss = append(txhss, txput.Txid)
			}
			dataLen = int32(len(txhss))
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, at.Txoutput[s].Txid)
				}
			}
		}
	case stypeF:
		for _, txput := range at.Txoutput {
			if txput.Spend == wtxmgr.SpendF && txput.SpendTo != nil {
				txhssin = append(txhssin, txput.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txhssin))
		if page < 0 {
			txhss = append(txhss, txhssin...)
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, txhssin[s])
				}
			}
		}
	case stypeT:
		for _, txput := range at.Txoutput {
			txhss = append(txhss, txput.Txid)
			if txput.Spend == wtxmgr.SpendF && txput.SpendTo != nil {
				txhss = append(txhss, txput.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txhss))
		if page < 0 {
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, txhssin[s])
				}
			}
		}
	default:
		return nil,fmt.Errorf("err stype")
	}
	if stype == stypeZ {
		dataLen = int32(len(at.Txoutput))
		if page < 0 {
			for _, txput := range at.Txoutput {
				txhss = append(txhss, txput.Txid)
			}
			dataLen = int32(len(txhss))
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, at.Txoutput[s].Txid)
				}
			}
		}
	} else if stype == stypeF{
		for _, txput := range at.Txoutput {
			if txput.Spend == wtxmgr.SpendF && txput.SpendTo != nil {
				txhssin = append(txhssin, txput.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txhssin))
		if page < 0 {
			txhss = append(txhss, txhssin...)
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, txhssin[s])
				}
			}
		}
	} else {
		for _, txput := range at.Txoutput {
			txhss = append(txhss, txput.Txid)
			if txput.Spend == wtxmgr.SpendF && txput.SpendTo != nil {
				txhss = append(txhss, txput.SpendTo.TxHash)
			}
		}
		dataLen = int32(len(txhss))
		if page < 0 {
			page = defaultPage
			pageSize = defaultMaxPageSize
		} else {
			if startIndex > dataLen {
				return nil, fmt.Errorf("No data")
			} else {
				if (startIndex + pageSize) > dataLen {
					endIndex = dataLen
				} else {
					endIndex = (startIndex + pageSize)
				}
				for s := startIndex; s < endIndex; s++ {
					txhss = append(txhss, txhssin[s])
				}
			}
		}
	}
	result.Page = page
	result.PageSize = pageSize
	result.Total = dataLen
	var transactions []corejson.TxRawResult
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		txns := ns.NestedReadBucket(wtxmgr.BucketTxJson)
		for _, txhs := range txhss {
			v := txns.Get(txhs.Bytes())
			if v == nil {
				return fmt.Errorf("db uploadblock err tx:%s non-existent", txhs.String())
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

func (w *Wallet) GetBalance(addr string, requiredConfs int32) (*Balance, error) {
	if addr == "" {
		return nil, errors.New("addr is nil")
	}
	res, err := w.getAddrAndAddrTxOutputByAddr(addr, requiredConfs)
	if err != nil {
		return nil, err
	}
	return &res.balance, nil
}

func (w *Wallet) insertTx(txins []types.TxOutPoint, txouts []wtxmgr.AddrTxOutput, trrs []corejson.TxRawResult) error {
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		txns := ns.NestedReadWriteBucket(wtxmgr.BucketTxJson)
		//inns:=ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxin)
		outns := ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		for _, tr := range trrs {
			//log.Info("save txid :",tr.Txid)
			k, err := hash.NewHashFromStr(tr.Txid)
			if err != nil {
				return err
			}
			v, err := json.Marshal(tr)
			if err != nil {
				return err
			}
			ks := k.Bytes()
			err = txns.Put(ks, v)
			if err != nil {
				return err
			}
		}
		for _, txo := range txouts {
			err := w.TxStore.UpdateAddrTxOut(outns, &txo)
			if err != nil {
				return err
			}
		}
		for _, txi := range txins {
			v := txns.Get(txi.Hash.Bytes())
			if v == nil {
				continue
			}
			var txr corejson.TxRawResult
			err := json.Unmarshal(v, &txr)
			if err != nil {
				return err
			}
			addr := txr.Vout[txi.OutIndex].ScriptPubKey.Addresses[0]
			spendedOut, err := w.TxStore.GetAddrTxOut(outns, addr, txi)
			if err != nil {
				return err
			}
			if spendedOut.Spend != wtxmgr.SpendF {
				txHash, err := hash.NewHashFromStr(txr.Txid)
				if err != nil {
					return err
				}
				spendto := wtxmgr.SpendTo{
					Index:  txi.OutIndex,
					TxHash: *txHash,
				}
				spendedOut.Spend = wtxmgr.SpendF
				spendedOut.Address = addr
				spendedOut.SpendTo = &spendto

				err = w.TxStore.UpdateAddrTxOut(outns, spendedOut)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func (w *Wallet) SyncTx(order int64) (clijson.BlockHttpResult, error) {
	var block clijson.BlockHttpResult
	blockByte, err := w.Httpclient.getBlockByOrder(order)
	if err != nil {
		return block, err
	}
	//log.Info("SyncTx order:",order)
	if err := json.Unmarshal(blockByte, &block); err == nil {
		if !block.Txsvalid {
			log.Trace(fmt.Sprintf("block:%v err,txsvalid is false", block.Hash))
			return block, nil
		}
		txins, txouts, trrs, err := parseBlockTxs(block)
		if err != nil {
			return block, err
		}
		err = w.insertTx(txins, txouts, trrs)
		if err != nil {
			return block, err
		}

	} else {
		log.Error(err.Error())
		return block, err
	}
	//log.Info("tx:",tx)
	return block, nil
}

func parseTx(tr corejson.TxRawResult, height int32) ([]types.TxOutPoint, []wtxmgr.AddrTxOutput, error) {
	var txins []types.TxOutPoint
	var txouts []wtxmgr.AddrTxOutput
	blockhash, err := hash.NewHashFromStr(tr.BlockHash)
	if err != nil {
		return nil, nil, err
	}
	block := wtxmgr.Block{
		Hash:   *blockhash,
		Height: height,
	}
	txid, err := hash.NewHashFromStr(tr.Txid)
	if err != nil {
		return nil, nil, err
	}
	spend := wtxmgr.SpendZ
	if tr.Confirmations < config.Cfg.Confirmations{
		spend=wtxmgr.SpendT
	}
	for j := 0; j < len(tr.Vin); j++ {
		vi := tr.Vin[j]
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
				txin := types.TxOutPoint{
					Hash:     *hs,
					OutIndex: vi.Vout,
				}
				txins = append(txins, txin)
			}
		}
	}
	for k := 0; k < len(tr.Vout); k++ {
		vo := tr.Vout[k]
		if len(vo.ScriptPubKey.Addresses) == 0 {
			continue
		} else {
			txout := wtxmgr.AddrTxOutput{
				Address: vo.ScriptPubKey.Addresses[0],
				Txid:    *txid,
				Index:   uint32(k),
				Amount:  types.Amount(vo.Amount),
				Block:   block,
				Spend:   spend,
			}
			txouts = append(txouts, txout)
		}
	}

	return txins, txouts, nil
}

func parseBlockTxs(block clijson.BlockHttpResult) ([]types.TxOutPoint, []wtxmgr.AddrTxOutput, []corejson.TxRawResult, error) {
	var txins []types.TxOutPoint
	var txouts []wtxmgr.AddrTxOutput
	var tx []corejson.TxRawResult
	for _, tr := range block.Transactions {
		tx = append(tx, tr)
		tin, tout, err := parseTx(tr, block.Order)
		if err != nil {
			return nil, nil, nil, err
		} else {
			txins = append(txins, tin...)
			txouts = append(txouts, tout...)
		}
	}
	return txins, txouts, tx, nil
}

func (w *Wallet) GetSynceBlockHeight() int32 {
	height := w.Manager.SyncedTo().Height
	return height
}

var orderchan = make(chan int64, 20)

func (w *Wallet) handleBlock(order int64) {
	//for  {
	//order := <- orderchan
	_, er := w.SyncTx(order)
	if er != nil {
		fmt.Errorf("SyncTx err :", er.Error())
		return
	}
	//hs, err := hash.NewHashFromStr(br.Hash)
	//if err != nil {
	//	log.Info("blockhash string to hash  err:", err.Error())
	//	return
	//}
	//stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: br.Order}
	//err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
	//	ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
	//	err := w.Manager.SetSyncedTo(ns, stamp)
	//	if err != nil {
	//		log.Info("db err:", err.Error())
	//		return err
	//	}
	//	return nil
	//})
	//if err != nil {
	//	log.Info("blockhash string to hash  err:", err.Error())
	//	//continue
	//	return
	//}
	//}
}

func (w *Wallet) handleBlockSynced(order int64) error {
	//for  {
	//order := <- orderchan
	br, er := w.SyncTx(order)
	if er != nil {
		return er
	}
	hs, err := hash.NewHashFromStr(br.Hash)
	if err != nil {
		return fmt.Errorf("blockhash string to hash  err:", err.Error())
	}
	if br.Confirmations > config.Cfg.Confirmations {
		stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: br.Order}
		err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
			ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
			err := w.Manager.SetSyncedTo(ns, stamp)
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

func (w *Wallet) Updateblock(toHeight int64) error {
	var blockcount string
	var err error
	if toHeight == 0 {
		blockcount, err = w.Httpclient.getblockCount()
		if err != nil {
			//log.Info("getblockcount err:", err.Error())
			return err
		}
	} else {
		blockcount = strconv.FormatInt(toHeight, 10)
	}
	blockheight, err := strconv.ParseInt(blockcount, 10, 32)
	if err != nil {
		return err
	}
	h := int64((w.Manager.SyncedTo().Height))
	if h < blockheight {
		log.Trace(fmt.Sprintf("localheight:%d,blockheight:%d", h, blockheight))
		for h < blockheight {
			//orderchan <- h
			err := w.handleBlockSynced(h)
			if err != nil {
				return err
			} else {
				w.SyncHeight = int32(h)
				fmt.Fprintf(os.Stdout, "update blcok:%s/%s\r", strconv.FormatInt(h, 10), strconv.FormatInt(blockheight-1, 10))
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
func (w *Wallet) NextAccount(scope waddrmgr.KeyScope, name string) (uint32, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var (
		account uint32
		//props   *waddrmgr.AccountProperties
	)
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.NewAccount(addrmgrNs, name)
		if err != nil {
			return err
		}
		_, err = manager.AccountProperties(addrmgrNs, account)
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
func (w *Wallet) AccountBalances(scope waddrmgr.KeyScope,
	requiredConfs int32) ([]AccountBalanceResult, error) {
	aaas, err := w.GetAccountAndAddress(scope, requiredConfs)
	if err != nil {
		return nil, err
	}
	results := make([]AccountBalanceResult, len(aaas))
	for index, aaa := range aaas {
		results[index].AccountNumber = aaa.AccountNumber
		results[index].AccountName = aaa.AccountName
		unspendAmount := types.Amount(0)
		for _, addr := range aaa.AddrsOutput {
			unspendAmount = unspendAmount + addr.balance.UnspendAmount
		}
		if err != nil {
			return nil, err
		}
		results[index].AccountBalance = unspendAmount
	}
	return results, nil
}

// AccountNumber returns the account number for an account name under a
// particular key scope.
func (w *Wallet) AccountNumber(scope waddrmgr.KeyScope, accountName string) (uint32, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var account uint32
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.LookupAccount(addrmgrNs, accountName)
		return err
	})
	return account, err
}

// NewAddress returns the next external chained address for a wallet.
func (w *Wallet) NewAddress(
	scope waddrmgr.KeyScope, account uint32) (types.Address, error) {
	//chainClient, err := w.requireChainClient()
	//if err != nil {
	//	return nil, err
	//}

	var (
		addr types.Address
		//props *waddrmgr.AccountProperties
	)
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		addr, _, err = w.newAddress(addrmgrNs, account, scope)
		return err
	})
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func (w *Wallet) newAddress(addrmgrNs walletdb.ReadWriteBucket, account uint32,
	scope waddrmgr.KeyScope) (types.Address, *waddrmgr.AccountProperties, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, nil, err
	}

	// Get next address from wallet.
	addrs, err := manager.NextExternalAddresses(addrmgrNs, account, 1)
	if err != nil {
		return nil, nil, err
	}

	props, err := manager.AccountProperties(addrmgrNs, account)
	if err != nil {
		log.Error(fmt.Sprintf("Cannot fetch account properties for notification "+
			"after deriving next external address: %v", err))
		return nil, nil, err
	}

	return addrs[0].Address(), props, nil
}

// DumpWIFPrivateKey returns the WIF encoded private key for a
// single wallet address.
func (w *Wallet) DumpWIFPrivateKey(addr types.Address) (string, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		waddrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = w.Manager.Address(waddrmgrNs, addr)
		return err
	})
	if err != nil {
		return "", err
	}
	pka, ok := maddr.(waddrmgr.ManagedPubKeyAddress)
	if !ok {
		return "", fmt.Errorf("address %s is not a key type", addr)
	}
	//pri, err := pka.PrivKey()
	//log.Info("pri:%x\n", pri.SerializeSecret())
	wif, err := pka.ExportPrivKey()
	if err != nil {
		return "", err
	}
	return wif.String(), nil
}
func (w *Wallet) getPrivateKey(addr types.Address) (waddrmgr.ManagedPubKeyAddress, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		waddrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = w.Manager.Address(waddrmgrNs, addr)
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

func confirmed(minconf, txHeight, curHeight int32) bool {
	return confirms(txHeight, curHeight) >= minconf
}

// confirms returns the number of confirmations for a transaction in a block at
// height txHeight (or -1 for an unconfirmed tx) given the chain height
// curHeight.
func confirms(txHeight, curHeight int32) int32 {
	switch {
	case txHeight == -1, txHeight > curHeight:
		return 0
	default:
		return curHeight - txHeight
	}
}

// Unlock unlocks the wallet's address manager and relocks it after timeout has
// expired.  If the wallet is already unlocked and the new passphrase is
// correct, the current timeout is replaced with the new one.  The wallet will
// be locked if the passphrase is incorrect or any other error occurs during the
// unlock.
func (w *Wallet) Unlock(passphrase []byte, lock <-chan time.Time) error {
	log.Trace("wallet Unlock")
	err := make(chan error, 1)
	w.unlockRequests <- unlockRequest{
		passphrase: passphrase,
		lockAfter:  lock,
		err:        err,
	}
	log.Trace("wallet Unlock end")
	return <-err
}

//// Lock locks the wallet's address manager.
func (w *Wallet) Lock() {
	w.lockRequests <- struct{}{}
}

//// Locked returns whether the account manager for a wallet is locked.
func (w *Wallet) Locked() bool {
	return <-w.lockState
}

// quitChan atomically reads the quit channel.
func (w *Wallet) quitChan() <-chan struct{} {
	w.quitMu.Lock()
	c := w.quit
	w.quitMu.Unlock()
	return c
}

func (w *Wallet) UnLockManager(passphrase []byte) error {
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return w.Manager.Unlock(addrmgrNs, passphrase)
	})
	if err != nil {
		return err
	}
	return nil
}

// walletLocker manages the locked/unlocked state of a wallet.
func (w *Wallet) walletLocker() {
	log.Trace("wallet walletLocker")
	var timeout <-chan time.Time
	//holdChan := make(heldUnlock)
	quit := w.quitChan()
out:
	for {
		select {
		case req := <-w.unlockRequests:
			log.Trace("walletLocker,unlockRequests")
			err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
				addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
				return w.Manager.Unlock(addrmgrNs, req.passphrase)
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



		case w.lockState <- w.Manager.IsLocked():
			continue

		case <-quit:
			break out

		//case <-w.lockRequests:
		case <-timeout:
		}

		// Select statement fell through by an explicit lock or the
		// timer expiring.  Lock the manager here.
		timeout = nil
		err := w.Manager.Lock()
		if err != nil && !waddrmgr.IsError(err, waddrmgr.ErrLocked) {
			log.Error("Could not lock wallet", "err", err)
		} else {
			log.Info("The wallet has been locked")
		}
	}
	w.wg.Done()
}

// AccountAddresses returns the addresses for every created address for an
// account.
func (w *Wallet) AccountAddresses(account uint32) (addrs []types.Address, err error) {
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return w.Manager.ForEachAccountAddress(addrmgrNs, account, func(maddr waddrmgr.ManagedAddress) error {
			addrs = append(addrs, maddr.Address())
			return nil
		})
	})
	return
}

// AccountOfAddress finds the account that an address is associated with.
func (w *Wallet) AccountOfAddress(a types.Address) (uint32, error) {
	var account uint32
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		_, account, err = w.Manager.AddrAccount(addrmgrNs, a)
		return err
	})
	return account, err
}

// AccountName returns the name of an account.
func (w *Wallet) AccountName(scope waddrmgr.KeyScope, accountNumber uint32) (string, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	var accountName string
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		accountName, err = manager.AccountName(addrmgrNs, accountNumber)
		return err
	})
	return accountName, err
}

func (w *Wallet) GetUtxo(addr string) ([]wtxmgr.Utxo, error) {
	var txouts []wtxmgr.AddrTxOutput
	var utxos []wtxmgr.Utxo
	//var txins []*types.TxOutPoint
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadBucket(wtxmgr.BucketAddrtxout)
		hsoutns := outns.NestedReadBucket(hs)
		if hsoutns != nil {
			hsoutns.ForEach(func(k, v []byte) error {
				to := wtxmgr.AddrTxOutput{}
				err := wtxmgr.ReadAddrTxOutput(v, &to)
				if err != nil {
					log.Error("ReadAddrTxOutput err", "err", err.Error())
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
		uo := wtxmgr.Utxo{}
		if txout.Spend == wtxmgr.SpendZ {
			uo.Txid = txout.Txid.String()
			uo.Index = txout.Index
			uo.Amount = txout.Amount
			utxos = append(utxos, uo)
		}
	}
	return utxos, nil
}

// SendOutputs creates and sends payment transactions. It returns the
// transaction upon success.
func (w *Wallet) SendOutputs(outputs []*types.TxOutput, account int64, //uint32,
	minconf int32, satPerKb types.Amount) (*string, error) {

	// Ensure the outputs to be created adhere to the network's consensus
	// rules.
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
	aaars, err := w.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044, minconf)
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

		for _, output := range aaar.AddrsOutput {
			log.Trace(fmt.Sprintf("addr:%s,unspend:%v",output.Addr,output.balance.UnspendAmount))
			if output.balance.UnspendAmount > (payAmout + types.Amount(feeAmout)) {
				addr, err := address.DecodeAddress(output.Addr)
				if err != nil {
					return nil, err
				}
				frompkscipt, err := txscript.PayToAddrScript(addr)
				if err != nil {
					return nil, err
				}
				pri, err := w.getPrivateKey(addr)
				if err != nil {
					return nil, err
				}
				prikey, err := pri.PrivKey()
				if err != nil {
					return nil, err
				}
				prk = hex.EncodeToString(prikey.SerializeSecret())
				for _, output1 := range output.Txoutput {
					output1.Address = output.Addr
					if output1.Spend == wtxmgr.SpendZ && output1.SpendTo == nil {
						if output1.Amount >= payAmout && payAmout > types.Amount(0) {
							pre := types.NewOutPoint(&output1.Txid, output1.Index)
							tx.AddTxIn(types.NewTxInput(pre, nil))
							selfTxOut := types.NewTxOutput(uint64(output1.Amount-payAmout), frompkscipt)
							feeAmout = util.CalcMinRequiredTxRelayFee(int64(tx.SerializeSize()+selfTxOut.SerializeSize()), types.Amount(config.Cfg.MinTxFee))
							if (output1.Amount - payAmout) >= types.Amount(feeAmout) {
								selfTxOut.Amount = uint64(output1.Amount - payAmout - types.Amount(feeAmout))
								tx.AddTxOut(selfTxOut)
								payAmout = types.Amount(0)
								feeAmout = int64(0)
								sendAddrTxOutput = append(sendAddrTxOutput, output1)
								break b
							} else {
								if uint64(output1.Amount - payAmout) >0{
									selfTxOut.Amount = uint64(output1.Amount - payAmout)
									tx.AddTxOut(selfTxOut)
								}
								sendAddrTxOutput = append(sendAddrTxOutput, output1)
								payAmout = types.Amount(0)

							}
						} else if output1.Amount < payAmout && payAmout > types.Amount(0) {
							pre := types.NewOutPoint(&output1.Txid, output1.Index)
							tx.AddTxIn(types.NewTxInput(pre, nil))
							payAmout = payAmout - output1.Amount
							sendAddrTxOutput = append(sendAddrTxOutput, output1)
						} else if output1.Amount >= types.Amount(feeAmout) && payAmout == types.Amount(0) {
							pre := types.NewOutPoint(&output1.Txid, output1.Index)
							feeTxin := types.NewTxInput(pre, nil)
							feeTxOut := types.NewTxOutput(uint64(output1.Amount-types.Amount(feeAmout)), frompkscipt)
							feeAmout = util.CalcMinRequiredTxRelayFee(int64(tx.SerializeSize()+feeTxOut.SerializeSize()+feeTxin.SerializeSize()), types.Amount(config.Cfg.MinTxFee))
							if output1.Amount >= types.Amount(feeAmout) {
								feeTxOut.Amount = uint64(output1.Amount - types.Amount(feeAmout))
								tx.AddTxIn(feeTxin)
								tx.AddTxOut(feeTxOut)
								payAmout = types.Amount(0)
								feeAmout = int64(0)
								sendAddrTxOutput = append(sendAddrTxOutput, output1)
								break b
							}
						}
					}
				}
			}
		}
	}
	if payAmout.ToCoin() != types.Amount(0).ToCoin() || feeAmout != 0 {
		log.Trace("payAmout", "payAmout", payAmout)
		log.Trace("feeAmout", "feeAmout", feeAmout)
		//log.Info("balance is not enough")
		return nil, fmt.Errorf("balance is not enough")
	}
	b, err := tx.Serialize()
	if err != nil {
		return nil, err
	}
	signTx, err := qx.TxSign(prk, hex.EncodeToString(b), w.chainParams.Name)
	if err != nil {
		return nil, err
	}
	log.Trace(fmt.Sprintf("signTx size:%v", len(signTx)), "signTx", signTx)
	msg, err := w.Httpclient.SendRawTransaction(signTx, false)
	if err != nil{
		return nil, err
	} else {
		msg=strings.ReplaceAll(msg,"\"","")
		log.Info("SendRawTransaction txSign response msg", "msg", msg)
	}

	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		for _, txoutput := range sendAddrTxOutput {
			txoutput.Spend = wtxmgr.SpendF
			err = w.TxStore.UpdateAddrTxOut(outns, &txoutput)
			if err != nil {
				log.Error("UpdateAddrTxOut to spend err", "err", err.Error())
				return err
			}
		}
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
func (w *Wallet)  SendPairs( amounts map[string]types.Amount,
	account int64 /*uint32*/, minconf int32, feeSatPerKb types.Amount) (string, error) {
	//check,err := w.Httpclient.CheckSyncUpdate(int64(w.SyncHeight))
	//
	//if check ==false{
	//	return "",err
	//}
	outputs, err := makeOutputs(amounts, w.ChainParams())
	if err != nil {
		return "", err
	}
	tx, err := w.SendOutputs(outputs, account, minconf, feeSatPerKb)
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
func makeOutputs(pairs map[string]types.Amount, chainParams *chaincfg.Params) ([]*types.TxOutput, error) {
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
