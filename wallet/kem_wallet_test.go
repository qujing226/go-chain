package wallet

import (
	"encoding/json"
	"fmt"
	chain_did "github.com/qujing226/blockchain/did"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// TestNewKemWallet tests the NewKemWallet function
func TestNewKemWallet(t *testing.T) {
	kw := NewKemWallet()
	assert.NotNil(t, kw)
	assert.NotEmpty(t, kw.EncapsulationKey)
	assert.NotEmpty(t, kw.DecapsulationKey)
}

// TestReceiveSecretKey tests the ReceiveSecretKey method
func TestReceiveSecretKey(t *testing.T) {
	kw := NewKemWallet()
	_, ciphertext, err := chain_did.EncryptWithKEM(kw.EncapsulationKey)
	assert.NoError(t, err)

	err = kw.ReceiveSecretKey(ciphertext)
	assert.NoError(t, err)
	fmt.Println("SharedSecretReceiver:", kw.SharedSecretReceiver)
	assert.NotEmpty(t, kw.SharedSecretReceiver)
}

// TestGetAddress tests the GetAddress method
func TestGetAddress(t *testing.T) {
	kw := NewKemWallet()
	address := kw.GetAddress()
	assert.NotNil(t, address)
	assert.NotEmpty(t, address)
}

// TestKemWallets_NewKemWallets tests the NewKemWallets function
func TestKemWallets_NewKemWallets(t *testing.T) {
	kws, err := NewKemWallets()
	assert.Error(t, err) // Expect error since the file does not exist

	// Create a test file
	testData := KemWallets{
		KWallets: make(map[string]*KemWallet),
	}
	k := NewKemWallet()
	testData.KWallets["testAddress"] = &k

	testDataJSON, _ := json.Marshal(testData)
	os.WriteFile(kemWalletFile, testDataJSON, 0644)
	defer os.Remove(kemWalletFile)

	kws, err = NewKemWallets()
	assert.NoError(t, err)
	assert.NotNil(t, kws)
	assert.NotEmpty(t, kws.KWallets)
}

// TestKemWallets_CreateWallet tests the CreateWallet method
func TestKemWallets_CreateWallet(t *testing.T) {
	kws := &KemWallets{
		KWallets: make(map[string]*KemWallet),
	}

	address := kws.CreateWallet()
	assert.NotEmpty(t, address)
	assert.NotNil(t, kws.KWallets[address])
}

// TestKemWallets_GetAddresses tests the GetAddresses method
func TestKemWallets_GetAddresses(t *testing.T) {
	kws := &KemWallets{
		KWallets: make(map[string]*KemWallet),
	}

	// Create two wallets
	address1 := kws.CreateWallet()
	address2 := kws.CreateWallet()

	// Ensure the addresses are unique
	assert.NotEqual(t, address1, address2)

	addresses := kws.GetAddresses()
	assert.Len(t, addresses, 2)
	assert.Contains(t, addresses, address1)
	assert.Contains(t, addresses, address2)
}

// TestKemWallets_GetWallet tests the GetWallet method
func TestKemWallets_GetWallet(t *testing.T) {
	kws := &KemWallets{
		KWallets: make(map[string]*KemWallet),
	}

	address := kws.CreateWallet()
	kw := kws.GetWallet(address)
	assert.NotNil(t, kw)
}

// TestKemWallets_LoadFromFile tests the LoadFromFile method
func TestKemWallets_LoadFromFile(t *testing.T) {
	kws := &KemWallets{
		KWallets: make(map[string]*KemWallet),
	}

	testData := KemWallets{
		KWallets: make(map[string]*KemWallet),
	}
	k := NewKemWallet()
	testData.KWallets["testAddress"] = &k

	testDataJSON, _ := json.Marshal(testData)
	os.WriteFile(kemWalletFile, testDataJSON, 0644)
	defer os.Remove(kemWalletFile)

	err := kws.LoadFromFile()
	assert.NoError(t, err)
	assert.NotEmpty(t, kws.KWallets)
}

// TestKemWallets_SaveToFile tests the SaveToFile method
func TestKemWallets_SaveToFile(t *testing.T) {
	kws := &KemWallets{
		KWallets: make(map[string]*KemWallet),
	}

	kws.CreateWallet()

	err := kws.SaveToFile()
	assert.NoError(t, err)

	// Verify the file exists
	_, err = os.Stat(kemWalletFile)
	assert.NoError(t, err)
}
