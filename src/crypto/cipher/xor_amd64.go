// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import (
	"internal/cpu"
	"runtime"
)

// xorBytes xors the bytes in a and b. The destination should have enough
// space, otherwise xorBytes will panic. Returns the number of bytes xor'd.
func xorBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	switch {
	case runtime.GOARCH == "amd64" && cpu.X86.HasSSE2:
		xorBytesSSE2(&dst[0], &a[0], &b[0], n)
	default:
		fastXORBytes(dst, a, b, n)
	}
	return n
}

//go:noescape
func xorBytesSSE2(dst, a, b *byte, n int)
