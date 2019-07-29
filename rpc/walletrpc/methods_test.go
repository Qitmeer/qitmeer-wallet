package walletrpc
import (
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-wallet/json/qitmeerjson"
	"github.com/HalalChain/qitmeer-wallet/util"
	"github.com/HalalChain/qitmeer-wallet/wallet"
	"github.com/HalalChain/qitmeer-lib/crypto/ecc/secp256k1"
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
	account:="luoshan"
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
	address:="TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
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
func test_wallet_importPrivKey(w *wallet.Wallet)( interface{}, error){
	v:=false
	cmd:=&qitmeerjson.ImportPrivKeyCmd{
		PrivKey :"9QwXzXVQBFNm1fxP8jCqHJG9jZKjqrUKjYiTvaRxEbFobiNrvzhgZ",
		Rescan:&v,
		//PrivKey :"7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b",
	}
	msg,err:=importPrivKey(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_dumpPrivKey(w *wallet.Wallet)( interface{}, error){
	address:="TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
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
func test_wallet_getAccountAndAddress(w *wallet.Wallet)( interface{}, error){
	msg,err:=getAccountAndAddress(w,16)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}
func test_wallet_sendToAddress(w *wallet.Wallet)( interface{}, error){
	cmd:=&qitmeerjson.SendToAddressCmd{
		Address:"TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF",
		Amount :   float64(10),
	}
	msg,err:=sendToAddress(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	return msg,nil
}

func test_wif(w *wallet.Wallet) error{
	pri:="7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b"
	data, err := hex.DecodeString(pri)
	if err!=nil {
		return err
	}
	privkey, _ := secp256k1.PrivKeyFromBytes(data)
	wif,err:=util.NewWIF(privkey,w.ChainParams(),true)
	if err!=nil {
		return err
	}
	fmt.Println("wif:",wif)
	wif1,err:=util.DecodeWIF(wif.String(),w.ChainParams())
	if err!=nil {
		return err
	}
	fmt.Println("wif1:",wif1)
	fmt.Printf("wif1:%x\n",wif1.PrivKey.SerializeSecret())
	return nil
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


	m,err:=test_wallet_listAccounts(w)
	if(err!=nil){
		fmt.Errorf("test_wallet_listAccounts err:%s",err.Error())
	}else{
		fmt.Println("test_wallet_listAccounts :",m)
	}

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

	//pri,err:=test_wallet_dumpPrivKey(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_dumpPrivKey err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_dumpPrivKey :",pri)
	//}

	//result,err:=test_wallet_importPrivKey(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_importPrivKey err:%s",err.Error())
	//}else{
	//	fmt.Println("test_wallet_importPrivKey :",result)
	//}

	result,err:=test_wallet_sendToAddress(w)
	if(err!=nil){
		fmt.Errorf("test_wallet_sendToAddress err:%s",err.Error())
	}else{
		fmt.Println("test_wallet_sendToAddress :",result)
	}


	//err=test_wif(w)
	//if(err!=nil){
	//	fmt.Println("test_wif err:",err.Error())
	//}

	//result,err:=test_wallet_getAccountAndAddress(w)
	//if(err!=nil){
	//	fmt.Errorf("test_wallet_getAccountAndAddress err:%s",err.Error())
	//}else{
	//	a:=result.([]wallet.AccountAndAddressResult)
	//	fmt.Println("test_wallet_getAccountAndAddress :",a)
	//	fmt.Println("test_wallet_getAccountAndAddress :",a[1].AddrsOutput[0].Addr)
	//	//fmt.Println("test_wallet_getAccountAndAddress :",a[1].Addrs[1].ScriptAddress())
	//	//fmt.Println("test_wallet_getAccountAndAddress :",a[1].Addrs[2].Hash160())
	//}
}
