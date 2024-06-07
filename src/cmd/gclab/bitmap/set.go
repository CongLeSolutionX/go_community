// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitmap

import (
	"iter"
	"log"
	"math/bits"
)

type Set[K ~uint64] struct {
	bits []uint64
}

func NewSet[K ~uint64](nBits K) Set[K] {
	return Set[K]{make([]uint64, (nBits+63)/64)}
}

func FromWords[K ~uint64](words []uint64) Set[K] {
	return Set[K]{words}
}

func (b Set[K]) Has(i K) bool {
	return i/64 < K(len(b.bits)) && (b.bits[i/64]&(1<<(i%64))) != 0
}

func (b Set[K]) Add(i K) {
	b.bits[i/64] |= 1 << (i % 64)
}

func (b Set[K]) Remove(i K) {
	b.bits[i/64] &^= 1 << (i % 64)
}

func (b Set[K]) Len() int {
	var sum int
	for _, w := range b.bits {
		sum += bits.OnesCount64(w)
	}
	return sum
}

func (b Set[K]) LenRange(start, end K) K {
	var sum K
	if start%64 != 0 {
		word := b.bits[start/64] >> (start % 64)
		sum += K(bits.OnesCount64(word))
	}
	if end%64 != 0 {
		word := b.bits[end/64] << (64 - end%64)
		sum += K(bits.OnesCount64(word))
	}

	for _, w := range b.bits[(start+63)/64 : end/64] {
		sum += K(bits.OnesCount64(w))
	}
	return sum
}

func (b Set[K]) Words(start K, len uint) []uint64 {
	if uint64(start)%64 != 0 {
		log.Fatalf("start offset %v not a multiple of 64", start)
	}
	return b.bits[start/64:][:(len+63)/64]
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
