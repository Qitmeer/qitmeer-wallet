package config

import (
	"testing"

	"github.com/Qitmeer/qitmeer-wallet/rpc/client"
)

func TestSave(t *testing.T) {

	cfg := NewDefaultConfig()

	cfg.QitmeerdList = make(map[string]*client.Config)
	cfg.QitmeerdList["local"] = &client.Config{
		RPCServer: "2.2.2.2:8080",
	}
	cfg.QitmeerdSelect = "local"

	t.Log(cfg.Save("config.toml"))
}
