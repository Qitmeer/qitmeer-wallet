package main

import (
	"fmt"

	"github.com/HalalChain/qitmeer-wallet/wallet"
)

func main() {

	wallet.Start()

	ch := make(chan int)

	<-ch

	fmt.Println("wallet")
}
