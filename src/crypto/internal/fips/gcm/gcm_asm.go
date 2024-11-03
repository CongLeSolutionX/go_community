// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (amd64 || arm64) && !purego

package gcm

import (
	"crypto/internal/fips/aes"
	"crypto/internal/fips/alias"
	"crypto/subtle"
	"internal/cpu"
)

// The following functions are defined in gcm_*.s.

//go:noescape
func gcmAesInit(productTable *[16]gcmFieldElement, ks []uint32)

//go:noescape
func gcmAesData(productTable *[16]gcmFieldElement, data []byte, T *[16]byte)

//go:noescape
func gcmAesEnc(productTable *[16]gcmFieldElement, dst, src []byte, ctr, T *[16]byte, ks []uint32)

//go:noescape
func gcmAesDec(productTable *[16]gcmFieldElement, dst, src []byte, ctr, T *[16]byte, ks []uint32)

//go:noescape
func gcmAesFinish(productTable *[16]gcmFieldElement, tagMask, T *[16]byte, pLen, dLen uint64)

var supportsAESGCM = cpu.X86.HasAES && cpu.X86.HasPCLMULQDQ || cpu.ARM64.HasAES && cpu.ARM64.HasPMULL

func initGCM(g *GCM) {
	b, ok := g.cipher.(*aes.Block)
	if !ok || !supportsAESGCM {
		initGCMGeneric(g)
		return
	}

	gcmAesInit(&g.productTable, b.EncryptionKeySchedule())
}

func seal(g *GCM, dst, nonce, plaintext, data []byte) []byte {
	b, ok := g.cipher.(*aes.Block)
	if !ok || !supportsAESGCM {
		return sealGeneric(g, dst, nonce, plaintext, data)
	}

	var counter, tagMask [gcmBlockSize]byte

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmAesData(&g.productTable, nonce, &counter)
		gcmAesFinish(&g.productTable, &tagMask, &counter, uint64(len(nonce)), uint64(0))
	}

	b.Encrypt(tagMask[:], counter[:])

	var tagOut [gcmTagSize]byte
	gcmAesData(&g.productTable, data, &tagOut)

	ret, out := sliceForAppend(dst, len(plaintext)+g.tagSize)
	if alias.InexactOverlap(out[:len(plaintext)], plaintext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	if len(plaintext) > 0 {
		gcmAesEnc(&g.productTable, out, plaintext, &counter, &tagOut, b.EncryptionKeySchedule())
	}
	gcmAesFinish(&g.productTable, &tagMask, &tagOut, uint64(len(plaintext)), uint64(len(data)))
	copy(out[len(plaintext):], tagOut[:])

	return ret
}

func open(g *GCM, dst, nonce, ciphertext, data []byte) ([]byte, error) {
	b, ok := g.cipher.(*aes.Block)
	if !ok || !supportsAESGCM {
		return openGeneric(g, dst, nonce, ciphertext, data)
	}

	tag := ciphertext[len(ciphertext)-g.tagSize:]
	ciphertext = ciphertext[:len(ciphertext)-g.tagSize]

	// See GCM spec, section 7.1.
	var counter, tagMask [gcmBlockSize]byte

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmAesData(&g.productTable, nonce, &counter)
		gcmAesFinish(&g.productTable, &tagMask, &counter, uint64(len(nonce)), uint64(0))
	}

	b.Encrypt(tagMask[:], counter[:])

	var expectedTag [gcmTagSize]byte
	gcmAesData(&g.productTable, data, &expectedTag)

	ret, out := sliceForAppend(dst, len(ciphertext))
	if alias.InexactOverlap(out, ciphertext) {
		panic("crypto/cipher: invalid buffer overlap")
	}
	if len(ciphertext) > 0 {
		gcmAesDec(&g.productTable, out, ciphertext, &counter, &expectedTag, b.EncryptionKeySchedule())
	}
	gcmAesFinish(&g.productTable, &tagMask, &expectedTag, uint64(len(ciphertext)), uint64(len(data)))

	if subtle.ConstantTimeCompare(expectedTag[:g.tagSize], tag) != 1 {
		clear(out)
		return nil, errOpen
	}

	return ret, nil
}
