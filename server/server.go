package server

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/fatih/color"
	"github.com/qujing226/blockchain/block_chain"
	"io"
	"log"
	"net"
	"time"
)

const (
	protocol      = "tcp"
	nodeVersion   = 1
	commandLength = 12
)

var (
	nodeAddress string
	// miningAddress 只会在矿工节点上设置。如果有两笔或者更多的交易则开始挖矿。
	miningAddress string
	// knowNodes 是区块链的节点池
	// “localhost:3000”这是对中心节点的地址进行硬编码：因为每个节点必须知道从何处开始初始化。
	knownNodes = []string{"localhost:3000"}
	// blocksInTransit 跟踪已下载的块。这能够让我们从不同的节点下载块。
	// 在将块置于传送状态时，我们给 inv 消息的发送者发送 getData 命令并更新 blocksInTransit。
	blocksInTransit [][]byte
	// memPool 存储所有交易，直到被挖出块。
	memPool = make(map[string]chain.Transaction)
)

type addr struct {
	AddrList []string
}

// block 用来发送 Block 消息。
type block struct {
	AddrFrom string
	Block    []byte
}

// tx 用来发送交易消息。
type tx struct {
	AddFrom     string
	Transaction []byte
}

// getBlocks 用来请求 BlockHashes。
type getBlocks struct {
	AddrFrom string
}

// getData 用于某个块或交易的请求，它可以仅包含一个块或交易的 ID。
type getData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

// inv 用来向其他节点展示当前节点有什么块和交易。它没有包含完整的区块链和交易，仅仅是哈希而已。
type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

// 节点通过消息（message）进行交流。 当一个新的节点开始运行时，
// 它会从一个 DNS 种子获取几个节点，发送 version 消息
type version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

// sendVersion 用于发送本节点的版本信息。
// 如果当前节点不是中心节点，则必须向中心节点发送version信息
// 通过 Height 来进行确认本节点是否是最新的节点。
func sendVersion(addr string, bc *chain.BlockChain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(version{
		Version:    nodeVersion,
		BestHeight: bestHeight,
		AddrFrom:   nodeAddress,
	})
	request := append(commandToBytes("version"), payload...)

	sendData(addr, request)
}

// sendInv 用于发送 inv 消息。
// inv 来向其他节点展示当前节点有什么块和交易。它没有包含完整的区块链和交易，仅仅是哈希而已。
func sendInv(addr, command string, items [][]byte) {
	inventory := inv{nodeAddress, command, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(addr, request)
}

// sendGetBlocks 用于发送 getBlocks 消息。期望对方返回所有 BlockHashes。
func sendGetBlocks(addr string) {
	payload := gobEncode(getBlocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(addr, request)
}

// sendGetData 用于发送 getData 消息。
func sendGetData(addr, kind string, id []byte) {
	payload := gobEncode(getData{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(addr, request)
}

// sendBlock 用于发送一个 block 消息。
// block 消息指定了节点地址，附带一个块的二进制序列。
func sendBlock(addr string, b *chain.Block) {
	data := block{nodeAddress, b.Serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

// sendTx 用于发送一个 tx 消息。
func sendTx(addr string, t *chain.Transaction) {
	data := tx{nodeAddress, t.Serialize()}
	payload := gobEncode(data)
	// 尝试通过json进行编码
	//payload, err := json.Marshal(data)
	//if err != nil {
	//	fmt.Println("json marshal error")
	//}

	request := append(commandToBytes("tx"), payload...)
	sendData(addr, request)
}

func SendTx(tx *chain.Transaction) {
	sendTx(knownNodes[0], tx)
}
func sendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updateNodes []string

		for _, node := range knownNodes {
			if node != addr {
				updateNodes = append(updateNodes, node)
			}
		}
		knownNodes = updateNodes

		return
	}
	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

// handlerVersion 通过比较本节点的版本号和远程节点的版本号（height）进行处理
// 如果本节点的版本号小于远程节点的版本号，则发送 getBlocks 消息。否则发送version
func handlerVersion(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload version
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight
	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	// sendAddr(payload.AddrFrom)
	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleAddr(request []byte) {
	var buff bytes.Buffer
	var payload addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	knownNodes = append(knownNodes, payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

// handleGetBlocks 用于处理 getBlocks 消息。
// 它通过 inv 消息向 <对方> 返回 <当前节点> 所有的 BlockHashes。
func handleGetBlocks(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload getBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

// handleInv 用于处理 inv 消息。
// inv 消息来源于对方，它包含对方的所有 BlockHashes。
func handleInv(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		var newInTransit [][]byte
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	} else if payload.Type == "tx" {
		txID := payload.Items[0]
		txIDHex := hex.EncodeToString(txID)
		if memPool[txIDHex].ID == nil {
			fmt.Printf("Transaction %s not found in memPool, sending getData request\n", txIDHex)
			sendGetData(payload.AddrFrom, "tx", txID)
		} else {
			fmt.Printf("Transaction %s has found in memPool\n", txIDHex)
		}
	}
}

// handleGetData 用于处理 getData 消息。
func handleGetData(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload getData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memPool[txID]
		sendTx(payload.AddrFrom, &tx)

	}
}

// handleBlock 用于处理 block 消息。
func handleBlock(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	blockData := payload.Block
	b := chain.DeSerializeBlock(blockData)

	fmt.Println("a new block received!")
	bc.AddBlock(b)

	fmt.Printf("Added block %x \n", b.Hash)
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)
		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := chain.UTXOSet{Blockchain: bc}
		UTXOSet.Reindex()
	}
}

func handleTx(request []byte, bc *chain.BlockChain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	//err := json.Unmarshal(request[commandLength:], &payload)
	//if err != nil {
	//	log.Panic(err)
	//}

	txData := payload.Transaction
	tx := chain.DeserializeTransaction(txData)
	memPool[hex.EncodeToString(tx.ID)] = tx

	// 如果是中心节点，就将挖矿信息推广到除自身和挖矿节点之外的节点。
	// 中心节点是不会挖矿的。
	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		// miningAddress 只会在矿工节点上设置。如果有两笔或者更多的交易则开始挖矿。
		if len(memPool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*chain.Transaction

			for id := range memPool {
				tx := memPool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}
			// 验证后的交易被放到一个块里，同时还有附带奖励的 coinbase 交易。
			// 当块被挖出来以后，UTXO 集会被重新索引。
			cbTx := chain.NewCoinBaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := chain.UTXOSet{Blockchain: bc}
			UTXOSet.Reindex()

			fmt.Println("New block mined!")

			// 删除已经挖出的块里的交易
			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(memPool, txID)
			}

			// 当前节点所连接到的所有其他节点，接收带有新块哈希的 inv 消息。
			// 在处理完消息后，它们可以对块进行请求。
			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}
			if len(memPool) > 0 {
				goto MineTransactions
			}
		}
	}
}
func handleConnection(conn net.Conn, bc *chain.BlockChain) {
	fmt.Printf("--> Received message from %s | Time: %v\n", conn.RemoteAddr(), time.Now().Format(" 15:04:05"))

	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("--> Message : %s command, Dealing...\n", command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handlerVersion(request, bc)
	default:
		fmt.Println("Unknown command received!")
	}
	_ = conn.Close()
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	bc := chain.NewBlockChain(nodeID)
	if nodeAddress != knownNodes[0] {

		sendVersion(knownNodes[0], bc)
	}

	printInformation(nodeID, minerAddress)

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}
}

// ... 其他导入和代码 ...

func printInformation(nodeID string, minerAddress string) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	magenta := color.New(color.FgMagenta).SprintFunc()

	fmt.Printf("%s %s %s\n", green("==="), green("Date:"), green(time.Now().Format("2006-01-02 15:04:05")))
	fmt.Printf("%s %s %s %s %s\n", green("==="), yellow("Node:"), yellow(nodeID), yellow("is Handling Connection... | address:"), yellow(nodeAddress))
	if minerAddress != "" {
		fmt.Printf("%s %s\n", green("==="), magenta("INFO: This is a miner Node!"))
	}
	if nodeAddress == knownNodes[0] {
		fmt.Printf("%s %s\n", green("==="), cyan("INFO: This is the Genesis Node!"))
	}
	fmt.Println(green("==="))
}

// gobEncode 用于将数据(any)编码成二进制流。
func gobEncode(data any) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

// commandToBytes 用于将命令(string)编码成二进制流。
func commandToBytes(command string) []byte {
	var bytes [commandLength]byte
	for i, c := range command {
		bytes[i] = byte(c)
	}
	return bytes[:]
}

// bytesToCommand 用于将二进制流解码成命令(string)。
// 命令的长度不同，因此需要有 != 0x0 的判断。
func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

// nodeIsKnown 用于判断节点是否已经存在于 knownNodes 中。
func nodeIsKnown(addr string) bool {
	for _, val := range knownNodes {
		if val == addr {
			return true
		}
	}
	return false
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}
