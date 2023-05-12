// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64.v3

package runtime

import "unsafe"

// This exists for architectures which require a different implementation to clear pointers.

// memclrPointers takes in a number of pointer to clear instead of number of bytes.
//
//go:noescape
func memclrPointers(ptr unsafe.Pointer, n uintptr)

//go:nosplit
func doMemclrWithPointers(ptr unsafe.Pointer, n uintptr) {
	bulk := n / unsafe.Sizeof(ptr)
	memclrPointers(ptr, bulk)
	ptr = unsafe.Add(ptr, bulk)
	if n&4 != 0 {
		memclrNoHeapPointers(ptr, 4)
		ptr = unsafe.Add(ptr, 4)
	}
	if n&2 != 0 {
		memclrNoHeapPointers(ptr, 2)
		ptr = unsafe.Add(ptr, 2)
	}
	if n&1 != 0 {
		memclrNoHeapPointers(ptr, 1)
	}
}
