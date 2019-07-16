package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	cg"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/utils"
	"github.com/HalalChain/qitmeer-wallet/wallet"
)

var (
	configFile string
)

func init() {

	log.SetLevel(log.TraceLevel)
	log.SetOutput(os.Stdout)

	defaultDataDir := utils.GetUserDataDir()
	flag.StringVar(&configFile, "c", defaultDataDir+"/config.toml", "-c path/to/config.toml")
	flag.Parse()
}

func main() {
	cfg, err := cg.Load(configFile)
	if err != nil {
		log.Printf("main: %s", err)
		return
	}
	if err := cfg.Init(); err != nil {
		log.Printf("main: %s", err)
		return
	}

	w, err := wallet.NewWallet(cfg)
	if err != nil {
		log.Printf("main: %s", err)
		return
	}
	w.Start()

	ch := make(chan int)
	<-ch
}
