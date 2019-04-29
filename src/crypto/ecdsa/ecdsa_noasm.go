// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !s390x gccgo

// Package ecdsa implements the Elliptic Curve Digital Signature Algorithm, as
// defined in FIPS 186-3.
//
// This implementation  derives the nonce from an AES-CTR CSPRNG keyed by
// ChopMD(256, SHA2-512(priv.D || entropy || hash)). The CSPRNG key is IRO by
// a result of Coron; the AES-CTR stream is IRO under standard assumptions.

package ecdsa

import (
	"crypto/cipher"
	"crypto/elliptic"
	"math/big"
)

func sign(priv *PrivateKey, csprng *cipher.StreamReader, c elliptic.Curve, e *big.Int) (r, s *big.Int, err error) {
	r, s, err = signGeneric(priv, csprng, c, e)
	return
}
