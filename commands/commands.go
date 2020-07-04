package commands

import (
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "qitmeer-wallet",
}

type JsonCmdHelper struct {
	JsonCmd interface{}
	Run     func(interface{}, *wallet.Wallet) (interface{}, error)
}

func (h *JsonCmdHelper) Call() (interface{}, error) {
	var err error
	var res interface{}
	if res, err = h.Run(h.JsonCmd, w); err == nil {
		if res, err = json.MarshalIndent(res, "", " "); err == nil {
			fmt.Printf("%s\n", res)
			return res, nil
		}
	}
	return nil, err
}

var helper *JsonCmdHelper

var preCfg *config.Config
var fileCfg = config.Cfg

func bindFlags() {
	preCfg = &config.Config{}
	RootCmd.PersistentFlags().StringVarP(&preCfg.ConfigFile, "configfile", "c", "config.toml", "config file")
	RootCmd.PersistentFlags().StringVarP(&preCfg.DebugLevel, "debuglevel", "d", "info", "Logging level {trace, debug, info, warn, error, critical}")
	RootCmd.PersistentFlags().StringVarP(&preCfg.AppDataDir, "appdatadir", "a", "", "wallet db path")
	RootCmd.PersistentFlags().StringVarP(&preCfg.LogDir, "logdir", "l", "", "log data path")
	RootCmd.PersistentFlags().StringVarP(&preCfg.Network, "network", "n", "testnet", "network")
	RootCmd.PersistentFlags().BoolVar(&preCfg.Create, "create", false, "Create a new wallet")
	RootCmd.PersistentFlags().StringVarP(&preCfg.QServer, "qserver", "s", "127.0.0.1:8030", "qitmeer node server")
	RootCmd.PersistentFlags().StringVarP(&preCfg.QUser, "quser", "u", "admin", "qitmeer node user")
	RootCmd.PersistentFlags().StringVarP(&preCfg.QPass, "qpass", "p", "123456", "qitmeer node password")
	RootCmd.PersistentFlags().StringVarP(&preCfg.WalletPass, "pubwalletpass", "P", "public", "data encryption password")
	RootCmd.PersistentFlags().BoolVar(&preCfg.QNoTLS, "qnotls", fileCfg.QNoTLS, "disable TLS")
	RootCmd.PersistentFlags().StringVar(&preCfg.QCert, "qcert", fileCfg.QCert, "Certificate path")
	RootCmd.PersistentFlags().BoolVar(&preCfg.QTLSSkipVerify, "qtlsskipverify", fileCfg.QTLSSkipVerify, "skip TLS verification")

	RootCmd.PersistentFlags().Int64Var(&preCfg.Confirmations, "confirmations", 10, "Number of block confirmations ")
	RootCmd.PersistentFlags().BoolVar(&preCfg.UI, "ui", true, "Start Wallet with RPC and webUI interface")
	RootCmd.PersistentFlags().StringArrayVar(&preCfg.Listeners, "listeners", fileCfg.Listeners, "rpc listens")
	RootCmd.PersistentFlags().StringVar(&preCfg.RPCUser, "rpcUser", fileCfg.RPCUser, "rpc user,default by random")
	RootCmd.PersistentFlags().StringVar(&preCfg.RPCPass, "rpcPass", fileCfg.RPCPass, "rpc pass,default by random")
	RootCmd.PersistentFlags().Int64Var(&preCfg.MinTxFee, "mintxfee", fileCfg.MinTxFee, "The minimum transaction fee in QIT/kB default 20000 (aka. 0.0002 MEER/KB)")
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

func init() {
	bindFlags()

	RootCmd.AddCommand(QcCmd)
	RootCmd.AddCommand(QxCmd)
	RootCmd.AddCommand(WebCmd)
	RootCmd.AddCommand(ConsoleCmd)

	AddQcCommand()
	AddQxCommand()
	AddWebCommand()
	AddConsoleCommand()
}
