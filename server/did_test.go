package server

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/qujing226/blockchain/wallet"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGenerateDIDChallengeSignature(t *testing.T) {
	tc := []struct {
		name      string
		address   string
		challenge string // base64 编码后的 challenge
	}{
		{
			name:      "success",
			address:   "1J7v5bua2tLRovJmoQ32C7xrhMqvX1z8Qw",
			challenge: "lrFseGmdlWkjS/zENsz5xz6KVqBUP45E+d/av32Vlow=",
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			// 创建或加载钱包集合（假设节点ID已经定义）
			ws, err := newWallets("3003") // 请替换为实际节点ID
			require.NoError(t, err)

			// 根据地址获得钱包实例
			w := ws.GetWallet(tt.address)
			require.NotNil(t, w)

			// 将 base64 的 challenge 解码成字节
			challengeBytes, err := base64.StdEncoding.DecodeString(tt.challenge)
			require.NoError(t, err)

			// 使用私钥签名 challenge。注意，此处 challengeBytes 应该是消息的哈希值，
			// 如果 challenge 尚未进行哈希处理，请先使用例如 sha256 对其哈希后再签名。
			r, s, err := ecdsa.Sign(rand.Reader, &w.PrivateKey, challengeBytes)
			require.NoError(t, err)

			// 固定 r 和 s 的字节长度为 32 字节（对于 P256 曲线）
			rBytes := fixedBytes(r.Bytes(), 32)
			sBytes := fixedBytes(s.Bytes(), 32)

			// 将两个部分拼接成最终签名（64 字节）
			signature := append(rBytes, sBytes...)

			// 若需要调试或传输，也可以将签名进行 base64 编码
			signatureEncoded := base64.StdEncoding.EncodeToString(signature)
			fmt.Printf("Generated DID challenge signature: %s\n", signatureEncoded)

		})
	}
}

func newWallets(nodeID string) (*wallet.Wallets, error) {
	walletFile := fmt.Sprintf("../components/wallet_%s.dat", nodeID)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return nil, err
	}

	fileContent, _ := os.ReadFile(walletFile)
	var wallets wallet.Wallets

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err := decoder.Decode(&wallets)
	if err != nil {
		return nil, err
	}

	return &wallets, nil
}
