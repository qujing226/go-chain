package chain

import (
	"encoding/hex"
	"errors"
	"go.etcd.io/bbolt"
	"log"
)

const utxoBucket = "chainState"

type UTXOSet struct {
	Blockchain *BlockChain
}

// FindSpendableOutPuts finds and returns unspent outputs to reference in inputs
func (u *UTXOSet) FindSpendableOutPuts(pubkeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubkeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

func (u *UTXOSet) FindUTXO(pubkeyHash []byte) []TXOutput {
	UTXO := make([]TXOutput, 0)
	db := u.Blockchain.Db
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)
			//fmt.Printf("FindUTXO: txID=%x, outs=%+v\n", k, outs)
			for _, out := range outs.Outputs {
				//fmt.Printf("Compare: stored=%x, expecting=%x\n", out.PubKeyHash, pubkeyHash)
				if out.IsLockedWithKey(pubkeyHash) {
					UTXO = append(UTXO, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXO
}

// CountTransactions returns the number of transactions in the UTXO set
func (u *UTXOSet) CountTransactions() int {
	db := u.Blockchain.Db
	counter := 0

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return counter
}

// Reindex rebuilds the UTXO set
func (u *UTXOSet) Reindex() {
	db := u.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil {
			if !errors.Is(err, bbolt.ErrBucketNotFound) {
				return err
			}
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				log.Panic(err)
			}
			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the Block
// is considered to be the tip of a blockchain
func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Db

	err := db.Update(func(tx *bbolt.Tx) error {
		// 获取 utxoBucket
		b := tx.Bucket([]byte(utxoBucket))
		if b == nil {
			//return fmt.Errorf("bucket %s not found", utxoBucket)
		}

		// 遍历当前区块中的每一笔交易
		for _, tx := range block.Transactions {
			// 如果不是 coinbase 交易：
			if !tx.IsCoinbase() {
				// 遍历该交易中的每个 vin（输入）
				for _, vin := range tx.Vin {
					//fmt.Printf("Processing vin with Txid: %x, Vout: %d\n", vin.Txid, vin.Vout)

					updatedOuts := TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					if outsBytes == nil {
						//fmt.Printf("No UTXO found for Txid: %x\n", vin.Txid)
						continue
					}
					outs := DeserializeOutputs(outsBytes)
					//fmt.Printf("Original outputs for Txid %x: %+v\n", vin.Txid, outs)

					// 剔除已经被引用的输出
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						} else {
							//fmt.Printf("Spending output index %d from Txid: %x\n", outIdx, vin.Txid)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						// 如果 <txid> 下所有输出都已被花费，则删除该键
						err := b.Delete(vin.Txid)
						if err != nil {
							log.Panic(err)
						}
						//fmt.Printf("Deleted UTXO entry for Txid: %x\n", vin.Txid)
					} else {
						// 否则，更新该键对应的 value
						err := b.Put(vin.Txid, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
						//fmt.Printf("Updated UTXO entry for Txid: %x, new outputs: %+v\n", vin.Txid, updatedOuts)
					}
					//fmt.Println("-----")
				}
			}

			// 将当前交易的输出写入数据库：无论是否 coinbase
			newOutputs := TXOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//fmt.Printf("Added new UTXO for transaction %x: %+v\n", tx.ID, newOutputs)
			//fmt.Println("===========")
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}
