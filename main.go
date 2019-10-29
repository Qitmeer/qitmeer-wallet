package main

import (
	"github.com/Qitmeer/qitmeer-wallet/console"
	"github.com/Qitmeer/qitmeer/log"
	"os"
)
var rootCmd =console.Command


func init()  {
	console.BindFlags()
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}