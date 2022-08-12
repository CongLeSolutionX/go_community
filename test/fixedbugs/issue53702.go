// compile
//go:build goexperiment.unified

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Elem struct{}

func (*Elem) Wait(callback func()) {}

type Base struct {
	elem [8]*Elem
}

var g_val = 1

func (s *Base) Do() *int {
	resp := &g_val
	for *resp != 0 {
		s.elem[0].Wait(func() { *resp = 0 })
	}
	return resp
}

type Sub struct {
	*Base
}

func main() {
	a := Sub{}
	a.Do()
}
