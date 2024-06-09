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

func TestCondenseAsm(t *testing.T) {
	for sizeClass, bytes := range class_to_size {
		if sizeClass == 0 {
			// Large objects.
			continue
		}
		f := int(bytes) / 8
		t.Run(fmt.Sprintf("f=%d", f), func(t *testing.T) {
			if f == 1 || f >= 64 {
				t.Skipf("unimplemented: bytes=%d", bytes)
			}

			// Test each single bit.
			nPages := int(class_to_allocnpages[sizeClass])
			nBits := nPages * 1024
			packed := make([]uint64, (nBits+63)/64)
			for bit := 0; bit < nBits/f*f; bit++ {
				var unpacked [8]uint64
				var want [8]uint64
				clear(packed)
				packed[bit/64] = 1 << (bit % 64)
				wantBit := bit / f
				want[wantBit/64] = 1 << (wantBit % 64)

				condenseAsm(sizeClass, &packed[0], &unpacked)

				if unpacked != want {
					t.Fatalf("unexpected output for input bit %d, output bit %d:\n%s", bit, wantBit, fmtBits(unpacked[:]))
				}
			}
		})
	}
}
