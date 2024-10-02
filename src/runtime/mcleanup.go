// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/abi"
	"unsafe"
)

// AddCleanup attaches a cleanup function to ptr. Some time after ptr is no longer
// reachable, the runtime will (probably) call cleanup(arg) in a separate goroutine.
//
// If ptr is reachable from cleanup or arg, ptr will never be collected
// and the cleanup will never run. AddCleanup panics if arg == ptr.
//
// The cleanup(arg) call is not always guaranteed to run; in particular it is not
// guaranteed to run before program exit.
//
// A single goroutine runs all cleanup calls for a program, sequentially. If a cleanup
// function must run for a long time, it should create a new goroutine.
//
// TODO(amedee) cleanups must happen after finalizers
// TODO(amedee) cleanups may run in a new goroutine
func AddCleanup[T, S any](ptr *T, cleanup func(S), arg S) Cleanup {
	ptr = abi.Escape(ptr)
	cleanup = abi.Escape(cleanup)
	arg = abi.Escape(arg)

	if ptr == nil {
		throw("runtime.AddCleanup: ptr is nil")
	}
	usptr := uintptr(unsafe.Pointer(ptr))

	// ensure that arg is not equal to the ptr
	if unsafe.Pointer(&arg) == unsafe.Pointer(ptr) {
		throw("runtime.AddCleanup: ptr is equal to arg")
	}
	if inUserArenaChunk(usptr) {
		// Arena-allocated objects are not eligible for cleanup.
		throw("runtime.AddCleanup: ptr was allocated into an arena")
	}
	if debug.sbrk != 0 {
		// debug.sbrk never frees memory, so no finalizers run
		// (and we don't have the data structures to record them).
		// return a noop cleanup.
		return Cleanup{}
	}

	var fn func()
	fn = func() {
		cleanup(arg)
	}
	// closure must escape
	fn = abi.Escape(fn)
	fv := *(**funcval)(unsafe.Pointer(&fn))
	fv = abi.Escape(fv)

	base, _, _ := findObject(usptr, 0, 0)
	if base == 0 {
		if isGoPointerWithoutSpan(unsafe.Pointer(ptr)) {
			return Cleanup{}
		}
		throw("runtime.AddCleanup: ptr not in allocated block")
	}

	// ensure we have a finalizer goroutine running.
	createfing()

	systemstack(func() {
		addCleanup(unsafe.Pointer(ptr), fv)
	})

	KeepAlive(arg)
	KeepAlive(cleanup)
	KeepAlive(usptr)
	KeepAlive(fn)
	KeepAlive(fv)
	return Cleanup{}
}

// Cleanup is a handle to a cleanup call for a specific object.
type Cleanup struct{}

// Stop cancels the cleanup call. Stop will have no effect if the cleanup call
// has already been queued for execution (because ptr became unreachable).
// To guarantee that Stop removes the cleanup function, the caller must ensure
// that the pointer that was passed to AddCleanup is reachable across the call to Stop.
//
// TODO(amedee) do not work on stop until after the rest of the implementation is complete.
func (c Cleanup) Stop() {}
