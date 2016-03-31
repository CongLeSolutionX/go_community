// +build amd64
// run

// Copyright 2016 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	defer func() {
		if e := recover(); e == nil {
			panic("expected divide by zero panic")
		}
	}()
	v := 1
	_ = 5 / (1 - v)
}
