// Copyright (c) 2018-2020 The qitmeer developers
// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package wallet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/util"
	"github.com/Qitmeer/qng/common/marshal"
	"github.com/Qitmeer/qng/qx"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	wt "github.com/Qitmeer/qitmeer-wallet/types"
	"github.com/Qitmeer/qng/common/hash"
	"github.com/Qitmeer/qng/core/address"
	corejson "github.com/Qitmeer/qng/core/json"
	j "github.com/Qitmeer/qng/core/json"
	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/engine/txscript"
	"github.com/Qitmeer/qng/log"
	chaincfg "github.com/Qitmeer/qng/params"
	"github.com/Qitmeer/qng/rpc/client"
	"github.com/Qitmeer/qng/rpc/client/cmds"

	"github.com/Qitmeer/qitmeer-wallet/config"
	clijson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/waddrmgs"
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
	tokenmgrNamespaceKey = []byte("tknmgr")
)

const CoinBaseMaturity = 720

type Wallet struct {
	cfg *config.Config

	// Data stores
	db      walletdb.DB
	Manager *waddrmgr.Manager
	TxStore *wtxmgr.Store
	tokens  *QitmeerToken

	HttpClient *httpConfig

	notificationRpc *client.Client

	// Channels for the manager locker.
	unlockRequests chan unlockRequest
	lockRequests   chan struct{}
	lockState      chan bool

	chainParams *chaincfg.Params
	wg          *sync.WaitGroup

	started   bool
	UploadRun bool

	quit   chan struct{}
	quitMu sync.Mutex

	syncAll    bool
	syncLatest bool
	syncOrder  uint32
	toOrder    uint32
	syncQuit   chan struct{}
	syncWg     *sync.WaitGroup
	scanEnd    chan struct{}
	orderMutex sync.RWMutex
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

		//updateBlockTicker := time.NewTicker(webUpdateBlockTicker * time.Second)
		updateBlockTicker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-updateBlockTicker.C:
				if w.UploadRun == false {
					log.Trace("Updateblock start")
					w.UploadRun = true
					err := w.UpdateBlock(0)
					if err != nil {
						log.Error("Start.Updateblock err", "err", err.Error())
					}
					w.UploadRun = false
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

type Amount struct {
	Value int64
	Id    types.CoinID `json:"-"`
}

func NewAmount(value int64, id types.CoinID) *Amount {
	return &Amount{
		Value: 0,
		Id:    0,
	}
}

func (a *Amount) ToCoin() float64 {
	return (&types.Amount{Value: a.Value, Id: a.Id}).ToCoin()
}

type Value struct {
	TotalAmount       int64 // 总余额
	UnspentAmount     int64 // 可用余额
	LockAmount        int64 // 锁定
	UnconfirmedAmount int64 // 待确认
	SpendAmount       int64 // 已花费
}

type Balance struct {
	TotalAmount       *Amount // 总余额
	UnspentAmount     *Amount // 可用余额
	LockAmount        *Amount // 锁定
	UnconfirmedAmount *Amount // 待确认
	SpendAmount       *Amount // 已花费
}

func NewBalance(coinId types.CoinID) *Balance {
	return &Balance{
		TotalAmount:       &Amount{Value: 0, Id: coinId},
		UnspentAmount:     &Amount{Value: 0, Id: coinId},
		LockAmount:        &Amount{Value: 0, Id: coinId},
		UnconfirmedAmount: &Amount{Value: 0, Id: coinId},
		SpendAmount:       &Amount{Value: 0, Id: coinId},
	}
}

// AccountBalanceResult is a single result for the Wallet.AccountBalances method.
type AccountBalanceResult struct {
	AccountNumber      uint32 `json:"-"`
	AccountName        string
	AccountBalanceList []Balance
}
type AccountAndAddressResult struct {
	AccountNumber uint32
	AccountName   string
	AddrsOutput   []AddrAndAddrTxOutput
}
type AccountAddress struct {
	AccountNumber uint32
	AccountName   string
	Addrs         []types.Address
}
type AddrAndAddrTxOutput struct {
	Addr        string
	balanceMap  map[types.CoinID]Balance
	TxoutputMap map[types.CoinID][]wtxmgr.AddrTxOutput
}

func NewAddrAndAddrTxOutput() *AddrAndAddrTxOutput {
	return &AddrAndAddrTxOutput{
		Addr:        "",
		balanceMap:  map[types.CoinID]Balance{},
		TxoutputMap: map[types.CoinID][]wtxmgr.AddrTxOutput{},
	}
}

func (w *Wallet) SetConfig(cfg *config.Config) {
	w.cfg = cfg
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
		_, err = tx.CreateTopLevelBucket(tokenmgrNamespaceKey)
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
		tokens  *QitmeerToken
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
		tokenBucket := tx.ReadWriteBucket(tokenmgrNamespaceKey)
		tokens = NewQitmeerToken(tokenBucket)
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
		wg:             &sync.WaitGroup{},
		syncWg:         &sync.WaitGroup{},
		cfg:            cfg,
		db:             db,
		Manager:        addrMgr,
		TxStore:        txMgr,
		tokens:         tokens,
		unlockRequests: make(chan unlockRequest),
		lockRequests:   make(chan struct{}),
		lockState:      make(chan bool),
		chainParams:    params,
		quit:           make(chan struct{}),
		syncQuit:       make(chan struct{}, 1),
		scanEnd:        make(chan struct{}, 1),
	}

	return w, nil
}

func NewNotificationRpc(cfg *config.Config, handlers client.NotificationHandlers) (*client.Client, error) {
	connCfg := &client.ConnConfig{
		Host:               cfg.QServer,
		Endpoint:           "ws",
		User:               cfg.QUser,
		Pass:               cfg.QPass,
		DisableTLS:         cfg.QNoTLS,
		HTTPPostMode:       false,
		InsecureSkipVerify: cfg.QTLSSkipVerify,
	}
	if !connCfg.DisableTLS {
		certs, err := ioutil.ReadFile(cfg.QCert)
		if err != nil {
			return nil, err
		}
		connCfg.Certificates = certs
	}

	client, err := client.New(connCfg, &handlers)
	if err != nil {
		return nil, err
	}
	// Register for block connect and disconnect notifications.
	if err := client.NotifyBlocks(); err != nil {
		return nil, err
	}
	return client, nil
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

func (w *Wallet) GetAccountAddress(scope waddrmgr.KeyScope) ([]types.Address, error) {
	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, err
	}
	var accountList []AccountAddress
	var rs []types.Address
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		lastAcct, err := manager.LastAccount(addrNs)
		if err != nil {
			return err
		}
		accountList = make([]AccountAddress, lastAcct+2)
		for i := range accountList[:len(accountList)-1] {
			accountName, err := manager.AccountName(addrNs, uint32(i))
			if err != nil {
				return err
			}
			accountList[i].AccountNumber = uint32(i)
			accountList[i].AccountName = accountName
		}
		accountList[len(accountList)-1].AccountNumber = waddrmgr.ImportedAddrAccount
		accountList[len(accountList)-1].AccountName = waddrmgr.ImportedAddrAccountName
		for k := range accountList {
			adds, err := w.AccountAddresses(accountList[k].AccountNumber)
			if err != nil {
				return err
			}
			if adds != nil {
				rs = append(rs, adds...)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rs, err
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

func (w *Wallet) GetAddress(scope waddrmgr.KeyScope, account int) ([]AccountAndAddressResult, error) {
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

func (w *Wallet) getAddrTxOutputByCoin(addr string, coin types.CoinID) (wtxmgr.AddrTxOutputs, error) {
	var txOuts wtxmgr.AddrTxOutputs
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outNs := ns.NestedReadBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, types.CoinID(coin)))
		if outNs == nil {
			return nil
		}
		hsOutNs := outNs.NestedReadBucket(hs)
		if hsOutNs != nil {
			err := hsOutNs.ForEach(func(k, v []byte) error {
				to, err := wtxmgr.DecodeAddrTxOutput(v)
				if err != nil {
					return err
				}
				txOuts = append(txOuts, *to)
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
	return txOuts, nil
}

func (w *Wallet) getAddrAndAddrTxOutputByAddr(addr string) (*AddrAndAddrTxOutput, error) {
	height := w.Manager.ChainHeight()
	ato := NewAddrAndAddrTxOutput()
	for _, token := range w.tokens.tokens {
		b := Balance{}
		txOuts, err := w.getAddrTxOutputByCoin(addr, types.CoinID(token.CoinId))
		if err != nil {
			return nil, err
		}
		var spendAmount = NewAmount(0, types.CoinID(token.CoinId))
		var usableAmount = NewAmount(0, types.CoinID(token.CoinId))
		var UnconfirmedAmount = NewAmount(0, types.CoinID(token.CoinId))
		var totalAmount = NewAmount(0, types.CoinID(token.CoinId))
		var lockAmount = NewAmount(0, types.CoinID(token.CoinId))

		for _, txOut := range txOuts {
			if txOut.Status == wtxmgr.TxStatusConfirmed {
				if txOut.Spend == wtxmgr.SpendStatusSpend {
					spendAmount.Value += txOut.Amount.Value
				} else {
					if txOut.Locked > height {
						lockAmount.Value += txOut.Amount.Value
					} else {
						usableAmount.Value += txOut.Amount.Value
					}
				}
			} else if txOut.Status == wtxmgr.TxStatusUnConfirmed {
				if txOut.Spend == wtxmgr.SpendStatusSpend {
					spendAmount.Value += txOut.Amount.Value
				} else {
					UnconfirmedAmount.Value += txOut.Amount.Value
				}
			} else if txOut.Status == wtxmgr.TxStatusMemPool {
				if txOut.Spend == wtxmgr.SpendStatusSpend {
					spendAmount.Value += txOut.Amount.Value
				} else {
					UnconfirmedAmount.Value += txOut.Amount.Value
				}
			}
		}
		totalAmount.Value = usableAmount.Value + lockAmount.Value + UnconfirmedAmount.Value

		b.UnspentAmount = usableAmount
		b.UnconfirmedAmount = UnconfirmedAmount
		b.LockAmount = lockAmount
		b.SpendAmount = spendAmount
		b.TotalAmount = totalAmount
		ato.balanceMap[types.CoinID(token.CoinId)] = b
		ato.TxoutputMap[types.CoinID(token.CoinId)] = txOuts
	}
	ato.Addr = addr
	return ato, nil
}

func (w *Wallet) getAddrAndAddrTxOutputByCoin(addr string, coin int) (*AddrAndAddrTxOutput, error) {
	height := w.Manager.ChainHeight()
	ato := NewAddrAndAddrTxOutput()
	b := Balance{}
	txOuts, err := w.getAddrTxOutputByCoin(addr, types.CoinID(coin))
	if err != nil {
		return nil, err
	}
	token, err := w.tokens.GetToken(types.CoinID(coin))
	if err != nil {
		return nil, err
	}
	var spendAmount = NewAmount(0, types.CoinID(token.CoinId))
	var usableAmount = NewAmount(0, types.CoinID(token.CoinId))
	var UnconfirmedAmount = NewAmount(0, types.CoinID(token.CoinId))
	var totalAmount = NewAmount(0, types.CoinID(token.CoinId))
	var lockAmount = NewAmount(0, types.CoinID(token.CoinId))

	for _, txOut := range txOuts {
		if txOut.Status == wtxmgr.TxStatusConfirmed {
			if txOut.Spend == wtxmgr.SpendStatusSpend {
				spendAmount.Value += txOut.Amount.Value
			} else {
				if txOut.Locked > height {
					lockAmount.Value += txOut.Amount.Value
				} else {
					usableAmount.Value += txOut.Amount.Value
				}
			}
		} else if txOut.Status == wtxmgr.TxStatusUnConfirmed {
			if txOut.Spend == wtxmgr.SpendStatusSpend {
				spendAmount.Value += txOut.Amount.Value
			} else {
				UnconfirmedAmount.Value += txOut.Amount.Value
			}
		} else if txOut.Status == wtxmgr.TxStatusMemPool {
			if txOut.Spend == wtxmgr.SpendStatusSpend {
				spendAmount.Value += txOut.Amount.Value
			} else {
				UnconfirmedAmount.Value += txOut.Amount.Value
			}
		}
	}
	totalAmount.Value = usableAmount.Value + lockAmount.Value + UnconfirmedAmount.Value

	b.UnspentAmount = usableAmount
	b.UnconfirmedAmount = UnconfirmedAmount
	b.LockAmount = lockAmount
	b.SpendAmount = spendAmount
	b.TotalAmount = totalAmount
	ato.balanceMap[types.CoinID(token.CoinId)] = b
	ato.TxoutputMap[types.CoinID(token.CoinId)] = txOuts
	ato.Addr = addr
	return ato, nil
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
	//TODO
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

	for _, at := range at.TxoutputMap {
		for _, o := range at {
			txOut, found := allMap[o.TxId]
			if found {
				txOut.Variation += o.Amount.Value
			} else {
				txOut.TxID = o.TxId
				txOut.Variation = o.Amount.Value
				txOut.BlockHash = o.Block.Hash
				txOut.BlockOrder = uint32(o.Block.Order)
			}
			//log.Debug(fmt.Sprintf("%s %v %v", o.TxId.String(), float64(o.Amount)/math.Pow10(8), float64(txOut.Amount)/math.Pow10(8)))

			allMap[o.TxId] = txOut

			if o.SpendTo != nil {
				txOut, found := allMap[o.SpendTo.TxId]
				if found {
					txOut.Variation -= o.Amount.Value
				} else {
					txOut.TxID = o.SpendTo.TxId
					txOut.Variation = -o.Amount.Value
					// ToDo: add Block to SpendTo
					txOut.BlockHash = o.Block.Hash
					txOut.BlockOrder = uint32(o.Block.Order)
				}
				allMap[o.SpendTo.TxId] = txOut
				//log.Debug(fmt.Sprintf("%s %v %v", o.SpendTo.TxHash.String(), float64(-o.Amount)/math.Pow10(8), float64(txOut.Amount)/math.Pow10(8)))
			}
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

func (w *Wallet) GetBalanceByCoin(addr string, coin types.CoinID) (map[string]Value, error) {
	balanceMap := map[string]Value{}
	if addr == "" {
		return nil, errors.New("addr is nil")
	}
	res, err := w.getAddrAndAddrTxOutputByCoin(addr, int(coin))
	if err != nil {
		return nil, err
	}
	for key, val := range res.balanceMap {
		balanceMap[key.Name()] = Value{
			TotalAmount:       val.TotalAmount.Value,
			UnspentAmount:     val.UnspentAmount.Value,
			LockAmount:        val.LockAmount.Value,
			UnconfirmedAmount: val.UnconfirmedAmount.Value,
			SpendAmount:       val.SpendAmount.Value,
		}
	}
	return balanceMap, nil
}

func (w *Wallet) GetBalance(addr string) (map[types.CoinID]Balance, error) {
	balanceMap := map[types.CoinID]Balance{}
	if addr == "" {
		return nil, errors.New("addr is nil")
	}
	res, err := w.getAddrAndAddrTxOutputByAddr(addr)
	if err != nil {
		return nil, err
	}
	for key, val := range res.balanceMap {
		balanceMap[key] = val
	}
	return balanceMap, nil
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
			outNrb := rb.NestedAndCreateReadWriteBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, types.CoinID(vOut.CoinId)))
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

func (w *Wallet) insertTx(order uint32, txins []wtxmgr.TxInputPoint, txouts []wtxmgr.AddrTxOutput, status []wtxmgr.TxConfirmed, trrs []corejson.TxRawResult) error {
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		txNs := ns.NestedReadWriteBucket(wtxmgr.BucketTxJson)
		unTxNs := ns.NestedReadWriteBucket(wtxmgr.BucketUnConfirmed)
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
			outNs := ns.NestedAndCreateReadWriteBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, txo.Amount.Id))
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
			outNs := ns.NestedAndCreateReadWriteBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, types.CoinID(txr.Vout[txi.TxOutPoint.OutIndex].CoinId)))
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
		for _, s := range status {
			k, err := hash.NewHashFromStr(s.TxId)
			if err != nil {
				return err
			}
			if s.TxStatus == wtxmgr.TxStatusUnConfirmed {
				value := &wtxmgr.UnconfirmTx{
					Order:         order,
					Confirmations: s.Confirmations,
				}
				unTxNs.Put(k.Bytes(), value.Marshal())
			} else {
				unTxNs.Delete(k.Bytes())
			}
		}
		return nil
	})
	return err
}

func (w *Wallet) parseTx(tx *j.DecodeRawTransactionResult) ([]wtxmgr.TxInputPoint, []wtxmgr.AddrTxOutput, []wtxmgr.TxConfirmed, []corejson.TxRawResult, error) {
	var txIns []wtxmgr.TxInputPoint
	var txOuts []wtxmgr.AddrTxOutput
	var status []wtxmgr.TxConfirmed
	var txRaws []corejson.TxRawResult
	var confirmations uint32
	confirmations = config.Cfg.Confirmations
	tr := corejson.TxRawResult{
		Txid:          tx.Txid,
		TxHash:        tx.Hash,
		Version:       tx.Version,
		LockTime:      tx.LockTime,
		Timestamp:     tx.Time,
		Vin:           tx.Vin,
		Vout:          tx.Vout,
		BlockHash:     tx.BlockHash,
		BlockOrder:    tx.Order,
		Confirmations: int64(tx.Confirms),
		Duplicate:     tx.Duplicate,
		Txsvalid:      tx.Txvalid,
	}
	txRaws = append(txRaws, tr)
	tin, tout, txStatus, isCoinBase, err := w.parseTxDetail(tr, uint32(tr.BlockOrder), tx.IsBlue)
	if err != nil {
		return nil, nil, nil, nil, err
	} else {
		if isCoinBase {
			confirmations = uint32(w.chainParams.CoinbaseMaturity)
		}
		status = append(status, wtxmgr.TxConfirmed{
			TxId:          tr.Txid,
			Confirmations: confirmations,
			TxStatus:      txStatus,
		})
		txIns = append(txIns, tin...)
		txOuts = append(txOuts, tout...)
	}
	return txIns, txOuts, status, txRaws, nil
}

func (w *Wallet) parseTxDetail(tr corejson.TxRawResult, order uint32, isBlue bool) ([]wtxmgr.TxInputPoint, []wtxmgr.AddrTxOutput, wtxmgr.TxStatus, bool, error) {
	var txins []wtxmgr.TxInputPoint
	var txouts []wtxmgr.AddrTxOutput
	var isCoinBase bool
	var inMemPool bool
	if tr.BlockHash == "" {
		inMemPool = true
	}
	blockhash, err := hash.NewHashFromStr(tr.BlockHash)
	if err != nil {
		return nil, nil, wtxmgr.TxStatusUnConfirmed, isCoinBase, err
	}
	block := wtxmgr.Block{
		Hash:  *blockhash,
		Order: int32(order),
	}
	txId, err := hash.NewHashFromStr(tr.Txid)
	if err != nil {
		return nil, nil, wtxmgr.TxStatusUnConfirmed, isCoinBase, err
	}
	for i, vi := range tr.Vin {
		if vi.Coinbase != "" {
			isCoinBase = true
			continue
		}
		if vi.Txid == "" && vi.Vout == 0 {
			continue
		} else {
			hs, err := hash.NewHashFromStr(vi.Txid)
			if err != nil {
				return nil, nil, wtxmgr.TxStatusUnConfirmed, isCoinBase, err
			} else {
				txOutPoint := types.TxOutPoint{
					Hash:     *hs,
					OutIndex: vi.Vout,
				}
				spendTo := wtxmgr.SpendTo{
					Index: uint32(i),
					TxId:  *txId,
				}
				txIn := wtxmgr.TxInputPoint{
					TxOutPoint: txOutPoint,
					SpendTo:    spendTo,
				}
				txins = append(txins, txIn)
			}
		}
	}
	txStatus := w.txStatus(uint32(tr.Confirmations), tr.Txsvalid, isBlue, isCoinBase, inMemPool)
	for index, vo := range tr.Vout {
		var lock uint64
		switch vo.ScriptPubKey.Type {
		case "cltvpubkeyhash":
			if len(vo.ScriptPubKey.Addresses) == 0 {
				continue
			} else {
				codes := strings.Split(vo.ScriptPubKey.Asm, " ")
				if len(codes) == 0 {
					return nil, nil, 0, false, fmt.Errorf("cltvpubkeyhash vout error,  %s", vo.ScriptPubKey.Asm)
				}
				lock, err = littleHexToUint64(codes[0])
				if err != nil {
					return nil, nil, 0, false, fmt.Errorf("little hex %s to uint64 error, %s", codes[0], err.Error())
				}
			}
		case "nonstandard":
			continue
		}
		if len(vo.ScriptPubKey.Addresses) == 0 {
			fmt.Printf("")
		}
		txOut := wtxmgr.AddrTxOutput{
			Address: vo.ScriptPubKey.Addresses[0],
			TxId:    *txId,
			Index:   uint32(index),
			Amount:  types.Amount{Value: int64(vo.Amount), Id: types.CoinID(vo.CoinId)},
			Block:   block,
			Spend:   wtxmgr.SpendStatusUnspent,
			IsBlue:  isBlue,
			Status:  txStatus,
			SpendTo: &wtxmgr.SpendTo{
				Index: 0,
				TxId:  hash.Hash{},
			},
			Locked:   uint32(lock),
			PkScript: vo.ScriptPubKey.Hex,
		}
		txouts = append(txouts, txOut)
	}

	return txins, txouts, txStatus, isCoinBase, nil
}

func (w *Wallet) txStatus(confirmations uint32, txsvalid, isBlue, isCoinBase, inMemPool bool) wtxmgr.TxStatus {
	if isCoinBase {
		if confirmations < uint32(w.chainParams.CoinbaseMaturity) {
			return wtxmgr.TxStatusUnConfirmed
		} else if !txsvalid {
			return wtxmgr.TxStatusFailed
		} else if isBlue {
			return wtxmgr.TxStatusConfirmed
		} else {
			return wtxmgr.TxStatusRead
		}
	} else {
		if confirmations < config.Cfg.Confirmations {
			if inMemPool {
				return wtxmgr.TxStatusMemPool
			}
			return wtxmgr.TxStatusUnConfirmed
		} else if !txsvalid {
			return wtxmgr.TxStatusFailed
		} else {
			return wtxmgr.TxStatusConfirmed
		}
	}
}

func (w *Wallet) GetSyncBlockOrder() uint32 {
	order := w.Manager.SyncedTo().Order
	return order
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
		stamp := &waddrmgr.BlockStamp{Hash: *hs, Order: block.Order}
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

func (w *Wallet) updateTokens() error {
	tokens, err := w.HttpClient.GetTokenInfo()
	if err != nil {
		return err
	}
	for _, token := range tokens {
		w.tokens.Add(token)
		walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
			bucket := tx.ReadWriteBucket(tokenmgrNamespaceKey)

			types.CoinNameMap[types.CoinID(token.CoinId)] = types.CoinID(token.CoinId).Name()
			bucket.Put([]byte(types.CoinID(token.CoinId).Name()), EncodeToken(token))
			return nil
		})
	}
	return nil
}

func (w *Wallet) ClearTxData() error {
	return walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		if err := tx.DeleteTopLevelBucket(wtxmgrNamespaceKey); err != nil {
			return nil
		}
		ns, err := tx.CreateTopLevelBucket(wtxmgrNamespaceKey)
		if err != nil {
			return err
		}
		if err := wtxmgr.Create(ns); err != nil {
			return err
		}
		addrMgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		h, _ := hash.NewHashFromStr("")
		stamp := &waddrmgr.BlockStamp{Hash: *h, Order: 0}
		err = w.Manager.SetSyncedTo(addrMgrNs, stamp)
		if err != nil {
			return err
		}
		return nil
	})
}

func (w *Wallet) UpdateBlock(toOrder uint64) error {
	var err error
	w.syncLatest = false
	w.syncAll = true
	w.syncQuit = make(chan struct{}, 1)
	if toOrder != 0 {
		w.syncAll = false
	}
	_ = w.updateTokens()
	addrs, err := w.walletAddress()
	if err != nil {
		return err
	}
	err = w.updateSyncToOrder(uint32(toOrder))
	if err != nil {
		return err
	}
	w.setOrder(w.Manager.SyncedTo().Order)
	w.scanEnd <- struct{}{}
	ntfnHandlers := client.NotificationHandlers{
		OnBlockConnected:    w.OnBlockConnected,
		OnTxConfirm:         w.OnTxConfirm,
		OnTxAcceptedVerbose: w.OnTxAcceptedVerbose,
		OnRescanProgress:    w.OnRescanProgress,
		OnRescanFinish:      w.OnRescanFinish,
		OnNodeExit:          w.OnNodeExit,
	}

	w.notificationRpc, err = NewNotificationRpc(w.cfg, ntfnHandlers)
	if err != nil {
		return err
	}
	if err = w.notifyBlock(); err != nil {
		return err
	}

	if err = w.notifyTxByAddr(addrs); err != nil {
		return err
	}

	if err := w.notifyNewTransaction(); err != nil {
		return err
	}

	w.syncWg.Add(1)
	go w.notifyTxConfirmed()

	w.syncWg.Add(1)
	go w.notifyScanTxByAddr(addrs)

	w.notificationRpc.WaitForShutdown()
	w.syncWg.Wait()
	log.Info("Stop notify sync process")
	return nil
}

func (w *Wallet) notifyScanTxByAddr(addrs []string) {
	defer w.syncWg.Done()
	var startScan bool

	for {
		select {
		case <-w.syncQuit:
			log.Info("Stop scan block")
			return
		case <-w.scanEnd:
			if !w.syncAll && (startScan || w.getToOrder() <= w.getSyncOrder()+1) {
				fmt.Fprintf(os.Stdout, "update history block:%d/%d\n", w.getSyncOrder(), w.getToOrder()-1)
				w.notificationRpc.Shutdown()
				return
			} else {
				startScan = true
				if err := w.updateSyncToOrder(0); err != nil {
					w.stopSync()
					break
				}
				if w.getToOrder() > w.getSyncOrder()+1 {
					w.syncLatest = false
					log.Info("notification rescan block", "start", w.getSyncOrder(), "end", w.getToOrder()-1)
					err := w.notificationRpc.Rescan(uint64(w.getSyncOrder()), uint64(w.getToOrder()), addrs, nil)
					if err != nil {
						return
					}
				} else {
					w.syncLatest = true
					fmt.Fprintf(os.Stdout, "update history block:%d/%d\r", w.getSyncOrder(), w.getToOrder()-1)
					return
				}
			}
		default:
			time.Sleep(time.Second * 5)
		}
	}
}

func (w *Wallet) OnTxConfirm(txConfirm *cmds.TxConfirmResult) {
	if err := w.updateTxConfirm(txConfirm); err != nil {
		log.Warn("updateTxConfirm", "error", err)
	}
	fmt.Printf("Confirm %d %d %s\n", txConfirm.Order, txConfirm.Confirms, txConfirm.Tx)
}

func (w *Wallet) OnTxAcceptedVerbose(c *client.Client, tx *j.DecodeRawTransactionResult) {
	if tx.Duplicate {
		return
	}
	if !w.syncLatest {
		if tx.BlockHash == "" {
			return
		}
		if tx.Order >= uint64(w.getToOrder()) {
			return
		}
	}
	txIns, txOuts, status, trRs, err := w.parseTx(tx)
	if err != nil {
		log.Error("OnTxAcceptedVerbose parse tx", "error", err)
		return
	}
	err = w.insertTx(uint32(tx.Order), txIns, txOuts, status, trRs)
	if err != nil {
		log.Error("OnTxAcceptedVerbose insert tx", "error", err)
		return
	}

	blockHash, _ := hash.NewHashFromStr(tx.BlockHash)
	err = w.updateBlockTemp(*blockHash, uint32(tx.Order))
	if err != nil {
		return
	}
	if w.syncLatest {
		_, _ = fmt.Fprintf(os.Stdout, "update new transaction:%d %s\r", tx.Order, tx.Txid)
	}
}

func (w *Wallet) OnRescanProgress(rescanPro *cmds.RescanProgressNtfn) {
	//log.Info("scan block progress", "order", rescanPro.Order)
	_, _ = fmt.Fprintf(os.Stdout, "update history blcok:%d/%d\r", rescanPro.Order, w.getToOrder()-1)
}

func (w *Wallet) updateSyncToOrder(toOrder uint32) error {
	maxOrder, err := w.maxBlockOrder()
	if err != nil {
		return err
	}
	if toOrder == 0 {
		toOrder = uint32(maxOrder)
	} else if toOrder > uint32(maxOrder) {
		return fmt.Errorf("the target Order %d cannot be larger than the number of existing blocks  %d on the node", toOrder, maxOrder)
	}
	w.setToOrder(toOrder)
	return nil
}

func (w *Wallet) OnBlockConnected(hash *hash.Hash, height int64, order int64, t time.Time, txs []*types.Transaction) {
	if err := w.updateChainHeight(uint32(height)); err != nil {
		log.Warn("update chain height", "error", err.Error())
	}
}

func (w *Wallet) OnRescanFinish(rescanFinish *cmds.RescanFinishedNtfn) {
	defer func() {
		w.scanEnd <- struct{}{}
	}()

	hash, err := w.HttpClient.getBlockHashByOrder(int64(w.getToOrder() - 1))
	if err != nil {
		log.Warn("get block hash by order", "error", err)
		return
	}
	err = w.updateBlockTemp(*hash, w.getToOrder()-1)
	if err != nil {
		return
	}
	w.setOrder(w.getToOrder() - 1)
}

func (w *Wallet) OnNodeExit(nodeExit *cmds.NodeExitNtfn) {
	w.notificationRpc.Shutdown()
	w.stopSync()
}

func (w *Wallet) stopSync() {
	close(w.syncQuit)
}

func (w *Wallet) updateBlockTemp(hash hash.Hash, localOrder uint32) error {
	stamp := &waddrmgr.BlockStamp{Hash: hash, Order: localOrder}
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		err := w.Manager.SetSyncedTo(ns, stamp)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (w *Wallet) updateChainHeight(height uint32) error {
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		err := w.Manager.SetChainHeight(ns, height)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (w *Wallet) notifyTxByAddr(addrs []string) error {
	err := w.notificationRpc.NotifyTxsByAddr(false, addrs, nil)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wallet) notifyBlock() error {
	err := w.notificationRpc.NotifyBlocks()
	if err != nil {
		return err
	}
	return nil
}

func (w *Wallet) notifyNewTransaction() error {
	err := w.notificationRpc.NotifyNewTransactions(true)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wallet) notifyTxConfirmed() {
	defer w.syncWg.Done()

	t := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-w.syncQuit:
			log.Info("Stop notify tx confirmed block")
			return
		case <-t.C:
			unTxs := []cmds.TxConfirm{}
			err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
				ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
				unTxNs := ns.NestedReadWriteBucket(wtxmgr.BucketUnConfirmed)
				err := unTxNs.ForEach(func(k, v []byte) error {
					hashTxId, err := hash.NewHash(k)
					if err != nil {
						return err
					}
					u, err := wtxmgr.UnMarshalUnconfirmTx(v)
					if err != nil {
						return err
					}
					unTxs = append(unTxs, cmds.TxConfirm{
						Txid:          hashTxId.String(),
						Order:         uint64(u.Order),
						Confirmations: int32(u.Confirmations),
					})
					return nil
				})
				return err
			})
			if err != nil {
				log.Error(err.Error())
				continue
			}

			if len(unTxs) > 0 {
				err := w.notificationRpc.NotifyTxsConfirmed(unTxs)
				if err != nil {
					log.Error(err.Error())
					continue
				}
			}
		}
	}
}

func (w *Wallet) setOrder(syncOrder uint32) {
	w.orderMutex.Lock()
	defer w.orderMutex.Unlock()

	w.syncOrder = syncOrder
}

func (w *Wallet) setToOrder(toOrder uint32) {
	w.orderMutex.Lock()
	defer w.orderMutex.Unlock()

	w.toOrder = toOrder
}

func (w *Wallet) getSyncOrder() uint32 {
	w.orderMutex.RLock()
	defer w.orderMutex.RUnlock()

	return w.syncOrder
}

func (w *Wallet) getToOrder() uint32 {
	w.orderMutex.RLock()
	defer w.orderMutex.RUnlock()

	return w.toOrder
}

func (w *Wallet) walletAddress() ([]string, error) {
	manager, err := w.Manager.FetchScopedKeyManager(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}
	var results []AccountAndAddressResult
	var addresses = []string{}
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
			adds, _ := w.AccountAddresses(results[k].AccountNumber)
			for _, addr := range adds {
				addresses = append(addresses, addr.String())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return addresses, nil
}

func (w *Wallet) updateTxConfirm(confirmRs *cmds.TxConfirmResult) error {
	tx, err := w.GetTx(confirmRs.Tx)
	if err != nil {
		log.Error("wallet can not find tx", "txid", confirmRs.Tx)
		return err
	}
	status := w.txStatus(uint32(confirmRs.Confirms), confirmRs.IsValid, confirmRs.IsBlue, wtxmgr.TxRawIsCoinBase(tx), false)
	if status < wtxmgr.TxStatusConfirmed {
		log.Warn("updateTxConfirm tx status is unconfirmed", "TxConfirmResult", confirmRs)
		return nil
	}
	return w.updateTxStatus(tx, status)
}

func (w *Wallet) updateTxStatus(txRaw corejson.TxRawResult, status wtxmgr.TxStatus) error {
	coinBucket := map[types.CoinID]walletdb.ReadWriteBucket{}
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		var bucket walletdb.ReadWriteBucket
		var ok bool
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		unTxNs := ns.NestedReadWriteBucket(wtxmgr.BucketUnConfirmed)
		txHash, err := hash.NewHashFromStr(txRaw.Txid)
		if err != nil {
			return err
		}
		for i, vout := range txRaw.Vout {
			if vout.ScriptPubKey.Addresses == nil {
				continue
			}
			coinID := types.CoinID(vout.CoinId)
			if bucket, ok = coinBucket[coinID]; ok {
				bucket = coinBucket[coinID]
			} else {
				bucket = ns.NestedAndCreateReadWriteBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, coinID))
				coinBucket[coinID] = bucket
			}
			out, err := w.TxStore.GetAddrTxOut(bucket, vout.ScriptPubKey.Addresses[0], types.TxOutPoint{
				Hash:     *txHash,
				OutIndex: uint32(i),
			})
			if err != nil {
				return err
			}
			out.Status = status
			err = w.TxStore.UpdateAddrTxOut(bucket, out)
			if err != nil {
				return err
			}
		}
		if err := unTxNs.Delete(txHash.Bytes()); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (w *Wallet) maxBlockOrder() (uint64, error) {
	var blockCount string
	var err error
	blockCount, err = w.HttpClient.getblockCount()
	if err != nil {
		return 0, err
	}
	Order, err := strconv.ParseUint(blockCount, strIntBase, strIntBitSize32)
	if err != nil {
		return 0, err
	}
	return Order, nil
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
	for _, token := range w.tokens.tokens {
		id := types.CoinID(token.CoinId)
		for index, aaa := range aaaRs {
			results[index].AccountNumber = aaa.AccountNumber
			results[index].AccountName = aaa.AccountName
			balance := NewBalance(id)
			for _, addr := range aaa.AddrsOutput {
				balance.UnspentAmount.Value += addr.balanceMap[id].UnspentAmount.Value
				balance.LockAmount.Value += addr.balanceMap[id].LockAmount.Value
				balance.SpendAmount.Value += addr.balanceMap[id].SpendAmount.Value
				balance.TotalAmount.Value += addr.balanceMap[id].TotalAmount.Value
				balance.UnconfirmedAmount.Value += addr.balanceMap[id].UnconfirmedAmount.Value

				balance.LockAmount.Id = addr.balanceMap[id].LockAmount.Id
				balance.UnspentAmount.Id = addr.balanceMap[id].UnspentAmount.Id
				balance.UnconfirmedAmount.Id = addr.balanceMap[id].UnconfirmedAmount.Id
				balance.TotalAmount.Id = addr.balanceMap[id].TotalAmount.Id
				balance.SpendAmount.Id = addr.balanceMap[id].SpendAmount.Id

			}
			results[index].AccountBalanceList = append(results[index].AccountBalanceList, *balance)
		}
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
	return wif, nil
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
			// Get private key from wallet if it exists.
			pka, ok := mAddr.(waddrmgr.ManagedPubKeyAddress)
			if !ok {
				return fmt.Errorf("address %s is not a key type", mAddr.Address())
			}
			pkaddr, err := address.NewSecpPubKeyAddress(pka.PubKey().SerializeCompressed(), w.chainParams)
			if err != nil {
				log.Error("PubKey Create Failed", mAddr.Address())
				return err
			}
			addrs = append(addrs, pkaddr)
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

func (w *Wallet) GetUnspentUTXO(addr string, coin types.CoinID) ([]wtxmgr.UTxo, error) {
	var utxos []wtxmgr.UTxo
	outputs, err := w.GetUnspentAddrOutput(addr, coin)
	if err != nil {
		return nil, err
	}
	for _, outPut := range outputs {
		utxos = append(utxos, wtxmgr.UTxo{
			TxId:    outPut.TxId.String(),
			Index:   outPut.Index,
			Amount:  outPut.Amount,
			Address: outPut.Address,
		})
	}
	return utxos, nil
}

func (w *Wallet) GetUnspentAddrOutput(addr string, coin types.CoinID) ([]*wtxmgr.AddrTxOutput, error) {
	height := w.Manager.ChainHeight()
	var utxos []*wtxmgr.AddrTxOutput
	err := walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		hs := []byte(addr)
		ns := tx.ReadBucket(wtxmgrNamespaceKey)
		outns := ns.NestedReadBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, types.CoinID(coin)))
		if outns == nil {
			return nil
		}
		hsoutns := outns.NestedReadBucket(hs)
		if hsoutns != nil {
			_ = hsoutns.ForEach(func(k, v []byte) error {
				outPut := wtxmgr.NewAddrTxOutput()
				outPut, err := wtxmgr.DecodeAddrTxOutput(v)
				if err != nil {
					log.Error("readAddrTxOutput err", "err", err.Error())
					return err
				}

				if outPut.Spend == wtxmgr.SpendStatusUnspent && outPut.Status == wtxmgr.TxStatusConfirmed && outPut.Locked <= height {
					utxos = append(utxos, outPut)
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		log.Error("ReadAddrTxOutput err", "err", err)
		return nil, err
	}
	return utxos, nil
}

// Sendoutputs can only be accessed by a single thread at the same time to prevent the referenced utxo from being referenced again under the concurrency
var syncSendOutputs = new(sync.Mutex)

// SendOutputs creates and sends payment transactions. It returns the
// transaction upon success.
func (w *Wallet) SendOutputs(coin2outputs []*TxOutput, coinId types.CoinID, account int64, satPerKb int64, byAddr string) (*string, error) {
	// Ensure the outputs to be created adhere to the network's consensus
	// rules.
	syncSendOutputs.Lock()
	defer syncSendOutputs.Unlock()
	var addrs = make([]types.Address, 0)
	var err error
	if account == waddrmgr.AccountMergePayNum {
		addrs, err = w.GetAccountAddress(waddrmgr.KeyScopeBIP0044)
	} else {
		addrs, err = w.AccountAddresses(uint32(account))
	}
	if byAddr != "" {
		addr, err := address.DecodeAddress(byAddr)
		if err != nil {
			return nil, err
		}
		addrs = []types.Address{addr}
	}
	if err != nil {
		return nil, err
	}
	log.Info("SendOutputs", "addrs", addrs)
	signedRaw, payAmount, allSpentUTXO, err := w.createTx(addrs, coin2outputs, coinId, 0, satPerKb)
	if err != nil {
		return nil, err
	}
	fees := w.fees(signedRaw, coinId)
	payAmount = payAmount + fees
	signedRaw, payAmount, allSpentUTXO, err = w.createTx(addrs, coin2outputs, coinId, fees, satPerKb)
	if err != nil {
		return nil, err
	}

	log.Trace(fmt.Sprintf("signTx size:%v", len(signedRaw)), "signTx", signedRaw)
	msg, err := w.HttpClient.SendRawTransaction(signedRaw, false)
	if err != nil {
		log.Trace("SendRawTransaction txSign err ", "err", err.Error())
		return nil, err
	} else {
		msg = strings.ReplaceAll(msg, "\"", "")
		log.Trace("SendRawTransaction txSign response msg", "msg", msg)
	}

	txId, _ := hash.NewHashFromStr(msg)
	w.updateUTXOSpent(allSpentUTXO, &wtxmgr.SpendTo{
		TxId: *txId,
	})
	return &msg, nil
}

func (w *Wallet) createTx(addrs []types.Address, coin2outputs []*TxOutput, coinId types.CoinID, fees int64, satPerKb int64) (string, int64, []*wtxmgr.AddrTxOutput, error) {
	var sum int64
	outputs := make([]qx.Output, 0)
	inputs := make([]qx.Input, 0)
	priKeyList := make([]string, 0)
	payAmount := types.Amount{Id: coinId}
	for _, output := range coin2outputs {
		if err := txrules.CheckOutput(types.NewTxOutput(output.Amount, output.PkScript), satPerKb); err != nil {
			return "", 0, nil, err
		}
		payAmount.Value += output.Amount.Value
		outputs = append(outputs, qx.Output{
			TargetLockTime: int64(output.LockHeight),
			TargetAddress:  output.Address,
			Amount: types.Amount{
				Value: payAmount.Value,
				Id:    payAmount.Id,
			},
			OutputType: types.TxTypeRegular,
		})
	}

	payAmount.Value = payAmount.Value + fees
	uxtoList, sum, err := w.GetUTXOByAddress(addrs, payAmount)
	if err != nil {
		return "", 0, nil, err
	}
	change := sum - payAmount.Value
	if change > 0 {
		addr, _ := address.DecodeAddress(uxtoList[0].Address)
		addrScript, _ := txscript.PayToAddrScript(addr)
		changeOut := types.NewTxOutput(types.Amount{
			Value: change,
			Id:    coinId,
		}, addrScript)
		if err := txrules.CheckOutput(changeOut, satPerKb); err == nil {
			outputs = append(outputs, qx.Output{
				TargetLockTime: 0,
				Amount: types.Amount{
					Value: change,
					Id:    coinId,
				},
				TargetAddress: uxtoList[0].Address,
				OutputType:    types.TxTypeRegular,
			})
		}
	}

	outputVal := uint64(0)
	for _, v := range outputs {
		outputVal += uint64(v.Amount.Value)
	}
	log.Info("output all val is: ", "val", outputVal)
	var vinPkScript = make([]string, 0)
	for _, utxo := range uxtoList {
		addr, _ := address.DecodeAddress(utxo.Address)

		vinPkScript = append(vinPkScript, utxo.PkScript)
		typ := types.TxTypeRegular
		pkhAddr := addr
		switch addr.(type) {
		case *address.SecpPubKeyAddress:
			typ = types.TxTypeCrossChainExport
			pkaddr := addr.(*address.SecpPubKeyAddress)
			pkhAddr = pkaddr.PKHAddress()
		default:
		}
		if fees > 0 {
			log.Info("uxtoList", "utxo detail", *utxo, "pkhAddr", pkhAddr.String())
		}
		inputs = append(inputs, qx.Input{
			TxID:      utxo.TxId.String(),
			InputType: typ,
			OutIndex:  utxo.Index})
		pri, err := w.getPrivateKey(pkhAddr)
		if err != nil {
			return "", 0, nil, err
		}
		priKey, err := pri.PrivKey()
		if err != nil {
			return "", 0, nil, err
		}
		priKeyList = append(priKeyList, hex.EncodeToString(priKey.Serialize()))
	}
	timeNow := time.Now()
	raw, err := qx.TxEncode(1, w.Manager.ChainHeight(), &timeNow, inputs, outputs)
	if err != nil {
		return "", 0, nil, err
	}
	signedRaw, err := qx.TxSign(priKeyList, raw, config.Cfg.Network)
	if err != nil {
		return "", 0, nil, err
	}
	log.Info("signedRaw", "str", signedRaw)
	return signedRaw, payAmount.Value, uxtoList, nil
}

func (w *Wallet) updateUTXOSpent(UTXOs []*wtxmgr.AddrTxOutput, spentTx *wtxmgr.SpendTo) error {
	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		ns := tx.ReadWriteBucket(wtxmgrNamespaceKey)
		for _, txoutput := range UTXOs {
			outns := ns.NestedAndCreateReadWriteBucket(wtxmgr.CoinBucket(wtxmgr.BucketAddrtxout, txoutput.Amount.Id))
			txoutput.Spend = wtxmgr.SpendStatusSpend
			txoutput.SpendTo = spentTx
			err := w.TxStore.UpdateAddrTxOut(outns, txoutput)
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
		return err
	}
	return err
}

func (w *Wallet) GetUTXOByAddress(addrs []types.Address, amount types.Amount) ([]*wtxmgr.AddrTxOutput, int64, error) {
	otxoList := make([]*wtxmgr.AddrTxOutput, 0)
	var sum int64
FindUTXO:
	for _, addr := range addrs {
		uxtoList, err := w.GetUnspentAddrOutput(addr.String(), amount.Id)
		if err != nil {
			log.Warn("Failed to get address utxo", "address", addr.String(), "coinId", amount.Id.Name())
			continue
		}
		for _, utxo := range uxtoList {
			sum += utxo.Amount.Value
			otxoList = append(otxoList, utxo)
			if sum >= amount.Value {
				break FindUTXO
			}
		}
	}

	if sum < amount.Value {
		return nil, 0, fmt.Errorf("the balance is not enough to send %v", amount)
	}
	return otxoList, sum, nil
}

func (w *Wallet) fees(rawTx string, coinId types.CoinID) int64 {
	bytes, _ := hex.DecodeString(rawTx)
	txLen := len(bytes)
	return util.CalcMinRequiredTxRelayFee(int64(txLen), config.Cfg.MinTxFee)
}

// Multi address merge signature
func (w *Wallet) multiAddressMergeSign(redeemTx types.Transaction, txInPkScript [][]byte) (string, error) {
	for i, txIn := range redeemTx.TxIn {
		addrByte := txIn.SignScript
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
		sigScript, err := txscript.SignatureScript(&redeemTx, i, txInPkScript[i], txscript.SigHashAll, priKey, true)
		if err != nil {
			return "", err
		}
		redeemTx.TxIn[i].SignScript = sigScript
	}

	mtxHex, err := marshal.MessageToHex(&redeemTx)
	if err != nil {
		return "", err
	}
	return mtxHex, nil
}

//sendPairs creates and sends payment transactions.
//It returns the transaction hash in string format upon success
//All errors are returned in btcjson.RPCError format
func (w *Wallet) SendPairs(amounts map[string]types.Amount,
	account int64, feeSatPerKb int64, lockHeight uint64, byAddress string) (string, error) {
	//check, err := w.HttpClient.CheckSyncUpdate(int64(w.Manager.SyncedTo().Order))
	log.Info("SendPairs", "amounts", amounts, "byAddress", byAddress)
	/*if check == false {
		return "", err
	}*/
	outputs, coinId, err := makeOutputs(amounts, lockHeight)
	if err != nil {
		return "", err
	}
	tx, err := w.SendOutputs(outputs, coinId, account, feeSatPerKb, byAddress)
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

func (w *Wallet) CoinID(coin types.CoinID) (types.CoinID, error) {
	token, err := w.tokens.GetToken(coin)
	if err != nil {
		return 0, err
	}
	return types.CoinID(token.CoinId), nil
}

type TxOutput struct {
	Amount     types.Amount
	Address    string
	PkScript   []byte
	LockHeight uint64
}

// makeOutputs creates a slice of transaction outputs from a pair of address
// strings to amounts.  This is used to create the outputs to include in newly
// created transactions from a JSON object describing the output destinations
// and amounts.
func makeOutputs(pairs map[string]types.Amount, lockHeight uint64) ([]*TxOutput, types.CoinID, error) {
	coin2outputs := make([]*TxOutput, 0)
	var coinId types.CoinID
	var pkScript []byte
	for addrStr, amt := range pairs {
		coinId = amt.Id
		addr, err := address.DecodeAddress(addrStr)
		if err != nil {
			return nil, coinId, fmt.Errorf("cannot decode address: %s,address:%s", err, addrStr)
		}
		if lockHeight != 0 {
			pkScript, err = txscript.PayToCLTVPubKeyHashScript(addr.Script(), int64(lockHeight))
			if err != nil {
				return coin2outputs, coinId, fmt.Errorf("cannot create txout script: %s", err)
			}
		} else {
			pkScript, err = txscript.PayToAddrScript(addr)
			if err != nil {
				return coin2outputs, coinId, fmt.Errorf("cannot create txout script: %s", err)
			}
		}
		coin2outputs = append(coin2outputs, &TxOutput{
			Amount:     amt,
			Address:    addrStr,
			PkScript:   pkScript,
			LockHeight: lockHeight,
		})
	}
	return coin2outputs, coinId, nil
}

func littleHexToUint64(hexStr string) (uint64, error) {
	if len(hexStr) == 1 {
		hexStr = "0" + hexStr
	}
	src, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, err
	}
	dst := make([]byte, 8)
	copy(dst, src)
	var number uint64
	bytesBuffer := bytes.NewBuffer(dst)
	err = binary.Read(bytesBuffer, binary.LittleEndian, &number)
	return number, err
}
