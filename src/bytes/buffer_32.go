// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 arm mips amd64p32 nacl

package bytes

// This file contains the definition of the Buffer type for platforms where
// pointers are 32 bits.

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf      []byte // contents are the bytes buf[off : len(buf)]
	off      int    // read at &buf[off], write at &buf[len(buf)]
	lastRead readOp // last read operation, so that Unread* can work correctly.

	// memory to hold first slice; helps small buffers avoid allocation. on 32 bit
	// platforms the bootstrap buffer is 44 bytes long so that the Buffer type is
	// 1 cacheline long (64 bytes).
	// FIXME: lastRead can be shurnk down to 1 byte, and the resulting space used
	// to enlarge bootstrap.
	bootstrap [44]byte
}
