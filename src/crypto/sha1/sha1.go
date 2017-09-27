// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha1 implements the SHA-1 hash algorithm as defined in RFC 3174.
//
// SHA-1 is cryptographically broken and should not be used for secure
// applications.
package sha1

import (
	"crypto"
	"errors"
	"hash"
)

func init() {
	crypto.RegisterHash(crypto.SHA1, New)
}

// The size of a SHA-1 checksum in bytes.
const Size = 20

// The blocksize of SHA-1 in bytes.
const BlockSize = 64

const (
	chunk = 64
	init0 = 0x67452301
	init1 = 0xEFCDAB89
	init2 = 0x98BADCFE
	init3 = 0x10325476
	init4 = 0xC3D2E1F0
)

const marshaledDigestSize = 4 + 5*4 + chunk + 4 + 8

// digest represents the partial evaluation of a checksum.
type digest struct {
	h   [5]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, marshaledDigestSize)
	b[0], b[1], b[2], b[3] = 's', 'h', 'a', 0x01

	b[4], b[5], b[6], b[7] = byte(d.h[0]>>24), byte(d.h[0]>>16), byte(d.h[0]>>8), byte(d.h[0])
	b[8], b[9], b[10], b[11] = byte(d.h[1]>>24), byte(d.h[1]>>16), byte(d.h[1]>>8), byte(d.h[1])
	b[12], b[13], b[14], b[15] = byte(d.h[2]>>24), byte(d.h[2]>>16), byte(d.h[2]>>8), byte(d.h[2])
	b[16], b[17], b[18], b[19] = byte(d.h[3]>>24), byte(d.h[3]>>16), byte(d.h[3]>>8), byte(d.h[3])
	b[20], b[21], b[22], b[23] = byte(d.h[4]>>24), byte(d.h[4]>>16), byte(d.h[4]>>8), byte(d.h[4])

	copy(b[24:], d.x[:])

	b[88], b[89], b[90], b[91] = byte(d.nx>>24), byte(d.nx>>16), byte(d.nx>>8), byte(d.nx)

	b[92], b[93], b[94], b[95] = byte(d.len>>56), byte(d.len>>48), byte(d.len>>40), byte(d.len>>32)
	b[96], b[97], b[98], b[99] = byte(d.len>>24), byte(d.len>>16), byte(d.len>>8), byte(d.len)

	return b, nil
}

func (d *digest) UnmarshalBinary(data []byte) error {
	if len(data) != marshaledDigestSize || data[0] != 's' || data[1] != 'h' || data[2] != 'a' || data[3] != 0x01 {
		return errors.New("crypto/sha1: invalid state")
	}

	d.h[0] = uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	d.h[1] = uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11])
	d.h[2] = uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15])
	d.h[3] = uint32(data[16])<<24 | uint32(data[17])<<16 | uint32(data[18])<<8 | uint32(data[19])
	d.h[4] = uint32(data[20])<<24 | uint32(data[21])<<16 | uint32(data[22])<<8 | uint32(data[23])

	copy(d.x[:], data[24:])

	d.nx = int(data[88])<<24 | int(data[89])<<16 | int(data[90])<<8 | int(data[91])

	d.len = uint64(data[92])<<56 | uint64(data[93])<<48 | uint64(data[94])<<40 | uint64(data[95])<<32 |
		uint64(data[96])<<24 | uint64(data[97])<<16 | uint64(data[98])<<8 | uint64(data[99])

	return nil
}

func (d *digest) Reset() {
	d.h[0] = init0
	d.h[1] = init1
	d.h[2] = init2
	d.h[3] = init3
	d.h[4] = init4
	d.nx = 0
	d.len = 0
}

// New returns a new hash.Hash computing the SHA1 checksum.
func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return BlockSize }

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == chunk {
			block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}
	if len(p) >= chunk {
		n := len(p) &^ (chunk - 1)
		block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

func (d0 *digest) Sum(in []byte) []byte {
	// Make a copy of d0 so that caller can keep writing and summing.
	d := *d0
	hash := d.checkSum()
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	len := d.len
	// Padding.  Add a 1 bit and 0 bits until 56 bytes mod 64.
	var tmp [64]byte
	tmp[0] = 0x80
	if len%64 < 56 {
		d.Write(tmp[0 : 56-len%64])
	} else {
		d.Write(tmp[0 : 64+56-len%64])
	}

	// Length in bits.
	len <<= 3
	putUint64(tmp[:], len)
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte

	putUint32(digest[0:], d.h[0])
	putUint32(digest[4:], d.h[1])
	putUint32(digest[8:], d.h[2])
	putUint32(digest[12:], d.h[3])
	putUint32(digest[16:], d.h[4])

	return digest
}

// ConstantTimeSum computes the same result of Sum() but in constant time
func (d0 *digest) ConstantTimeSum(in []byte) []byte {
	d := *d0
	hash := d.constSum()
	return append(in, hash[:]...)
}

func (d *digest) constSum() [Size]byte {
	var length [8]byte
	l := d.len << 3
	for i := uint(0); i < 8; i++ {
		length[i] = byte(l >> (56 - 8*i))
	}

	nx := byte(d.nx)
	t := nx - 56                 // if nx < 56 then the MSB of t is one
	mask1b := byte(int8(t) >> 7) // mask1b is 0xFF iff one block is enough

	separator := byte(0x80) // gets reset to 0x00 once used
	for i := byte(0); i < chunk; i++ {
		mask := byte(int8(i-nx) >> 7) // 0x00 after the end of data

		// if we reached the end of the data, replace with 0x80 or 0x00
		d.x[i] = (^mask & separator) | (mask & d.x[i])

		// zero the separator once used
		separator &= mask

		if i >= 56 {
			// we might have to write the length here if all fit in one block
			d.x[i] |= mask1b & length[i-56]
		}
	}

	// compress, and only keep the digest if all fit in one block
	block(d, d.x[:])

	var digest [Size]byte
	for i, s := range d.h {
		digest[i*4] = mask1b & byte(s>>24)
		digest[i*4+1] = mask1b & byte(s>>16)
		digest[i*4+2] = mask1b & byte(s>>8)
		digest[i*4+3] = mask1b & byte(s)
	}

	for i := byte(0); i < chunk; i++ {
		// second block, it's always past the end of data, might start with 0x80
		if i < 56 {
			d.x[i] = separator
			separator = 0
		} else {
			d.x[i] = length[i-56]
		}
	}

	// compress, and only keep the digest if we actually needed the second block
	block(d, d.x[:])

	for i, s := range d.h {
		digest[i*4] |= ^mask1b & byte(s>>24)
		digest[i*4+1] |= ^mask1b & byte(s>>16)
		digest[i*4+2] |= ^mask1b & byte(s>>8)
		digest[i*4+3] |= ^mask1b & byte(s)
	}

	return digest
}

// Sum returns the SHA-1 checksum of the data.
func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}

func putUint64(x []byte, s uint64) {
	_ = x[7]
	x[0] = byte(s >> 56)
	x[1] = byte(s >> 48)
	x[2] = byte(s >> 40)
	x[3] = byte(s >> 32)
	x[4] = byte(s >> 24)
	x[5] = byte(s >> 16)
	x[6] = byte(s >> 8)
	x[7] = byte(s)
}

func putUint32(x []byte, s uint32) {
	_ = x[3]
	x[0] = byte(s >> 24)
	x[1] = byte(s >> 16)
	x[2] = byte(s >> 8)
	x[3] = byte(s)
}
