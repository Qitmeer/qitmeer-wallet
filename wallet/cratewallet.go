package wallet

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Qitmeer/qitmeer-lib/crypto/ecc/secp256k1"

	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/internal/legacy/keystore"
	"github.com/Qitmeer/qitmeer-wallet/internal/prompt"
	"github.com/Qitmeer/qitmeer-wallet/utils"
)

// createWallet prompts the user for information needed to generate a new wallet
// and generates the wallet accordingly.  The new wallet will reside at the
// provided path.
func CreateWallet(cfg *config.Config, walletPass string) error {

	activNetParams := utils.GetNetParams(cfg.Network)

	dbDir := filepath.Join(cfg.AppDataDir, cfg.Network)
	loader := NewLoader(activNetParams, dbDir, 250, cfg)

	// When there is a legacy keystore, open it now to ensure any errors
	// don't end up exiting the process after the user has spent time
	// entering a bunch of information.
	netDir := filepath.Join(cfg.AppDataDir, cfg.Network)
	keystorePath := filepath.Join(netDir, keystore.Filename)
	var legacyKeyStore *keystore.Store
	_, err := os.Stat(keystorePath)
	if err != nil && !os.IsNotExist(err) {
		// A stat error not due to a non-existant file should be
		// returned to the caller.
		return err
	} else if err == nil {
		// Keystore file exists.
		legacyKeyStore, err = keystore.OpenDir(netDir)
		if err != nil {
			return err
		}
	}

	// Start by prompting for the private passphrase.  When there is an
	// existing keystore, the user will be promped for that passphrase,
	// otherwise they will be prompted for a new one.
	reader := bufio.NewReader(os.Stdin)
	privPass, err := prompt.PrivatePass(reader, legacyKeyStore)
	if err != nil {
		return err
	}

	// When there exists a legacy keystore, unlock it now and set up a
	// callback to import all keystore keys into the new walletdb
	// wallet
	if legacyKeyStore != nil {
		err = legacyKeyStore.Unlock(privPass)
		if err != nil {
			return err
		}

		// Import the addresses in the legacy keystore to the new wallet if
		// any exist, locking each wallet again when finished.
		loader.RunAfterLoad(func(w *Wallet) {
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
		[]byte(InsecurePubPassphrase), []byte(walletPass)) // []byte(cfg.WalletPass))
	if err != nil {
		return err
	}
	fmt.Println("pubPass:", string(pubPass))
	// Ascertain the wallet generation seed.  This will either be an
	// automatically generated value the user has already confirmed or a
	// value the user has entered which has already been validated.
	seed, err := prompt.Seed(reader)
	if err != nil {
		return err
	}

	fmt.Println("Creating the wallet...")
	w, err := loader.CreateNewWallet(pubPass, privPass, seed, time.Now())
	if err != nil {
		return err
	}

	w.Manager.Close()
	fmt.Println("The wallet has been created successfully.")
	return nil
}

// convertLegacyKeystore converts all of the addresses in the passed legacy
// key store to the new waddrmgr.Manager format.  Both the legacy keystore and
// the new manager must be unlocked.
func convertLegacyKeystore(legacyKeyStore *keystore.Store, w *Wallet) error {
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

			wif, err := utils.NewWIF((*secp256k1.PrivateKey)(privKey),
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
