// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cipher

import (
	"runtime"
	"unsafe"
)

const ptrSize = int(unsafe.Sizeof(uintptr(0)))
const supportsUnaligned = runtime.GOARCH == "386" || runtime.GOARCH == "amd64"

// ppc64:  possibly word aligned (4 bytes) can be done efficiently too
const checkAlignment = runtime.GOARCH == "ppc64le" || runtime.GOARCH == "ppc64"

// fastXORBytes xors in bulk. It only works on architectures that
// support unaligned read/writes.
func fastXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	w := n / ptrSize
	if w > 0 {
		dw := *(*[]uintptr)(unsafe.Pointer(&dst))
		aw := *(*[]uintptr)(unsafe.Pointer(&a))
		bw := *(*[]uintptr)(unsafe.Pointer(&b))
		for i := 0; i < w; i++ {
			dw[i] = aw[i] ^ bw[i]
		}
	}

	for i := (n - n%ptrSize); i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}

	return n
}

// Determine efficient alignment.
func efficientAlignment(dst, a, b []byte) bool {
	// find alignment with respect to ptrSize
	da := int(uintptr(unsafe.Pointer(&dst[0]))) & (ptrSize - 1)
	aa := int(uintptr(unsafe.Pointer(&a[0]))) & (ptrSize - 1)
	ba := int(uintptr(unsafe.Pointer(&b[0]))) & (ptrSize - 1)

	// If any are not aligned to 0, don't try
	if da|ba|aa != 0 {
		return false
	}
	return true
}

func safeXORBytes(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

// xorBytes xors the bytes in a and b. The destination is assumed to have enough
// space. Returns the number of bytes xor'd.
func xorBytes(dst, a, b []byte) int {
	if supportsUnaligned {
		return fastXORBytes(dst, a, b)
	} else {
		// TODO(hanwen): if (dst, a, b) have common alignment
		// we could still try fastXORBytes. It is not clear
		// how often this happens, and it's only worth it if
		// the block encryption itself is hardware
		// accelerated.

		// LAB:  This has been done for Power, not sure
		// how this applies to other GOARCHes, so I didn't
		// remove the TODO
		if checkAlignment && efficientAlignment(dst, a, b) {
			return fastXORBytes(dst, a, b)
		}
		return safeXORBytes(dst, a, b)
	}
}

// fastXORWords XORs multiples of 4 or 8 bytes (depending on architecture.)
// The arguments are assumed to be of equal length.
func fastXORWords(dst, a, b []byte) {
	dw := *(*[]uintptr)(unsafe.Pointer(&dst))
	aw := *(*[]uintptr)(unsafe.Pointer(&a))
	bw := *(*[]uintptr)(unsafe.Pointer(&b))
	n := len(b) / ptrSize
	for i := 0; i < n; i++ {
		dw[i] = aw[i] ^ bw[i]
	}
}

func xorWords(dst, a, b []byte) {
	if supportsUnaligned {
		fastXORWords(dst, a, b)
	} else {
		if checkAlignment && efficientAlignment(dst, a, b) {
			fastXORBytes(dst, a, b)
		} else {
			safeXORBytes(dst, a, b)
		}
	}
}
