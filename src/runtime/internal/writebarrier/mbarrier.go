// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: write barriers.
//
// For the concurrent garbage collector, the Go compiler implements
// updates to pointer-valued fields that may be in heap objects by
// emitting calls to write barriers. This file contains the actual write barrier
// implementation, markwb, and the various wrappers called by the
// compiler to implement pointer assignment, slice assignment,
// typed memmove, and so on.

package writebarrier

import (
	_iface "runtime/internal/iface"
)

//go:nosplit
func writebarrierstring(dst *[2]uintptr, src [2]uintptr) {
	_iface.Writebarrierptr(&dst[0], src[0])
	dst[1] = src[1]
}

//go:nosplit
func writebarrierslice(dst *[3]uintptr, src [3]uintptr) {
	_iface.Writebarrierptr(&dst[0], src[0])
	dst[1] = src[1]
	dst[2] = src[2]
}

//go:nosplit
func writebarrieriface(dst *[2]uintptr, src [2]uintptr) {
	_iface.Writebarrierptr(&dst[0], src[0])
	_iface.Writebarrierptr(&dst[1], src[1])
}
