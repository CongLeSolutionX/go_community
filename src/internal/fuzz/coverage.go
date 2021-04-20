// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuzz

import (
	"fmt"
	"internal/unsafeheader"
	"unsafe"
)

// coverageCopy returns a copy of the current bytes provided by coverage().
func coverageCopy() []byte {
	cov := coverage()
	ret := make([]byte, len(cov))
	copy(ret, cov)
	return ret
}

// resetCovereage sets all of the counters for each edge of the instrumented
// source code to 0.
func resetCoverage() {
	cov := coverage()
	for i := range cov {
		cov[i] = 0
	}
}

// expandsCoverage returns true if newCov expands coverage (ie. a new edge was
// hit). If doUpdate is true, then for every edge, if newCov shows a counter
// value that's larger than the current counter value in cov, then cov will be
// updated with this larger value.
func expandsCoverage(cov, newCov []byte, doUpdate bool) bool {
	if len(newCov) != len(cov) {
		panic(fmt.Sprintf("num edges changed at runtime: %d, expected %d", len(newCov), len(cov)))
	}
	newEdge := false
	for i := range cov {
		if newCov[i] > cov[i] {
			if cov[i] == 0 {
				newEdge = true
				if !doUpdate {
					// A new edge was hit, and cov does not need to be updated, so
					// return early indicating that newCov expands coverage.
					return true

				}
			}
			cov[i] = newCov[i]
		}
	}
	return newEdge
}

// coverage returns a []byte containing unique 8-bit counters for each edge of
// the instrumented source code. This coverage data will only be generated if
// `-d=libfuzzer` is set at build time. This can be used to understand the code
// coverage of a test execution.
func coverage() []byte {
	addr := unsafe.Pointer(&_counters)
	size := uintptr(unsafe.Pointer(&_ecounters)) - uintptr(addr)
	if size == 0 {
		// Test binary was built on a platform that doesn't support coverage
		// instrumentation.
		return []byte{}
	}

	var res []byte
	*(*unsafeheader.Slice)(unsafe.Pointer(&res)) = unsafeheader.Slice{
		Data: addr,
		Len:  int(size),
		Cap:  int(size),
	}
	return res
}

// _counters and _ecounters mark the start and end, respectively, of where
// the 8-bit coverage counters reside in memory. They're known to cmd/link,
// which specially assigns their addresses for this purpose.
var _counters, _ecounters [0]byte
