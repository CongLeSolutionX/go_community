// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that we handle large constants correctly.
// On amd64, most instructions only accept a signed
// 32-bit constant or offset.

package main

const c = 0x1122334455

// Make sure big pointer arithmetic works.
// We don't actually call these f's, but make sure
// they compile.
func f1(s []byte) byte {
	return s[c]
}

func f2(s []byte) *byte {
	return &s[c]
}

type T struct {
	pad [c]byte
	x   int
}

func f3(t *T) int {
	return t.x
}
func f4(t *T) *int {
	return &t.x
}

type U struct {
	pad [c]byte
	x   [8]byte
}

func f5(u *U, i int) byte {
	return u.x[i]
}
func f6(u *U, i int) *byte {
	return &u.x[i]
}

type W struct {
	pad [c]byte
	x   [8]int16
}

func f7(w *W, i int) int16 {
	return w.x[i]
}
func f8(w *W, i int) *int16 {
	return &w.x[i]
}

type X struct {
	pad [c]byte
	x   [8]int32
}

func f9(x *X, i int) int32 {
	return x.x[i]
}
func f10(x *X, i int) *int32 {
	return &x.x[i]
}

type Y struct {
	pad [c]byte
	x   [8]int64
}

func f11(y *Y, i int) int64 {
	return y.x[i]
}
func f12(y *Y, i int) *int64 {
	return &y.x[i]
}

// Make sure forming *const opcodes doesn't lose
// the high-order bits of constants.

//go:noinline
func add(x int64) int64 {
	return x + c
}

//go:noinline
func sub(x int64) int64 {
	return x - c
}

//go:noinline
func mul(x int64) int64 {
	return x * c
}

//go:noinline
func and(x int64) int64 {
	return x & c
}

//go:noinline
func or(x int64) int64 {
	return x | c
}

//go:noinline
func xor(x int64) int64 {
	return x ^ c
}

//go:noinline
func cmp(x int64) bool {
	return x < c
}

//go:noinline
func lea1(x, y int64) int64 {
	return x + y + c
}

//go:noinline
func lea2(x, y int64) int64 {
	return x + 2*y + c
}

//go:noinline
func lea4(x, y int64) int64 {
	return x + 4*y + c
}

//go:noinline
func lea8(x, y int64) int64 {
	return x + 8*y + c
}

func main() {
	const d = 77
	failed := false
	if want, got := int64(d+c), add(d); got != want {
		println("add wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(d-c), sub(d); got != want {
		println("sub wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(d*c), mul(d); got != want {
		println("mul wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(d&c), and(d); got != want {
		println("and wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(d|c), or(d); got != want {
		println("or wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(d^c), xor(d); got != want {
		println("xor wanted", want, "got", got)
		failed = true
	}
	if want, got := true, cmp(c-1); got != want {
		println("cmp wanted", want, "got", got)
		failed = true
	}
	if want, got := false, cmp(c); got != want {
		println("cmp wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(c+2*d), lea1(d, d); got != want {
		println("lea1 wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(c+3*d), lea2(d, d); got != want {
		println("lea2 wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(c+5*d), lea4(d, d); got != want {
		println("lea4 wanted", want, "got", got)
		failed = true
	}
	if want, got := int64(c+9*d), lea8(d, d); got != want {
		println("lea8 wanted", want, "got", got)
		failed = true
	}
	if failed {
		panic("failed")
	}
}
