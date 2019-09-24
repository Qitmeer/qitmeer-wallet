package wserver

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Qitmeer/qitmeer-lib/crypto/bip32"
	"github.com/Qitmeer/qitmeer-lib/crypto/bip39"
	"github.com/Qitmeer/qitmeer-lib/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer-lib/crypto/seed"

	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
)

// API wallet
type API struct {
	cfg  *config.Config
	wSvr *WalletServer
}

// NewAPI api make
func NewAPI(cfg *config.Config, wSvr *WalletServer) *API {
	return &API{
		cfg:  cfg,
		wSvr: wSvr,
	}
}

// Status wallet info
func (api *API) Status() (status *ResStatus, err error) {
	status = &ResStatus{}

	wtExist, err := api.wSvr.WtLoader.WalletExists()
	if err != nil {
		log.Warnf("api Status WalletExists err: %s", err)
		return nil, fmt.Errorf("check wallet exist err: %s", err)
	}

	if !wtExist {
		status.Stats = "nil"
		return
	}

	if api.wSvr.Wt == nil {
		status.Stats = "closed"
		return
	}

	// if api.wSvr.Wt.Locked() {
	// 	status.Stats = "lock"
	// } else {
	status.Stats = "unlock"
	// }
	log.Debug("wallet api: status", status)
	return
}

//Create wallet by seed
func (api *API) Create(seed string, walletPass string) error {
	seedBuf, err := hex.DecodeString(seed)
	if err != nil {
		return &crateError{Code: -1, Msg: fmt.Sprintf("seed hex err: %s ", err)}
	}

	return api.createWallet(seedBuf, walletPass)
}

//Recove wallet by mnemonic
func (api *API) Recove(mnemonic string, walletPass string) error {
	fmt.Println("mnemonic string:", mnemonic)
	seedBuf, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		fmt.Println("errr:", err.Error())
		return &crateError{Code: -1, Msg: fmt.Sprintf("seed hex err: %s ", err)}
	}
	return api.createWallet(seedBuf, walletPass)
}

//Open wallet
func (api *API) Open(walletPubPass string) error {
	if api.wSvr.Wt != nil {
		log.Trace("api open wallet already open ")
		return nil
	}

	walletPubPassBuf := []byte(wallet.InsecurePubPassphrase)
	wt, err := api.wSvr.WtLoader.OpenExistingWallet(walletPubPassBuf, false)
	if err != nil {
		log.Warnf("api OpenExistingWallet err: %s", err)
		return fmt.Errorf("open wallet err: %s", err)
	}
	log.Trace("api open ok")
	api.wSvr.Wt = wt

	api.wSvr.WtLoader.RunAfterLoad(func(w *wallet.Wallet) {

		w.Start()

		log.Trace("api open RunAfterLoad")

		lockChan := make(chan time.Time, 1)
		defer func() {
			lockChan <- time.Time{}
		}()
		err := w.Unlock([]byte("123456"), lockChan)
		if err != nil {
			fmt.Printf("ERR: Failed to unlock new wallet "+
				"during old wallet key import: %v", err)
			return
		}
		log.Trace("api open RunAfterLoad end")
	})

	api.wSvr.StartAPI()
	log.Trace("api open wallet start")
	return nil
}

// createWallet by seed and walletPass
func (api *API) createWallet(seed []byte, walletPass string) error {
	log.Trace("createWallet", api.cfg.Network)
	fmt.Printf("seed:%x\n", seed)

	activeNetParams := utils.GetNetParams(api.cfg.Network)
	log.Trace("createWallet", activeNetParams.Name)

	dbDir := filepath.Join(api.cfg.AppDataDir, activeNetParams.Name)
	loader := wallet.NewLoader(activeNetParams, dbDir, 250, api.cfg)

	walletExist, err := loader.WalletExists()
	if err != nil {
		log.Errorf("createWallet load wallet err: %s", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("load Wallet err: %s ", err)}
	}
	if walletExist {
		return &crateError{Code: -100, Msg: "wallet exist"}
	}

	wt, err := loader.CreateNewWallet([]byte(wallet.InsecurePubPassphrase), []byte(walletPass), seed, time.Now())
	if err != nil {
		log.Errorf("createWallet loader CreateNewWallet err: %s", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("createWallet loader CreateNewWallet err: %s ", err)}
	}

	//import master key addr
	seedKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		log.Errorf("createWallet NewMasterKey err: %s", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("createWallet NewMasterKey err: %err", err)}
	}
	log.Tracef("createWallet import master key: %x\n", seedKey.Key)

	pri, _ := secp256k1.PrivKeyFromBytes(seedKey.Key)
	wif, err := utils.NewWIF(pri, activeNetParams, true)
	if err != nil {
		log.Errorf("createWallet private key decode failed: %s", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("createWallet private key decode failed: %s", err)}
	}
	if !wif.IsForNet(activeNetParams) {
		log.Errorf("createWallet Key is not intended for: %s %s", activeNetParams.Name, err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("createWallet Key is not intended for: %s", err)}
	}
	wt.UnLockManager([]byte(walletPass))

	_, err = wt.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif, nil, false)
	if err != nil {
		log.Errorf("createWallet ImportPrivateKey err: %s", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("createWallet ImportPrivateKey err: %s", err)}
	}

	wt.Manager.Close()
	//todo,not close,reopen slow
	wt.Database().Close()

	return nil
}

//ResStatus statusinfo
type ResStatus struct {
	Stats string `json:"stats"` //err,nil,closed,lock,unlock
}

// MakeSeed wallet HD seed and mnemonic
func (api *API) MakeSeed() (*ResSeed, error) {
	entropyBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		return nil, fmt.Errorf("Generate entropy err: %s", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropyBuf)
	if err != nil {
		return nil, fmt.Errorf("NewMnemonic err: %s", err)
	}

	seedBuf, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf("NewSeed err: %s", err)
	}

	return &ResSeed{
		Seed:     hex.EncodeToString(seedBuf),
		Mnemonic: mnemonic,
	}, nil
}

//ResSeed make seed
type ResSeed struct {
	Seed     string `json:"seed"`
	Mnemonic string `json:"mnemonic"`
}

type crateError struct {
	Code int
	Msg  string
}

func (e *crateError) ErrorCode() int { return e.Code }
func (e *crateError) Error() string  { return e.Msg }
