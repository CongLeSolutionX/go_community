// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"unsafe"
)

// covmetablob is a container for holding the meta-data symbol (an
// RODATA variable) for an instrumented Go package. Here "P" points to
// the symbol itself, "Len" is the length of the sym in bytes, and
// "Hash" is an md5sum for the sym computed by the compiler.
type covmetablob struct {
	// Important: any changes to this struct should also be made in
	// the runtime/coverage/emit.go, since that code expects a
	// specific format here.
	p       *byte
	len     uint32
	hash    [16]byte
	pkgpath string
	pkid    int
	cmode   uint8 // coverage.CounterMode
}

var covmetalist []covmetablob

var covpkgmap map[int]int

var hardcodedListNeedsUpdating bool

func reportInsanityInHardcodedList(pkgId int32) {
	println("internal error in coverage meta-data tracking:")
	println("encountered bad pkg ID ", pkgId)
	println("list of hard-coded runtime package IDs needs revising.")
	println("[see the comment on the 'rtPkgs' var in ")
	println(" <goroot>/src/cmd/compile/internal/coverage/coverage.go]")
	println("registered list:")
	for k, b := range covmetalist {
		print("slot: ", k, " path='", b.pkgpath, "' ")
		if b.pkid != -1 {
			print(" hard-coded id: ", b.pkid)
		}
		println("")
	}
	println("remap table:")
	for from, to := range covpkgmap {
		println("from ", from, " to ", to)
	}
}

// addcovmeta is invoked during package "init" functions by the
// compiler when compiling for coverage instrumentation; here 'p' is a
// meta-data blob of length 'dlen' for the package in question, 'hash'
// is a compiler-computed md5.sum for the blob, 'pkpath' is the
// package path, and 'pkid' is the hard-coded ID that the compiler is
// using for the package (or -1 if the compiler doesn't think a
// hard-coded ID is needed). Return value is the ID for the package.
func addcovmeta(p unsafe.Pointer, dlen uint32, hash [16]byte, pkpath string, pkid int, cmode uint8) uint32 {
	slot := len(covmetalist)
	covmetalist = append(covmetalist,
		covmetablob{
			p:       (*byte)(p),
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

func getcovmetalist() []covmetablob {
	return covmetalist
}

// covcounterblob is a container for encapsulating a counter section
// (BSS variable) for an instrumented Go module. Here "Counters" points to
// the counter payload and Len is the number of uint32 entries in the
// section.
type covcounterblob struct {
	// Important: any changes to this struct should also be made in
	// the runtime/coverage/emit.go, since that code expects a
	// specific format here.
	Counters *uint32
	Len      uint64
}

func getcovcounterlist() []covcounterblob {
	res := []covcounterblob{}
	u32sz := unsafe.Sizeof(uint32(0))
	for datap := &firstmoduledata; datap != nil; datap = datap.next {
		if datap.covctrs == datap.ecovctrs {
			continue
		}
		res = append(res, covcounterblob{
			Counters: (*uint32)(unsafe.Pointer(datap.covctrs)),
			Len:      uint64((datap.ecovctrs - datap.covctrs) / u32sz),
		})
	}
	return res
}

func getcovpkgmap() map[int]int {
	return covpkgmap
}
