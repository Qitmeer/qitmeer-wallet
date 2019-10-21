package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/console"
	log "github.com/sirupsen/logrus"
	//"github.com/spf13/cobra"
	"os"


	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/wserver"
)
var rootCmd =console.Command


func init()  {
	console.BindFlags()
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//if config.Cfg.Create{
	//	console.CreatWallet()
	//}else {
	//	logLevel, err := log.ParseLevel(config.Cfg.DebugLevel)
	//	if err != nil {
	//		return
	//	}
	//	log.SetLevel(logLevel)
	//	QitmeerMain(config.Cfg)
	//}


	//fmt.Println(config.Cfg.WalletPass)
	//fmt.Println(config.Cfg.AppDataDir)
	//fmt.Println(config.Cfg.DebugLevel)
	//fmt.Println(config.Cfg.LogDir)
	//if len(os.Args) > 1 && os.Args[1] == "-console" {
	//	console.StartConsole()
	//}else{
	//	cmd := newCmd()
	//	err := cmd.Execute()
	//	if err != nil {
	//		log.Errorf("main err: %s", err)
	//	}
	//}
}
//
//
//func newCmd() (cmd *cobra.Command) {
//	// root Command
//	cmd = &cobra.Command{
//		Use:   "qitmeer-wallet",
//		Short: `qitmeer-wallet`,
//		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
//			var err error
//			//cfg, err = config.LoadConfig(cfg.ConfigFile, false, cfg)
//			//if err != nil {
//			//	return fmt.Errorf("cmd PersistentPreRunE err: %s", err)
//			//}
//			logLevel, err := log.ParseLevel(config.Cfg.DebugLevel)
//			if err != nil {
//				return fmt.Errorf("cmd LogLevl err: %s", err)
//			}
//			log.SetLevel(logLevel)
//
//			err = config.Cfg.Check()
//			if err != nil {
//				return err
//			}
//			return nil
//		},
//		Run: func(cmd *cobra.Command, args []string) {
//			QitmeerMain(config.Cfg)
//		},
//	}
//
//	// Create Wallet Command
//	createCmd := &cobra.Command{
//		Use:   "create",
//		Short: "create new wallet or recover wallet from seed",
//		Run: func(cmd *cobra.Command, args []string) {
//			fmt.Println("wallet create")
//		},
//	}
//
//	// version Command
//	versionCmd := &cobra.Command{
//		Use:   "version",
//		Short: "show version",
//		Run: func(cmd *cobra.Command, args []string) {
//			fmt.Println(version.Version())
//		},
//	}
//
//	cmd.AddCommand(versionCmd)
//	cmd.AddCommand(createCmd)
//
//	return
//}

// QitmeerMain wallet main
func QitmeerMain(cfg *config.Config) {
	log.Trace("Qitmeer Main")
	wsvr, err := wserver.NewWalletServer(cfg)
	if err != nil {
		log.Errorf("NewWalletServer err: %s", err)
		return
	}
	wsvr.Start()

	exitCh := make(chan int)
	<-exitCh
}
