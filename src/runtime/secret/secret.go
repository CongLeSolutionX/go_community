// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret

import (
	"runtime"
)

// Do invokes f.
//
// Do ensures that any temporary storage used by f is erased in a
// timely manner. (In this context, "f" is shorthand for the
// entire call tree initiated by f.)
//  - Any registers used by f are erased before Do returns.
//  - Any stack used by f is erased before Do returns.
//  - Any heap allocation done by f is erased as soon as the garbage
//    collector realizes that it is no longer reachable.
//  - Do works even if f panics or calls runtime.Goexit.  As part of
//    that, any panic raised by f will appear as if it originates from
//    Do itself.
//
// Limitations:
//  - Currently works only on amd64 and arm64.  On unsupported
//    platforms Do will immediately panic.
//  - Protection does not extend to any new goroutines made by f.
//  - Protection does not extend to any global variables written by f.
//  - If f calls runtime.Goexit, erasure can be delayed by defers
//    higher up on the call stack.
//  - Heap allocations will only be erased if the program drops all
//    references to those allocations, and then the garbage collector
//    notices that those references are gone. The former is under
//    control of the program, but the latter is at the whim of the
//    runtime.
//  - Any value panicked by f may point to allocations from within
//    f. Those allocations will not be erased until (at least) the
//    panicked value is dead.
func Do(f func()) {
	switch runtime.GOARCH {
	default:
		panic("unsupported architecture")
	case "amd64", "arm64":
	}

	// Place to store any panic value.
	var p any

	// Step 1: increment the nesting count.
	inc()

	// Step 2: call helper. The helper just calls f
	// and captures (recovers) any panic result.
	p = doHelper(f)

	// Step 3: decrement the nesting count.
	dec()

	// Step 4: erase everything used by f (stack, registers).
	eraseSecrets()

	// Step 5: re-raise any caught panic.
	// This will make the panic appear to come
	// from a stack whose bottom frame is
	// runtime/secret.Do.
	// Anything below that to do with f will be gone.
	//
	// Note that the panic value is not erased. It behaves
	// like any other value that escapes from f. If it is
	// heap allocated, it will be erased when the garbage
	// collector notices it is no longer referenced.
	if p != nil {
		panic(p)
	}

	// Note: if f calls runtime.Goexit, step 3 and above will not
	// happen, as Goexit is unrecoverable. We handle that case in
	// runtime/proc.go:goexit0.
}

func doHelper(f func()) (p any) {
	// Step 2b: Pop the stack up to the secret.doHelper frame
	// if we are in the process of panicking.
	// (It is a no-op if we are not panicking.)
	// We return any panicked value to secret.Do, who will
	// re-panic it.
	defer func() {
		// Note: we rely on the go1.21+ behavior that
		// if we are panicking, recover returns non-nil.
		p = recover()
	}()

	// Step 2a: call the secret function.
	f()

	return
}

// Enabled reports whether [Do] appears anywhere on the call stack.
func Enabled() bool {
	return count() > 0
}

// implemented in runtime
func count() uint32
func inc()
func dec()
func eraseSecrets()
func getStack() (uintptr, uintptr) // for testing
