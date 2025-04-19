package chain_did

import (
	"github.com/nuts-foundation/go-did/did"
	"testing"
)

func Test_addLatticeKeyToDidDocument(t *testing.T) {
	before := func(s string) *did.Document {
		doc, err := did.ParseDocument(s)
		if err != nil {
			panic(err)
		}
		return doc
	}
	tests := []struct {
		name      string
		doc       string
		didStr    string
		publicKey string
	}{
		// TODO: Add test cases.
		{
			name: "success",
			doc: `{
    "@context": "https://www.w3.org/ns/did/v1",      // 定义文档语义环境，DID 版本信息
    "assertionMethod": [                            // 制定了哪些验证方法可用于创建断言（assertions），比如在需要证明其身份时用来签名挑战或事务。
            "did:example:123#key-1"                 // 直接引用verificationMethod定义的密钥表明进行身份认证或生命操作时，用使用此密钥对数据进行签名
        ],
    "id": "did:example:123",                        // 唯一标识符
    "verificationMethod": [
        {
            "controller": "did:example:123",        // 表示密钥的控制者，指向链方，也可指向本人
            "id": "did:example:123#key-1",          // 验证方法的唯一标识符（使用DID的fragment语法），表示出DID内的具体以密钥
            "publicKeyJwk": {
                "crv": "P-256",                     // 椭圆曲线标准
                "kty": "EC",                        // 密钥类型、键类型
                "x": "szMZLf4xQXiIdaelIQ6YaZZf8M2Y1DxN6ceAb3evwtA",  // 公钥的x坐标
                "y": "ybr-cOxWHeRMRhcVLZa-bGXDOSsRmADoyVUMkoqmAGk"   // 公钥的y坐标
            },
            "type": "JsonWebKey2020"
        }
    ]
}`,
			didStr: "did:example:123",
			publicKey: `{
    "crv": "P-256",
    "kty": "EC",
    "x": "szMZLf4xQXiIdaelIQ6YaZZf8M2Y1DxN6ceAb3evwtA",
    "y": "ybr-cOxWHeRMRhcVLZa-bGXDOSsRmADoyVUMkoqmAGk"
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := before(tt.doc)
			addLatticeKeyToDidDocument(doc, tt.didStr, tt.publicKey)
		})
	}
}
