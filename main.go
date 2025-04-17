package main

import (
	"github.com/qujing226/blockchain/cli"
)

func main() {
	//ch := make(chan struct{}, 1)
	//
	//go func() {
	//	if os.Getenv("NODE_ID") == "3003" {
	//		server.InitConfig()
	//		server.StartDidService()
	//	}
	//	ch <- struct{}{}
	//}()

	c := cli.CLI{}
	c.Run()
}
