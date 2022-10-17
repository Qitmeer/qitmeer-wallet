package testutils

import (
	"context"
	"github.com/Qitmeer/qitmeer-wallet/config"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/params"
	"github.com/shopspring/decimal"
	"testing"
	"time"
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
	mnemonic     = "dune school cash fancy post theory sense again earth divide balcony always"
	path         = "m/44'/60'/0'/0/0"
)

func TestCreateWallet(t *testing.T) {
	if err := clearWallet(walletCfg, activeParams); err != nil {
		t.Errorf("faild to clear wallet, %s", err.Error())
		return
	}
	w, err := createWallet(walletCfg, activeParams, pass, mnemonic, "")
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
	if addrs[0].String() != "RmGQG7xfaEm1dRi9Grb31NMT8kvsW1EAd4B" {
		t.Errorf("test failed, expect %v , but got %v", "RmGQG7xfaEm1dRi9Grb31NMT8kvsW1EAd4B", addrs[0].String())
	}
	clearWallet(walletCfg, activeParams)
	w, err = createWallet(walletCfg, activeParams, pass, mnemonic, path)
	if err != nil {
		t.Errorf("failed to create wallet, %s", err)
		return
	}
	account, err = w.AccountNumber(waddrmgr.KeyScopeBIP0044, "imported")
	if err != nil {
		t.Errorf("failed to get account number, %s", err)
		return
	}
	addrs, err = w.AccountAddresses(account)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	for _, addr := range addrs {
		t.Logf("account address = %s", addr)
	}

	addrs1, err := w.AccountEVMAddresses(account)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	if addrs1[0].String() != "0x2cb3aD95bE524F9d34E17Da37a901F63fa12Ba35" {
		t.Errorf("test failed, expect %v , but got %v", "0x2cb3aD95bE524F9d34E17Da37a901F63fa12Ba35", addrs1[0].String())
	}
}

func TestExportAmountToEvm(t *testing.T) {
	args := []string{"--modules=miner", "--modules=qitmeer", "--notls"}
	h, err := NewHarnessWithMnemonic(t, mnemonic, path, params.PrivNetParam.Params, args...)
	defer h.Teardown()

	if err != nil {
		t.Errorf("new harness failed: %v", err)
		return
	}
	err = h.Setup()
	if err != nil {
		t.Errorf("setup harness failed:%v", err)
		return
	}

	h.WaitWalletInit()
	if info, err := h.Client.NodeInfo(); err != nil {
		t.Errorf("test failed : %v", err)
		return
	} else {
		expect := "privnet"
		if info.Network != expect {
			t.Errorf("test failed, expect %v , but got %v", expect, info.Network)
			return
		}

	}
	GenerateBlock(t, h, 20)
	time.Sleep(10 * time.Second)

	GenerateBlock(t, h, 1)
	time.Sleep(5 * time.Second)
	b, err := h.wallet.Balance(types.MEERA)
	if err != nil {
		t.Errorf("test failed:%v", err)
		return
	}
	if b.UnspentAmount.Value != 250000000000 {
		t.Errorf("test failed, expect balance %d, but got %d", 250000000000, b.UnspentAmount.Value)
		return
	}
	account, err := h.wallet.wallet.AccountNumber(waddrmgr.KeyScopeBIP0044, "imported")
	if err != nil {
		t.Errorf("failed to get account number, %s", err)
		return
	}
	addrs, err := h.wallet.wallet.AccountAddresses(account)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	//
	_, err = h.wallet.SendToAddress(addrs[1].String(), types.MEERB, 500)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	GenerateBlock(t, h, 1)
	addrs1, err := h.wallet.wallet.AccountEVMAddresses(account)
	if err != nil {
		t.Errorf("failed to get account address, %s", err)
		return
	}
	ba, _ := h.evmClient.BalanceAt(context.Background(), addrs1[0], nil)
	baD := decimal.NewFromBigInt(ba, 0)
	baD = baD.Div(decimal.NewFromFloat(1e18))
	if baD.Cmp(decimal.NewFromFloat(500)) != 0 {
		t.Errorf("failed to get account balance expect 500 , but got %s ", baD.String())
	}
}
