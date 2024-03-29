package main

import (
	"github.com/Qitmeer/qitmeer-wallet/commands"
	"github.com/Qitmeer/qng/log"
	"os"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
