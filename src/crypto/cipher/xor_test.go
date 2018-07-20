// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import (
	"bytes"
	"testing"
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

func BenchmarkXORBytes(b *testing.B) {
	benchmarks := []struct {
		name string
		n    int
	}{
		{"8B", 8},
		{"16B", 16},
		{"64B", 64},
		{"256B", 256},
		{"1K", 1024},
		{"4K", 4096},
	}
	dst := make([]byte, 4096)
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			s0 := make([]byte, bm.n)
			s1 := make([]byte, bm.n)
			b.SetBytes(int64(bm.n))
			for i := 0; i < b.N; i++ {
				xorBytes(dst, s0, s1)
			}
		})
	}
}
