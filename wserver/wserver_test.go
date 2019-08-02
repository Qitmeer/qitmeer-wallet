package wserver

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/HalalChain/qitmeer-lib/crypto/bip39"
	"github.com/HalalChain/qitmeer-lib/crypto/seed"
)

func TestSeed(t *testing.T) {

	seedBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		t.Log(fmt.Errorf("GenerateSeed err: %s", err))
		return
	}

	t.Log(hex.EncodeToString(seedBuf))

	mnemonic, err := bip39.NewMnemonic(seedBuf)

	t.Log(mnemonic, err)

 

	s3, err := bip39.EntropyFromMnemonic(mnemonic)
	t.Log(hex.EncodeToString(s3), err)
}
