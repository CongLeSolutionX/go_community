// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build 386 amd64 amd64p32 s390x arm arm64 ppc64 ppc64le mips mipsle wasm

package bytealg

import _ "unsafe" // For go:linkname

//go:noescape
func Compare(a, b []byte) int

// The following are defined in assembly in this package, but exported
// to other packages. Provide Go declarations to go with their
// assembly definitions.

//go:linkname bytes_Compare bytes.Compare
func bytes_Compare(a, b []byte) int

//go:linkname runtime_cmpstring runtime.cmpstring
func runtime_cmpstring(a, b string) int
