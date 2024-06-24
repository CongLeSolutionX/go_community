// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cryptotest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
	"testing"
)

// MakeBlockMode returns a cipher.BlockMode instance.
// It expects len(iv) == b.BlockSize().
type MakeBlockMode func(b cipher.Block, iv []byte) cipher.BlockMode

// TestBlockMode performs a set of tests on cipher.BlockMode implementations,
// checking the documented requirements of CryptBlocks.
func TestBlockMode(t *testing.T, makeEncrypter, makeDecrypter MakeBlockMode) {

	// Test BlockMode using AES block cipher for each of its key lengths
	for _, keylen := range []int{128, 192, 256} {
		t.Run(fmt.Sprintf("AES-%d", keylen), func(t *testing.T) {
			rng := newRandReader(t)

			// Generate a random IV and key to instantiate an AES block cipher
			iv := make([]byte, aes.BlockSize)
			rng.Read(iv)
			key := make([]byte, keylen/8)
			rng.Read(key)
			cipher, err := aes.NewCipher(key)
			if err != nil {
				panic(err)
			}
			testBlockModePair(t, makeEncrypter, makeDecrypter, cipher, iv)
		})
	}

	t.Run("DES", func(t *testing.T) {
		rng := newRandReader(t)

		// Generate a random IV and key to instantiate a DES block cipher
		key := make([]byte, 8)
		rng.Read(key)
		iv := make([]byte, des.BlockSize)
		rng.Read(iv)
		cipher, err := des.NewCipher(key)
		if err != nil {
			panic(err)
		}
		testBlockModePair(t, makeEncrypter, makeDecrypter, cipher, iv)
	})
}

func testBlockModePair(t *testing.T, enc, dec MakeBlockMode, b cipher.Block, iv []byte) {
	t.Run("Encryption", func(t *testing.T) {
		testBlockMode(t, enc, b, iv)
	})

	t.Run("Decryption", func(t *testing.T) {
		testBlockMode(t, dec, b, iv)
	})

	t.Run("Roundtrip", func(t *testing.T) {
		rng := newRandReader(t)

		blockSize := enc(b, iv).BlockSize()
		if decBlockSize := dec(b, iv).BlockSize(); decBlockSize != blockSize {
			t.Errorf("decryption blocksize different than encryption's; got %d, want %d", decBlockSize, blockSize)
		}

		before, dst, after := make([]byte, blockSize*2), make([]byte, blockSize*2), make([]byte, blockSize*2)
		rng.Read(before)

		enc(b, iv).CryptBlocks(dst, before)
		dec(b, iv).CryptBlocks(after, dst)
		if !bytes.Equal(after, before) {
			t.Errorf("plaintext is different after an encrypt/decrypt cycle; got %x, want %x", after, before)
		}
	})
}

func testBlockMode(t *testing.T, bm MakeBlockMode, b cipher.Block, iv []byte) {
	blockSize := bm(b, iv).BlockSize()

	t.Run("WrongIVLen", func(t *testing.T) {
		iv := make([]byte, b.BlockSize()+1)
		mustPanic(t, "IV length must equal block size", func() { bm(b, iv) })
	})

	t.Run("Aliasing", func(t *testing.T) {
		rng := newRandReader(t)

		src, dst, before := make([]byte, blockSize*2), make([]byte, blockSize*2), make([]byte, blockSize*2)
		for _, length := range []int{0, blockSize, blockSize * 2} {
			rng.Read(src)
			copy(before, src)

			bm(b, iv).CryptBlocks(dst[:length], src[:length])
			if !bytes.Equal(src, before) {
				t.Errorf("CryptBlocks modified src; got %x, want %x", src, before)
			}

			// CryptBlocks on the same src data should yield same output even if dst=src
			copy(src, before)
			bm(b, iv).CryptBlocks(src[:length], src[:length])
			if !bytes.Equal(src[:length], dst[:length]) {
				t.Errorf("CryptBlocks behaves differently when dst = src; got %x, want %x", src[:length], dst[:length])
			}

			rng.Read(src)
			rng.Read(dst)
			copy(before, dst)
			bm(b, iv).CryptBlocks(dst, src[:length])
			if !bytes.Equal(dst[length:], before[length:]) {
				t.Errorf("CryptBlocks modified dst past len(src); got %x, want %x", dst[length:], before[length:])
			}
		}
	})

	t.Run("BufferOverlap", func(t *testing.T) {
		rng := newRandReader(t)

		buff := make([]byte, blockSize*2)
		rng.Read(buff)

		// Make src and dst slices point to same array with inexact overlap
		src := buff[:blockSize]
		dst := buff[1 : blockSize+1]
		mustPanic(t, "invalid buffer overlap", func() { bm(b, iv).CryptBlocks(dst, src) })

		// Only overlap on one byte
		src = buff[:blockSize]
		dst = buff[blockSize-1 : 2*blockSize-1]
		mustPanic(t, "invalid buffer overlap", func() { bm(b, iv).CryptBlocks(dst, src) })

		// src comes after dst with inexact overlap
		src = buff[blockSize-1 : 2*blockSize-1]
		dst = buff[:blockSize]
		mustPanic(t, "invalid buffer overlap", func() { bm(b, iv).CryptBlocks(dst, src) })
	})

	// Input to CryptBlocks should be a multiple of BlockSize
	t.Run("PartialBlocks", func(t *testing.T) {
		// Check a few cases of not being a multiple of BlockSize
		for _, srcSize := range []int{blockSize - 1, blockSize + 1, 2*blockSize - 1, 2*blockSize + 1} {
			src := make([]byte, srcSize)
			dst := make([]byte, 3*blockSize) // Make a dst large enough for all src
			mustPanic(t, "input not full blocks", func() { bm(b, iv).CryptBlocks(dst, src) })
		}
	})

	t.Run("OutOfBoundsWrite", func(t *testing.T) { // Issue 21104
		rng := newRandReader(t)

		src, dst := make([]byte, blockSize*3), make([]byte, blockSize*3)
		rng.Read(src)
		rng.Read(dst)

		initPrefix, initSuffix := make([]byte, blockSize), make([]byte, blockSize)

		endOfPrefix, startOfSuffix := blockSize, blockSize*2
		copy(initPrefix, dst[:endOfPrefix])
		copy(initSuffix, dst[startOfSuffix:])

		// Shouldn't write to anything outside of dst on a valid CryptBlocks call
		bm(b, iv).CryptBlocks(dst[endOfPrefix:startOfSuffix], src[endOfPrefix:startOfSuffix])
		if !bytes.Equal(dst[startOfSuffix:], initSuffix) {
			t.Errorf("block cipher did out of bounds write after end of dst slice; got %x, want %x", dst[startOfSuffix:], initSuffix)
		}

		if !bytes.Equal(dst[:endOfPrefix], initPrefix) {
			t.Errorf("block cipher did out of bounds write before beginning of dst slice; got %x, want %x", dst[:endOfPrefix], initPrefix)
		}

		// Shouldn't write to anything outside of dst even if src is bigger
		mustPanic(t, "output smaller than input", func() {
			bm(b, iv).CryptBlocks(dst[endOfPrefix:startOfSuffix], src)
		})

		if !bytes.Equal(dst[startOfSuffix:], initSuffix) {
			t.Errorf("block cipher did out of bounds write after end of dst slice; got %x, want %x", dst[startOfSuffix:], initSuffix)
		}
		if !bytes.Equal(dst[:endOfPrefix], initPrefix) {
			t.Errorf("block cipher did out of bounds write before beginning of dst slice; got %x, want %x", dst[:endOfPrefix], initPrefix)
		}
	})

	// Check that output of cipher isn't affected by adjacent data beyond input
	// slice scope.
	t.Run("OutOfBoundsRead", func(t *testing.T) {
		rng := newRandReader(t)

		src := make([]byte, blockSize)
		rng.Read(src)
		ctrlDst := make([]byte, blockSize) // Record a control output
		bm(b, iv).CryptBlocks(ctrlDst, src)

		// Make a buffer with src in the middle and data on either end
		buff := make([]byte, blockSize*3)
		endOfPrefix, startOfSuffix := blockSize, blockSize*2

		copy(buff[endOfPrefix:startOfSuffix], src)
		rng.Read(buff[:endOfPrefix])
		rng.Read(buff[startOfSuffix:])

		testDst := make([]byte, blockSize)
		bm(b, iv).CryptBlocks(testDst, buff[endOfPrefix:startOfSuffix])

		if !bytes.Equal(testDst, ctrlDst) {
			t.Errorf("CryptBlocks affected by data outside of src slice bounds; got %x, want %x", testDst, ctrlDst)
		}
	})

	t.Run("KeepState", func(t *testing.T) {
		rng := newRandReader(t)

		src, serialDst, compositeDst := make([]byte, blockSize*4), make([]byte, blockSize*4), make([]byte, blockSize*4)
		rng.Read(src)

		length, block := 2*blockSize, bm(b, iv)
		block.CryptBlocks(serialDst, src[:length])
		block.CryptBlocks(serialDst[length:], src[length:])

		bm(b, iv).CryptBlocks(compositeDst, src)

		if !bytes.Equal(serialDst, compositeDst) {
			t.Errorf("two successive CryptBlocks calls returned a different result than a single one; got %x, want %x", serialDst, compositeDst)
		}
	})
}
