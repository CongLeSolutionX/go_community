package cipher_test

import (
	"crypto/cipher"
	"crypto/internal/cryptotest"
	"testing"
)

func TestCBCBlockMode(t *testing.T) {
	cryptotest.TestBlockMode(t, cipher.NewCBCEncrypter, cipher.NewCBCDecrypter)
}
