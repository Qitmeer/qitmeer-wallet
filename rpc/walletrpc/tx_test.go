package walletrpc

import (
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/Qitmeer/qitmeer-lib/common/hash"

	"github.com/Qitmeer/qitmeer-lib/core/address"
	"github.com/Qitmeer/qitmeer-lib/core/types"
	"github.com/Qitmeer/qitmeer-lib/engine/txscript"
	"github.com/Qitmeer/qitmeer-lib/qx"
	//"github.com/Qitmeer/qitmeer-lib/params"
	"testing"
)

func Test_tx(t *testing.T)  {
	private_key:="ca462c7e9582955cc8c2a6cbb03282861fa50d7f2f7be7035414f9b765b2920b"
	from:="TmgD1mu8zMMV9aWmJrXqQYnWRhR9SBfDZG6"
	to:="TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
	amount:=uint64(1000000)
	//if(err!=nil){
	//	fmt.Println("err:",err.Error())
	//	return
	//}
	utxoHash:="81a846bf4a78d30e0040e10b8c8415f6f7b5213b9d0683e286f2799576df0b59"
	utxoScript:="76a914bd4d1888cb054b2755d65d93c356573e4d283ead88ac"
	sign ,_:=hex.DecodeString(utxoScript)
	fmt.Println("sign:",sign)
	outputindex:=uint32(0)
	fromAdds, err := address.DecodeAddress(from)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return
	}
	toAdds, err := address.DecodeAddress(to)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return
	}
	//chainParams := &params.TestNetParams
	//locktime := int64(1558702865)
	frompkscipt,err:=txscript.PayToAddrScript(fromAdds)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return
	}
	topkscipt,err:=txscript.PayToAddrScript(toAdds)
	if(err!=nil){
		fmt.Println("err:",err.Error())
		return
	}
	outpointhash,_:=hash.NewHashFromStr(utxoHash)
	outpoint:=types.NewOutPoint(outpointhash,outputindex)
	outpoint_amount:=uint64(100000000)
	txinput:=types.NewTxInput(outpoint,nil)
	txoutput:=types.NewTxOutput(uint64(outpoint_amount)-amount,frompkscipt)
	txoutput1:=types.NewTxOutput(amount,topkscipt)

	tx := types.NewTransaction()
	tx.AddTxIn(txinput)
	tx.AddTxOut(txoutput)
	tx.AddTxOut(txoutput1)
	b,err:=tx.Serialize()
	if err != nil {
		fmt.Println("err:",err.Error())
		return
	}
	fmt.Println("tran json:",hex.EncodeToString(b))
	signTx,err:=qx.TxSign(private_key,hex.EncodeToString(b),"testnet")
	//signTx,err:=txSign(private_key,hex.EncodeToString(b),"testnet")
	fmt.Println("signTx:",signTx)
}
func blake256(input string)[]byte{
	data, err :=hex.DecodeString(input)
	if err != nil {
		fmt.Printf("%x\n",err)
		return nil
	}
	hasher := crypto.BLAKE2b_256.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	// fmt.Printf("%x\n",hash[:])
	return hash[:]
}