// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

type X int

func (x X) M() X { return x }

func F[T interface{ M() U }, U interface{ M() T }]() {
	type Y[V interface{ M() W }, W interface{ M() V }] int

	var _ Y[X, X]
	var _ Y[T, U]
}

func G() { F[X, X]() }
