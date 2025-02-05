package chain

import (
	"bytes"
	"encoding/gob"
	"time"
)

type Block struct {
	TimeStamp    int64
	Data         []byte
	PreBlockHash []byte
	Hash         []byte
	Nonce        int
}

func NewBlock(data string, preBlockHash []byte) *Block {
	block := &Block{
		TimeStamp:    time.Now().Unix(),
		Data:         []byte(data),
		PreBlockHash: preBlockHash,
		Hash:         []byte{},
		Nonce:        0,
	}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
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

func (b *Block) DeSerialize(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}
	return &block
}

//func (b *Block) SetHash() {
//	timeStamp := []byte(strconv.FormatInt(b.TimeStamp, 10))
//	headers := bytes.Join([][]byte{b.PreBlockHash, b.Data, timeStamp}, []byte{})
//	hash := sha256.Sum256(headers)
//
//	b.Hash = hash[:]
//}
