package console

import (
	"encoding/json"
	"fmt"
	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/json/qitmeerjson"
	"github.com/HalalChain/qitmeer-wallet/rpc/walletrpc"
	"github.com/HalalChain/qitmeer-wallet/util"
	qjson "github.com/HalalChain/qitmeer-wallet/json"
	"github.com/HalalChain/qitmeer-wallet/wallet"
	"log"
	"path/filepath"
	"runtime"
	"strconv"
)

const (
	Name     = "wallet-cli"
)
var w *wallet.Wallet
var isWin = runtime.GOOS == "windows"

func StartConsole()  {
	log.Println("config.Cfg.AppDataDir：",config.Cfg.AppDataDir)
	b:=checkWalletIeExist(config.Cfg)
	var err error
	if b {
		log.Println("db is exist",filepath.Join(networkDir(config.Cfg.AppDataDir, config.ActiveNet), config.WalletDbName))
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
	}
	for {
		cmd, arg1, arg2 := printPrompt()
		fmt.Println("arg1:",arg1,"arg2:",arg2)
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
		case "getbalance":
			iarg1,err:=strconv.Atoi(arg1)
			if(err!=nil){
				fmt.Print("getbalance err :",err.Error())
				break
			}
			getbalance(iarg1,arg2)
			break
		case "listAccountsBalance":
			iarg1,err:=strconv.Atoi(arg1)
			if(err!=nil){
				fmt.Print("listAccountsBalance err :",err.Error())
				break
			}
			listAccountsBalance(iarg1)
			break
		case "getlisttxbyaddr":
			getlisttxbyaddr(arg1)
			break
		case "getNewAddress":
			getNewAddress(arg1)
			break
		case "getAddressesByAccount":
			getAddressesByAccount(arg1)
			break
		case "getAccountByAddress":
			getAccountByAddress(arg1)
			break
		case "importPrivKey":
			importPrivKey(arg1)
			break
		case "importWifPrivKey":
			importWifPrivKey(arg1)
			break
		case "dumpPrivKey":
			dumpPrivKey(arg1)
			break
		case "getAccountAndAddress":
			i32,err := strconv.ParseInt(arg1,10,32)
			if(err!=nil){
				fmt.Print("getAccountAndAddress err :",err.Error())
				break
			}
			getAccountAndAddress(int32(i32))
			break
		case "sendToAddress":
			i32,err := strconv.ParseInt(arg2,10,32)
			if(err!=nil){
				fmt.Print("getAccountAndAddress err :",err.Error())
				break
			}
			sendToAddress(arg1,int32(i32))
			break
		case "updateblock":
			updateblock()
			break
		//case "help":
		//	printHelp()
		default:
			printError("Wrong command " + cmd)
			break
		}
	}
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	r:=b.(*wallet.Balance)
	fmt.Println("getbalance :",b)
	fmt.Println("getbalance  ConfirmAmount:",r.ConfirmAmount)
	fmt.Println("getbalance  UnspendAmount:",r.UnspendAmount)
	fmt.Println("getbalance  SpendAmount:",r.SpendAmount)
	fmt.Println("getbalance  TotalAmount:",r.TotalAmount)
	return b, nil
}
func  listAccountsBalance(min int)( interface{}, error){
	cmd:=&qitmeerjson.ListAccountsCmd{
		MinConf:&min,
	}
	msg, err := walletrpc.ListAccounts(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
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
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("dumpPrivKey :",msg)
	return msg, nil
}
func getAccountAndAddress(minconf int32) (interface{}, error) {
	msg, err := walletrpc.GetAccountAndAddress(w, minconf)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	a:=msg.([]wallet.AccountAndAddressResult)
	fmt.Println("getAccountAndAddress :",a)
	fmt.Println("getAccountAndAddress :",a[1].AddrsOutput[0].Addr)
	return msg, nil
}
func sendToAddress(address string ,amount int32)( interface{}, error){
	cmd:=&qitmeerjson.SendToAddressCmd{
		Address:"address",
		Amount :   float64(amount),
	}
	msg, err := walletrpc.SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("sendToAddress :",msg)
	return msg, nil
}
func updateblock()(  error){
	cmd:=&qitmeerjson.UpdateBlockToCmd{
		Toheight:0,
	}
	err := walletrpc.Updateblock(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return err
	}
	return nil
}