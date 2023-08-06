// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

func chacha8setup(counter uint64, seed *[8]uint32, b32 *[16][4]uint32) {
	b := (*[16][2]uint64)(unsafe.Pointer(b32))

	b[0][0] = 0x61707865_61707865
	b[0][1] = 0x61707865_61707865

	b[1][0] = 0x3320646e_3320646e
	b[1][1] = 0x3320646e_3320646e

	b[2][0] = 0x79622d32_79622d32
	b[2][1] = 0x79622d32_79622d32

	b[3][0] = 0x6b206574_6b206574
	b[3][1] = 0x6b206574_6b206574

	var x64 uint64
	var x uint32

	x = seed[0]
	x64 = uint64(x)<<32 | uint64(x)
	b[4][0] = x64
	b[4][1] = x64

	x = seed[1]
	x64 = uint64(x)<<32 | uint64(x)
	b[5][0] = x64
	b[5][1] = x64

	x = seed[2]
	x64 = uint64(x)<<32 | uint64(x)
	b[6][0] = x64
	b[6][1] = x64

	x = seed[3]
	x64 = uint64(x)<<32 | uint64(x)
	b[7][0] = x64
	b[7][1] = x64

	x = seed[4]
	x64 = uint64(x)<<32 | uint64(x)
	b[8][0] = x64
	b[8][1] = x64

	x = seed[5]
	x64 = uint64(x)<<32 | uint64(x)
	b[9][0] = x64
	b[9][1] = x64

	x = seed[6]
	x64 = uint64(x)<<32 | uint64(x)
	b[10][0] = x64
	b[10][1] = x64

	x = seed[7]
	x64 = uint64(x)<<32 | uint64(x)
	b[11][0] = x64
	b[11][1] = x64

	b[12][0] = uint64(uint32(counter)) | uint64(uint32(counter+1))<<32
	b[12][1] = uint64(uint32(counter+2)) | uint64(uint32(counter+3))<<32

	b[13][0] = uint64(uint32(counter>>32)) | uint64(uint32((counter+1)>>32))<<32
	b[13][1] = uint64(uint32((counter+2)>>32)) | uint64(uint32((counter+3)>>32))<<32

	b[14][0] = 0
	b[14][1] = 0

	b[15][0] = 0
	b[15][1] = 0
}

func chacha8block_generic(counter uint64, seedBytes *[32]byte, buf *[32]uint64) {
	seed := (*[8]uint32)(unsafe.Pointer(seedBytes))
	b := (*[16][4]uint32)(unsafe.Pointer(buf))

	chacha8setup(counter, seed, b)

	for i := range b[0] {
		b0 := b[0][i]
		b1 := b[1][i]
		b2 := b[2][i]
		b3 := b[3][i]
		b4 := b[4][i]
		b5 := b[5][i]
		b6 := b[6][i]
		b7 := b[7][i]
		b8 := b[8][i]
		b9 := b[9][i]
		b10 := b[10][i]
		b11 := b[11][i]
		b12 := b[12][i]
		b13 := b[13][i]
		b14 := b[14][i]
		b15 := b[15][i]

		for round := 0; round < 4; round++ {
			b0, b4, b8, b12 = chacha8qr(b0, b4, b8, b12)
			b1, b5, b9, b13 = chacha8qr(b1, b5, b9, b13)
			b2, b6, b10, b14 = chacha8qr(b2, b6, b10, b14)
			b3, b7, b11, b15 = chacha8qr(b3, b7, b11, b15)

			b0, b5, b10, b15 = chacha8qr(b0, b5, b10, b15)
			b1, b6, b11, b12 = chacha8qr(b1, b6, b11, b12)
			b2, b7, b8, b13 = chacha8qr(b2, b7, b8, b13)
			b3, b4, b9, b14 = chacha8qr(b3, b4, b9, b14)
		}

		b[0][i] = b0
		b[1][i] = b1
		b[2][i] = b2
		b[3][i] = b3
		b[4][i] = b4
		b[5][i] = b5
		b[6][i] = b6
		b[7][i] = b7
		b[8][i] = b8
		b[9][i] = b9
		b[10][i] = b10
		b[11][i] = b11
		b[12][i] = b12
		b[13][i] = b13
		b[14][i] = b14
		b[15][i] = b15
	}
}

func chacha8qr(a, b, c, d uint32) (_a, _b, _c, _d uint32) {
	a += b
	d ^= a
	d = d<<16 | d>>16
	c += d
	b ^= c
	b = b<<12 | b>>20
	a += b
	d ^= a
	d = d<<8 | d>>24
	c += d
	b ^= c
	b = b<<7 | b>>25
	return a, b, c, d
}
