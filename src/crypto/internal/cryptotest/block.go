package cryptotest

import (
	"bytes"
	"crypto/cipher"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type MakeBlock func(key []byte) (cipher.Block, error)

func TestBlock(t *testing.T, keySize int, mb MakeBlock) {
	// Create one random number generator to use for all Block tests
	rng := newRandReader(t)

	// Generate random key
	key := make([]byte, keySize)
	rng.Read(key)

	block, err := mb(key)
	if err != nil {
		panic(err)
	}

	blockSize := block.BlockSize()

	t.Run("Encryption", func(t *testing.T) {
		testCipher(t, rng, block.Encrypt, blockSize)
	})

	t.Run("Decryption", func(t *testing.T) {
		testCipher(t, rng, block.Decrypt, blockSize)
	})

	// Checks baseline Encrypt/Decrypt functionality.  More thorough
	// implementation-specific characterization/golden tests should be done
	// for each block cipher implementation test.
	t.Run("Cycle", func(t *testing.T) {
		// Check Decrypt inverts Encrypt
		before, ciphertext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Encrypt(ciphertext, before)
		block.Decrypt(after, ciphertext)

		expectEqual(t, after, before, "plaintext is different after an encrypt/decrypt cycle")

		// Check Encrypt inverts Decrypt (assumes block ciphers are deterministic)
		before, plaintext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Decrypt(plaintext, before)
		block.Encrypt(after, plaintext)

		expectEqual(t, after, before, "ciphertext is different after a decrypt/encrypt cycle")
	})

}

func testCipher(t *testing.T, rng io.Reader, cipher func(dst, src []byte), blockSize int) {
	t.Run("Aliasing", func(t *testing.T) {
		src, dst, before := make([]byte, blockSize*2), make([]byte, blockSize*2), make([]byte, blockSize*2)

		rng.Read(src)
		copy(before, src)
		cipher(dst[:blockSize], src[:blockSize])
		expectEqual(t, src, before, "block cipher modified src")

		// cipher on the same src data should yield same output even if dst=src
		copy(src, before)
		cipher(src[:blockSize], src[:blockSize])
		expectEqual(t, src[:blockSize], dst[:blockSize], "block cipher behaves differently when dst = src")

		rng.Read(src)
		rng.Read(dst)
		copy(before, dst)
		cipher(dst, src[:blockSize])
		expectEqual(t, dst[blockSize:], before[blockSize:], "block cipher modified dst past BlockSize bytes")

		// cipher should yield same dst when given same src even if src is longer
		// than BlockSize
		copy(before, dst)
		cipher(dst, src) // src unchanged from last cipher call except longer
		expectEqual(t, dst[:blockSize], before[:blockSize], "block cipher affected by src data beyond BlockSize bytes")

	})

	t.Run("OutOfBoundsWrite", func(t *testing.T) {
		src, dst := make([]byte, blockSize), make([]byte, blockSize*3)
		rng.Read(src)
		rng.Read(dst)

		beforePrefix, beforeSuffix := make([]byte, blockSize), make([]byte, blockSize)
		copy(beforePrefix, dst[:blockSize])
		copy(beforeSuffix, dst[blockSize*2:])

		cipher(dst[blockSize:blockSize*2], src)

		expectEqual(t, dst[blockSize*2:], beforeSuffix, "block cipher did out of bounds write after end of dst slice")
		expectEqual(t, dst[:blockSize], beforePrefix, "block cipher did out of bounds write before beginning of dst slice")
	})

	// Check that output of cipher isn't affected by adjacent data beyond input
	// slice scope.
	// For encryption, this assumes block ciphers encrypt deterministically.
	t.Run("OutOfBoundsRead", func(t *testing.T) {
		src := make([]byte, blockSize)
		rng.Read(src)
		ctrlDst := make([]byte, blockSize) // Record a control ciphertext
		cipher(ctrlDst, src)

		// Make a buffer with src in the middle and data on either end
		buff := make([]byte, blockSize*3)
		copy(buff[blockSize:blockSize*2], src)
		rng.Read(buff[:blockSize])
		rng.Read(buff[blockSize*2:])

		testDst := make([]byte, blockSize)
		cipher(testDst, buff[blockSize:blockSize*2])

		expectEqual(t, testDst, ctrlDst, "block cipher affected by data outside of src slice bounds")
	})

	t.Run("BufferOverlap", func(t *testing.T) {
		// Make src and dst slices point to same array with inexact overlap
		src := make([]byte, blockSize+1)
		rng.Read(src)
		dst := src[1:]

		mustPanic(t, "invalid buffer overlap", func() { cipher(dst, src) })
	})

	t.Run("ShortBlock", func(t *testing.T) {
		// Returns slice of n bytes of an n+1 length array.  Lets us test that a
		// slice is still considered too short even if the underlying array it
		// points to is large enough.
		bytes := func(n int) []byte { return make([]byte, n+1)[0:n] }

		mustPanic(t, "input not full block", func() { cipher(bytes(blockSize), bytes(blockSize-1)) })
		mustPanic(t, "output not full block", func() { cipher(bytes(blockSize-1), bytes(blockSize)) })
	})
}

func newRandReader(t *testing.T) io.Reader {
	seed := time.Now().UnixNano()
	t.Logf("Deterministic RNG seed: 0x%x", seed)
	return rand.New(rand.NewSource(seed))
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

func expectEqual(t *testing.T, got, want []byte, msg string) {
	t.Helper()

	if !bytes.Equal(got, want) {
		t.Errorf("%s; got %x, want %x", msg, got, want)
	}
}
