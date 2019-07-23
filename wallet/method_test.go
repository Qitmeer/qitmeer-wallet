package wallet

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-wallet/waddrmgs"
	"github.com/HalalChain/qitmeer-wallet/walletdb"
	"testing"
	//"time"
)

func TestWallet_NextAccount(t *testing.T) {
	dbpath:="C:\\Users\\luoshan\\AppData\\Local\\Qitwallet\\testnet"
	activeNet:=&params.TestNetParams
	load:=NewLoader(activeNet,dbpath,250)
	w,err:=load.OpenExistingWallet([]byte("public"),false)
	if(err!=nil){
		fmt.Println("openWallet err:",err.Error())
		return
	}
	err = walletdb.View(w.db, func(tx walletdb.ReadTx) error {
		addrmgrNs := tx.ReadBucket(waddrmgrNamespaceKey)
		return w.Manager.Unlock(addrmgrNs, []byte("123456"))
	})
	if(err!=nil){
		fmt.Println("Manager.Unlock err:",err.Error())
		return
	}
	_,err=w.NextAccount(waddrmgr.KeyScopeBIP0044,"luoshan2")
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return
	}
	fmt.Println("succc")
}
