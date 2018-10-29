// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestRotateSecretOffset(t *testing.T) {
	// This function is used with SHA-1 and SHA-256 which are 20 and 30
	// bytes, respectively.
	for _, l := range []int{20, 32} {
		b := make([]byte, l)
		rand.Read(b)
		b2 := make([]byte, l)
		copy(b2, b)

		for i := range b {
			for j := range b2 {
				b2[j] = b[(j+i)%len(b)]
			}
			rotateSecretOffset(b, i)

			if !bytes.Equal(b, b2) {
				t.Fatalf("Rotating %d bytes by %d produced %x, wanted %x", l, i, b, b2)
			}
		}
	}
}

func TestCopySecretSlice(t *testing.T) {
	in := make([]byte, 256+32)
	rand.Read(in)

	// This function is used with SHA-1 and SHA-256 which are 20 and 30
	// bytes, respectively.
	for _, l := range []int{20, 32} {
		// CBC-mode ciphers will call copySecretSlice on multiples of
		// 16 up to 256+l (small padded plaintext unconstrained by the
		// maximum padding) and 256+l (the maximum padding is 256).
		// Test a range of possible values here.
		//
		// We also use this codepath for RC4, which is unconstrained by
		// block sizes, so additionally test l and l+1 to capture
		// boundary cases.
		for _, tot := range []int{l, l + 1, 32, 48, 64, 96, 128, 192, 256, 256 + l} {
			b := in[:tot]
			for i := 0; i < tot-l; i++ {
				ret := copySecretSlice(b, i, l)
				if !bytes.Equal(ret, b[i:i+l]) {
					t.Errorf("Copying b[%d:%d] produced %x, wanted %x (len(b) = %d)", i, i+l, ret, b[i:i+l], tot)
				}
			}
		}
	}
}
