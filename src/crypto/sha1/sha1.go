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
	"crypto/subtle"
	"encoding/binary"
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

// The base 2 logarithm of BlockSize for right-shifting.
const blockSizeLog2 = 6

const (
	chunk = 64
	init0 = 0x67452301
	init1 = 0xEFCDAB89
	init2 = 0x98BADCFE
	init3 = 0x10325476
	init4 = 0xC3D2E1F0
)

// digest represents the partial evaluation of a checksum.
type digest struct {
	h   [5]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

const (
	magic         = "sha\x01"
	marshaledSize = len(magic) + 5*4 + chunk + 8
)

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, marshaledSize)
	b = append(b, magic...)
	b = appendUint32(b, d.h[0])
	b = appendUint32(b, d.h[1])
	b = appendUint32(b, d.h[2])
	b = appendUint32(b, d.h[3])
	b = appendUint32(b, d.h[4])
	b = append(b, d.x[:d.nx]...)
	b = b[:len(b)+len(d.x)-int(d.nx)] // already zero
	b = appendUint64(b, d.len)
	return b, nil
}

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic) || string(b[:len(magic)]) != magic {
		return errors.New("crypto/sha1: invalid hash state identifier")
	}
	if len(b) != marshaledSize {
		return errors.New("crypto/sha1: invalid hash state size")
	}
	b = b[len(magic):]
	b, d.h[0] = consumeUint32(b)
	b, d.h[1] = consumeUint32(b)
	b, d.h[2] = consumeUint32(b)
	b, d.h[3] = consumeUint32(b)
	b, d.h[4] = consumeUint32(b)
	b = b[copy(d.x[:], b):]
	b, d.len = consumeUint64(b)
	d.nx = int(d.len % chunk)
	return nil
}

func appendUint64(b []byte, x uint64) []byte {
	var a [8]byte
	putUint64(a[:], x)
	return append(b, a[:]...)
}

func appendUint32(b []byte, x uint32) []byte {
	var a [4]byte
	putUint32(a[:], x)
	return append(b, a[:]...)
}

func consumeUint64(b []byte) ([]byte, uint64) {
	_ = b[7]
	x := uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
	return b[8:], x
}

func consumeUint32(b []byte) ([]byte, uint32) {
	_ = b[3]
	x := uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	return b[4:], x
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

// New returns a new hash.Hash computing the SHA1 checksum. The Hash also
// implements encoding.BinaryMarshaler and encoding.BinaryUnmarshaler to
// marshal and unmarshal the internal state of the hash.
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

func (d *digest) Sum(in []byte) []byte {
	// Make a copy of d so that caller can keep writing and summing.
	d0 := *d
	hash := d0.checkSum()
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

// ConstantTimeSumWithData computes the same result as Write(data[:l]) followed
// by Sum(in) but does not modify the digest state. It treats l as secret and
// len(data) as public. The output is undefined if l < 0, l > len(data), or
// len(data) >= 2**31.
func (d *digest) ConstantTimeSumWithData(in, data []byte, l int) []byte {
	d0 := *d
	hash := d0.constSumWithData(data, l)
	return append(in, hash[:]...)
}

func (d *digest) constSumWithData(data []byte, l int) [Size]byte {
	if chunk != BlockSize {
		// This logic assumes that chunk and block sizes match. If this
		// ever changes, this logic should check for d.nx >= BlockSize
		// and, if so, process d.x down to a partial block. (d.nx is
		// public, so this can be done without timing precautions.)
		panic("sha1: code currently assumes the chunk and block sizes are the same")
	}

	// The final SHA-1 hash is determined by the hash state (d.h) after
	// incorporating the following, divided into blocks:
	//
	// - d.x[:d.nx], the pending partial block.
	// - data[:l]
	// - a single byte 0x80
	// - however many zero bytes are needed for the length below to end on
	//   a block boundary
	// - the total length in bits, encoded as an 8-byte big-endian integer
	//
	// We must do so without leaking l. Note d.nx is public as it was
	// computed by Write().

	// First, compute and assemble the encoded length, in bits.
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], (uint64(l)+d.len)<<3)

	// Compute the number of blocks we actually wish to process. Division
	// is variable-time, so we count blocks with blockSizeLog2. (The
	// compiler will likely optimize division to a shift, but this avoids
	// depending on compiler optimizations.)
	numBlocks := d.nx + l + 1 + 8
	numBlocks = (numBlocks + BlockSize - 1) >> blockSizeLog2

	// numBlocks is secret due to its use of l. It is publicly upper
	// bounded by len(data), so compute the public upper bound on the
	// number of blocks. We must process this many blocks, even if we
	// ignore most of them.
	maxBlocks := d.nx + len(data) + 1 + 8
	maxBlocks = (maxBlocks + BlockSize - 1) >> blockSizeLog2

	// Assemble and process maxBlocks blocks. For i < numBlocks, we
	// assemble the correct blocks in constant-time. Blocks beyond that may
	// have any arbitrary value (this code uses all zeros).
	var digest [Size]byte
	var dataIdx int
	for i := 0; i < maxBlocks; i++ {
		isLastBlock := subtle.ConstantTimeEq(int32(i), int32(numBlocks-1))
		isLastBlockMask := uint8(^(isLastBlock - 1))
		var start int
		if i == 0 {
			// For the first block, skip over the existing partial
			// block.
			start = d.nx
		}
		if dataIdx < len(data) {
			// Fill the block with data, including data past l. The
			// subsequent loop will zero bytes to discard.
			copy(d.x[start:BlockSize], data[dataIdx:])
		}
		for j := start; j < BlockSize; j++ {
			// Zero the byte of data if we are past l.
			isPastLastByte := subtle.ConstantTimeLessOrEq(l, dataIdx)
			d.x[j] &= uint8(isPastLastByte - 1)

			// The byte immediately after data[:l] is 0x80.
			isPadByte := subtle.ConstantTimeEq(int32(l), int32(dataIdx))
			d.x[j] |= 0x80 & uint8(^(isPadByte - 1))

			// The final block's last 8 bytes are the length.
			if j >= BlockSize-8 {
				d.x[j] |= length[j-(BlockSize-8)] & isLastBlockMask
			}

			dataIdx++
		}

		// Compress and only keep the digest if this is the last block.
		block(d, d.x[:BlockSize])
		for i, s := range d.h {
			digest[i*4] |= isLastBlockMask & byte(s>>24)
			digest[i*4+1] |= isLastBlockMask & byte(s>>16)
			digest[i*4+2] |= isLastBlockMask & byte(s>>8)
			digest[i*4+3] |= isLastBlockMask & byte(s)
		}
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
