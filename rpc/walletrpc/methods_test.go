package walletrpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer/log"

	"testing"

	qjson "github.com/Qitmeer/qitmeer-wallet/json"
	"github.com/Qitmeer/qitmeer-wallet/json/qitmeerjson"
	util "github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/params"
)

func TestOpenWallet(t *testing.T) {

}

func TestSetSyncedToNum(t *testing.T) {
	w, err := openWallet()
	if err != nil {
		log.Error("openWallet fail", "err", err.Error())
		return
	}
	err = SetSyncedToNum(100000, w)
	if err != nil {
		log.Error("SetSyncedToNum fail", "err", err.Error())
		return
	}
	fmt.Printf("TestSetSyncedToNum succ")
}
func TestGetSyncHeight(t *testing.T) {
	w, err := openWallet()
	if err != nil {
		log.Error("openWallet fail", "err", err.Error())
		return
	}
	fmt.Printf("TestGetSyncOrder: %v\n", w.Manager.SyncedTo().Order)
}

func openWallet() (*wallet.Wallet, error) {
	dbpath := "/Users/luoshan/Library/Application Support/Qitwallet/testnet"
	tomlpath := "/Users/luoshan/GolandProjects/qitmeer-wallet/config.toml"
	pubpass := "public"
	dbpass := "123456"
	err := config.LoadConfig(tomlpath)
	if err != nil {
		log.Error("TestLoadConfig err", "err", err.Error())
		fmt.Println("TestLoadConfig err :" + err.Error())
		return nil, err
	}
	activeNet := &params.TestNetParams
	load := wallet.NewLoader(activeNet, dbpath, 250, config.Cfg)
	w, err := load.OpenExistingWallet([]byte(pubpass), false)
	if err != nil {
		log.Error("openWallet err", "err", err.Error())
		return nil, err
	}

	err = w.UnLockManager([]byte(dbpass))
	if err != nil {
		fmt.Errorf("UnLockManager err:%s", err.Error())
		return nil, err
	}
	w.HttpClient, err = wallet.NewHtpc(config.Cfg)
	if err != nil {
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
		return err
	}
	log.Info("test_wallet_createNewAccount :", msg)
	return nil
}
func test_wallet_getBalance(w *wallet.Wallet) (*wallet.Balance, error) {
	cmd := &qitmeerjson.GetBalanceByAddressCmd{
		Address: "TmgD1mu8zMMV9aWmJrXqQYnWRhR9SBfDZG6",
		//Address:"TmfDniZnvsjdH98GsH4aetL3XQKFUTWPp4e",
		//Address:"TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF",
	}
	b, err := GetBalance(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	r := b.(*wallet.Balance)
	fmt.Printf("test_wallet_getBalance  UnspendAmount:%v\n", r.UnspentAmount)
	//log.Info("test_wallet_getBalance :",b)
	//log.Info("test_wallet_getBalance  ConfirmAmount:",r.ConfirmAmount)
	//log.Info("test_wallet_getBalance  UnspendAmount:",r.UnspendAmount)
	//log.Info("test_wallet_getBalance  SpendAmount:",r.SpendAmount)
	//log.Info("test_wallet_getBalance  TotalAmount:",r.TotalAmount)
	return r, nil
}
func test_wallet_listAccounts(w *wallet.Wallet) (interface{}, error) {

	msg, err := ListAccounts(w)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func test_wallet_getlisttxbyaddr(w *wallet.Wallet) (interface{}, error) {
	cmd := &qitmeerjson.GetListTxByAddrCmd{
		Address:  "TmYaYXRU58ppifMLwqsk6YRPQDrEvdm4dW1",
		Stype:    int32(0),
		Page:     int32(1),
		PageSize: int32(100),
	}
	result, err := GetListTxByAddr(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	if err != nil {
		fmt.Errorf("test_wallet_getlisttxbyaddr err:%s", err.Error())
	} else {
		a := result.(*qjson.PageTxRawResult)
		log.Info("test_wallet_getlisttxbyaddr msg a.Total:", a.Total)
		for _, t := range a.Transactions {
			b, err := json.Marshal(t)
			if err != nil {
				return nil, err
			}
			log.Info("test_wallet_getlisttxbyaddr ", "result", string(b))
		}
	}
	return result, nil
}

func test_wallet_getBillByAddr(w *wallet.Wallet) (interface{}, error) {
	cmd := &qitmeerjson.GetBillByAddrCmd{
		Address:  "Tme9dVJ4GeWRninBygrA6oDwCAGYbBvNxY7",
		Filter:   int32(0),
		PageNo:   int32(1),
		PageSize: int32(100),
	}
	result, err := GetBillByAddr(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	if err != nil {
		fmt.Errorf("test_wallet_getBillByAddr err:%s", err.Error())
	} else {
		a := result.(*qjson.PagedBillResult)
		log.Info("test_wallet_getBillByAddr msg a.Total:", a.Total)
		for _, b := range a.Bill {
			bill, err := json.Marshal(b)
			if err != nil {
				return nil, err
			}
			log.Info("test_wallet_getBillByAddr ", "result", string(bill))
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
		log.Error("test_wallet_getNewAddress", "err", err.Error())
		return nil, err
	}
	log.Info("test_wallet_getNewAddress ", "result", msg)
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
		log.Info("errr:", err.Error())
		return nil, err
	}
	log.Info("test_wallet_getAddressesByAccount ", "result", msg)
	return msg, nil
}
func test_wallet_getAccountByAddress(w *wallet.Wallet) (interface{}, error) {
	address := "TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
	cmd := &qitmeerjson.GetAccountCmd{
		Address: address,
	}
	msg, err := GetAccountByAddress(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	log.Info("test_wallet_getAccountByAddress ", "result", msg)
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
		log.Info("errr:", err.Error())
		return nil, err
	}
	log.Info("test_wallet_importPrivKey ", "result", msg)
	return msg, nil
}
func test_wallet_importrivKey(w *wallet.Wallet) (interface{}, error) {
	v := false
	cmd := &qitmeerjson.ImportPrivKeyCmd{
		PrivKey: "9QwXzXVQBFNm1fxP8jCqHJG9jZKjqrUKjYiTvaRxEbFobiNrvzhgZ",
		Rescan:  &v,
	}
	msg, err := ImportWifPrivKey(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
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
		log.Info("errr:", err.Error())
		return nil, err
	}
	log.Info("test_wallet_dumpPrivKey ", "result", msg)
	return msg, nil
}
func test_wallet_getAccountAndAddress(w *wallet.Wallet) (interface{}, error) {
	msg, err := GetAccountAndAddress(w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	a := msg.([]wallet.AccountAndAddressResult)
	log.Info("test_wallet_getAccountAndAddress ", "result", a)
	//log.Info("test_wallet_getAccountAndAddress :",a[1].AddrsOutput[0].Addr)
	return msg, nil
}
func test_wallet_sendToAddress(w *wallet.Wallet) (interface{}, error) {
	cmd := &qitmeerjson.SendToAddressCmd{
		Address: "TmgD1mu8zMMV9aWmJrXqQYnWRhR9SBfDZG6",
		//Address:"TmbCBKbZF8PeSdj5Chm22T4hZRMJY5D8XyX",
		//Address:"TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF",
		Amount: float64(1),
	}
	msg, err := SendToAddress(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
		return nil, err
	}
	log.Info("test_wallet_sendToAddress", "result", msg)
	return msg, nil
}
func test_wallet_updateblock(w *wallet.Wallet) error {
	cmd := &qitmeerjson.UpdateBlockToCmd{
		Toheight: 0,
	}
	err := UpdateBlock(cmd, w)
	if err != nil {
		log.Info("errr:", err.Error())
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
	log.Info("test_wif", "wif:", wif)
	wif1, err := util.DecodeWIF(wif.String(), w.ChainParams())
	if err != nil {
		return err
	}
	log.Info("test_wif", "wif1:", wif1)
	log.Info("test_wif", "wif1", wif1.PrivKey.SerializeSecret())
	return nil
}

func TestWallet_Method(t *testing.T) {
	//w, err := open_wallet()
	//if err != nil {
	//	log.Info("open_wallet err:", err)
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
	//test_wallet_getBalance(w)
	////
	////
	//msg,err:=test_wallet_sendToAddress(w)
	//if err!=nil{
	//	fmt.Printf(err.Error())
	//}else{
	//	fmt.Printf("%s\n",msg)
	//}
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
	//	log.Info("GetTx err:",err.Error())
	//	return
	//}
	//log.Info("GetTx :",str)
}
