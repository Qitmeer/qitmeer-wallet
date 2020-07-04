package commands

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/config"
	"github.com/Qitmeer/qitmeer-wallet/wserver"
	"github.com/Qitmeer/qitmeer/log"
	"github.com/spf13/cobra"
)

// web mode

var WebCmd = &cobra.Command{
	Use:   "web",
	Short: "web administration UI",
	Example: `
		Enter web mode
		`,
	Args: cobra.NoArgs,
	PersistentPreRun: LoadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("web model")
		qitmeerMain(fileCfg)
	},
}

func AddWebCommand() {

}

func qitmeerMain(cfg *config.Config) {
	log.Trace("Qitmeer Main")
	wsvr, err := wserver.NewWalletServer(cfg)
	if err != nil {
		log.Error(fmt.Sprintf("NewWalletServer err: %s", err))
		return
	}
	wsvr.Start()

	exitCh := make(chan int)
	<-exitCh
}
