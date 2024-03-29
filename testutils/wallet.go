package testutils

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/internal/legacy/keystore"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer-wallet/wallet/txrules"
	"github.com/Qitmeer/qng/core/protocol"
	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/crypto/bip32"
	"github.com/Qitmeer/qng/crypto/bip39"
	"github.com/Qitmeer/qng/crypto/ecc/secp256k1"
	chaincfg "github.com/Qitmeer/qng/params"
	"github.com/Qitmeer/qng/qx"
	wallet2 "github.com/Qitmeer/qng/wallet"
	"os"
	"path/filepath"
	"time"
)

type Wallet struct {
	wallet  *wallet.Wallet
	address string
}

func NewWallet(cfg *config.Config, net protocol.Network, mnemonic, path string) (*Wallet, error) {
	activParams := chaincfg.PrivNetParams
	switch net {
	case protocol.MainNet:
		activParams = chaincfg.MainNetParams
	case protocol.MixNet:
		activParams = chaincfg.MixNetParams
	case protocol.TestNet:
		activParams = chaincfg.TestNetParams
	case protocol.PrivNet:
		activParams = chaincfg.PrivNetParams
	default:
		return nil, fmt.Errorf("unknown network type %v", net)
	}
	if err := clearWallet(cfg, &activParams); err != nil {
		return nil, err
	}
	w, err := createWallet(cfg, &activParams, cfg.WalletPass, mnemonic, path)
	if err != nil {
		return nil, err
	}
	return &Wallet{wallet: w}, nil
}

func (w *Wallet) Start() error {
	w.wallet.Start()
	return nil
}

func (w *Wallet) Stop() {
	w.wallet.Stop()
}

func (w *Wallet) GenerateAddress(usePkAddr bool) (string, error) {
	account, err := w.wallet.AccountNumber(waddrmgr.KeyScopeBIP0044, "imported")
	if err != nil {
		return "", err
	}
	addrs, err := w.wallet.AccountAddresses(account)
	if err != nil {
		return "", err
	}
	if len(addrs) != 2 {
		return "", fmt.Errorf("wrong address")
	}
	w.address = addrs[0].String()
	if usePkAddr {
		w.address = addrs[1].String()
	}
	return w.address, nil
}

func (w *Wallet) Balance(coin types.CoinID) (*wallet.Balance, error) {
	balance, err := w.wallet.GetBalance(w.address)
	if err != nil {
		return nil, nil
	}
	if b, ok := balance[coin]; ok {
		return &b, nil
	} else {
		return nil, fmt.Errorf("no coin %s", coin.Name())
	}
}

func (w *Wallet) BalanceByAddr(coin types.CoinID, addr string) (*wallet.Balance, error) {
	balance, err := w.wallet.GetBalance(addr)
	if err != nil {
		return nil, nil
	}
	if b, ok := balance[coin]; ok {
		return &b, nil
	} else {
		return nil, fmt.Errorf("no coin %s", coin.Name())
	}
}

func (w *Wallet) SendToAddress(addr string, coin types.CoinID, amount uint64) (string, error) {
	// Check that signed integer parameters are positive.
	if amount < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	coinId, err := w.wallet.CoinID(coin)
	if err != nil {
		return "", err
	}
	amt := types.Amount{Value: int64(amount * types.AtomsPerCoin), Id: coinId}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addr: amt,
	}

	return w.wallet.SendPairs(pairs, waddrmgr.AccountMergePayNum, txrules.DefaultRelayFeePerKb, 0, "")
}

func (w *Wallet) EvmToAddress(addr string, coin types.CoinID, amount uint64) (string, error) {
	// Check that signed integer parameters are positive.
	if amount < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	coinId, err := w.wallet.CoinID(coin)
	if err != nil {
		return "", err
	}
	amt := types.Amount{Value: int64(amount * types.AtomsPerCoin), Id: coinId}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addr: amt,
	}

	return w.wallet.EVMToUTXO(pairs, waddrmgr.AccountMergePayNum, txrules.DefaultRelayFeePerKb, 0, "")
}

func networkDir(dataDir string, chainParams *chaincfg.Params) string {
	netname := chainParams.Name

	return filepath.Join(dataDir, netname)
}

func createWallet(cfg *config.Config, params *chaincfg.Params, pass, mnemonic, path string) (*wallet.Wallet, error) {
	dbDir := networkDir(cfg.AppDataDir, params)
	loader := wallet.NewLoader(params, dbDir, 250, &config.Config{})

	keystorePath := filepath.Join(dbDir, keystore.Filename)
	var legacyKeyStore *keystore.Store
	_, err := os.Stat(keystorePath)
	if err != nil && !os.IsNotExist(err) {
		// A stat error not due to a non-existant file should be
		// returned to the caller.
		return nil, err
	} else if err == nil {
		// Keystore file exists.
		legacyKeyStore, err = keystore.OpenDir(dbDir)
		if err != nil {
			return nil, err
		}
	}

	privPass := []byte(pass)

	if legacyKeyStore != nil {
		err = legacyKeyStore.Unlock(privPass)
		if err != nil {
			return nil, err
		}

	}

	seed, err := bip32.NewSeed()
	if err != nil {
		return nil, err
	}
	if mnemonic != "" {
		seed, err = bip39.NewSeedWithErrorChecking(mnemonic, "")
		if err != nil {
			return nil, err
		}
	}
	seedKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	w, err := loader.CreateNewWallet(privPass, privPass, seed, time.Now())
	if err != nil {
		return nil, err
	}
	var ck = seedKey
	if path != "" {
		derivePath := qx.DerivePathFlag{Path: wallet2.DerivationPath{}}
		derivePath.Set(path)
		for _, i := range derivePath.Path {
			ck, err = ck.NewChildKey(i)
			if err != nil {
				return nil, err
			}
		}
	}
	pri, _ := secp256k1.PrivKeyFromBytes(ck.Key)
	wif, err := utils.NewWIF(pri, w.ChainParams(), true)
	if err != nil {
		return nil, err
	}
	if !wif.IsForNet(w.ChainParams()) {
		return nil, err
	}

	w.UnLockManager(privPass)
	_, err = w.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	if err != nil {
		return nil, err
	}

	w.SetConfig(cfg)
	w.HttpClient, err = wallet.NewHtpc(cfg)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func clearWallet(cfg *config.Config, params *chaincfg.Params) error {
	dbDir := networkDir(cfg.AppDataDir, params)
	return os.RemoveAll(dbDir)
}

func newWalletConfig(homeDir string) *config.Config {
	var walletCfg = &config.Config{
		Network:        "privnet",
		DisableTLS:     true,
		Confirmations:  10,
		QServer:        "127.0.0.1:38131",
		QUser:          "testuser",
		QPass:          "testpass",
		QCert:          "",
		QNoTLS:         true,
		QTLSSkipVerify: false,
		WalletPass:     "111111",
		AppDataDir:     homeDir,
	}
	return walletCfg
}
