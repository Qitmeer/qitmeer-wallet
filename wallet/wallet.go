package wallet

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"

	qitmeerConfig "github.com/HalalChain/qitmeer-lib/config"

	"github.com/HalalChain/qitmeer-wallet/assets"
	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/rpc/server"
	"github.com/HalalChain/qitmeer-wallet/services"
	"github.com/HalalChain/qitmeer-wallet/utils"


	"encoding/json"
	"errors"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	chaincfg "github.com/HalalChain/qitmeer-lib/params"
	clijson "github.com/HalalChain/qitmeer-wallet/json"
	corejson "github.com/HalalChain/qitmeer-lib/core/json"
	"github.com/HalalChain/qitmeer-wallet/waddrmgs"
	"github.com/HalalChain/qitmeer-wallet/walletdb"
	"github.com/HalalChain/qitmeer-wallet/wtxmgr"
	"github.com/HalalChain/qitmeer-lib/core/types"
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

// Namespace bucket keys.
var (
	waddrmgrNamespaceKey = []byte("waddrmgr")
	wtxmgrNamespaceKey   = []byte("wtxmgr")
	// 地址对应的交易in out 桶
	//waddrtrNamespaceKey   = []byte("waddrtr")
)



// Start Wallet start
func (w *Wallet) Start() error {

	log.Trace("wallet start")

	w.RPCSvr.Start()

	go w.runSvr()

	//open home in web browser

	utils.OpenBrowser("http://" + w.cfg.Listen)

	return nil
}

// NewWallet make wallet server
func NewWallet(cfg *config.Config) (w *Wallet, err error) {

	w = &Wallet{
		cfg: cfg,
	}

	RPCSvrCfg := qitmeerConfig.Config{
		RPCUser:       cfg.RPCUser,
		RPCPass:       cfg.RPCPass,
		RPCCert:       cfg.RPCCert,
		RPCKey:        cfg.RPCKey,
		RPCMaxClients: 100,
		DisableRPC:    false,
		DisableTLS:    cfg.DisableTLS,
	}

	w.RPCSvr, err = server.NewRPCServer(&RPCSvrCfg)
	if err != nil {
		return nil, fmt.Errorf("NewWallet: %s", err)
	}

	for _, api := range cfg.Apis {
		switch api {
		case "account":
			w.RPCSvr.RegisterService("account", &services.AccountAPI{})
		case "tx":
			w.RPCSvr.RegisterService("tx", &services.TxAPI{})
		}
	}

	return
}

//
func (w *Wallet) runSvr() {
	defer func() {
		if rev := recover(); rev != nil {
			log.Println("server run recover: ", rev)
		}
		go w.runSvr()
	}()

	log.Trace("wallet runSvr")

	router := httprouter.New()

	staticF, err := assets.GetStatic()
	if err != nil {
		log.Println("server run err: ", err)
		return
	}

	router.ServeFiles("/app/*filepath", staticF)
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.Redirect(w, r, "app/index.html", http.StatusMovedPermanently)
	})

	router.POST("/api", w.HandleAPI)

	log.Fatal(http.ListenAndServe(":38130", router))
}

// HandleAPI RPC Method
func (w *Wallet) HandleAPI(ResW http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.RPCSvr.HandleFunc(ResW, r)
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