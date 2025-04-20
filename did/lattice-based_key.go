package chain_did

import (
	"crypto"
	"encoding/base64"
	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/did"
	"log"
)

func addLatticeKeyToDidDocument(doc *did.Document, didStr string, publicKey [1184]byte) *did.Document {
	keyID, err := did.ParseDIDURL(didStr + "#lattice-key")
	if err != nil {
		panic(err)
	}
	newDid, err := did.ParseDID(didStr)
	if err != nil {
		log.Fatalf("解析 DID 失败: %v", err)
	}

	// 创建一个新的验证方法
	verificationMethod, err := did.NewVerificationMethod(*keyID, ssi.KemJsonKey2025, *newDid, publicKey)
	if err != nil {
		panic(err)
	}
	// 将新的验证方法添加到文档的 verificationMethod 数组
	doc.VerificationMethod = append(doc.VerificationMethod, verificationMethod)
	doc.AddAssertionMethod(verificationMethod)
	doc.Context = append(doc.Context, did.DIDContextV1URI())
	return doc
}

type JWKPublicKey struct {
	JWK map[string]interface{}
}

func (j JWKPublicKey) PublicKey() crypto.PublicKey {
	return j
}
func ConvertToJWK(publicKey [1184]byte) JWKPublicKey {
	return JWKPublicKey{
		JWK: map[string]interface{}{
			"kty": "KYBER",                                            // 密钥类型
			"crv": "Kyber768",                                         // 安全级别
			"x":   base64.RawURLEncoding.EncodeToString(publicKey[:]), // 公钥数据
		},
	}
}
