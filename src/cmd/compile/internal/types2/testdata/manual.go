// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is tested when running "go test -run Manual"
// without source arguments. Use for one-off debugging.

package p

type I interface {
	F() error
}

func Foo() error {
	return nil
}

func Do(i I) {
	_ = i.F()
}

func M() {
	imp := I{
		F: Foo,
	}
	Do(imp)
}

func P() {
	imp := I{
		Foo,
	}
	Do(imp)
}

type MI interface {
	F(x int) error
}

func Do2(i MI) {
	_ = i.F(24)
}

func N() {
	imp := MI{
		F: func(x int) error {
			return nil
		},
	}
	Do2(imp)
}

type MII interface {
	F(x int) error
	G() error
}

func Do3(i MII) {
	_ = i.F(24)
}

func O() {
	imp := MII{
		F: func(x int) error {
			return nil
		},
		G: func() error { return nil },
	}
	Do3(imp)
}
