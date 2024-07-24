// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret

import (
	"internal/cpu"
	"unsafe"
)

func Do(f func()) {
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

func Enabled() bool {
	return count() > 0
}

// implemented in runtime
func count() uint32
func inc()
func dec()
func eraseSecrets()
func getStack() (uintptr, uintptr) // for testing

// Assembly in this package (for testing only)

// load data from p into test registers.
//
//go:noescape
func loadRegisters(p unsafe.Pointer)

// spill data from test registers into p.
// Returns the amount of space filled in.
//
//go:noescape
func spillRegisters(p unsafe.Pointer) uintptr

const (
	offsetX86HasAVX     = unsafe.Offsetof(cpu.X86.HasAVX)
	offsetX86HasAVX512F = unsafe.Offsetof(cpu.X86.HasAVX512F)
)
