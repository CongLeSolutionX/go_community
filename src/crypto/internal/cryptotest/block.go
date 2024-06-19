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

	// Checks baseline Encrypt/Decrypt functionality.  More thorough
	// implementation-specific characterization/golden tests should be done
	// for each block cipher implementation test.
	t.Run("InvertibleEncryptDecrypt", func(t *testing.T) {
		// Generate random key
		key := make([]byte, keySize)
		rng.Read(key)

		block, err := mb(key)

		if err != nil {
			panic(err)
		}

		blockSize := block.BlockSize()

		// Check Decrypt inverts Encrypt
		before, ciphertext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Encrypt(ciphertext, before)
		block.Decrypt(after, ciphertext)

		if !bytes.Equal(before, after) {
			t.Errorf("Encrypted message %x. Decrypt should invert Encrypt but got %x", before, after)
		}

		// Check Encrypt inverts Decrypt (assumes block ciphers are deterministic)
		before, plaintext, after := make([]byte, blockSize), make([]byte, blockSize), make([]byte, blockSize)

		rng.Read(before)

		block.Decrypt(plaintext, before)
		block.Encrypt(after, plaintext)

		if !bytes.Equal(before, after) {
			t.Errorf("Decrypted ciphertext %x. Encrypt should deterministically invert Decrypt but got %x", before, after)
		}

	})

	t.Run("BlockScope", func(t *testing.T) {
		key := make([]byte, keySize)
		rng.Read(key)

		block, err := mb(key)

		if err != nil {
			panic(err)
		}

		blockSize := block.BlockSize()

		extraBuff := []byte("This shouldn't be touched")

		// Check that Encrypt doesn't touch output beyond block size
		src := make([]byte, blockSize)
		rng.Read(src)
		dst := make([]byte, blockSize+len(extraBuff))
		copy(dst[blockSize:], extraBuff) // Put extraBuff as suffix

		if block.Encrypt(dst, src); !bytes.Equal(dst[blockSize:], extraBuff) {
			t.Errorf("Encrypt writing to more bytes than block size.\n Last %v bytes of dst got %x, want %x", len(extraBuff), dst[blockSize:], extraBuff)
		}

		src = make([]byte, blockSize)
		rng.Read(src)
		dst = make([]byte, len(extraBuff)+blockSize)

		prefix := dst[:len(extraBuff)]
		copy(prefix, extraBuff) //Put extraBuff as prefix
		suffix := dst[len(extraBuff):]

		if block.Encrypt(suffix, src); !bytes.Equal(prefix, extraBuff) {
			t.Errorf("Encrypt writing to data preceding dst slice.\n First %v bytes of array that dst references got %x, want %x", len(extraBuff), prefix, extraBuff)
		}

		// Check that Decrypt doesn't touch output beyond block size
		src = make([]byte, blockSize)
		rng.Read(src)
		dst = make([]byte, blockSize+len(extraBuff))
		copy(dst[blockSize:], extraBuff) // Put extraBuff as suffix

		if block.Decrypt(dst, src); !bytes.Equal(dst[blockSize:], extraBuff) {
			t.Errorf("Decrypt writing to more bytes than block size.\n Last %v bytes of dst got %x, want %x", len(extraBuff), dst[blockSize:], extraBuff)
		}

		src = make([]byte, blockSize)
		rng.Read(src)
		dst = make([]byte, len(extraBuff)+blockSize)

		prefix = dst[:len(extraBuff)]
		copy(prefix, extraBuff) //Put extraBuff as prefix
		suffix = dst[len(extraBuff):]

		if block.Decrypt(suffix, src); !bytes.Equal(prefix, extraBuff) {
			t.Errorf("Decrypt writing to data preceding dst slice.\n First %v bytes of array that dst references got %x, want %x", len(extraBuff), prefix, extraBuff)
		}

		// Check that output of Encrypt isn't affected by out-of-blocksize-scope input
		// This test assumes block ciphers encrypt deterministically.
		src = make([]byte, blockSize)
		ctrlDst := make([]byte, blockSize) // Record a control ciphertext
		block.Encrypt(ctrlDst, src)

		src = make([]byte, blockSize+6) // We will pad src on either side with data
		testDst := make([]byte, blockSize)

		copy(src[:3], "abc")
		copy(src[len(src)-3:], "def")
		src = src[3 : len(src)-3] // Only will pass in the relevant part w/o padding

		if block.Encrypt(testDst, src); !bytes.Equal(testDst, ctrlDst) {
			t.Errorf("Encrypt using input beyond BlockSize when encrypting %x.\n Floating data %x on the left of src and %x on the right affects Encrypt.\n Encrypt got %x\n Want %x", src, []byte("abc"), []byte("def"), testDst, ctrlDst)
		}

		// Check that output of Decrypt isn't affected by out-of-blocksize-scope input
		src = make([]byte, blockSize)
		ctrlDst = make([]byte, blockSize) // Record a control plaintext
		block.Decrypt(ctrlDst, src)

		src = make([]byte, blockSize+6) // We will pad src on either side with data
		testDst = make([]byte, blockSize)

		copy(src[:3], "abc")
		copy(src[len(src)-3:], "def")
		src = src[3 : len(src)-3] // Only will pass in the relevant part w/o padding

		if block.Decrypt(testDst, src); !bytes.Equal(testDst, ctrlDst) {
			t.Errorf("Decrypt using input beyond BlockSize when decrypting %x.\n Floating data %x on the left of src and %x on the right affects Decrypt.\n Decrypt got %x\n Want %x", src, []byte("abc"), []byte("def"), testDst, ctrlDst)
		}
	})

	t.Run("BufferOverlap", func(t *testing.T) {
		key := make([]byte, keySize)
		rng.Read(key)

		block, err := mb(key)

		if err != nil {
			panic(err)
		}

		// Make src and dst slices point to same array with inexact overlap
		src := make([]byte, block.BlockSize()+1)
		rng.Read(src)
		dst := src[1:]

		mustPanic(t, "invalid buffer overlap", func() { block.Encrypt(dst, src) })
	})

	t.Run("ShortBlock", func(t *testing.T) {
		key := make([]byte, keySize)
		rng.Read(key)

		block, err := mb(key)

		if err != nil {
			panic(err)
		}

		// Returns slice of n bytes of an n+1 length array.  Lets us test that a
		// slice is still considered too short even if the underlying array it
		// points to is large enough.
		bytes := func(n int) []byte { return make([]byte, n+1)[0:n] }

		blockSize := block.BlockSize()

		mustPanic(t, "input not full block", func() { block.Encrypt(bytes(blockSize), bytes(blockSize-1)) })
		mustPanic(t, "input not full block", func() { block.Decrypt(bytes(blockSize), bytes(blockSize-1)) })
		mustPanic(t, "output not full block", func() { block.Encrypt(bytes(blockSize-1), bytes(blockSize)) })
		mustPanic(t, "output not full block", func() { block.Decrypt(bytes(blockSize-1), bytes(blockSize)) })
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
