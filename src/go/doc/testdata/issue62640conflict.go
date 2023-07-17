// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue62640conflict

type A struct{
	// B should be hidden within S because it conflicts with u.S.
	B int
}

type c struct{
	// B should be hidden within S because it conflicts with e.S.
	B

	// D should be exposed by S.
	D
}

type B struct{
	// E should be unambiguous within s even though u.B has a conflict.
	E int
}

type D struct {
	// F should be visible within s because it is embedded in u,
	// even though u itself is unexported.
	F int
}

// S has visible selectors A, D, and F,
// which should hide deep methods in Outer.
type S struct {
	A
	c
}

type Shallow struct {
	Deep
}

type Deep struct {
	Deeper
}

type Deeper struct {
	Deepest
}

// Deepest declares methods that may or may not be visible in Outer.
type Deepest struct{}

func (Deepest) A() {}
func (Deepest) B() {}
func (Deepest) c() {}
func (Deepest) D() {}
// E should be masked by S.c.B.E.
func (Deepest) E() {}
// F should be masked by S.c.D.F.
func (Deepest) F() {}

type Outer struct {
	S
	Shallow
}
