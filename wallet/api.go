package wallet

import (
	"github.com/HalalChain/qitmeer-wallet/config"
	waddrmgr "github.com/HalalChain/qitmeer-wallet/waddrmgs"
)

// AccountAPI wallet
type AccountAPI struct {
	cfg *config.Config

	wt *Wallet
}

// NewAccountAPI api make
func NewAccountAPI(cfg *config.Config) *AccountAPI {
	return &AccountAPI{
		cfg: cfg,
	}
}

// ListAccount
func (api *AccountAPI) ListAccounts() (map[string]float64, error) {
	accountBalances := map[string]float64{}
	results, err := api.wt.AccountBalances(waddrmgr.KeyScopeBIP0044, int32(16))
	if err != nil {
		return nil, err
	}

	for _, result := range results {
		accountBalances[result.AccountName] = result.AccountBalance.ToCoin()
	}
	// Return the map.  This will be marshaled into a JSON object.
	return accountBalances, nil
}
