// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gcm

import (
	"crypto/internal/fips/aes"
	"crypto/internal/fips/alias"
	"crypto/internal/fips/drbg"
	"crypto/internal/fips/subtle"
	"errors"
	"internal/byteorder"
)

// gcmFieldElement represents a value in GF(2¹²⁸). In order to reflect the GCM
// standard and make binary.BigEndian suitable for marshaling these values, the
// bits are stored in big endian order. For example:
//
//	the coefficient of x⁰ can be obtained by v.low >> 63.
//	the coefficient of x⁶³ can be obtained by v.low & 1.
//	the coefficient of x⁶⁴ can be obtained by v.high >> 63.
//	the coefficient of x¹²⁷ can be obtained by v.high & 1.
type gcmFieldElement struct {
	low, high uint64
}

// Block is a subset of cipher.Block to avoid a circular dependency.
type Block interface {
	BlockSize() int
	Encrypt(dst, src []byte)
}

// GCM represents a Galois Counter Mode with a specific key. See
// https://csrc.nist.gov/groups/ST/toolkit/BCM/documents/proposedmodes/GCM/GCM-revised-spec.pdf
type GCM struct {
	cipher    Block
	nonceSize int
	tagSize   int
	// productTable contains the first sixteen powers of the key, H.
	// However, they are in bit reversed order. See initGCMGeneric.
	//
	// TODO: the assembly implementations don't store the table in the same
	// format, in violation of AssemblyPolicy. They need to be reconciled or at
	// least documented.
	productTable [16]gcmFieldElement
}

func New(cipher Block, nonceSize, tagSize int) (*GCM, error) {
	if tagSize < gcmMinimumTagSize || tagSize > gcmBlockSize {
		return nil, errors.New("cipher: incorrect tag size given to GCM")
	}
	if nonceSize <= 0 {
		return nil, errors.New("cipher: the nonce can't have zero length")
	}
	if cipher.BlockSize() != gcmBlockSize {
		return nil, errors.New("cipher: NewGCM requires 128-bit block cipher")
	}
	g := &GCM{cipher: cipher, nonceSize: nonceSize, tagSize: tagSize}
	initGCM(g)
	return g, nil
}

func initGCMGeneric(g *GCM) {
	checkGenericIsExpected(g.cipher)

	var key [gcmBlockSize]byte
	g.cipher.Encrypt(key[:], key[:])

	// We precompute 16 multiples of |key|. However, when we do lookups
	// into this table we'll be using bits from a field element and
	// therefore the bits will be in the reverse order. So normally one
	// would expect, say, 4*key to be in index 4 of the table but due to
	// this bit ordering it will actually be in index 0010 (base 2) = 2.
	x := gcmFieldElement{
		byteorder.BeUint64(key[:8]),
		byteorder.BeUint64(key[8:]),
	}
	g.productTable[reverseBits(1)] = x

	for i := 2; i < 16; i += 2 {
		g.productTable[reverseBits(i)] = gcmDouble(&g.productTable[reverseBits(i/2)])
		g.productTable[reverseBits(i+1)] = gcmAdd(&g.productTable[reverseBits(i)], &x)
	}
}

const (
	gcmBlockSize         = 16
	gcmTagSize           = 16
	gcmMinimumTagSize    = 12 // NIST SP 800-38D recommends tags with 12 or more bytes.
	gcmStandardNonceSize = 12
)

func (g *GCM) NonceSize() int {
	return g.nonceSize
}

func (g *GCM) Overhead() int {
	return g.tagSize
}

func (g *GCM) Seal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	if uint64(len(plaintext)) > uint64((1<<32)-2)*gcmBlockSize {
		panic("crypto/cipher: message too large for GCM")
	}

	ret, out := sliceForAppend(dst, len(plaintext)+g.tagSize)
	if alias.InexactOverlap(out, plaintext) {
		panic("crypto/cipher: invalid buffer overlap of output and input")
	}
	if alias.AnyOverlap(out, data) {
		panic("crypto/cipher: invalid buffer overlap of output and additional data")
	}

	seal(out, g, nonce, plaintext, data)
	return ret
}

func sealGeneric(out []byte, g *GCM, nonce, plaintext, data []byte) {
	checkGenericIsExpected(g.cipher)

	var counter, tagMask [gcmBlockSize]byte
	g.deriveCounter(&counter, nonce)

	g.cipher.Encrypt(tagMask[:], counter[:])
	gcmInc32(&counter)

	g.counterCrypt(out, plaintext, &counter)

	var tag [gcmTagSize]byte
	g.auth(tag[:], out[:len(plaintext)], data, &tagMask)
	copy(out[len(plaintext):], tag[:])
}

var errOpen = errors.New("cipher: message authentication failed")

func (g *GCM) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	// Sanity check to prevent the authentication from always succeeding if an implementation
	// leaves tagSize uninitialized, for example.
	if g.tagSize < gcmMinimumTagSize {
		panic("crypto/cipher: incorrect GCM tag size")
	}

	if len(ciphertext) < g.tagSize {
		return nil, errOpen
	}
	if uint64(len(ciphertext)) > uint64((1<<32)-2)*gcmBlockSize+uint64(g.tagSize) {
		return nil, errOpen
	}

	ret, out := sliceForAppend(dst, len(ciphertext)-g.tagSize)
	if alias.InexactOverlap(out, ciphertext) {
		panic("crypto/cipher: invalid buffer overlap of output and input")
	}
	if alias.AnyOverlap(out, data) {
		panic("crypto/cipher: invalid buffer overlap of output and additional data")
	}

	if err := open(out, g, nonce, ciphertext, data); err != nil {
		// We decrypt and authenticate concurrently, so we overwrite dst in the
		// event of a tag mismatch. To be consistent across platforms and to
		// avoid releasing unauthenticated plaintext, we clear the buffer in the
		// event of an error.
		clear(out)
		return nil, err
	}
	return ret, nil
}

func openGeneric(out []byte, g *GCM, nonce, ciphertext, data []byte) error {
	checkGenericIsExpected(g.cipher)

	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]

	var counter, tagMask [gcmBlockSize]byte
	g.deriveCounter(&counter, nonce)

	g.cipher.Encrypt(tagMask[:], counter[:])
	gcmInc32(&counter)

	var expectedTag [gcmTagSize]byte
	g.auth(expectedTag[:], ciphertext, data, &tagMask)

	if subtle.ConstantTimeCompare(expectedTag[:g.tagSize], tag) != 1 {
		return errOpen
	}

	g.counterCrypt(out, ciphertext, &counter)

	return nil
}

// GCMWithRandomNonce is an AEAD that automatically generates random nonces and
// prepends them to the ciphertext. The nonce size exposed through the AEAD
// interface is zero, and the nonce size is folded into the overhead.
type GCMWithRandomNonce struct {
	g GCM
}

func NewWithRandomNonce(cipher *aes.Block) *GCMWithRandomNonce {
	g := &GCMWithRandomNonce{GCM{cipher: cipher, nonceSize: gcmStandardNonceSize, tagSize: gcmTagSize}}
	initGCM(&g.g)
	return g
}

func (g *GCMWithRandomNonce) NonceSize() int {
	return 0
}

func (g *GCMWithRandomNonce) Overhead() int {
	return gcmStandardNonceSize + gcmTagSize
}

// Seal appends a random nonce and the encryption of plaintext to dst.
// nonce must be empty.
func (g *GCMWithRandomNonce) Seal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != 0 {
		panic("crypto/cipher: non-empty nonce passed to GCMWithRandomNonce")
	}
	if uint64(len(plaintext)) > uint64((1<<32)-2)*gcmBlockSize {
		panic("crypto/cipher: message too large for GCM")
	}

	ret, out := sliceForAppend(dst, gcmStandardNonceSize+len(plaintext)+gcmTagSize)
	if alias.InexactOverlap(out, plaintext) {
		panic("crypto/cipher: invalid buffer overlap of output and input")
	}
	if alias.AnyOverlap(out, data) {
		panic("crypto/cipher: invalid buffer overlap of output and additional data")
	}

	// The AEAD interface allows using plaintext[:0] or ciphertext[:0] as dst.
	//
	// This is kind of a problem when trying to prepend or trim a nonce, because the
	// actual AES-CTR blocks end up overlapping but not exactly.
	//
	// In Open, we write the output *before* the input, so unless we do something
	// weird like working through a chunk of block backwards, it works out.
	//
	// In Seal, we could work through the input backwards or intentionally load
	// ahead before writing, but for now we just do a memmove if we detect overlap.
	//
	//     ┌───────────────────────────┬ ─ ─
	//     │PPPPPPPPPPPPPPPPPPPPPPPPPPP│    │
	//     └▽─────────────────────────▲┴ ─ ─
	//       ╲ Seal                    ╲
	//        ╲                    Open ╲
	//     ┌───▼─────────────────────────△──┐
	//     │NN|CCCCCCCCCCCCCCCCCCCCCCCCCCC|T│
	//     └────────────────────────────────┘
	//
	if alias.ExactOverlap(out, plaintext) {
		copy(out[gcmStandardNonceSize:], plaintext)
		plaintext = out[gcmStandardNonceSize : gcmStandardNonceSize+len(plaintext)]
	}

	drbg.Read(out[:gcmStandardNonceSize])
	seal(out[gcmStandardNonceSize:], &g.g, out[:gcmStandardNonceSize], plaintext, data)
	return ret
}

// Open extracts the nonce from the ciphertext and appends the plaintext to dst.
// nonce must be empty.
func (g *GCMWithRandomNonce) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(nonce) != 0 {
		panic("crypto/cipher: non-empty nonce passed to GCMWithRandomNonce")
	}

	if len(ciphertext) < gcmStandardNonceSize+gcmTagSize {
		return nil, errOpen
	}
	if uint64(len(ciphertext)) > gcmStandardNonceSize+((1<<32)-2)*gcmBlockSize+gcmTagSize {
		return nil, errOpen
	}

	ret, out := sliceForAppend(dst, len(ciphertext)-gcmStandardNonceSize-gcmTagSize)
	if alias.InexactOverlap(out, ciphertext) {
		panic("crypto/cipher: invalid buffer overlap of output and input")
	}
	if alias.AnyOverlap(out, data) {
		panic("crypto/cipher: invalid buffer overlap of output and additional data")
	}

	if err := open(out, &g.g, ciphertext[:gcmStandardNonceSize], ciphertext[gcmStandardNonceSize:], data); err != nil {
		// We decrypt and authenticate concurrently, so we overwrite dst in the
		// event of a tag mismatch. To be consistent across platforms and to
		// avoid releasing unauthenticated plaintext, we clear the buffer in the
		// event of an error.
		clear(out)
		return nil, err
	}
	return ret, nil
}

// reverseBits reverses the order of the bits of 4-bit number in i.
func reverseBits(i int) int {
	i = ((i << 2) & 0xc) | ((i >> 2) & 0x3)
	i = ((i << 1) & 0xa) | ((i >> 1) & 0x5)
	return i
}

// gcmAdd adds two elements of GF(2¹²⁸) and returns the sum.
func gcmAdd(x, y *gcmFieldElement) gcmFieldElement {
	// Addition in a characteristic 2 field is just XOR.
	return gcmFieldElement{x.low ^ y.low, x.high ^ y.high}
}

// gcmDouble returns the result of doubling an element of GF(2¹²⁸).
func gcmDouble(x *gcmFieldElement) (double gcmFieldElement) {
	msbSet := x.high&1 == 1

	// Because of the bit-ordering, doubling is actually a right shift.
	double.high = x.high >> 1
	double.high |= x.low << 63
	double.low = x.low >> 1

	// If the most-significant bit was set before shifting then it,
	// conceptually, becomes a term of x^128. This is greater than the
	// irreducible polynomial so the result has to be reduced. The
	// irreducible polynomial is 1+x+x^2+x^7+x^128. We can subtract that to
	// eliminate the term at x^128 which also means subtracting the other
	// four terms. In characteristic 2 fields, subtraction == addition ==
	// XOR.
	if msbSet {
		double.low ^= 0xe100000000000000
	}

	return
}

var gcmReductionTable = []uint16{
	0x0000, 0x1c20, 0x3840, 0x2460, 0x7080, 0x6ca0, 0x48c0, 0x54e0,
	0xe100, 0xfd20, 0xd940, 0xc560, 0x9180, 0x8da0, 0xa9c0, 0xb5e0,
}

// mul sets y to y*H, where H is the GCM key, fixed during NewGCMWithNonceSize.
func (g *GCM) mul(y *gcmFieldElement) {
	var z gcmFieldElement

	for i := 0; i < 2; i++ {
		word := y.high
		if i == 1 {
			word = y.low
		}

		// Multiplication works by multiplying z by 16 and adding in
		// one of the precomputed multiples of H.
		for j := 0; j < 64; j += 4 {
			msw := z.high & 0xf
			z.high >>= 4
			z.high |= z.low << 60
			z.low >>= 4
			z.low ^= uint64(gcmReductionTable[msw]) << 48

			// the values in |table| are ordered for
			// little-endian bit positions. See the comment
			// in NewGCMWithNonceSize.
			t := &g.productTable[word&0xf]

			z.low ^= t.low
			z.high ^= t.high
			word >>= 4
		}
	}

	*y = z
}

// updateBlocks extends y with more polynomial terms from blocks, based on
// Horner's rule. There must be a multiple of gcmBlockSize bytes in blocks.
func (g *GCM) updateBlocks(y *gcmFieldElement, blocks []byte) {
	for len(blocks) > 0 {
		y.low ^= byteorder.BeUint64(blocks)
		y.high ^= byteorder.BeUint64(blocks[8:])
		g.mul(y)
		blocks = blocks[gcmBlockSize:]
	}
}

// update extends y with more polynomial terms from data. If data is not a
// multiple of gcmBlockSize bytes long then the remainder is zero padded.
func (g *GCM) update(y *gcmFieldElement, data []byte) {
	fullBlocks := (len(data) >> 4) << 4
	g.updateBlocks(y, data[:fullBlocks])

	if len(data) != fullBlocks {
		var partialBlock [gcmBlockSize]byte
		copy(partialBlock[:], data[fullBlocks:])
		g.updateBlocks(y, partialBlock[:])
	}
}

// gcmInc32 treats the final four bytes of counterBlock as a big-endian value
// and increments it.
func gcmInc32(counterBlock *[16]byte) {
	ctr := counterBlock[len(counterBlock)-4:]
	byteorder.BePutUint32(ctr, byteorder.BeUint32(ctr)+1)
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

// counterCrypt crypts in to out using g.cipher in counter mode.
func (g *GCM) counterCrypt(out, in []byte, counter *[gcmBlockSize]byte) {
	var mask [gcmBlockSize]byte

	for len(in) >= gcmBlockSize {
		g.cipher.Encrypt(mask[:], counter[:])
		gcmInc32(counter)

		subtle.XORBytes(out, in, mask[:])
		out = out[gcmBlockSize:]
		in = in[gcmBlockSize:]
	}

	if len(in) > 0 {
		g.cipher.Encrypt(mask[:], counter[:])
		gcmInc32(counter)
		subtle.XORBytes(out, in, mask[:])
	}
}

// deriveCounter computes the initial GCM counter state from the given nonce.
// See NIST SP 800-38D, section 7.1. This assumes that counter is filled with
// zeros on entry.
func (g *GCM) deriveCounter(counter *[gcmBlockSize]byte, nonce []byte) {
	// GCM has two modes of operation with respect to the initial counter
	// state: a "fast path" for 96-bit (12-byte) nonces, and a "slow path"
	// for nonces of other lengths. For a 96-bit nonce, the nonce, along
	// with a four-byte big-endian counter starting at one, is used
	// directly as the starting counter. For other nonce sizes, the counter
	// is computed by passing it through the GHASH function.
	if len(nonce) == gcmStandardNonceSize {
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		var y gcmFieldElement
		g.update(&y, nonce)
		y.high ^= uint64(len(nonce)) * 8
		g.mul(&y)
		byteorder.BePutUint64(counter[:8], y.low)
		byteorder.BePutUint64(counter[8:], y.high)
	}
}

// auth calculates GHASH(ciphertext, additionalData), masks the result with
// tagMask and writes the result to out.
func (g *GCM) auth(out, ciphertext, additionalData []byte, tagMask *[gcmTagSize]byte) {
	var y gcmFieldElement
	g.update(&y, additionalData)
	g.update(&y, ciphertext)

	y.low ^= uint64(len(additionalData)) * 8
	y.high ^= uint64(len(ciphertext)) * 8

	g.mul(&y)

	byteorder.BePutUint64(out, y.low)
	byteorder.BePutUint64(out[8:], y.high)

	subtle.XORBytes(out, out, tagMask[:])
}
