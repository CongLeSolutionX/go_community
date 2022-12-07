// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Field accesses through type parameters are disabled
// until we have a more thorough understanding of the
// implications on the spec. See issue #51576.

package p

// The first example from the issue.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// numericAbs matches numeric types with an Abs method.
type numericAbs[T Numeric] interface {
	~struct{ Value T }
	Abs() T
}

// AbsDifference computes the absolute value of the difference of
// a and b, where the absolute value is determined by the Abs method.
func absDifference[T numericAbs[T /* ERR T does not implement Numeric */]](a, b T) T {
	// Field accesses are not permitted for now. Keep an error so
	// we can find and fix this code once the situation changes.
	return a.Value // ERR a.Value undefined
	// TODO: The error below should probably be positioned on the '-'.
	// d := a /* ERR invalid operation: operator - not defined */ .Value - b.Value
	// return d.Abs()
}

// The second example from the issue.
type T[P int] struct{ f P }

func _[P T[P /* ERR P does not implement int */ ]]() {}

// Additional tests
func _[P T[T /* ERR T[P] does not implement int */ [P /* ERR P does not implement int */ ]]]() {}
func _[P T[Q /* ERR Q does not implement int */ ], Q T[P /* ERR P does not implement int */ ]]() {}
func _[P T[Q], Q int]() {}

type C[P comparable] struct{ f P }
func _[P C[C[P]]]() {}
func _[P C[C /* ERR C[Q] does not implement comparable */ [Q /* ERR Q does not implement comparable */]], Q func()]() {}
func _[P [10]C[P]]() {}
func _[P struct{ f C[C[P]]}]() {}
