# go-chain
使用Go搭建的简单区块链

创世块地址 14sYRQmgWqnfm1fezbBdZaYTae9hq4ohFj
address1 15oNPjxBvSWdTuqXWYMaJjC8S7q5U46KQJ
address2 1J7v5bua2tLRovJmoQ32C7xrhMqvX1z8Qw    did

矿工地址  16DgKNaFExb43aaXoiWGcDrQd3XUm1KPky  


# DID 设计说明

## 1. DID 文档构成
- **id**: 全局唯一的 DID 标识，例如：did:example:abcd1234
- **publicKey**: 用户的公钥字符串，用于数字签名验证
- **authentication**: 描述验证方法，例如 "EcdsaSecp256k1"。在认证时，将使用该公钥对来自用户的数字签名进行验证
- **service**: （选填）服务端点信息，未来可用于扩展更多互动功能

## 2. 用户注册与存根
用户需先构造一笔 **DID 注册交易**：
1. 填入上述 DID 信息，生成对应的交易。此交易将链上记录，作为用户身份的依据。
2. 后续用户在交易过程中，如签名交易或调用智能合约时，需要携带该 DID 标识，并用对应私钥生成签名。

## 3. 后续 DID 认证流程
1. 用户发起认证请求，附加 DID 标识、需要验证的内容及签名。
2. 系统获取链上对应 DID 交易记录中的公钥，利用公钥验证用户提交的签名，完成认证。


``` text
{
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
}
```


``` md
{
    "did_document": {
        "@context": [                                                                 // 定义文档语义环境，DID 版本信息
            "https://www.w3.org/ns/did/v1",
            "https://www.w3.org/ns/did/v1.1",
        ],
        "assertionMethod": [                                                          // 制定了哪些验证方法可用于创建断言（assertions），比如在需要证明其身份时用来签名挑战或事务。
            "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk#authentication-key",          // 直接引用verificationMethod定义的密钥表明进行身份认证或生命操作时，用使用此密钥对数据进行签名
            "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk#lattice-key"
        ],
        "id": "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk",                           // 唯一标识符
        "verificationMethod": [
            {
                "controller": "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk",           // 表示密钥的控制者，指向链方，也可指向本人
                "id": "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk#authentication-key",// 验证方法的唯一标识符（使用DID的fragment语法），表示出DID内的具体以密钥
                "publicKeyJwk": {
                    "crv": "P-256",
                    "kty": "EC",
                    "x": "dHddVNV6HAlbB3XxG2eROxfEJ1SsuBxJEamT5kLOjLw",
                    "y": "GPDYdgJDqs_RzrtI9s13hL5LogG6Z400lmwEY86lWWE"
                },
                "type": "JsonWebKey2020"
            },
            { 
                "controller": "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk",
                "id": "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk#lattice-key",
                "publicKeyJwk": {
                    "crv": "Kyber768",
                    "kty": "KYBER",
                    "x": "LAtS3tlfw9hi_4t2oHyUv8kVFQSlOOgfTvQAMtEk1eiAqISqSDk0UrZBfzVsGUJuvQeua7iOuAgliSwqloNwMNd8roW-FMBngpImsHlvBKWQ8XJafmCMTEwA8obGmro7z7YL1TkCOmhOH3u29yZq5NxekuMPNqDGbyDOdcwmXhw78mMj1OxQdsbCSLCO6vqKxfU6vqCnSMiLxuw5zphOYKsYQSFwu5RNknqWVomfJ7rCQYQ1v_PNwtCpjLF6GetjMgSJkiZb3gMn72RUTmYZlXXBbfEeB2ifYAx1G8RD_GWO79ON45xcOPSDhEo3FoA9NVqRfouUhzh5JrSATRPM0zcQZqEVMcu_iTSaiqGHSyqCAjzGdrCkGwDCdaaZT8c9MyJjywJ35JRHj2cTX-GAQjKuNFVNlLiOuIDFblejZfaXtWRHswly2XPD2RcfROeAwbdFDSRdiUx6fpLKD2F-8eixUSRahbEqGJINwdCcyfuWMlw5VtA6uhoLq-kngVWwQ1E1QVishMgM0RokAoYG3qqpWrli-BgYLDqlU2mPniR7QATHYdIe9HY5kaQbnzM6NlbE7vuk1EJWLfLJCbVD14KDNuFI52FFMKXL-yW0TIN-Ksg25fzKATW-NuTGfrqFw3omjqFIiDeii1yRqgCbMoYsIYeAeoxLj8uhXzoma_xXlrE-rzrGtgguuJCaLDG-uxNfs2MEaqYKPExstjcQuQcLlngbXSIQ_qE7kkDLKyAi7yaRUqdNobfEy2xbnvtWj9wnDkoaiRd-_TKwaLl0U0uFgMxdVvvEc0cRsHOIC3UTg9J2qzFTW0obDnOS5ZwDP6sPN-qPMPaSQBg384O2UxEZWCwNTTx52LZnUMlg5jwgN5TGW9lR64c3Kkenqzx3uTIjT1Mi8GebPckJUKBPw_QZXKV52QeFK1lRwDy6W2Jhy9RKbdizXiKnYmzK34hLiYuxL4BxP7o_7RhcSQBi5Hy5KvbLA7SEnINyFJDEAXKqyOlJdccPUkpQv-Cp1-I6rki6A9lgvgiXXHiF_ridY5qjkjo1wWkBWnNZYbhk2GcJ20CdUekWo2WiRaun_UC5NlkVzTtpXvWIqDV_y_hU2UjBZzIuzdJZK4U98mZy7knKwswYavEpXfGH9jlAoiQhH7Qe3WcYLik6bjqg4Mdm63wuVZlp34BtjWcsQBiYs0TIQnt1Z8EbejNfL-Ihrmq3l0pE_QQ6ypVqFnI8JQCGmYgmF4o7REtcEOrM_vVArvwqVTIKsbUPhSu6vDxrCDQB9HO8mgwB8nJ8sWsf49uOSINoEMlylgGBHbuo6ec3S8yl2ObLHIQGgZhIO7yIMqWkrGGdoKwGxFIn3tgK-JqyvhdRgvkZVvuqgHHPEuEQ1EBipVBxKEXG13bLYbC6MnnH8CN_3yhbF3uKf9CeZ0HKWTeifwSemRIEwOCyGYciFIJEpNlaTWQEu4gz7NUK7fKgCEgXoiQ-_mORZiN1n5aqCyenpqwW2hufmhehKFGPg7CYxeYOj_w8lMOs2Pgcp5qTyXPk2tUgLf6VFXhgGkdQq98EDga5scRk92zZlKETdWU"
                },
                "type": "KemJsonKey2025"                                              // 标识用于KEM加密
            }
        ]
    }
}
```
``` md
type Block struct {
	TimeStamp    int64
	PreBlockHash []byte
	Hash         []byte
	Nonce        int
	Height       int
	Transactions []*Transaction // 存储所有交易（其中包含 DID 文档相关交易）
}


type Transaction struct {
	ID        []byte
	Vin       []TXInput
	Vout      []TXOutput
	TimeStamp int64
	Payload   []string
}


type Document struct {
	Context              []interface{}             
	ID                   DID                       
	Controller           []DID                     
	AlsoKnownAs          []ssi.URI                 
	VerificationMethod   VerificationMethods       
	Authentication       VerificationRelationships 
	AssertionMethod      VerificationRelationships 
	KeyAgreement         VerificationRelationships 
	CapabilityInvocation VerificationRelationships 
	CapabilityDelegation VerificationRelationships 
	Service              []Service                 
}

```