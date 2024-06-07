// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitmap

import (
	"slices"
)

type DynSet[K ~uint64] struct {
	bits []uint64
}

func (b DynSet[K]) Has(i K) bool {
	if i/64 < K(len(b.bits)) {
		return (b.bits[i/64] & (1 << (i % 64))) != 0
	}
	return false
}

func (b *DynSet[K]) Add(i K) {
	if i/64 >= K(len(b.bits)) {
		l := int(i/64 + 1)
		bits := slices.Grow(b.bits, l-len(b.bits))
		b.bits = bits[:cap(bits)]
	}
	b.bits[i/64] |= 1 << (i % 64)
}

func (b *DynSet[K]) Remove(i K) {
	if int(i/64) < len(b.bits) {
		b.bits[i/64] &^= 1 << (i % 64)
	}
}

func (b DynSet[K]) Set() Set[K] {
	return Set[K]{b.bits}
}

func (b *DynSet[K]) Drop() {
	b.bits = nil
}
