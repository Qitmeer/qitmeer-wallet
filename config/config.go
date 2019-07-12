package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/HalalChain/qitmeer-wallet/utils"
)

// Config Qitmeer Wallet Config
type Config struct {
	Network string // blockchain network: main/test/sim

	DataDir   string
	KeysDir   string // default {datadir}/{network}/keys
	WalletDir string //default {datadir}/{network}/wallet

	Listen     string // 127.0.0.1:18130
	RPCUser    string
	RPCPass    string
	RPCCert    string
	RPCKey     string
	DisableTLS bool

	Apis []string
}

// Init check config
func (c *Config) Init() error {
	//check dir and make it

	return nil
}

// NewDefaultConfig make a config by default set
func NewDefaultConfig() (cfg *Config) {
	//make config file
	network := "test"
	dataDir := utils.GetUserDataDir()

	cfg = &Config{
		Network:   "test",
		DataDir:   dataDir,
		KeysDir:   filepath.Join(dataDir, network, "keys"),
		WalletDir: filepath.Join(dataDir, network, "wallet"),

		Listen:     "127.0.0.1:38130",
		RPCUser:    "",
		RPCPass:    "",
		RPCCert:    "",
		RPCKey:     "",
		DisableTLS: true,

		Apis: []string{"account", "tx"},
	}

	return
}

// Load load config from config file or make a new config file
func Load(configFile string) (cfg *Config, err error) {
	_, err = os.Stat(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = NewDefaultConfig()

			//save
			buf := new(bytes.Buffer)
			if err := toml.NewEncoder(buf).Encode(cfg); err != nil {
				return nil, fmt.Errorf("config load Encode err: %s", err)
			}

			err = utils.MakeDirAll(filepath.Dir(configFile))
			if err != nil {
				return nil, fmt.Errorf("Load configFile: mkDir err: %s", err)
			}

			err = ioutil.WriteFile(configFile, buf.Bytes(), 0666)

			return cfg, err
		}
		return nil, fmt.Errorf("Load configFile: stat err: %s", err)
	}

	cfg = &Config{}
	_, err = toml.DecodeFile(configFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("Load configFile: DecodeFile err: %s", err)
	}

	return cfg, nil
}
