// errorcheck -0 -m
//go:build goexperiment.unified

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package foo

func f5() *byte { // ERROR "can inline f5"
	type T struct {
		x [1]byte
	}
	t := new(T) // ERROR "new.T. escapes to heap"
	return &t.x[0]
}

func f6() *byte { // ERROR "can inline f6"
	type T struct {
		x struct {
			y byte
		}
	}
	t := new(T) // ERROR "new.T. escapes to heap"
	return &t.x.y
}
