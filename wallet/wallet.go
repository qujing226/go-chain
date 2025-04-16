package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"math/big"
)

var (
	version            = []byte{0x00}
	addressChecksumLen = 4
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := &Wallet{private, public}

	return wallet
}

func (w *Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionedPayload := append(version, pubKeyHash...)
	check := checksum(versionedPayload)
	fullPayload := append(versionedPayload, check...)
	address := []byte(base58.Encode(fullPayload))
	return address
}

// BytesToPublicKey converts a byte slice to an ecdsa.PublicKey
func BytesToPublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	curve := elliptic.P256()
	keyLen := curve.Params().BitSize / 8

	if len(pubKeyBytes) != 2*keyLen {
		return nil, fmt.Errorf("invalid public key length: expected %d, got %d", 2*keyLen, len(pubKeyBytes))
	}

	x := new(big.Int).SetBytes(pubKeyBytes[:keyLen])
	y := new(big.Int).SetBytes(pubKeyBytes[keyLen:])

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		fmt.Println("HashPubKey err:", err)
		return nil
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

func ValidateAddress(address string) bool {
	pubKeyHash := base58.Decode(address)
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	vers := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetCheckSum := checksum(append([]byte{vers}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetCheckSum) == 0
}
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("生成私钥失败")
		return *new(ecdsa.PrivateKey), nil
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:4]
}

type _PrivateKey struct {
	D          *big.Int
	PublicKeyX *big.Int
	PublicKeyY *big.Int
}

func (w *Wallet) GobEncode() ([]byte, error) {
	privKey := &_PrivateKey{
		D:          w.PrivateKey.D,
		PublicKeyX: w.PrivateKey.PublicKey.X,
		PublicKeyY: w.PrivateKey.PublicKey.Y,
	}

	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(privKey)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(w.PublicKey)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	var privKey _PrivateKey

	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&privKey)
	if err != nil {
		return err
	}

	w.PrivateKey = ecdsa.PrivateKey{
		D: privKey.D,
		PublicKey: ecdsa.PublicKey{
			X:     privKey.PublicKeyX,
			Y:     privKey.PublicKeyY,
			Curve: elliptic.P256(),
		},
	}
	w.PublicKey = make([]byte, buf.Len())
	_, err = buf.Read(w.PublicKey)
	if err != nil {
		return err
	}

	return nil
}
