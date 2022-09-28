package wallet

import (
	"encoding/hex"
	corejson "github.com/Qitmeer/qng/core/json"
	"time"

	"github.com/Qitmeer/qitmeer-wallet/config"
	clijson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet/txrules"
	"github.com/Qitmeer/qitmeer-wallet/wtxmgr"
	"github.com/Qitmeer/qng/core/address"
	"github.com/Qitmeer/qng/core/types"
	"github.com/Qitmeer/qng/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qng/log"
)

// API for wallet
type API struct {
	cfg *config.Config
	wt  *Wallet
}

// NewAPI make api
func NewAPI(cfg *config.Config, wt *Wallet) *API {
	return &API{
		cfg: cfg,
		wt:  wt,
	}
}

//SyncStats block update stats
type SyncStats struct {
	Order uint32
}

// SyncStats block update stats
func (api *API) SyncStats() (*SyncStats, error) {

	stats := &SyncStats{}

	stats.Order = api.wt.getSyncOrder() //api.wt.Manager.SyncedTo().Height

	return stats, nil
}

//Unlock wallet
func (api *API) Unlock(walletPriPass string, second int64) error {
	//if api.wSvr.Wt.Locked() {
	var dur <-chan time.Time
	if second < 0 {
		dur = nil
	} else {
		dur = time.After(time.Duration(second) * time.Second)
	}
	err := api.wt.Unlock([]byte(walletPriPass), dur)
	if err != nil {
		log.Error("Failed to unlock new wallet during old wallet key import", "err", err)
		return err
	}

	return nil
}

//Lock wallet
func (api *API) Lock() error {
	api.wt.Lock()
	return nil
}

// GetAccountsAndBalance List all accounts[{account,balance}]
func (api *API) GetAccountsAndBalance(coin types.CoinID) (map[string]*Value, error) {
	accountsBalances := make(map[string]*Value)
	aaas, err := api.wt.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}
	coinID, err := api.wt.CoinID(types.CoinID(coin))
	if err != nil {
		return map[string]*Value{}, err
	}
	for _, aaa := range aaas {

		if _, ok := accountsBalances[aaa.AccountName]; !ok {
			accountsBalances[aaa.AccountName] = &Value{}
		}

		accountBalance := accountsBalances[aaa.AccountName]
		for _, addr := range aaa.AddrsOutput {
			accountBalance.TotalAmount += addr.balanceMap[coinID].TotalAmount.Value
			accountBalance.SpendAmount += addr.balanceMap[coinID].SpendAmount.Value
			accountBalance.UnspentAmount += addr.balanceMap[coinID].UnspentAmount.Value
			accountBalance.UnconfirmedAmount += addr.balanceMap[coinID].UnconfirmedAmount.Value
			accountBalance.LockAmount += addr.balanceMap[coinID].LockAmount.Value
		}

	}
	return accountsBalances, nil
}

// GetBalanceByAccount get account balance
func (api *API) GetBalanceByAccount(name string, coin types.CoinID) (*Value, error) {
	results, err := api.wt.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}

	coinID, err := api.wt.CoinID(types.CoinID(coin))
	if err != nil {
		return nil, err
	}
	accountBalance := &Value{}
	for _, result := range results {
		if result.AccountName == name {
			for _, addr := range result.AddrsOutput {
				accountBalance.TotalAmount += addr.balanceMap[coinID].TotalAmount.Value
				accountBalance.SpendAmount += addr.balanceMap[coinID].SpendAmount.Value
				accountBalance.UnspentAmount += addr.balanceMap[coinID].UnspentAmount.Value
				accountBalance.UnconfirmedAmount += addr.balanceMap[coinID].UnconfirmedAmount.Value
				accountBalance.LockAmount += addr.balanceMap[coinID].LockAmount.Value
			}
		}
	}

	return accountBalance, nil
}

// GetUTxo addr unSpend UTxo
func (api *API) GetUTxo(addr string, coin types.CoinID) ([]wtxmgr.UTxo, error) {
	results, err := api.wt.GetUnspentUTXO(addr, coin)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetTx get transaction by ID
func (api *API) GetTx(txID string) (*corejson.TxRawResult, error) {
	result, err := api.wt.GetTx(txID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateAccount create account
func (api *API) CreateAccount(name string) error {
	// The wildcard * is reserved by the rpc server with the special meaning
	// of "all accounts", so disallow naming accounts to this string.
	if name == "*" {
		return &qitmeerjson.ErrReservedAccountName
	}

	_, err := api.wt.NextAccount(waddrmgr.KeyScopeBIP0044, name)
	if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
		return &qitmeerjson.RPCError{
			Code: qitmeerjson.ErrRPCWalletUnlockNeeded,
			Message: "Creating an account requires the wallet to be unlocked. " +
				"Enter the wallet passphrase with walletpassphrase to unlock",
		}
	}
	return nil
}

// CreateAddress by accountName
func (api *API) CreateAddress(accountName string) (string, error) {
	if accountName == "" {
		accountName = "default"
	}
	account, err := api.wt.AccountNumber(waddrmgr.KeyScopeBIP0044, accountName)
	if err != nil {
		return "", err
	}
	addr, err := api.wt.NewAddress(waddrmgr.KeyScopeBIP0044, account)
	if err != nil {
		return "", err
	}
	// Return the new payment address string.
	return addr.Encode(), nil
}

// GetAddressesByAccount by account
func (api *API) GetAddressesByAccount(accountName string) ([]string, error) {
	account, err := api.wt.AccountNumber(waddrmgr.KeyScopeBIP0044, accountName)
	if err != nil {
		return nil, err
	}

	adds, err := api.wt.AccountAddresses(account)
	if err != nil {
		return nil, err
	}

	addrStr := make([]string, len(adds))
	for i, a := range adds {
		addrStr[i] = a.Encode()
	}
	return addrStr, nil
}

// GetAccountByAddress get account name
func (api *API) GetAccountByAddress(addrStr string) (string, error) {
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return "", err
	}
	// Fetch the associated account
	account, err := api.wt.AccountOfAddress(addr)
	if err != nil {
		return "", &qitmeerjson.ErrAddressNotInWallet
	}

	acctName, err := api.wt.AccountName(waddrmgr.KeyScopeBIP0044, account)
	if err != nil {
		return "", &qitmeerjson.ErrAccountNameNotFound
	}
	return acctName, nil
}

// DumpPrivKey dump a single address private key
//
// dumpPriKey handles a DumpPrivKey request with the private key
// for a single address, or an appropiate error if the wallet
// is locked.
func (api *API) DumpPrivKey(addrStr string) (string, error) {
	addr, err := address.DecodeAddress(addrStr)
	if err != nil {
		return "", err
	}

	key, err := api.wt.DumpWIFPrivateKey(addr)
	if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
		// Address was found, but the private key isn't
		// accessible.
		return "", &qitmeerjson.ErrWalletUnlockNeeded
	}
	return key, err
}

// ImportWifPrivKey import a WIF-encoded private key and adding it to an account
// a WIF-encoded private key and adding it to an account.
func (api *API) ImportWifPrivKey(accountName string, key string) error {
	// Ensure that private keys are only imported to the correct account.
	if accountName != "" && accountName != waddrmgr.ImportedAddrAccountName {
		return &qitmeerjson.ErrNotImportedAccount
	}

	wif, err := utils.DecodeWIF(key, api.wt.ChainParams())
	if err != nil {
		return &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "WIF decode failed: " + err.Error(),
		}
	}
	if !wif.IsForNet(api.wt.ChainParams()) {
		return &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "Key is not intended for " + api.wt.ChainParams().Name,
		}
	}

	// Import the private key, handling any errors.
	_, err = api.wt.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return &qitmeerjson.ErrWalletUnlockNeeded
	}

	return err
}

// ImportPrivKey import pri key
func (api *API) ImportPrivKey(accountName string, key string) error {
	// Ensure that private keys are only imported to the correct account.
	//
	// Yes, Label is the account name.
	if accountName != "" && accountName != waddrmgr.ImportedAddrAccountName {
		return &qitmeerjson.ErrNotImportedAccount
	}

	priHash, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	pri, _ := secp256k1.PrivKeyFromBytes(priHash)
	wif, err := utils.NewWIF(pri, api.wt.ChainParams(), true)
	if err != nil {
		return &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "private key decode failed: " + err.Error(),
		}
	}
	if !wif.IsForNet(api.wt.ChainParams()) {
		return &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "Key is not intended for " + api.wt.ChainParams().Name,
		}
	}

	// Import the private key, handling any errors.
	_, err = api.wt.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return &qitmeerjson.ErrWalletUnlockNeeded
	}

	return err
}

type ApiAmount struct {
	Value float64
	Coin  string
}

//SendToAddress handles a sendtoaddress RPC request by creating a new
//transaction spending unspent transaction outputs for a wallet to another
//payment address.  Leftover inputs not sent to the payment address or a fee
//for the miner are sent back to a new address in the wallet.  Upon success,
//the TxID for the created transaction is returned.
func (api *API) SendToAddress(addressStr string, amount float64, coin types.CoinID, byAddress string) (string, error) {

	// Check that signed integer parameters are positive.
	if amount < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	coinID, err := api.wt.CoinID(types.CoinID(coin))
	if err != nil {
		return "", err
	}
	amt := types.Amount{Value: int64(amount * types.AtomsPerCoin), Id: coinID}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addressStr: amt,
	}

	return api.wt.SendPairs(pairs, waddrmgr.AccountMergePayNum, txrules.DefaultRelayFeePerKb, 0, byAddress)
}

//SendToAddress handles a sendtoaddress RPC request by creating a new
//transaction spending unspent transaction outputs for a wallet to another
//payment address.  Leftover inputs not sent to the payment address or a fee
//for the miner are sent back to a new address in the wallet.  Upon success,
//the TxID for the created transaction is returned.
func (api *API) SendLockedToAddress(addressStr string, amount float64, coin types.CoinID, lockHeight uint64) (string, error) {

	// Check that signed integer parameters are positive.
	if amount < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	id, err := api.wt.CoinID(types.CoinID(coin))
	if err != nil {
		return "", err
	}
	amt, err := types.NewAmount(amount)
	if err != nil {
		return "", err
	}
	amt.Id = id
	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addressStr: *amt,
	}

	return api.wt.SendPairs(pairs, waddrmgr.AccountMergePayNum, txrules.DefaultRelayFeePerKb, lockHeight, "")
}

func (api *API) SendToMany(addAmounts map[string]float64, coin types.CoinID, byAddress string) (string, error) {

	pairs := make(map[string]types.Amount)
	for addr, amount := range addAmounts {
		if amount < 0 {
			return "", qitmeerjson.ErrNeedPositiveAmount
		}
		coinID, err := api.wt.CoinID(types.CoinID(coin))
		if err != nil {
			return "", err
		}
		amt := types.Amount{Value: int64(amount * types.AtomsPerCoin), Id: coinID}

		pairs[addr] = amt
	}

	return api.wt.SendPairs(pairs, waddrmgr.AccountMergePayNum, txrules.DefaultRelayFeePerKb, 0, byAddress)
}

// SendToAddressByAccount by account
func (api *API) SendToAddressByAccount(accountName string, addressStr string, amount float64, coin types.CoinID, comment string, commentTo string) (string, error) {

	accountNum, err := api.wt.AccountNumber(waddrmgr.KeyScopeBIP0044, accountName)
	if err != nil {
		return "", err
	}

	// Check that signed integer parameters are positive.
	if amount < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	coinID, err := api.wt.CoinID(types.CoinID(coin))
	if err != nil {
		return "", nil
	}
	amt := types.Amount{Value: int64(amount * types.AtomsPerCoin), Id: coinID}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addressStr: amt,
	}

	return api.wt.SendPairs(pairs, int64(accountNum), txrules.DefaultRelayFeePerKb, 0, "")
}

//GetBalanceByAddr get balance by address
func (api *API) GetBalanceByAddr(addrStr string, coin types.CoinID) (map[string]Value, error) {
	m, err := api.wt.GetBalanceByCoin(addrStr, coin)
	if err != nil {
		return nil, err
	}
	return m, nil
}

//GetTxListByAddr get transactions affecting specific address, one transaction could affect MULTIPLE addresses
func (api *API) GetTxListByAddr(addr string, sType int, page int, pageSize int) (*clijson.PageTxRawResult, error) {
	rs, err := api.wt.GetListTxByAddr(addr, sType, page, pageSize)
	return rs, err
}

//GetBillByAddr get bill of payments affecting specific address, one payment could affect ONE address
func (api *API) GetBillByAddr(addr string, filter int, page int, pageSize int) (*clijson.PagedBillResult, error) {
	rs, err := api.wt.GetBillByAddr(addr, filter, page, pageSize)
	return rs, err
}
