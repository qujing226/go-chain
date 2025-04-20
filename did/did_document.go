package chain_did

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	ssi "github.com/nuts-foundation/go-did"
	"github.com/nuts-foundation/go-did/did"
	"log"
	"math/big"
)

const method = "easyblock"

func GenerateDidDocument(pubKey *ecdsa.PublicKey) *did.Document {
	hash := pubKeyHash(pubKey)
	idBytes := hash[len(hash)-20:]
	// 构造唯一标识符
	identifier := base58.Encode(idBytes)
	didStr := fmt.Sprintf("did:%s:%s", method, identifier)
	fmt.Println(didStr)
	// 解析didStr 为DID类型
	newDid, err := did.ParseDID(didStr)
	if err != nil {
		log.Fatalf("解析 DID 失败: %v", err)
	}
	// 构造 DID document
	doc := &did.Document{
		Context: []any{"https://www.w3.org/ns/did/v1", "https://www.w3.org/ns/did/v1.1"},
		ID:      *newDid,
	}

	// 构造用于验证 DID URL （分片）
	keyID, err := did.ParseDIDURL(didStr + "#authentication-key")
	if err != nil {
		log.Fatalf("解析 DID URL 失败: %v", err)
	}

	verifucationMethod, err := did.NewVerificationMethod(*keyID, ssi.JsonWebKey2020, *newDid, pubKey)
	if err != nil {
		log.Fatalf("创建验证方法失败: %v", err)
	}

	// 将验证方法添加到 DID document 中作为assertionMethod
	doc.AddAssertionMethod(verifucationMethod)

	return doc
}

func UpdateDidDocument(doc *did.Document, pubKey [1184]byte) *did.Document {
	didStr := doc.ID.String()
	return addLatticeKeyToDidDocument(doc, didStr, pubKey)
}

func SerializeDidDocument(doc *did.Document) ([]byte, error) {
	docJson, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return []byte{}, err
	}
	return docJson, nil
}

func DeserializeDidDocument(docJson []byte) (*did.Document, error) {
	var doc did.Document
	if err := json.Unmarshal(docJson, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

func VerifyDidDocument(doc *did.Document, pubKey ecdsa.PublicKey) bool {
	if doc == nil {
		return false
	}
	for _, m := range doc.AssertionMethod {
		if publicKeyEqual(m.PublicKey, pubKey) {
			return true
		}
	}
	return false
}

func GenerateChallenge() []byte {
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		log.Fatalf("生成随机数失败: %v", err)
	}
	return challenge
}

func VerifyChallengeSignature(challenge []byte, r, s *big.Int, pubKey *ecdsa.PublicKey) bool {
	return ecdsa.Verify(pubKey, challenge, r, s)
}

// publicKeyEqual 用于判断存储在 VerifyMethod 的公钥和传入的公钥是否相等
func publicKeyEqual(keyStored any, pubKey ecdsa.PublicKey) bool {
	switch k := keyStored.(type) {
	case *ecdsa.PublicKey:
		return k.X.Cmp(pubKey.X) == 0 && k.Y.Cmp(pubKey.Y) == 0 && k.Curve == pubKey.Curve
	case ecdsa.PublicKey:
		return k.X.Cmp(pubKey.X) == 0 && k.Y.Cmp(pubKey.Y) == 0 && k.Curve == pubKey.Curve
	default:
		return false
	}
}

func pubKeyHash(key *ecdsa.PublicKey) []byte {
	pubHash := elliptic.MarshalCompressed(key.Curve, key.X, key.Y)

	hash := sha256.Sum256(pubHash)
	return hash[:]
}
