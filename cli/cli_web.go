package cli

import "github.com/qujing226/blockchain/server"

func (cli *CLI) startWeb() {
	server.InitConfig()
	server.StartDidService()
}
