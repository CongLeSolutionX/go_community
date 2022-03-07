// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

// Unix cryptographically secure pseudorandom number
// generator.

package rand

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

import "crypto/internal/boring"

const urandomDevice = "/dev/urandom"

func init() {
<<<<<<< HEAD   (768804 [dev.boringcrypto] misc/boring: add new releases to RELEASES)
	if boring.Enabled {
		Reader = boring.RandReader
		return
	}
	if runtime.GOOS == "plan9" {
		Reader = newReader(nil)
	} else {
		Reader = &devReader{name: urandomDevice}
	}
=======
	Reader = &reader{}
>>>>>>> BRANCH (c9b606 crypto/rand: separate out plan9 X9.31 /dev/random expander)
}

// A reader satisfies reads by reading from urandomDevice
type reader struct {
	f    io.Reader
	mu   sync.Mutex
	used int32 // atomic; whether this reader has been used
}

// altGetRandom if non-nil specifies an OS-specific function to get
// urandom-style randomness.
var altGetRandom func([]byte) (ok bool)

func warnBlocked() {
	println("crypto/rand: blocked for 60 seconds waiting to read random data from the kernel")
}

<<<<<<< HEAD   (768804 [dev.boringcrypto] misc/boring: add new releases to RELEASES)
func (r *devReader) Read(b []byte) (n int, err error) {
	boring.Unreachable()
=======
func (r *reader) Read(b []byte) (n int, err error) {
>>>>>>> BRANCH (c9b606 crypto/rand: separate out plan9 X9.31 /dev/random expander)
	if atomic.CompareAndSwapInt32(&r.used, 0, 1) {
		// First use of randomness. Start timer to warn about
		// being blocked on entropy not being available.
		t := time.AfterFunc(time.Minute, warnBlocked)
		defer t.Stop()
	}
	if altGetRandom != nil && altGetRandom(b) {
		return len(b), nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.f == nil {
		f, err := os.Open(urandomDevice)
		if err != nil {
			return 0, err
		}
		r.f = bufio.NewReader(hideAgainReader{f})
	}
	return r.f.Read(b)
}

// hideAgainReader masks EAGAIN reads from /dev/urandom.
// See golang.org/issue/9205
type hideAgainReader struct {
	r io.Reader
}

func (hr hideAgainReader) Read(p []byte) (n int, err error) {
	n, err = hr.r.Read(p)
	if errors.Is(err, syscall.EAGAIN) {
		err = nil
	}
	return
}
<<<<<<< HEAD   (768804 [dev.boringcrypto] misc/boring: add new releases to RELEASES)

// Alternate pseudo-random implementation for use on
// systems without a reliable /dev/urandom.

// newReader returns a new pseudorandom generator that
// seeds itself by reading from entropy. If entropy == nil,
// the generator seeds itself by reading from the system's
// random number generator, typically /dev/random.
// The Read method on the returned reader always returns
// the full amount asked for, or else it returns an error.
//
// The generator uses the X9.31 algorithm with AES-128,
// reseeding after every 1 MB of generated data.
func newReader(entropy io.Reader) io.Reader {
	if entropy == nil {
		entropy = &devReader{name: "/dev/random"}
	}
	return &reader{entropy: entropy}
}

type reader struct {
	mu                   sync.Mutex
	budget               int // number of bytes that can be generated
	cipher               cipher.Block
	entropy              io.Reader
	time, seed, dst, key [aes.BlockSize]byte
}

func (r *reader) Read(b []byte) (n int, err error) {
	boring.Unreachable()
	r.mu.Lock()
	defer r.mu.Unlock()
	n = len(b)

	for len(b) > 0 {
		if r.budget == 0 {
			_, err := io.ReadFull(r.entropy, r.seed[0:])
			if err != nil {
				return n - len(b), err
			}
			_, err = io.ReadFull(r.entropy, r.key[0:])
			if err != nil {
				return n - len(b), err
			}
			r.cipher, err = aes.NewCipher(r.key[0:])
			if err != nil {
				return n - len(b), err
			}
			r.budget = 1 << 20 // reseed after generating 1MB
		}
		r.budget -= aes.BlockSize

		// ANSI X9.31 (== X9.17) algorithm, but using AES in place of 3DES.
		//
		// single block:
		// t = encrypt(time)
		// dst = encrypt(t^seed)
		// seed = encrypt(t^dst)
		ns := time.Now().UnixNano()
		binary.BigEndian.PutUint64(r.time[:], uint64(ns))
		r.cipher.Encrypt(r.time[0:], r.time[0:])
		for i := 0; i < aes.BlockSize; i++ {
			r.dst[i] = r.time[i] ^ r.seed[i]
		}
		r.cipher.Encrypt(r.dst[0:], r.dst[0:])
		for i := 0; i < aes.BlockSize; i++ {
			r.seed[i] = r.time[i] ^ r.dst[i]
		}
		r.cipher.Encrypt(r.seed[0:], r.seed[0:])

		m := copy(b, r.dst[0:])
		b = b[m:]
	}

	return n, nil
}
=======
>>>>>>> BRANCH (c9b606 crypto/rand: separate out plan9 X9.31 /dev/random expander)
