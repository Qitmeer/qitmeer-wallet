package commands

import (
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/rpc/walletrpc"
	"github.com/Qitmeer/qitmeer-wallet/util"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer-wallet/wtxmgr"
	"github.com/Qitmeer/qitmeer/crypto/bip39"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/qx"
	"github.com/spf13/cobra"
	"path/filepath"
	"runtime"
)

const (
	Name = "wallet-cli:"
)

var w *wallet.Wallet
var isWin = runtime.GOOS == "windows"

// interactive mode

var ConsoleCmd = &cobra.Command{
	Use:   "console",
	Short: "interactive mode",
	Example: `
		Enter console mode
		`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		b := checkWalletIeExist(config.Cfg)
		if b == false {
			fmt.Println("Please create a wallet first,[qitmeer-wallet qc create ]")
			return
		}
		needMn := ""
		if len(args) >= 1 {
			needMn = args[0]
		}
		startConsole(needMn)
	},
}

func AddConsoleCommand() {

}

func CreatWallet(needMnemonic string) {
	b := checkWalletIeExist(config.Cfg)
	if b {
		fmt.Println("db is exist", filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
		return
	} else {
		_, err := createWallet(needMnemonic)
		if err != nil {
			fmt.Println("createWallet err:", err.Error())
			return
		} else {
			fmt.Println("createWallet succ")
		}
		return
	}
}

func OpenWallet() error {
	b := checkWalletIeExist(config.Cfg)
	var err error
	if b {
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250, config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			return fmt.Errorf("openWallet err:%s\n", err.Error())
		}
	} else {
		return fmt.Errorf("Please create a wallet first,[qitmeer-wallet qc create ]")
	}
	return nil
}

func UnLock(password string) error {
	err := w.UnLockManager([]byte(password))
	if err != nil {
		return fmt.Errorf("password error")
	}
	return nil
}

func startConsole(needMnemonic string) {
	b := checkWalletIeExist(config.Cfg)
	var err error
	if b {
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250, config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			fmt.Println("openWallet err:", err.Error())
			return
		}
		w.Start()
	} else {
		w, err = createWallet(needMnemonic)
		if err != nil {
			fmt.Println("createWallet err:", err.Error())
			return
		}
		return
	}
	c := Config{Prompt: Name}
	con, err := New(c)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	con.Interactive()
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("\t<command> [arguments]")
	fmt.Println("\tThe commands are:")
	fmt.Println("\t<createNewAccount> : Create a new account. Parameter: [account]")
	fmt.Println("\t<getBalance> : Query the specified address balance. Parameter: [address]")
	fmt.Println("\t<listAccountsBalance> : Obtain all account balances. Parameter: []")
	fmt.Println("\t<getTx> : Gets transaction by ID. Parameter: [txID]")
	fmt.Println("\t<getListTxByAddr> : Gets all transactions that affect specified address, one transaction could affect MULTIPLE addresses. Parameter: [address] [stype:in,out,all]")
	fmt.Println("\t<getBillByAddr> : Gets all payments that affect specified address, one payment could affect only ONE address. Parameter: [address] [filter:in,out,all]")
	fmt.Println("\t<getNewAddress> : Create a new address under the account. Parameter: [account]")
	fmt.Println("\t<getAddressesByAccount> : Check all addresses under the account. Parameter: [account]")
	fmt.Println("\t<getAccountByAddress> : Inquire about the account number of the address. Parameter: [address]")
	fmt.Println("\t<importPrivKey> : Import private key. Parameter: [priKey]")
	fmt.Println("\t<importWifPrivKey> : Import wif format private key. Parameter: [priKey]")
	fmt.Println("\t<dumpPrivKey> : Export wif format private key by address. Parameter: [address]")
	fmt.Println("\t<getAccountAndAddress> : Check all accounts and addresses. Parameter: []")
	fmt.Println("\t<sendToAddress> : Transfer transaction. Parameter: [address] [coin] [num]")
	fmt.Println("\t<updateblock> : Update Wallet Block. Parameter: []")
	fmt.Println("\t<syncheight> : Current Synchronized Data Height. Parameter: []")
	fmt.Println("\t<unlock> : Unlock Wallet. Parameter: [password]")
	fmt.Println("\t<help> : help")
	fmt.Println("\t<exit> : Exit command mode")
	fmt.Println("")
}

func checkWalletIeExist(cfg *config.Config) bool {
	netDir := networkDir(cfg.AppDataDir, config.ActiveNet)
	err := checkCreateDir(netDir)
	if err != nil {
		return false
	}
	dbPath := filepath.Join(netDir, config.WalletDbName)
	if fi, err := util.FileExists(dbPath); err != nil {
		fmt.Println("FileExists ", "err", err.Error())
		return false
	} else {
		return fi
	}
}

func createNewAccount(arg string) error {
	cmd := &qitmeerjson.CreateNewAccountCmd{
		Account: arg,
	}
	msg, err := walletrpc.CreateNewAccount(cmd, w)
	if err != nil {
		fmt.Println("createNewAccount", "err", err.Error())
		return err
	}
	fmt.Printf("%s", msg)
	return nil
}
func getBalance(addr string) (map[string]wallet.Balance, error) {
	cmd := &qitmeerjson.GetBalanceByAddressCmd{
		Address: addr,
	}
	b, err := walletrpc.GetBalance(cmd, w)
	if err != nil {
		fmt.Println("getBalance", "err", err.Error())
		return nil, err
	}
	r := b.(map[string]wallet.Balance)
	return r, nil
}

func GetTxSpendInfo(txId string) ([]*wtxmgr.AddrTxOutput, error) {
	b, err := walletrpc.GetTxSpendInfo(txId, w)
	if err != nil {
		fmt.Println("GetTxSpendInfo", "err", err.Error())
		return nil, err
	}
	r := b.([]*wtxmgr.AddrTxOutput)
	return r, nil
}

func listAccountsBalance() (interface{}, error) {
	helper = &JsonCmdHelper{
		JsonCmd: nil,
		Run: func(cmd interface{}, w *wallet.Wallet) (interface{}, error) {
			return walletrpc.ListAccounts(w)
		},
	}
	return helper.Call()
}

func getListTxByAddr(addr string, filter int, pageNo int, pageSize int) (interface{}, error) {
	helper = &JsonCmdHelper{
		JsonCmd: &qitmeerjson.GetListTxByAddrCmd{
			Address:  addr,
			Stype:    int32(filter),
			Page:     int32(pageNo),
			PageSize: int32(pageSize),
		},
		Run: func(cmd interface{}, w *wallet.Wallet) (interface{}, error) {
			return walletrpc.GetListTxByAddr(cmd, w)
		},
	}
	return helper.Call()
}

func getBillByAddr(addr string, filter int, pageNo int, pageSize int) (interface{}, error) {
	helper = &JsonCmdHelper{
		JsonCmd: &qitmeerjson.GetBillByAddrCmd{
			Address:  addr,
			Filter:   int32(filter),
			PageNo:   int32(pageNo),
			PageSize: int32(pageSize),
		},
		Run: func(cmd interface{}, w *wallet.Wallet) (interface{}, error) {
			return walletrpc.GetBillByAddr(cmd, w)
		},
	}
	return helper.Call()
}

func getNewAddress(account string) (interface{}, error) {
	if account == waddrmgr.ImportedAddrAccountName {
		fmt.Sprintf("Imported account cannot generate address")
		return nil, fmt.Errorf("imported account cannot generate address")
	}
	if account == "" {
		account = "default"
	}
	cmd := &qitmeerjson.GetNewAddressCmd{
		Account: &account,
	}
	msg, err := walletrpc.GetNewAddress(cmd, w)
	if err != nil {
		fmt.Println("getNewAddress", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}

func SetSyncedToNum(order int64) error {
	return walletrpc.SetSyncedToNum(order, w)
}

func getAddressesByAccount(account string) (interface{}, error) {
	if account == "" {
		account = "default"
	}
	cmd := &qitmeerjson.GetAddressesByAccountCmd{
		Account: account,
	}
	msg, err := walletrpc.GetAddressesByAccount(cmd, w)
	if err != nil {
		fmt.Println("getAddressesByAccount", "err", err.Error())
		return nil, err
	}
	for _, addr := range msg {
		fmt.Printf("%s\n", addr)
	}
	return msg, nil
}
func getAccountByAddress(address string) (interface{}, error) {
	cmd := &qitmeerjson.GetAccountCmd{
		Address: address,
	}
	msg, err := walletrpc.GetAccountByAddress(cmd, w)
	if err != nil {
		fmt.Println("getAccountByAddress", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}

func getTx(txID string) (interface{}, error) {
	helper = &JsonCmdHelper{
		JsonCmd: txID,
		Run: func(cmd interface{}, w *wallet.Wallet) (interface{}, error) {
			txID := cmd.(string)
			return walletrpc.GetTx(txID, w)
		},
	}
	return helper.Call()
}

func importPrivKey(priKey string) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: priKey,
		Rescan:  &v,
	}
	msg, err := walletrpc.ImportPrivKey(cmd, w)
	if err != nil {
		fmt.Println("importPrivKey", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func importWifPrivKey(wifPriKey string) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: wifPriKey,
		Rescan:  &v,
	}
	msg, err := walletrpc.ImportWifPrivKey(cmd, w)
	if err != nil {
		fmt.Println("importWifPrivKey", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func dumpPrivKey(address string) (interface{}, error) {
	cmd := &qitmeerjson.DumpPrivKeyCmd{
		Address: address,
	}
	msg, err := walletrpc.DumpPrivKey(cmd, w)
	if err != nil {
		fmt.Println("dumpPrivKey", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func getAccountAndAddress() (interface{}, error) {
	msg, err := walletrpc.GetAccountAndAddress(w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	a := msg.([]wallet.AccountAndAddressResult)
	for _, result := range a {
		for _, r := range result.AddrsOutput {
			fmt.Printf("account:%s,address:%s\n", result.AccountName, r.Addr)
		}
	}
	return msg, nil
}
func sendToAddress(address string, amount float64, coin string) (interface{}, error) {
	cmd := &qitmeerjson.SendToAddressCmd{
		Address: address,
		Amount:  amount,
		Coin:    coin,
	}
	msg, err := walletrpc.SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("sendToAddress:", "error", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func sendLockedToAddress(address string, amount float64, lockedHeight uint64, coin string) (interface{}, error) {
	cmd := &qitmeerjson.SendLockedToAddressCmd{
		Address:      address,
		Amount:       amount,
		Coin:         coin,
		LockedHeight: lockedHeight,
	}
	msg, err := walletrpc.SendLockedToAddress(cmd, w)
	if err != nil {
		fmt.Println("sendToAddress:", "error", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func updateblock(height int64) error {
	cmd := &qitmeerjson.UpdateBlockToCmd{
		ToOrder: height,
	}
	err := walletrpc.UpdateBlock(cmd, w)
	if err != nil {
		fmt.Println("updateblock:", "error", err.Error())
		return err
	}
	return nil
}
func syncheight() error {
	fmt.Printf("%d\n", w.Manager.SyncedTo().Order)
	return nil
}

func ClearTxData() error {
	err := walletrpc.ClearTxData(w)
	if err != nil {
		fmt.Println("clearTxData:", "error", err.Error())
		return err
	}
	return nil
}

func mnemonicToSeed(mnemonic string) (string, error) {
	seedBuf, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", seedBuf[:]), nil
}

func mnemonicToAddr(mnemonic string, network string) (string, error) {
	seed, err := mnemonicToSeed(mnemonic)
	if err != nil {
		return "", err
	}
	return seedToAddr(seed, network)
}

func seedToAddr(seed string, network string) (string, error) {
	pri, err := qx.EcNew("secp256k1", seed)
	if err != nil {
		return "", err
	}
	return priToAddr(pri, network)
}
func priToAddr(pri string, network string) (string, error) {
	pub, err := priToPub(pri, false)
	if err != nil {
		return "", err
	}
	addr, err := pubToAddr(pub, network)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func priToPub(pri string, uncompressed bool) (string, error) {
	msg, err := qx.EcPrivateKeyToEcPublicKey(uncompressed, pri)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", msg), nil
}
func pubToAddr(pub string, net string) (string, error) {

	serializedPubKey, err := hex.DecodeString(pub)
	pubkey, err := secp256k1.ParsePubKey(serializedPubKey)
	msg, err := qx.EcPubKeyToAddress(net, hex.EncodeToString(pubkey.SerializeCompressed()))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", msg), nil
}

func unlock(password string) error {
	err := walletrpc.Unlock(password, w)
	if err != nil {
		fmt.Println("unlock", "err", err.Error())
		return err
	} else {
		fmt.Println("unlock succ")
	}
	return nil
}
