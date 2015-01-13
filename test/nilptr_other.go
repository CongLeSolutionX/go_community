// +build !s390 !s390x
// run nilptr.go

// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Platform specific test configuration

package main

// Having a big address space means that indexing
// at a 256 MB offset from a nil pointer might not
// cause a memory access fault. This test checks
// that Go is doing the correct explicit checks to catch
// these nil pointer accesses, not just relying on the hardware.
const inMemSize uintptr = 256 << 20 // 256 MiB
const maxlen uintptr = (1 << 30) - 2 // 0x40000000
const inMaxlenArray uintptr = 256 << 20
var dummy [256 << 20]byte // give us a big address space

func sanityCheck() {
	// the test only tests what we intend to test
	// if dummy starts in the first 256 MB of memory.
	// otherwise there might not be anything mapped
	// at the address that might be accidentally
	// dereferenced below.
	if uintptr(unsafe.Pointer(&dummy)) > 256<<20 {
		panic("dummy too far out")
	}
}
