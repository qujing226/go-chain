package chain

import (
	"fmt"
	"go.etcd.io/bbolt"
)

type BlockChain struct {
	tip []byte
	db  *bbolt.DB
}

const dbFile = "blockchain.db"
const blocksBucket = "block"

func NewBlockChain() *BlockChain {
	var tip []byte

	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		panic(err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		if bucket == nil {
			genesisBlock := NewBlock("Genesis Block", []byte{})
			bucket, err = tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}
			err = bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				return err
			}
			err = bucket.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				return err
			}
			tip = genesisBlock.Hash
		} else {
			tip = bucket.Get([]byte("l"))
		}
		return err
	})

	return &BlockChain{
		tip: tip,
		db:  db,
	}
}

func (bc *BlockChain) AddBlock(data string) {
	var latestHash []byte
	err := bc.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		latestHash = bucket.Get([]byte("l"))

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	newBlock := NewBlock(data, latestHash)
	err = bc.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		err = bucket.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = bucket.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})
}
