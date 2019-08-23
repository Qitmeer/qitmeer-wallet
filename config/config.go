// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"log"
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"

	"github.com/HalalChain/qitmeer-lib/params"

	"github.com/HalalChain/qitmeer-wallet/utils"
)

const (
	defaultCAFilename       = "qit.cert"
	defaultConfigFilename   = "config.toml"
	defaultLogLevel         = "debug"
	defaultLogDirname       = "logs"
	defaultLogFilename      = "wallet.log"
	defaultRPCMaxClients    = 10
	defaultRPCMaxWebsockets = 25

	WalletDbName = "wallet.db"
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

	WalletPass string

	// //qitmeerd RPC config
	// QitmeerdSelect string // QitmeerdList[QitmeerdSelect]
	// QitmeerdList   map[string]*client.Config
}

var Cfg *Config
var ActiveNet = &params.MainNetParams
var once sync.Once

func init(){
	once.Do(func() {
		fmt.Println("执行init------------------------")
		Cfg=NewDefaultConfig()
		_, err := toml.DecodeFile("config.toml", &Cfg)
		if err != nil {
			log.Println(err)
		}
		if Cfg.AppDataDir ==""{
			appData:=cleanAndExpandPath(Cfg.AppDataDir)
			log.Println("appData：",appData)
			Cfg.AppDataDir=appData
		}
		ActiveNet = utils.GetNetParams(Cfg.Network)
	})
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
		QServer:        "127.0.0.1:18130",
		QUser:          "",
		QPass:          "",
		QCert:          "",
		QNoTLS:         true,
		QTLSSkipVerify: true,
		QProxy:         "",
		QProxyUser:     "",
		QProxyPass:     "",
		WalletPass:     "public",
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
// cleanAndExpandPath expands environement variables and leading ~ in the
// passed path, cleans the result, and returns it.
func cleanAndExpandPath(path string) string {
	// NOTE: The os.ExpandEnv doesn't work with Windows cmd.exe-style
	// %VARIABLE%, but they variables can still be expanded via POSIX-style
	// $VARIABLE.
	path = os.ExpandEnv(path)

	if !strings.HasPrefix(path, "~") {
		return filepath.Clean(path)
	}

	// Expand initial ~ to the current user's home directory, or ~otheruser
	// to otheruser's home directory.  On Windows, both forward and backward
	// slashes can be used.
	path = path[1:]

	var pathSeparators string
	if runtime.GOOS == "windows" {
		pathSeparators = string(os.PathSeparator) + "/"
	} else {
		pathSeparators = string(os.PathSeparator)
	}

	userName := ""
	if i := strings.IndexAny(path, pathSeparators); i != -1 {
		userName = path[:i]
		path = path[i:]
	}

	homeDir := ""
	var u *user.User
	var err error
	if userName == "" {
		u, err = user.Current()
	} else {
		u, err = user.Lookup(userName)
	}
	if err == nil {
		homeDir = u.HomeDir
	}
	// Fallback to CWD if user lookup fails or user has no home directory.
	if homeDir == "" {
		homeDir = "."
	}

	return filepath.Join(homeDir, path)
}