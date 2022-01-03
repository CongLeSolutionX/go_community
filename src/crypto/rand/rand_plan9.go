// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Plan9 cryptographically secure pseudorandom number
// generator.

package rand

import (
	"golang.org/x/crypto/chacha20"
	"io"
	"os"
	"sync"
	"time"
)

const randomDevice = "/dev/random"

func init() {
	Reader = &reader{}
}

// reader is a new pseudorandom generator that seeds itself by
// reading from /dev/random. The Read method on the returned
// reader always returns the full amount asked for, or else it
// returns an error. The generator is a fast erasure RNG.
type reader struct {
	mu      sync.Mutex
	seeded  sync.Once
	seedErr error
	key     [chacha20.KeySize]byte
}

func (r *reader) Read(b []byte) (n int, err error) {
	r.seeded.Do(func() {
		t := time.AfterFunc(time.Minute, func() {
			println("crypto/rand: blocked for 60 seconds waiting to read random data from the kernel")
		})
		defer t.Stop()
		entropy, err := os.Open(randomDevice)
		if err != nil {
			r.seedErr = err
		}
		_, err = io.ReadFull(entropy, r.key[:])
		r.seedErr = err
	})
	if r.seedErr != nil {
		return 0, r.seedErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	n = len(b)
	var zeroNonce [chacha20.NonceSize]byte
	stream, err := chacha20.NewUnauthenticatedCipher(r.key[:], zeroNonce[:])
	if err != nil {
		return 0, err
	}
	var zeroKey [chacha20.KeySize]byte
	stream.XORKeyStream(r.key[:], zeroKey[:])

	stream.XORKeyStream(b, b) // Wart: needless XORing with whatever junk is in b already.
	return n, nil
}
