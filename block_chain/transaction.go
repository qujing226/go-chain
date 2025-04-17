package chain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/qujing226/blockchain/wallet"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

// 交易输出

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))
	return txo

}

func (o *TXOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(string(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	o.PubKeyHash = pubKeyHash
}

func (o *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(o.PubKeyHash, pubKeyHash) == 0
}

type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

type Transaction struct {
	ID        []byte
	Vin       []TXInput
	Vout      []TXOutput
	TimeStamp int64
	Payload   []string
}

// NewUTXOTransaction 创建一个 unspent transaction output 交易
func NewUTXOTransaction(w *wallet.Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	pubKeyHash := wallet.HashPubKey(w.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutPuts(pubKeyHash, amount)
	if acc < amount {
		fmt.Println("ERROR: Not enough funds")
		os.Exit(1)
	}
	// Build a list of inouts
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, out := range outs {
			input := TXInput{txID, out, nil, nil}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	from := fmt.Sprintf("%s", w.GetAddress())
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		// a change
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}
	tx := &Transaction{nil, inputs, outputs, time.Now().UnixMilli(), []string{}}
	tx.ID = tx.Hash()

	UTXOSet.Blockchain.SignTransaction(tx, w.PrivateKey)
	return tx
}

// NewCoinBaseTX 创建一个 coinbase 交易
// coinbase 交易的输入和输出都是由系统自动生成的，不需要用户参与，因此 coinbase 交易没有输入和输出。
// coinbase 交易的 Vin 数组长度为 1，并且第一个输入的 Txid 为空字节，Vout 为 -1。
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}
	txin := TXInput{[]byte{}, -1, []byte{}, []byte{}}
	txout := NewTXOutput(20, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, time.Now().UnixMilli(), []string{data}}
	tx.ID = tx.Hash()
	return &tx
}

// String returns a human-readable representation of a transaction
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

	for i, input := range tx.Vin {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		// 根据 vin.Txid 得到引用的前交易
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		if prevTx.ID == nil {
			fmt.Println("ERROR: Previous transaction is missing")
			os.Exit(1)
		}
		// 设置当前输入：清空 Signature，PubKey 填充为引用输出的 PubKeyHash
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		// 计算待签名数据
		dataToSign := txCopy.Hash()

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, dataToSign)
		if err != nil {
			log.Panic(err)
		}

		// 固定 32 字节长度（P256 的字段均为32字节，不足补0）
		rBytes := fixedBytes(r.Bytes(), 32)
		sBytes := fixedBytes(s.Bytes(), 32)
		signature := append(rBytes, sBytes...)
		tx.Vin[inID].Signature = signature

		// 固定公钥的两部分长度
		xBytes := fixedBytes(privateKey.PublicKey.X.Bytes(), 32)
		yBytes := fixedBytes(privateKey.PublicKey.Y.Bytes(), 32)
		fullPubKey := append(xBytes, yBytes...)
		tx.Vin[inID].PubKey = fullPubKey

		// 清空临时拷贝中对应的 PubKey
		txCopy.Vin[inID].PubKey = nil
	}
}

// TrimmedCopy 返回一个没有签名的副本。
// 副本中输入的签名和公钥被设置为 nil。
// 签名的副本无法被其他节点所验证，因为签名的私钥是当前节点的私钥，而签名的公钥是当前节点的公钥。
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs, 0, tx.Payload}

	return txCopy
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	curve := elliptic.P256()
	txCopy := tx.TrimmedCopy()

	fmt.Printf("%+v\n", *tx)

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		if prevTx.ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}

		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		dataToVerify := txCopy.Hash()

		if len(vin.Signature) != 64 {
			fmt.Printf("ERROR: Signature length is invalid: got %d, expected %d\n", len(vin.Signature), 64)
			return false
		}

		// 固定分成 32 字节的两部分
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		if len(vin.PubKey) != 64 {
			fmt.Printf("ERROR: Public key length is invalid: got %d, expected %d\n", len(vin.PubKey), 64)
			return false
		}
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		if !ecdsa.Verify(&rawPubKey, dataToVerify, &r, &s) {
			fmt.Println("ERROR: Signature is not valid")
			return false
		}
	}
	return true
}

// Serialize returns a serialized Transaction
func (tx *Transaction) Serialize() []byte {
	buf, err := json.Marshal(*tx)
	if err != nil {
		log.Panic(err)
	}

	return buf
}

func (tx *Transaction) SerializeV1() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	err := json.Unmarshal(data, &transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

func DeserializeTransactionV1(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}

// Hash 返回当前交易的哈希值。消除了 ID 字段的影响。
func (tx *Transaction) Hash() []byte {
	txCopy := *tx
	txCopy.ID = []byte{}
	txCopy.Payload = []string{}
	txCopy.TimeStamp = 0

	hash := sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

// IsCoinbase 判断当前交易是否为 coinbase 交易
// coinbase 交易的输入和输出都是由系统自动生成的，不需要用户参与，因此 coinbase 交易没有输入和输出。
// coinbase 交易的 Vin 数组长度为 1，并且第一个输入的 Txid 为空字节，Vout 为 -1。
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// fixedBytes 将字节数组 b 固定为 size 长度，不足时在前面补 0
func fixedBytes(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}
