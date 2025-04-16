package cli

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	chain "github.com/qujing226/blockchain/block_chain"
	"github.com/qujing226/blockchain/server"
	"github.com/qujing226/blockchain/wallet"
	"log"
)

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := chain.NewBlockChain(nodeID)
	UTXOSet := chain.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := chain.NewBlockChain(nodeID)
	UTXOSet := chain.UTXOSet{Blockchain: bc}
	defer bc.Close()

	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := chain.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := chain.NewCoinBaseTX(from, "")
		txs := []*chain.Transaction{cbTx, tx}
		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		server.SendTx(tx)
	}

	fmt.Println("Success!")
}

func (cli *CLI) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := chain.NewBlockChain(nodeID)
	UTXOSet := chain.UTXOSet{Blockchain: bc}
	defer bc.Close()

	balance := 0
	pubKeyHash := base58.Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	//fmt.Println(UTXOs, pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
