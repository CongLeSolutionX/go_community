// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sha512 implements the SHA-384, SHA-512, SHA-512/224, and SHA-512/256
// hash algorithms as defined in FIPS 180-4.
package sha512

import (
	"crypto"
	"errors"
	"hash"
)

func init() {
	crypto.RegisterHash(crypto.SHA384, New384)
	crypto.RegisterHash(crypto.SHA512, New)
	crypto.RegisterHash(crypto.SHA512_224, New512_224)
	crypto.RegisterHash(crypto.SHA512_256, New512_256)
}

const (
	// Size is the size, in bytes, of a SHA-512 checksum.
	Size = 64

	// Size224 is the size, in bytes, of a SHA-512/224 checksum.
	Size224 = 28

	// Size256 is the size, in bytes, of a SHA-512/256 checksum.
	Size256 = 32

	// Size384 is the size, in bytes, of a SHA-384 checksum.
	Size384 = 48

	// BlockSize is the block size, in bytes, of the SHA-512/224,
	// SHA-512/256, SHA-384 and SHA-512 hash functions.
	BlockSize = 128
)

const (
	chunk     = 128
	init0     = 0x6a09e667f3bcc908
	init1     = 0xbb67ae8584caa73b
	init2     = 0x3c6ef372fe94f82b
	init3     = 0xa54ff53a5f1d36f1
	init4     = 0x510e527fade682d1
	init5     = 0x9b05688c2b3e6c1f
	init6     = 0x1f83d9abfb41bd6b
	init7     = 0x5be0cd19137e2179
	init0_224 = 0x8c3d37c819544da2
	init1_224 = 0x73e1996689dcd4d6
	init2_224 = 0x1dfab7ae32ff9c82
	init3_224 = 0x679dd514582f9fcf
	init4_224 = 0x0f6d2b697bd44da8
	init5_224 = 0x77e36f7304c48942
	init6_224 = 0x3f9d85a86a1d36c8
	init7_224 = 0x1112e6ad91d692a1
	init0_256 = 0x22312194fc2bf72c
	init1_256 = 0x9f555fa3c84c64c2
	init2_256 = 0x2393b86b6f53b151
	init3_256 = 0x963877195940eabd
	init4_256 = 0x96283ee2a88effe3
	init5_256 = 0xbe5e1e2553863992
	init6_256 = 0x2b0199fc2c85b8aa
	init7_256 = 0x0eb72ddc81c52ca2
	init0_384 = 0xcbbb9d5dc1059ed8
	init1_384 = 0x629a292a367cd507
	init2_384 = 0x9159015a3070dd17
	init3_384 = 0x152fecd8f70e5939
	init4_384 = 0x67332667ffc00b31
	init5_384 = 0x8eb44a8768581511
	init6_384 = 0xdb0c2e0d64f98fa7
	init7_384 = 0x47b5481dbefa4fa4
)

const marshaledDigestSize = 4 + 8*8 + chunk + 8 + 8

// digest represents the partial evaluation of a checksum.
type digest struct {
	h        [8]uint64
	x        [chunk]byte
	nx       int
	len      uint64
	function crypto.Hash
}

func (d *digest) Reset() {
	switch d.function {
	case crypto.SHA384:
		d.h[0] = init0_384
		d.h[1] = init1_384
		d.h[2] = init2_384
		d.h[3] = init3_384
		d.h[4] = init4_384
		d.h[5] = init5_384
		d.h[6] = init6_384
		d.h[7] = init7_384
	case crypto.SHA512_224:
		d.h[0] = init0_224
		d.h[1] = init1_224
		d.h[2] = init2_224
		d.h[3] = init3_224
		d.h[4] = init4_224
		d.h[5] = init5_224
		d.h[6] = init6_224
		d.h[7] = init7_224
	case crypto.SHA512_256:
		d.h[0] = init0_256
		d.h[1] = init1_256
		d.h[2] = init2_256
		d.h[3] = init3_256
		d.h[4] = init4_256
		d.h[5] = init5_256
		d.h[6] = init6_256
		d.h[7] = init7_256
	default:
		d.h[0] = init0
		d.h[1] = init1
		d.h[2] = init2
		d.h[3] = init3
		d.h[4] = init4
		d.h[5] = init5
		d.h[6] = init6
		d.h[7] = init7
	}
	d.nx = 0
	d.len = 0
}

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, marshaledDigestSize)
	b[0], b[1], b[2] = 's', 'h', 'a'
	switch d.function {
	case crypto.SHA384:
		b[3] = 0x04
	case crypto.SHA512_224:
		b[3] = 0x05
	case crypto.SHA512_256:
		b[3] = 0x06
	case crypto.SHA512:
		b[3] = 0x07
	}

	b[4], b[5], b[6], b[7] = byte(d.h[0]>>56), byte(d.h[0]>>48), byte(d.h[0]>>40), byte(d.h[0]>>32)
	b[8], b[9], b[10], b[11] = byte(d.h[0]>>24), byte(d.h[0]>>16), byte(d.h[0]>>8), byte(d.h[0])
	b[12], b[13], b[14], b[15] = byte(d.h[1]>>56), byte(d.h[1]>>48), byte(d.h[1]>>40), byte(d.h[1]>>32)
	b[16], b[17], b[18], b[19] = byte(d.h[1]>>24), byte(d.h[1]>>16), byte(d.h[1]>>8), byte(d.h[1])
	b[20], b[21], b[22], b[23] = byte(d.h[2]>>56), byte(d.h[2]>>48), byte(d.h[2]>>40), byte(d.h[2]>>32)
	b[24], b[25], b[26], b[27] = byte(d.h[2]>>24), byte(d.h[2]>>16), byte(d.h[2]>>8), byte(d.h[2])
	b[28], b[29], b[30], b[31] = byte(d.h[3]>>56), byte(d.h[3]>>48), byte(d.h[3]>>40), byte(d.h[3]>>32)
	b[32], b[33], b[34], b[35] = byte(d.h[3]>>24), byte(d.h[3]>>16), byte(d.h[3]>>8), byte(d.h[3])
	b[36], b[37], b[38], b[39] = byte(d.h[4]>>56), byte(d.h[4]>>48), byte(d.h[4]>>40), byte(d.h[4]>>32)
	b[40], b[41], b[42], b[43] = byte(d.h[4]>>24), byte(d.h[4]>>16), byte(d.h[4]>>8), byte(d.h[4])
	b[44], b[45], b[46], b[47] = byte(d.h[5]>>56), byte(d.h[5]>>48), byte(d.h[5]>>40), byte(d.h[5]>>32)
	b[48], b[49], b[50], b[51] = byte(d.h[5]>>24), byte(d.h[5]>>16), byte(d.h[5]>>8), byte(d.h[5])
	b[52], b[53], b[54], b[55] = byte(d.h[6]>>56), byte(d.h[6]>>48), byte(d.h[6]>>40), byte(d.h[6]>>32)
	b[56], b[57], b[58], b[59] = byte(d.h[6]>>24), byte(d.h[6]>>16), byte(d.h[6]>>8), byte(d.h[6])
	b[60], b[61], b[62], b[63] = byte(d.h[7]>>56), byte(d.h[7]>>48), byte(d.h[7]>>40), byte(d.h[7]>>32)
	b[64], b[65], b[66], b[67] = byte(d.h[7]>>24), byte(d.h[7]>>16), byte(d.h[7]>>8), byte(d.h[7])

	copy(b[68:], d.x[:])

	b[196], b[197], b[198], b[199] = byte(d.nx>>24), byte(d.nx>>16), byte(d.nx>>8), byte(d.nx)

	b[200], b[201], b[202], b[203] = byte(d.len>>56), byte(d.len>>48), byte(d.len>>40), byte(d.len>>32)
	b[204], b[205], b[206], b[207] = byte(d.len>>24), byte(d.len>>16), byte(d.len>>8), byte(d.len)

	return b, nil
}

func (d *digest) UnmarshalBinary(data []byte) error {
	if len(data) != marshaledDigestSize || data[0] != 's' || data[1] != 'h' || data[2] != 'a' {
		return errors.New("crypto/sha512: invalid state")
	}
	switch {
	case d.function == crypto.SHA384 && data[3] == 0x04:
	case d.function == crypto.SHA512_224 && data[3] == 0x05:
	case d.function == crypto.SHA512_256 && data[3] == 0x06:
	case d.function == crypto.SHA512 && data[3] == 0x07:
	default:
		return errors.New("crypto/sha512: invalid state")
	}

	d.h[0] = uint64(data[4])<<56 | uint64(data[5])<<48 | uint64(data[6])<<40 | uint64(data[7])<<32 |
		uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	d.h[1] = uint64(data[12])<<56 | uint64(data[13])<<48 | uint64(data[14])<<40 | uint64(data[15])<<32 |
		uint64(data[16])<<24 | uint64(data[17])<<16 | uint64(data[18])<<8 | uint64(data[19])
	d.h[2] = uint64(data[20])<<56 | uint64(data[21])<<48 | uint64(data[22])<<40 | uint64(data[23])<<32 |
		uint64(data[24])<<24 | uint64(data[25])<<16 | uint64(data[26])<<8 | uint64(data[27])
	d.h[3] = uint64(data[28])<<56 | uint64(data[29])<<48 | uint64(data[30])<<40 | uint64(data[31])<<32 |
		uint64(data[32])<<24 | uint64(data[33])<<16 | uint64(data[34])<<8 | uint64(data[35])
	d.h[4] = uint64(data[36])<<56 | uint64(data[37])<<48 | uint64(data[38])<<40 | uint64(data[39])<<32 |
		uint64(data[40])<<24 | uint64(data[41])<<16 | uint64(data[42])<<8 | uint64(data[43])
	d.h[5] = uint64(data[44])<<56 | uint64(data[45])<<48 | uint64(data[46])<<40 | uint64(data[47])<<32 |
		uint64(data[48])<<24 | uint64(data[49])<<16 | uint64(data[50])<<8 | uint64(data[51])
	d.h[6] = uint64(data[52])<<56 | uint64(data[53])<<48 | uint64(data[54])<<40 | uint64(data[55])<<32 |
		uint64(data[56])<<24 | uint64(data[57])<<16 | uint64(data[58])<<8 | uint64(data[59])
	d.h[7] = uint64(data[60])<<56 | uint64(data[61])<<48 | uint64(data[62])<<40 | uint64(data[63])<<32 |
		uint64(data[64])<<24 | uint64(data[65])<<16 | uint64(data[66])<<8 | uint64(data[67])

	copy(d.x[:], data[68:])

	d.nx = int(data[196])<<24 | int(data[197])<<16 | int(data[198])<<8 | int(data[199])

	d.len = uint64(data[200])<<56 | uint64(data[201])<<48 | uint64(data[202])<<40 | uint64(data[203])<<32 |
		uint64(data[204])<<24 | uint64(data[205])<<16 | uint64(data[206])<<8 | uint64(data[207])

	return nil
}

// New returns a new hash.Hash computing the SHA-512 checksum.
func New() hash.Hash {
	d := &digest{function: crypto.SHA512}
	d.Reset()
	return d
}

// New512_224 returns a new hash.Hash computing the SHA-512/224 checksum.
func New512_224() hash.Hash {
	d := &digest{function: crypto.SHA512_224}
	d.Reset()
	return d
}

// New512_256 returns a new hash.Hash computing the SHA-512/256 checksum.
func New512_256() hash.Hash {
	d := &digest{function: crypto.SHA512_256}
	d.Reset()
	return d
}

// New384 returns a new hash.Hash computing the SHA-384 checksum.
func New384() hash.Hash {
	d := &digest{function: crypto.SHA384}
	d.Reset()
	return d
}

func (d *digest) Size() int {
	switch d.function {
	case crypto.SHA512_224:
		return Size224
	case crypto.SHA512_256:
		return Size256
	case crypto.SHA384:
		return Size384
	default:
		return Size
	}
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
	d := new(digest)
	*d = *d0
	hash := d.checkSum()
	switch d.function {
	case crypto.SHA384:
		return append(in, hash[:Size384]...)
	case crypto.SHA512_224:
		return append(in, hash[:Size224]...)
	case crypto.SHA512_256:
		return append(in, hash[:Size256]...)
	default:
		return append(in, hash[:]...)
	}
}

func (d *digest) checkSum() [Size]byte {
	// Padding. Add a 1 bit and 0 bits until 112 bytes mod 128.
	len := d.len
	var tmp [128]byte
	tmp[0] = 0x80
	if len%128 < 112 {
		d.Write(tmp[0 : 112-len%128])
	} else {
		d.Write(tmp[0 : 128+112-len%128])
	}

	// Length in bits.
	len <<= 3
	for i := uint(0); i < 16; i++ {
		tmp[i] = byte(len >> (120 - 8*i))
	}
	d.Write(tmp[0:16])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	h := d.h[:]
	if d.function == crypto.SHA384 {
		h = d.h[:6]
	}

	var digest [Size]byte
	for i, s := range h {
		digest[i*8] = byte(s >> 56)
		digest[i*8+1] = byte(s >> 48)
		digest[i*8+2] = byte(s >> 40)
		digest[i*8+3] = byte(s >> 32)
		digest[i*8+4] = byte(s >> 24)
		digest[i*8+5] = byte(s >> 16)
		digest[i*8+6] = byte(s >> 8)
		digest[i*8+7] = byte(s)
	}

	return digest
}

// Sum512 returns the SHA512 checksum of the data.
func Sum512(data []byte) [Size]byte {
	d := digest{function: crypto.SHA512}
	d.Reset()
	d.Write(data)
	return d.checkSum()
}

// Sum384 returns the SHA384 checksum of the data.
func Sum384(data []byte) (sum384 [Size384]byte) {
	d := digest{function: crypto.SHA384}
	d.Reset()
	d.Write(data)
	sum := d.checkSum()
	copy(sum384[:], sum[:Size384])
	return
}

// Sum512_224 returns the Sum512/224 checksum of the data.
func Sum512_224(data []byte) (sum224 [Size224]byte) {
	d := digest{function: crypto.SHA512_224}
	d.Reset()
	d.Write(data)
	sum := d.checkSum()
	copy(sum224[:], sum[:Size224])
	return
}

// Sum512_256 returns the Sum512/256 checksum of the data.
func Sum512_256(data []byte) (sum256 [Size256]byte) {
	d := digest{function: crypto.SHA512_256}
	d.Reset()
	d.Write(data)
	sum := d.checkSum()
	copy(sum256[:], sum[:Size256])
	return
}
