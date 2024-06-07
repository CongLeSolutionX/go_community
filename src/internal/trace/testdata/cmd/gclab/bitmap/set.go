// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitmap

import (
	"iter"
	"math/bits"
)

type Set[K ~uint64] struct {
	bits []uint64
}

func NewSet[K ~uint64](nBits K) Set[K] {
	return Set[K]{make([]uint64, (nBits+63)/64)}
}

func (b Set[K]) Has(i K) bool {
	return (b.bits[i/64] & (1 << (i % 64))) != 0
}

func (b Set[K]) Add(i K) {
	b.bits[i/64] |= 1 << (i % 64)
}

func (b Set[K]) Remove(i K) {
	b.bits[i/64] &^= 1 << (i % 64)
}

func (b Set[K]) Len() K {
	var sum K
	for _, w := range b.bits {
		sum += K(bits.OnesCount64(w))
	}
	return sum
}

func (b Set[K]) All() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i, val := range b.bits {
			for range bits.OnesCount64(val) {
				bitI := bits.TrailingZeros64(val)
				yield(K(i*64 + bitI))
				val &^= 1 << bitI
			}
		}
	}
}

func (b Set[K]) Range(start, end K) iter.Seq[K] {
	if start%64 != 0 || end%64 != 0 {
		// We could of course support this, but we don't need it and it
		// complicates a hot path.
		panic("start and len must be multiples of 64")
	}

	return func(yield func(K) bool) {
		for i, val := range b.bits[start/64 : end/64] {
			base := start + K(i*64)
			for range bits.OnesCount64(val) {
				bitI := bits.TrailingZeros64(val)
				yield(base + K(bitI))
				val &^= 1 << bitI
			}
		}
	}
}
