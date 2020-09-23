// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type T1 struct {
	f2 T2
}

type T2 struct { // ERROR "invalid recursive type: T2" "T2 refers to$" "T1 refers to$" "T2$"
	f1 T1
}

type a b
type b c // ERROR "invalid recursive type: b" "b refers to$" "c refers to$" "b$"
type c b

type d e
type e f
type f f // ERROR "invalid recursive type: f" "f refers to$" "f$"
