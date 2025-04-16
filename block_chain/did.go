package chain

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/nuts-foundation/go-did/did"
	chain_did "github.com/qujing226/blockchain/did"
	"github.com/qujing226/blockchain/wallet"
	"log"
	"strings"
	"time"
)

// NewDidDocumentTransaction 创建一个did文档交易
func NewDidDocumentTransaction(w *wallet.Wallet, data []byte) *Transaction {
	// todo: utxo中的签名是写在input中的，对于did来说应该写在下一个payload中
	tx := &Transaction{nil, nil, nil, time.Now().UnixMilli(), []string{string(data)}}
	tx = signDidDocument(w, tx)
	tx.ID = tx.Hash()

	return tx
}

func FindDidDocument(bc *BlockChain, targetDID string) *did.Document {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, data := range tx.Payload {
					if strings.Contains(data, targetDID) {
						doc, err := chain_did.DeserializeDidDocument([]byte(data))
						if err != nil {
							fmt.Println("deserialize did document error")
							return nil
						}
						return doc
					}
				}
			}
		}

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return nil
}

func signDidDocument(w *wallet.Wallet, tx *Transaction) *Transaction {
	dataToSign := tx.Serialize()
	r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, dataToSign)
	if err != nil {
		log.Panic(err)
	}
	signature := append(r.Bytes(), s.Bytes()...)
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	tx.Payload = append(tx.Payload, encodedSignature)
	return tx
}
