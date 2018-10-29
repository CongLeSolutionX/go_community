// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"hash"
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

type hmacTest struct {
	hash      func() hash.Hash
	key       []byte
	in        []byte
	out       string
	size      int
	blocksize int
}

var hmacTests = []hmacTest{
	// Tests from US FIPS 198
	// https://csrc.nist.gov/publications/fips/fips198/fips-198a.pdf
	{
		sha1.New,
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
		},
		[]byte("Sample #1"),
		"4f4ca3d5d68ba7cc0a1208c9c61e9c5da0403c0a",
		sha1.Size,
		sha1.BlockSize,
	},
	{
		sha1.New,
		[]byte{
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
			0x40, 0x41, 0x42, 0x43,
		},
		[]byte("Sample #2"),
		"0922d3405faa3d194f82a45830737d5cc6c75d24",
		sha1.Size,
		sha1.BlockSize,
	},

	// Tests from https://csrc.nist.gov/groups/ST/toolkit/examples.html
	// (truncated tag tests are left out)
	{
		sha1.New,
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
		},
		[]byte("Sample message for keylen=blocklen"),
		"5fd596ee78d5553c8ff4e72d266dfd192366da29",
		sha1.Size,
		sha1.BlockSize,
	},
	{
		sha1.New,
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13,
		},
		[]byte("Sample message for keylen<blocklen"),
		"4c99ff0cb1b31bd33f8431dbaf4d17fcd356a807",
		sha1.Size,
		sha1.BlockSize,
	},
}

func TestConstantTimeHMAC(t *testing.T) {
	for i, tt := range hmacTests {
		h := newConstantTimeHMAC(tt.hash, tt.key)
		if s := h.Size(); s != tt.size {
			t.Errorf("Size: got %v, want %v", s, tt.size)
		}
		if b := h.BlockSize(); b != tt.blocksize {
			t.Errorf("BlockSize: got %v, want %v", b, tt.blocksize)
		}
		const maxExtra = 256
		padded := make([]byte, len(tt.in)+maxExtra)
		copy(padded, tt.in)
		rand.Read(padded[len(tt.in):])
		for j := range tt.in {
			n, err := h.Write(tt.in[:j])
			if n != j || err != nil {
				t.Errorf("test %d.%d: Write(%d) = %d, %v", i, j, j, n, err)
				continue
			}

			// Repeated ConstantTimeSumWithData() calls should return the same value
			for k := 0; k < maxExtra; k++ {
				sum := fmt.Sprintf("%x", h.ConstantTimeSumWithData(nil, padded[j:len(tt.in)+k], len(tt.in)-j))
				if sum != tt.out {
					t.Errorf("test %d.%d.%d: have %s want %s\n", i, j, k, sum, tt.out)
				}
			}

			h.Reset()
		}
	}
}
