// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecdh

import (
	"crypto/internal/nistec"
	"crypto/internal/randutil"
	"errors"
	"io"
)

type nistCurve[Point nistPoint[Point]] struct {
	newPoint   func() Point
	scalarSize int
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
	key := make([]byte, 66)[:c.scalarSize]
	randutil.MaybeReadByte(rand)
	for {
		if _, err := io.ReadFull(rand, key); err != nil {
			return nil, err
		}

		// Mask off any excess bits if the size of the underlying field is not a
		// whole number of bytes.
		// key[0] &= mask[bitSize%8] // TODO

		// In tests, rand will return all zeros and we don't want to return a
		// private key that will always fail ECDH (and which has the point at
		// infinity as its public key). This also makes this function consistent
		// with crypto/elliptic.GenerateKey.
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
	if len(key) != c.scalarSize {
		return nil, errors.New("crypto/ecdh: invalid private key size")
	}
	if !c.checkScalar(key) {
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

func (c *nistCurve[Point]) checkScalar(s []byte) bool {
	return true // TODO, also check for zero
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
	newPoint:   nistec.NewP256Point,
	scalarSize: 32,
}

// P384 returns a Curve which implements NIST P-384 (FIPS 186-3, section D.2.4),
// also known as secp384r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P384() Curve { return p384 }

var p384 = &nistCurve[*nistec.P384Point]{
	newPoint:   nistec.NewP384Point,
	scalarSize: 48,
}

// P521 returns a Curve which implements NIST P-521 (FIPS 186-3, section D.2.5),
// also known as secp521r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P521() Curve { return p521 }

var p521 = &nistCurve[*nistec.P521Point]{
	newPoint:   nistec.NewP521Point,
	scalarSize: 66,
}
