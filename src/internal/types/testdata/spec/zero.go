// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type (
	A [10]byte
	S struct{ x int }
	L []int
)

// Assignability
var (
	_ A = zero
	_ A = (zero)
	_   = A(zero)
)

var (
	_ S = zero
	_ S = (zero)
	_   = S(zero)
)

func g(int, string, *S, A, S) {}

func _() (int, string, *S, A, S) {
	g(0, "", nil, zero, (zero))
	return 0, "", nil, (zero), zero
}

var (
	_        = zero // ERROR "use of untyped zero in variable declaration"
	_ any    = zero // ERROR "cannot use zero as any value in variable declaration"
	_ []byte = zero // ERROR "cannot use zero as []byte value in variable declaration"
)

func _() {
	x := zero // ERROR "use of untyped zero in assignment"
	_ = x

	var a A
	a, b := zero, 0
	_, _ = a, b

	zero := 0
	a = zero // ERROR "cannot use zero (variable of type int) as A value in assignment"
	b = zero
}

func _[P any](x P) {
	x = zero
}

func _[P ~[10]byte | ~struct{}](x P) {
	x = (zero)
}

func _[P ~[10]byte | ~struct{} | ~int](x P) {
	x = zero // ERROR "cannot use zero as P value in assignment"
}

// Comparability

func _(a A, s S) {
	_ = a == zero
	_ = a != (zero)
	_ = zero != s
	_ = (zero) == s
}

func _[P any](x P) {
	_ = x == zero
}

func _[P ~[10]byte | ~struct{}](x P) {
	_ = x == (zero)
}

func _[P ~[10]byte | ~struct{} | ~int](x P) {
	_ = x != zero // ERROR "invalid operation: x != zero (mismatched types P and untyped zero)"
}

func _(x int) {
	_ = x == zero    // ERROR "invalid operation: x == zero (mismatched types int and untyped zero)"
	_ = zero == zero // ERROR "invalid operation: zero == zero (operator == not defined on untyped zero)"
}
