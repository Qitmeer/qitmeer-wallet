package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/HalalChain/qitmeer-wallet/config"
	"github.com/HalalChain/qitmeer-wallet/version"
)

func main() {
	cmd := newCmd()
	err := cmd.Execute()
	if err != nil {
		log.Errorf("main err: %s", err)
	}
}

func newCmd() (cmd *cobra.Command) {
	var cfg *config.Config

	// root Command
	cmd = &cobra.Command{
		Use:   "qitmeer-wallet",
		Short: `qitmeer-wallet`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			cfg, err = config.LoadConfig(cfg.ConfigFile, false)
			if err != nil {
				return fmt.Errorf("cmd PersistentPreRunE err: %s", err)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("flags ConfigFile: ", cfg)
			//wallet.Main(cfg)
			fmt.Println("start wallet")
		},
	}

	// Create Wallet Command
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "create new wallet or recover wallet from seed",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("wallet create")
		},
	}

	// version Command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version())
		},
	}

	cmd.AddCommand(versionCmd)
	cmd.AddCommand(createCmd)

	//cmd flags
	gFlags := cmd.PersistentFlags()
	cfg = config.NewDefaultConfig()
	gFlags.StringVarP(&cfg.ConfigFile, "config", "C", cfg.ConfigFile, "Path to configuration file")
	gFlags.StringVarP(&cfg.AppDataDir, "appdata", "A", cfg.AppDataDir, "Application data directory for wallet config, databases and logs")
	gFlags.StringVarP(&cfg.DebugLevel, "debuglevel", "d", cfg.DebugLevel, "Logging level {trace, debug, info, warn, error, critical}")

	return
}
