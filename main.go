package main

import (
	"github.com/qujing226/blockchain/cli"
	"github.com/qujing226/blockchain/server"
	"os"
)

func main() {
	ch := make(chan struct{}, 1)

	go func() {
		if os.Getenv("NODE_ID") == "3003" {
			server.InitConfig()
			server.StartDidService()
		}
		ch <- struct{}{}
	}()

	c := cli.CLI{}
	c.Run()
	<-ch
}
