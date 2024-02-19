// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestCompactors(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 8))
	n := 1000000
	if testing.Short() {
		n = 1000
	}

	for class, compactor := range runtime.Compactors {
		if compactor == nil {
			continue
		}
		size := uintptr(runtime.Class2Size[class])
		nq := size / runtime.ObjectQuantum

		for i := 0; i < n; i++ {
			// Pick a random argument. Choose a bit set with probability
			// 1/nq, which gives a result bit set with prob ~50%.
			x := uintptr(0)
			for j := 0; j < 64; j++ {
				if r.IntN(int(nq+1)) == 0 {
					x |= 1 << j
				}
			}
			got := compactor(x)
			want := uintptr(0)
			for j := 0; j < 64; j++ {
				if x>>j&1 != 0 {
					want |= 1 << (uintptr(j) / nq)
				}
			}
			if got != want {
				t.Errorf("compactor(%d)(%x) = %x, want %x", size, x, got, want)
			}
		}
	}
}
