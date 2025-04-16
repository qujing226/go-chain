package server

import (
	"context"
	"crypto/ecdsa"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nuts-foundation/go-did/did"
	chain "github.com/qujing226/blockchain/block_chain"
	chaindid "github.com/qujing226/blockchain/did"
	"github.com/qujing226/blockchain/wallet"
	"github.com/redis/go-redis/v9"
	"log"
	"math/big"
	"net/http"
	"os"
)

//go:embed store_publicKey.lua
var storePubKey string

//go:embed get_publicKey.lua
var getPubKey string

var nodeID string //
var bc *chain.BlockChain
var ws *wallet.Wallets

var rdb redis.Cmdable

func InitConfig() {
	nodeID = os.Getenv("NODE_ID") //
	bc = chain.NewBlockChain(nodeID)
	wallets, err := wallet.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	ws = wallets
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:16379",
		Password: "",
	})
}

func StartDidService() {
	server := gin.Default()
	RegisterRoutes(server)
	_ = server.Run(":8080")
}

func RegisterRoutes(s *gin.Engine) {
	s.POST("/did/create", CreateDidDocument)
	s.POST("/did/challenge", SendChallenge)
	s.POST("/did/verify", VerifyDid)
}

func CreateDidDocument(ctx *gin.Context) {
	var req struct {
		Address string `json:"address"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"message": "address error",
		})
	}
	// 这段代码用于参数接受公钥的情况
	//pubkey, err := decodePubKey(req.pubkey)

	w := ws.GetWallet(req.Address)
	p, err := decodePubKey(w.PublicKey)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(200, gin.H{
			"message": "decode pulickey error",
		})
	}
	doc := chaindid.GenerateDidDocument(p)

	go func() {
		SaveDocToBlockChain(&w, doc)
	}()
	// doc 上链

	ctx.JSON(200, gin.H{
		"did_document": doc,
	})
}

type verifyDidDocumentReq struct {
	DID     string `json:"did"`
	Address string `json:"address"`
	// Challenge 经由服务器生成后发给用户，用户计算私钥签名后传回，假设以 hex 格式编码
	Challenge string `json:"challenge"`

	// 签名组成部分（r, s），均以 hex 字符串传输
	SignatureR string `json:"signature_r"`
	SignatureS string `json:"signature_s"`
}

func SendChallenge(ctx *gin.Context) {
	var req verifyDidDocumentReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"message": "请求格式错误"})
		return
	}
	// 根据链查询用户传入的did对应的document
	doc := chain.FindDidDocument(bc, req.DID)
	if doc == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "未找到 DID document"})
		return
	}
	// 检查 DID Document 是否拥有 assertionMethod（验证方法）
	if len(doc.AssertionMethod) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "DID document missing assertion method",
		})
		return
	}

	// 提取 DID Document 中第一个验证方法的公钥。注意：在 nuts-foundation/go-did 中，
	// VerificationMethod 的 PublicKey 字段通常是一个函数，需调用后获取实际公钥
	rawPub, err := doc.AssertionMethod[0].PublicKey()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "获取公钥失败", "error": err.Error()})
		return
	}
	pubKey, ok := rawPub.(*ecdsa.PublicKey)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "DID document 中公钥类型错误"})
		return
	}

	// todo:将公钥数据编码后存入redis
	err = storePubKeyToCache(req.DID, pubKey)
	if err != nil {
		fmt.Println("store pubkey error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "redis存储公钥失败", "error": err.Error()})
	}

	ch := chaindid.GenerateChallenge()

	ctx.JSON(200, gin.H{"message": "success",
		"challenge": ch})
}

// VerifyDid 实现整个验证流程
//
// 流程说明：
//  1. Server 根据请求中的 DID 查询链上存储的 DID Document。
//  2. 从 DID Document 中提取登记时的公钥（assertionMethod 部分）。
//  3. 将请求中的 challenge 和签名（r、s）解码（假设均为 hex 格式）。
//  4. 使用 ECDSA 签名验证算法验证该签名是否由该公钥产生。
//  5. 如果本地验证通过，Server 将签名数据（例如 DID、Address、challenge、签名）转给 Issuer；
//     此处示例采用一个模拟的函数 simulateIssuerVerification 来表示这一过程。
//  6. Issuer 返回验证结果后，Server 返回最终结果给客户端。
func VerifyDid(ctx *gin.Context) {
	var req verifyDidDocumentReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误", "error": err.Error()})
		return
	}

	// 1. 解码 challenge（hex 编码转为字节数组）
	challengeBytes, err := hex.DecodeString(req.Challenge)
	if err != nil {
		fmt.Println("challenge decode error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "challenge decode error", "error": err.Error()})
		return
	}

	// 2. 解码签名的 r 和 s 部分
	r := new(big.Int)
	s := new(big.Int)
	if _, ok := r.SetString(req.SignatureR, 16); !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "signature_r 格式错误"})
		return
	}

	// 3. 从内存中读取
	if _, ok := s.SetString(req.SignatureS, 16); !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "signature_s 格式错误"})
		return
	}

	pubKey, err := getPubKeyFromCache(req.DID)
	if err != nil {
		fmt.Println("get pubkey error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "获取公钥失败", "error": err.Error()})
		return
	}

	if !chaindid.VerifyChallengeSignature(challengeBytes, r, s, pubKey) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "本地签名验证失败"})
		return
	}

	// 若一切验证成功，则返回 success
	ctx.JSON(http.StatusOK, gin.H{"message": "Verification successful"})
}

func SaveDocToBlockChain(w *wallet.Wallet, doc *did.Document) {
	b, err := chaindid.SerializeDidDocument(doc)
	if err != nil {
		log.Println("serialize did document error")
	}
	tx := chain.NewDidDocumentTransaction(w, b)
	// 向中心节点发送 did 交易
	fmt.Println("正在将 DID document 发送至中心节点")
	SendTx(tx)
}

// decodePubKey 用于解码字符串 ——> 公钥，但本项目采用地址读取公私钥
func decodePubKey(pubKey []byte) (*ecdsa.PublicKey, error) {
	// 解码公钥字符串，返回一个 *ecdsa.PublicKey 类型的公钥
	// 此处略去详细实现逻辑
	// 解码Base64字符串为字节
	//pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKey)
	//if err != nil {
	//	return nil, err
	//}

	// 将字节转换为公钥
	pubkey, err := wallet.BytesToPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	return pubkey, nil
}

func storePubKeyToCache(didStr string, pubKey *ecdsa.PublicKey) error {
	buf, err := json.MarshalIndent(pubKey, "", "  ")
	if err != nil {
		return err
	}
	err = rdb.Eval(context.Background(), storePubKey, []string{didStr}, buf).Err()
	if err != nil {
		return err
	}
	return nil
}

func getPubKeyFromCache(didStr string) (*ecdsa.PublicKey, error) {
	pubKey, err := rdb.Eval(context.Background(), getPubKey, []string{didStr}, nil).Result()
	if err != nil {
		return nil, err
	}
	var pubkey *ecdsa.PublicKey
	if err := json.Unmarshal([]byte(pubKey.(string)), &pubkey); err != nil {
		return nil, err
	}
	return pubkey, nil
}
