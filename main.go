package main

import (
	"github.com/HalalChain/qitmeer-wallet/console"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-console" {
		console.StartConsole()
	}else{
		log.Println("test")
	}
}
