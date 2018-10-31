// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 amd64 amd64p32 s390x arm arm64 ppc64 ppc64le mips mipsle mips64 mips64le wasm

package bytealg

import _ "unsafe"

//go:noescape
func IndexByte(b []byte, c byte) int

//go:noescape
func IndexByteString(s string, c byte) int

// The declarations below generate ABI wrappers for functions
// implemented in assembly in this package but declared in another
// package.

//go:linkname abigen_bytes_IndexByte bytes.IndexByte
func abigen_bytes_IndexByte(b []byte, c byte) int

//go:linkname abigen_strings_IndexByte strings.IndexByte
func abigen_strings_IndexByte(s string, c byte) int
