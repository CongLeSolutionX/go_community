// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rand

import "math/bits"

// https://numpy.org/devdocs/reference/random/upgrading-pcg64.html
// https://github.com/imneme/pcg-cpp/commit/871d0494ee9c9a7b7c43f753e3d8ca47c26f8005

type PCG struct {
	hi uint64
	lo uint64
}

func NewPCG(seed1, seed2 uint64) *PCG {
	return &PCG{seed1, seed2}
}

func (p *PCG) next() (hi, lo uint64) {
	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L161
	const (
		mulHi = 2549297995355413924
		mulLo = 4865540595714422341
		incHi = 6364136223846793005
		incLo = 1442695040888963407
	)

	// state = state * mul + inc
	hi, lo = bits.Mul64(p.lo, mulLo)
	hi += p.hi*mulLo + p.lo*mulHi
	lo, c := bits.Add64(lo, incLo, 0)
	hi, _ = bits.Add64(hi, incHi, c)
	p.lo = lo
	p.hi = hi
	return hi, lo
}

func (p *PCG) xslrr() uint64 {
	// XSL-RR
	hi, lo := p.next()
	return bits.RotateLeft64(lo^hi, -int(hi>>58))
}

func (p *PCG) Uint64() uint64 {
	hi, lo := p.next()

	// DXSM "double xorshift multiply"
	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L1015

	// https://github.com/imneme/pcg-cpp/blob/428802d1a5/include/pcg_random.hpp#L176
	const cheapMul = 0xda942042e4dd58b5
	hi ^= hi >> 32
	hi *= cheapMul
	hi ^= hi >> 48
	hi *= (lo | 1)
	return hi
}
