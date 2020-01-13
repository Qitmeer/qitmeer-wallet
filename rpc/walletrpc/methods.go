package walletrpc

import (
	"encoding/hex"
	"fmt"
	"time"

	util "github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wallet/txrules"
	"github.com/Qitmeer/qitmeer/core/address"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/log"

	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer/core/types"

	"github.com/Qitmeer/qitmeer-wallet/wallet"
)





// createNewAccount handles a createnewaccount request by creating and
// returning a new account. If the last account has no transaction history
// as per BIP 0044 a new account cannot be created so an error will be returned.
func CreateNewAccount(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.CreateNewAccountCmd)

	// The wildcard * is reserved by the rpc server with the special meaning
	// of "all accounts", so disallow naming accounts to this string.
	if cmd.Account == "*" {
		return nil, &qitmeerjson.ErrReservedAccountName
	}

	_, err := w.NextAccount(waddrmgr.KeyScopeBIP0044, cmd.Account)
	if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
		return nil, &qitmeerjson.RPCError{
			Code: qitmeerjson.ErrRPCWalletUnlockNeeded,
			Message: "Creating an account requires the wallet to be unlocked. " +
				"Enter the wallet passphrase with walletpassphrase to unlock",
		}
	}
	return "succ", err
}

// listAccounts handles a listaccounts request by returning a map of account
// names to their balances.
func ListAccounts(w *wallet.Wallet) (interface{}, error) {
	accountBalances := map[string]float64{}
	results, err := w.AccountBalances(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		accountBalances[result.AccountName] = result.AccountBalance.ToCoin()

	}
	// Return the map.  This will be marshaled into a JSON object.
	return accountBalances, nil
}

// getNewAddress handles a GetNewAddress request by returning a new
// address for an account.  If the account does not exist an appropiate
// error is returned.
// the gap limit.
func GetNewAddress(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.GetNewAddressCmd)

	acctName := "default"
	if cmd.Account != nil {
		acctName = *cmd.Account
	}
	if acctName == "imported" {
		return nil, fmt.Errorf("Import account cannot create subaddress.")
	}
	account, err := w.AccountNumber(waddrmgr.KeyScopeBIP0044, acctName)
	if err != nil {
		return nil, err
	}
	addr, err := w.NewAddress(waddrmgr.KeyScopeBIP0044, account)
	if err != nil {
		return nil, err
	}
	// Return the new payment address string.
	return addr.Encode(), nil
}

// getAddressesByAccount handles a getaddressesbyaccount request by returning
// all addresses for an account, or an error if the requested account does
// not exist.
func GetAddressesByAccount(iCmd interface{}, w *wallet.Wallet) ([]string, error) {
	cmd := iCmd.(*qitmeerjson.GetAddressesByAccountCmd)

	account, err := w.AccountNumber(waddrmgr.KeyScopeBIP0044, cmd.Account)
	if err != nil {
		return nil, err
	}

	addrs, err := w.AccountAddresses(account)
	if err != nil {
		return nil, err
	}

	addrStrs := make([]string, len(addrs))
	for i, a := range addrs {
		addrStrs[i] = a.Encode()
	}
	return addrStrs, nil
}

func GetAccountAndAddress(w *wallet.Wallet) (interface{}, error) {
	a, err := w.GetAccountAndAddress(waddrmgr.KeyScopeBIP0044)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// getAccount handles a getaccount request by returning the account name
// associated with a single address.
func GetAccountByAddress(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.GetAccountCmd)

	addr, err := address.DecodeAddress(cmd.Address)
	if err != nil {
		return nil, err
	}

	// Fetch the associated account
	account, err := w.AccountOfAddress(addr)
	if err != nil {
		return nil, &qitmeerjson.ErrAddressNotInWallet
	}

	acctName, err := w.AccountName(waddrmgr.KeyScopeBIP0044, account)
	if err != nil {
		return nil, &qitmeerjson.ErrAccountNameNotFound
	}
	return acctName, nil
}

// dumpPrivKey handles a dumpprivkey request with the private key
// for a single address, or an appropiate error if the wallet
// is locked.
func DumpPrimKey(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.DumpPrivKeyCmd)

	addr, err := address.DecodeAddress(cmd.Address)
	if err != nil {
		return nil, err
	}

	key, err := w.DumpWIFPrivateKey(addr)
	if waddrmgr.IsError(err, waddrmgr.ErrLocked) {
		// Address was found, but the private key isn't
		// accessible.
		return nil, &qitmeerjson.ErrWalletUnlockNeeded
	}
	return key, err
}

// importPrivKey handles an importprivkey request by parsing
// a WIF-encoded private key and adding it to an account.
func ImportWifiPrimKey(icmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := icmd.(*qitmeerjson.ImportPrivKeyCmd)

	// Ensure that private keys are only imported to the correct account.
	//
	// Yes, Label is the account name.
	if cmd.Label != nil && *cmd.Label != waddrmgr.ImportedAddrAccountName {
		return nil, &qitmeerjson.ErrNotImportedAccount
	}

	wif, err := util.DecodeWIF(cmd.PrivKey, w.ChainParams())
	if err != nil {
		return nil, &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "WIF decode failed: " + err.Error(),
		}
	}
	if !wif.IsForNet(w.ChainParams()) {
		return nil, &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "Key is not intended for " + w.ChainParams().Name,
		}
	}

	// Import the private key, handling any errors.
	_, err = w.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil, nil
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return nil, &qitmeerjson.ErrWalletUnlockNeeded
	}

	return nil, err
}
func ImportPrimKey(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.ImportPrivKeyCmd)
	// Ensure that private keys are only imported to the correct account.
	//
	// Yes, Label is the account name.
	if cmd.Label != nil && *cmd.Label != waddrmgr.ImportedAddrAccountName {
		return nil, &qitmeerjson.ErrNotImportedAccount
	}

	prihash, err := hex.DecodeString(cmd.PrivKey)
	if err != nil {
		return nil, err
	}
	pri, _ := secp256k1.PrivKeyFromBytes(prihash)
	wif, err := util.NewWIF(pri, w.ChainParams(), true)
	if err != nil {
		return nil, &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "private key decode failed: " + err.Error(),
		}
	}
	if !wif.IsForNet(w.ChainParams()) {
		return nil, &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCInvalidAddressOrKey,
			Message: "Key is not intended for " + w.ChainParams().Name,
		}
	}

	// Import the private key, handling any errors.
	_, err = w.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	switch {
	case waddrmgr.IsError(err, waddrmgr.ErrDuplicateAddress):
		// Do not return duplicate key errors to the client.
		return nil, fmt.Errorf("private key imported")
	case waddrmgr.IsError(err, waddrmgr.ErrLocked):
		return nil, &qitmeerjson.ErrWalletUnlockNeeded
	}

	return "ok", err
}

//sendToAddress handles a sendtoaddress RPC request by creating a new
//transaction spending unspent transaction outputs for a wallet to another
//payment address.  Leftover inputs not sent to the payment address or a fee
//for the miner are sent back to a new address in the wallet.  Upon success,
//the TxID for the created transaction is returned.
func SendToAddress(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.SendToAddressCmd)

	// Transaction comments are not yet supported.  Error instead of
	// pretending to save them.
	if !isNilOrEmpty(cmd.Comment) || !isNilOrEmpty(cmd.CommentTo) {
		return nil, &qitmeerjson.RPCError{
			Code:    qitmeerjson.ErrRPCUnimplemented,
			Message: "Transaction comments are not yet supported",
		}
	}

	amt, err := types.NewAmount(cmd.Amount)
	if err != nil {
		return nil, err
	}

	// Check that signed integer parameters are positive.
	if amt < 0 {
		return nil, qitmeerjson.ErrNeedPositiveAmount
	}

	// Mock up map of address and amount pairs.
	pairs := map[string]types.Amount{
		cmd.Address: amt,
	}

	return w.SendPairs(pairs, int64(waddrmgr.AccountMergePayNum),  txrules.DefaultRelayFeePerKb)
}

func UpdateBlock(iCmd interface{}, w *wallet.Wallet) error {
	cmd := iCmd.(*qitmeerjson.UpdateBlockToCmd)
	err := w.UpdateBlock(cmd.Toheight)
	if err != nil {
		return err
	}
	return nil
}
func GetTx(txId string, w *wallet.Wallet) (interface{}, error) {
	tx, err := w.GetTx(txId)
	if err != nil {
		return "", err
	}
	return tx, nil
}
func GetBalance(iCmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := iCmd.(*qitmeerjson.GetBalanceByAddressCmd)
	m, err := w.GetBalance(cmd.Address)
	if err != nil {
		log.Error("GetBalance ", "err ", err.Error())
		return nil, err
	}
	return m, nil
}

func Unlock(password string, w *wallet.Wallet) error {
	return w.Unlock([]byte(password), time.After(10*time.Minute))
}

func GetListTxByAddr(icmd interface{}, w *wallet.Wallet) (interface{}, error) {
	cmd := icmd.(*qitmeerjson.GetListTxByAddrCmd)
	m, err := w.GetListTxByAddr(cmd.Address, cmd.Stype, cmd.Page, cmd.PageSize)
	if err != nil {
		log.Error("GetListTxByAddr ", " err", err.Error())
		return nil, err
	}
	return m, nil
}


func isNilOrEmpty(s *string) bool {
	return s == nil || *s == ""
}
