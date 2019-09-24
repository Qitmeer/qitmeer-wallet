package walletrpc

import (
	"fmt"
	//"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/params"
	"github.com/Qitmeer/qitmeer-lib/core/address"
	//"github.com/Qitmeer/qitmeer-lib/common/hash"
	"encoding/hex"
	"testing"
)

func Test_text(t *testing.T)  {

	s,_:=hex.DecodeString("cdd006e3adcd991a15f46c7e975686c1aec20773")
	par:=params.TestNetParams
	//b:=hash.Hash160(s)
	//fmt.Printf("1:%x",b[:])
	//fmt.Println("hex:",hex.EncodeToString(b))
	//script,_:=hex.DecodeString("304402201edae2abc5761ea6025268202aff6a8e9d475f31fa69f7a52b0a652ef243e1c2022039b54520a2ceb04b4781863e846a1dcf143dfc7c4def57680fa03702801e7c5a01 0354455a60d86273d322eebb913d87f428988ce97922a366f0a0867a426df78bc9")
	////ha,err:=hash.NewHashFromStr(script)
	////if(err!=nil){
	////	fmt.Println("err:",err.Error())
	////	return
	////}
	//_, addresses, _, err :=txscript.ExtractPkScriptAddrs(txscript.DefaultScriptVersion,script,&par)
	//if(err!=nil){
	//	fmt.Println("err:",err.Error())
	//	return
	//}
	//fmt.Println("addresses:",addresses[0].String())

	a,err:=address.NewPubKeyHashAddressByNetId(s,par.PubKeyHashAddrID)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return
	}
	fmt.Println("addresses:",a.String())

}
