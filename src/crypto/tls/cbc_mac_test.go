// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestRotateSecretOffset(t *testing.T) {
	// SHA-1 uses 20-byte MACs.
	b := make([]byte, 20)
	rand.Read(b)
	b2 := make([]byte, 20)
	copy(b2, b)

	for i := range b {
		for j := range b2 {
			b2[j] = b[(j+i)%len(b)]
		}
		rotateSecretOffset(b, i)

		if !bytes.Equal(b, b2) {
			t.Fatalf("Rotating by %d produced %x, wanted %x", i, b, b2)
		}
	}
}

func TestCopySecretSlice(t *testing.T) {
	// SHA-1 uses 20-byte MACs.
	const l = 20
	b := make([]byte, l+256)
	rand.Read(b)
	for i := 0; i < len(b)-l; i++ {
		ret := copySecretSlice(b, i, l)
		if !bytes.Equal(ret, b[i:i+l]) {
			t.Errorf("Copying b[%d:%d] produced %x, wanted %x", i, i+l, ret, b[i:i+l])
		}
	}
}
