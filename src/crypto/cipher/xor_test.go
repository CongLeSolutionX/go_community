// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import (
	"bytes"
	"fmt"
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
				n := min(p, q)
				for i := 0; i < n; i++ {
					d2[i] = p[i] ^ q[i]
				}
				if !bytes.Equal(d1, d2) {
					t.Error("not equal")
				}
			}
		}
	}
}

func min(a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	return n
}

func BenchmarkXORBytes(b *testing.B) {
	dst := make([]byte, 1<<15)
	data0 := make([]byte, 1<<15)
	data1 := make([]byte, 1<<15)
	for j := 1 << 3; j <= 1<<15; j <<= 4 {
		b.Run(fmt.Sprintf("%dBytes", j), func(b *testing.B) {
			s0 := data0[:j]
			s1 := data1[:j]
			b.SetBytes(int64(j))
			for i := 0; i < b.N; i++ {
				xorBytes(dst, s0, s1)
			}
		})
	}
}
