package config

import "testing"

func TestLoad(t *testing.T) {
	cfg, err := Load("config.toml")
	t.Log(err)
	t.Log(cfg)
}
