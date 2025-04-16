package chain

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	TimeStamp    int64
	PreBlockHash []byte
	Hash         []byte
	Nonce        int
	Height       int
	Transactions []*Transaction // 存储所有交易（其中可能包含 DID 文档相关交易）
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	fmt.Println("Generating genesis block...")
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func NewBlock(transactions []*Transaction, preBlockHash []byte, height int) *Block {
	block := &Block{
		TimeStamp:    time.Now().Unix(),
		Transactions: transactions,
		PreBlockHash: preBlockHash,
		Hash:         []byte{},
		Nonce:        0,
		Height:       height,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// HashTransactions 计算块中所有交易的哈希值
func (b *Block) HashTransactions() []byte {
	// 以总交易哈希值作为块哈希值
	//var txHashes [][]byte
	//var txHash [32]byte
	//for _, tx := range b.Transactions {
	//	txHashes = append(txHashes, tx.ID)
	//}
	//txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	//return txHash[:]

	// 优化：使用 merkle tree, 返回根节点的 hash
	var transactions [][]byte
	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}
	return result.Bytes()
}

func DeSerializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}
	return &block
}
