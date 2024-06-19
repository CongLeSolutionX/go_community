package cryptotest

import (
	"bytes"
	"crypto/cipher"
	"testing"
)

type MakeBlock func(key []byte) (cipher.Block, error)

// TestBlock performs a set of tests on cipher.Block implementations, checking
// the documented requirements of BlockSize, Encrypt, and Decrypt.
func TestBlock(t *testing.T, keySize int, mb MakeBlock) {
	// Generate random key
	key := make([]byte, keySize)
	newRandReader(t).Read(key)
	t.Logf("Cipher key: 0x%x", key)

	block, err := mb(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()

	t.Run("Encryption", func(t *testing.T) {
		testCipher(t, block.Encrypt, blockSize)
	})

	t.Run("Decryption", func(t *testing.T) {
		testCipher(t, block.Decrypt, blockSize)
	})

	// Checks baseline Encrypt/Decrypt functionality.  More thorough
	// implementation-specific characterization/golden tests should be done
	// for each block cipher implementation test.
	t.Run("Cycle", func(t *testing.T) {
		rng := newRandReader(t)

		// Check Decrypt inverts Encrypt
		before, ciphertext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Encrypt(ciphertext, before)
		block.Decrypt(after, ciphertext)

		if !bytes.Equal(after, before) {
			t.Errorf("plaintext is different after an encrypt/decrypt cycle; got %x, want %x", after, before)
		}

		// Check Encrypt inverts Decrypt (assumes block ciphers are deterministic)
		before, plaintext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Decrypt(plaintext, before)
		block.Encrypt(after, plaintext)

		if !bytes.Equal(after, before) {
			t.Errorf("ciphertext is different after a decrypt/encrypt cycle; got %x, want %x", after, before)
		}
	})

}

func testCipher(t *testing.T, cipher func(dst, src []byte), blockSize int) {
	t.Run("Aliasing", func(t *testing.T) {
		rng := newRandReader(t)

		src, dst, before := make([]byte, blockSize*2), make([]byte, blockSize*2), make([]byte, blockSize*2)

		rng.Read(src)
		copy(before, src)
		cipher(dst[:blockSize], src[:blockSize])
		if !bytes.Equal(src, before) {
			t.Errorf("block cipher modified src; got %x, want %x", src, before)
		}

		// cipher on the same src data should yield same output even if dst=src
		copy(src, before)
		cipher(src[:blockSize], src[:blockSize])
		if !bytes.Equal(src[:blockSize], dst[:blockSize]) {
			t.Errorf("block cipher behaves differently when dst = src; got %x, want %x", src[:blockSize], dst[:blockSize])
		}

		rng.Read(src)
		rng.Read(dst)
		copy(before, dst)
		cipher(dst, src[:blockSize])
		if !bytes.Equal(dst[blockSize:], before[blockSize:]) {
			t.Errorf("block cipher modified dst past BlockSize bytes; got %x, want %x", dst[blockSize:], before[blockSize:])
		}

		// cipher should yield same dst when given same src even if src is longer
		// than BlockSize
		copy(before, dst)
		cipher(dst, src) // src unchanged from last cipher call except longer
		if !bytes.Equal(dst[:blockSize], before[:blockSize]) {
			t.Errorf("block cipher affected by src data beyond BlockSize bytes; got %x, want %x", dst[:blockSize], before[:blockSize])
		}
	})

	t.Run("OutOfBoundsWrite", func(t *testing.T) {
		rng := newRandReader(t)

		src, dst := make([]byte, blockSize), make([]byte, blockSize*3)
		rng.Read(src)
		rng.Read(dst)

		initPrefix, initSuffix := make([]byte, blockSize), make([]byte, blockSize)

		endOfPrefix, startOfSuffix := blockSize, blockSize*2
		copy(initPrefix, dst[:endOfPrefix])
		copy(initSuffix, dst[startOfSuffix:])

		cipher(dst[endOfPrefix:startOfSuffix], src)

		if !bytes.Equal(dst[startOfSuffix:], initSuffix) {
			t.Errorf("block cipher did out of bounds write after end of dst slice; got %x, want %x", dst[startOfSuffix:], initSuffix)
		}
		if !bytes.Equal(dst[:endOfPrefix], initPrefix) {
			t.Errorf("block cipher did out of bounds write before beginning of dst slice; got %x, want %x", dst[:endOfPrefix], initPrefix)
		}
	})

	// Check that output of cipher isn't affected by adjacent data beyond input
	// slice scope.
	// For encryption, this assumes block ciphers encrypt deterministically.
	t.Run("OutOfBoundsRead", func(t *testing.T) {
		rng := newRandReader(t)

		src := make([]byte, blockSize)
		rng.Read(src)
		ctrlDst := make([]byte, blockSize) // Record a control ciphertext
		cipher(ctrlDst, src)

		// Make a buffer with src in the middle and data on either end
		buff := make([]byte, blockSize*3)
		endOfPrefix, startOfSuffix := blockSize, blockSize*2

		copy(buff[endOfPrefix:startOfSuffix], src)
		rng.Read(buff[:endOfPrefix])
		rng.Read(buff[startOfSuffix:])

		testDst := make([]byte, blockSize)
		cipher(testDst, buff[endOfPrefix:startOfSuffix])

		if !bytes.Equal(testDst, ctrlDst) {
			t.Errorf("block cipher affected by data outside of src slice bounds; got %x, want %x", testDst, ctrlDst)
		}
	})

	t.Run("BufferOverlap", func(t *testing.T) {
		rng := newRandReader(t)

		buff := make([]byte, blockSize*2)
		rng.Read((buff))

		// Make src and dst slices point to same array with inexact overlap
		src := buff[:blockSize]
		dst := buff[1 : blockSize+1]
		mustPanic(t, "invalid buffer overlap", func() { cipher(dst, src) })

		// Only overlap on one byte
		src = buff[:blockSize]
		dst = buff[blockSize-1 : 2*blockSize-1]
		mustPanic(t, "invalid buffer overlap", func() { cipher(dst, src) })

		// src comes after dst with inexact overlap
		src = buff[blockSize-1 : 2*blockSize-1]
		dst = buff[:blockSize]
		mustPanic(t, "invalid buffer overlap", func() { cipher(dst, src) })
	})

	t.Run("ShortBlock", func(t *testing.T) {
		// Returns slice of n bytes of an n+1 length array.  Lets us test that a
		// slice is still considered too short even if the underlying array it
		// points to is large enough.
		byteSlice := func(n int) []byte { return make([]byte, n+1)[0:n] }

		mustPanic(t, "input not full block", func() { cipher(byteSlice(blockSize), byteSlice(blockSize-1)) })
		mustPanic(t, "output not full block", func() { cipher(byteSlice(blockSize-1), byteSlice(blockSize)) })
	})
}

func mustPanic(t *testing.T, msg string, f func()) {
	t.Helper()

	defer func() {
		t.Helper()

		err := recover()

		if err == nil {
			t.Errorf("function did not panic for %q", msg)
		}
	}()
	f()
}
