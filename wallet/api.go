package wallet

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/utils"
)

// API wallet
type API struct {
	cfg *config.Config

	wt *Wallet
}

// NewAPI api make
func NewAPI(cfg *config.Config) *API {
	return &API{}
}

// Status wallet info
func (api *API) Status() (status *ResStatus) {
	status = &ResStatus{}

	if api.wt == nil {
		status.Stats = "nil"
		return
	}

	return nil
}

//Create wallet
func (api *API) Create(walletPass string) error {
	log.Trace("CreateAPI CreateWallet")

	activeNetParams := utils.GetNetParams(api.cfg.Network)

	dbDir := filepath.Join(api.cfg.AppDataDir, activeNetParams.Name)
	loader := NewLoader(activeNetParams, dbDir, 250)

	walletExist, err := loader.WalletExists()
	if err != nil {
		log.Errorf("qitmeer start err: load wallet ", err)
		return fmt.Errorf("load Wallet err: %s ", err)
	}
	if !walletExist {
		err := CreateWallet(api.cfg, walletPass)
		if err != nil {
			log.Errorf("qitmeer start err: crate wallet ", err)
			return fmt.Errorf("crate wallet err: %s", err)
		}
	}
	wt, err := loader.OpenExistingWallet([]byte(walletPass), false)
	if err != nil {
		log.Errorf("newWallet err: %s", err)
		return fmt.Errorf("open wallet err: %s", err)
	}
	api.wt = wt

	return nil
}

//Open wallet
func (api *API) Open(walletPass string) error {
	return nil
}

//ResStatus statusinfo
type ResStatus struct {
	Stats string //nil,closed,opened
}
