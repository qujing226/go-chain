package chain_did

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_generateKEM(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{
			name: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kem1, q1, _ := GenerateKEM()
			kem2, q2, _ := GenerateKEM()
			fmt.Println(kem1)
			fmt.Println(kem2)
			fmt.Println(q1)
			fmt.Println(q2)
			assert.NotEqual(t, kem1, kem2)

		})
	}
}

//func Test_EncryptWithKEM(t *testing.T) {
//	tests := []struct {
//		name string
//		encapsulationKey []byte
//
//	}{
//		// TODO: Add test cases.
//		{
//			name: "test",
//			encapsulationKey: []byte(),
//		},
//	}
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			sharedSecret,text,err:= EncryptWithKEM(tc.encapsulationKey)
//			require.NoError(t,err)
//			fmt.Println(sharedSecret,text)
//		})
//	}
//}

func Test_DecryptWithKEM(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "test",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			de, en, err := GenerateKEM()
			require.NoError(t, err)
			sharedSecret, secText, err := EncryptWithKEM(en)
			require.NoError(t, err)
			sharedSecretReceiver, err := DecryptWithKEM(de, secText)
			require.NoError(t, err)
			require.Equal(t, sharedSecret, sharedSecretReceiver)
		})
	}
}
