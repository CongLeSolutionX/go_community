// compile

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure that literal value can be passed to struct
// blank field of array type, see issue #38690.

package main

func cmpS(a, b S) {
	defer func() {
		recover()
	}()
	_ = a == b
}

type S struct {
	x int
	_ [2]int
}

func main() {
	cmpS(S{1, [2]int{2}}, S{1, [2]int{3}})
}
