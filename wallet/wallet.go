// Copyright (c) 2018-2020 The qitmeer developers
// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer/common/marshal"
	"github.com/Qitmeer/qitmeer/core/message"
	"github.com/Qitmeer/qitmeer/crypto/ecc"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Qitmeer/qitmeer-wallet/util"

	wt "github.com/Qitmeer/qitmeer-wallet/types"
	"github.com/Qitmeer/qitmeer/common/hash"
	"github.com/Qitmeer/qitmeer/core/address"
	corejson "github.com/Qitmeer/qitmeer/core/json"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/log"
	chaincfg "github.com/Qitmeer/qitmeer/params"

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
	InsecurePubPassphrase   = "public"
	webUpdateBlockTicker    = 30
	defaultNewAddressNumber = 1
)

var (
	// Namespace bucket keys.
	waddrmgrNamespaceKey = []byte("waddrmgr")
	wtxmgrNamespaceKey   = []byte("wtxmgr")
)

var UploadRun = false

type Wallet struct {
	cfg *config.Config

	// Data stores
	db      walletdb.DB
	Manager *waddrmgr.Manager
	TxStore *wtxmgr.Store

	HttpClient *httpConfig

	// Channels for the manager locker.
	unlockRequests chan unlockRequest
	lockRequests   chan struct{}
	lockState      chan bool

	chainParams *chaincfg.Params
	wg          sync.WaitGroup

	started bool
	quit    chan struct{}
	quitMu  sync.Mutex

	SyncHeight     int32
	SyncMainHeight int32
}

// Start starts the goroutines necessary to manage a wallet.
func (w *Wallet) Start() {
	w.quitMu.Lock()
	select {
	case <-w.quit:
		// Restart the wallet goroutines after shutdown finishes.
		w.WaitForShutdown()
		w.quit = make(chan struct{})
	default:
		// Ignore when the wallet is still running.
		if w.started {
			w.quitMu.Unlock()
			return
		}
		w.started = true
	}
	w.quitMu.Unlock()

	w.wg.Add(1)
	go w.walletLocker()

	go func() {

		updateBlockTicker := time.NewTicker(webUpdateBlockTicker * time.Second)
		for {
			select {
			case <-updateBlockTicker.C:
				if UploadRun == false {
					log.Trace("Updateblock start")
					UploadRun = true
					err := w.UpdateBlock(0)
					if err != nil {
						log.Error("Start.Updateblock err", "err", err.Error())
					}
					UploadRun = false
				}
			}
		}

	}()
}

// quitChan atomically reads the quit channel.
func (w *Wallet) quitChan() <-chan struct{} {
	w.quitMu.Lock()
	c := w.quit
	w.quitMu.Unlock()
	return c
}

// Stop signals all wallet goroutines to shutdown.
func (w *Wallet) Stop() {
	w.quitMu.Lock()
	quit := w.quit
	w.quitMu.Unlock()

	select {
	case <-quit:
	default:
		close(quit)
	}
}

// ShuttingDown returns whether the wallet is currently in the process of
// shutting down or not.
func (w *Wallet) ShuttingDown() bool {
	select {
	case <-w.quitChan():
		return true
	default:
		return false
	}
}

// WaitForShutdown blocks until all wallet goroutines have finished executing.
func (w *Wallet) WaitForShutdown() {
	w.wg.Wait()
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

// ImportPrivateKey imports a private key to the wallet and writes the new
// wallet to disk.
//
// NOTE: If a block stamp is not provided, then the wallet's birthday will be
// set to the genesis block of the corresponding chain.
func (w *Wallet) ImportPrivateKey(scope waddrmgr.KeyScope, wif *utils.WIF) (string, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	// Attempt to import private key into wallet.
	var addr types.Address
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
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

	var (
		addrMgr *waddrmgr.Manager
		txMgr   *wtxmgr.Store
	)

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
		txMgr, err = wtxmgr.Open(txMgrBucket, params)
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
		cfg:            cfg,
		db:             db,
		Manager:        addrMgr,
		TxStore:        txMgr,
		unlockRequests: make(chan unlockRequest),
		lockRequests:   make(chan struct{}),
		lockState:      make(chan bool),
		chainParams:    params,
		quit:           make(chan struct{}),
	}

	return w, nil
}

func (w *Wallet) GetTx(txId string) (corejson.TxRawResult, error) {

	trx := corejson.TxRawResult{}
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
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

func (w *Wallet) GetAccountAndAddress(scope waddrmgr.KeyScope) ([]AccountAndAddressResult, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, err
	}
	var results []AccountAndAddressResult
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
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
			adds, err := w.AccountAddresses(results[k].AccountNumber)
			if err != nil {
				return err
			}
			var addrOutputs []AddrAndAddrTxOutput
			for _, addr := range adds {
				addrOutput, err := w.getAddrAndAddrTxOutputByAddr(addr.Encode())
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

func (w *Wallet) getAddrAndAddrTxOutputByAddr(addr string) (*AddrAndAddrTxOutput, error) {

	ato := AddrAndAddrTxOutput{}
	b := Balance{}
	var txOuts wtxmgr.AddrTxOutputs
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
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
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(txOuts))

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
	PageUseDefault  = -1
	PageDefaultNo   = 1
	PageDefaultSize = 10
	PageMaxSize     = 1000000000
	FilterIn        = 0
	FilterOut       = 1
	FilterAll       = 2
)

/**
request all the transactions that affect a specific address,
a transaction can have MULTIPLE payments and affect MULTIPLE addresses

sType 0 Turn in 1 Turn out 2 all no page
*/
func (w *Wallet) GetListTxByAddr(addr string, sType int, pageNo int, pageSize int) (*clijson.PageTxRawResult, error) {

	bill, err := w.getPagedBillByAddr(addr, sType, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	result := clijson.PageTxRawResult{}
	result.Page = int32(pageNo)
	result.PageSize = int32(pageSize)
	result.Total = int32(bill.Len())

	var transactions []corejson.TxRawResult
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		txNs := ns.NestedReadBucket(wtxmgr.BucketTxJson)
		for _, b := range *bill {
			txHs := b.TxID
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

// request the bill of a specific address, a bill is the log of payments,
// which are the effects that a transaction makes on a specific address
// a payment can affect only ONE address
func (w *Wallet) GetBillByAddr(addr string, filter int, pageNo int, pageSize int) (*clijson.PagedBillResult, error) {
	bill, err := w.getPagedBillByAddr(addr, filter, pageNo, pageSize)
	if err != nil {
		return nil, err
	}

	res := clijson.PagedBillResult{}
	res.PageNo = int32(pageNo)
	res.PageSize = int32(pageSize)
	res.Total = int32(bill.Len())

	for _, p := range *bill {
		res.Bill = append(res.Bill, clijson.PaymentResult{
			TxID:      p.TxID.String(),
			Variation: p.Variation,
		})
	}

	return &res, nil
}

func (w *Wallet) getPagedBillByAddr(addr string, filter int, pageNo int, pageSize int) (*wt.Bill, error) {
	at, err := w.getAddrAndAddrTxOutputByAddr(addr)
	if err != nil {
		return nil, err
	}
	if pageNo == 0 {
		pageNo = PageDefaultNo
	}
	if pageSize == 0 {
		pageSize = PageDefaultSize
	}
	startIndex := (pageNo - 1) * pageSize
	var endIndex int
	var allTxs wt.Bill
	var inTxs wt.Bill
	var outTxs wt.Bill
	var dataLen int

	allMap := make(map[hash.Hash]wt.Payment)

	for _, o := range at.Txoutput {

		txOut, found := allMap[o.TxId]
		if found {
			txOut.Variation += int64(o.Amount)
		} else {
			txOut.TxID = o.TxId
			txOut.Variation = int64(o.Amount)
			txOut.BlockHash = o.Block.Hash
			txOut.BlockOrder = uint32(o.Block.Height)
		}
		//log.Debug(fmt.Sprintf("%s %v %v", o.TxId.String(), float64(o.Amount)/math.Pow10(8), float64(txOut.Amount)/math.Pow10(8)))

		allMap[o.TxId] = txOut

		if o.SpendTo != nil {
			txOut, found := allMap[o.SpendTo.TxHash]
			if found {
				txOut.Variation -= int64(o.Amount)
			} else {
				txOut.TxID = o.SpendTo.TxHash
				txOut.Variation = -int64(o.Amount)
				// ToDo: add Block to SpendTo
				txOut.BlockHash = o.Block.Hash
				txOut.BlockOrder = uint32(o.Block.Height)
			}
			allMap[o.SpendTo.TxHash] = txOut
			//log.Debug(fmt.Sprintf("%s %v %v", o.SpendTo.TxHash.String(), float64(-o.Amount)/math.Pow10(8), float64(txOut.Amount)/math.Pow10(8)))
		}
	}

	for _, out := range allMap {
		if out.Variation > 0 {
			inTxs = append(inTxs, out)
		} else {
			outTxs = append(outTxs, out)
		}
	}

	switch filter {
	case FilterIn:
		allTxs = inTxs
	case FilterOut:
		allTxs = outTxs
	case FilterAll:
		allTxs = append(inTxs, outTxs...)
	default:
		return nil, fmt.Errorf("err filter:%d", filter)
	}

	sort.Sort(allTxs)

	dataLen = len(allTxs)
	if pageNo < 0 {
		pageNo = PageDefaultNo
		pageSize = PageMaxSize
	} else {
		if startIndex > dataLen {
			return nil, fmt.Errorf("no data, index:%d len:%d", startIndex, dataLen)
		} else {
			if (startIndex + pageSize) > dataLen {
				endIndex = dataLen
			} else {
				endIndex = startIndex + pageSize
			}
			allTxs = allTxs[startIndex:endIndex]
		}
	}
	return &allTxs, nil
}

func (w *Wallet) GetBalance(addr string) (*Balance, error) {
	if addr == "" {
		return nil, errors.New("addr is nil")
	}
	res, err := w.getAddrAndAddrTxOutputByAddr(addr)
	if err != nil {
		return nil, err
	}
	return &res.balance, nil
}
func (w *Wallet) GetTxSpendInfo(txId string) ([]*wtxmgr.AddrTxOutput, error) {
	var atos []*wtxmgr.AddrTxOutput
	txHash, err := hash.NewHashFromStr(txId)
	if err != nil {
		return nil, err
	}
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		rb := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		txNrb := rb.NestedReadWriteBucket(wtxmgr.BucketTxJson)
		outNrb := rb.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		v := txNrb.Get(txHash.Bytes())
		if v == nil {
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
			var ato, err = w.TxStore.GetAddrTxOut(outNrb, addr, top)
			if err != nil {
				return err
			}
			ato.Address = addr
			atos = append(atos, ato)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return atos, nil
}

func (w *Wallet) insertTx(txins []wtxmgr.TxInputPoint, txouts []wtxmgr.AddrTxOutput, trrs []corejson.TxRawResult) error {
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
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
			err := w.TxStore.InsertAddrTxOut(outNs, &txo)
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
			spendOut, err := w.TxStore.GetAddrTxOut(outNs, addr, txi.TxOutPoint)
			if err != nil {
				return err
			}

			spendOut.Spend = wtxmgr.SpendStatusSpend
			spendOut.Address = addr
			spendOut.SpendTo = &txi.SpendTo
			err = w.TxStore.UpdateAddrTxOut(outNs, spendOut)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (w *Wallet) SyncTx(order int64) (clijson.BlockHttpResult, error) {
	var block clijson.BlockHttpResult
	blockByte, err := w.HttpClient.getBlockByOrder(order)
	if err != nil {
		return block, err
	}
	if err := json.Unmarshal(blockByte, &block); err == nil {
		if !block.Txsvalid {
			log.Trace(fmt.Sprintf("block:%v err,txsvalid is false", block.Hash))
			return block, nil
		}
		isBlue, err := w.HttpClient.isBlue(block.Hash)
		if err != nil {
			return block, err
		}
		block.IsBlue = isBlue
		if !block.IsBlue {
			log.Trace(fmt.Sprintf("block:%v is not blue", block.Hash))
		}
		txIns, txOuts, trRs, err := parseBlockTxs(block)
		if err != nil {
			return block, err
		}
		err = w.insertTx(txIns, txOuts, trRs)
		if err != nil {
			return block, err
		}

	} else {
		log.Error(err.Error())
		return block, err
	}
	return block, nil
}

func parseTx(tr corejson.TxRawResult, height, mainHeight int32, isBlue bool) ([]wtxmgr.TxInputPoint, []wtxmgr.AddrTxOutput, error) {
	var txins []wtxmgr.TxInputPoint
	var txouts []wtxmgr.AddrTxOutput
	blockhash, err := hash.NewHashFromStr(tr.BlockHash)
	if err != nil {
		return nil, nil, err
	}
	block := wtxmgr.Block{
		Hash:       *blockhash,
		Height:     height,
		MainHeight: mainHeight,
	}
	txId, err := hash.NewHashFromStr(tr.Txid)
	if err != nil {
		return nil, nil, err
	}
	spend := wtxmgr.SpendStatusUnspent
	if tr.Confirmations < config.Cfg.Confirmations {
		spend = wtxmgr.SpendStatusUnconfirmed
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
				spendTo := wtxmgr.SpendTo{
					Index:  uint32(i),
					TxHash: *txId,
				}
				txIn := wtxmgr.TxInputPoint{
					TxOutPoint: txOutPoint,
					SpendTo:    spendTo,
				}
				txins = append(txins, txIn)
				spend = wtxmgr.SpendStatusUnspent
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
				IsBlue:  isBlue,
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
		tin, tout, err := parseTx(tr, block.Order, block.Height, block.IsBlue)
		if err != nil {
			return nil, nil, nil, err
		} else {
			txIns = append(txIns, tin...)
			txOuts = append(txOuts, tout...)
		}
	}
	return txIns, txOuts, tx, nil
}

func (w *Wallet) GetSyncBlockHeight() int32 {
	height := w.Manager.SyncedTo().Height
	return height
}

func (w *Wallet) SetSyncedToNum(order int64) error {
	var block clijson.BlockHttpResult
	blockByte, err := w.HttpClient.getBlockByOrder(order)
	if err != nil {
		return err
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
		return nil
	} else {
		log.Error(err.Error())
		return err
	}
}

func (w *Wallet) handleBlockSynced(order int64) error {
	br, er := w.SyncTx(order)
	if er != nil {
		return er
	}
	hs, err := hash.NewHashFromStr(br.Hash)
	if err != nil {
		return fmt.Errorf("blockhash string to hash  err:%s", err.Error())
	}
	if br.Confirmations > config.Cfg.Confirmations {
		stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: br.Order}
		err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
			ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
			err := w.Manager.SetSyncedTo(ns, stamp)
			if err != nil {
				return err
			}
			w.SyncMainHeight = br.Height
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Wallet) UpdateBlock(toHeight int64) error {
	var blockCount string
	var err error
	if toHeight == 0 {
		blockCount, err = w.HttpClient.getblockCount()
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
	h := int64(w.Manager.SyncedTo().Height)
	if h < blockHeight {
		log.Trace(fmt.Sprintf("localheight:%d,blockHeight:%d", h, blockHeight))
		for h < blockHeight {
			err := w.handleBlockSynced(h)
			if err != nil {
				return err
			} else {
				w.SyncHeight = int32(h)
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
func (w *Wallet) NextAccount(scope waddrmgr.KeyScope, name string) (uint32, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var (
		account uint32
	)
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
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
func (w *Wallet) AccountBalances(scope waddrmgr.KeyScope) ([]AccountBalanceResult, error) {
	aaaRs, err := w.GetAccountAndAddress(scope)
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
func (w *Wallet) AccountNumber(scope waddrmgr.KeyScope, accountName string) (uint32, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return 0, err
	}

	var account uint32
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.LookupAccount(addrMgrNs, accountName)
		return err
	})
	return account, err
}

// NewAddress returns the next external chained address for a wallet.
func (w *Wallet) NewAddress(
	scope waddrmgr.KeyScope, account uint32) (types.Address, error) {
	var (
		addr types.Address
	)
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrMgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		addr, _, err = w.newAddress(addrMgrNs, account, scope)
		return err
	})
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func (w *Wallet) newAddress(addrMgrNs walletdb.ReadWriteBucket, account uint32,
	scope waddrmgr.KeyScope) (types.Address, *waddrmgr.AccountProperties, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
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
func (w *Wallet) DumpWIFPrivateKey(addr types.Address) (string, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		waddrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = w.Manager.Address(waddrMgrNs, addr)
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
func (w *Wallet) getPrivateKey(addr types.Address) (waddrmgr.ManagedPubKeyAddress, error) {
	var maddr waddrmgr.ManagedAddress
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		waddrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		// Get private key from wallet if it exists.
		var err error
		maddr, err = w.Manager.Address(waddrMgrNs, addr)
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

func (w *Wallet) UnLockManager(passphrase []byte) error {
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return w.Manager.Unlock(addrMgrNs, passphrase)
	})
	if err != nil {
		return err
	}
	return nil
}

// walletLocker manages the locked/unlocked state of a wallet.
func (w *Wallet) walletLocker() {
	var timeout <-chan time.Time
	quit := w.quitChan()
out:
	for {
		select {
		case req := <-w.unlockRequests:
			err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
				addMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
				return w.Manager.Unlock(addMgrNs, req.passphrase)
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

		case <-w.lockRequests:
		case <-timeout:
		}

		// Select statement fell through by an explicit lock or the
		// timer expiring.  Lock the manager here.
		timeout = nil
		err := w.Manager.Lock()
		if err != nil && !waddrmgr.IsError(err, waddrmgr.ErrLocked) {
			log.Error("Could not lock wallet: ", err.Error())
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
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return w.Manager.ForEachAccountAddress(addrMgrNs, account, func(mAddr waddrmgr.ManagedAddress) error {
			addrs = append(addrs, mAddr.Address())
			return nil
		})
	})
	return
}

// AccountOfAddress finds the account that an address is associated with.
func (w *Wallet) AccountOfAddress(a types.Address) (uint32, error) {
	var account uint32
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		_, account, err = w.Manager.AddrAccount(addrMgrNs, a)
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
		addrMgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		var err error
		accountName, err = manager.AccountName(addrMgrNs, accountNumber)
		return err
	})
	return accountName, err
}

func (w *Wallet) GetUtxo(addr string) ([]wtxmgr.UTxo, error) {
	var txouts []wtxmgr.AddrTxOutput
	var utxos []wtxmgr.UTxo
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
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
func (w *Wallet) SendOutputs(outputs []*types.TxOutput, account int64, satPerKb types.Amount) (*string, error) {

	// Ensure the outputs to be created adhere to the network's consensus
	// rules.
	syncSendOutputs.Lock()
	defer syncSendOutputs.Unlock()
	tx := types.NewTransaction()
	payAmount := types.Amount(0)
	feeAmount := int64(0)
	for _, output := range outputs {
		if err := txrules.CheckOutput(output, satPerKb); err != nil {
			return nil, err
		}
		payAmount = payAmount + types.Amount(output.Amount)
		tx.AddTxOut(output)
	}
	aaars, err := w.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}

	var sendAddrTxOutput []wtxmgr.AddrTxOutput
	//var prk string
b:
	for _, aaar := range aaars {

		if int64(aaar.AccountNumber) != account && account != waddrmgr.AccountMergePayNum {
			continue
		}

		for _, addroutput := range aaar.AddrsOutput {
			log.Trace(fmt.Sprintf("addr:%s,unspend:%v", addroutput.Addr, addroutput.balance.UnspendAmount))
			if addroutput.balance.UnspendAmount > 0 {
				addr, err := address.DecodeAddress(addroutput.Addr)
				if err != nil {
					return nil, err
				}
				frompkscipt, err := txscript.PayToAddrScript(addr)
				if err != nil {
					return nil, err
				}
				addrByte := []byte(addroutput.Addr)

				for _, output := range addroutput.Txoutput {
					output.Address = addroutput.Addr

					mature := false
					if outTx, err := w.GetTx(output.TxId.String()); err == nil {
						if outTx.Vin[0].IsCoinBase() {
							if output.IsBlue {
								confirms := uint16(w.SyncMainHeight - output.Block.MainHeight + 1)
								if confirms > uint16(config.Cfg.CoinbaseMaturity) {
									mature = true
								}
							}
						} else {
							mature = true
						}
					}

					if output.Spend == wtxmgr.SpendStatusUnspent && mature {
						if payAmount > 0 && feeAmount == 0 {
							if output.Amount > payAmount {
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, addrByte))
								selfTxOut := types.NewTxOutput(uint64(output.Amount-payAmount), frompkscipt)
								feeAmount = util.CalcMinRequiredTxRelayFee(int64(tx.SerializeSize()+selfTxOut.SerializeSize()), types.Amount(config.Cfg.MinTxFee))
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								if (output.Amount - payAmount - types.Amount(feeAmount)) >= 0 {
									selfTxOut.Amount = uint64(output.Amount - payAmount - types.Amount(feeAmount))
									if selfTxOut.Amount > 0 {
										tx.AddTxOut(selfTxOut)
									}
									payAmount = 0
									feeAmount = 0
									break b
								} else {
									selfTxOut.Amount = uint64(output.Amount - payAmount)
									payAmount = 0
									tx.AddTxOut(selfTxOut)
								}

							} else {
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, addrByte))
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								payAmount = payAmount - output.Amount
								if payAmount == 0 {
									feeAmount = util.CalcMinRequiredTxRelayFee(int64(tx.SerializeSize()), types.Amount(config.Cfg.MinTxFee))
								}
							}
						} else if payAmount == 0 && feeAmount > 0 {
							if output.Amount >= types.Amount(feeAmount) {
								input := types.NewOutPoint(&output.TxId, output.Index)
								tx.AddTxIn(types.NewTxInput(input, addrByte))
								selfTxOut := types.NewTxOutput(uint64(output.Amount-types.Amount(feeAmount)), frompkscipt)
								if selfTxOut.Amount > 0 {
									tx.AddTxOut(selfTxOut)
								}
								sendAddrTxOutput = append(sendAddrTxOutput, output)
								feeAmount = 0
								break b
							} else {
								log.Trace("utxo < feeAmount")
							}

						} else {
							log.Trace(fmt.Sprintf("system err payAmount :%v ,feeAmount :%v\n", payAmount, feeAmount))
							return nil, fmt.Errorf("system err payAmount :%v ,feeAmount :%v\n", payAmount, feeAmount)
						}
					}
				}
			}
			//}
		}
	}
	if payAmount.ToCoin() != types.Amount(0).ToCoin() || feeAmount != 0 {
		log.Trace("payAmount", "payAmount", payAmount)
		log.Trace("feeAmount", "feeAmount", feeAmount)
		return nil, fmt.Errorf("balance is not enough,please deduct the service charge:%v", types.Amount(feeAmount).ToCoin())
	}

	signTx, err := w.multiAddressMergeSign(*tx, w.chainParams.Name)
	if err != nil {
		return nil, err
	}
	log.Trace(fmt.Sprintf("signTx size:%v", len(signTx)), "signTx", signTx)
	msg, err := w.HttpClient.SendRawTransaction(signTx, false)
	if err != nil {
		log.Trace("SendRawTransaction txSign err ", "err", err.Error())
		return nil, err
	} else {
		msg = strings.ReplaceAll(msg, "\"", "")
		log.Trace("SendRawTransaction txSign response msg", "msg", msg)
	}

	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
		for _, txoutput := range sendAddrTxOutput {
			txoutput.Spend = wtxmgr.SpendStatusSpend
			err = w.TxStore.UpdateAddrTxOut(outns, &txoutput)
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

// Multi address merge signature
func (w *Wallet) multiAddressMergeSign(redeemTx types.Transaction, network string) (string, error) {

	var param *chaincfg.Params
	switch network {
	case "mainnet":
		param = &chaincfg.MainNetParams
	case "testnet":
		param = &chaincfg.TestNetParams
	case "privnet":
		param = &chaincfg.PrivNetParams
	case "mixnet":
		param = &chaincfg.MixNetParams
	}

	var sigScripts [][]byte
	for i := range redeemTx.TxIn {
		addrByte := redeemTx.TxIn[i].SignScript
		addr, err := address.DecodeAddress(string(addrByte))
		if err != nil {
			return "", err
		}
		pri, err := w.getPrivateKey(addr)
		if err != nil {
			return "", err
		}
		priKey, err := pri.PrivKey()
		if err != nil {
			return "", err
		}
		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", err
		}
		var kdb txscript.KeyClosure = func(types.Address) (ecc.PrivateKey, bool, error) {
			return priKey, true, nil // compressed is true
		}
		sigScript, err := txscript.SignTxOutput(param, &redeemTx, i, pkScript, txscript.SigHashAll, kdb, nil, nil, ecc.ECDSA_Secp256k1)
		if err != nil {
			return "", err
		}
		sigScripts = append(sigScripts, sigScript)
	}

	for i2 := range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx: &redeemTx})
	if err != nil {
		return "", err
	}
	return mtxHex, nil
}

//sendPairs creates and sends payment transactions.
//It returns the transaction hash in string format upon success
//All errors are returned in btcjson.RPCError format
func (w *Wallet) SendPairs(amounts map[string]types.Amount,
	account int64, feeSatPerKb types.Amount) (string, error) {
	check, err := w.HttpClient.CheckSyncUpdate(int64(w.Manager.SyncedTo().Height))

	if check == false {
		return "", err
	}
	outputs, err := makeOutputs(amounts)
	if err != nil {
		return "", err
	}
	tx, err := w.SendOutputs(outputs, account, feeSatPerKb)
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
			return nil, fmt.Errorf("cannot decode address: %s,address:%s", err, addrStr)
		}

		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("cannot create txout script: %s", err)
		}

		outputs = append(outputs, types.NewTxOutput(uint64(amt), pkScript))
	}
	return outputs, nil
}
