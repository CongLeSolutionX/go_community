// errorcheckwithauto -0 -d=ssa/expand_calls/debug=-2

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// Check that the method wrapper with pointer receivers (in both the wrapper and the method) uses tail call.

// ERRORAUTO "rewrite TailCall.*<\[2\]int,mem>.*\(\*Foo\).Get2Vals"

//go:noinline
func (f *Foo) Get2Vals() [2]int { return [2]int{f.Val, f.Val + 1} }
func (f *Foo) Get3Vals() [3]int { return [3]int{f.Val, f.Val + 1, f.Val + 2} }

type Foo struct{ Val int }

type Bar struct {
	int64
	*Foo // needs a method wrapper
	string
}

var i any

func init() {
	i = Bar{1, nil, "first"}
}
