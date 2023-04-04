// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Modifications to support ifma
// Copyright (C) 2023 Intel Corporation
// SPDX-License-Identifier: BSD-3-Clause
//
// Based on OpenSSL
// Copyright 1995-2022 The OpenSSL Project Authors. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.openssl.org/source/license.html
//

//
// References
//
// [1] Gueron, S. Efficient software implementations of modular exponentiation.
//     DOI: 10.1007/s13389-012-0031-5
//

package rsa

import (
	"math/big"
	"unsafe"
)

//go:noescape
func cvt_16x64_to_20x52(_a52, _b52, _m52 *big.Word, a0, a1, b0, b1, m0, m1 []big.Word)
func cvt_20x52_norm_to_16x64(x64 []big.Word, _x52 *big.Word)
func amm_52x20_x1_ifma256(_res, _a, _b, _m *big.Word, _k0 uint64)
func amm_52x20_x2_ifma256(_out, _a, _b, _m *big.Word, _k0 *uint64)
func bn_reduce_once_in_place_16(_z, _a, _b *big.Word) big.Word
func bn_mul_mont_16(_r, _a, _b, _n *big.Word, _n0 uint64)
func bn_mul_mont(_r, _a, _b, _n *big.Word, _n0 *uint64, _num uint64)
func bn_from_mont(_bn, _r, _n *big.Word, _n0 big.Word, _num uint64)
func extract_mul_2x20_win5(_res, _table *big.Word, i1, i2 big.Word)

const modulus_bitsize = 1024
const exp_digits = 16
const exp_win_size = 5
const exp_win_mask = 31
const rem = 1024 % exp_win_size
const table_idx_mask = big.Word(exp_win_mask)

// montgomery BN representation
type mont struct {
	ri int       // number of bits in R
	RR *big.Int  // used to convert to montgomery form, zero-padded if necessary
	N  *big.Int  // modulus
	Ni *big.Int  // R*(1/R mod N) - N*Ni = 1 (Ni is only stored for bignum algorithm)
	N0 [2]uint64 // least significant word(s) of Ni
}

// ////////////////////////////////////////////////////////////////////////////
//
// newMont()
//
// initialize montgomery struct
func newMont() *mont {
	m := new(mont)
	m.ri = 0
	m.RR = new(big.Int)
	m.N = new(big.Int)
	m.Ni = new(big.Int)
	return m
}

// ////////////////////////////////////////////////////////////////////////////
//
// setMont()
//
// initialize montgomery val
// assumes m.RR=0 on input, otherwise set explicitly
func setMont(mod *big.Int, m *mont, numBits int) int {
	var bnZero = new(big.Int).SetInt64(0)
	var bnOne = new(big.Int).SetInt64(1)
	var bnRi = new(big.Int).SetInt64(0)
	var bnTmod = new(big.Int).SetBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if 0 == mod.Cmp(bnZero) {
		return 0
	}
	m.N.Set(mod)
	m.ri = len(mod.Bits()) * 64
	m.RR.SetBit(m.RR, 64, 1)
	bnTmod.SetBits(mod.Bits()[:1])
	if 0 != bnTmod.Cmp(bnOne) {
		bnRi.ModInverse(m.RR, bnTmod) // Ri = R^-1 mod N
	}
	bnRi.Lsh(bnRi, 64)
	if 0 != bnRi.Cmp(bnZero) {
		bnRi.Sub(bnRi, bnOne)
	} else {
		panic("decryptAndCheckIfma2048.setMont() internal error")
	}
	bnRi.Div(bnRi, bnTmod)
	m.N0[0] = uint64(bnRi.Bits()[0])
	m.N0[1] = 0

	// setup RR for conversions
	m.RR.SetUint64(0)
	m.RR.SetBit(m.RR, 2*m.ri, 1)
	m.RR.Mod(m.RR, m.N)

	// high-order zeros in the bit count are truncated
	if (m.RR.BitLen()+63)/64 != (m.N.BitLen()+63)/64 {
		panic("decryptAndCheckIfma2048.setMont() internal error")
	}
	return 1
}

// ////////////////////////////////////////////////////////////////////////////
//
// bn_mod_sub_fixed_top()
//
// Montgomery BN modular subtraction for both a and b non-negative,
// a is less than m, while b is of same bit width as m. Implemented
// as subtraction followed by two conditional additions.
//
//	0 <= a < m
//	0 <= b < 2^w < 2*m
//
// after subtraction
//
//	-2*m < r = a - b < m
//
// thus up to two conditional additions required to make |r| positive
//
// Derived from the OpenSSL internal crypto big number function
// bn_mod_sub_fixed_top()
// * Copyright 1998-2021 The OpenSSL Project Authors. All Rights Reserved.
// *
// * Licensed under the Apache License 2.0 (the "License").  You may not use
// * this file except in compliance with the License.  You can obtain a copy
// * in the file LICENSE in the source distribution or at
// * https://www.openssl.org/source/license.html
func bn_mod_sub_fixed_top(a, b, m []big.Word) []big.Word {
	var i, ai, bi, borrow, carry, ta, tb, mask, mtop, btop, atop big.Word
	var r [16]big.Word
	mtop = big.Word(len(m))
	atop = big.Word(len(a))
	btop = big.Word(len(b))
	for i, ai, bi, borrow = 0, 0, 0, 0; i < mtop; {
		mask = big.Word(big.Word(0) - ((i - atop) >> (8*8 - 1)))
		ta = a[ai] & mask
		mask = big.Word(big.Word(0) - ((i - btop) >> (8*8 - 1)))
		tb = b[bi] & mask
		r[i] = ta - tb - borrow
		if ta != tb {
			if ta < tb {
				borrow = 1
			} else {
				borrow = 0
			}
		}
		i++
		ai += (i - 16) >> (8*8 - 1)
		bi += (i - 16) >> (8*8 - 1)
	}
	for i, mask, carry = 0, 0-borrow, 0; i < mtop; i++ {
		ta = ((m[i] & mask) + carry)
		if ta < carry {
			carry = 1
		} else {
			carry = 0
		}
		r[i] = (r[i] + ta)
		if r[i] < ta {
			carry++
		}
	}
	borrow -= carry
	for i, mask, carry = 0, 0-borrow, 0; i < mtop; i++ {
		ta = ((m[i] & mask) + carry)
		if ta < carry {
			carry = 1
		} else {
			carry = 0
		}
		r[i] = (r[i] + ta)
		if r[i] < ta {
			carry++
		}
	}
	return r[:]
}

// ////////////////////////////////////////////////////////////////////////////
//
// bn_mod_exp( a, p, m )
//
// big number modular exponentiation; bn return val = a^p mod m;
// makes use of BN_window_bits_for_exponent_size -- sliding window mod_exp functions
// ror window size 'w' (w >= 2) and a random 'b' bits exponent,
// the number of multiplications is a constant plus on average
//
//	2^(w-1) + (b-w)/(w+1);
//
// here  2^(w-1)  is for precomputing the table (we actually need
// entries only for windows that have the lowest bit set), and
// (b-w)/(w+1)  is an approximation for the expected number of
// w-bit windows, not counting the first one.
// Thus we should use
//
//	w >= 6  if        b > 671
//	w = 5  if  671 > b > 239
//	w = 4  if  239 > b >  79
//	w = 3  if   79 > b >  23
//	w <= 2  if   23 > b
//
// (with draws in between).  Very small exponents are often selected
// with low Hamming weight, so we use  w = 1  for b <= 23.
//
// Derived from openSSL crypto rsa function bn_mod_exp()
// * Copyright 1998-2021 The OpenSSL Project Authors. All Rights Reserved.
// *
// * Licensed under the Apache License 2.0 (the "License").  You may not use
// * this file except in compliance with the License.  You can obtain a copy
// * in the file LICENSE in the source distribution or at
// * https://www.openssl.org/source/license.html
func bn_mod_exp(a, p, m *big.Int) *big.Int {
	bits := p.BitLen()
	if bits == 0 {
		panic("decryptAndCheckIfma2048.bn_mod_exp() zero length p")
	}
	mont := newMont()
	setMont(m, mont, 1024)

	// init temp full-length bn vectors for a, m
	var A [32]big.Word
	var RR [32]big.Word
	var N [32]big.Word
	for i := range a.Bits() {
		A[i] = a.Bits()[i]
	}
	for i := range mont.RR.Bits() {
		RR[i] = mont.RR.Bits()[i]
	}
	for i := range mont.N.Bits() {
		N[i] = mont.N.Bits()[i]
	}
	var val [64][32]big.Word
	var dval [32]big.Word
	num := uint64(len(mont.N.Bits()))
	mN := &(N[0])
	mN0 := &(mont.N0[0])
	d := &(dval[0])
	bn_mul_mont(&(val[0][0]), &(A[0]), &(RR[0]), mN, mN0, num)

	// init exp window
	var w int
	switch {
	case bits > 671:
		w = 6
	case bits > 239:
		w = 5
	case bits > 79:
		w = 4
	case bits > 23:
		w = 3
	default:
		w = 1
	}
	if w > 1 {
		bn_mul_mont(d, &(val[0][0]), &(val[0][0]), mN, mN0, num)
		j := 1 << (w - 1)
		for i := 1; i < j; i++ {
			bn_mul_mont(&(val[i][0]), &(val[i-1][0]), d, mN, mN0, num)
		}
	}

	// avoid multiplication etc
	// when there is only the value '1' in the buffer
	start := 1
	wvalue := 0        // win val
	wstart := bits - 1 // win top bit
	wend := 0          // win bottom bit

	// by Shay Gueron's suggestion
	j := len(m.Bits())
	mw := m.Bits()

	// rw.[]Words must be 2x length of m to hold carry words during conversion
	// from mont currently largest expected m is 32
	var rw [64]big.Word
	var rrw [64]big.Word
	var bnOne [32]big.Word
	bnOne[0] = 1
	r := &(rw[0])
	rr := &(rrw[0])
	if mw[j-1]&0x8000000000000000 != 0 {
		/* 2^(top*BN_BITS2) - m */
		rw[0] = 0 - mw[0]
		for i := 1; i < j; i++ {
			rw[i] = ^mw[i]
		}
	} else {
		bn_mul_mont(r, &(bnOne[0]), &(RR[0]), mN, mN0, num)
	}
	dbg := 0
	for true {
		dbg++
		if p.Bit(wstart) == 0 {
			if start == 0 {
				bn_mul_mont(r, r, r, mN, mN0, num)
			}
			if wstart == 0 {
				break
			}
			wstart--
			continue
		}

		// wstart is on a 'set' bit; next determine win size by scanning forward
		// until the last set bit before window end
		wvalue = 1
		wend = 0
		for i := 1; i < w; i++ {
			if wstart-i < 0 {
				break
			}
			if p.Bit(wstart-i) == 1 {
				wvalue <<= (i - wend)
				wvalue |= 1
				wend = i
			}
		}

		// wend is the size of the current window
		j = wend + 1

		// add the 'bytes above'
		if start == 0 {
			for i := 0; i < j; i++ {
				bn_mul_mont(r, r, r, mN, mN0, num)
			}
		}

		// wvalue will be an odd number < 2^window
		bn_mul_mont(r, r, &(val[wvalue>>1][0]), mN, mN0, num)

		// move 'window' down further
		wstart -= wend + 1
		wvalue = 0
		start = 0
		if wstart < 0 {
			break
		}
	}
	bn_from_mont(rr, r, mN, big.Word(mont.N0[0]), num)
	ret := new(big.Int)
	ret.SetBits(rrw[:(num + 1)])
	return ret
}

// ////////////////////////////////////////////////////////////////////////////
//
// getSliceWordAlign64()
//
// get 64-bit aligned slice
func getSliceWordAlign64(n int) []big.Word {
	const align = 64
	v := make([]big.Word, n+align/8)
	for i := range v {
		if uintptr(unsafe.Pointer(&v[i]))%align == 0 {
			return v[i:]
		}
	}
	panic("could get 64-byte aligned []big.Word")
}

// ////////////////////////////////////////////////////////////////////////////
//
// getAlginedBuffers()
//
// get aligned buffers for ifma ops in decryptAndCheck()
func getAlignedBuffers() (Base52, M52, RR52, Coeff, Table, X, Y, Expz []big.Word) {
	buf := getSliceWordAlign64(1534)
	Base52 = buf[:40]
	M52 = buf[40:80]
	RR52 = buf[80:120]
	Coeff = buf[120:140]
	Table = buf[140:1420]
	X = buf[1420:1460]
	Y = buf[1460:1500]
	Expz = buf[1500:1534]
	return Base52, M52, RR52, Coeff, Table, X, Y, Expz
}

// ////////////////////////////////////////////////////////////////////////////
//
// exponentiation()
//
// w-ary modular exponentiation using prime moduli of
// the same bit size using Almost Montgomery Multiplication, with
// the parameter w (window size) = 5;
//
// Derived from OpenSSL crypto big number function RSAZ_mod_exp_x2_ifma256()
// * Copyright 1998-2021 The OpenSSL Project Authors. All Rights Reserved.
// *
// * Licensed under the Apache License 2.0 (the "License").  You may not use
// * this file except in compliance with the License.  You can obtain a copy
// * in the file LICENSE in the source distribution or at
// * https://www.openssl.org/source/license.html
func exponentiation(x, y, m, table *big.Word, k0 *uint64, expz []big.Word) {
	exp_bit_no := modulus_bitsize - rem
	exp_chunk_no := exp_bit_no / 64
	exp_chunk_shift := exp_bit_no % 64

	// process 1-st exp window - just init result
	// the function operates with fixed moduli sizes divisible by 64,
	// thus table index here is always in supported range [0, EXP_WIN_SIZE).
	idx0 := big.Word(expz[exp_chunk_no]) >> exp_chunk_shift
	idx1 := big.Word(expz[exp_chunk_no+(exp_digits+1)]) >> exp_chunk_shift
	extract_mul_2x20_win5(y, table, idx0, idx1)

	// process other exp windows
	for exp_bit_no -= exp_win_size; exp_bit_no >= 0; exp_bit_no -= exp_win_size {

		// extract pre-computed multiplier from the table
		exp_chunk_no = exp_bit_no / 64
		exp_chunk_shift = exp_bit_no % 64
		idx0 = expz[exp_chunk_no]
		T := expz[exp_chunk_no+1]
		idx0 >>= exp_chunk_shift

		// get additional bits from then next quadword
		// when 64-bit boundaries are crossed
		if exp_chunk_shift > 64-exp_win_size {
			T <<= (64 - exp_chunk_shift)
			idx0 ^= T
		}
		idx0 &= table_idx_mask
		idx1 = expz[exp_chunk_no+1*(exp_digits+1)]
		T = expz[exp_chunk_no+1+1*(exp_digits+1)]
		idx1 >>= exp_chunk_shift

		// get additional bits from then next quadword
		// when 64-bit boundaries are crossed.
		if exp_chunk_shift > 64-exp_win_size {
			T <<= (64 - exp_chunk_shift)
			idx1 ^= T
		}
		idx1 &= table_idx_mask
		extract_mul_2x20_win5(x, table, idx0, idx1)

		// series of squaring
		amm_52x20_x2_ifma256(y, y, y, m, k0)
		amm_52x20_x2_ifma256(y, y, y, m, k0)
		amm_52x20_x2_ifma256(y, y, y, m, k0)
		amm_52x20_x2_ifma256(y, y, y, m, k0)
		amm_52x20_x2_ifma256(y, y, y, m, k0)
		amm_52x20_x2_ifma256(y, y, x, m, k0)
	}
}

// ////////////////////////////////////////////////////////////////////////////
//
// decryptIfma2048()
//
// implements crypto/rsa function decryptAndCheck(),
// optmized for 2048-bit keys using avx-512 and ifma instructions
//
// Derived from OpenSSL crypto rsa function rsa_ossl_mod_exp()
// * Copyright 1998-2021 The OpenSSL Project Authors. All Rights Reserved.
// *
// * Licensed under the Apache License 2.0 (the "License").  You may not use
// * this file except in compliance with the License.  You can obtain a copy
// * in the file LICENSE in the source distribution or at
// * https://www.openssl.org/source/license.html
func decryptIfma2048(priv *PrivateKey, ciphertext []byte, check bool) ([]byte, error) {

	const RSA_MAX_NUM_PRIMES = 5

	// switch on number of primes, either 2 or >2
	c := new(big.Int).SetBytes(ciphertext)
	switch len(priv.Primes) {
	case 2:
		R0 := new(big.Int)
		R1 := new(big.Int).Mod(c, priv.Primes[0])
		M1 := new(big.Int).Mod(c, priv.Primes[1])
		Mp := newMont()
		Mq := newMont()
		P := priv.Primes[0]
		Q := priv.Primes[1]
		setMont(priv.Primes[0], Mp, 1024)
		setMont(priv.Primes[1], Mq, 1024)

		// align intermediate result buffers on 64-byte boundaries
		Base52, M52, RR52, Coeff, Table, X, Y, Expz := getAlignedBuffers()

		// 2048-bit radix-52 bn buffer pointers
		// each buffer contains 2 consecutive big numbers represented in Go nat format
		// i.e., for 2048-bits a little-endian vector of 20 x radix-52-bit uwords (LS word first) --> "nat_r52_LE_2048"
		base_0 := &(Base52[0]) // base_0; Base52 holds 2 consecutive bn base_0, base_1, each nat_r52_LE_2048
		rr_0 := &(RR52[0])     // rr_0 nat_r52_2048
		rr_1 := &(RR52[20])    // rr_1 	  "  "
		m_0 := &(M52[0])       // m_0  	  "  "
		m_1 := &(M52[20])      // m52_1 	"  "
		coeff := &(Coeff[0])
		x := &(X[0])
		y := &(Y[0])

		// convert base_i, m_i, rr_i, from radix 64 to radix 52
		cvt_16x64_to_20x52(base_0, m_0, rr_0, M1.Bits(), R1.Bits(), Q.Bits(), P.Bits(), Mq.RR.Bits(), Mp.RR.Bits())

		// Compute target domain Montgomery converters RR' for each modulus
		// based on precomputed original domain's RR.
		//   RR -> RR' transformation steps:
		//    (1) coeff = 2^k
		//    (2) t = AMM(RR,RR) = RR^2 / R' mod m
		//    (3) RR' = AMM(t, coeff) = RR^2 * 2^k / R'^2 mod m
		//   where
		//    k = 4 * (52 * digits52 - modlen)
		//    R  = 2^(64 * ceil(modlen/64)) mod m
		//    RR = R^2 mod m
		//    R' = 2^(52 * ceil(modlen/52)) mod m
		//    EX/ modlen = 1024: k = 64, RR = 2^2048 mod m, RR' = 2^2080 mod m
		Coeff[1] = 0x1000                                      // (1), using radix 52
		amm_52x20_x1_ifma256(rr_0, rr_0, rr_0, m_0, Mq.N0[0])  // (2) for m1
		amm_52x20_x1_ifma256(rr_0, rr_0, coeff, m_0, Mq.N0[0]) // (3) for m1
		amm_52x20_x1_ifma256(rr_1, rr_1, rr_1, m_1, Mp.N0[0])  // (2) for m2
		amm_52x20_x1_ifma256(rr_1, rr_1, coeff, m_1, Mp.N0[0]) // (3) for m2

		// Compute table of powers base^i, i = 0, ..., (2^EXP_WIN_SIZE) - 1
		//  table[0] = mont(x^0) = mont(1)
		//  table[1] = mont(x^1) = mont(x)
		X[0] = 1
		X[20] = 1
		k0 := [2]uint64{Mq.N0[0], Mp.N0[0]}
		amm_52x20_x2_ifma256(&(Table[0]), x, rr_0, m_0, &(k0[0]))
		amm_52x20_x2_ifma256(&(Table[40]), base_0, rr_0, m_0, &(k0[0]))
		for idx := 1; idx < 16; idx++ {
			amm_52x20_x2_ifma256(&(Table[(2*idx)*40]), &(Table[(1*idx)*40]), &(Table[(1*idx)*40]), m_0, &(k0[0]))
			amm_52x20_x2_ifma256(&(Table[(2*idx+1)*40]), &(Table[(2*idx)*40]), &(Table[40]), m_0, &(k0[0]))
		}

		// copy and expand exponents
		copy(Expz, priv.Precomputed.Dq.Bits())
		copy(Expz[17:], priv.Precomputed.Dp.Bits())
		exponentiation(x, y, m_0, &(Table[0]), &(k0[0]), Expz)

		// after the last AMM of exponentiation in Montgomery domain, the result
		// may be (modulus_bitsize + 1), but the conversion out of Montgomery domain
		// performs an AMM(x,1) which guarantees that the final result is less than
		// |m|, so no conditional subtraction is needed here. See [1] for details.
		// convert result back in regular 2^52 domain
		for i := range X {
			X[i] = 0
		}
		X[0] = 1
		X[20] = 1
		amm_52x20_x2_ifma256(rr_0, y, x, m_0, &(k0[0]))

		// convert results back to radix 2^64
		cvt_20x52_norm_to_16x64(X, rr_0)
		cvt_20x52_norm_to_16x64(Y, rr_1)
		bn_reduce_once_in_place_16(rr_0, x, &(Q.Bits()[0]))
		bn_reduce_once_in_place_16(&(RR52[16]), y, &(P.Bits()[0]))
		M1.SetBits(RR52[0:16])
		R1.SetBits(RR52[16:32])
		R1.SetBits(bn_mod_sub_fixed_top(R1.Bits(), M1.Bits(), P.Bits()))

		// r1 = r1 * iqmp mod p
		r1 := &(R1.Bits()[0])
		bn_mul_mont_16(r1, r1, &(Mp.RR.Bits()[0]), &(Mp.N.Bits()[0]), Mp.N0[0])
		bn_mul_mont_16(r1, r1, &(priv.Precomputed.Qinv.Bits()[0]), &(Mp.N.Bits()[0]), Mp.N0[0])
		R0.Mul(R1, Q)
		R0.Add(R0, M1)

		// verify
		//		if check {
		//			ct, _ := encrypt(&priv.PublicKey, R0.Bytes())
		//			if bytes.Compare(ct, ciphertext) != 0 {
		//				return nil, ErrDecryption
		//			}
		//		}
		return R0.Bytes(), nil

	// multi-prime
	default:
		var M [RSA_MAX_NUM_PRIMES - 2]*big.Int
		var PP [RSA_MAX_NUM_PRIMES]*big.Int

		// I mod Q
		R1 := new(big.Int).Mod(c, priv.Primes[1])
		R2 := new(big.Int)
		M1 := bn_mod_exp(R1, priv.Precomputed.Dq, priv.Primes[1])

		// I mod P
		R1 = R1.Mod(c, priv.Primes[0])

		// R0 = R1^dmp1 mod P
		R0 := bn_mod_exp(R1, priv.Precomputed.Dp, priv.Primes[0])

		for i := 2; i < len(priv.Primes); i++ {
			// I mod P
			R1 = R1.Mod(c, priv.Primes[i])

			// M[i] = R1 ^ dmp[i] mod P[i]
			M[i-2] = bn_mod_exp(R1, priv.Precomputed.CRTValues[i-2].Exp, priv.Primes[i])
		}

		// stop size of r0 increasing, which does affect the multiply if it optimised for a power of 2 size
		R0.Sub(R0, M1)
		if R0.Sign() < 0 {
			R0.Add(R0, priv.Primes[0]) // R0 = R0 - P
		}
		R1.Mul(R0, priv.Precomputed.Qinv)
		R0.Mod(R1, priv.Primes[0])

		// if p < q it is occasionally possible for the correction of adding 'p'
		// if r0 is negative above to leave the result still negative. This can
		// break the private key operations: the following second correction
		// should *always* correct this rare occurrence.
		if R0.Sign() < 0 {
			R0.Add(R0, priv.Primes[0])
		}
		R1.Mul(R0, priv.Primes[1])
		R0.Add(R1, M1)

		// use R1 to compute PP[i]
		PP[2] = new(big.Int).Mul(priv.Primes[0], priv.Primes[1])
		for i := 3; i < len(priv.Primes); i++ {
			PP[i] = new(big.Int)
			PP[i].Mul(priv.Primes[i], PP[i-1])
		}

		for i := 2; i < len(priv.Primes); i++ {
			R1.Sub(M[i-2], R0)
			R2.Mul(R1, priv.Precomputed.CRTValues[i-2].Coeff)
			R1.Mod(R2, priv.Primes[i])
			if R1.Sign() < 0 {
				R1.Add(R1, priv.Primes[i])
			}
			R1.Mul(R1, PP[i])
			R0.Add(R0, R1)
		}

		// verify
		//		if check {
		//			ct, _ := encrypt(&priv.PublicKey, R0.Bytes())
		//			if bytes.Compare(ct, ciphertext) != 0 {
		//				return nil, ErrDecryption
		//			}
		//		}
		return R0.Bytes(), nil
	}
}
