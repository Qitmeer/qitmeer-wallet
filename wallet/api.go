package wallet

import (
	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/json/qitmeerjson"
	waddrmgr "github.com/HalalChain/qitmeer-wallet/waddrmgs"
	
)

// AccountAPI wallet
type AccountAPI struct {
	cfg *config.Config

	wt *Wallet
}

// NewAccountAPI api make
func NewAccountAPI(cfg *config.Config, wt *Wallet) *AccountAPI {
	return &AccountAPI{
		cfg: cfg,
		wt:  wt,
	}
}

// List all accounts[{account,balance}]
func (api *AccountAPI) List() (map[string]float64, error) {
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

// Create a account
func (api *AccountAPI) Create(name string) error {

	// The wildcard * is reserved by the rpc server with the special meaning
	// of "all accounts", so disallow naming accounts to this string.
	if name == "*" {
		return &qitmeerjson.ErrReservedAccountName
	}

	_, err :=api.wt.NextAccount(waddrmgr.KeyScopeBIP0044,name)
	if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
		return  &qitmeerjson.RPCError{
			Code: qitmeerjson.ErrRPCWalletUnlockNeeded,
			Message: "Creating an account requires the wallet to be unlocked. " +
				"Enter the wallet passphrase with walletpassphrase to unlock",
		}
	}

	return nil
}
