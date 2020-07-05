// Copyright (c) 2014-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Qitmeer/qitmeer/crypto/bip32"
	btcec "github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/internal/legacy/keystore"
	"github.com/Qitmeer/qitmeer-wallet/internal/prompt"
	"github.com/Qitmeer/qitmeer-wallet/utils"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"

	"github.com/Qitmeer/qitmeer/crypto/ecc/secp256k1"
	chaincfg "github.com/Qitmeer/qitmeer/params"
	"github.com/Qitmeer/qitmeer-wallet/wallet"
	"github.com/Qitmeer/qitmeer-wallet/walletdb"
)

// networkDir returns the directory name of a network directory to hold wallet
// files.
func networkDir(dataDir string, chainParams *chaincfg.Params) string {
	netname := chainParams.Name

	return filepath.Join(dataDir, netname)
}

// checkCreateDir checks that the path exists and is a directory.
// If path does not exist, it is created.
func checkCreateDir(path string) error {
	if fi, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Attempt data directory creation
			if err = os.MkdirAll(path, 0700); err != nil {
				return fmt.Errorf("cannot create directory: %s", err)
			}
		} else {
			return fmt.Errorf("error checking directory: %s", err)
		}
	} else {
		if !fi.IsDir() {
			return fmt.Errorf("path '%s' is not a directory", path)
		}
	}

	return nil
}

// createSimulationWallet is intended to be called from the rpcclient
// and used to create a wallet for actors involved in simulations.
func createSimulationWallet() error {
	// Simulation wallet password is 'password'.
	privPass := []byte("password")

	// Public passphrase is the default.
	pubPass := []byte(wallet.InsecurePubPassphrase)

	netDir := networkDir(config.Cfg.AppDataDir, config.ActiveNet)

	// Create the wallet.
	dbPath := filepath.Join(netDir, config.WalletDbName)
	fmt.Println("Creating the wallet...")

	// Create the wallet database backed by bolt db.
	db, err := walletdb.Create("bdb", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	// Create the wallet.
	err = wallet.Create(db, pubPass, privPass, nil, config.ActiveNet, time.Now())
	if err != nil {
		return err
	}

	fmt.Println("The wallet has been created successfully.")
	return nil
}

// createWallet prompts the user for information needed to generate a new wallet
// and generates the wallet accordingly.  The new wallet will reside at the
// provided path.
func createWallet() (*wallet.Wallet,error) {
	dbDir := networkDir(config.Cfg.AppDataDir, config.ActiveNet)
	loader := wallet.NewLoader(config.ActiveNet, dbDir, 250, &config.Config{})

	// When there is a legacy keystore, open it now to ensure any errors
	// don't end up exiting the process after the user has spent time
	// entering a bunch of information.
	netDir := networkDir(config.Cfg.AppDataDir, config.ActiveNet)
	keystorePath := filepath.Join(netDir, keystore.Filename)
	var legacyKeyStore *keystore.Store
	_, err := os.Stat(keystorePath)
	if err != nil && !os.IsNotExist(err) {
		// A stat error not due to a non-existant file should be
		// returned to the caller.
		return nil,err
	} else if err == nil {
		// Keystore file exists.
		legacyKeyStore, err = keystore.OpenDir(netDir)
		if err != nil {
			return nil,err
		}
	}

	// Start by prompting for the private passphrase.  When there is an
	// existing keystore, the user will be promped for that passphrase,
	// otherwise they will be prompted for a new one.
	reader := bufio.NewReader(os.Stdin)
	privPass, err := prompt.PrivatePass(reader, legacyKeyStore)
	if err != nil {
		return nil,err
	}

	// When there exists a legacy keystore, unlock it now and set up a
	// callback to import all keystore keys into the new walletdb
	// wallet
	if legacyKeyStore != nil {
		err = legacyKeyStore.Unlock(privPass)
		if err != nil {
			return nil,err
		}

		// Import the addresses in the legacy keystore to the new wallet if
		// any exist, locking each wallet again when finished.
		loader.RunAfterLoad(func(w *wallet.Wallet) {
			defer legacyKeyStore.Lock()

			fmt.Println("Importing addresses from existing wallet...")

			lockChan := make(chan time.Time, 1)
			defer func() {
				lockChan <- time.Time{}
			}()
			err := w.Unlock(privPass, lockChan)
			if err != nil {
				fmt.Printf("ERR: Failed to unlock new wallet "+
					"during old wallet key import: %v", err)
				return
			}

			err = convertLegacyKeystore(legacyKeyStore, w)
			if err != nil {
				fmt.Printf("ERR: Failed to import keys from old "+
					"wallet format: %v", err)
				return
			}
			// Remove the legacy key store.
			err = os.Remove(keystorePath)
			if err != nil {
				fmt.Printf("WARN: Failed to remove legacy wallet "+
					"from'%s'\n", keystorePath)
			}
		})
	}

	// Ascertain the public passphrase.  This will either be a value
	// specified by the user or the default hard-coded public passphrase if
	// the user does not want the additional public data encryption.
	pubPass, err := prompt.PublicPass(reader, privPass,
		[]byte(wallet.InsecurePubPassphrase), []byte(config.Cfg.WalletPass))
	if err != nil {
		return nil,err
	}
	fmt.Println("pubPass:", string(pubPass))
	// Ascertain the wallet generation seed.  This will either be an
	// automatically generated value the user has already confirmed or a
	// value the user has entered which has already been validated.
	seed, err := prompt.Seed(reader)
	if err != nil {
		return nil,err
	}
	seedKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		fmt.Println("failed to derive master extended key.")
		return nil,err
	}
	fmt.Println("Creating the wallet...")
	w, err := loader.CreateNewWallet(pubPass, privPass, seed, time.Now())
	if err != nil {
		return nil,err
	}
	fmt.Printf("pri:%x\n", seedKey.Key)
	pri, _ := secp256k1.PrivKeyFromBytes(seedKey.Key)
	wif, err := utils.NewWIF(pri, w.ChainParams(), true)
	if err != nil {
		fmt.Println("private key decode failed:", err.Error())
		return nil,err
	}
	if !wif.IsForNet(w.ChainParams()) {
		fmt.Println("Key is not intended for", w.ChainParams().Name, err.Error())
		return nil,err
	}
	w.UnLockManager(privPass)
	_, err = w.ImportPrivateKey(waddrmgr.KeyScopeBIP0044, wif)
	if err != nil {
		fmt.Println("ImportPrivateKey err:", err.Error())
		return nil,err
	}
	//w.Manager.Close()
	fmt.Println("The wallet has been created successfully.")
	return w,nil
}

// convertLegacyKeystore converts all of the addresses in the passed legacy
// key store to the new waddrmgr.Manager format.  Both the legacy keystore and
// the new manager must be unlocked.
func convertLegacyKeystore(legacyKeyStore *keystore.Store, w *wallet.Wallet) error {
	netParams := legacyKeyStore.Net()
	//blockStamp := waddrmgr.BlockStamp{
	//	Height: 0,
	//	Hash:   *netParams.GenesisHash,
	//}
	for _, walletAddr := range legacyKeyStore.ActiveAddresses() {
		switch addr := walletAddr.(type) {
		case keystore.PubKeyAddress:
			privKey, err := addr.PrivKey()
			if err != nil {
				fmt.Printf("WARN: Failed to obtain private key "+
					"for address %v: %v\n", addr.Address(),
					err)
				continue
			}

			wif, err := utils.NewWIF((*btcec.PrivateKey)(privKey),
				netParams, addr.Compressed())
			if err != nil {
				fmt.Printf("WARN: Failed to create wallet "+
					"import format for address %v: %v\n",
					addr.Address(), err)
				continue
			}
			fmt.Println("wif:", wif)
			//_, err = w.ImportPrivateKey(waddrmgr.KeyScopeBIP0044,
			//	wif, &blockStamp, false)
			//if err != nil {
			//	fmt.Printf("WARN: Failed to import private "+
			//		"key for address %v: %v\n",
			//		addr.Address(), err)
			//	continue
			//}

		//case keystore.ScriptAddress:
		//	_, err := w.ImportP2SHRedeemScript(addr.Script())
		//	if err != nil {
		//		fmt.Printf("WARN: Failed to import "+
		//			"pay-to-script-hash script for "+
		//			"address %v: %v\n", addr.Address(), err)
		//		continue
		//	}

		default:
			fmt.Printf("WARN: Skipping unrecognized legacy "+
				"keystore type: %T\n", addr)
			continue
		}
	}

	return nil
}
