package walletrpc

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"github.com/HalalChain/qitmeer-lib/common/hash"
	"github.com/HalalChain/qitmeer-lib/common/marshal"
	"github.com/HalalChain/qitmeer-lib/core/message"
	"github.com/HalalChain/qitmeer-lib/crypto/ecc"
	"github.com/HalalChain/qitmeer-lib/params"

	"github.com/HalalChain/qitmeer-lib/core/address"
	"github.com/HalalChain/qitmeer-lib/core/types"
	"github.com/HalalChain/qitmeer-lib/engine/txscript"
	//"github.com/HalalChain/qitmeer-lib/params"
	"testing"
)

func Test_tx(t *testing.T)  {
	private_key:="7e445aa5ffd834cb2d3b2db50f8997dd21af29bec3d296aaa066d902b93f484b"
	from:="TmbsdsjwzuGboFQ9GcKg6EUmrr3tokzozyF"
	to:="TmT5dipuqvrWR2cSF4rFgRDsAAQXLh6qw3S"
	amount:=uint64(1250000000)
	//if(err!=nil){
	//	fmt.Println("err:",err.Error())
	//	return
	//}
	utxoHash:="ee28d445807a6e8512a1f78f7674469d3ad4ba084404fe42886044b80017cb5c"
	utxoScript:="76a91444d959afb6db4ad730a6e2c0daf46ceeb98c53a088ac"
	sign ,_:=hex.DecodeString(utxoScript)
	fmt.Println("sign:",sign)
	outputindex:=uint32(1)
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
	outpoint_amount:=uint64(2250000000)
	txinput:=types.NewTxInput(outpoint,outpoint_amount,nil)
	txoutput:=types.NewTxOutput(uint64(outpoint_amount)-amount,frompkscipt)
	txoutput1:=types.NewTxOutput(amount,topkscipt)

	tx := types.NewTransaction()
	tx.AddTxIn(txinput)
	tx.AddTxOut(txoutput)
	tx.AddTxOut(txoutput1)
	s:=types.TxSerializeFull
	b,err:=tx.Serialize(s)
	if err != nil {
		fmt.Println("err:",err.Error())
		return
	}
	fmt.Println("tran json:",hex.EncodeToString(b))
	signTx,err:=txSign(private_key,hex.EncodeToString(b),"testnet")
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

func txSign(privkeyStr string, rawTxStr string, network string) (string, error) {
	privkeyByte, err := hex.DecodeString(privkeyStr)
	if err != nil {
		return "", err
	}
	if len(privkeyByte) != 32 {
		return "", fmt.Errorf("invaid ec private key bytes: %d", len(privkeyByte))
	}
	privateKey, pubKey := ecc.Secp256k1.PrivKeyFromBytes(privkeyByte)
	h160 := hash.Hash160(pubKey.SerializeCompressed())
	fmt.Println("hex.EncodeToString(h160)：",hex.EncodeToString(h160))

	var param *params.Params
	switch network {
	case "mainnet":
		param = &params.MainNetParams
	case "testnet":
		param = &params.TestNetParams
	case "privnet":
		param = &params.PrivNetParams
	}
	addr, err := address.NewPubKeyHashAddress(h160, param, ecc.ECDSA_Secp256k1)
	if err != nil {
		return "", err
	}
	// Create a new script which pays to the provided address.
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}

	if len(rawTxStr)%2 != 0 {
		return "", fmt.Errorf("invaild raw transaction : %s", rawTxStr)
	}
	serializedTx, err := hex.DecodeString(rawTxStr)
	if err != nil {
		return "", err
	}

	var redeemTx types.Transaction
	err = redeemTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return "", err
	}
	var kdb txscript.KeyClosure = func(types.Address) (ecc.PrivateKey, bool, error) {
		return privateKey, true, nil // compressed is true
	}
	var sigScripts [][]byte
	for i := range redeemTx.TxIn {
		sigScript, err := txscript.SignTxOutput(param, &redeemTx, i, pkScript, txscript.SigHashAll, kdb, nil, nil, ecc.ECDSA_Secp256k1)
		if err != nil {
			return "", err
		}
		sigScripts = append(sigScripts, sigScript)
	}

	for i2 := range sigScripts {
		redeemTx.TxIn[i2].SignScript = sigScripts[i2]
	}

	mtxHex, err := marshal.MessageToHex(&message.MsgTx{Tx: &redeemTx})
	if err != nil {
		return "", err
	}
	return mtxHex, nil
}