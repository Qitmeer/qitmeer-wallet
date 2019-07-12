package services

import (
	"encoding/hex"

	"github.com/HalalChain/qitmeer-lib/crypto/seed"
)

// AccountAPI key,address manage api
type AccountAPI struct {
}

//MakeEntropy generate a cryptographically secure pseudorandom entropy
func (c *AccountAPI) MakeEntropy() (string, error) {
	seedBuf, err := seed.GenerateSeed(uint16(32))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(seedBuf), err
}

//MakeHdKey create a new HD(BIP32) private key from an entropy
func (c *AccountAPI) MakeHdKey() (st string) {
	return "ok"
}

//ListAccount list all account
func (c *AccountAPI) ListAccount() ([]string, error) {

	//
	//{"alias":"defalut","keys":""}

	return []string{"a", "b"}, nil
}

//NewAccount make a new account
func (c *AccountAPI) NewAccount(alias string, pass string) ([]string, error) {
	return []string{}, nil
}

//ListAddresses list account all address
func (c *AccountAPI) ListAddresses(alias string) ([]string, error) {
	return []string{}, nil
}

//NewAddress make new address account
func (c *AccountAPI) NewAddress(alias string) (string, error) {

	return "", nil
}
