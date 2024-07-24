// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64 || amd64

package secret_test

import (
	"runtime"
	"runtime/secret"
	"strings"
	"testing"
	"time"
	"unsafe"
)

type secretType int64

const secretValue = 0x53c237_53c237

// S is a type that might have some secrets in it.
type S [100]secretType

// makeS makes an S with secrets in it.
//
//go:noinline
func makeS() S {
	// Note: noinline ensures this doesn't get inlined and
	// completely optimized away.
	var s S
	for i := range s {
		s[i] = secretValue
	}
	return s
}

// heapS allocates an S on the heap with secrets in it.
//
//go:noinline
func heapS() *S {
	// Note: noinline forces heap allocation
	s := makeS()
	return &s
}

// Test that when we allocate inside secret.Do, the resulting
// allocations are zeroed by the garbage collector when they
// are freed.
// See runtime/mheap.go:freeSpecial.
func TestHeap(t *testing.T) {
	var u uintptr
	secret.Do(func() {
		u = uintptr(unsafe.Pointer(heapS()))
	})

	runtime.GC()

	// Check that object got zeroed.
	checkRangeForSecret(t, u, u+unsafe.Sizeof(S{}))
	// Also check our stack, just because we can.
	checkStackForSecret(t)
}

// Test that when we return from secret.Do, we zero the stack used
// by the argument to secret.Do.
// See runtime/secret.go:secret_dec.
func TestStack(t *testing.T) {
	checkStackForSecret(t) // if this fails, something is wrong with the test

	secret.Do(func() {
		s := makeS()
		use(&s)
	})

	checkStackForSecret(t)
}

//go:noinline
func use(s *S) {
	// Note: noinline prevents dead variable elimination.
}

// Test that when we copy a stack, we zero the old one.
// See runtime/stack.go:copystack.
func TestStackCopy(t *testing.T) {
	checkStackForSecret(t) // if this fails, something is wrong with the test

	var lo, hi uintptr
	secret.Do(func() {
		// Put some secrets on the current stack frame.
		s := makeS()
		use(&s)
		// Remember the current stack.
		lo, hi = secret.GetStack()
		// Use a lot more stack to force a stack copy.
		growStack()
	})
	checkRangeForSecret(t, lo, hi) // pre-grow stack
	checkStackForSecret(t)         // post-grow stack (just because we can)
}

func growStack() {
	growStack1(1000)
}
func growStack1(n int) {
	if n == 0 {
		return
	}
	growStack1(n - 1)
}

func TestPanic(t *testing.T) {
	checkStackForSecret(t) // if this fails, something is wrong with the test

	defer func() {
		checkStackForSecret(t)

		p := recover()
		if p == nil {
			t.Errorf("panic squashed")
			return
		}
		var e error
		var ok bool
		if e, ok = p.(error); !ok {
			t.Errorf("panic not an error")
		}
		if !strings.Contains(e.Error(), "divide by zero") {
			t.Errorf("panic not a divide by zero error: %s", e.Error())
		}
		var pcs [10]uintptr
		n := runtime.Callers(0, pcs[:])
		frames := runtime.CallersFrames(pcs[:n])
		for {
			frame, more := frames.Next()
			if strings.Contains(frame.Function, "dividePanic") {
				t.Errorf("secret function in traceback")
			}
			if !more {
				break
			}
		}
	}()
	secret.Do(dividePanic)
}

func dividePanic() {
	s := makeS()
	use(&s)
	_ = 8 / zero
}

var zero int

func TestGoExit(t *testing.T) {
	checkStackForSecret(t) // if this fails, something is wrong with the test

	c := make(chan uintptr, 2)

	go func() {
		// Run the test in a separate goroutine
		defer func() {
			// Tell original goroutine what our stack is
			// so it can check it for secrets.
			lo, hi := secret.GetStack()
			c <- lo
			c <- hi
		}()
		secret.Do(func() {
			s := makeS()
			use(&s)
			runtime.Goexit()
		})
		t.Errorf("goexit didn't happen")
	}()
	lo := <-c
	hi := <-c
	// We want to wait until the other goroutine has finished Goexiting and
	// cleared its stack. There's no signal for that, so just wait a bit.
	time.Sleep(1 * time.Millisecond)

	checkRangeForSecret(t, lo, hi)
}

func checkStackForSecret(t *testing.T) {
	t.Helper()
	lo, hi := secret.GetStack()
	checkRangeForSecret(t, lo, hi)
}
func checkRangeForSecret(t *testing.T, lo, hi uintptr) {
	t.Helper()
	for p := lo; p < hi; p += unsafe.Sizeof(secretType(0)) {
		v := secretType(secret.Read(p))
		if v == secretValue {
			t.Errorf("secret found in [%x,%x] at %x", lo, hi, p)
		}
	}
}

func TestRegisters(t *testing.T) {
	secret.Do(func() {
		s := makeS()
		secret.LoadRegisters(unsafe.Pointer(&s))
	})
	var spillArea [64]secretType
	n := secret.SpillRegisters(unsafe.Pointer(&spillArea))
	if n > unsafe.Sizeof(spillArea) {
		t.Fatalf("spill area overrun %d\n", n)
	}
	for i, v := range spillArea {
		if v == secretValue {
			t.Errorf("secret found in spill slot %d", i)
		}
	}
}
