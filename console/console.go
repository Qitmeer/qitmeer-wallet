package console

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer/crypto/bip39"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"

	//"github.com/Qitmeer/qitmeer/crypto/bip39"
	//"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	//"github.com/Qitmeer/qitmeer/crypto/seed"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	"github.com/Qitmeer/qitmeer-wallet/rpc/walletrpc"
	"github.com/Qitmeer/qitmeer-wallet/util"
	qjson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer/qx"
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
		fmt.Println("db is exist",filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
		return
	}else{
		_,err:=createWallet()
		if err!=nil{
			fmt.Println("createWallet err:",err.Error())
			return
		}else{
			fmt.Println("createWallet succ")
		}
		return
	}
}

func OpenWallet() error{
	b:=checkWalletIeExist(config.Cfg)
	var err error
	if b {
		load := wallet.NewLoader(config.ActiveNet, networkDir(config.Cfg.AppDataDir, config.ActiveNet), 250,config.Cfg)
		w, err = load.OpenExistingWallet([]byte(config.Cfg.WalletPass), false)
		if err != nil {
			return fmt.Errorf("openWallet err:%s\n", err.Error())
		}
	}else{
		return fmt.Errorf("Please create a wallet first,[qitmeer-wallet create ]")
	}
	return nil
}

func UnLock(password string) error{
	err := w.UnLockManager([]byte(password))
	if err != nil {
		return fmt.Errorf("password error")
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
			fmt.Println("openWallet err:", err.Error())
			return
		}
	}else{
		fmt.Println("Please create a wallet first,[qitmeer-wallet create --create]")
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
			fmt.Println("openWallet err:", err.Error())
			return
		}
	}else{
		w,err=createWallet()
		if err!=nil{
			fmt.Println("createWallet err:",err.Error())
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
				fmt.Println("getbalance err ","err",err.Error())
				break
			}
			company:="i"
			b,err:=getbalance(Default_minconf,arg1)
			if err!=nil{
				fmt.Println(err.Error())
				return
			}
			if arg2!="" &&  arg2 !="i"{
				company="f"
			}
			if company == "i"{
				fmt.Printf("%s\n",b.UnspendAmount.String())
			}else{
				fmt.Printf("%f\n",b.UnspendAmount.ToCoin())
			}
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
			fmt.Printf("Wrong command %s\n " , cmd)
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
	fmt.Println("")
}

func printPrompt() (cmd string, arg1 string, arg2 string) {
	if isWin {
		fmt.Printf("[%s]:", Name)
	} else {
		//fmt.Printf("[%s]:", Name)
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
func checkWalletIeExist(cfg *config.Config) bool{
	netDir := networkDir(cfg.AppDataDir, config.ActiveNet)
	err:=checkCreateDir(netDir)
	if err!=nil{
		return false
	}
	dbPath := filepath.Join(netDir, config.WalletDbName)
	if fi,err:=util.FileExists(dbPath);err!=nil{
		fmt.Println("FileExists ","err",err.Error())
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
		fmt.Println("createNewAccount","err", err.Error())
		return err
	}
	fmt.Printf("%s",msg)
	return nil
}
func getbalance(minconf int ,addr string) ( *wallet.Balance, error){
	cmd:=&qitmeerjson.GetBalanceByAddressCmd{
		Address:addr,
		MinConf:minconf,
	}
	b,err:=walletrpc.Getbalance(cmd,w)
	if(err!=nil){
		fmt.Println("getbalance","err",err.Error())
		return nil,err
	}
	r:=b.(*wallet.Balance)
	//fmt.Println("getbalance :",b)
	//fmt.Println("getbalance  ConfirmAmount:",r.ConfirmAmount)
	//fmt.Printf("%d\n",r.UnspendAmount)
	//fmt.Println("getbalance  SpendAmount:",r.SpendAmount)
	//fmt.Println("getbalance  TotalAmount:",r.TotalAmount)
	return r, nil
}
func  listAccountsBalance(min int)( interface{}, error){
	cmd:=&qitmeerjson.ListAccountsCmd{
		MinConf:&min,
	}
	msg, err := walletrpc.ListAccounts(cmd, w)
	if err != nil {
		fmt.Println("listAccountsBalance","err", err.Error())
		return nil, err
	}
	fmt.Printf("%v\n",msg)
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
		fmt.Println("getlisttxbyaddr","err", err.Error())
		return nil, err
	}else{
		a:=result.(*qjson.PageTxRawResult)
		for _, t := range a.Transactions {
			b,err:=json.Marshal(t)
			if err!=nil{
				fmt.Println("getlisttxbyaddr err:",err.Error())
				return nil, err
			}
			fmt.Printf("%s\n",string(b))
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
		fmt.Println("getNewAddress","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
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
		fmt.Println("getAddressesByAccount","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
	return msg, nil
}
func getAccountByAddress(address string) (interface{}, error) {
	cmd := &qitmeerjson.GetAccountCmd{
		Address: address,
	}
	msg, err := walletrpc.GetAccountByAddress(cmd, w)
	if err != nil {
		fmt.Println("getAccountByAddress","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
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
		fmt.Println("importPrivKey","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
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
		fmt.Println("importWifPrivKey","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
	return msg, nil
}
func dumpPrivKey(address string) (interface{}, error) {
	cmd := &qitmeerjson.DumpPrivKeyCmd{
		Address: address,
	}
	msg, err := walletrpc.DumpPrivKey(cmd, w)
	if err != nil {
		fmt.Println("dumpPrivKey","err", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
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
			fmt.Printf("account:%s,address:%s\n",result.AccountName,r.Addr)
		}
	}
	return msg, nil
}
func sendToAddress(address string ,amount float64)( interface{}, error){
	cmd:=&qitmeerjson.SendToAddressCmd{
		Address: address,
		Amount :   amount,
	}
	msg, err := walletrpc.SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("sendToAddress:","error", err.Error())
		return nil, err
	}
	fmt.Printf("%s\n",msg)
	return msg, nil
}
func updateblock(height int64)(  error){
	cmd:=&qitmeerjson.UpdateBlockToCmd{
		Toheight:height,
	}
	err := walletrpc.Updateblock(cmd, w)
	if err != nil {
		fmt.Println("updateblock:","error", err.Error())
		return err
	}
	return nil
}
func syncheight()(  error){
	fmt.Printf("%d\n",w.Manager.SyncedTo().Height)
	return nil
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
		fmt.Println("unlock","err", err.Error())
		return err
	}else{
		fmt.Println("unlock succ")
	}
	return nil
}