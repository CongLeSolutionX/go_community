// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm

package sha1

import (
	"reflect"
	"unsafe"
)

// Defined in sha1block_arm.s.
//
// Because the hashing function for SHA1 on ARM exceeds the stack limit
// for NOSPLIT, we cannot safely JMP from another assembly routine
// like elsewhere (e.g. amd64). To ensure we have a well-formed stack
// that can be correctly grown, we merge the string and []byte paths
// here in Go and call into the shared block function in assembly.
//go:noescape
func blockPtr(dig *digest, p uintptr, n int)

func blockString(dig *digest, s string) {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	blockPtr(dig, hdr.Data, hdr.Len)
}

func block(dig *digest, p []byte) {
	blockPtr(dig, uintptr(unsafe.Pointer(&p[0])), len(p))
}
