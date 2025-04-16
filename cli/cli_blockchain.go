package cli

import (
	"fmt"
	"github.com/qujing226/blockchain/block_chain"
	"github.com/qujing226/blockchain/wallet"
	"log"
	"strconv"
)

func (cli *CLI) createBlockchain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := chain.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()

	UTXOSet := chain.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) printChain(nodeID string) {
	bc := chain.NewBlockChain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PreBlockHash)
		pow := chain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
}
