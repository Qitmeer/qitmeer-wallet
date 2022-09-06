package testutils

import (
	"github.com/Qitmeer/qitmeer-wallet/config"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qng/params"
	"testing"
)

var walletCfg = &config.Config{
	Network:        "privnet",
	UI:             false,
	Listeners:      nil,
	RPCCert:        "",
	RPCKey:         "",
	RPCMaxClients:  0,
	DisableRPC:     false,
	DisableTLS:     true,
	Confirmations:  10,
	QServer:        "127.0.0.1:38131",
	QUser:          "testuser",
	QPass:          "testpass",
	QCert:          "",
	QNoTLS:         true,
	QTLSSkipVerify: false,
	WalletPass:     "111111",
}

var (
	activeParams = &params.PrivNetParams
	pass         = "111111"
	private      = "33f9e4e405c054fb267c4d0717afa2978376486b560415b0abb5a015db36da1e"
	public       = "02e177d3e179d31b7df4986656bee887700a401fb1edf20b12fbddb8cf0f81ab88"
)

func TestCreateWallet(t *testing.T) {
	if err := clearWallet(walletCfg, activeParams); err != nil {
		t.Errorf("faild to clear wallet, %s", err.Error())
		return
	}
	w, err := createWallet(walletCfg, activeParams, pass)
	if err != nil {
		t.Errorf("failed to create wallet, %s", err)
		return
	}
	account, err := w.AccountNumber(waddrmgr.KeyScopeBIP0044, "imported")
	if err != nil {
		t.Errorf("failed to get account number, %s", err)
		return
	}
	addrs, err := w.AccountAddresses(account)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	for _, addr := range addrs {
		t.Logf("account address = %s", addr)
	}
}
