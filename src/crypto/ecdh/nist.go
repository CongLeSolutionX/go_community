// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecdh

import (
	"crypto/internal/nistec"
	"crypto/internal/randutil"
	"encoding/binary"
	"errors"
	"io"
	"math/bits"
)

type nistCurve[Point nistPoint[Point]] struct {
	newPoint    func() Point
	scalarOrder []byte
}

// nistPoint is a generic constraint for the nistec Point types.
type nistPoint[T any] interface {
	Bytes() []byte
	BytesX() ([]byte, error)
	SetBytes([]byte) (T, error)
	ScalarMult(T, []byte) (T, error)
	ScalarBaseMult([]byte) (T, error)
}

// private implements the Curve interface.
func (nistCurve[Point]) private() {}

func (c *nistCurve[Point]) GenerateKey(rand io.Reader) (*PrivateKey, error) {
	key := make([]byte, 66)[:len(c.scalarOrder)]
	randutil.MaybeReadByte(rand)
	for {
		if _, err := io.ReadFull(rand, key); err != nil {
			return nil, err
		}

		// Mask off any excess bits if the size of the underlying field is not a
		// whole number of bytes, which is only the case for P-521. We use a
		// pointer to the scalarOrder field because comparing generic and
		// instantiated types is not supported.
		if &c.scalarOrder[0] == &p521Order[0] {
			key[0] &= 0b0000_0001
		}

		// In tests, rand will return all zeros and NewPrivateKey will reject
		// the zero key as it generates the identity as a public key. This also
		// makes this function consistent with crypto/elliptic.GenerateKey.
		key[1] ^= 0x42

		k, err := c.NewPrivateKey(key)
		if err == errInvalidPrivateKey {
			continue
		}
		return k, err
	}
}

var errInvalidPrivateKey = errors.New("crypto/ecdh: invalid private key")

func (c *nistCurve[Point]) NewPrivateKey(key []byte) (*PrivateKey, error) {
	if len(key) != len(c.scalarOrder) {
		return nil, errors.New("crypto/ecdh: invalid private key size")
	}
	if isZero(key) || !isLess(key, c.scalarOrder) {
		return nil, errInvalidPrivateKey
	}

	p, err := c.newPoint().ScalarBaseMult(key)
	if err != nil {
		return nil, err
	}

	k := &PrivateKey{}
	k.curve = c
	k.privateKey = append(k.buf[:0], key...)
	k.publicKey = append(k.PublicKey.buf[:0], p.Bytes()...)

	return k, nil
}

// isZero returns whether a is all zeroes in constant time.
func isZero(a []byte) bool {
	var acc byte
	for _, b := range a {
		acc |= b
	}
	return acc == 0
}

// isLess returns whether a < b, where a and b are big-endian buffers of the
// same length and shorter than 72 bytes.
func isLess(a, b []byte) bool {
	if len(a) != len(b) {
		panic("crypto/ecdh: internal error: mismatched isLess inputs")
	}

	// Copy the values into a fixed-size preallocated little-endian buffer.
	// 72 bytes is enough for every scalar in this package, and having a fixed
	// size lets us avoid heap allocations.
	if len(a) > 72 {
		panic("crypto/ecdh: internal error: isLess input too large")
	}
	bufA, bufB := make([]byte, 72), make([]byte, 72)
	for i := range a {
		bufA[i], bufB[i] = a[len(a)-i-1], b[len(b)-i-1]
	}

	// Perform a subtraction with borrow.
	var borrow uint64
	for i := 0; i < len(bufA); i += 8 {
		limbA, limbB := binary.LittleEndian.Uint64(bufA[i:]), binary.LittleEndian.Uint64(bufB[i:])
		_, borrow = bits.Sub64(limbA, limbB, borrow)
	}

	// If there is a borrow at the end of the operation, then a < b.
	return borrow == 1
}

func (c *nistCurve[Point]) NewPublicKey(key []byte) (*PublicKey, error) {
	// Reject the point at infinity and compressed encodings.
	if len(key) == 0 || key[0] != 4 {
		return nil, errors.New("crypto/ecdh: invalid public key")
	}
	if _, err := c.newPoint().SetBytes(key); err != nil {
		return nil, err
	}

	k := &PublicKey{curve: c}
	k.publicKey = append(k.buf[:0], key...)
	return k, nil
}

func (c *nistCurve[Point]) ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error) {
	p, err := c.newPoint().SetBytes(remote.publicKey)
	if err != nil {
		return nil, err
	}
	if _, err := p.ScalarMult(p, local.privateKey); err != nil {
		return nil, err
	}
	// BytesX will return an error if p is the point at infinity.
	return p.BytesX()
}

// P256 returns a Curve which implements NIST P-256 (FIPS 186-3, section D.2.3),
// also known as secp256r1 or prime256v1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P256() Curve { return p256 }

var p256 = &nistCurve[*nistec.P256Point]{
	newPoint:    nistec.NewP256Point,
	scalarOrder: p256Order,
}

var p256Order = []byte{0xff, 0xff, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xbc, 0xe6, 0xfa, 0xad, 0xa7, 0x17, 0x9e, 0x84, 0xf3, 0xb9, 0xca, 0xc2, 0xfc, 0x63, 0x25, 0x51}

// P384 returns a Curve which implements NIST P-384 (FIPS 186-3, section D.2.4),
// also known as secp384r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P384() Curve { return p384 }

var p384 = &nistCurve[*nistec.P384Point]{
	newPoint:    nistec.NewP384Point,
	scalarOrder: p384Order,
}

var p384Order = []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xc7, 0x63, 0x4d, 0x81, 0xf4, 0x37, 0x2d, 0xdf, 0x58, 0x1a, 0xd, 0xb2, 0x48, 0xb0, 0xa7, 0x7a, 0xec, 0xec, 0x19, 0x6a, 0xcc, 0xc5, 0x29, 0x73}

// P521 returns a Curve which implements NIST P-521 (FIPS 186-3, section D.2.5),
// also known as secp521r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P521() Curve { return p521 }

var p521 = &nistCurve[*nistec.P521Point]{
	newPoint:    nistec.NewP521Point,
	scalarOrder: p521Order,
}

var p521Order = []byte{0x1, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfa, 0x51, 0x86, 0x87, 0x83, 0xbf, 0x2f, 0x96, 0x6b, 0x7f, 0xcc, 0x1, 0x48, 0xf7, 0x9, 0xa5, 0xd0, 0x3b, 0xb5, 0xc9, 0xb8, 0x89, 0x9c, 0x47, 0xae, 0xbb, 0x6f, 0xb7, 0x1e, 0x91, 0x38, 0x64, 0x9}
