// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is tested when running "go test -run Manual"
// without source arguments. Use for one-off debugging.

package p

import "fmt"

type F[A, B any] int

func G[A, B any](F[A, B]) {
}

func _() {
	// TODO(gri) only report one error below (issue #50932)
	var x F /* ERR got 1 arguments but 2 type parameters */ [int]
	G(x /* ERR does not match */)
}

// test case from issue
// (lots of errors but doesn't crash anymore)

type RC[G any, RG any] interface {
	~[]RG
}

type RG[G any] struct{}

type RSC[G any] []*RG[G]

type M[Rc RC[G, RG], G any, RG any] struct {
	Fn func(Rc)
}

type NFn[Rc RC[G, RG], G any, RG any] func(Rc)

func NC[Rc RC[G, RG], G any, RG any](nFn NFn[Rc, G, RG]) {
	var empty Rc
	nFn(empty)
}

func NSG[G any](c RSC[G]) {
	fmt.Println(c)
}

func MMD[Rc RC /* ERR got 1 arguments */ [RG], RG any, G any]() M /* ERR got 2 arguments */ [Rc, RG] {

	var nFn NFn /* ERR got 2 arguments */ [Rc, RG]

	var empty Rc
	switch any(empty).(type) {
	case BC /* ERR undefined: BC */ :

	case RSC[G]:
		nFn = NSG /* ERR cannot use NSG[G] */ [G]
	}

	return M /* ERR got 2 arguments */ [Rc, RG]{
		Fn: func(rc Rc) {
			NC(nFn /* ERR does not match */ )
		},
	}

	return M /* ERR got 2 arguments */ [Rc, RG]{}
}
