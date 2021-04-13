package wserver

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/Qitmeer/qitmeer-wallet/utils"
	"github.com/Qitmeer/qitmeer/crypto/bip32"
	"github.com/Qitmeer/qitmeer/crypto/bip39"
	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer/crypto/seed"

	"github.com/Qitmeer/qitmeer/qx"
)

func TestSeed(t *testing.T) {

	activeNetParams := utils.GetNetParams("testnet")

	seedBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		t.Log(fmt.Errorf("GenerateSeed err: %s", err))
		return
	}
	seed := hex.EncodeToString(seedBuf)
	if seed == ""{
		t.Error("failed to encode seed to string")
	}

	mnemonic, err := bip39.NewMnemonic(seedBuf)

	s3, err := bip39.EntropyFromMnemonic(mnemonic)

	if hex.EncodeToString(s3) != seed{
		t.Errorf("generate entropy by mnemonic, Got:%s, Expected:%s", s3, seed)
	}

	//import master key addr
	seedKey, err := bip32.NewMasterKey(seedBuf)
	if err != nil {
		t.Logf("createWallet NewMasterKey err: %s", err)
		return
	}

	pri, _ := secp256k1.PrivKeyFromBytes(seedKey.Key)
	wif, err := utils.NewWIF(pri, activeNetParams, true)
	if err != nil {
		t.Logf("createWallet private key decode failed: %s", err)
		return
	}
	if !wif.IsForNet(activeNetParams) {
		t.Logf("createWallet Key is not intended for: %s %s", activeNetParams.Name, err)
		return
	}
}

func TestGan(t *testing.T) {
	//func GenerateAddr() string {

	ver := "mainnet"


	s, err := qx.NewEntropy(32)
	if err != nil {
		t.Logf("An error occurred generating s，%s", err)
	}
	prv, err := qx.EcNew("secp256k1", s)
	if err != nil {
		t.Logf("An error occurred generating private key，%s", err)
	}
	pub, err := qx.EcPrivateKeyToEcPublicKey(false, prv)
	if err != nil {
		t.Logf("An error occurred generating public key，%s", err)
	}
	addr, err := qx.EcPubKeyToAddress(ver, pub)
	if err != nil {
		t.Logf("An error occurred generating address，%s", err)
	}

	t.Log(addr)

}
