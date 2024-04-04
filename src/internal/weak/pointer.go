// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
The weak package is a package for managing weak references.
The implementation is general, but not efficient, since it
requires at least one heap allocation to create a weak pointer
for a single value. It is designed primarily to support the
narrow use-case of the unique package, where we're explicitly
deduplicating weak pointers by value, limiting their scope.
*/
package weak

import (
	"internal/abi"
	"runtime"
	"unsafe"
)

// Pointer is a weak pointer to a value of type T.
//
// This value is comparable and is guaranteed to compare equal
// for multiple pointers created for the same value.
type Pointer[T any] struct {
	u unsafe.Pointer
}

// Make creates a weak pointer from a strong pointer to some value of
// type T. ptr must not be an interior pointer, and Make will panic if
// it is. ptr MUST NOT have a finalizer, and must never get one after
// this point.
func Make[T any](ptr *T) Pointer[T] {
	// Explicitly force ptr to escape to the heap.
	ptr = abi.Escape(ptr)

	var u unsafe.Pointer
	if ptr != nil {
		u = runtime_registerWeakPointer(unsafe.Pointer(ptr))
	}
	runtime.KeepAlive(ptr)
	return Pointer[T]{u}
}

// Strong creates a strong pointer from the weak pointer. Returns nil
// if the original value for the weak pointer was garbage collected.
func (p Pointer[T]) Strong() *T {
	return (*T)(runtime_makeStrongFromWeak(unsafe.Pointer(p.u)))
}

// Implemented in runtime.

//go:linkname runtime_registerWeakPointer
func runtime_registerWeakPointer(unsafe.Pointer) unsafe.Pointer

//go:linkname runtime_makeStrongFromWeak
func runtime_makeStrongFromWeak(unsafe.Pointer) unsafe.Pointer
