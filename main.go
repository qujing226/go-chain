package main

import (
	"github.com/qujing226/blockchain/chain"
)

func main() {
	c := chain.NewBlockChain()
	c.AddBlock("Send 1 BTC to Ivan")
	c.AddBlock("Send 2 more BTC to Ivan")

	//for _, block := range c.db {
	//	pow := chain.NewProofOfWork(block)
	//	fmt.Printf("Pow : %s\n", strconv.FormatBool(pow.Validate()))
	//}
}
