// errorcheck -0 -m

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test escape analysis with label.

package p

func g() bool

//go:noescape
func h(*int)

func f1() {
	var p *int
	if g() {
	x:
		goto x
	} else {
		p = new(int) // ERROR "f1 new\(int\) does not escape$"
	}
	h(p)
}
