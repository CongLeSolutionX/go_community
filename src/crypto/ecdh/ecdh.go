// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ecdh implements Elliptic Curve Diffie-Hellman over
// NIST curves and Curve25519.
package ecdh

import (
	"crypto"
	"crypto/internal/nistec"
	"crypto/internal/randutil"
	"crypto/subtle"
	"errors"
	"io"

	"golang.org/x/crypto/curve25519"
)

type Curve interface {
	// ECDH performs a ECDH exchange and returns the shared secret.
	//
	// For NIST curves, this performs ECDH as specified in SEC 1, Version 2.0,
	// Section 3.3.1, and returns the x-coordinate encoded according to SEC 1,
	// Version 2.0, Section 2.3.5. In particular, if the result is the point at
	// infinity, ECDH returns an error. (Note that for NIST curves, that's only
	// possible if the private key is the all-zero value.)
	//
	// For X25519, this performs ECDH as specified in RFC 7748, Section 6.1. If
	// the result is the all-zero value, ECDH returns an error.
	ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error)

	// GenerateKey generates a new PrivateKey from rand.
	GenerateKey(rand io.Reader) (*PrivateKey, error)

	// NewPrivateKey checks that key is valid and returns a PrivateKey.
	//
	// For NIST curves, this follows SEC 1, Version 2.0, Section 2.3.6, which
	// amounts to decoding the bytes as a fixed length big endian integer and
	// checking that it is lower than the order of the curve.
	//
	// For X25519, this only checks the scalar length.
	NewPrivateKey(key []byte) (*PrivateKey, error)

	// NewPublicKey checks that key is valid and returns a PublicKey.
	//
	// For NIST curves, this decodes an uncompressed point according to SEC 1,
	// Version 2.0, Section 2.3.4. Compressed encodings and the point at
	// infinity are rejected.
	//
	// For X25519, this only checks the u-coordinate length. Some public keys
	// will cause ECDH to return an error.
	NewPublicKey(key []byte) (*PublicKey, error)

	// private is a private method to allow us to expand the ECDH interface with
	// more methods in the future without breaking backwards compatibility.
	private()
}

const (
	p256PublicKeySize    = 1 + 2*32
	p256PrivateKeySize   = 32
	p256SharedSecretSize = 32

	p384PublicKeySize    = 1 + 2*48
	p384PrivateKeySize   = 48
	p384SharedSecretSize = 48

	p521PublicKeySize    = 1 + 2*66
	p521PrivateKeySize   = 66
	p521SharedSecretSize = 66

	x25519PublicKeySize    = curve25519.PointSize
	x25519PrivateKeySize   = curve25519.ScalarSize
	x25519SharedSecretSize = curve25519.PointSize
)

// PublicKey is an ECDH public key, usually a peer's ECDH share sent over the wire.
type PublicKey struct {
	curve     Curve
	publicKey []byte

	// buf is the backing array of publicKey, to prevent extra heap allocations.
	// Its size is sufficient to hold any public key supported by this package.
	buf [p521PublicKeySize]byte
}

// Bytes returns a copy of the encoding of the public key.
func (k *PublicKey) Bytes() []byte {
	// Copy the public key to a fixed size buffer that can get allocated on the
	// caller's stack after inlining.
	var buf [p521PublicKeySize]byte
	return append(buf[:0], k.publicKey...)
}

func (k *PublicKey) Equal(x crypto.PublicKey) bool {
	xx, ok := x.(*PublicKey)
	if !ok {
		return false
	}
	if k.curve != xx.curve {
		return false
	}
	return subtle.ConstantTimeCompare(k.publicKey, xx.publicKey) == 1
}

func (k *PublicKey) Curve() Curve {
	return k.curve
}

// PrivateKey is an ECDH private key, usually kept secret.
type PrivateKey struct {
	PublicKey
	privateKey []byte

	// buf is the backing array of privateKey, to prevent extra heap allocations.
	// Its size is sufficient to hold any private key supported by this package.
	buf [p521PrivateKeySize]byte
}

// Bytes returns a copy of the encoding of the private key.
func (k *PrivateKey) Bytes() []byte {
	// Copy the private key to a fixed size buffer that can get allocated on the
	// caller's stack after inlining.
	var buf [p521PrivateKeySize]byte
	return append(buf[:0], k.privateKey...)
}

func (k *PrivateKey) Equal(x crypto.PrivateKey) bool {
	xx, ok := x.(*PrivateKey)
	if !ok {
		return false
	}
	if !k.PublicKey.Equal(&xx.PublicKey) {
		return false
	}
	return subtle.ConstantTimeCompare(k.privateKey, xx.privateKey) == 1
}

func (k *PrivateKey) Public() crypto.PublicKey {
	return &k.PublicKey
}

// P256 returns a Curve which implements NIST P-256 (FIPS 186-3, section D.2.3),
// also known as secp256r1 or prime256v1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P256() Curve { return p256 }

var p256 = &p256Curve{}

type p256Curve struct{}

func (p256Curve) private() {}

func (c *p256Curve) GenerateKey(rand io.Reader) (*PrivateKey, error) {
	k := &PrivateKey{}
	k.curve = c
	k.privateKey = k.buf[:p256PrivateKeySize]
	k.publicKey = k.PublicKey.buf[:p256PublicKeySize]

	randutil.MaybeReadByte(rand)
	for {
		if _, err := io.ReadFull(rand, k.privateKey); err != nil {
			return nil, err
		}
		if p256CheckScalar(k.privateKey) {
			break
		}
	}

	p, err := nistec.NewP256Point().ScalarBaseMult(k.privateKey)
	if err != nil {
		return nil, err
	}

	copy(k.publicKey, p.Bytes())
	return k, nil
}

func (c *p256Curve) NewPrivateKey(key []byte) (*PrivateKey, error) {
	panic("unimplemented") // TODO
}

func p256CheckScalar(s []byte) bool {
	return true // TODO
}

func (c *p256Curve) NewPublicKey(key []byte) (*PublicKey, error) {
	// Reject the point at infinity and compressed encodings.
	if len(key) == 0 || key[0] != 4 {
		return nil, errors.New("crypto/ecdh: invalid public key")
	}
	if _, err := nistec.NewP256Point().SetBytes(key); err != nil {
		return nil, err
	}

	k := &PublicKey{curve: c}
	k.publicKey = append(k.buf[:0], key...)
	return k, nil
}

func (c *p256Curve) ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error) {
	p, err := nistec.NewP256Point().SetBytes(remote.publicKey)
	if err != nil {
		return nil, err
	}
	if _, err := p.ScalarMult(p, local.privateKey); err != nil {
		return nil, err
	}
	return p.BytesX()
}

// P384 returns a Curve which implements NIST P-384 (FIPS 186-3, section D.2.4),
// also known as secp384r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P384() Curve { return p384 }

var p384 = &p384Curve{}

type p384Curve struct{}

func (p384Curve) private() {}

func (c *p384Curve) GenerateKey(rand io.Reader) (*PrivateKey, error) {
	k := &PrivateKey{}
	k.curve = c
	k.privateKey = k.buf[:p384PrivateKeySize]
	k.publicKey = k.PublicKey.buf[:p384PublicKeySize]

	randutil.MaybeReadByte(rand)
	for {
		if _, err := io.ReadFull(rand, k.privateKey); err != nil {
			return nil, err
		}
		if p384CheckScalar(k.privateKey) {
			break
		}
	}

	p, err := nistec.NewP384Point().ScalarBaseMult(k.privateKey)
	if err != nil {
		return nil, err
	}

	copy(k.publicKey, p.Bytes())
	return k, nil
}

func (c *p384Curve) NewPrivateKey(key []byte) (*PrivateKey, error) {
	panic("unimplemented") // TODO
}

func p384CheckScalar(s []byte) bool {
	return true // TODO
}

func (c *p384Curve) NewPublicKey(key []byte) (*PublicKey, error) {
	// Reject the point at infinity and compressed encodings.
	if len(key) == 0 || key[0] != 4 {
		return nil, errors.New("crypto/ecdh: invalid public key")
	}
	if _, err := nistec.NewP384Point().SetBytes(key); err != nil {
		return nil, err
	}

	k := &PublicKey{curve: c}
	k.publicKey = append(k.buf[:0], key...)
	return k, nil
}

func (c *p384Curve) ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error) {
	p, err := nistec.NewP384Point().SetBytes(remote.publicKey)
	if err != nil {
		return nil, err
	}
	if _, err := p.ScalarMult(p, local.privateKey); err != nil {
		return nil, err
	}
	return p.BytesX()
}

// P521 returns a Curve which implements NIST P-521 (FIPS 186-3, section D.2.5),
// also known as secp521r1.
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func P521() Curve { return p521 }

var p521 = &p521Curve{}

type p521Curve struct{}

func (p521Curve) private() {}

func (c *p521Curve) GenerateKey(rand io.Reader) (*PrivateKey, error) {
	k := &PrivateKey{}
	k.curve = c
	k.privateKey = k.buf[:p521PrivateKeySize]
	k.publicKey = k.PublicKey.buf[:p521PublicKeySize]

	randutil.MaybeReadByte(rand)
	for {
		if _, err := io.ReadFull(rand, k.privateKey); err != nil {
			return nil, err
		}
		if p521CheckScalar(k.privateKey) {
			break
		}
	}

	p, err := nistec.NewP521Point().ScalarBaseMult(k.privateKey)
	if err != nil {
		return nil, err
	}

	copy(k.publicKey, p.Bytes())
	return k, nil
}

func (c *p521Curve) NewPrivateKey(key []byte) (*PrivateKey, error) {
	panic("unimplemented") // TODO
}

func p521CheckScalar(s []byte) bool {
	return true // TODO
}

func (c *p521Curve) NewPublicKey(key []byte) (*PublicKey, error) {
	// Reject the point at infinity and compressed encodings.
	if len(key) == 0 || key[0] != 4 {
		return nil, errors.New("crypto/ecdh: invalid public key")
	}
	if _, err := nistec.NewP521Point().SetBytes(key); err != nil {
		return nil, err
	}

	k := &PublicKey{curve: c}
	k.publicKey = append(k.buf[:0], key...)
	return k, nil
}

func (c *p521Curve) ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error) {
	p, err := nistec.NewP521Point().SetBytes(remote.publicKey)
	if err != nil {
		return nil, err
	}
	if _, err := p.ScalarMult(p, local.privateKey); err != nil {
		return nil, err
	}
	return p.BytesX()
}

// X25519 returns a Curve which implements the X25519 function over Curve25519
// (RFC 7748, Section 5).
//
// Multiple invocations of this function will return the same value, so it can
// be used for equality checks and switch statements.
func X25519() Curve { return x25519 }

var x25519 = &x25519Curve{}

type x25519Curve struct{}

func (x25519Curve) private() {}

func (c *x25519Curve) GenerateKey(rand io.Reader) (*PrivateKey, error) {
	k := &PrivateKey{}
	k.curve = c
	k.privateKey = k.buf[:x25519PrivateKeySize]
	k.publicKey = k.PublicKey.buf[:x25519PublicKeySize]

	randutil.MaybeReadByte(rand)
	if _, err := io.ReadFull(rand, k.privateKey); err != nil {
		return nil, err
	}

	publicKey, err := curve25519.X25519(k.privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}

	copy(k.publicKey, publicKey)
	return k, nil
}

func (c *x25519Curve) NewPrivateKey(key []byte) (*PrivateKey, error) {
	panic("unimplemented") // TODO
}

func (c *x25519Curve) NewPublicKey(key []byte) (*PublicKey, error) {
	if len(key) != x25519PublicKeySize {
		return nil, errors.New("crypto/ecdh: invalid public key")
	}

	k := &PublicKey{curve: c}
	k.publicKey = append(k.buf[:0], key...)
	return k, nil
}

func (c *x25519Curve) ECDH(local *PrivateKey, remote *PublicKey) ([]byte, error) {
	return curve25519.X25519(local.privateKey, remote.publicKey)
}
