// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

type covmetablob struct {
	p    unsafe.Pointer
	len  uint32
	hash [16]byte
}

var covmetalist []covmetablob

// addcovmeta is invoked during package "init" functions by the
// compiler when compiling for coverage instrumentation; here 'p' is a
// meta-data blob of length 'dlen' for the package in question, 'hash'
// is a compiler-computed md5.sum for the blob, 'pkpath' is the
// package path, and 'pkid' is the hard-coded ID that the compiler is
// using for the package (or -1 if the compiler doesn't think a
// hard-coded ID is needed). Return value is the ID for the package.
func addcovmeta(p unsafe.Pointer, dlen uint32, hash [16]byte, pkpath string, pkid int) uint32 {
	slot := len(covmetalist)
	covmetalist = append(covmetalist,
		covmetablob{
			p:    p,
			len:  dlen,
			hash: hash,
		})
	if pkid != -1 && pkid-1 != slot {
		// If this assert is firing, it means that the compiler's
		// snapshot of the expected runtime package dependencies has
		// changed (maybe something has been added or deleted), and
		// the compiler needs to be updated. See the hardCodedPkgId
		// code in cmd/compile/internal/coverage/coverage.go.
		println("reserved runtime package ID clash on '",
			pkpath, "': slot='", slot, "' pkid='", pkid)
		throw("runtime.addcovmeta: fatal error")
	}

	// ID zero is reserved as invalid.
	return uint32(slot + 1)
}
