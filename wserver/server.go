package wserver

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"github.com/HalalChain/qitmeer-wallet/assets"
	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/rpc/server"
	"github.com/HalalChain/qitmeer-wallet/utils"
	"github.com/HalalChain/qitmeer-wallet/wallet"
)

//WalletServer wallet api server
type WalletServer struct {
	cfg *config.Config

	RPCSvr *server.RpcServer
}

//NewWalletServer make a wallet api server
func NewWalletServer(cfg *config.Config, wt *wallet.Wallet) (wSvr *WalletServer, err error) {
	wSvr = &WalletServer{
		cfg: cfg,
	}

	RPCSvrCfg := &server.Config{
		RPCUser:       cfg.RPCUser,
		RPCPass:       cfg.RPCPass,
		RPCCert:       cfg.RPCCert,
		RPCKey:        cfg.RPCKey,
		RPCMaxClients: 100,
		DisableRPC:    false,
		DisableTLS:    cfg.DisableTLS,
	}

	wSvr.RPCSvr, err = server.NewRPCServer(RPCSvrCfg)
	if err != nil {
		return nil, fmt.Errorf("NewWallet: %s", err)
	}

	for _, api := range cfg.APIs {
		switch api {
		case "account":
			//wSvr.RPCSvr.RegisterService("account", wallet.NewAPI(wt))
		case "tx":
			//wSvr.RPCSvr.RegisterService("tx", &services.TxAPI{})
		}
	}

	return
}

//
func (wsvr *WalletServer) runSvr() {
	defer func() {
		if rev := recover(); rev != nil {
			log.Println("server run recover: ", rev)
		}
		go wsvr.runSvr()
	}()

	log.Trace("wallet runSvr")

	router := httprouter.New()

	if wsvr.cfg.UI {
		staticF, err := assets.GetStatic()
		if err != nil {
			log.Println("server run err: ", err)
			return
		}

		router.ServeFiles("/app/*filepath", staticF)
		router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			http.Redirect(w, r, "app/index.html", http.StatusMovedPermanently)
		})
	}

	router.POST("/api", wsvr.HandleAPI)

	for _, lis := range wsvr.cfg.Listeners {

		go func() {
			log.Infof("Experimental RPC server listening on %s", lis)
			err := http.ListenAndServe(":38130", router)
			log.Tracef("Finished serving expimental RPC: %v", err)
		}()
	}
}

// HandleAPI RPC Method
func (wsvr *WalletServer) HandleAPI(ResW http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wsvr.RPCSvr.HandleFunc(ResW, r)
}

// Start routine
func (wsvr *WalletServer) Start() error {

	log.Trace("wallet server start")

	wsvr.RPCSvr.Start()

	//open home in web browser

	utils.OpenBrowser("http://" + wsvr.cfg.Listeners[0])

	return nil
}
