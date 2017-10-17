// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm64

package aes

import (
	"crypto/cipher"
	"crypto/subtle"
	"errors"
)

//go:noescape
func gcmHash(y, productTable *byte, input []byte)
//go:noescape
func gcmAesCtrEncAsm(out, plaintext *byte, datalen uint32, counter *byte, ks *uint32, nk int)
//go:noescape
func gcmMul(dst, a, b *byte)
//go:noescape
func gcmConvertBytes(out, in *byte)
//go:noescape
func gcmMergeLen(out *byte, len0, len1 uint64)

const (
	gcmCounterSize       = 16
	gcmBlockSize         = 16
	gcmTagSize           = 16
	gcmStandardNonceSize = 12
	gcmMinimumTagSize    = 12
)

var errOpen = errors.New("cipher: message authentication failed")

// aesCipherGCM implements crypto/cipher.gcmAble so that crypto/cipher.NewGCM
// will use the optimised implementation in this file when possible. Instances
// of this type only exist when supportsAesGcm() returns true.
type aesCipherGCM struct {
	aesCipherAsm
}

// Assert that aesCipherGCM implements the gcmAble interface.
var _ gcmAble = (*aesCipherGCM)(nil)

// NewGCM returns the AES cipher wrapped in Galois Counter Mode. This is only
// called by crypto/cipher.NewGCM via the gcmAble interface.
func (c *aesCipherGCM) NewGCM(nonceSize, tagSize int) (cipher.AEAD, error) {
	g := &gcmAsm{ks: c.enc, nonceSize: nonceSize, tagSize: tagSize}
	h := make([]byte, gcmBlockSize)
	c.Encrypt(h, h)
	gcmConvertBytes(&h[0], &h[0])
	// pre-compute  table H^0 -- H^8
	gcmAesInit(&g.productTable, h)
	return g, nil
}

type gcmAsm struct {
	// ks is the key schedule, the length of which depends on the size of
	// the AES key.
	ks []uint32
	// productTable contains pre-computed multiples of the binary-field
	// element used in GHASH.
	productTable [128]byte
	// nonceSize contains the expected size of the nonce, in bytes.
	nonceSize int
	// tagSize contains the size of the tag, in bytes.
	tagSize int
}

func (g *gcmAsm) NonceSize() int {
	return g.nonceSize
}

func (g *gcmAsm) Overhead() int {
	return g.tagSize
}

// sliceForAppend takes a slice and a requested number of bytes. It returns a
// slice with the contents of the given slice followed by that many bytes and a
// second slice that aliases into it and contains only the extra bytes. If the
// original slice has sufficient capacity then no allocation is performed.
func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}

// Seal encrypts and authenticates plaintext. See the cipher.AEAD interface for
// details.
func (g *gcmAsm) Seal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != g.nonceSize {
		panic("cipher: incorrect nonce length given to GCM")
	}
	if uint64(len(plaintext)) > ((1<<32)-2)*BlockSize {
		panic("cipher: message too large for GCM")
	}

	counter := make([]byte, gcmCounterSize)

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmHashData(g.productTable, nonce, counter)

		lens := make([]byte, gcmCounterSize)
		gcmMergeLen(&lens[0], 0, uint64(len(nonce))*8)

		gcmHashData(g.productTable, lens[:], counter)
		gcmConvertBytes(&counter[0], &counter[0])
	}

	// record orginal counter
	initCounter := make([]byte, gcmCounterSize)
	copy(initCounter[:], counter[:])

	gcmInc32(counter[:], 1, gcmCounterSize)

	tagOut := make([]byte, gcmTagSize)
	gcmHashData(g.productTable, data, tagOut)

	ret, out := sliceForAppend(dst, len(plaintext))
	if len(plaintext) > 0 {
		gcmAesEnc(g.productTable, out, plaintext, counter, tagOut, g.ks)
	}

	ret, out = sliceForAppend(out, g.tagSize)
	gcmAesFinish(g.productTable, tagOut, uint64(len(plaintext)), uint64(len(data)), initCounter, g.ks)
	copy(out[:], tagOut[:])

	return ret
}

// Open authenticates and decrypts ciphertext. See the cipher.AEAD interface
// for details.
func (g *gcmAsm) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(nonce) != g.nonceSize {
		panic("cipher: incorrect nonce length given to GCM")
	}

	// Sanity check to prevent the authentication from always succeeding if an implementation
	// leaves tagSize uninitialized, for example.
	if g.tagSize < gcmMinimumTagSize {
		panic("cipher: incorrect GCM tag size")
	}

	if len(ciphertext) < g.tagSize {
		return nil, errOpen
	}
	if uint64(len(ciphertext)) > ((1<<32)-2)*BlockSize+uint64(g.tagSize) {
		return nil, errOpen
	}

	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]

	counter := make([]byte, gcmCounterSize)

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmHashData(g.productTable, nonce, counter)

		lens := make([]byte, gcmCounterSize)
		gcmMergeLen(&lens[0], 0, uint64(len(nonce))*8)

		gcmHashData(g.productTable, lens[:], counter)
		gcmConvertBytes(&counter[0], &counter[0])
	}

	initCounter := make([]byte, gcmCounterSize)
	copy(initCounter[:], counter[:])

	gcmInc32(counter[:], 1, gcmCounterSize)

	expectedTag := make([]byte, gcmTagSize)
	gcmHashData(g.productTable, data, expectedTag)

	ret, out := sliceForAppend(dst, len(ciphertext))
	if len(ciphertext) > 0 {
		gcmAesDec(g.productTable, out, ciphertext, counter, expectedTag, g.ks)
	}
	gcmAesFinish(g.productTable, expectedTag, uint64(len(ciphertext)), uint64(len(data)), initCounter, g.ks)

	if subtle.ConstantTimeCompare(expectedTag[:g.tagSize], tag) != 1 {
		for i := range out {
			out[i] = 0
		}
		return nil, errOpen
	}

	return ret, nil
}

func gcmAesFinish(productTable [128]byte, tagOut []byte, clen, alen uint64, counter []byte, ks []uint32) {
	t := make([]byte, gcmCounterSize)
	gcmMergeLen(&t[0], alen*8, clen*8)
	gcmHash(&tagOut[0], &productTable[0], t)

	gcmConvertBytes(&tagOut[0], &tagOut[0])
	out := make([]byte, 16)
	gcmAesCtrEnc(out, tagOut, counter, ks)

	copy(tagOut[:], out[:])

	return
}

func gcmAesInit(productTable *[128]byte, h []byte) {
	copy((*productTable)[:16], h)

	for i := 1; i < 8; i++ {
		gcmMul(&((*productTable)[i*16]), &((*productTable)[(i - 1)*16]), &((*productTable)[0]))
	}
	return
}

func gcmHashData(productTable [128]byte, data []byte, tagOut []byte) {
	fullBlocks := (len(data) >> 4) << 4
	gcmHash(&tagOut[0], &productTable[0], data[:fullBlocks])

	if len(data) != fullBlocks {
		partialBlock := make([]byte, 16)
		copy(partialBlock, data[fullBlocks:])
		gcmHash(&tagOut[0], &productTable[0], partialBlock)
	}
	return
}

func gcmAesEnc(productTable [128]byte, out []byte, plaintext []byte, counter []byte, tagout []byte, ks []uint32) {
	gcmAesCtrEnc(out, plaintext, counter, ks)
	gcmHashData(productTable, out, tagout)
	return
}

func gcmAesDec(productTable [128]byte, out []byte, ciphertext []byte, counter []byte, tagout []byte, ks []uint32) {
	gcmHashData(productTable, ciphertext, tagout)
	gcmAesCtrEnc(out, ciphertext, counter, ks)
	return
}

func gcmAesCtrEnc(out []byte, plaintext []byte, counter []byte, ks []uint32) {
	gcmAesCtrEncAsm(&out[0], &plaintext[0], uint32(len(plaintext)), &counter[0], &ks[0], len(ks)/4-1)
	return
}

func gcmInc32(a []byte, step uint, length uint) {
	var d uint32

	if length < 4 {
		return
	}

	lastuint32 := a[length-4:]
	d = uint32(lastuint32[3]) | uint32(lastuint32[2])<<8 | uint32(lastuint32[1])<<16 | uint32(lastuint32[0])<<24
	d += uint32(step)

	lastuint32[3] = byte(d)
	lastuint32[2] = byte(d >> 8)
	lastuint32[1] = byte(d >> 16)
	lastuint32[0] = byte(d >> 24)
	return
}
