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
)

// Wallet qitmeer-wallet
type Wallet struct {
	cfg *config.Config

	RPCSvr *server.RpcServer
}

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
