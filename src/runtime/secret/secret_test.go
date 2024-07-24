// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret_test

import (
	"runtime"
	"runtime/secret"
	"testing"
	"unsafe"
)

const secretValue = 0x53c237

type secretType int

// Test that when we allocate inside secret.Do, the resulting
// allocations are zeroed by the garbage collector when they
// are freed.
// See runtime/mheap.go:freeSpecial.
func TestHeap(t *testing.T) {
	var u1, u2 uintptr

	secret.Do(func() {
		t2 := makeT2()

		u2 = uintptr(unsafe.Pointer(t2))
		u1 = uintptr(unsafe.Pointer(t2.t1))
	})

	runtime.GC()

	// Resurrect objects and check them.
	t1 := (*T1)(unsafe.Pointer(u1))
	t2 := (*T2)(unsafe.Pointer(u2))
	if t1.a == secretValue || t1.b == secretValue {
		t.Errorf("t1 not cleared!")
	}
	if t2.a == secretValue || t2.b == secretValue {
		t.Errorf("t2 not cleared!")
	}
}

type T1 struct {
	a, b secretType
}
type T2 struct {
	t1   *T1
	a, b secretType
}

// go:noinline
func makeT1() *T1 {
	// Note: noinline forces heap allocation
	return &T1{a: secretValue, b: secretValue}
}

//go:noinline
func makeT2() *T2 {
	// Note: noinline forces heap allocation
	return &T2{t1: makeT1(), a: secretValue, b: secretValue}
}

// Test that when we return from secret.Do, we zero the stack used
// by the argument to secret.Do.
// See runtime/secret.go:secret_dec.
func TestStack(t *testing.T) {
	var u uintptr
	secret.Do(func() {
		var s [100]T1
		for i := range s {
			s[i].a = secretValue
			s[i].b = secretValue
		}
		use(&s)
		u = uintptr(unsafe.Pointer(&s[0]))
	})

	// Resurrect array and check it.
	p := (*[100]T1)(unsafe.Pointer(u))
	for i := range p {
		if p[i].a == secretValue || p[i].b == secretValue {
			t.Errorf("p[%d] not cleared", i)
			break
		}
	}
}

//go:noinline
func use(t *[100]T1) {
	// Note: noinline prevents dead variable elimination.
}

// Test that when we copy a stack, we zero the old one.
// See runtime/stack.go:copystack.
func TestStackCopy(t *testing.T) {
	var u uintptr
	secret.Do(func() {
		var t T1
		t.a = secretValue
		t.b = secretValue
		u = uintptr(unsafe.Pointer(&t))
		growStack()
	})

	// Resurrect object and check it.
	p := (*T1)(unsafe.Pointer(u))
	if p.a == secretValue || p.b == secretValue {
		t.Errorf("p not cleared")
	}

}

func growStack() {
	growStack1(100000)
}
func growStack1(n int) {
	if n == 0 {
		return
	}
	growStack1(n - 1)
}
