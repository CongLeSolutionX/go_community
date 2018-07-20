// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import (
	"bytes"
	"testing"

	"internal/cpu"
)

func TestXOR(t *testing.T) {
	for alignP := 0; alignP < 2; alignP++ {
		for alignQ := 0; alignQ < 2; alignQ++ {
			for alignD := 0; alignD < 2; alignD++ {
				p := make([]byte, 1024)[alignP:]
				q := make([]byte, 1024)[alignQ:]
				d1 := make([]byte, 1024+alignD)[alignD:]
				d2 := make([]byte, 1024+alignD)[alignD:]
				xorBytes(d1, p, q)
				safeXORBytes(d2, p, q)
				if !bytes.Equal(d1, d2) {
					t.Error("not equal")
				}
			}
		}
	}
}

func TestXORSSE2(t *testing.T) {
	if cpu.X86.HasSSE2 {
		act := make([]byte, 128+64+32+16+8+7)
		exp := make([]byte, 128+64+32+16+8+7)
		for i := 1; i < len(act); i++ {
			a := make([]byte, i)
			b := make([]byte, i+1)
			xorSSE2(act, a, b, i)
			safeXORBytes(exp, a, b)
			if !bytes.Equal(exp, act) {
				t.Error("not equal")
			}
		}
	}
}

func BenchmarkTestXORBytes1K(b *testing.B) {
	dst := make([]byte, 1024)
	s0 := make([]byte, 1024)
	s1 := make([]byte, 1024)
	b.SetBytes(1024)
	for i := 0; i < b.N; i++ {
		xorBytes(dst, s0, s1)
	}
}
