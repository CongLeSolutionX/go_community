// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64 || arm64

package nistec

import (
	"encoding/binary"
	"testing"
)

func TestP256PrecomputedTable(t *testing.T) {
	basePoint := NewP256Generator()
	t1 := NewP256Point()
	t2 := NewP256Generator()

	zInv := new(p256Element)
	zInvSq := new(p256Element)
	for j := 0; j < 32; j++ {
		t1.Set(t2)
		for i := 0; i < 43; i++ {
			// The window size is 6 so we need to double 6 times.
			if i != 0 {
				for k := 0; k < 6; k++ {
					p256PointDoubleAsm(t1, t1)
				}
			}
			// Convert the point to affine form. (Its values are
			// still in Montgomery form however.)
			p256Inverse(zInv, &t1.z)
			p256Sqr(zInvSq, zInv, 1)
			p256Mul(zInv, zInv, zInvSq)

			p256Mul(&t1.x, &t1.x, zInvSq)
			p256Mul(&t1.y, &t1.y, zInv)

			t1.z = basePoint.z

			buf := make([]byte, 0, 8*8)
			for _, u := range t1.x {
				buf = binary.LittleEndian.AppendUint64(buf, u)
			}
			for _, u := range t1.y {
				buf = binary.LittleEndian.AppendUint64(buf, u)
			}
			start := i*32*8*8 + j*8*8
			if got, want := p256Precomputed[start:start+64], string(buf); got != want {
				t.Fatalf("Unexpected table entry at [%d][%d:%d]: got %v, want %v", i, j*8, (j*8)+8, got, want)
			}
		}
		if j == 0 {
			p256PointDoubleAsm(t2, basePoint)
		} else {
			p256PointAddAsm(t2, t2, basePoint)
		}
	}

}
