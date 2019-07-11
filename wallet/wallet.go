package wallet

import (
	"github.com/HalalChain/qitmeer-wallet/server"
	"github.com/HalalChain/qitmeer-wallet/tray"
)

// Start Wallet
func Start() {
	server.Start()

	tray.Open("http://127.0.0.1:1236")
}
