package server

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/HalalChain/qitmeer-lib/config"

	"github.com/HalalChain/qitmeer-wallet/rpc/server"
	"github.com/HalalChain/qitmeer-wallet/tools"
)

var s *server.RpcServer

// Start web api
func Start() {

	cfg := config.Config{
		RPCUser:       "admin",
		RPCPass:       "123",
		RPCCert:       "",
		RPCKey:        "",
		RPCMaxClients: 100,
		DisableRPC:    false,
		DisableTLS:    true,
	}

	var err error
	s, err = server.NewRPCServer(&cfg)
	if err != nil {
		log.Println("server start err:", err)
		return
	}

	s.Start()

	s.RegisterService(server.DefaultServiceNameSpace, &CliAPI{})

	go run()
}

func run() {
	defer func() {
		if rev := recover(); rev != nil {
			log.Println("server run recover: ", rev)
		}
		go run()
	}()

	router := httprouter.New()
	router.ServeFiles("/app/*filepath", http.Dir("../../electron/dist/"))

	router.GET("/", Index)

	router.POST("/api", API)

	log.Fatal(http.ListenAndServe(":1236", router))
}

// Index app home page
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "app/index.html", http.StatusMovedPermanently)
}

// API RPC Method
func API(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	s.HandleFunc(w, r)
}

// CliAPI cli api
type CliAPI struct {
}

//MakeEntropy generate a cryptographically secure pseudorandom entropy
func (c *CliAPI) MakeEntropy() (seed string, err error) {
	seedBuf, err := tools.NewEntropy(32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(seedBuf), err
}

//MakeHdKey create a new HD(BIP32) private key from an entropy
func (c *CliAPI) MakeHdKey() (st string) {
	return "ok"
}

//ListAccount list all account
func (c *CliAPI) ListAccount() ([]string, error) {

	//
	//{"alias":"defalut","keys":""}

	return []string{"a", "b"}, nil
}

//NewAccount make a new account
func (c *CliAPI) NewAccount(alias string, pass string) ([]string, error) {
	return []string{}, nil
}

//ListAddresses list account all address
func (c *CliAPI) ListAddresses(alias string) ([]string, error) {
	return []string{}, nil
}

//NewAddress make new address account
func (c *CliAPI) NewAddress(alias string) (string, error) {

	return "", nil
}
