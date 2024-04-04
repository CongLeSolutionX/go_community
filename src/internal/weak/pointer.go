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
	"sync/atomic"
	"unsafe"
)

// Pointer is a weak pointer to a value of type T.
//
// This value is comparable and is guaranteed to compare equal
// for multiple pointers created for the same value.
type Pointer[T any] struct {
	u *atomic.Uint64
}

// Make creates a weak pointer from a strong pointer to some value of
// type T. ptr must not be an interior pointer, and Make will panic if
// it is. ptr MUST NOT have a finalizer, and must never get one after
// this point.
func Make[T any](ptr *T) Pointer[T] {
	// Explicitly force ptr to escape to the heap.
	ptr = abi.Escape(ptr)

	var u *atomic.Uint64
	if ptr != nil {
		u = (*atomic.Uint64)(runtime_registerWeakPointer(unsafe.Pointer(ptr)))
	}
	runtime.KeepAlive(ptr)
	return Pointer[T]{u}
}

// Strong creates a strong pointer from the weak pointer. Returns nil
// if the original value for the weak pointer was garbage collected.
func (p Pointer[T]) Strong() *T {
	// We must be non-preemptible between loading the weak pointer
	// and turning it into a strong pointer. Otherwise there's a
	// chance that if we get descheduled at just the wrong point
	// for more than one GC cycle (the wrong point being right in
	// between the load and the conversion to the strong pointer),
	// we may create a strong pointer from a now-invalid pointer.
	runtime_procPin()
	for {
		wp := pointer[T](p.u.Load())
		if !wp.dying() {
			ptr := wp.strong()
			runtime_procUnpin()
			return ptr
		}
		if p.u.CompareAndSwap(uint64(wp), uint64(wp.resurrect())) {
			ptr := wp.strong()
			runtime_procUnpin()
			return ptr
		}
	}
}

// Nil returns true if the pointer is nil. This is useful for checking
// if a weak pointer is dead without creating a strong pointer, which
// may potentially keep the weak pointer live if called to often just
// to check.
//
// This condition is stable in that once it begins returning true, it
// will return true forever more.
func (p Pointer[T]) Nil() bool {
	return p.u.Load() == 0
}

const dyingBit = 1 << 63

// pointer represents the weak pointer state machine. Weak pointers
// degrade from regular pointers, to dying pointers, and finally to
// nil. The garbage collector in the runtime understands this state
// machine as well.
//
// Creating a strong pointer from a dying pointer will resurrect
// the object and make the corresponding weak pointer a regular
// pointer.
type pointer[T any] uint64

func (p pointer[T]) dying() bool {
	return uint64(p)&dyingBit != 0
}

func (p pointer[T]) resurrect() pointer[T] {
	return pointer[T](uint64(p &^ dyingBit))
}

func (p pointer[T]) strong() *T {
	return (*T)(unsafe.Pointer(uintptr(uint64(p &^ dyingBit))))
}

// Implemented in runtime.

//go:linkname runtime_registerWeakPointer
func runtime_registerWeakPointer(unsafe.Pointer) unsafe.Pointer

//go:linkname runtime_procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin
func runtime_procUnpin()
