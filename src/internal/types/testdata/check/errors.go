// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

// Testing precise operand formatting in error messages
// (matching messages are regular expressions, hence the \'s).
func f(x int, m map[string]int) {
	// no values
	_ = f /* ERR f(0, m) (no value) used as value */ (0, m)

	// built-ins
	_ = println // ERR println (built-in) must be called

	// types
	_ = complex128 // ERR complex128 (type) is not an expression

	// constants
	const c1 = 991
	const c2 float32 = 0.5
	const c3 = "foo"
	0 // ERR 0 (untyped int constant) is not used
	0.5 // ERR 0.5 (untyped float constant) is not used
	"foo" // ERR "foo" (untyped string constant) is not used
	c1 // ERR c1 (untyped int constant 991) is not used
	c2 // ERR c2 (constant 0.5 of type float32) is not used
	c1 /* ERR c1 + c2 (constant 991.5 of type float32) is not used */ + c2
	c3 // ERR c3 (untyped string constant "foo") is not used

	// variables
	x // ERR x (variable of type int) is not used

	// values
	nil // ERR nil is not used
	( /* ERR (*int)(nil) (value of type *int) is not used */ *int)(nil)
	x /* ERR x != x (untyped bool value) is not used */ != x
	x /* ERR x + x (value of type int) is not used */ + x

	// value, ok's
	const s = "foo"
	m /* ERR m[s] (map index expression of type int) is not used */ [s]
}

// Valid ERROR comments can have a variety of forms.
func _() {
	0 /* ERROR 0 .* is not used */
	0 /* ERROR 0 .* is not used */
	0 // ERROR 0 .* is not used
	0 // ERROR 0 .* is not used
}

// Don't report spurious errors as a consequence of earlier errors.
// Add more tests as needed.
func _() {
	if err := foo /* ERR undefined */ (); err != nil /* no error here */ {}
}

// Use unqualified names for package-local objects.
type T struct{}
var _ int = T /* ERR value of type T */ {} // use T in error message rather then errors.T

// Don't report errors containing "invalid type" (issue #24182).
func _(x *missing /* ERR undefined: missing */ ) {
	x.m() // there shouldn't be an error here referring to *invalid type
}
