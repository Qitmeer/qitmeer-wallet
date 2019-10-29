// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Qitmeer/qitmeer/params"

	"github.com/Qitmeer/qitmeer-wallet/utils"
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
	activeParams *params.Params
)

// Config wallet config
type Config struct {
	ConfigFile string
	AppDataDir string
	DebugLevel string
	LogDir     string
	Create bool

	Network string

	//WalletRPC
	UI            bool
	Listeners     []string
	RPCUser       string
	RPCPass       string
	RPCCert       string
	RPCKey        string
	RPCMaxClients int64
	DisableRPC    bool
	DisableTLS    bool

	//walletAPI
	APIs []string

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

	WalletPass string `short:"w" long:"wp" description:"Path to configuration file"`

	// //qitmeerd RPC config
	// QitmeerdSelect string // QitmeerdList[QitmeerdSelect]
	// QitmeerdList   map[string]*client.Config
}

var Cfg =NewDefaultConfig()
var ActiveNet = &params.MainNetParams
var once sync.Once

// Check config rule
func (cfg *Config) Check() error {

	activeNetParams := utils.GetNetParams(cfg.Network)
	if activeNetParams == nil {
		return fmt.Errorf("network not found: %s", cfg.Network)
	}

	return nil
}

// NewDefaultConfig make config by default value
func NewDefaultConfig() (cfg *Config) {
	cfg = &Config{
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
		UI:true,
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