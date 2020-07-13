package commands

import (
	"encoding/hex"
	"encoding/json"
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
	Args:             cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		b := checkWalletIeExist(config.Cfg)
		if b == false {
			fmt.Println("Please create a wallet first,[qitmeer-wallet qc create ]")
			return
		}
		startConsole()
	},
}

func AddConsoleCommand() {

}

func CreatWallet() {
	b := checkWalletIeExist(config.Cfg)
	if b {
		fmt.Println("db is exist", filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
		return
	} else {
		_, err := createWallet()
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

func startConsole() {
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
		w, err = createWallet()
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
	//fmt.Println("\t<listAccountsBalance> : Obtain all account balances. Parameter: []")
	fmt.Println("\t<getListTxByAddr> : Gets all transaction records at the specified address. Parameter: [address] [stype:in,out,all]")
	fmt.Println("\t<getNewAddress> : Create a new address under the account. Parameter: [account]")
	fmt.Println("\t<getAddressesByAccount> : Check all addresses under the account. Parameter: [account]")
	fmt.Println("\t<getAccountByAddress> : Inquire about the account number of the address. Parameter: [address]")
	fmt.Println("\t<importPrivKey> : Import private key. Parameter: [priKey]")
	fmt.Println("\t<importWifPrivKey> : Import wif format private key. Parameter: [priKey]")
	fmt.Println("\t<dumpPrivKey> : Export wif format private key by address. Parameter: [address]")
	fmt.Println("\t<getAccountAndAddress> : Check all accounts and addresses. Parameter: []")
	fmt.Println("\t<sendToAddress> : Transfer transaction. Parameter: [address] [num]")
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
func getBalance(addr string) (*wallet.Balance, error) {
	cmd := &qitmeerjson.GetBalanceByAddressCmd{
		Address: addr,
	}
	b, err := walletrpc.GetBalance(cmd, w)
	if err != nil {
		fmt.Println("getBalance", "err", err.Error())
		return nil, err
	}
	r := b.(*wallet.Balance)
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
	msg, err := walletrpc.ListAccounts(w)
	if err != nil {
		fmt.Println("listAccountsBalance", "err", err.Error())
		return nil, err
	}
	for k, v := range msg.(map[string]float64) {
		fmt.Printf("%v:%v\n", k, v)
	}
	return msg, nil
}

func getListTxByAddr(addr string, page int32, pageSize int32, sType int32) (interface{}, error) {
	helper = &JsonCmdHelper{
		JsonCmd: &qitmeerjson.GetListTxByAddrCmd{
			Address:  addr,
			Stype:    sType,
			Page:     page,
			PageSize: pageSize,
		},
		Run: func(cmd interface{}, w *wallet.Wallet) (interface{}, error) {
			return walletrpc.GetListTxByAddr(cmd, w)
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

func SetSynceToNum(order int64) error {
	return walletrpc.SetSynceToNum(order, w)
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
func getTx(txid string) (interface{}, error) {
	msg, err := walletrpc.GetTx(txid, w)
	if err != nil {
		fmt.Println("getTx", "err", err.Error())
		return nil, err
	}
	b, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("getTx", "err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", b)
	return msg, nil
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
func sendToAddress(address string, amount float64) (interface{}, error) {
	cmd := &qitmeerjson.SendToAddressCmd{
		Address: address,
		Amount:  amount,
	}
	msg, err := walletrpc.SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("sendToAddress:", "error", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n", msg)
	return msg, nil
}
func updateblock(height int64) error {
	cmd := &qitmeerjson.UpdateBlockToCmd{
		Toheight: height,
	}
	err := walletrpc.UpdateBlock(cmd, w)
	if err != nil {
		fmt.Println("updateblock:", "error", err.Error())
		return err
	}
	return nil
}
func syncheight() error {
	fmt.Printf("%d\n", w.Manager.SyncedTo().Height)
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
