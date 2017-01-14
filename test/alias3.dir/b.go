// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import (
	"./a"
	. "go/build"
)

func F(x float64) a.Float64 {
	return x
}

type MyContext = Context // = build.Context

var C a.Context = Default

type S struct{}

func (S) M1(a.IntAlias) float64
func (S) M2() Context

var _ a.I1 = S{}
var _ a.I2 = S{}
