package main

import (
	"fmt"
	"github.com/Qitmeer/qitmeer-wallet/console"
	"os"
)
var rootCmd =console.Command


func init()  {
	console.BindFlags()
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}