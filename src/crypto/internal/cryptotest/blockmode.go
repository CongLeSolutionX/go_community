package cryptotest

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"
)

// MakeBlockMode returns a cipher.BlockMode instance.
// It expects len(iv) == b.BlockSize().
type MakeBlockMode func(b cipher.Block, iv []byte) cipher.BlockMode

// TestBlockMode performs a set of tests on cipher.BlockMode implementations,
// checking the documented requirements of CryptBlocks.
func TestBlockMode(t *testing.T, makeEncrypter, makeDecrypter MakeBlockMode) {
	// Create one random number generator to use for all BlockMode tests
	rng := newRandReader(t)

	// Test BlockMode using AES block cipher for each of its key lengths
	for _, keylen := range []int{128, 192, 256} {
		t.Run(fmt.Sprintf("AES-%d", keylen), func(t *testing.T) {
			// Generate a random IV and key to instantiate an AES block cipher
			iv := make([]byte, aes.BlockSize)
			rng.Read(iv)
			key := make([]byte, keylen/8)
			rng.Read(key)
			cipher, err := aes.NewCipher(key)
			if err != nil {
				panic(err)
			}
			testBlockModePair(t, rng, makeEncrypter, makeDecrypter, cipher, iv)
		})
	}

	t.Run("DES", func(t *testing.T) {
		// Generate a random IV and key to instantiate a DES block cipher
		key := make([]byte, 8)
		rng.Read(key)
		iv := make([]byte, des.BlockSize)
		rng.Read(iv)
		cipher, err := des.NewCipher(key)
		if err != nil {
			panic(err)
		}
		testBlockModePair(t, rng, makeEncrypter, makeDecrypter, cipher, iv)
	})
}

func testBlockModePair(t *testing.T, rng io.Reader, e, d MakeBlockMode, b cipher.Block, iv []byte) {
	t.Run("Encryption", func(t *testing.T) {
		testBlockMode(t, rng, e, b, iv)
	})

	t.Run("Decryption", func(t *testing.T) {
		testBlockMode(t, rng, d, b, iv)
	})

	t.Run("Cycle", func(t *testing.T) {
		bs := e(b, iv).BlockSize()
		if d(b, iv).BlockSize() != bs {
			t.Skip("mismatching encryption and decryption blocksizes")
		}

		before, dst, after := make([]byte, bs*2), make([]byte, bs*2), make([]byte, bs*2)
		rng.Read(before)

		e(b, iv).CryptBlocks(dst, before)
		d(b, iv).CryptBlocks(after, dst)
		expectEqual(t, after, before, "plaintext is different after a encrypt/decrypt cycle")
	})
}

func testBlockMode(t *testing.T, rng io.Reader, bm MakeBlockMode, b cipher.Block, iv []byte) {
	bs := bm(b, iv).BlockSize()

	t.Run("WrongIVLen", func(t *testing.T) {
		iv := make([]byte, b.BlockSize()+1)
		mustPanic(t, "IV length must equal block size", func() { bm(b, iv) })
	})

	t.Run("Aliasing", func(t *testing.T) {
		src, dst, before := make([]byte, bs*2), make([]byte, bs*2), make([]byte, bs*2)
		for _, length := range []int{0, bs, bs * 2} {
			rng.Read(src)
			copy(before, src)

			bm(b, iv).CryptBlocks(dst[:length], src[:length])
			expectEqual(t, src, before, "CryptBlocks modified src")

			// CryptBlocks on the same src data should yield same output even if dst=src
			copy(src, before)
			bm(b, iv).CryptBlocks(src[:length], src[:length])
			expectEqual(t, src[:length], dst[:length], "CryptBlocks behaves differently when dst = src")

			rng.Read(src)
			rng.Read(dst)
			copy(before, dst)
			bm(b, iv).CryptBlocks(dst, src[:length])
			expectEqual(t, dst[length:], before[length:], "CryptBlocks modified dst past len(src)")
		}
	})

	t.Run("BufferOverlap", func(t *testing.T) {
		src := make([]byte, bs+1)
		rng.Read(src)

		// Make src and dst slices point to same array with inexact overlap
		dst := src[1:]
		src = src[:len(src)-1]

		mustPanic(t, "invalid buffer overlap", func() { bm(b, iv).CryptBlocks(dst, src) })
	})

	// Input to CryptBlocks should be a multiple of BlockSize
	t.Run("PartialBlocks", func(t *testing.T) {
		// Check a few cases of not being a multiple of BlockSize
		for _, srcSize := range []int{bs - 1, bs + 1, 2*bs - 1, 2*bs + 1} {
			src := make([]byte, srcSize)
			dst := make([]byte, 3*bs) // Make a dst large enough all src
			mustPanic(t, "input not full blocks", func() { bm(b, iv).CryptBlocks(dst, src) })
		}
	})

	t.Run("OutOfBoundsWrite", func(t *testing.T) { // Issue 21104
		src, dst := make([]byte, bs*3), make([]byte, bs*3)
		rng.Read(src)
		copy(dst, src)
		mustPanic(t, "output smaller than input", func() { bm(b, iv).CryptBlocks(dst[bs:bs*2], src) })

		expectEqual(t, dst[bs*2:], src[bs*2:], "CryptBlocks did out of bounds write after end of dst slice")
		expectEqual(t, dst[:bs], src[:bs], "CryptBlocks did out of bounds write before beginning of dst slice")
	})

	// Check that output of cipher isn't affected by adjacent data beyond input
	// slice scope.
	t.Run("OutOfBoundsRead", func(t *testing.T) {
		src := make([]byte, bs)
		rng.Read(src)
		ctrlDst := make([]byte, bs) // Record a control output
		bm(b, iv).CryptBlocks(ctrlDst, src)

		// Make a buffer with src in the middle and data on either end
		buff := make([]byte, bs*3)
		copy(buff[bs:bs*2], src)
		rng.Read(buff[:bs])
		rng.Read(buff[bs*2:])

		testDst := make([]byte, bs)
		bm(b, iv).CryptBlocks(testDst, buff[bs:bs*2])

		expectEqual(t, testDst, ctrlDst, "CryptBlocks affected by data outside of src slice bounds")
	})

	t.Run("KeepState", func(t *testing.T) {
		src, serialDst, compositeDst := make([]byte, bs*4), make([]byte, bs*4), make([]byte, bs*4)
		rng.Read(src)

		length, block := 2*bs, bm(b, iv)
		block.CryptBlocks(serialDst, src[:length])
		block.CryptBlocks(serialDst[length:], src[length:])

		bm(b, iv).CryptBlocks(compositeDst, src)

		expectEqual(t, serialDst, compositeDst, "two successive CryptBlocks calls returned a different result than a single one")
	})
}

func newRandReader(t *testing.T) io.Reader {
	seed := time.Now().UnixNano()
	t.Logf("Deterministic RNG seed: 0x%x", seed)
	return rand.New(rand.NewSource(seed))
}

func expectEqual(t *testing.T, got, want []byte, msg string) {
	if !bytes.Equal(got, want) {
		t.Errorf("%s; got %x, want %x", msg, got, want)
	}
}

// Check if function f panics with message msg and error otherwise.
func mustPanic(t *testing.T, msg string, f func()) {
	t.Helper()

	defer func() {
		t.Helper()

		err := recover()

		if err == nil {
			t.Errorf("function did not panic, wanted %q", msg)
		} else {
			// Cast err to string
			err := fmt.Sprintf("%v", err)

			// Split message of form "path/to/block/directory: error message" at colon
			panicMsg := strings.SplitN(err, ": ", 2)

			// If the message didn't have a colon, it should
			if len(panicMsg) == 1 {
				t.Errorf(
					"got panic %q, wanted %q",
					panicMsg[0], "path/to/block/directory: "+msg)
				return
			}

			panicFrom := panicMsg[0]
			panicGot := panicMsg[1]
			if panicGot != msg {
				t.Errorf("%v got panic %q, wanted %q", panicFrom, panicGot, msg)
			}
		}
	}()
	f()
}
