// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type T1 struct {
	f2 T2
}

type T2 struct { // ERROR "(?s)invalid recursive type: T2.\tT2 refers to.\tT1 refers to.\tT2$"
	f1 T1
}

type a b
type b c // ERROR "(?s)invalid recursive type: b.\tb refers to.\tc refers to.\tb$"
type c b

type d e
type e f
type f f // ERROR "(?s)invalid recursive type: f.\tf refers to.\tf$"

type g struct {
	h struct { // ERROR "(?s)invalid recursive type: struct { g }.\tg refers to.\tstruct { g }$"
		g
	}
}
