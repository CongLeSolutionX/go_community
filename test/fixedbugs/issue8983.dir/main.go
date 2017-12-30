// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"

	"./a1"
	"./a2"
)

type Closer int

func newOne1() Closer {
	return a1.NewCloser()
}

func newOne2() Closer {
	return a2.NilCloser()
}

type error int

func F() error {
	return errors.New("x")
}
