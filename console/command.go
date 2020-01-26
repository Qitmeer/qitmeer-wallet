package console

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wserver"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "qitmeer-wallet",
}

var Command = &cobra.Command{
	Use:              "qc",
	Short:            "qitmeer wallet command",
	Long:             `qitmeer wallet command`,
	PersistentPreRun: LoadConfig,
}

var preCfg *config.Config
var fileCfg = config.Cfg

func BindFlags() {
	preCfg = &config.Config{}
	Command.PersistentFlags().StringVarP(&preCfg.ConfigFile, "configfile", "c", "config.toml", "config file")
	Command.PersistentFlags().StringVarP(&preCfg.DebugLevel, "debuglevel", "d", "info", "Logging level {trace, debug, info, warn, error, critical}")
	Command.PersistentFlags().StringVarP(&preCfg.AppDataDir, "appdatadir", "a", "", "wallet db path")
	Command.PersistentFlags().StringVarP(&preCfg.LogDir, "logdir", "l", "", "log data path")
	Command.PersistentFlags().StringVarP(&preCfg.Network, "network", "n", "testnet", "network")
	Command.PersistentFlags().BoolVar(&preCfg.Create, "create", false, "Create a new wallet")
	Command.PersistentFlags().StringVarP(&preCfg.QServer, "qserver", "s", "127.0.0.1:8030", "qitmeer node server")
	Command.PersistentFlags().StringVarP(&preCfg.QUser, "quser", "u", "admin", "qitmeer node user")
	Command.PersistentFlags().StringVarP(&preCfg.QPass, "qpass", "p", "123456", "qitmeer node password")
	Command.PersistentFlags().StringVarP(&preCfg.WalletPass, "pubwalletpass", "P", "public", "data encryption password")
	Command.PersistentFlags().BoolVar(&preCfg.QNoTLS, "qnotls", fileCfg.QNoTLS, "tls")
	Command.PersistentFlags().StringVar(&preCfg.QCert, "qcert", fileCfg.QCert, "Certificate path")
	Command.PersistentFlags().BoolVar(&preCfg.QTLSSkipVerify, "qtlsskipverify", fileCfg.QTLSSkipVerify, "tls skipverify")

	Command.PersistentFlags().Int64Var(&preCfg.Confirmations, "confirmations", 10, "Number of block confirmations ")
	Command.PersistentFlags().BoolVar(&preCfg.UI, "ui", true, "Start Wallet with RPC and webUI interface")
	Command.PersistentFlags().StringArrayVar(&preCfg.Listeners, "listeners", fileCfg.Listeners, "rpc listens")
	Command.PersistentFlags().StringVar(&preCfg.RPCUser, "rpcUser", fileCfg.RPCUser, "rpc user,default by random")
	Command.PersistentFlags().StringVar(&preCfg.RPCPass, "rpcPass", fileCfg.RPCPass, "rpc pass,default by random")
	Command.PersistentFlags().Int64Var(&preCfg.MinTxFee, "mintxfee", fileCfg.MinTxFee, "The minimum transaction fee in AtomMEER/kB default 20000 (aka. 0.0002 Qitmeer/kB)")
}



// LoadConfig config file and flags
func LoadConfig(cmd *cobra.Command, args []string) {
	// load configfile ane merge command ,but don't udpate configfile
	_, err := toml.DecodeFile(preCfg.ConfigFile, fileCfg)
	if err != nil {
		if cmd.Flag("configfile").Changed {
			if fExit, _ := utils.FileExists(preCfg.ConfigFile); fExit {
				log.Error(fmt.Sprintf("config file err: %s", err))
				return
			}
			return
		}
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
	if cmd.Flag("create").Changed {
		fileCfg.Create = preCfg.Create
	}
	if cmd.Flag("listeners").Changed {
		fileCfg.Listeners = preCfg.Listeners
	}
	if cmd.Flag("rpcUser").Changed {
		fileCfg.RPCUser = preCfg.RPCUser
	}
	if cmd.Flag("rpcPass").Changed {
		fileCfg.RPCPass = preCfg.RPCPass
	}
	if cmd.Flag("ui").Changed {
		fileCfg.UI = preCfg.UI
	}
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
	if cmd.Flag("qnotls").Changed {
		fileCfg.QNoTLS = preCfg.QNoTLS
	}
	if cmd.Flag("qcert").Changed {
		fileCfg.QCert = preCfg.QCert
	}
	if cmd.Flag("mintxfee").Changed {
		fileCfg.MinTxFee = preCfg.MinTxFee
	}
	if cmd.Flag("confirmations").Changed {
		fileCfg.Confirmations = preCfg.Confirmations
	}
	if cmd.Flag("qtlsskipverify").Changed {
		fileCfg.QTLSSkipVerify = preCfg.QTLSSkipVerify
	}
	config.ActiveNet = utils.GetNetParams(fileCfg.Network)
	funcName := "LoadConfig"
	// Parse, validate, and set debug log level(s).
	if err := parseAndSetDebugLevels(fileCfg.DebugLevel); err != nil {
		err := fmt.Errorf("%s: %v", funcName, err.Error())
		fmt.Fprintln(os.Stderr, err)
		return
	}
	InitLogRotator(filepath.Join(fileCfg.LogDir, "wallet.log"))

	if fileCfg.MinTxFee < config.DefaultMinRelayTxFee {
		fileCfg.MinTxFee = config.DefaultMinRelayTxFee
	}

	return

}

func parseAndSetDebugLevels(debugLevel string) error {
	// When the specified string doesn't have any delimters, treat it as
	// the log level for all subsystems.
	if !strings.Contains(debugLevel, ",") && !strings.Contains(debugLevel, "=") {
		// Validate debug log level.
		lvl, err := log.LvlFromString(debugLevel)
		if err != nil {
			str := "the specified debug level [%v] is invalid"
			return fmt.Errorf(str, debugLevel)
		}
		// Change the logging level for all subsystems.
		Glogger().Verbosity(lvl)
		return nil
	}
	// TODO support log for subsystem
	return nil
}

var createWalletCmd = &cobra.Command{
	Use:     "create",
	Short:   "create",
	Example: "create wallte",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		CreatWallet()
	},
}
var createNewAccountCmd = &cobra.Command{
	Use:     "createnewaccount {account} {pripassword}",
	Short:   "create new account",
	Example: "createnewaccount test password",
	Args:    cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = UnLock(args[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		_ = createNewAccount(args[0])
	},
}
var getnewaddressCmd = &cobra.Command{
	Use:     "getnewaddress {account}",
	Short:   "create new address by account",
	Example: "getnewaddress default",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		getNewAddress(args[0])
	},
}

var getBalanceCmd = &cobra.Command{
	Use:   "getbalance {address} {string ,company : i(int64),f(float),default i } {bool ,detail : true,false,default false }",
	Short: "getbalance",
	Example: `
		getbalance TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5	i true
		getbalance TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5	f false
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		company := "i"
		detail := "false"
		b, err := getBalance( args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(args) > 1 {
			if args[1] != "i" {
				company = "f"
			}
			if len(args) > 2 {
				detail = args[2]
			}
		}
		if company == "i" {
			if detail == "true" {
				fmt.Printf("unspend:%s\n", b.UnspendAmount.String())
				fmt.Printf("unconfirmed:%s\n", b.ConfirmAmount.String())
				fmt.Printf("totalamount:%s\n", b.TotalAmount.String())
				fmt.Printf("spendamount:%s\n", b.SpendAmount.String())
			} else {
				fmt.Printf("%s\n", b.UnspendAmount.String())
			}
		} else {
			if detail == "true" {
				fmt.Printf("unspend:%f\n", b.UnspendAmount.ToCoin())
				fmt.Printf("unconfirmed:%f\n", b.ConfirmAmount.ToCoin())
				fmt.Printf("totalamount:%f\n", b.TotalAmount.ToCoin())
				fmt.Printf("spendamount:%f\n", b.SpendAmount.ToCoin())
			} else {
				fmt.Printf("%f\n", b.UnspendAmount.ToCoin())
			}
		}

	},
}
var sendToAddressCmd = &cobra.Command{
	Use:   "sendtoaddress {address} {amount} {pripassword} ",
	Short: "send transaction ",
	Example: `
		sendtoaddress TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5 10 pripassword
		`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		f32, err := strconv.ParseFloat(args[1], 32)
		if err != nil {
			log.Error("sendtoaddress ", "error", err.Error())
			return
		}
		err = UnLock(args[2])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		sendToAddress(args[0], float64(f32))
	},
}
var getTxByTxIdCmd=&cobra.Command{
	Use:"gettx {txid}",
	Short:"Access to transaction information ",
	Example:`
		gettx 81278a6ba67d4ea2fc49fb469f2a45f6adb2306b82146747b9d5f3bd655e5030
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err:=OpenWallet()
		if err!=nil{
			fmt.Println(err.Error())
			return
		}
		getTx(args[0])
	},
}
var getAddressesByAccountCmd=&cobra.Command{
	Use:"getaddressesbyaccount {string ,account,defalut imported} ",
	Short:"get addresses by account ",
	Example:`
		getaddressesbyaccount imported
		`,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		account := "imported"
		if len(args) > 0 {
			account = args[0]
		}
		getAddressesByAccount(account)
	},
}
var importPriKeyCmd = &cobra.Command{
	Use:   "importprivkey {priKey} {pripassword}",
	Short: "import priKey ",
	Example: `
		importprivkey priKey pripassword
		`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		err = UnLock(args[1])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		importPrivKey(args[0])
	},
}
var listAccountsBalanceCmd = &cobra.Command{
	Use:   "listaccountsbalance ",
	Short: "list Accounts Balance",
	Example: `
		listaccountsbalance
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		listAccountsBalance()
	},
}
var getlisttxbyaddrCmd = &cobra.Command{
	Use:   "getlisttxbyaddr {address} {String ,Transaction type : in ,out ,all ,default all } ",
	Short: "get all transactions for address",
	Example: `
		getlisttxbyaddr Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ in
		getlisttxbyaddr Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ out 
		getlisttxbyaddr Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ all
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		stype := int32(2)
		if len(args) > 1 {
			if args[1] == "in" {
				stype = int32(0)
			} else if args[1] == "out" {
				stype = int32(1)
			} else {
				stype = int32(2)
			}
		}
		getListTxByAddr(args[0], int32(-1), int32(100), stype)
	},
}
var updateblockCmd = &cobra.Command{
	Use:   "updateblock {int,Update to the specified block, 0 is updated to the latest by default,defalut 0}",
	Short: "Update local block data",
	Example: `
		updateblock
		updateblock 12
		`,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		var height = int64(0)
		if len(args) > 0 {
			var err error
			height, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				log.Info("Argument is not of type int")
				return
			}
		}
		updateblock(height)
	},
}
var syncheightCmd = &cobra.Command{
	Use:   "syncheight",
	Short: "Get the number of local synchronization blocks",
	Example: `
		syncheight
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := OpenWallet()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		syncheight()
	},
}

// interactive mode

var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "console",
	Example: `
		Enter console mode
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		b := checkWalletIeExist(config.Cfg)
		if b == false {
			fmt.Println("Please create a wallet first,[qitmeer-wallet qc create ]")
			return
		}
		startConsole()
	},
}

// web mode

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "web",
	Example: `
		Enter web mode
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// b:=checkWalletIeExist(config.Cfg)
		// if b ==false{
		// 	fmt.Println("Please create a wallet first,[qitmeer-wallet qc create ]")
		// 	return
		// }
		fmt.Println("web model")
		qitmeerMain(fileCfg)
	},
}

func qitmeerMain(cfg *config.Config) {
	log.Trace("Qitmeer Main")
	wsvr, err := wserver.NewWalletServer(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("NewWalletServer err: %s", err))
		return
	}
	wsvr.Start()

	exitCh := make(chan int)
	<-exitCh
}

func init() {
	RootCmd.AddCommand(Command)
	RootCmd.AddCommand(QxCmd)
	QxCmd.AddCommand(generatemnemonicCmd)
	QxCmd.AddCommand(mnemonictoseedCmd)
	QxCmd.AddCommand(seedtopriCmd)
	QxCmd.AddCommand(pritopubCmd)
	QxCmd.AddCommand(mnemonictoaddrCmd)
	QxCmd.AddCommand(seedtoaddrCmd)
	QxCmd.AddCommand(pritoaddrCmd)
	QxCmd.AddCommand(pubtoaddrCmd)

	Command.AddCommand(createWalletCmd)
	Command.AddCommand(createNewAccountCmd)
	Command.AddCommand(getnewaddressCmd)
	Command.AddCommand(getBalanceCmd)
	Command.AddCommand(getlisttxbyaddrCmd)
	Command.AddCommand(updateblockCmd)
	Command.AddCommand(syncheightCmd)
	Command.AddCommand(sendToAddressCmd)
	Command.AddCommand(importPriKeyCmd)
	Command.AddCommand(getAddressesByAccountCmd)
	Command.AddCommand(listAccountsBalanceCmd)
	Command.AddCommand(consoleCmd)
	Command.AddCommand(getTxByTxIdCmd)
	Command.AddCommand(webCmd)
	Command.AddCommand(QxCmd)

}
