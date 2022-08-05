// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build (!amd64 && !arm64 && !ppc64 && !ppc64le) || testgenericxor

package subtle

import (
	"runtime"
	"unsafe"
)

const wordSize = unsafe.Sizeof(uintptr(0))

// GOARCH==amd64 is only here for better coverage
// when testing with 'go test -tags testgenericxor'.
const supportsUnaligned = runtime.GOARCH == "386" ||
	runtime.GOARCH == "amd64" ||
	runtime.GOARCH == "s390x"

func xorBytes(dst, x, y *byte, n int) {
	if supportsUnaligned || aligned(dst, x, y, uintptr(n)) {
		nw := uintptr(n) / wordSize
		xorWordsLoop(
			(*uintptr)(unsafe.Pointer(dst)),
			(*uintptr)(unsafe.Pointer(x)),
			(*uintptr)(unsafe.Pointer(y)),
			int(nw))
		nb := nw * wordSize
		if uintptr(n) == nb {
			return
		}
		n -= int(nb)
		dst = (*byte)(unsafe.Add(unsafe.Pointer(dst), nb))
		x = (*byte)(unsafe.Add(unsafe.Pointer(x), nb))
		y = (*byte)(unsafe.Add(unsafe.Pointer(y), nb))
	}
	xorBytesLoop(dst, x, y, int(n))
}

func aligned(dst, x, y *byte, n uintptr) bool {
	return (uintptr(unsafe.Pointer(dst))|uintptr(unsafe.Pointer(x))|uintptr(unsafe.Pointer(y))|n)&(wordSize-1) == 0
}

func xorWordsLoop(dst, x, y *uintptr, n int) {
	dstw := unsafe.Slice(dst, n)
	xw := unsafe.Slice(x, n)
	yw := unsafe.Slice(y, n)
	for i := 0; i < n; i++ {
		dstw[i] = xw[i] ^ yw[i]
	}
}

func xorBytesLoop(dst, x, y *byte, n int) {
	dstb := unsafe.Slice(dst, n)
	xb := unsafe.Slice(x, n)
	yb := unsafe.Slice(y, n)
	for i := 0; i < n; i++ {
		dstb[i] = xb[i] ^ yb[i]
	}
}
