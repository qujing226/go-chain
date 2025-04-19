package chain_did

import (
	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/did"
	"log"
)

func addLatticeKeyToDidDocument(doc *did.Document, didStr string, publicKey string) *did.Document {
	keyID, err := did.ParseDIDURL(didStr + "#lattice-key")
	if err != nil {
		panic(err)
	}
	newDid, err := did.ParseDID(didStr)
	if err != nil {
		log.Fatalf("解析 DID 失败: %v", err)
	}
	// 创建一个新的验证方法
	verificationMethod, err := did.NewVerificationMethod(*keyID, ssi.JsonWebKey2020, *newDid, publicKey)
	if err != nil {
		panic(err)
	}
	// 将新的验证方法添加到文档的 verificationMethod 数组
	doc.VerificationMethod = append(doc.VerificationMethod, verificationMethod)
	doc.AddAssertionMethod(verificationMethod)
	return doc
}
