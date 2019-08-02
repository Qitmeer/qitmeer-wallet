// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/HalalChain/qitmeer-lib/params"

	"github.com/HalalChain/qitmeer-wallet/utils"
)

const (
	defaultCAFilename       = "qit.cert"
	defaultConfigFilename   = "wallet.toml"
	defaultLogLevel         = "debug"
	defaultLogDirname       = "logs"
	defaultLogFilename      = "wallet.log"
	defaultRPCMaxClients    = 10
	defaultRPCMaxWebsockets = 25

	walletDbName = "wallet.db"
)

var (
	defaultAppDataDir  = utils.AppDataDir("qitwallet", false)
	DefaultConfigFile  = filepath.Join(defaultAppDataDir, defaultConfigFilename)
	defaultRPCKeyFile  = filepath.Join(defaultAppDataDir, "rpc.key")
	defaultRPCCertFile = filepath.Join(defaultAppDataDir, "rpc.cert")
	defaultLogDir      = filepath.Join(defaultAppDataDir, defaultLogDirname)
)

var (
	InsecurePubPassphrase = "public"

	activeParams *params.Params
)

// Config wallet config
type Config struct {
	ConfigFile string
	AppDataDir string
	DebugLevel string
	LogDir     string

	Network string // mainnet testnet privnet

	//WalletRPC
	UI            bool     // local web server UI
	Listeners     []string // ["127.0.0.1:38130"]
	RPCUser       string
	RPCPass       string
	RPCCert       string
	RPCKey        string
	RPCMaxClients int64
	DisableRPC    bool
	DisableTLS    bool

	//walletAPI
	APIs []string // rpc support api list

	//Qitmeerd
	isLocal        bool
	QServer        string
	QUser          string
	QPass          string
	QCert          string
	QNoTLS         bool
	QTLSSkipVerify bool
	QProxy         string
	QProxyUser     string
	QProxyPass     string

	// //qitmeerd RPC config
	// QitmeerdSelect string // QitmeerdList[QitmeerdSelect]
	// QitmeerdList   map[string]*client.Config
}

// Check config rule
func (cfg *Config) Check() error {

	activeNetParams := utils.GetNetParams(cfg.Network)
	if activeNetParams == nil {
		return fmt.Errorf("network not found: %s", cfg.Network)
	}

	return nil
}

// LoadConfig load config from file
func LoadConfig(configFile string, isCreate bool, preCfg *Config) (cfg *Config, err error) {

	if isCreate {
		cfg = NewDefaultConfig()

		//save default
		buf := new(bytes.Buffer)
		if err = toml.NewEncoder(buf).Encode(cfg); err != nil {

			return nil, fmt.Errorf("LoadConfig err: %s", err)
		}
		err = ioutil.WriteFile(configFile, buf.Bytes(), 0666)
		if err != nil {
			return nil, fmt.Errorf("LoadConfig err: %s", err)
		}
		return
	}

	var fileExist bool
	fileExist, err = utils.FileExists(configFile)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig err: %s", err)
	}
	if !fileExist && configFile == DefaultConfigFile {
		return preCfg, nil
	}

	_, err = toml.DecodeFile(configFile, preCfg)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig err: %s", err)
	}

	preCfg.ConfigFile = configFile

	return preCfg, nil
}

// NewDefaultConfig make config by default value
func NewDefaultConfig() (cfg *Config) {
	cfg = &Config{
		ConfigFile: DefaultConfigFile,
		AppDataDir: defaultAppDataDir,
		DebugLevel: defaultLogLevel,
		LogDir:     defaultLogDir,

		Network: "mainnet",

		Listeners:     []string{"127.0.0.1:38130"},
		RPCUser:       randStr(8),
		RPCPass:       randStr(24),
		RPCCert:       defaultRPCCertFile,
		RPCKey:        defaultRPCKeyFile,
		RPCMaxClients: defaultRPCMaxClients,
		DisableRPC:    false,
		DisableTLS:    false,

		APIs: []string{"account", "wallet"},

		isLocal:        true,
		QServer:        "127.0.0.1",
		QUser:          "18130",
		QPass:          "",
		QCert:          "",
		QNoTLS:         false,
		QTLSSkipVerify: true,
		QProxy:         "",
		QProxyUser:     "",
		QProxyPass:     "",
	}
	return
}

func randStr(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
