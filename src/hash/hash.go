// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hash provides interfaces for hash functions.
package hash

import "io"

// The marshaled binary representation of hashes in the standard library
// follow the following format:
//   [3 bytes hash magic identifier]
//   [1 byte hash kind]
//   [n bytes internal representation]
//
// The hash identifiers and kinds are:
//  md5.New()           "fnv" 0x01
//  sha1.New()          "sha" 0x01
//  sha256.New224()     "sha" 0x02
//  sha256.New()        "sha" 0x03
//  sha512.New384()     "sha" 0x04
//  sha512.New512_224() "sha" 0x05
//  sha512.New512_256() "sha" 0x06
//  sha512.New()        "sha" 0x07
//  adler32.New()       "adl" 0x01
//  crc32.New()         "crc" 0x01
//  crc64.New()         "crc" 0x02
//  fnv.New32()         "fnv" 0x01
//  fnv.New32a()        "fnv" 0x02
//  fnv.New64()         "fnv" 0x03
//  fnv.New64a()        "fnv" 0x04
//  fnv.New128()        "fnv" 0x05
//  fnv.New128a()       "fnv" 0x06

// Hash is the common interface implemented by all hash functions.
//
// Hash implementations in the standard library (e.g. hash/crc32 and
// crypto/sha256) implement the encoding.BinaryMarshaler and
// encoding.BinaryUnmarshaler interfaces. Marshaling a hash implementation
// allows its internal state to be saved and used for additional processing
// later, without having to re-write the data previously written to the hash.
type Hash interface {
	// Write (via the embedded io.Writer interface) adds more data to the running hash.
	// It never returns an error.
	io.Writer

	// Sum appends the current hash to b and returns the resulting slice.
	// It does not change the underlying hash state.
	Sum(b []byte) []byte

	// Reset resets the Hash to its initial state.
	Reset()

	// Size returns the number of bytes Sum will return.
	Size() int

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if all writes
	// are a multiple of the block size.
	BlockSize() int
}

// Hash32 is the common interface implemented by all 32-bit hash functions.
type Hash32 interface {
	Hash
	Sum32() uint32
}

// Hash64 is the common interface implemented by all 64-bit hash functions.
type Hash64 interface {
	Hash
	Sum64() uint64
}
