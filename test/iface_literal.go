// run -ldflags=-prunedeadmeth=0

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type S struct {
	_do func()
}

func (s *S) Do() { s._do() }

type I interface {
	Do()
}

func Doit(i I) {
	i.Do()
}

func main() {
	si := &S{}
	Doit(si)

	_ = I{
		Do: func() {},
	}
}
