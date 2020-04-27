// compile

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure that literal value can be passed to struct
// blank field of array type, see issue #38690.

package main

func cmpS1(a, b S1) {
	defer func() {
		recover()
	}()
	_ = a == b
}

type S1 struct {
	x int
	_ [2]int
}

type T2 = [1]int

type S2 struct {
	x int
	_ T2
}

type T3 = struct{ y int }

type S3 struct {
	x int
	_ T3
}

func main() {
	cmpS1(S1{1, [2]int{2}}, S1{1, [2]int{3}})
	_ = S2{1, T2{2}}.x == S2{1, T2{3}}.x
	_ = S3{1, T3{3}}.x == S3{1, T3{3}}.x
}
