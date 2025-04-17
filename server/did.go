package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	_ "embed"
	"encoding/base64"
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
	"time"
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
	DID       string `json:"did"`
	Address   string `json:"address"`
	Challenge string `json:"challenge"`
	Signature string `json:"signature"` // 完整的 base64 格式签名（64字节签名编码后的字符串）
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
		return
	}

	ch := chaindid.GenerateChallenge()
	err = storeChallengeFromCache(req.DID, ch)
	if err != nil {
		fmt.Println("store challenge error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "redis存储challenge失败", "error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "success",
		"challenge": ch})
}

// VerifyDid 验证 challenge
func VerifyDid(ctx *gin.Context) {
	var req verifyDidDocumentReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误", "error": err.Error()})
		return
	}

	// 1. 解码 challenge（hex 编码转为字节数组）
	ch, err := getChallengeFromCache(req.DID)
	if err != nil {
		fmt.Println("get challenge error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "get challenge error", "error": err.Error()})
		return
	}

	sigBytes, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		fmt.Println("signature decode error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "signature decode error", "error": err.Error()})
		return
	}

	// 2. 解码签名的 r 和 s 部分
	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])

	pubKey, err := getPubKeyFromCache(req.DID)
	if err != nil {
		fmt.Println("get pubkey error")
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "获取公钥失败", "error": err.Error()})
		return
	}

	if !chaindid.VerifyChallengeSignature(ch, r, s, pubKey) {
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
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	// 确保 xBytes 和 yBytes 长度为 32 字节
	xBytes = fixedBytes(xBytes, 32)
	yBytes = fixedBytes(yBytes, 32)

	// 拼接 xBytes 和 yBytes
	buf := append(xBytes, yBytes...)

	err := rdb.Eval(context.Background(), storePubKey, []string{didStr}, buf).Err()
	if err != nil {
		return err
	}
	return nil
}

func fixedBytes(b []byte, size int) []byte {
	if len(b) > size {
		return b[:size]
	}
	return append(make([]byte, size-len(b)), b...)
}

func getPubKeyFromCache(didStr string) (*ecdsa.PublicKey, error) {
	pubKeyData, err := rdb.Eval(context.Background(), getPubKey, []string{didStr}, []interface{}{}).Result()
	if err != nil {
		return nil, err
	}

	var buf []byte
	switch v := pubKeyData.(type) {
	case []byte:
		buf = v
	case string:
		buf = []byte(v)
	default:
		return nil, fmt.Errorf("invalid data type: expected []byte or string, got %T", pubKeyData)
	}

	if len(buf) != 64 {
		return nil, fmt.Errorf("invalid public key length: expected 64 bytes, got %d bytes", len(buf))
	}

	xBytes := buf[:32]
	yBytes := buf[32:]

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)

	curve := elliptic.P256()
	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}

	return pubKey, nil
}

func storeChallengeFromCache(didStr string, ch []byte) error {
	err := rdb.SetEx(context.Background(), didStr+":challenge", ch, 5*time.Minute).Err()
	if err != nil {
		return err
	}
	return nil
}

func getChallengeFromCache(didStr string) ([]byte, error) {
	challengeData, err := rdb.Get(context.Background(), didStr+":challenge").Result()
	if err != nil {
		return nil, err
	}
	return []byte(challengeData), nil
}
