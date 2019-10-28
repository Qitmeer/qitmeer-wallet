package walletrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	//"os"
	//"path/filepath"
	"testing"

	"github.com/Qitmeer/qitmeer-lib/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	util "github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	qjson "github.com/Qitmeer/qitmeer-wallet/json"
	//"time"
	//"time"
)

//func TestListAccounts(t *testing.T) {
//	w, err := open_wallet()
//	if err != nil {
//		t.Log("open wallet err", err)
//		return
//	}
//
//	l, err := test_wallet_listAccounts(w)
//	if err != nil {
//		t.Log(err)
//		return
//	}
//
//	t.Log(l)
//
//}

func open_wallet() (*wallet.Wallet, error) {
	//dbpath, _ := os.Getwd() //  "C:\\Users\\luoshan\\AppData\\Local\\Qitwallet\\testnet"
	dbpath:="C:\\Users\\luoshan\\AppData\\Local\\Qitwallet\\testnet"
	//dbpath = filepath.Join(dbpath, "testnet")
	activeNet := &params.TestNetParams
	load := wallet.NewLoader(activeNet, dbpath, 250,nil)
	w, err := load.OpenExistingWallet([]byte("public"), false)
	if err != nil {
		fmt.Println("openWallet err:", err.Error())
		return nil, err
	}
	//w.Start()
	//err=w.Unlock([]byte("123456"),time.After(10*time.Minute))
	//if err!=nil{
	//	fmt.Println("err:",err.Error())
	//	return nil,err
	//}
	//err = w.UnLockManager([]byte("123456"))
	err = w.UnLockManager([]byte("123456"))
	if err != nil {
		fmt.Errorf("UnLockManager err:%s", err.Error())
		return nil, err
	}
	w.Httpclient ,err= wallet.NewHtpc("admin", "123456", "47.88.220.44:8131",
		"", true, true, "", "", "")
	if err!=nil{
		fmt.Errorf("NewHtpc err:%s", err.Error())
		return nil, err
	}
	return w, nil
}
func test_wallet_createNewAccount(w *wallet.Wallet) error {
	cmd := &qitmeerjson.CreateNewAccountCmd{
		Account: "luoshan4",
	}
	msg, err := CreateNewAccount(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return err
	}
	fmt.Println("test_wallet_createNewAccount :",msg)
	return nil
}
func test_wallet_getbalance(w *wallet.Wallet) ( interface{}, error){
	minconf:=3
	cmd:=&qitmeerjson.GetBalanceByAddressCmd{
		//Address:"Tmjc34zWMTAASHTwcNtPppPujFKVK5SeuaJ",
		//Address:"TmcAh3FGNCEZMNtmU6RWme18D5GxQGwE3xb",
		Address:"TmaTi4yt947FXPcWTAkMNDqtRELKceEFBb5",
		//Address:"TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF",
		MinConf:minconf,
	}
	b,err:=Getbalance(cmd,w)
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return nil,err
	}
	r:=b.(*wallet.Balance)
	fmt.Println("test_wallet_getbalance :",b)
	fmt.Println("test_wallet_getbalance  ConfirmAmount:",r.ConfirmAmount)
	fmt.Println("test_wallet_getbalance  UnspendAmount:",r.UnspendAmount)
	fmt.Println("test_wallet_getbalance  SpendAmount:",r.SpendAmount)
	fmt.Println("test_wallet_getbalance  TotalAmount:",r.TotalAmount)
	return b, nil
}
func test_wallet_listAccounts(w *wallet.Wallet)( interface{}, error){
	min:=3
	cmd:=&qitmeerjson.ListAccountsCmd{
		MinConf:&min,
	}
	msg, err := ListAccounts(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_listAccounts :",msg)
	return msg, nil
}

func test_wallet_getlisttxbyaddr(w *wallet.Wallet)( interface{}, error){
	cmd:=&qitmeerjson.GetListTxByAddrCmd{
		Address:"TmYaYXRU58ppifMLwqsk6YRPQDrEvdm4dW1",
		Stype:int32(0),
		Page:int32(1),
		PageSize:int32(100),
	}
	result, err := Getlisttxbyaddr(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	if(err!=nil){
		fmt.Errorf("test_wallet_getlisttxbyaddr err:%s",err.Error())
	}else{
		a:=result.(*qjson.PageTxRawResult)
		fmt.Println("test_wallet_getlisttxbyaddr msg a.Total:",a.Total)
		for _, t := range a.Transactions {
			b,err:=json.Marshal(t)
			if err!=nil{
				fmt.Println("test_wallet_getlisttxbyaddr err:",err.Error())
				return nil, err
			}
			fmt.Println("test_wallet_getlisttxbyaddr :",string(b))
		}
	}
	return result, nil
}

func test_wallet_getNewAddress(w *wallet.Wallet) (interface{}, error) {
	//account := "default"
	account := "imported"
	cmd := &qitmeerjson.GetNewAddressCmd{
		Account: &account,
	}
	msg, err := GetNewAddress(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_getNewAddress :",msg)
	return msg, nil
}
func test_wallet_getAddressesByAccount(w *wallet.Wallet) (interface{}, error) {
	//account := "default"
	account := "imported"
	cmd := &qitmeerjson.GetAddressesByAccountCmd{
		Account: account,
	}
	msg, err := GetAddressesByAccount(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_getAddressesByAccount :",msg)
	return msg, nil
}
func test_wallet_getAccountByAddress(w *wallet.Wallet) (interface{}, error) {
	address := "TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
	cmd := &qitmeerjson.GetAccountCmd{
		Address: address,
	}
	msg, err := GetAccountByAddress(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_getAccountByAddress :",msg)
	return msg, nil
}
func test_wallet_importPrivKey(w *wallet.Wallet) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: "7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b",
		Rescan:  &v,
	}
	msg, err := ImportPrivKey(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_importPrivKey :",msg)
	return msg, nil
}
func test_wallet_importWifPrivKey(w *wallet.Wallet) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: "9QwXzXVQBFNm1fxP8jCqHJG9jZKjqrUKjYiTvaRxEbFobiNrvzhgZ",
		Rescan:  &v,
	}
	msg, err := ImportWifPrivKey(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	return msg, nil
}
func test_wallet_dumpPrivKey(w *wallet.Wallet) (interface{}, error) {
	address := "TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
	cmd := &qitmeerjson.DumpPrivKeyCmd{
		Address: address,
	}
	msg, err := DumpPrivKey(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_dumpPrivKey :",msg)
	return msg, nil
}
func test_wallet_getAccountAndAddress(w *wallet.Wallet) (interface{}, error) {
	msg, err := GetAccountAndAddress(w, 16)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	a:=msg.([]wallet.AccountAndAddressResult)
	fmt.Println("test_wallet_getAccountAndAddress :",a)
	fmt.Println("test_wallet_getAccountAndAddress :",a[1].AddrsOutput[0].Addr)
	return msg, nil
}
func test_wallet_sendToAddress(w *wallet.Wallet)( interface{}, error){
	cmd:=&qitmeerjson.SendToAddressCmd{
		Address:"TmZQiY7WZarVk6Fax1NgUJCoVmonrEFRzwy",
		//Address:"TmbCBKbZF8PeSdj5Chm22T4hZRMJY5D8XyX",
		//Address:"TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF",
		Amount :   float64(31),
	}
	msg, err := SendToAddress(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return nil, err
	}
	fmt.Println("test_wallet_sendToAddress :",msg)
	return msg, nil
}
func test_wallet_updateblock(w *wallet.Wallet)(  error){
	cmd:=&qitmeerjson.UpdateBlockToCmd{
		Toheight:0,
	}
	err := Updateblock(cmd, w)
	if err != nil {
		fmt.Println("errr:", err.Error())
		return err
	}
	return nil
}

func test_wif(w *wallet.Wallet) error {
	pri := "7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b"
	data, err := hex.DecodeString(pri)
	if err != nil {
		return err
	}
	privkey, _ := secp256k1.PrivKeyFromBytes(data)
	wif, err := util.NewWIF(privkey, w.ChainParams(), true)
	if err != nil {
		return err
	}
	fmt.Println("wif:", wif)
	wif1, err := util.DecodeWIF(wif.String(), w.ChainParams())
	if err != nil {
		return err
	}
	fmt.Println("wif1:", wif1)
	fmt.Printf("wif1:%x\n", wif1.PrivKey.SerializeSecret())
	return nil
}

func TestWallet_Method(t *testing.T) {
	serializedPubKey, err := hex.DecodeString("049ed083b4d2a923c9736c3186444610de68576e4314b03e2b638b4738a1741c460f23f8921dbdf6cbf0a3b67fc53734ad5c05116f55d41abc8d78f10aa1e4f2a2")
	pu,err:=secp256k1.ParsePubKey(serializedPubKey)
	if err!=nil{
		fmt.Print(err.Error())
		return
	}
	fmt.Printf("pu.SerializeCompressed()：%x\n",pu.SerializeCompressed())
	fmt.Printf("pu.SerializeUncompressed()：%x\n",pu.SerializeUncompressed())

	//w, err := open_wallet()
	//if err != nil {
	//	fmt.Println("open_wallet err:", err)
	//	return
	//}
	//w.UpdateMempool()
	//
	//test_wallet_createNewAccount(w)
	//
	//
	//test_wallet_importPrivKey(w)
	//
	//
	//test_wallet_getNewAddress(w)
	//test_wallet_getAddressesByAccount(w)
	//test_wallet_listAccounts(w)
	//
	//
	//
	//
	//
	//
	//test_wallet_getAccountByAddress(w)
	//
	//
	//test_wallet_dumpPrivKey(w)
	//
	//
	//
	//test_wallet_getbalance(w)
	//
	//
	//test_wallet_sendToAddress(w)
	//
	//
	//test_wif(w)
	//
	//test_wallet_getAccountAndAddress(w)
	//
	//
	//test_wallet_getlisttxbyaddr(w)

	//test_wallet_updateblock(w)
	//str,err:=w.GetTx("e44b7a7c361c7f220811f07a6c051ea95967c56dff0d255e62c29908597c320d")
	////str,err:=w.GetTx("2c0cbf455ee3ae055261db248efa136e09c9742634b1a769c6f1be49c4a689f0")
	//if(err!=nil){
	//	fmt.Println("GetTx err:",err.Error())
	//	return
	//}
	//fmt.Println("GetTx :",str)
}
