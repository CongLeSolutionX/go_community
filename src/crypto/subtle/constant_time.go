// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package subtle implements functions that are often useful in cryptographic
// code but require careful thought to use correctly.
package subtle

import (
	"unsafe"
)

// ConstantTimeCompare returns 1 if and only if the two slices, x
// and y, have equal contents. The time taken is a function of the length of
// the slices and is independent of the contents.
func ConstantTimeCompare(x, y []byte) int {
	if len(x) != len(y) {
		return 0
	}

	if len(x) == 0 {
		return 1
	}

	var v int32

	trailstart := 0
	if len(x) >= 12 {
		xp := (uintptr)(unsafe.Pointer(&x[0]))
		yp := (uintptr)(unsafe.Pointer(&y[0]))

		switch {
		// same alignment on 4 byte boundary
		case xp&0x7 == 0x4 && yp&0x7 == 0x4:

			var vq int64

			vq |= int64(*(*int32)(unsafe.Pointer(xp)) ^ *(*int32)(unsafe.Pointer(yp)))
			xp += 4
			yp += 4
			length := uintptr((len(x) - 4) / 8)

			trailstart = int(length*8 + 4)
			for i := uintptr(0); i < length; i++ {
				vq |= *(*int64)(unsafe.Pointer(xp + i*8)) ^ *(*int64)(unsafe.Pointer(yp + i*8))
			}
			v = int32(vq) | int32(vq>>32)

		// aligned on 8 byte boundary
		case xp&0x7 == 0 && yp&0x7 == 0:
			length := uintptr(len(x) / 8)

			trailstart = int(length * 8)
			var vq int64

			for i := uintptr(0); i < length; i++ {
				vq |= *(*int64)(unsafe.Pointer(xp + i*8)) ^ *(*int64)(unsafe.Pointer(yp + i*8))
			}
			v = int32(vq) | int32(vq>>32)

		// aligned on mismatched 4 byte boundary
		case xp&0x3 == 0 && yp&0x3 == 0:
			length := uintptr(len(x) / 4)
			trailstart = int(length * 4)
			for i := uintptr(0); i < length; i++ {
				v |= *(*int32)(unsafe.Pointer(xp + i*4)) ^ *(*int32)(unsafe.Pointer(yp + i*4))
			}

		// not aligned
		default:
			trailstart = len(x)
			var vb byte
			for i := 0; i < len(x); i++ {
				vb |= x[i] ^ y[i]
			}
			v = int32(vb)
		}
	}

	for i := trailstart; i < len(x); i++ {
		v |= int32(x[i]) ^ int32(y[i])
	}
	return ConstantTimeEq(v, 0)
}

// ConstantTimeSelect returns x if v is 1 and y if v is 0.
// Its behavior is undefined if v takes any other value.
func ConstantTimeSelect(v, x, y int) int { return ^(v-1)&x | (v-1)&y }

// ConstantTimeByteEq returns 1 if x == y and 0 otherwise.
func ConstantTimeByteEq(x, y uint8) int {
	z := ^(x ^ y)
	z &= z >> 4
	z &= z >> 2
	z &= z >> 1

	return int(z)
}

// ConstantTimeEq returns 1 if x == y and 0 otherwise.
func ConstantTimeEq(x, y int32) int {
	z := ^(x ^ y)
	z &= z >> 16
	z &= z >> 8
	z &= z >> 4
	z &= z >> 2
	z &= z >> 1

	return int(z & 1)
}

// ConstantTimeCopy copies the contents of y into x (a slice of equal length)
// if v == 1. If v == 0, x is left unchanged. Its behavior is undefined if v
// takes any other value.
func ConstantTimeCopy(v int, x, y []byte) {
	if len(x) != len(y) {
		panic("subtle: slices have different lengths")
	}

	xmask := byte(v - 1)
	ymask := byte(^(v - 1))
	for i := 0; i < len(x); i++ {
		x[i] = x[i]&xmask | y[i]&ymask
	}
}

// ConstantTimeLessOrEq returns 1 if x <= y and 0 otherwise.
// Its behavior is undefined if x or y are negative or > 2**31 - 1.
func ConstantTimeLessOrEq(x, y int) int {
	x32 := int32(x)
	y32 := int32(y)
	return int(((x32 - y32 - 1) >> 31) & 1)
}
