// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// covCounterBlob is a container for encapsulating a counter section
// (BSS variable) for an instrumented Go module. Here "counters"
// points to the counter payload and "len" is the number of uint32
// entries in the section.
type covCounterBlob struct {
	// Important: any changes to this struct should also be made in
	// the runtime/coverage/emit.go, since that code expects a
	// specific format here.
	counters *uint32
	len      uint64
}

// getCovCounterList returns a list of blobs storing pointers to the
// counter segments for the currently running coverage-instrumented
// program.
func getCovCounterList() []covCounterBlob {
	res := []covCounterBlob{}
	u32sz := unsafe.Sizeof(uint32(0))
	for datap := &firstmoduledata; datap != nil; datap = datap.next {
		if datap.covctrs == datap.ecovctrs {
			continue
		}
		res = append(res, covCounterBlob{
			counters: (*uint32)(unsafe.Pointer(datap.covctrs)),
			len:      uint64((datap.ecovctrs - datap.covctrs) / u32sz),
		})
	}
	return res
}
