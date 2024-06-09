// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"fmt"
	"slices"
	"testing"
)

func TestCondenser(t *testing.T) {
	// Test all widths up to a width that includes complete interior words.
	const limit = 1024
	for w := 1; w < 130; w++ {
		t.Run(fmt.Sprintf("w=%d", w), func(t *testing.T) {
			c := newCondenser(heap.Words(w), limit)
			t.Logf("%+v", c)

			src := make([]uint64, (limit+63)/64)
			dstBits := (limit + w - 1) / w
			dst := make([]uint64, (dstBits+63)/64)
			wantDst := make([]uint64, len(dst))

			for sBit := 0; sBit < limit; sBit++ {
				src[sBit/64] = 1 << (sBit % 64)
				oBit := sBit / w
				wantDst[oBit/64] = 1 << (oBit % 64)

				c.do(src, dst)
				if !slices.Equal(dst, wantDst) {
					t.Errorf("source bit %d:\nwant %#016x\ngot  %#016x", sBit, wantDst, dst)
				}

				src[sBit/64] = 0
				wantDst[oBit/64] = 0
			}
		})
	}
}
