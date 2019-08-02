package wserver

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/HalalChain/qitmeer-lib/crypto/bip39"
	"github.com/HalalChain/qitmeer-lib/crypto/seed"

	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/utils"
	"github.com/HalalChain/qitmeer-wallet/wallet"
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

	if api.wSvr.Wt.Locked() {
		status.Stats = "lock"
	} else {
		status.Stats = "unlock"
	}
	log.Debug("wallet api: status", status)
	return
}

//Create wallet
func (api *API) Create(seed string, walletPass string) error {
	log.Trace("CreateAPI CreateWallet", api.cfg.Network)

	activeNetParams := utils.GetNetParams(api.cfg.Network)
	log.Trace("CreateAPI CreateWallet ", activeNetParams.Name)

	dbDir := filepath.Join(api.cfg.AppDataDir, activeNetParams.Name)
	loader := wallet.NewLoader(activeNetParams, dbDir, 250)

	walletExist, err := loader.WalletExists()
	if err != nil {
		log.Errorf("qitmeer start err: load wallet ", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("load Wallet err: %s ", err)}
	}
	if walletExist {
		return &crateError{Code: -100, Msg: "wallet exist"}
	}

	seedBuf, err := hex.DecodeString(seed)
	if err != nil {
		return &crateError{Code: -1, Msg: fmt.Sprintf("seed hex err: %s ", err)}
	}

	wt, err := loader.CreateNewWallet([]byte(wallet.InsecurePubPassphrase), []byte(walletPass), seedBuf, time.Now())
	if err != nil {
		log.Errorf("qitmeer start err: crate wallet ", err)
		return &crateError{Code: -1, Msg: fmt.Sprintf("CreateNewWallet err: %s ", err)}
	}

	//wt.Manager.Close()

	// wt, err := loader.OpenExistingWallet([]byte(walletPass), false)
	// if err != nil {
	// 	log.Errorf("newWallet err: %s", err)
	// 	return fmt.Errorf("open wallet err: %s", err)
	// }
	api.wSvr.Wt = wt

	return nil
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
	wt.Start()

	api.wSvr.StartAPI()
	log.Trace("api open wallet start")
	return nil
}

//ResStatus statusinfo
type ResStatus struct {
	Stats string `json:"stats"` //err,nil,closed,lock,unlock
}

// MakeSeed wallet HD seed and mnemonic
func (api *API) MakeSeed() (*ResSeed, error) {
	seedBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		return nil, fmt.Errorf("GenerateSeed err: %s", err)
	}

	mnemonic, err := bip39.NewMnemonic(seedBuf)
	if err != nil {
		return nil, fmt.Errorf("NewMnemonic err: %s", err)
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
