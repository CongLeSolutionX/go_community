// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package obj

import (
	"testing"
)

func TestSymgrow(t *testing.T) {
	var tests = []struct {
		sym      *LSym
		size     int64
		len, cap int // expected len and cap after Symgrow
	}{
		{sym: &LSym{}, size: 0, len: 0, cap: 0},
		{sym: &LSym{P: make([]byte, 4)}, size: 0, len: 4, cap: 4},
		{sym: &LSym{P: make([]byte, 4)}, size: 4, len: 4, cap: 4},
		{sym: &LSym{P: make([]byte, 4)}, size: 8, len: 8, cap: 8},
		{sym: &LSym{P: make([]byte, 4)}, size: 12, len: 12, cap: 16},
		{sym: &LSym{P: make([]byte, 4, 8)}, size: 4, len: 4, cap: 8},
		{sym: &LSym{P: make([]byte, 4, 8)}, size: 5, len: 5, cap: 8},
	}

	for _, tt := range tests {
		got := *tt.sym
		Symgrow(&got, tt.size)
		if len(got.P) != tt.len || cap(got.P) != tt.cap {
			t.Errorf("Symgrow(%v): want: len=%v, cap=%v, got: len=%v, cap=%v", tt.size, tt.len, tt.cap, len(got.P), cap(got.P))
		}
	}
}

func BenchmarkSymgrowSmall(b *testing.B) {
	var buf [8]byte
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		sym := LSym{P: buf[0:4:8]}
		Symgrow(&sym, 12)
	}
}

func BenchmarkSymgrowMed(b *testing.B) {
	var buf [100]byte
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		sym := LSym{P: buf[0:50:100]}
		Symgrow(&sym, 120)
	}
}

func BenchmarkSymgrowLarge(b *testing.B) {
	var buf [3000]byte
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		sym := LSym{P: buf[0:1000:3000]}
		Symgrow(&sym, 5000)
	}
}
