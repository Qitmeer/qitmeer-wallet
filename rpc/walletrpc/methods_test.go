package walletrpc
import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-wallet/json/qitmeerjson"
	"github.com/HalalChain/qitmeer-wallet/wallet"
	"testing"
	//"time"

	//"time"
)

func open_wallet() (*wallet.Wallet,error) {
	dbpath:="C:\\Users\\luoshan\\AppData\\Local\\Qitwallet\\testnet"
	activeNet:=&params.TestNetParams
	load:=wallet.NewLoader(activeNet,dbpath,250)
	w,err:=load.OpenExistingWallet([]byte("public"),false)
	if(err!=nil){
		fmt.Println("openWallet err:",err.Error())
		return nil,err
	}
	err=w.UnLockManager([]byte("123456"))
	if(err!=nil){
		fmt.Errorf("UnLockManager err:%s",err.Error())
		return nil,err
	}
	return w,nil
}
func test_wallet_createNewAccount(w *wallet.Wallet) error{
	cmd:=&qitmeerjson.CreateNewAccountCmd{
		Account:"luoshan4",
	}
	_,err:=createNewAccount(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return err
	}
	return nil
}
func test_wallet_listAccounts(w *wallet.Wallet)( interface{}, error){
	min:=16
	cmd:=&qitmeerjson.ListAccountsCmd{
		MinConf:&min,
	}
	msg,err:=listAccounts(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_getNewAddress(w *wallet.Wallet)( interface{}, error){
	account:="luoshan"
	cmd:=&qitmeerjson.GetNewAddressCmd{
		Account:&account,
	}
	msg,err:=getNewAddress(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_getAddressesByAccount(w *wallet.Wallet)( interface{}, error){
	account:="luoshan1"
	cmd:=&qitmeerjson.GetAddressesByAccountCmd{
		Account:account,
	}
	msg,err:=getAddressesByAccount(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_getAccountByAddress(w *wallet.Wallet)( interface{}, error){
	address:="TmT5dipuqvrWR2cSF4rFgRDsAAQXLh6qw3S"
	cmd:=&qitmeerjson.GetAccountCmd{
		Address:address,
	}
	msg,err:=getAccountByAddress(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_dumpPrivKey(w *wallet.Wallet)( interface{}, error){
	address:="TmT5dipuqvrWR2cSF4rFgRDsAAQXLh6qw3S"
	cmd:=&qitmeerjson.DumpPrivKeyCmd{
		Address:address,
	}
	msg,err:=dumpPrivKey(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}

func TestWallet_Method(t *testing.T) {
	w,err:=open_wallet()
	if(err!=nil){
		fmt.Println("open_wallet err:",err)
		return
	}
	//err=test_wallet_createNewAccount(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_createNewAccount err:%s",err.Error())
	//}
	//m,err:=test_wallet_listAccounts(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_listAccounts err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_listAccounts :",m)
	//}

	//address,err:=test_wallet_getNewAddress(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_getNewAddress err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_getNewAddress :",address)
	//}

	//adds,err:=test_wallet_getAddressesByAccount(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_getAddressesByAccount err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_getAddressesByAccount :",adds)
	//}

	//account,err:=test_wallet_getAccountByAddress(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_getAccountByAddress err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_getAccountByAddress :",account)
	//}

	pri,err:=test_wallet_dumpPrivKey(w)
	if(err!=nil){
		fmt.Errorf("test_wallet_dumpPrivKey err:%s",err.Error())
	}else{
		fmt.Println("test_wallet_dumpPrivKey :",pri)
	}

}
