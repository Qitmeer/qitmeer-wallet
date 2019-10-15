package wserver

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"

	"github.com/Qitmeer/qitmeer-wallet/assets"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/rpc/server"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
)

//WalletServer wallet api server
type WalletServer struct {
	cfg *config.Config

	WtLoader *wallet.Loader
	Wt       *wallet.Wallet

	RPCSvr *server.RpcServer

	exitCh chan bool
}

//NewWalletServer make a wallet api server
func NewWalletServer(cfg *config.Config) (wSvr *WalletServer, err error) {

	activeNetParams := utils.GetNetParams(cfg.Network)
	dbDir := filepath.Join(cfg.AppDataDir, cfg.Network)
	wtLoader := wallet.NewLoader(activeNetParams, dbDir, 250, cfg)
	wtExist, err := wtLoader.WalletExists()
	if err != nil {
		return nil, fmt.Errorf("load wallet err: %s", err)
	}
	if !wtExist && !cfg.UI {
		return nil, fmt.Errorf("not wallet exist,please run crate command")
	}

	wSvr = &WalletServer{
		cfg: cfg,

		WtLoader: wtLoader,
		// wt:     &wallet.Wallet{},

		exitCh: make(chan bool),
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

	// for _, api := range cfg.APIs {
	// 	switch api {
	// 	case "account":
	// 		wSvr.RPCSvr.RegisterService("account", &services.AccountAPI{})
	// 	case "tx":
	// 		//wSvr.RPCSvr.RegisterService("tx", &services.TxAPI{})
	// 	}
	// }

	wSvr.RPCSvr.RegisterService("wallet", NewAPI(cfg, wSvr))

	// if !wtExist && cfg.UI {
	// 	wSvr.RPCSvr.RegisterService("crate", wallet.NewCreateAPI(cfg, wSvr.wt))
	// }
	return
}

//
func (wsvr *WalletServer) run() {
	defer func() {
		if rev := recover(); rev != nil {
			log.Println("WalletServer run recover: ", rev)
			go wsvr.run()
		}
	}()
	go func() {
		for {
			select {
			case <-wsvr.exitCh:
				os.Exit(1)
			}
		}
	}()
	log.Trace("WalletServer run")

	router := httprouter.New()

	if wsvr.cfg.UI {
		staticF, err := assets.GetStatic()
		if err != nil {
			log.Println("server run err: ", err)
			return
		}
		myStaticF := assets.NewMyStatic(staticF)

		fmt.Println("wsvr.cfg.ApiUrl:",wsvr.cfg.ApiUrl)

		myStaticF.AddFilter("/config.js", func() []byte {

			//update config.js
			tmpl := `
			//config
			window.QitmeerConfig = {
				RPCAddr: "{{api_url}}",
				RPCUser: "{{rpc_user}}",
				RPCPass: "{{rpc_pass}}"
			};
			`
			tmpl = strings.Replace(tmpl, "{{api_url}}", "http://"+wsvr.cfg.ApiUrl+"/api", -1)
			tmpl = strings.Replace(tmpl, "{{rpc_user}}", wsvr.cfg.RPCUser, -1)
			tmpl = strings.Replace(tmpl, "{{rpc_pass}}", wsvr.cfg.RPCPass, -1)

			return []byte(tmpl)
		})

		router.ServeFiles("/app/*filepath", myStaticF)
		router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			http.Redirect(w, r, "app/index.html", http.StatusMovedPermanently)
		})

		//ajx post options
		router.OPTIONS("/api", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			//log.Trace("api OPTIONS")
			w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8080")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
			w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers, Authorization")
			return
		})
	}

	router.POST("/api", wsvr.HandleAPI)

	for _, addr := range wsvr.cfg.Listeners {
		go func() {
			log.Infof("WalletServer listening on %s", addr)
			err := http.ListenAndServe(addr, router)
			if err != nil {
				log.Errorf("server listen err: %v", err)
				wsvr.exitCh <- true
				return
			}
		}()
	}
}

// HandleAPI RPC Method
func (wsvr *WalletServer) HandleAPI(ResW http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	ResW.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8080")
	ResW.Header().Set("Access-Control-Allow-Credentials", "true")
	ResW.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT")
	ResW.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin, Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers, Authorization")

	wsvr.RPCSvr.HandleFunc(ResW, r)
}

// Start routine
func (wsvr *WalletServer) Start() error {
	log.Trace("WalletServer start")

	wsvr.RPCSvr.Start()

	go wsvr.run()

	//open home in web browser
	if wsvr.cfg.UI {
		utils.OpenBrowser("http://" + wsvr.cfg.Listeners[0])
	}

	return nil
}

// StartAPI if wallet open ok start api
func (wsvr *WalletServer) StartAPI() {
	log.Trace("StartAPI", wsvr.cfg.APIs)
	for _, api := range wsvr.cfg.APIs {
		switch api {
		case "account":
			wsvr.RPCSvr.RegisterService("account", wallet.NewAPI(wsvr.cfg, wsvr.Wt))
		case "tx":
			//wSvr.RPCSvr.RegisterService("tx", &services.TxAPI{})
		}
	}
}
