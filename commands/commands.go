package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
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

var userConf = config.Cfg
var defaultConf = config.NewDefaultConfig()

func checkDefaultConf() error {
	dcf := config.DefaultConfigFile
	if exists, _ := utils.FileExists(dcf); !exists {
		fmt.Printf("Required to create default config %s, continue? [y/n]\n", dcf)

		reader := bufio.NewReader(os.Stdin)
		if answer, err := reader.ReadByte(); (answer == 'y' || answer == 'Y') && err == nil {
			if _, err := utils.FileCopy(config.SampleConfigFile, dcf); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return err
			}
		} else {
			err := errors.New(fmt.Sprintf("Denied to create %s", dcf))
			fmt.Fprintln(os.Stderr, err)
			return err
		}
	}

	return nil
}

func bindFlags() error {
	cobra.OnInitialize(initConfig)

	dcf := config.DefaultConfigFile
	if err := checkDefaultConf(); err != nil {
		return err
	}

	viper.SetConfigFile(dcf)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	uc := userConf
	viper.Unmarshal(uc)

	pf := RootCmd.PersistentFlags()

	pf.StringVarP(&uc.ConfigFile, "configfile", "c", uc.ConfigFile, "config file")
	pf.StringP("appdatadir", "a", uc.AppDataDir, "wallet db path")
	pf.StringP("debuglevel", "d", uc.DebugLevel, "Logging level {trace, debug, info, warn, error, critical}")
	pf.StringP("logdir", "l", uc.LogDir, "log data path")
	pf.Bool("create", uc.Create, "Create a new wallet")
	pf.StringP("network", "n", uc.Network, "network")

	pf.Bool("ui", uc.UI, "Start Wallet with RPC and webUI interface")
	pf.StringArray("listeners", uc.Listeners, "rpc listens")
	pf.String("rpcuser", uc.RPCUser, "RPC username,default by random")
	pf.String("rpcpass", uc.RPCPass, "RPC password,default by random")
	pf.String("rpccert", uc.RPCCert, "RPC certificate file")
	pf.String("rpckey", uc.RPCKey, "RPC certificate key file")
	pf.Int64("rpcmaxclients", uc.RPCMaxClients, "RPC max clients number")
	pf.Bool("disablerpc", uc.DisableRPC, "disable RPC server")
	pf.Bool("disabletls", uc.DisableTLS, "disable TLS for the RPC server")

	pf.Int64("confirmations", uc.Confirmations, "Number of block confirmations ")
	pf.Int64("mintxfee", uc.MinTxFee, "The minimum transaction fee in QIT/kB default 20000 (aka. 0.0002 MEER/KB)")
	pf.StringArray("apis", uc.APIs, "enabled APIs")

	pf.StringP("qserver", "s", uc.QServer, "qitmeer node server, overwritten by qitmeerdselect")
	pf.StringP("quser", "u", uc.QUser, "qitmeer node username")
	pf.StringP("qpass", "p", uc.QPass, "qitmeer node password")
	pf.String("qcert", uc.QCert, "Certificate path")
	pf.Bool("qnotls", uc.QNoTLS, "disable TLS")
	pf.Bool("qtlsskipverify", uc.QTLSSkipVerify, "skip TLS verification")
	pf.String("qproxy", uc.QProxy, "qitmeer node proxy address")
	pf.String("qproxyuser", uc.QProxyUser, "qitmeer node proxy username")
	pf.String("qproxypass", uc.QProxyPass, "qitmeer node proxy password")
	pf.StringP("walletpass", "P", uc.WalletPass, "data encryption password")
	pf.String("qitmeerdselect", uc.QitmeerdSelect,
		"select qitmeer RPC config defined in Qitmeerds section of config file, overwrite qserver")

	dc := defaultConf
	viper.SetDefault("ConfigFile", dc.ConfigFile)
	viper.SetDefault("AppDataDir", dc.AppDataDir)
	viper.SetDefault("DebugLevel", dc.DebugLevel)
	viper.SetDefault("LogDir", dc.LogDir)
	viper.SetDefault("Create", dc.Create)
	viper.SetDefault("Network", dc.Network)
	viper.SetDefault("UI", dc.UI)
	viper.SetDefault("Listeners", dc.Listeners)
	viper.SetDefault("RPCUser", dc.RPCUser)
	viper.SetDefault("RPCPass", dc.RPCPass)
	viper.SetDefault("RPCKey", dc.RPCKey)
	viper.SetDefault("RPCMaxClients", dc.RPCMaxClients)
	viper.SetDefault("DisableRPC", dc.DisableRPC)
	viper.SetDefault("DisableTLS", dc.DisableTLS)
	viper.SetDefault("Confirmations", dc.Confirmations)
	viper.SetDefault("MinTxFee", dc.MinTxFee)
	viper.SetDefault("APIs", dc.APIs)
	viper.SetDefault("QServer", dc.QServer)
	viper.SetDefault("QUser", dc.QUser)
	viper.SetDefault("QPass", dc.QPass)
	viper.SetDefault("QCert", dc.QCert)
	viper.SetDefault("QNoTLS", dc.QNoTLS)
	viper.SetDefault("QTLSSkipVerify", dc.QTLSSkipVerify)
	viper.SetDefault("QProxy", dc.QPass)
	viper.SetDefault("QProxyUser", dc.QProxyUser)
	viper.SetDefault("WalletPass", dc.WalletPass)
	viper.SetDefault("QitmeerdSelect", dc.QitmeerdSelect)
	viper.SetDefault("Qitmeerds", dc.Qitmeerds)

	viper.BindPFlag("ConfigFile", pf.Lookup("configfile"))
	viper.BindPFlag("AppDataDir", pf.Lookup("appdatadir"))
	viper.BindPFlag("DebugLevel", pf.Lookup("debuglevel"))
	viper.BindPFlag("LogDir", pf.Lookup("logdir"))
	viper.BindPFlag("Create", pf.Lookup("create"))
	viper.BindPFlag("Network", pf.Lookup("network"))

	viper.BindPFlag("UI", pf.Lookup("ui"))
	viper.BindPFlag("Listeners", pf.Lookup("listeners"))
	viper.BindPFlag("RPCUser", pf.Lookup("rpcuser"))
	viper.BindPFlag("RPCPass", pf.Lookup("rpcpass"))
	viper.BindPFlag("RPCCert", pf.Lookup("rpccert"))
	viper.BindPFlag("RPCKey", pf.Lookup("rpckey"))
	viper.BindPFlag("RPCMaxClients", pf.Lookup("rpcmaxclients"))
	viper.BindPFlag("DisableRPC", pf.Lookup("disablerpc"))
	viper.BindPFlag("DisableTLS", pf.Lookup("disabletls"))

	viper.BindPFlag("Confirmations", pf.Lookup("confirmations"))
	viper.BindPFlag("MinTxFee", pf.Lookup("mintxfee"))
	viper.BindPFlag("APIs", pf.Lookup("apis"))

	viper.BindPFlag("QServer", pf.Lookup("qserver"))
	viper.BindPFlag("QUser", pf.Lookup("quser"))
	viper.BindPFlag("QPass", pf.Lookup("qpass"))
	viper.BindPFlag("QCert", pf.Lookup("qcert"))
	viper.BindPFlag("QNoTLS", pf.Lookup("qnotls"))
	viper.BindPFlag("QTLSSkipVerify", pf.Lookup("qtlsskipverify"))
	viper.BindPFlag("QProxy", pf.Lookup("qproxy"))
	viper.BindPFlag("QProxyUser", pf.Lookup("qproxyuser"))
	viper.BindPFlag("QProxyPass", pf.Lookup("qproxypass"))
	viper.BindPFlag("WalletPass", pf.Lookup("walletpass"))
	viper.BindPFlag("QitmeerdSelect", pf.Lookup("qitmeerdselect"))
	return nil
}

func initConfig() {
	// save as default first
	if userConf.ConfigFile != defaultConf.ConfigFile {
		// Use config file from the flag.
		viper.SetConfigFile(userConf.ConfigFile)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Failed config file:", viper.ConfigFileUsed(), " error: ", err.Error())
		os.Exit(-1)
	}

	cfg := config.NewDefaultConfig()
	viper.Unmarshal(cfg)

	userConf = cfg

	config.ActiveNet = utils.GetNetParams(userConf.Network)
	// Parse, validate, and set debug log level(s).
	if lvl, err := log.LvlFromString(userConf.DebugLevel); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		Glogger().Verbosity(lvl)
	}

	InitLogRotator(filepath.Join(userConf.LogDir, "wallet.log"))

	if userConf.MinTxFee < config.DefaultMinRelayTxFee {
		userConf.MinTxFee = config.DefaultMinRelayTxFee
	}
}

func init() {
	if err := bindFlags(); err != nil {
		os.Exit(-1)
	}

	RootCmd.AddCommand(QcCmd)
	RootCmd.AddCommand(QxCmd)
	RootCmd.AddCommand(WebCmd)
	RootCmd.AddCommand(ConsoleCmd)

	AddQcCommand()
	AddQxCommand()
	AddWebCommand()
	AddConsoleCommand()
}
