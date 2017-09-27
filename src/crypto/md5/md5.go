// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gen.go -full -output md5block.go

// Package md5 implements the MD5 hash algorithm as defined in RFC 1321.
//
// MD5 is cryptographically broken and should not be used for secure
// applications.
package md5

import (
	"crypto"
	"errors"
	"hash"
)

func init() {
	crypto.RegisterHash(crypto.MD5, New)
}

// The size of an MD5 checksum in bytes.
const Size = 16

// The blocksize of MD5 in bytes.
const BlockSize = 64

const (
	chunk = 64
	init0 = 0x67452301
	init1 = 0xEFCDAB89
	init2 = 0x98BADCFE
	init3 = 0x10325476
)

const marshaledDigestSize = 4 + 4*4 + chunk + 4 + 8

// digest represents the partial evaluation of a checksum.
type digest struct {
	s   [4]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

func (d *digest) Reset() {
	d.s[0] = init0
	d.s[1] = init1
	d.s[2] = init2
	d.s[3] = init3
	d.nx = 0
	d.len = 0
}

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, marshaledDigestSize)
	b[0], b[1], b[2], b[3] = 'm', 'd', '5', 0x01

	b[4], b[5], b[6], b[7] = byte(d.s[0]>>24), byte(d.s[0]>>16), byte(d.s[0]>>8), byte(d.s[0])
	b[8], b[9], b[10], b[11] = byte(d.s[1]>>24), byte(d.s[1]>>16), byte(d.s[1]>>8), byte(d.s[1])
	b[12], b[13], b[14], b[15] = byte(d.s[2]>>24), byte(d.s[2]>>16), byte(d.s[2]>>8), byte(d.s[2])
	b[16], b[17], b[18], b[19] = byte(d.s[3]>>24), byte(d.s[3]>>16), byte(d.s[3]>>8), byte(d.s[3])

	copy(b[20:], d.x[:])

	b[84], b[85], b[86], b[87] = byte(d.nx>>24), byte(d.nx>>16), byte(d.nx>>8), byte(d.nx)

	b[88], b[89], b[90], b[91] = byte(d.len>>56), byte(d.len>>48), byte(d.len>>40), byte(d.len>>32)
	b[92], b[93], b[94], b[95] = byte(d.len>>24), byte(d.len>>16), byte(d.len>>8), byte(d.len)

	return b, nil
}

func (d *digest) UnmarshalBinary(data []byte) error {
	if len(data) != marshaledDigestSize || data[0] != 'm' || data[1] != 'd' || data[2] != '5' || data[3] != 0x01 {
		return errors.New("crypto/md5: invalid state")
	}

	d.s[0] = uint32(data[4])<<24 | uint32(data[5])<<16 | uint32(data[6])<<8 | uint32(data[7])
	d.s[1] = uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11])
	d.s[2] = uint32(data[12])<<24 | uint32(data[13])<<16 | uint32(data[14])<<8 | uint32(data[15])
	d.s[3] = uint32(data[16])<<24 | uint32(data[17])<<16 | uint32(data[18])<<8 | uint32(data[19])

	copy(d.x[:], data[20:])

	d.nx = int(data[84])<<24 | int(data[85])<<16 | int(data[86])<<8 | int(data[87])

	d.len = uint64(data[88])<<56 | uint64(data[89])<<48 | uint64(data[90])<<40 | uint64(data[91])<<32 |
		uint64(data[92])<<24 | uint64(data[93])<<16 | uint64(data[94])<<8 | uint64(data[95])

	return nil
}

// New returns a new hash.Hash computing the MD5 checksum.
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
	// Padding. Add a 1 bit and 0 bits until 56 bytes mod 64.
	len := d.len
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
		tmp[i] = byte(len >> (8 * i))
	}
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte
	for i, s := range d.s {
		digest[i*4] = byte(s)
		digest[i*4+1] = byte(s >> 8)
		digest[i*4+2] = byte(s >> 16)
		digest[i*4+3] = byte(s >> 24)
	}

	return digest
}

// Sum returns the MD5 checksum of the data.
func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}
