// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rand implements a cryptographically secure
// random number generator.
package rand

import (
	"crypto/internal/boring"
	"io"
	"sync/atomic"
	"time"
	_ "unsafe"
)

// Reader is a global, shared instance of a cryptographically
// secure random number generator. It is safe for concurrent use.
//
//   - On Linux, FreeBSD, Dragonfly, and Solaris, Reader uses getrandom(2).
//   - On macOS and iOS, Reader uses arc4random_buf(3).
//   - On OpenBSD, Reader uses getentropy(2).
//   - On NetBSD, Reader uses the kern.arandom sysctl.
//   - On Windows, Reader uses the ProcessPrng API.
//   - On js/wasm, Reader uses the Web Crypto API.
//   - On wasip1/wasm, Reader uses random_get.
//
// All the platform APIs above are documented to never return an error
// when used as they are in this package.
var Reader io.Reader

func init() {
	if boring.Enabled {
		Reader = boring.RandReader
		return
	}
	Reader = &reader{}
}

var firstUse atomic.Bool

func warnBlocked() {
	println("crypto/rand: blocked for 60 seconds waiting to read random data from the kernel")
}

type reader struct{}

func (r *reader) Read(b []byte) (n int, err error) {
	boring.Unreachable()
	if firstUse.CompareAndSwap(false, true) {
		// First use of randomness. Start timer to warn about
		// being blocked on entropy not being available.
		t := time.AfterFunc(time.Minute, warnBlocked)
		defer t.Stop()
	}
	if err := read(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

// fatal is [runtime.fatal], pushed via linkname.
//
//go:linkname fatal
func fatal(string)

// Read fills b with cryptographically secure random bytes. It never returns an
// error, and always fills b entirely.
//
// If [Reader] is set to a non-default value, Read calls [io.ReadFull] on
// [Reader] and crashes the program irrecoverably if an error is returned.
func Read(b []byte) (n int, err error) {
	// We don't want b to escape to the heap, but escape analysis can't see
	// through a potentially overridden Reader, so we special-case the default
	// case which we can keep non-escaping, and in the general case we read into
	// a heap buffer and copy from it.
	if r, ok := Reader.(*reader); ok {
		_, err = r.Read(b)
	} else {
		bb := make([]byte, len(b))
		_, err = io.ReadFull(Reader, bb)
		copy(b, bb)
	}
	if err != nil {
		fatal("crypto/rand: failed to read random data (see https://go.dev/issue/66821): " + err.Error())
		panic("unreachable") // To be sure.
	}
	return len(b), nil
}
