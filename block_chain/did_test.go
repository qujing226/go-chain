package chain

import (
	"fmt"
	"testing"
)

func TestFindDidDocument(t *testing.T) {
	tests := []struct {
		name      string
		targetdid string
	}{
		// TODO: Add test cases.
		{
			name:      "success.",
			targetdid: "did:easyblock:4LxQvAd4QGvQsLWL3iK6LdN9Xzvk",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDidDocument(NewBlockChain("3003"), tt.targetdid)
			fmt.Printf("%+v", *got)
		})
	}
}
