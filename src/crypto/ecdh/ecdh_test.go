// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ecdh_test

import (
	"bytes"
	"crypto"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	"golang.org/x/crypto/chacha20"
)

// Check that PublicKey and PrivateKey implement the interfaces documented in
// crypto.PublicKey and crypto.PrivateKey.
var _ interface {
	Equal(x crypto.PublicKey) bool
} = &ecdh.PublicKey{}
var _ interface {
	Public() crypto.PublicKey
	Equal(x crypto.PrivateKey) bool
} = &ecdh.PrivateKey{}

func TestECDH(t *testing.T) {
	testAllCurves(t, func(t *testing.T, curve ecdh.Curve) {
		aliceKey, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		bobKey, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}

		alicePubKey, err := curve.NewPublicKey(aliceKey.PublicKey.Bytes())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(aliceKey.PublicKey.Bytes(), alicePubKey.Bytes()) {
			t.Error("encoded and decoded public keys are different")
		}
		if !aliceKey.PublicKey.Equal(alicePubKey) {
			t.Error("encoded and decoded public keys are different")
		}

		alicePrivKey, err := curve.NewPrivateKey(aliceKey.Bytes())
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(aliceKey.Bytes(), alicePrivKey.Bytes()) {
			t.Error("encoded and decoded private keys are different")
		}
		if !aliceKey.Equal(alicePrivKey) {
			t.Error("encoded and decoded private keys are different")
		}

		bobSecret, err := curve.ECDH(bobKey, &aliceKey.PublicKey)
		if err != nil {
			t.Fatal(err)
		}
		aliceSecret, err := curve.ECDH(aliceKey, &bobKey.PublicKey)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(bobSecret, aliceSecret) {
			t.Error("two ECDH computations came out different")
		}
	})
}

func testAllCurves(t *testing.T, f func(t *testing.T, curve ecdh.Curve)) {
	t.Run("P256", func(t *testing.T) { f(t, ecdh.P256()) })
	t.Run("P384", func(t *testing.T) { f(t, ecdh.P384()) })
	t.Run("P521", func(t *testing.T) { f(t, ecdh.P521()) })
	t.Run("X25519", func(t *testing.T) { f(t, ecdh.X25519()) })
}

func BenchmarkECDH(b *testing.B) {
	benchmarkAllCurves(b, func(b *testing.B, curve ecdh.Curve) {
		c, err := chacha20.NewUnauthenticatedCipher(make([]byte, 32), make([]byte, 12))
		if err != nil {
			b.Fatal(err)
		}
		rand := cipher.StreamReader{
			S: c, R: zeroReader,
		}

		peerKey, err := curve.GenerateKey(rand)
		if err != nil {
			b.Fatal(err)
		}
		peerShare := peerKey.PublicKey.Bytes()
		b.ResetTimer()
		b.ReportAllocs()

		var allocationsSink byte

		for i := 0; i < b.N; i++ {
			key, err := curve.GenerateKey(rand)
			if err != nil {
				b.Fatal(err)
			}
			share := key.PublicKey.Bytes()
			peerPubKey, err := curve.NewPublicKey(peerShare)
			if err != nil {
				b.Fatal(err)
			}
			secret, err := curve.ECDH(key, peerPubKey)
			if err != nil {
				b.Fatal(err)
			}
			allocationsSink ^= secret[0] ^ share[0]
		}
	})
}

func benchmarkAllCurves(b *testing.B, f func(b *testing.B, curve ecdh.Curve)) {
	b.Run("P256", func(b *testing.B) { f(b, ecdh.P256()) })
	b.Run("P384", func(b *testing.B) { f(b, ecdh.P384()) })
	b.Run("P521", func(b *testing.B) { f(b, ecdh.P521()) })
	b.Run("X25519", func(b *testing.B) { f(b, ecdh.X25519()) })
}

type zr struct{}

// Read replaces the contents of dst with zeros. It is safe for concurrent use.
func (zr) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

var zeroReader = zr{}
