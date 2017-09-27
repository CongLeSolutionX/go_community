// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha256 implements the SHA224 and SHA256 hash algorithms as defined
// in FIPS 180-4.
package sha256

import (
	"crypto"
	"errors"
	"hash"
)

func init() {
	crypto.RegisterHash(crypto.SHA224, New224)
	crypto.RegisterHash(crypto.SHA256, New)
}

// The size of a SHA256 checksum in bytes.
const Size = 32

// The size of a SHA224 checksum in bytes.
const Size224 = 28

// The blocksize of SHA256 and SHA224 in bytes.
const BlockSize = 64

const (
	chunk     = 64
	init0     = 0x6A09E667
	init1     = 0xBB67AE85
	init2     = 0x3C6EF372
	init3     = 0xA54FF53A
	init4     = 0x510E527F
	init5     = 0x9B05688C
	init6     = 0x1F83D9AB
	init7     = 0x5BE0CD19
	init0_224 = 0xC1059ED8
	init1_224 = 0x367CD507
	init2_224 = 0x3070DD17
	init3_224 = 0xF70E5939
	init4_224 = 0xFFC00B31
	init5_224 = 0x68581511
	init6_224 = 0x64F98FA7
	init7_224 = 0xBEFA4FA4
)

const marshaledDigestSize = 4 + 8*4 + chunk + 8 + 8

// digest represents the partial evaluation of a checksum.
type digest struct {
	h     [8]uint32
	x     [chunk]byte
	nx    int
	len   uint64
	is224 bool // mark if this digest is SHA-224
}

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, marshaledDigestSize)
	b[0], b[1], b[2] = 's', 'h', 'a'
	if d.is224 {
		b[3] = 0x02
	} else {
		b[3] = 0x03
	}

	b[4], b[5], b[6], b[7] = byte(d.h[0]>>24), byte(d.h[0]>>16), byte(d.h[0]>>8), byte(d.h[0])
	b[8], b[9], b[10], b[11] = byte(d.h[1]>>24), byte(d.h[1]>>16), byte(d.h[1]>>8), byte(d.h[1])
	b[12], b[13], b[14], b[15] = byte(d.h[2]>>24), byte(d.h[2]>>16), byte(d.h[2]>>8), byte(d.h[2])
	b[16], b[17], b[18], b[19] = byte(d.h[3]>>24), byte(d.h[3]>>16), byte(d.h[3]>>8), byte(d.h[3])
	b[20], b[21], b[22], b[23] = byte(d.h[4]>>24), byte(d.h[4]>>16), byte(d.h[4]>>8), byte(d.h[4])
	b[24], b[25], b[26], b[27] = byte(d.h[5]>>24), byte(d.h[5]>>16), byte(d.h[5]>>8), byte(d.h[5])
	b[28], b[29], b[30], b[31] = byte(d.h[6]>>24), byte(d.h[6]>>16), byte(d.h[6]>>8), byte(d.h[6])
	b[32], b[33], b[34], b[35] = byte(d.h[7]>>24), byte(d.h[7]>>16), byte(d.h[7]>>8), byte(d.h[7])

	copy(b[36:], d.x[:])

	b[100], b[101], b[102], b[103] = byte(d.nx>>24), byte(d.nx>>16), byte(d.nx>>8), byte(d.nx)

	b[104], b[105], b[106], b[107] = byte(d.len>>56), byte(d.len>>48), byte(d.len>>40), byte(d.len>>32)
	b[108], b[109], b[110], b[111] = byte(d.len>>24), byte(d.len>>16), byte(d.len>>8), byte(d.len)

	return b, nil
}

func (d *digest) UnmarshalBinary(data []byte) error {
	if len(data) != marshaledDigestSize || data[0] != 's' || data[1] != 'h' || data[2] != 'a' || (d.is224 && data[3] != 0x02) || (!d.is224 && data[3] != 0x03) {
		return errors.New("crypto/sha256: invalid state")
	}

	d.h[0] = uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	d.h[1] = uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11])
	d.h[2] = uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15])
	d.h[3] = uint32(data[16])<<24 | uint32(data[17])<<16 | uint32(data[18])<<8 | uint32(data[19])
	d.h[4] = uint32(data[20])<<24 | uint32(data[21])<<16 | uint32(data[22])<<8 | uint32(data[23])
	d.h[5] = uint32(data[24])<<24 | uint32(data[25])<<16 | uint32(data[26])<<8 | uint32(data[27])
	d.h[6] = uint32(data[28])<<24 | uint32(data[29])<<16 | uint32(data[30])<<8 | uint32(data[31])
	d.h[7] = uint32(data[32])<<24 | uint32(data[33])<<16 | uint32(data[34])<<8 | uint32(data[35])

	copy(d.x[:], data[36:])

	d.nx = int(data[100])<<24 | int(data[101])<<16 | int(data[102])<<8 | int(data[103])

	d.len = uint64(data[104])<<56 | uint64(data[105])<<48 | uint64(data[106])<<40 | uint64(data[107])<<32 |
		uint64(data[108])<<24 | uint64(data[109])<<16 | uint64(data[110])<<8 | uint64(data[111])

	return nil
}

func (d *digest) Reset() {
	if !d.is224 {
		d.h[0] = init0
		d.h[1] = init1
		d.h[2] = init2
		d.h[3] = init3
		d.h[4] = init4
		d.h[5] = init5
		d.h[6] = init6
		d.h[7] = init7
	} else {
		d.h[0] = init0_224
		d.h[1] = init1_224
		d.h[2] = init2_224
		d.h[3] = init3_224
		d.h[4] = init4_224
		d.h[5] = init5_224
		d.h[6] = init6_224
		d.h[7] = init7_224
	}
	d.nx = 0
	d.len = 0
}

// New returns a new hash.Hash computing the SHA256 checksum.
func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

// New224 returns a new hash.Hash computing the SHA224 checksum.
func New224() hash.Hash {
	d := new(digest)
	d.is224 = true
	d.Reset()
	return d
}

func (d *digest) Size() int {
	if !d.is224 {
		return Size
	}
	return Size224
}

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
	if d.is224 {
		return append(in, hash[:Size224]...)
	}
	return append(in, hash[:]...)
}

func (d *digest) checkSum() [Size]byte {
	len := d.len
	// Padding. Add a 1 bit and 0 bits until 56 bytes mod 64.
	var tmp [64]byte
	tmp[0] = 0x80
	if len%64 < 56 {
		d.Write(tmp[0 : 56-len%64])
	} else {
		d.Write(tmp[0 : 64+56-len%64])
	}

	// Length in bits.
	len <<= 3
	for i := uint(0); i < 8; i++ {
		tmp[i] = byte(len >> (56 - 8*i))
	}
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	h := d.h[:]
	if d.is224 {
		h = d.h[:7]
	}

	var digest [Size]byte
	for i, s := range h {
		digest[i*4] = byte(s >> 24)
		digest[i*4+1] = byte(s >> 16)
		digest[i*4+2] = byte(s >> 8)
		digest[i*4+3] = byte(s)
	}

	return digest
}

// Sum256 returns the SHA256 checksum of the data.
func Sum256(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}

// Sum224 returns the SHA224 checksum of the data.
func Sum224(data []byte) (sum224 [Size224]byte) {
	var d digest
	d.is224 = true
	d.Reset()
	d.Write(data)
	sum := d.checkSum()
	copy(sum224[:], sum[:Size224])
	return
}
