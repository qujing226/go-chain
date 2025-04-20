package chain_did

import (
	"log"

	kyberk2so "github.com/symbolicsoft/kyber-k2so"
)

// GenerateKEM 生成密钥对（私钥和公钥） 先返回私钥，在返回公钥
func GenerateKEM() ([2400]byte, [1184]byte, error) {
	return kyberk2so.KemKeypair768()
}

func EncryptWithKEM(encapsulationKey [1184]byte) (sharedSecretSender [32]byte, ciphertext [1088]byte, err error) {
	// 使用公钥加密（封装共享密钥）
	ciphertext, sharedSecretSender, err = kyberk2so.KemEncrypt768(encapsulationKey)
	if err != nil {
		log.Fatalf("加密失败: %v", err)
	}
	return sharedSecretSender, ciphertext, err

}

func DecryptWithKEM(decapsulationKey [2400]byte, ciphertext [1088]byte) (sharedSecretReceiver [32]uint8, err error) {
	// 使用私钥解密（解封共享密钥）
	sharedSecretReceiver, err = kyberk2so.KemDecrypt768(ciphertext, decapsulationKey)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
	}
	return sharedSecretReceiver, err
}
