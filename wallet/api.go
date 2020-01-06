package wallet

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/core/types"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/engine/txscript"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/Qitmeer/qitmeer/params"

	"github.com/Qitmeer/qitmeer-wallet/config"
	clijson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet/txrules"
	"github.com/Qitmeer/qitmeer-wallet/wtxmgr"
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
	Height int32
}

// SyncStats block update stats
func (api *API) SyncStats() (*SyncStats, error) {

	stats := &SyncStats{}

	stats.Height = api.wt.SyncHeight //api.wt.Manager.SyncedTo().Height

	return stats, nil
}

//Unlock wallet
func (api *API) Unlock(walletPriPass string, second int64) error {
	//if api.wSvr.Wt.Locked() {
	err := api.wt.Unlock([]byte(walletPriPass), time.After(time.Duration(second)*time.Second))
	if err != nil {
		log.Error("Failed to unlock new wallet during old wallet key import", "err", err)
		return err
	}
	// } else {
	// 	return nil
	// }
	return nil
}

//Lock wallet
func (api *API) Lock() error {
	api.wt.Lock()
	return nil
}

// GetAccountsAndBalance List all accounts[{account,balance}]
func (api *API) GetAccountsAndBalance() (map[string]*Balance, error) {
	accountsBalances := make(map[string]*Balance)
	aaas, err := api.wt.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044, int32(16))
	if err != nil {
		return nil, err
	}

	for _, aaa := range aaas {

		if _, ok := accountsBalances[aaa.AccountName]; !ok {
			accountsBalances[aaa.AccountName] = &Balance{}
		}

		accountBalance := accountsBalances[aaa.AccountName]

		for _, addr := range aaa.AddrsOutput {
			accountBalance.ConfirmAmount = accountBalance.ConfirmAmount + addr.balance.ConfirmAmount
			accountBalance.SpendAmount = accountBalance.SpendAmount + addr.balance.SpendAmount
			accountBalance.TotalAmount = accountBalance.TotalAmount + addr.balance.TotalAmount
			accountBalance.UnspendAmount = accountBalance.UnspendAmount + addr.balance.UnspendAmount
		}

	}
	return accountsBalances, nil
}

// GetBalanceByAccount get account balance
func (api *API) GetBalanceByAccount(name string) (*Balance, error) {
	aaas, err := api.wt.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044, int32(16))
	if err != nil {
		return nil, err
	}

	accountBalance := &Balance{}

	for _, aaa := range aaas {
		if aaa.AccountName == name {
			for _, addr := range aaa.AddrsOutput {
				accountBalance.ConfirmAmount = accountBalance.ConfirmAmount + addr.balance.ConfirmAmount
				accountBalance.SpendAmount = accountBalance.SpendAmount + addr.balance.SpendAmount
				accountBalance.TotalAmount = accountBalance.TotalAmount + addr.balance.TotalAmount
				accountBalance.UnspendAmount = accountBalance.UnspendAmount + addr.balance.UnspendAmount
			}
		}
	}
	return accountBalance, nil
}

// GetUtxo addr unspend utxo
func (api *API) GetUtxo(addr string) ([]wtxmgr.Utxo, error) {
	results, err := api.wt.GetUtxo(addr)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// CreateAccount create acccount
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

	addrs, err := api.wt.AccountAddresses(account)
	if err != nil {
		return nil, err
	}

	addrStrs := make([]string, len(addrs))
	for i, a := range addrs {
		addrStrs[i] = a.Encode()
	}
	return addrStrs, nil
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
// dumpPrivKey handles a dumpprivkey request with the private key
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
//
// importPrivKey handles an importprivkey request by parsing
// a WIF-encoded private key and adding it to an account.
func (api *API) ImportWifPrivKey(accountName string, privKey string, rescan bool) error {
	// Ensure that private keys are only imported to the correct account.
	//
	// todo
	if accountName != "" && accountName != waddrmgr.ImportedAddrAccountName {
		return &qitmeerjson.ErrNotImportedAccount
	}

	wif, err := utils.DecodeWIF(privKey, api.wt.ChainParams())
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
	_, err = api.wt.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif, nil, rescan)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return &qitmeerjson.ErrWalletUnlockNeeded
	}

	return err
}

// ImportPrivKey import priv key
//
// func importPrivKey(icmd interface{}, w *wallet.Wallet) (interface{}, error) {
// 	cmd := icmd.(*qitmeerjson.ImportPrivKeyCmd)
func (api *API) ImportPrivKey(accountName string, privKey string, rescan bool) error {
	// Ensure that private keys are only imported to the correct account.
	//
	// Yes, Label is the account name.
	if accountName != "" && accountName != waddrmgr.ImportedAddrAccountName {
		return &qitmeerjson.ErrNotImportedAccount
	}

	prihash, err := hex.DecodeString(privKey)
	if err != nil {
		return err
	}
	pri, _ := secp256k1.PrivKeyFromBytes(prihash)
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
	_, err = api.wt.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif, nil, rescan)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return &qitmeerjson.ErrWalletUnlockNeeded
	}

	return err
}

//SendToAddress handles a sendtoaddress RPC request by creating a new
//transaction spending unspent transaction outputs for a wallet to another
//payment address.  Leftover inputs not sent to the payment address or a fee
//for the miner are sent back to a new address in the wallet.  Upon success,
//the TxID for the created transaction is returned.
// func sendToAddress(icmd interface{}, w *wallet.Wallet) (interface{}, error) {
// 	cmd := icmd.(*qitmeerjson.SendToAddressCmd)
func (api *API) SendToAddress(addressStr string, amount float64, comment string, commentTo string) (string, error) {

	// Transaction comments are not yet supported.  Error instead of
	// pretending to save them.
	//if !isNilOrEmpty(cmd.Comment) || !isNilOrEmpty(cmd.CommentTo) {
	if comment != "" || commentTo != "" {
		return "", &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCUnimplemented,
			Message: "Transaction comments are not yet supported",
		}
	}

	amt, err := types.NewAmount(amount)
	if err != nil {
		return "", err
	}

	// Check that signed integer parameters are positive.
	if amt < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addressStr: amt,
	}

	// sendtoaddress always spends from the default account, this matches bitcoind
	//return sendPairs(api.wt, pairs, waddrmgr.DefaultAccountNum, 1, txrules.DefaultRelayFeePerKb)
	return sendPairs(api.wt, pairs, waddrmgr.AccountMergePayNum, 1, txrules.DefaultRelayFeePerKb)
}

// SendToAddressByAccount by account
func (api *API) SendToAddressByAccount(accountName string, addressStr string, amount float64, comment string, commentTo string) (string, error) {

	accountNum, err := api.wt.AccountNumber(waddrmgr.KeyScopeBIP0044, accountName)
	if err != nil {
		return "", err
	}

	// Transaction comments are not yet supported.  Error instead of
	// pretending to save them.
	//if !isNilOrEmpty(cmd.Comment) || !isNilOrEmpty(cmd.CommentTo) {
	if comment != "" || commentTo != "" {
		return "", &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCUnimplemented,
			Message: "Transaction comments are not yet supported",
		}
	}

	amt, err := types.NewAmount(amount)
	if err != nil {
		return "", err
	}

	// Check that signed integer parameters are positive.
	if amt < 0 {
		return "", qitmeerjson.ErrNeedPositiveAmount
	}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		addressStr: amt,
	}

	// sendtoaddress always spends from the default account, this matches bitcoind
	return sendPairs(api.wt, pairs, int64(accountNum), 1, txrules.DefaultRelayFeePerKb)
}

//GetBalanceByAddr get balance by address
func (api *API) GetBalanceByAddr(addrStr string, minConf int32) (*Balance, error) {
	m, err := api.wt.GetBalance(addrStr, minConf)
	if err != nil {
		return nil, err
	}
	return m, nil
}

//GetTxListByAddr get addr tx list
func (api *API) GetTxListByAddr(addr string, stype int32, page int32, pageSize int32) (clijson.PageTxRawResult, error) {
	rs, err := api.wt.GetListTxByAddr(addr, stype, page, pageSize)
	return *rs, err
}

//sendPairs creates and sends payment transactions.
//It returns the transaction hash in string format upon success
//All errors are returned in btcjson.RPCError format
func sendPairs(w *Wallet, amounts map[string]types.Amount,
	account int64 /*uint32*/, minconf int32, feeSatPerKb types.Amount) (string, error) {

	outputs, err := makeOutputs(amounts, w.ChainParams())
	if err != nil {
		return "", err
	}
	tx, err := w.SendOutputs(outputs, account, minconf, feeSatPerKb)
	if err != nil {
		if err == txrules.ErrAmountNegative {
			return "", qitmeerjson.ErrNeedPositiveAmount
		}
		if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
			return "", &qitmeerjson.ErrWalletUnlockNeeded
		}
		switch err.(type) {
		case qitmeerjson.RPCError:
			return "", err
		}

		return "", &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInternal.Code,
			Message: err.Error(),
		}
	}
	return *tx, nil
}

// makeOutputs creates a slice of transaction outputs from a pair of address
// strings to amounts.  This is used to create the outputs to include in newly
// created transactions from a JSON object describing the output destinations
// and amounts.
func makeOutputs(pairs map[string]types.Amount, chainParams *params.Params) ([]*types.TxOutput, error) {
	outputs := make([]*types.TxOutput, 0, len(pairs))
	for addrStr, amt := range pairs {
		addr, err := address.DecodeAddress(addrStr)
		if err != nil {
			return nil, fmt.Errorf("cannot decode address: %s", err)
		}

		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, fmt.Errorf("cannot create txout script: %s", err)
		}

		outputs = append(outputs, types.NewTxOutput(uint64(amt), pkScript))
	}
	return outputs, nil
}

func isNilOrEmpty(s *string) bool {
	return s == nil || *s == ""
}
