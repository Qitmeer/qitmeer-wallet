package main

import (
	"github.com/HalalChain/qitmeer-wallet/wallet"
)

func main() {

	wallet.Start()

	ch := make(chan int)
	<-ch
}
