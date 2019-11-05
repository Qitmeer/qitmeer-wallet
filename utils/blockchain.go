package utils

import (
	"github.com/Qitmeer/qitmeer/params"
)

// GetNetParams by network name
func GetNetParams(name string) *params.Params {
	switch name {
	case "mainnet":
		return &params.MainNetParams
	case "testnet":
		return &params.TestNetParams
	case "privnet":
		return &params.PrivNetParams
	default:
		return &params.TestNetParams
	}
}
