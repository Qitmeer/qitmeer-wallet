package main

import (
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/HalalChain/qitmeer-lib/log"
	"github.com/HalalChain/qitmeer-wallet/version"
)

var (
	cfg *config
)

func main() {
	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Work around defer not working after os.Exit.
	if err := walletMain(); err != nil {
		os.Exit(1)
	}
}

// walletMain is a work-around main function that is required since deferred
// functions (such as log flushing) are not called with calls to os.Exit.
// Instead, main runs this function and checks for a non-nil error, at which
// point any defers have already run, and if the error is non-nil, the program
// can be exited with an error exit status.
func walletMain() error {
	// Load configuration and parse command line.  This function also
	// initializes logging and configures it accordingly.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	// Show version at startup.
	log.Info("Version %s", version.Version())

	if cfg.Profile != "" {
		go func() {
			listenAddr := net.JoinHostPort("", cfg.Profile)
			log.Info("Profile server listening on %s", listenAddr)
			profileRedirect := http.RedirectHandler("/debug/pprof",
				http.StatusSeeOther)
			http.Handle("/", profileRedirect)
			log.Error("%v", http.ListenAndServe(listenAddr, nil))
		}()
	}

	//dbDir := networkDir(cfg.AppDataDir.Value, activeNet)
	//loader := wallet.NewLoader(activeNet, dbDir, 250)
	log.Info("Shutdown complete")
	return nil
}
