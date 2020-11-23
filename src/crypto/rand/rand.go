// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rand implements a cryptographically secure
// random number generator.
package rand

import "io"

// Reader is a global, shared instance of a cryptographically
// secure random number generator.
//
// On Linux and FreeBSD, Reader uses getrandom(2) if available, /dev/urandom otherwise.
// On OpenBSD, Reader uses getentropy(2).
// On other Unix-like systems, Reader reads from /dev/urandom.
// On Windows systems, Reader uses the RtlGenRandom API.
// On Wasm, Reader uses the Web Crypto API.
//
// It is recommended to instead invoke Read as Reader can be mutated easily.
var Reader io.Reader

// internalReader exists to prevent Reader from being mutated easily
// by rogue dependencies. Please see https://golang.org/issue/42713.
var internalReader io.Reader

func init() {
	Reader = internalReader
}

// Read is a helper function that calls the internal reaer's Read method using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(b []byte) (n int, err error) {
	return io.ReadFull(internalReader, b)
}
