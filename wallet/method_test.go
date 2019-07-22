package wallet

import (
	"fmt"
	"github.com/HalalChain/qitmeer-lib/params"
	"github.com/HalalChain/qitmeer-wallet/waddrmgs"
	"testing"
	"time"
)

func TestWallet_NextAccount(t *testing.T) {
	dbpath:="C:\\Users\\luoshan\\AppData\\Local\\qitmeer\\data\\wallet"
	activeNet:=&params.TestNetParams
	load:=NewLoader(activeNet,dbpath,250)
	w,err:=load.OpenWallet(time.Now())
	if(err!=nil){
		fmt.Println("openWallet err:",err.Error())
		return
	}
	_,err=w.NextAccount(waddrmgr.KeyScopeBIP0044,"luoshan")
	if(err!=nil){
		fmt.Println("errr:",err.Error())
		return
	}
	fmt.Println("succc")
}
