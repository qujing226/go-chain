package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	chain_did "github.com/qujing226/blockchain/did"
	"log"
	"os"
)

const kemWalletFile = "./components/kem_wallets.dat"

var KemWalletVersion = []byte{0x66}

type KemWallet struct {
	EncapsulationKey     [1184]byte
	DecapsulationKey     [2400]byte
	SharedSecretReceiver [32]uint8
}

func NewKemWallet() KemWallet {
	decapsulationKey, encapsulationKey, err := chain_did.GenerateKEM()
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}
	return KemWallet{
		EncapsulationKey: encapsulationKey,
		DecapsulationKey: decapsulationKey,
	}
}

func (k *KemWallet) ReceiveSecretKey(ciphertext [1088]byte) (err error) {
	// 使用私钥解密（解封共享密钥）
	sharedSecretReceiver, err := chain_did.DecryptWithKEM(k.DecapsulationKey, ciphertext)
	if err != nil {
		log.Fatalf("解密失败: %v", err)
		return err
	}
	k.SharedSecretReceiver = sharedSecretReceiver
	return nil
}

func (k *KemWallet) GetAddress() (address []byte) {
	e := make([]byte, 1184)
	copy(e, k.EncapsulationKey[:])
	pubKeyHash := HashPubKey(e)
	versionedPayload := append(KemWalletVersion, pubKeyHash...)
	check := checksum(versionedPayload)
	fullPayload := append(versionedPayload, check...)
	address = []byte(base58.Encode(fullPayload))
	return address
}

type KemWallets struct {
	KWallets map[string]*KemWallet
}

func NewKemWallets() (*KemWallets, error) {
	wallets := KemWallets{}
	wallets.KWallets = make(map[string]*KemWallet)
	err := wallets.LoadFromFile()
	return &wallets, err
}

func (kws *KemWallets) CreateWallet() string {
	kw := NewKemWallet()
	//fmt.Printf("Your kem pubKey: %s\n", base64.StdEncoding.EncodeToString(kw.EncapsulationKey[:]))
	address := fmt.Sprintf("%s", kw.GetAddress())
	kws.KWallets[address] = &kw
	return address
}

func (kws *KemWallets) GetAddresses() []string {
	var addresses []string

	for address := range kws.KWallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (kws *KemWallets) GetWallet(address string) KemWallet {
	return *kws.KWallets[address]
}

func (kws *KemWallets) LoadFromFile() error {
	if _, err := os.Stat(kemWalletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, _ := os.ReadFile(kemWalletFile)
	var kwallets KemWallets

	err := json.Unmarshal(fileContent, &kwallets)
	if err != nil {
		return err
	}

	kws.KWallets = kwallets.KWallets

	return nil
}

// SaveToFile saves wallets to a file
func (kws *KemWallets) SaveToFile() error {
	buf, err := json.Marshal(kws)
	if err != nil {
		return err
	}

	err = os.WriteFile(kemWalletFile, buf, 0644)
	return err
}
