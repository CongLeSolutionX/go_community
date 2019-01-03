// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that we don't consider a Go'd function's
// arguments as pointers when they aren't.

package main

import (
	"unsafe"
)

var badPtr uintptr

var sink []byte

func init() {
	// Allocate large enough to use largeAlloc.
	b := make([]byte, 1<<16-1)
	sink = b // force heap allocation
	//  Any space between the object and the end of page is invalid to point to.
	badPtr = uintptr(unsafe.Pointer(&b[len(b)-1])) + 1
}

var throttle = make(chan struct{}, 10)

func noPointerArgs(p *byte, a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15, a16, a17, a18, a19, a20, a21, a22, a23, a24, a25, a26, a27, a28, a29, a30, a31, a32 uintptr) {
	sink = make([]byte, 4096)
	<-throttle
}

func main() {
	const N = 1000
	for i := 0; i < N; i++ {
		throttle <- struct{}{}
		go noPointerArgs(nil, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr, badPtr)
		sink = make([]byte, 4096)
	}
}
