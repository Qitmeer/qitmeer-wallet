package console

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"strconv"
)

var Command = &cobra.Command{
	Use:               "qitmeer-wallet",
	Long:              `qitmeer wallet util`,
	PersistentPreRun:LoadConfig,
}

var preCfg *config.Config
var fileCfg =config.Cfg
func BindFlags(){
	preCfg=&config.Config{}
	Command.PersistentFlags().StringVarP(&preCfg.ConfigFile, "configfile", "c", "config.toml", "config file")
	Command.PersistentFlags().StringVarP(&preCfg.DebugLevel, "debuglevel", "d", "error", "log level")
	Command.PersistentFlags().StringVarP(&preCfg.AppDataDir, "appdatadir", "a", "", "wallet db path")
	Command.PersistentFlags().StringVarP(&preCfg.LogDir, "logdir", "l", "", "log data path")
	Command.PersistentFlags().StringVarP(&preCfg.Network, "network", "n", "testnet", "network")
	Command.PersistentFlags().BoolVar(&preCfg.Create, "create",false,"Create a new wallet")
	Command.PersistentFlags().StringVarP(&preCfg.QServer, "qserver", "s", "127.0.0.1:8030", "qitmeer node server")
	Command.PersistentFlags().StringVarP(&preCfg.QUser, "quser", "u", "admin", "qitmeer node user")
	Command.PersistentFlags().StringVarP(&preCfg.QPass, "qpass", "p", "123456", "qitmeer node password")
	Command.PersistentFlags().StringVarP(&preCfg.WalletPass, "pubwalletpass", "P", "public", "data encryption password")
}

// LoadConfig config file and flags
func LoadConfig(cmd *cobra.Command, args []string)  {
	// debug
	//if cmd.Flag("debug").Changed && preCfg.DebugLevel {
	//
	//	log.SetLevel(log.TraceLevel)
	//}

	// load configfile ane merge command ,but don't udpate configfile
	_, err := toml.DecodeFile(preCfg.ConfigFile, fileCfg)
	if err != nil {

		//if not set config file and default cli.toml decode err, use default set only.
		if !cmd.Flag("configfile").Changed {

			if fExit, _ := utils.FileExists(preCfg.ConfigFile); fExit {
				log.Fatalln("config file err: %s", err)
				return
			}

			return
		}
		log.Fatalln("config file err: %s", err)
		return
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
	//log.SetLevel(log.TraceLevel)
	//log.Debug("fileCfg: ", *fileCfg)

	config.ActiveNet = utils.GetNetParams(fileCfg.Network)


	//if fileCfg.Create{
	//	CreatWallet()
	//	return
	//}
	//
	//InitWallet()

	return

	//save
	// buf := new(bytes.Buffer)
	// if err := toml.NewEncoder(buf).Encode(*fileCfg); err != nil {
	// 	log.Fatal(err)
	// }

	//return ioutil.WriteFile(fileCfg.configFile, buf.Bytes(), 0666)
}

var createWalletCmd = &cobra.Command{
	Use:"create",
	Short:"create",
	Example:"create wallte",
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		CreatWallet()
	},
}
var createNewAccountCmd=&cobra.Command{
	Use:"createnewaccount {account} {pripassword}",
	Short:"create new account",
	Example:"createnewaccount test password",
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		UnLock(args[1])
		createNewAccount(args[0])
	},
}

var getbalanceCmd=&cobra.Command{
	Use:"getbalance {address}",
	Short:"getbalance",
	Example:`
		getbalance TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		getbalance(Default_minconf,args[0])
	},
}
var sendToAddressCmd=&cobra.Command{
	Use:"sendtoaddress {address} {amount} ",
	Short:"send transaction ",
	Example:`
		sendtoaddress TmWMuY9q5dUutUTGikhqTVKrnDMG34dEgb5 10
		`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		f32,err := strconv.ParseFloat(args[1],32)
		if(err!=nil){
			log.Fatal("getAccountAndAddress err :",err.Error())
			return
		}
		sendToAddress(args[0],float64(f32))
	},
}
var getAddressesByAccountCmd=&cobra.Command{
	Use:"getAddressesByAccount {string ,account,defalut imported} ",
	Short:"get addresses by account ",
	Example:`
		getAddressesByAccount imported
		`,
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		account :="imported"
		if len(args)>0{
			account=args[0]
		}
		getAddressesByAccount(account)
	},
}
var importPriKeyCmd=&cobra.Command{
	Use:"importprivkey {prikey} {pripassword}",
	Short:"import prikey ",
	Example:`
		importprivkey prikey pripassword
		`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		UnLock(args[1])
		importPrivKey(args[0])
	},
}
var listAccountsBalanceCmd=&cobra.Command{
	Use:"listaccountsbalance ",
	Short:"list Accounts Balance",
	Example:`
		listaccountsbalance
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		listAccountsBalance(Default_minconf)
	},
}
var getlisttxbyaddrCmd=&cobra.Command{
	Use:"getlisttxbyaddr {address}",
	Short:"get all transactions for address",
	Example:`
		getlisttxbyaddr Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ
		`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		getlisttxbyaddr(args[0])
	},
}
var updateblockCmd=&cobra.Command{
	Use:"updateblock {int,Update to the specified block, 0 is updated to the latest by default,defalut 0}",
	Short:"Update local block data",
	Example:`
		updateblock
		updateblock 12
		`,
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		var height = int64(0)
		if(len(args)>0){
			var err error
			height, err = strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Println("Argument is not of type int")
				return
			}
		}
		updateblock(height)
	},
}
var syncheightCmd=&cobra.Command{
	Use:"syncheight",
	Short:"Get the number of local synchronization blocks",
	Example:`
		syncheight
		`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		OpenWallet()
		syncheight()
	},
}

func init()  {
	Command.AddCommand(createWalletCmd)
	Command.AddCommand(createNewAccountCmd)
	Command.AddCommand(getbalanceCmd)
	Command.AddCommand(getlisttxbyaddrCmd)
	Command.AddCommand(updateblockCmd)
	Command.AddCommand(syncheightCmd)
	Command.AddCommand(sendToAddressCmd)
	Command.AddCommand(importPriKeyCmd)
	Command.AddCommand(getAddressesByAccountCmd)
	Command.AddCommand(listAccountsBalanceCmd)
}