package wallet

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Qitmeer/qitmeer-wallet/config"
	waddrmgr "github.com/Qitmeer/qitmeer-wallet/waddrmgs"
	"github.com/Qitmeer/qitmeer-wallet/walletdb"
	_ "github.com/Qitmeer/qitmeer-wallet/walletdb/bdb"
	"github.com/Qitmeer/qitmeer/log"
	chaincfg "github.com/Qitmeer/qitmeer/params"
)

const (
	walletDbName = "wallet.db"
)

var (
	// ErrLoaded describes the error condition of attempting to load or
	// create a wallet when the loader has already done so.
	ErrLoaded = errors.New("wallet already loaded")

	// ErrNotLoaded describes the error condition of attempting to close a
	// loaded wallet when a wallet has not been loaded.
	ErrNotLoaded = errors.New("wallet is not loaded")

	// ErrExists describes the error condition of attempting to create a new
	// wallet when one exists already.
	ErrExists = errors.New("wallet already exists")
)

type Loader struct {
	Cfg            *config.Config
	callbacks      []func(*Wallet)
	chainParams    *chaincfg.Params
	dbDirPath      string
	recoveryWindow uint32
	wallet         *Wallet
	db             walletdb.DB
	mu             sync.Mutex
}

func NewLoader(chainParams *chaincfg.Params, dbDirPath string,
	recoveryWindow uint32, cfg *config.Config) *Loader {

	return &Loader{
		Cfg:            cfg,
		chainParams:    chainParams,
		dbDirPath:      dbDirPath,
		recoveryWindow: recoveryWindow,
	}
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// onLoaded executes each added callback and prevents loader from loading any
// additional wallets.  Requires mutex to be locked.
func (l *Loader) onLoaded(w *Wallet, db walletdb.DB) {
	for _, fn := range l.callbacks {
		fn(w)
	}

	l.wallet = w
	l.db = db
	l.callbacks = nil // not needed anymore
}

// RunAfterLoad adds a function to be executed when the loader creates or opens
// a wallet.  Functions are executed in a single goroutine in the order they are
// added.
func (l *Loader) RunAfterLoad(fn func(*Wallet)) {
	l.mu.Lock()
	if l.wallet != nil {
		w := l.wallet
		l.mu.Unlock()
		fn(w)
	} else {
		l.callbacks = append(l.callbacks, fn)
		l.mu.Unlock()
	}
}

func (l *Loader) OpenWallet() (*Wallet, error) {
	b, err := l.WalletExists()
	if err != nil {
		return nil, err
	}
	if b {
		return l.OpenExistingWallet(nil, false)
	} else {
		return l.CreateNewWallet(nil, nil, nil, time.Now())
	}
}

// CreateNewWallet creates a new wallet using the provided public and private
// passphrases.  The seed is optional.  If non-nil, addresses are derived from
// this seed.  If nil, a secure random seed is generated.
func (l *Loader) CreateNewWallet(pubPassphrase, privPassphrase, seed []byte,
	bday time.Time) (*Wallet, error) {

	defer l.mu.Unlock()
	l.mu.Lock()

	if l.wallet != nil {
		return nil, ErrLoaded
	}

	dbPath := filepath.Join(l.dbDirPath, walletDbName)
	exists, err := fileExists(dbPath)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrExists
	}

	// Create the wallet database backed by bolt db.
	err = os.MkdirAll(l.dbDirPath, 0700)
	if err != nil {
		return nil, err
	}
	db, err := walletdb.Create("bdb", dbPath)
	if err != nil {
		return nil, err
	}

	// Initialize the newly created database for the wallet before opening.
	err = Create(
		db, pubPassphrase, privPassphrase, seed, l.chainParams, bday,
	)
	if err != nil {
		return nil, err
	}

	// Open the newly-created wallet.
	w, err := Open(db, pubPassphrase, nil, l.chainParams, l.recoveryWindow, l.Cfg)
	if err != nil {
		return nil, err
	}
	//w.Start()

	l.onLoaded(w, db)
	return w, nil
}

var errNoConsole = errors.New("db upgrade requires console access for additional input")

func noConsole() ([]byte, error) {
	return nil, errNoConsole
}

// OpenExistingWallet opens the wallet from the loader's wallet database path
// and the public passphrase.  If the loader is being called by a context where
// standard input prompts may be used during wallet upgrades, setting
// canConsolePrompt will enables these prompts.
func (l *Loader) OpenExistingWallet(pubPassphrase []byte, canConsolePrompt bool) (*Wallet, error) {
	defer l.mu.Unlock()
	l.mu.Lock()

	if l.wallet != nil {
		return nil, ErrLoaded
	}

	// Ensure that the network directory exists.
	if err := checkCreateDir(l.dbDirPath); err != nil {
		return nil, err
	}

	// Open the database using the boltdb backend.
	dbPath := filepath.Join(l.dbDirPath, walletDbName)
	log.Trace("OpenExistingWallet", "dbPath", dbPath)

	db, err := walletdb.Open("bdb", dbPath)
	if err != nil {
		log.Trace("Failed to open database", "err", err)
		return nil, err
	}
	log.Trace("OpenExistingWallet", "open db", "succ")

	var cbs *waddrmgr.OpenCallbacks
	if canConsolePrompt {
		//cbs = &waddrmgr.OpenCallbacks{
		//	ObtainSeed:        prompt.ProvideSeed,
		//	ObtainPrivatePass: prompt.ProvidePrivPassphrase,
		//}
	} else {
		cbs = &waddrmgr.OpenCallbacks{
			ObtainSeed:        noConsole,
			ObtainPrivatePass: noConsole,
		}
	}
	log.Trace("OpenExistingWallet", "open", "1")
	w, err := Open(db, pubPassphrase, cbs, l.chainParams, l.recoveryWindow, l.Cfg)
	if err != nil {
		log.Trace("OpenExistingWallet", "open error", err)
		// If opening the wallet fails (e.g. because of wrong
		// passphrase), we must close the backing database to
		// allow future calls to walletdb.Open().
		e := db.Close()
		if e != nil {
			log.Warn("Error closing database: %v", e)
		}
		return nil, err
	}
	log.Trace("OpenExistingWallet", "open ok", true)
	w.HttpClient, err = NewHtpc(config.Cfg)
	if err != nil {
		log.Error(fmt.Sprintf("wallet start, NewHtpc err: %s", err))
		return nil, err
	}

	l.onLoaded(w, db)
	return w, nil
}

// WalletExists returns whether a file exists at the loader's database path.
// This may return an error for unexpected I/O failures.
func (l *Loader) WalletExists() (bool, error) {
	dbPath := filepath.Join(l.dbDirPath, walletDbName)
	return fileExists(dbPath)
}

// LoadedWallet returns the loaded wallet, if any, and a bool for whether the
// wallet has been loaded or not.  If true, the wallet pointer should be safe to
// dereference.
func (l *Loader) LoadedWallet() (*Wallet, bool) {
	l.mu.Lock()
	w := l.wallet
	l.mu.Unlock()
	return w, w != nil
}

// UnloadWallet stops the loaded wallet, if any, and closes the wallet database.
// This returns ErrNotLoaded if the wallet has not been loaded with
// CreateNewWallet or LoadExistingWallet.  The Loader may be reused if this
// function returns without error.
func (l *Loader) UnloadWallet() error {
	defer l.mu.Unlock()
	l.mu.Lock()

	if l.wallet == nil {
		return ErrNotLoaded
	}

	l.wallet.Stop()
	l.wallet.WaitForShutdown()
	err := l.db.Close()
	if err != nil {
		return err
	}

	l.wallet = nil
	l.db = nil
	return nil
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
