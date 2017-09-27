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
	"encoding/binary"
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

const marshaledDigestSize = 4*4 + chunk + 8 + 8

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

	binary.BigEndian.PutUint32(b, d.s[0])
	binary.BigEndian.PutUint32(b[4:], d.s[1])
	binary.BigEndian.PutUint32(b[8:], d.s[2])
	binary.BigEndian.PutUint32(b[12:], d.s[3])

	copy(b[16:], d.x[:])

	binary.BigEndian.PutUint64(b[80:], uint64(d.nx))

	binary.BigEndian.PutUint64(b[88:], d.len)

	return b, nil
}

func (d *digest) UnmarshalBinary(data []byte) error {
	if len(data) != marshaledDigestSize {
		return hash.ErrMarshalState
	}

	d.s[0] = binary.BigEndian.Uint32(data)
	d.s[1] = binary.BigEndian.Uint32(data[4:])
	d.s[2] = binary.BigEndian.Uint32(data[8:])
	d.s[3] = binary.BigEndian.Uint32(data[12:])

	copy(d.x[:], data[16:])

	d.nx = int(binary.BigEndian.Uint64(data[80:]))

	d.len = binary.BigEndian.Uint64(data[88:])

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
