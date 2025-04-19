package cli

import (
	"fmt"
	"github.com/qujing226/blockchain/wallet"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := wallet.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) createKemWallet() {
	kemWallets, _ := wallet.NewKemWallets()
	address := kemWallets.CreateWallet()
	kemWallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}
