// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package rand

import "testing"

func BenchmarkPCG_DXSM(b *testing.B) {
	var p PCG
	var t uint64
	for n := b.N; n > 0; n-- {
		t += p.Uint64()
	}
	Sink = t
}

func BenchmarkPCG_XSLRR(b *testing.B) {
	var p PCG
	var t uint64
	for n := b.N; n > 0; n-- {
		t += p.xslrr()
	}
	Sink = t
}
