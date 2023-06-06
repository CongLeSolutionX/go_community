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

func TestPCGMarshal(t *testing.T) {
	var p PCG
	const (
		seed1 = 0x123456789abcdef0
		seed2 = 0xfedcba9876543210
		want  = "pcg:\x12\x34\x56\x78\x9a\xbc\xde\xf0\xfe\xdc\xba\x98\x76\x54\x32\x10"
	)
	data, err := p.MarshalBinary()
	if string(data) != want || err != nil {
		t.Errorf("MarshalBinary() = %q, %v, want %q, nil", data, err, want)
	}

	q := PCG{}
	if err := q.UnmarshalBinary([]byte(want)); err != nil {
		t.Fatalf("UnmarshalBinary(): %v", err)
	}
	if q.hi != seed1 || q.lo != seed2 {
		t.Fatalf("after UnmarshalBinary, hi:lo = %#x:%#x, want %#x:%#x", p.hi, p.lo, seed1, seed2)
	}
	if q != *p {
		t.Fatalf("after round trip, q = %#x, but p = %#x", q, *p)
	}

	qu := q.Uint64()
	pu := p.Uint64()
	if qu != pu {
		t.Errorf("after round trip, q.Uint64() = %#x, but p.Uint64() = %#x", qu, pu)
	}
}