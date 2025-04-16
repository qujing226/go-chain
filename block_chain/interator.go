package chain

import "go.etcd.io/bbolt"

// BlockChainIterator 迭代器
type BlockChainIterator struct {
	currentHash []byte
	db          *bbolt.DB
}

func (i *BlockChainIterator) Next() *Block {
	var block *Block
	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeSerializeBlock(encodedBlock)
		return nil
	})
	if err != nil {
		return nil
	}
	i.currentHash = block.PreBlockHash
	return block
}
