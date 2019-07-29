package wallet

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	"github.com/HalalChain/qitmeer-wallet/util"

	//"github.com/HalalChain/qitmeer-wallet/wallet/txrules"
	"sync"

	log "github.com/sirupsen/logrus"


	
	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/rpc/server"


	"encoding/json"
	"errors"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	corejson "github.com/HalalChain/qitmeer-lib/core/json"
	"github.com/HalalChain/qitmeer-lib/core/types"
	chaincfg "github.com/HalalChain/qitmeer-lib/params"
	clijson "github.com/HalalChain/qitmeer-wallet/json"
	"github.com/HalalChain/qitmeer-wallet/waddrmgs"
	"github.com/HalalChain/qitmeer-wallet/walletdb"
	"github.com/HalalChain/qitmeer-wallet/wtxmgr"
	"strconv"
	"time"
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

	walletDbWatchingOnlyName = "wowallet.db"

	// recoveryBatchSize is the default number of blocks that will be
	// scanned successively by the recovery manager, in the event that the
	// wallet is started in recovery mode.
	recoveryBatchSize = 2000
)
// Wallet qitmeer-wallet
type Wallet struct {
	cfg *config.Config

	RPCSvr *server.RpcServer

	// Data stores
	db      walletdb.DB
	Manager *waddrmgr.Manager
	TxStore *wtxmgr.Store

	chainParams *chaincfg.Params

	Httpclient  *htpc

	// Channels for the manager locker.
	unlockRequests     chan unlockRequest
	lockRequests       chan struct{}
	lockState          chan bool

	wg          sync.WaitGroup

	started bool
	quit    chan struct{}
	quitMu  sync.Mutex
}

// NewWallet make wallet
func NewWallet(cfg *config.Config)(wt *Wallet,err error){


	return
}

// Start wallet routine
func (wt *Wallet) Start(){
	log.Trace("wallet start")
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

type balance struct {
	TotalAmount float64 // 总数
	SpendAmount float64 // 已花费
	UnspendAmount float64 //未花费
}

// AccountBalanceResult is a single result for the Wallet.AccountBalances method.
type AccountBalanceResult struct {
	AccountNumber  uint32
	AccountName    string
	AccountBalance types.Amount
}
// Namespace bucket keys.
var (
	waddrmgrNamespaceKey = []byte("waddrmgr")
	wtxmgrNamespaceKey   = []byte("wtxmgr")
	// 地址对应的交易in out 桶
	//waddrtrNamespaceKey   = []byte("waddrtr")
)




// ImportPrivateKey imports a private key to the wallet and writes the new
// wallet to disk.
//
// NOTE: If a block stamp is not provided, then the wallet's birthday will be
// set to the genesis block of the corresponding chain.
func (w *Wallet) ImportPrivateKey(scope waddrmgr.KeyScope, wif *util.WIF,
	bs *waddrmgr.BlockStamp, rescan bool) (string, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return "", err
	}

	// Attempt to import private key into wallet.
	var addr types.Address
	var props *waddrmgr.AccountProperties
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		maddr, err := manager.ImportPrivateKey(addrmgrNs, wif, bs)
		if err != nil {
			return err
		}
		addr = maddr.Address()
		props, err = manager.AccountProperties(
			addrmgrNs, waddrmgr.ImportedAddrAccount,
		)
		if err != nil {
			return err
		}

		// We'll only update our birthday with the new one if it is
		// before our current one. Otherwise, if we do, we can
		// potentially miss detecting relevant chain events that
		// occurred between them while rescanning.
		//birthdayBlock, _, err := w.Manager.BirthdayBlock(addrmgrNs)
		//if err != nil {
		//	return err
		//}
		//if bs.Height >= birthdayBlock.Height {
		//	return nil
		//}
		//
		//err = w.Manager.SetBirthday(addrmgrNs, bs.Timestamp)
		//if err != nil {
		//	return err
		//}
		//
		//// To ensure this birthday block is correct, we'll mark it as
		//// unverified to prompt a sanity check at the next restart to
		//// ensure it is correct as it was provided by the caller.
		//return w.Manager.SetBirthdayBlock(addrmgrNs, *bs, false)
		return nil
	})
	if err != nil {
		return "", err
	}
	fmt.Println("props:",props)
	addrStr := addr.Encode()
	log.Infof("Imported payment address %s", addrStr)

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
	params *chaincfg.Params, recoveryWindow uint32) (*Wallet, error) {

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
		txMgr, err = wtxmgr.Open(txMgrBucket, params)
		if err != nil {
			return err
		}
		fmt.Println("txmgr：",txMgr)
		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Info("Opened wallet") // TODO: log balance? last sync height?

	w := &Wallet{
		//publicPassphrase:    pubPass,
		db:                  db,
		Manager:             addrMgr,
		//TxStore:             txMgr,
		//lockedOutpoints:     map[wire.OutPoint]struct{}{},
		//recoveryWindow:      recoveryWindow,
		//rescanAddJob:        make(chan *RescanJob),
		//rescanBatch:         make(chan *rescanBatch),
		//rescanNotifications: make(chan interface{}),
		//rescanProgress:      make(chan *RescanProgressMsg),
		//rescanFinished:      make(chan *RescanFinishedMsg),
		//createTxRequests:    make(chan createTxRequest),
		//unlockRequests:      make(chan unlockRequest),
		//lockRequests:        make(chan struct{}),
		//holdUnlockRequests:  make(chan chan heldUnlock),
		//lockState:           make(chan bool),
		//changePassphrase:    make(chan changePassphraseRequest),
		//changePassphrases:   make(chan changePassphrasesRequest),
		chainParams:         params,
		//quit:                make(chan struct{}),
	}

	//w.NtfnServer = newNotificationServer(w)
	//w.TxStore.NotifyUnspent = func(hash *chainhash.Hash, index uint32) {
	//	w.NtfnServer.notifyUnspentOutput(0, hash, index)
	//}

	return w, nil
}

func (w *Wallet) GetTx(txid string) (string,error){

	trx:=corejson.TxRawResult{}
	err:=walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		ns:=tx.ReadBucket(wtxmgrNamespaceKey)
		txns:=ns.NestedReadBucket(wtxmgr.BucketTxJson)
		k,err:=hash.NewHashFromStr(txid)
		if(err!=nil){
			fmt.Println("GetTx err:",err.Error())
			return err
		}
		v:=txns.Get(k.Bytes())
		if(v!=nil){
			err:=json.Unmarshal(v,&trx)
			if(err!=nil){
				fmt.Println(" Unmarshal err:",err.Error())
				return err
			}
		}else{
			return errors.New("GetTx fail ")
		}
		return nil
	})
	if(err!=nil){
		return "",err
	}
	b,err:=json.Marshal(trx)
	if(err!=nil){
		fmt.Println("json.Marshal err：",err.Error())
		return "",err
	}
	return string(b),nil
}

func (w *Wallet) GetBalance(addr string) (balance,error){
	b:=balance{}
	if(addr ==""){
		return b,errors.New("addr is nil")
	}
	add:=&addr
	var txouts []*wtxmgr.AddrTxOutput
	var txins []*types.TxOutPoint
	walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		hs:=[]byte(*add)
		ns:=tx.ReadBucket(wtxmgrNamespaceKey)
		inns:=ns.NestedReadBucket(wtxmgr.BucketAddrtxin)
		outns:=ns.NestedReadBucket(wtxmgr.BucketAddrtxout)
		hsinns:=inns.NestedReadBucket(hs)
		if(hsinns !=nil){
			hsinns.ForEach(func(k, v []byte) error {
				tin :=types.TxOutPoint{}
				err:=wtxmgr.ReadCanonicalOutPoint(v,&tin)
				if(err!=nil){
					fmt.Println("ReadCanonicalOutPoint err:",err.Error())
					return err
				}
				txins=append(txins,&tin)
				return nil
			})
		}
		hsoutns:=outns.NestedReadBucket(hs)
		if(hsoutns!=nil){
			hsoutns.ForEach(func(k, v []byte) error {
				to:=wtxmgr.AddrTxOutput{}
				err:=wtxmgr.ReadAddrTxOutput(v,&to)
				if err!=nil{
					fmt.Println("ReadAddrTxOutput err:",err.Error())
					return err
				}
				txouts=append(txouts,&to)
				return nil
			})
		}
		return nil
	})
	var spendAmount float64
	var unspendAmount float64
	var totalAmount float64
	for i:=0;i< len(txouts);i++  {
		to:=txouts[i]
		for j:=0; j<len(txins);j++  {
			ti:=txins[j]
			if(ti.Hash.IsEqual(&to.Txid)&&to.Index==ti.OutIndex){
				spendAmount+=to.Amount.ToCoin()
				totalAmount+=to.Amount.ToCoin()
				break
			}
		}
		totalAmount+=to.Amount.ToCoin()
		unspendAmount +=to.Amount.ToCoin()
	}
	b.UnspendAmount=unspendAmount
	b.SpendAmount=spendAmount
	b.TotalAmount=totalAmount
	return b,nil
}

func (w *Wallet) SyncTx(h string) error{
	tx,err:=w.Httpclient.getBlock(h,true)
	if(err!=nil){
		fmt.Println("getblockcount err:",err.Error())
		return err
	}
	if(tx==""){
		fmt.Println("tx is null")
		return err
	}
	fmt.Println("tx is :",tx)
	var block clijson.BlockHttpResult
	if err := json.Unmarshal([]byte(tx), &block); err == nil {
		//var txins []*types.TxOutPoint
		//var txouts []*wtxmgr.AddrTxOutput
		//var trrs []*corejson.TxRawResult
		txins ,txouts,trrs,err:=parseBlockTxs(block)
		if(err!=nil){
			fmt.Println("err :",err.Error())
			return err
		}
		walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
			ns:=tx.ReadWriteBucket(wtxmgrNamespaceKey)
			txns:=ns.NestedReadWriteBucket(wtxmgr.BucketTxJson)
			inns:=ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxin)
			outns:=ns.NestedReadWriteBucket(wtxmgr.BucketAddrtxout)
			for a:=0;a<len(trrs) ;a++  {
				tr:=trrs[a]
				k,err:=hash.NewHashFromStr(tr.Txid)
				if(err!=nil){
					fmt.Println("format tx key to byte err:",err.Error())
					return err
				}
				v,err:=json.Marshal(tr)
				if(err!=nil){
					fmt.Println("format tx value to byte err:",err.Error())
					return err
				}
				ks:=k.Bytes()
				txns.Put(ks,v)
			}
			for a:=0; a<len(txins);a++  {
				txi :=txins[a]
				v :=txns.Get(txi.Hash.Bytes())
				if(v==nil){
					continue
				}
				var txr corejson.TxRawResult
				err:=json.Unmarshal(v,&txr)
				if err!=nil{
					fmt.Println("Unmarshal tx err:",err.Error())
					return err
				}
				addr:=txr.Vout[txi.OutIndex].ScriptPubKey.Addresses[0]
				err =w.TxStore.UpdateAddrTxIn(inns,addr,txi)
				if err!=nil{
					fmt.Println("UpdateAddrTxIn err:",err.Error())
					return err
				}
			}
			for a:=0; a<len(txouts);a++  {
				txo:=txouts[a]
				err:=w.TxStore.UpdateAddrTxOut(outns,txo)
				if err!=nil{
					fmt.Println("UpdateAddrTxOut err:",err.Error())
					return err
				}
			}
			return err
		})
	} else {
		fmt.Println(err)
		return err
	}
	//fmt.Println("tx:",tx)
	return nil
}

func parseBlockTxs(block clijson.BlockHttpResult) ([]*types.TxOutPoint,[]*wtxmgr.AddrTxOutput, []*corejson.TxRawResult,error){
	//var txins []types.TxOutPoint
	//var txouts []wtxmgr.AddrTxOutput
	var txins []*types.TxOutPoint
	var txouts []*wtxmgr.AddrTxOutput
	var tx []*corejson.TxRawResult
	for i:=0;i< len(block.Transactions); i++ {
		tr:= block.Transactions[i]
		tx=append(tx,&tr)
		blockhash,err:=hash.NewHashFromStr(tr.BlockHash)
		if(err!=nil){
			fmt.Println("vin NewHashFromStr err :", err.Error())
			return nil,nil,nil,err
		}
		block:=wtxmgr.Block{
			Hash:*blockhash,
			Height:int32(tr.BlockHeight),
		}
		txid,err:=hash.NewHashFromStr(tr.Txid)
		if(err !=nil) {
			fmt.Println("vin NewHashFromStr err :", err.Error())
			return nil,nil,nil,err
		}
		for j:=0;j<len(tr.Vin) ;j++  {
			vi:=tr.Vin[j]
			if(vi.Txid==""&&vi.ScriptSig ==nil){
				continue
			}else{
				hs,err:=hash.NewHashFromStr(vi.Txid)
				if(err !=nil){
					fmt.Println("vin NewHashFromStr err :",err.Error())
					return nil,nil,nil,err
				}else{
					txin:=&types.TxOutPoint{
						Hash:*hs,
						OutIndex:vi.Vout,
					}
					txins=append(txins,txin)
				}
			}
		}
		for k:=0;k<len(tr.Vout) ; k++ {
			vo:=tr.Vout[k]
			if len(vo.ScriptPubKey.Addresses)==0{
				continue
			}else{
				txout:=&wtxmgr.AddrTxOutput{
					Address:vo.ScriptPubKey.Addresses[0],
					Txid:*txid,
					Index:uint32(k),
					Amount:types.Amount(vo.Amount),
					Block:block,
					Spend:"0",
				}
				txouts=append(txouts,txout)
			}
		}
	}
	return txins,txouts,tx, nil
}

func (w *Wallet) Updateblock(){
	blockcount,err:=w.Httpclient.getblockCount()
	if(err!=nil){
		fmt.Println("getblockcount err:",err.Error())
		return
	}
	fmt.Println("blockcount:",blockcount)
	fmt.Println("httpclienr:",w.Httpclient.RPCServer)
	fmt.Println("blockcount:",w.Httpclient.httpclient)
	if(blockcount!= ""){
		blockheight,err:= strconv.ParseInt(blockcount, 10, 32)
		if(err!=nil){
			fmt.Println("string to int  err:",err.Error())
			return
		}
		log.Info("getblockcount :",blockheight)
		localheight:=w.Manager.SyncedTo().Height+1
		for h :=localheight;localheight<=int32(blockheight) ;h++  {
			blockhash,err:=w.Httpclient.getBlockhash(int64(h))
			if(err!=nil){
				fmt.Println("getblockhash err:",err.Error())
				break
			}
			er:=w.SyncTx(blockhash)
			if(er!=nil){
				fmt.Println("SyncTx err :",err.Error())
				return
			}
			//fmt.Println(len(blockhash))
			//fmt.Println("1")
			log.Info("localheight:",h," blockhash:",blockhash)
			hs,err:=hash.NewHashFromStr(blockhash)
			//fmt.Println("hs:",hs)
			if(err!=nil){
				fmt.Println("blockhash string to hash  err:",err.Error())
			}
			stamp := &waddrmgr.BlockStamp{Hash: *hs, Height: h}
			walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
				ns:=tx.ReadWriteBucket(waddrmgrNamespaceKey)
				err:=w.Manager.SetSyncedTo(ns,stamp)
				if(err!=nil){
					fmt.Println("db err:",err.Error())
					return err
				}
				return nil
			})
			fmt.Println("localheight:",h," blockhash:",blockhash)
		}
	}else{
		fmt.Println("getblockcount err:",err.Error())
		return
	}
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
		props   *waddrmgr.AccountProperties
	)
	err = walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		account, err = manager.NewAccount(addrmgrNs, name)
		if err != nil {
			return err
		}
		props, err = manager.AccountProperties(addrmgrNs, account)
		fmt.Println("props：",props)
		return err
	})
	if err != nil {
		log.Errorf("Cannot fetch new account properties for notification "+
			"after account creation: %v", err)
	}
	return account, err
}

// AccountBalances returns all accounts in the wallet and their balances.
// Balances are determined by excluding transactions that have not met
// requiredConfs confirmations.
func (w *Wallet) AccountBalances(scope waddrmgr.KeyScope,
	requiredConfs int32) ([]AccountBalanceResult, error) {

	manager, err := w.Manager.FetchScopedKeyManager(scope)
	if err != nil {
		return nil, err
	}

	var results []AccountBalanceResult
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		txmgrNs := tx.ReadBucket(wtxmgrNamespaceKey)

		syncBlock := w.Manager.SyncedTo()

		// Fill out all account info except for the balances.
		lastAcct, err := manager.LastAccount(addrmgrNs)
		if err != nil {
			return err
		}
		results = make([]AccountBalanceResult, lastAcct+2)
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

		// Fetch all unspent outputs, and iterate over them tallying each
		// account's balance where the output script pays to an account address
		// and the required number of confirmations is met.
		unspentOutputs, err := w.TxStore.UnspentOutputs(txmgrNs)
		if err != nil {
			return err
		}
		for i := range unspentOutputs {
			output := &unspentOutputs[i]
			if !confirmed(requiredConfs, output.Height, syncBlock.Height) {
				continue
			}
			if output.FromCoinBase && !confirmed(int32(w.ChainParams().CoinbaseMaturity),
				output.Height, syncBlock.Height) {
				continue
			}
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(txscript.DefaultScriptVersion,output.PkScript, w.chainParams)
			if err != nil || len(addrs) == 0 {
				continue
			}
			outputAcct, err := manager.AddrAccount(addrmgrNs, addrs[0])
			if err != nil {
				continue
			}
			switch {
			case outputAcct == waddrmgr.ImportedAddrAccount:
				results[len(results)-1].AccountBalance += output.Amount
			case outputAcct > lastAcct:
				return errors.New("waddrmgr.Manager.AddrAccount returned account " +
					"beyond recorded last account")
			default:
				results[outputAcct].AccountBalance += output.Amount
			}
		}
		return nil
	})
	return results, err
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
	scope waddrmgr.KeyScope,account uint32) (types.Address, error) {
	//chainClient, err := w.requireChainClient()
	//if err != nil {
	//	return nil, err
	//}

	var (
		addr  types.Address
		props *waddrmgr.AccountProperties
	)
	err:= walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		var err error
		addr, props, err = w.newAddress(addrmgrNs, account, scope)
		return err
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("props:",props)
	//// Notify the rpc server about the newly created address.
	//err = chainClient.NotifyReceived([]btcutil.Address{addr})
	//if err != nil {
	//	return nil, err
	//}
	//
	//w.NtfnServer.notifyAccountProperties(props)

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
		log.Errorf("Cannot fetch account properties for notification "+
			"after deriving next external address: %v", err)
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

	wif, err := pka.ExportPrivKey()
	if err != nil {
		return "", err
	}
	return wif.String(), nil
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
		return curHeight - txHeight + 1
	}
}
// Unlock unlocks the wallet's address manager and relocks it after timeout has
// expired.  If the wallet is already unlocked and the new passphrase is
// correct, the current timeout is replaced with the new one.  The wallet will
// be locked if the passphrase is incorrect or any other error occurs during the
// unlock.
func (w *Wallet) Unlock(passphrase []byte, lock <-chan time.Time) error {
	err := make(chan error, 1)
	w.unlockRequests <- unlockRequest{
		passphrase: passphrase,
		lockAfter:  lock,
		err:        err,
	}
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

func (w *Wallet) UnLockManager(passphrase []byte) error{
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
	var timeout <-chan time.Time
	//holdChan := make(heldUnlock)
	quit := w.quitChan()
out:
	for {
		select {
		case req := <-w.unlockRequests:
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

		//case req := <-w.changePassphrase:
		//	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		//		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		//		return w.Manager.ChangePassphrase(
		//			addrmgrNs, req.old, req.new, req.private,
		//			&waddrmgr.DefaultScryptOptions,
		//		)
		//	})
		//	req.err <- err
		//	continue
		//
		//case req := <-w.changePassphrases:
		//	err := walletdb.Update(w.db, func(tx walletdb.ReadWriteTx) error {
		//		addrmgrNs := tx.ReadWriteBucket(waddrmgrNamespaceKey)
		//		err := w.Manager.ChangePassphrase(
		//			addrmgrNs, req.publicOld, req.publicNew,
		//			false, &waddrmgr.DefaultScryptOptions,
		//		)
		//		if err != nil {
		//			return err
		//		}
		//
		//		return w.Manager.ChangePassphrase(
		//			addrmgrNs, req.privateOld, req.privateNew,
		//			true, &waddrmgr.DefaultScryptOptions,
		//		)
		//	})
		//	req.err <- err
		//	continue
		//
		//case req := <-w.holdUnlockRequests:
		//	if w.Manager.IsLocked() {
		//		close(req)
		//		continue
		//	}
		//
		//	req <- holdChan
		//	<-holdChan // Block until the lock is released.
		//
		//	// If, after holding onto the unlocked wallet for some
		//	// time, the timeout has expired, lock it now instead
		//	// of hoping it gets unlocked next time the top level
		//	// select runs.
		//	select {
		//	case <-timeout:
		//		// Let the top level select fallthrough so the
		//		// wallet is locked.
		//	default:
		//		continue
		//	}

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
			log.Errorf("Could not lock wallet: %v", err)
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
// SendOutputs creates and sends payment transactions. It returns the
// transaction upon success.
//func (w *Wallet) SendOutputs(outputs []*types.TxOutput, account uint32,
//	minconf int32, satPerKb types.Amount) (*types.Transaction, error) {
//
//	// Ensure the outputs to be created adhere to the network's consensus
//	// rules.
//	for _, output := range outputs {
//		if err := txrules.CheckOutput(output, satPerKb); err != nil {
//			return nil, err
//		}
//	}
//
//	// Create the transaction and broadcast it to the network. The
//	// transaction will be added to the database in order to ensure that we
//	// continue to re-broadcast the transaction upon restarts until it has
//	// been confirmed.
//	createdTx, err := w.CreateSimpleTx(
//		account, outputs, minconf, satPerKb, false,
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	txHash, err := w.reliablyPublishTransaction(createdTx.Tx)
//	if err != nil {
//		return nil, err
//	}
//
//	// Sanity check on the returned tx hash.
//	if *txHash != createdTx.Tx.TxHash() {
//		return nil, errors.New("tx hash mismatch")
//	}
//
//	return createdTx.Tx, nil
//}