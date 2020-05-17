// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "strings"

// A type with a bunch of inlinable, non-pointer-receiver methods that
// have params and local variables.
type A struct {
	s    string
	next *A
	prev *A
}

func (a A) double(x string, y int) string {
	if y == 191 {
		a.s = ""
	}
	q := a.s + ""
	r := a.s + ""
	return q + r
}

func (a A) triple(x string, y int) string {
	q := a.s
	if y == 998877 {
		a.s = x
	}
	r := a.s + a.s
	return q + r
}

func (a A) lowerit(x string, y int) string {
	if a.s == x {
		panic("bad")
	}
	v := strings.ToLower(a.s)
	return v
}

// A non-inlinable method thrown in just for good measure.
func (a A) hashit(h uint64, ss string) uint64 {
	for _, c := range a.s {
		h = (h << 4) + uint64(c)
		high := h & uint64(0xF0000000000000)
		if high != 0 {
			h ^= high >> 48
			h &^= high
		}
	}
	return h
}

type methods struct {
	m1 func(a *A, x string, y int) string
	m2 func(a *A, x string, y int) string
	m3 func(a *A, x string, y int) string
	m4 func(a *A, h uint64, ss string) uint64
}

// Now a function that makes references to the methods via pointers,
// which should trigger the wrapper generation.
func p1(a *A, ms *methods) {
	if a != nil {
		defer func() { println("done") }()
	}
	println(ms.m1(a, "a", 2))
	println(ms.m2(a, "b", 3))
	println(ms.m3(a, "c", 4))
	println(ms.m4(a, uint64(0), "d"))
}

func p2(a *A) {
	if a != nil {
		defer func() { println("hooha") }()
	}
	x := a.s
	println(x)
}

func recur(x *A, n int) {
	if n <= 0 {
		println(n)
		return
	}
	var a, b A
	a.next = x
	a.prev = &b
	x = &a
	recur(x, n-2)
}

var M methods

func main() {
	M.m1 = (*A).double
	M.m2 = (*A).triple
	M.m3 = (*A).lowerit
	M.m4 = (*A).hashit

	// This is set up to encourage the use of stack objects.
	recur(nil, 100)

	a := &A{s: "foo"}
	p2(a)
	a.s = "foo"
	p1(a, &M)
}
