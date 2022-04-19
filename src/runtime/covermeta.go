// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// covmetablob is a container for holding the meta-data symbol (an
// RODATA variable) for an instrumented Go package. Here "P" points to
// the symbol itself, "Len" is the length of the sym in bytes, and
// "Hash" is an md5sum for the sym computed by the compiler.
type covmetablob struct {
	p       unsafe.Pointer
	len     uint32
	hash    [16]byte
	pkgpath string
	pkid    int
	cmode   uint8 // coverage.CounterMode
}

// covmetalist contains the list of currently registered meta-data
// blobs for the running program.
var covmetalist []covmetablob

// covpkgmap records mappings from hard-coded package IDs to
// slots in the covmetalist above.
var covpkgmap map[int]int

var hardcodedListNeedsUpdating bool

func reportInsanityInHardcodedList() {
	println("internal error in coverage meta-data tracking:")
	println("list of hard-coded runtime package IDs needs revising.")
	println("[see the comment on the 'rtPkgs' var in ")
	println(" <goroot>/src/internal/coverage/pkid.go]")
	println("registered list:")
	for k, b := range covmetalist {
		print("slot: ", k, " path='", b.pkgpath, "' ")
		if b.pkid != -1 {
			print(" hard-coded id: ", b.pkid)
		}
		println("")
	}
}

// addcovmeta is invoked during package "init" functions by the
// compiler when compiling for coverage instrumentation; here 'p' is a
// meta-data blob of length 'dlen' for the package in question, 'hash'
// is a compiler-computed md5.sum for the blob, 'pkpath' is the
// package path, 'pkid' is the hard-coded ID that the compiler is
// using for the package (or -1 if the compiler doesn't think a
// hard-coded ID is needed), and 'cmode' is the coverage counter mode
// requested by the user. Return value is the ID for the package for
// use by the package code itself.
func addcovmeta(p unsafe.Pointer, dlen uint32, hash [16]byte, pkpath string, pkid int, cmode uint8) uint32 {
	slot := len(covmetalist)
	covmetalist = append(covmetalist,
		covmetablob{
			p:       p,
			len:     dlen,
			hash:    hash,
			pkgpath: pkpath,
			pkid:    pkid,
			cmode:   cmode,
		})
	if pkid != -1 {
		if covpkgmap == nil {
			covpkgmap = make(map[int]int)
		}
		if _, ok := covpkgmap[pkid]; ok {
			throw("runtime.addcovmeta: fatal error: covpkgmap collision")
		}
		// Record the real slot (position on meta-list) for this
		// package. We'll use the map to fix things up later on.
		covpkgmap[pkid] = slot
	}

	// ID zero is reserved as invalid.
	return uint32(slot + 1)
}
