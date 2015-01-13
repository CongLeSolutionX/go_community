// +build s390x
// run nilptr.go nilptr_s390_common.go

// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Platform specific test configuration

package main

// Having a big address space means that indexing at a large
// offset from a nil pointer might not cause a memory access
// fault.  This test checks that Go is doing the correct explicit
// checks to catch these nil pointer accesses, not just relying on
// the hardware.
//
// Give us a big address space somewhere near minBssOffset.
const inMemSize uintptr = 256 << 20 // 256 MiB
const minBssOffset uintptr = 1 << 31 // 0x80000000
const maxlen uintptr = (1 << 32) - 1 // 0xffffffff
const inMaxlenArray uintptr = minBssOffset + inMemSize / 2
var dummy [inMemSize]byte
