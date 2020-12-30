// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import "./b"

type T int

func (T) Mv()  {}
func (*T) Mp() {}

var b1 int = (b.B).Mv  // ERROR "cannot use b\.B\.Mv"
var b2 int = (*b.B).Mv // ERROR "cannot use \(\*b\.B\)\.Mv"
var b3 int = (*b.B).Mp // ERROR "cannot use \(\*b\.B\)\.Mp"

var x3 int = (T).Mv  // ERROR "cannot use T\.Mv"
var x4 int = (*T).Mv // ERROR "cannot use \(\*T\)\.Mv"
var x5 int = (*T).Mp // ERROR "cannot use \(\*T\)\.Mp"
