package console

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/crypto/bip39"
	"github.com/Qitmeer/qitmeer-lib/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer-lib/crypto/seed"
	"github.com/Qitmeer/qitmeer-lib/qx"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/rpc/walletrpc"
	"github.com/Qitmeer/qitmeer-wallet/util"
	qjson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	Name     = "wallet-cli"
	Default_minconf =16
)
var w *wallet.Wallet
var isWin = runtime.GOOS == "windows"

func CreatWallet()  {
	b:=checkWalletIeExist(config.Cfg)
	if b {
		log.Fatalln("db is exist",filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
		return
	}else{
		_,err:=createWallet()
		if err!=nil{
			log.Fatalln("createWallet err:",err.Error())
			return
		}else{
			log.Println("createWallet succ")
		}
		return
	}
}

func OpenWallet(){
	b:=checkWalletIeExist(config.Cfg)
	var err error
	if b {
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250,config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			log.Fatalln("openWallet err:", err.Error())
			return
		}
	}else{
		log.Fatalln("Please create a wallet first,[qitmeer-wallet create ]")
		return
	}
}

func UnLock(password string) error{
	err := w.UnLockManager([]byte(password))
	if err != nil {
		log.Fatalf("UnLock err:%s", "password error")
		return err
	}
	return nil
}

func InitWallet(){
	//log.Println("config.Cfg.AppDataDir：",config.Cfg.AppDataDir)
	b:=checkWalletIeExist(config.Cfg)
	var err error
	if b {
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250,config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			log.Fatalln("openWallet err:", err.Error())
			return
		}
	}else{
		log.Fatalln("Please create a wallet first,[qitmeer-wallet create --create]")
		return
	}
}
func startConsole()  {
	b:=checkWalletIeExist(config.Cfg)
	var err error
	if b {
		//log.Println("db is exist",filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250,config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			log.Println("openWallet err:", err.Error())
			return
		}
	}else{
		w,err=createWallet()
		if err!=nil{
			log.Println("createWallet err:",err.Error())
			return
		}
		return
	}
	for {
		cmd, arg1, arg2 := printPrompt()
		//fmt.Println("arg1:",arg1,"arg2:",arg2)
		if cmd == "exit" {
			break
		}
		if cmd == "re" {
			continue
		}
		switch cmd {
		case "createNewAccount":
			createNewAccount(arg1)
			break
			//Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ
		case "getbalance":
			if arg1==""{
				fmt.Println("Please enter your address.")
				break
			}
			if(err!=nil){
				fmt.Println("getbalance err :",err.Error())
				break
			}
			getbalance(Default_minconf,arg1)
			break
		//case "listAccountsBalance":
		//	listAccountsBalance(Default_minconf)
		//	break
		case "getlisttxbyaddr":
			if arg1 == ""{
				fmt.Println("getlisttxbyaddr err :Please enter your address.")
				break
			}
			getlisttxbyaddr(arg1)
			break
		case "getNewAddress":
			if arg1 == ""{
				fmt.Println("getNewAddress err :Please enter your account.")
				break
			}
			getNewAddress(arg1)
			break
		case "getAddressesByAccount":
			if arg1 == ""{
				fmt.Println("getAddressesByAccount err :Please enter your account.")
				break
			}
			getAddressesByAccount(arg1)
			break
		case "getAccountByAddress":
			if arg1 == ""{
				fmt.Println("getAccountByAddress err :Please enter your address.")
				break
			}
			getAccountByAddress(arg1)
			break
		case "importPrivKey":
			if arg1 == ""{
				fmt.Println("importPrivKey err :Please enter your privkey.")
				break
			}
			importPrivKey(arg1)
			break
		case "importWifPrivKey":
			if arg1 == ""{
				fmt.Println("importWifPrivKey err :Please enter your wif privkey.")
				break
			}
			importWifPrivKey(arg1)
			break
		case "dumpPrivKey":
			if arg1 == ""{
				fmt.Println("dumpPrivKey err :Please enter your address.")
				break
			}
			dumpPrivKey(arg1)
			break
		case "getAccountAndAddress":
			getAccountAndAddress(int32(Default_minconf))
			break
		case "sendToAddress":
			if arg1 =="" {
				fmt.Println("getAccountAndAddress err : Please enter the receipt address.")
				break
			}
			if arg2 == ""{
				fmt.Println("getAccountAndAddress err : Please enter the amount of transfer.")
				break
			}
			f32,err := strconv.ParseFloat(arg2,32)
			if(err!=nil){
				fmt.Println("getAccountAndAddress err :",err.Error())
				break
			}
			sendToAddress(arg1,float64(f32))
			break
		case "updateblock":
			updateblock(0)
			break
		case "syncheight":
			syncheight()
			break
		case "unlock":
			unlock(arg1)
			break
		case "help":
			printHelp()
			break
		default:
			printError("Wrong command " + cmd)
			break
		}
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("\t<command> [arguments]")
	fmt.Println("\tThe commands are:")
	fmt.Println("\t<createNewAccount> : Create a new account. Parameter: [account]")
	fmt.Println("\t<getbalance> : Query the specified address balance. Parameter: [address]")
	//fmt.Println("\t<listAccountsBalance> : Obtain all account balances. Parameter: []")
	fmt.Println("\t<getlisttxbyaddr> : Gets all transaction records at the specified address. Parameter: [address]")
	fmt.Println("\t<getNewAddress> : Create a new address under the account. Parameter: [account]")
	fmt.Println("\t<getAddressesByAccount> : Check all addresses under the account. Parameter: [account]")
	fmt.Println("\t<getAccountByAddress> : Inquire about the account number of the address. Parameter: [address]")
	fmt.Println("\t<importPrivKey> : Import private key. Parameter: [prikey]")
	fmt.Println("\t<importWifPrivKey> : Import wif format private key. Parameter: [prikey]")
	fmt.Println("\t<dumpPrivKey> : Export wif format private key by address. Parameter: [address]")
	fmt.Println("\t<getAccountAndAddress> : Check all accounts and addresses. Parameter: []")
	fmt.Println("\t<sendToAddress> : Transfer transaction. Parameter: [address] [num]")
	fmt.Println("\t<updateblock> : Update Wallet Block. Parameter: []")
	fmt.Println("\t<syncheight> : Current Synchronized Data Height. Parameter: []")
	fmt.Println("\t<unlock> : Unlock Wallet. Parameter: [password]")
	fmt.Println("\t<help> : help")
	fmt.Println("\t<exit> : Exit command mode")
	fmt.Println()
}

func printPrompt() (cmd string, arg1 string, arg2 string) {
	if isWin {
		fmt.Printf("[%s]:", Name)
	} else {
		fmt.Printf("%c[4;%d;%dm[%s]: %c[0m", 0x1B, 0, 30, Name, 0x1B)
	}

	var c, a1, a2 string
	count, err := fmt.Scanln(&c, &a1, &a2)
	if err != nil {

	}
	if count == 0 {
		fmt.Println("there is nothing inputs")
		return "re", a1, a2
	}
	return c, a1, a2
}
func printError(msg string) {
	if isWin {
		fmt.Println("error:", msg)
	} else {
		fmt.Printf("%c[4;0;%dm%s!%c[0m\n", 0x1B, 30, msg, 0x1B)
	}

}
func checkWalletIeExist(cfg *config.Config) bool{
	netDir := networkDir(cfg.AppDataDir, config.ActiveNet)
	err:=checkCreateDir(netDir)
	if err!=nil{
		return false
	}
	dbPath := filepath.Join(netDir, config.WalletDbName)
	if fi,err:=util.FileExists(dbPath);err!=nil{
		log.Println("FileExists err:",err.Error())
		return false
	}else{
		return fi
	}
}


func createNewAccount(arg string) error {
	cmd := &qitmeerjson.CreateNewAccountCmd{
		Account: arg,
	}
	msg, err := walletrpc.CreateNewAccount(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return err
	}
	fmt.Println("createNewAccount :",msg)
	return nil
}
func getbalance(minconf int ,addr string) ( interface{}, error){
	cmd:=&qitmeerjson.GetBalanceByAddressCmd{
		Address:addr,
		MinConf:minconf,
	}
	b,err:=walletrpc.Getbalance(cmd,w)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return nil,err
	}
	r:=b.(*wallet.Balance)
	//fmt.Println("getbalance :",b)
	//fmt.Println("getbalance  ConfirmAmount:",r.ConfirmAmount)
	fmt.Println("getbalance  amount:",r.UnspendAmount)
	//fmt.Println("getbalance  SpendAmount:",r.SpendAmount)
	//fmt.Println("getbalance  TotalAmount:",r.TotalAmount)
	return b, nil
}
func  listAccountsBalance(min int)( interface{}, error){
	cmd:=&qitmeerjson.ListAccountsCmd{
		MinConf:&min,
	}
	msg, err := walletrpc.ListAccounts(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("listAccounts :",msg)
	return msg, nil
}

func getlisttxbyaddr(addr string)( interface{}, error){
	cmd:=&qitmeerjson.GetListTxByAddrCmd{
		Address:addr,
		Stype:int32(0),
		Page:int32(1),
		PageSize:int32(100),
	}
	result, err := walletrpc.Getlisttxbyaddr(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	if(err!=nil){
		fmt.Errorf("getlisttxbyaddr err:%s",err.Error())
	}else{
		a:=result.(*qjson.PageTxRawResult)
		fmt.Println("getlisttxbyaddr msg a.Total:",a.Total)
		for _, t := range a.Transactions {
			b,err:=json.Marshal(t)
			if err!=nil{
				fmt.Println("getlisttxbyaddr err:",err.Error())
				return nil, err
			}
			fmt.Println("getlisttxbyaddr :",string(b))
		}
	}
	return result, nil
}

func getNewAddress(account string) (interface{}, error) {
	if account==""{
		account = "default"
	}
	cmd := &qitmeerjson.GetNewAddressCmd{
		Account: &account,
	}
	msg, err := walletrpc.GetNewAddress(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("getNewAddress :",msg)
	return msg, nil
}
func getAddressesByAccount(account string) (interface{}, error) {
	if account==""{
		account = "default"
	}
	cmd := &qitmeerjson.GetAddressesByAccountCmd{
		Account: account,
	}
	msg, err := walletrpc.GetAddressesByAccount(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("getAddressesByAccount :",msg)
	return msg, nil
}
func getAccountByAddress(address string) (interface{}, error) {
	cmd := &qitmeerjson.GetAccountCmd{
		Address: address,
	}
	msg, err := walletrpc.GetAccountByAddress(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("getAccountByAddress :",msg)
	return msg, nil
}
func importPrivKey(privKey string) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: privKey,
		Rescan:  &v,
	}
	msg, err := walletrpc.ImportPrivKey(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("importPrivKey :",msg)
	return msg, nil
}
func importWifPrivKey(wifprivKey string) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: wifprivKey,
		Rescan:  &v,
	}
	msg, err := walletrpc.ImportWifPrivKey(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	return msg, nil
}
func dumpPrivKey(address string) (interface{}, error) {
	cmd := &qitmeerjson.DumpPrivKeyCmd{
		Address: address,
	}
	msg, err := walletrpc.DumpPrivKey(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("dumpPrivKey :",msg)
	return msg, nil
}
func getAccountAndAddress(minconf int32) (interface{}, error) {
	msg, err := walletrpc.GetAccountAndAddress(w, minconf)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	a:=msg.([]wallet.AccountAndAddressResult)
	for _ , result :=range a{
		for _,r:= range result.AddrsOutput{
			fmt.Println("account:",result.AccountName," ,address:",r.Addr)
		}
	}
	//fmt.Println("getAccountAndAddress :",a[1].AddrsOutput[0].Addr)
	return msg, nil
}
func sendToAddress(address string ,amount float64)( interface{}, error){
	cmd:=&qitmeerjson.SendToAddressCmd{
		Address: address,
		Amount :   amount,
	}
	msg, err := walletrpc.SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return nil, err
	}
	fmt.Println("sendToAddress :",msg)
	return msg, nil
}
func updateblock(height int64)(  error){
	cmd:=&qitmeerjson.UpdateBlockToCmd{
		Toheight:height,
	}
	err := walletrpc.Updateblock(cmd, w)
	if err != nil {
		fmt.Println("err:", err.Error())
		return err
	}
	//fmt.Printf("update to block :%v succ",height)
	return nil
}
func syncheight()(  error){
	fmt.Println(w.Manager.SyncedTo().Height)
	return nil
}

func generateMnemonic()(string,error){
	entropyBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		return "", fmt.Errorf("Generate entropy err: %s", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropyBuf)
	if err != nil {
		return "",err
	}
	return fmt.Sprintf("%s",mnemonic),nil
}

func  mnemonicToSeed(mnemonic string)(string,error){
	seedBuf, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "",err
	}
	return fmt.Sprintf("%x",seedBuf[:]),nil
}

func  mnemonicToAddr(mnemonic string,network string)(string,error){
	seed,err:=mnemonicToSeed(mnemonic)
	if err!=nil{
		return "",err
	}
	return seedToAddr(seed,network)
}

func seedToAddr(seed string,network string)(string,error){
	pri,err:=qx.EcNew("secp256k1",seed)
	if err!=nil{
		return "",err
	}
	return priToAddr(pri,network)
}
func priToAddr(pri string ,network string)(string,error){
	pub,err:=priToPub(pri,false)
	if err!=nil{
		return "",err
	}
	addr,err:=pubToAddr(pub,network)
	if err!=nil{
		return "",err
	}
	return addr,nil
}

func  priToPub(pri string ,uncompressed bool) (string,error) {
	msg,err:=qx.EcPrivateKeyToEcPublicKey(uncompressed,pri)
	if err != nil {
		return "",err
	}
	return fmt.Sprintf("%s",msg),nil
}
func  pubToAddr(pub string ,net string) (string,error) {
	serializedPubKey, err := hex.DecodeString(pub)
	pubkey,err:=secp256k1.ParsePubKey(serializedPubKey)
	msg,err:=qx.EcPubKeyToAddress(net,hex.EncodeToString(pubkey.SerializeCompressed()))
	if err != nil {
		return "",err
	}
	return fmt.Sprintf("%s",msg),nil
}

func unlock(password string) error{
	err :=walletrpc.Unlock(password,w)
	if err != nil {
		fmt.Println("unlock err:", err.Error())
		return err
	}else{
		fmt.Println("unlock succ")
	}
	return nil
}