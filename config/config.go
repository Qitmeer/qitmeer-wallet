// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/HalalChain/qitmeer-lib/params"

	"github.com/HalalChain/qitmeer-wallet/utils"
)

const (
	defaultCAFilename       = "qit.cert"
	defaultConfigFilename   = "wallet.toml"
	defaultLogLevel         = "info"
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

	Network string // mainnet testnet simnet

	//WalletRPC
	Listeners     []string // ["127.0.0.1:18131"]
	RPCUser       string
	RPCPass       string
	RPCCert       string
	RPCKey        string
	RPCMaxClients int64
	DisableRPC    bool
	DisableTLS    bool

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

// LoadConfig load config from file
func LoadConfig(configFile string, isCreate bool) (cfg *Config, err error) {
	cfg = NewDefaultConfig()

	if isCreate {
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
		return cfg, nil
	}

	_, err = toml.DecodeFile(configFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("LoadConfig err: %s", err)
	}

	//check rules
	if !validLogLevel(cfg.DebugLevel) {
		return nil, fmt.Errorf("LoadConfig validLogLevel err: %s", cfg.DebugLevel)
	}

	cfg.ConfigFile = configFile

	return
}

// NewDefaultConfig make config by default value
func NewDefaultConfig() (cfg *Config) {
	cfg = &Config{
		ConfigFile: DefaultConfigFile,
		AppDataDir: defaultAppDataDir,
		DebugLevel: defaultLogLevel,
		LogDir:     defaultLogDir,

		Network: "mainnet",

		Listeners:     []string{"127.0.0.1:18131"},
		RPCUser:       "",
		RPCPass:       "",
		RPCCert:       defaultRPCCertFile,
		RPCKey:        defaultRPCKeyFile,
		RPCMaxClients: defaultRPCMaxClients,
		DisableRPC:    false,
		DisableTLS:    false,

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

// validLogLevel returns whether or not logLevel is a valid debug log level.
func validLogLevel(logLevel string) bool {
	switch logLevel {
	case "trace":
		fallthrough
	case "debug":
		fallthrough
	case "info":
		fallthrough
	case "warn":
		fallthrough
	case "error":
		fallthrough
	case "critical":
		return true
	}
	return false
}
