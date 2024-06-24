// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher_test

import (
	"crypto/cipher"
	"crypto/internal/cryptotest"
	"testing"
)

func TestCBCBlockMode(t *testing.T) {
	cryptotest.TestBlockMode(t, cipher.NewCBCEncrypter, cipher.NewCBCDecrypter)
}
