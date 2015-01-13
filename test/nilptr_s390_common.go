// skip

// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Shared s390 and s390x specific test configuration

package main

import "unsafe"

func sanityCheck() {
	// The test only tests what we intend to test if dummy
	// starts near minBssOffset uintptr.  Otherwise there
	// might not be anything mapped at the address that might
	// be accidentally dereferenced below.
	if uintptr(unsafe.Pointer(&dummy)) > inMemSize + minBssOffset {
		panic("dummy too far out")
	} else if uintptr(unsafe.Pointer(&dummy)) < minBssOffset {
		panic("dummy too close")
	}
}
