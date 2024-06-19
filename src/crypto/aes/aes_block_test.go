package aes_test

import (
	"crypto/aes"
	"crypto/internal/cryptotest"
	"fmt"
	"testing"
)

func TestAESBlock(t *testing.T) {
	for _, keylen := range []int{128, 192, 256} {
		t.Run(fmt.Sprintf("AES-%d", keylen), func(t *testing.T) {
			cryptotest.TestBlock(t, keylen/8, aes.NewCipher)
		})
	}
}
