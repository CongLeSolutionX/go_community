// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure forming *const opcodes doesn't lose
// the high-order bits of constants.

package main

const c = 0x1122334455

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
