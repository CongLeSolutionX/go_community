// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuzz

import (
	"internal/unsafeheader"
	"unsafe"
)

// coverage returns a []byte containing unique 8-bit counters for each edge of
// the instrumented source code. This coverage data will only be generated if
// `-d=libfuzzer` is set at build time. This can be used to understand the code
// coverage of a test execution.
func coverage() []byte {
	addr := unsafe.Pointer(&_counters)
	size := uintptr(unsafe.Pointer(&_ecounters)) - uintptr(addr)

	var res []byte
	*(*unsafeheader.Slice)(unsafe.Pointer(&res)) = unsafeheader.Slice{
		Data: addr,
		Len:  int(size),
		Cap:  int(size),
	}
	return res
}

// ResetCovereage sets all of the counters for each edge of the instrumented
// source code to 0.
func ResetCoverage() {
	cov := coverage()
	for i := range cov {
		cov[i] = 0
	}
}

// SnapshotCoverage copies the current counter values into coverageSnapshot,
// preserving them for later inspection.
func SnapshotCoverage() {
	cov := coverage()
	if coverageSnapshot == nil {
		coverageSnapshot = make([]byte, len(cov))
	}
	copy(coverageSnapshot, cov)
	bucketCounters(coverageSnapshot)
}

func countEdges(cov []byte) int {
	n := 0
	for _, c := range cov {
		if c > 0 {
			n++
		}
	}
	return n
}

// bucketCounters quantizes coverage counters into a series of buckets. The buckets
// are chosen such that a counter jumping between buckets can be considered an indication
// of a significant change in the execution trace of a fuzz target.
func bucketCounters(cov []byte) {
	for i, c := range cov {
		switch {
		case c == 0 || c == 1 || c == 2:
			continue
		case c <= 4:
			cov[i] = 4
		case c <= 8:
			cov[i] = 8
		case c <= 16:
			cov[i] = 16
		case c <= 32:
			cov[i] = 32
		case c <= 64:
			cov[i] = 64
		case c <= 128:
			cov[i] = 128
		case c <= 255:
			cov[i] = 255
		}
	}
}

var coverageSnapshot []byte

// _counters and _ecounters mark the start and end, respectively, of where
// the 8-bit coverage counters reside in memory. They're known to cmd/link,
// which specially assigns their addresses for this purpose.
var _counters, _ecounters [0]byte
