// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// The original C code is written for MUSL and can be found at
// http://git.musl-libc.org/cgit/musl/tree/src/math/fma.c.
// The Go code is a simplified version of the original C.

import (
	"math/bits"
)

const (
	p          = 53
	emin       = 1 - bias
	zeroinfnan = mask - bias - p
)

func normalNum(f float64) (Mr uint64, Er int32, Sr int32) {
	uf := Float64bits(f)
	Er = int32(uf >> shift & mask)
	Sr = int32(uf >> 63)
	if Er == 0 {
		uf = Float64bits(f * (1 << 63))
		Er = int32(uf >> shift & mask)
		if Er == 0 {
			Er = mask + 1
		} else {
			Er = Er - 63
		}
	}
	Mr = uf & fracMask
	Mr |= 1 << shift
	Mr <<= 1
	Er -= bias + p
	return
}

// Fma returns (a * b) + c, computed as if to infinite
// precision with one rounding into a float64.
func Fma(a, b, c float64) float64 {
	Ma, Ea, Sa := normalNum(a)
	Mb, Eb, Sb := normalNum(b)
	Mc, Ec, Sc := normalNum(c)

	if Ea >= zeroinfnan || Eb >= zeroinfnan || Ec > zeroinfnan {
		return a*b + c
	}
	if Ec == zeroinfnan {
		return c
	}

	MabHi, MabLo := bits.Mul64(Ma, Mb)
	Er := Ea + Eb
	ediff := Ec - Er
	var McLo, McHi uint64
	if ediff > 0 {
		if ediff < 64 {
			McLo = Mc << uint(ediff)
			McHi = Mc >> (64 - uint(ediff))
		} else {
			McHi = Mc
			Er = Ec - 64
			ediff -= 64
			if ediff < 64 {
				MabLo = MabHi<<(64-uint(ediff)) | MabLo>>uint(ediff)
				if (MabLo << (64 - uint(ediff))) != 0 {
					MabLo |= 1
				}
				MabHi = MabHi >> uint(ediff)
			} else if ediff != 0 {
				MabLo = 1
				MabHi = 0
			}
		}
	} else {
		ediff = -ediff
		if ediff == 0 {
			McLo = Mc
		} else if ediff < 64 {
			McLo = Mc >> uint(ediff)
			if (Mc << (64 - uint(ediff))) != 0 {
				McLo |= 1
			}
		} else {
			McLo = 1
		}
	}

	Sxor := (Sa ^ Sb)
	Sr := Sxor != 0
	nz := true
	if (Sxor ^ Sc) == 0 {
		MabLo += McLo
		MabHi += McHi
		if MabLo < McLo {
			MabHi++
		}
	} else {
		tlo := MabLo
		MabLo -= McLo
		MabHi = MabHi - McHi
		if tlo < MabLo {
			MabHi--
		}
		if (MabHi >> 63) != 0 {
			MabLo = -MabLo
			MabHi = -MabHi
			if MabLo != 0 {
				MabHi--
			}
			Sr = !Sr
		}
		nz = MabHi != 0
	}

	if nz {
		Er += 64
		ediff = int32(bits.LeadingZeros64(MabHi) - 1)
		MabHi = MabHi<<uint(ediff) | MabLo>>(64-uint(ediff))
		if (MabLo << uint(ediff)) != 0 {
			MabHi |= 1
		}
	} else if MabLo != 0 {
		ediff = int32(bits.LeadingZeros64(MabLo) - 1)
		if ediff < 0 {
			MabHi = MabLo>>1 | (MabLo & 1)
		} else {
			MabHi = MabLo << uint(ediff)
		}
	} else {
		return a*b + c
	}
	Er -= ediff

	Mi := int64(MabHi)
	if Sr {
		Mi = -Mi
	}
	Mr := float64(Mi)
	if Er < emin-62 {
		if Er == emin-63 {
			z := float64(1 << 63)
			if Sr {
				z = -z
			}
			if Mr == z {
				if Sr {
					return -smallestNormal
				}
				return smallestNormal
			}
			if (MabHi << p) != 0 {
				Mi = int64(MabHi>>1 | (MabHi & 1) | 1<<62)
				if Sr {
					Mi = -Mi
				}
				Mr = float64(Mi)
				Mr = 2*Mr - z
			}
		} else {
			Mui := MabHi >> 10
			if (MabHi << shift) != 0 {
				Mui |= 1
			}
			Mi = int64(Mui << 10)
			if Sr {
				Mi = -Mi
			}
			Mr = float64(Mi)
		}
	}
	return Ldexp(Mr, int(Er))
}
