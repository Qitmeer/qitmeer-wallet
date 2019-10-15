package main

import (
	//"fmt"

	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"


	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/wserver"
)
var rootCmd = &cobra.Command{
	Use:               "qitmeer-wallet",
	Long:              `qitmeer wallet util`,
}
var preCfg *config.Config
func bindFlags(){
	preCfg=&config.Config{}
	rootCmd.PersistentFlags().StringVarP(&preCfg.ConfigFile, "configfile", "c", "config.toml", "config file")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "debuglevel", "d", "error", "log level")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "appdatadir", "a", "", "wallet db path")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "logdir", "l", "", "log data path")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "network", "n", "testnet", "network")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "qserver", "s", "127.0.0.1:8030", "qitmeer node server")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "quser", "u", "admin", "qitmeer node user")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "qpass", "p", "123456", "qitmeer node password")
	rootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "pubwalletpass", "P", "public", "data encryption password")
}

func init()  {
	bindFlags()
	rootCmd.AddCommand(config.GenerateCmd)
}
func main() {
	rootCmd.PersistentPreRunE=LoadConfig
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

// LoadConfig config file and flags
func LoadConfig(cmd *cobra.Command, args []string) (err error) {

	// debug
	//if cmd.Flag("debug").Changed && preCfg.DebugLevel {
	//
	//	log.SetLevel(log.TraceLevel)
	//}

	// load configfile ane merge command ,but don't udpate configfile
	fileCfg := config.Cfg
	_, err = toml.DecodeFile(preCfg.ConfigFile, fileCfg)
	if err != nil {

		//if not set config file and default cli.toml decode err, use default set only.
		if !cmd.Flag("configfile").Changed {

			if fExit, _ := utils.FileExists(preCfg.ConfigFile); fExit {
				return fmt.Errorf("config file err: %s", err)
			}

			return nil
		}
		return fmt.Errorf("config file err: %s", err)
	}

	fileCfg.ConfigFile = preCfg.ConfigFile

	if cmd.Flag("debuglevel").Changed {
		fileCfg.DebugLevel = preCfg.DebugLevel
	}
	if cmd.Flag("appdatadir").Changed {
		fileCfg.AppDataDir = preCfg.AppDataDir
	}
	if cmd.Flag("logdir").Changed {
		fileCfg.LogDir = preCfg.LogDir
	}
	if cmd.Flag("network").Changed {
		fileCfg.Network = preCfg.Network
	}
	//if cmd.Flag("listeners").Changed {
	//	fileCfg.Listeners = preCfg.Listeners
	//}
	//if cmd.Flag("rpcuser").Changed {
	//	fileCfg.RPCUser = preCfg.RPCUser
	//}
	//if cmd.Flag("rpcpass").Changed {
	//	fileCfg.RPCPass = preCfg.RPCPass
	//}
	//
	//if cmd.Flag("rpccert").Changed {
	//	fileCfg.RPCCert = preCfg.RPCCert
	//}
	//if cmd.Flag("rpckey").Changed {
	//	fileCfg.RPCKey = preCfg.RPCKey
	//}
	//if cmd.Flag("rpcmaxclients").Changed {
	//	fileCfg.RPCMaxClients = preCfg.RPCMaxClients
	//}
	//
	//if cmd.Flag("disablerpc").Changed {
	//	fileCfg.DisableRPC = preCfg.DisableRPC
	//}
	//
	//if cmd.Flag("disabletls").Changed {
	//	fileCfg.DisableTLS = preCfg.DisableTLS
	//}
	if cmd.Flag("qserver").Changed {
		fileCfg.QServer = preCfg.QServer
	}
	if cmd.Flag("quser").Changed {
		fileCfg.QUser = preCfg.QUser
	}
	if cmd.Flag("qpass").Changed {
		fileCfg.QPass = preCfg.QPass
	}
	if cmd.Flag("pubwalletpass").Changed {
		fileCfg.WalletPass = preCfg.WalletPass
	}
	log.SetLevel(log.TraceLevel)
	log.Debug("fileCfg: ", *fileCfg)


	return nil

	//save
	// buf := new(bytes.Buffer)
	// if err := toml.NewEncoder(buf).Encode(*fileCfg); err != nil {
	// 	log.Fatal(err)
	// }

	//return ioutil.WriteFile(fileCfg.configFile, buf.Bytes(), 0666)
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
