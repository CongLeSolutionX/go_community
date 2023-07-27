// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test conversion from function to single method interface.

package p

// wrong signature
type I interface{ f(int) int }

var _ I = I(func /* ERRORx `cannot convert .* to type I: func\(\) does not implement I \(missing method f\)` */ () {
})

// multiple method interface
type IM interface {
	f()
	b()
}

var _ IM = IM(func /* ERRORx `cannot convert .* to type IM: IM is not a single method interface` */ () {
})

// embed the wrong interface
type IMM interface{ IM }

var _ IMM = IMM(func /* ERRORx `cannot convert .* to type IMM: IMM is not a single method interface` */ () {
})

// base case
var _ I = I(func(i int) int { return i })

// embed an interface
type IE interface{ I }

var _ IE = IE(func(i int) int { return i })
