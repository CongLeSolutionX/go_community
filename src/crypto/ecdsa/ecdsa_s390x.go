// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build s390x,!gccgo

// Package ecdsa implements the Elliptic Curve Digital Signature Algorithm, as
// defined in FIPS 186-3.
//
// This implementation  derives the nonce from an AES-CTR CSPRNG keyed by
// ChopMD(256, SHA2-512(priv.D || entropy || )). The CSPRNG key is IRO by
// a result of Coron; the AES-CTR stream is IRO under standard assumptions.

package ecdsa

import (
	"crypto/cipher"
	"crypto/elliptic"
	"internal/cpu"
	"math/big"
)

// s390x accelerated signatures
//go:noescape
func kdsaSig(uint64, *byte) (uint64, uint64)

type signverify int

const (
	signing signverify = iota
	verifying
)

type bufferoffsets struct {
	baseSize       int
	hashSize       int
	offsetHash     int
	offsetKey1     int
	offsetRNorKey2 int
	offsetR        int
	offsetS        int
	functionCode   uint64
}

func preconditions(sv signverify, c elliptic.Curve, bo *bufferoffsets) (ret bool) {

	bitSize := c.Params().BitSize
	if !cpu.S390X.HasECDSA {
		return false
	}

	switch bitSize {
	case 256:
		bo.baseSize = 32
		bo.hashSize = 32
		bo.offsetHash = 64
		bo.offsetKey1 = 96
		bo.offsetRNorKey2 = 128
		bo.offsetR = 0
		bo.offsetS = 32
		if sv == signing {
			bo.functionCode = 9
		} else {
			bo.functionCode = 1
		}
		ret = true
	case 384:
		bo.baseSize = 48
		bo.hashSize = 48
		bo.offsetHash = 96
		bo.offsetKey1 = 144
		bo.offsetRNorKey2 = 192
		bo.offsetR = 0
		bo.offsetS = 48
		if sv == signing {
			bo.functionCode = 10
		} else {
			bo.functionCode = 2
		}
		ret = true
	case 521:
		bo.baseSize = 66
		bo.hashSize = 80
		bo.offsetHash = 160
		bo.offsetKey1 = 254
		bo.offsetRNorKey2 = 334
		bo.offsetR = 14
		bo.offsetS = 94
		if sv == signing {
			bo.functionCode = 11
			ret = false //not supporting Sign 521 for now
		} else {
			bo.functionCode = 3
			ret = true
		}
	default:
		ret = false
	}
	return
}

type parameters struct {
	baseSize, hashSize, offsetHash, offsetKey1, offsetRNorKey2, offsetR, offsetS int
	functionCode                                                                 uint64
}

func sign(priv *PrivateKey, csprng *cipher.StreamReader, c elliptic.Curve, e *big.Int) (r, s *big.Int, err error) {
	var k *big.Int
	var bo bufferoffsets
	if preconditions(signing, c, &bo) && e.Sign() != 0 {
		var buffer [1720]byte
		var filler [80]byte
		for {
			k, err = randFieldElement(c, csprng)
			if err != nil {
				r = nil
				return
			}
			rsize := bo.hashSize - len(e.Bytes())
			copy(buffer[bo.offsetHash:], filler[0:rsize])
			copy(buffer[bo.offsetHash+rsize:], e.Bytes())
			rsize = bo.baseSize - len(priv.D.Bytes())
			copy(buffer[bo.offsetKey1:], filler[0:rsize])
			copy(buffer[bo.offsetKey1+rsize:], priv.D.Bytes())
			rsize = bo.baseSize - len(k.Bytes())
			copy(buffer[bo.offsetRNorKey2:], filler[0:rsize])
			copy(buffer[bo.offsetRNorKey2+rsize:], k.Bytes())
			success, errn := kdsaSig(bo.functionCode, &buffer[0])
			if errn == 1 {
				return nil, nil, errZeroParam
			}
			if success == 0 { // success == 0 means successful signing
				r = new(big.Int)
				r.SetBytes(buffer[bo.offsetR : bo.offsetR+bo.baseSize])
				s = new(big.Int)
				s.SetBytes(buffer[bo.offsetS : bo.offsetS+bo.baseSize])
				break
			}
			//at this point, it must be that success == 1 and errn == 2: retry
		}
	} else {
		r, s, err = signGeneric(priv, csprng, c, e)
	}
	return
}

func verify(pub *PublicKey, c elliptic.Curve, e, r, s *big.Int) bool {
	var bo bufferoffsets
	if preconditions(verifying, c, &bo) && e.Sign() != 0 {
		var buffer [1720]byte
		var filler [80]byte
		rsize := bo.baseSize - len(r.Bytes())
		copy(buffer[bo.offsetR:], filler[0:rsize])
		copy(buffer[bo.offsetR+rsize:], r.Bytes())
		rsize = bo.baseSize - len(s.Bytes())
		copy(buffer[bo.offsetS:], filler[0:rsize])
		copy(buffer[bo.offsetS+rsize:], s.Bytes())
		rsize = bo.hashSize - len(e.Bytes())
		copy(buffer[bo.offsetHash:], filler[0:rsize])
		copy(buffer[bo.offsetHash+rsize:], e.Bytes())
		rsize = bo.baseSize - len(pub.X.Bytes())
		copy(buffer[bo.offsetKey1:], filler[0:rsize])
		copy(buffer[bo.offsetKey1+rsize:], pub.X.Bytes())
		rsize = bo.baseSize - len(pub.Y.Bytes())
		copy(buffer[bo.offsetRNorKey2:], filler[0:rsize])
		copy(buffer[bo.offsetRNorKey2+rsize:], pub.Y.Bytes())
		_, errn := kdsaSig(bo.functionCode, &buffer[0])
		if errn == 1 || errn == 2 {
			return false
		}
		// success == 0 means successful verification
		return true
	} else {
		return verifyGeneric(pub, c, e, r, s)
	}
}
