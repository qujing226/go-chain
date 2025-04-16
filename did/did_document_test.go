package chain_did

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"math/big"
	"testing"
)

func TestGenerateEcdsaKey(t *testing.T) {
	// 生成 ECDSA P-256 密钥对
	keyPair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("生成密钥失败: %v", err)
	}

	// 将私钥序列化为 DER 格式
	derPriv, err := x509.MarshalECPrivateKey(keyPair)
	if err != nil {
		t.Fatalf("序列化私钥失败: %v", err)
	}
	// 使用 PEM 格式进行编码，类型为 "EC PRIVATE KEY"
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: derPriv,
	})
	fmt.Printf("Private Key:\n%s\n", privPem)

	// 将公钥序列化为 DER 格式 (PKIX 格式)
	derPub, err := x509.MarshalPKIXPublicKey(&keyPair.PublicKey)
	if err != nil {
		t.Fatalf("序列化公钥失败: %v", err)
	}
	// 使用 PEM 格式进行编码，类型为 "PUBLIC KEY"
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPub,
	})
	fmt.Printf("Public Key:\n%s\n", pubPem)
}

func TestGenerateDidDocument(t *testing.T) {
	pubKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	t.Log("GenerateDidDocument")
	tc := []struct {
		name   string
		pubKey ecdsa.PublicKey
	}{
		{
			name:   "success",
			pubKey: pubKey.PublicKey,
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			doc := GenerateDidDocument(tt.pubKey)
			docJson, err := json.MarshalIndent(doc, "", "  ")
			require.NoError(t, err)
			fmt.Println(string(docJson))
		})
	}
}

func TestVerifyVerifyDidDocument(t *testing.T) {
	keyPair, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	doc := GenerateDidDocument(keyPair.PublicKey)
	docJSON, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)
	fmt.Println("generate did document: ")
	fmt.Println(string(docJSON))

	// 验证流程
	challenge := GenerateChallenge()
	fmt.Printf("challenge: %x\n", challenge)

	// 私钥签名
	r, s := signChallenge(challenge, keyPair)
	fmt.Printf("Signature: (r: %s,s %s)\n", r.String(), s.String())

	// 读取 DID document中的验证方法, 获得公钥
	vm := doc.AssertionMethod[0]

	// 此处提取公钥类型需要与生成密钥时一致
	rawPub, err := vm.PublicKey()
	require.NoError(t, err)
	extractedPub, ok := rawPub.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("提取公钥失败")
	}

	if VerifyChallengeSignature(challenge, r, s, extractedPub) {
		fmt.Println("challenge verify success")

	} else {
		fmt.Println("challenge verify failed")
	}

}

func signChallenge(challenge []byte, priv *ecdsa.PrivateKey) (r *big.Int, s *big.Int) {
	r, s, err := ecdsa.Sign(rand.Reader, priv, challenge)
	if err != nil {
		log.Fatalf("sign challenge failure %v", err)
	}
	return r, s
}
