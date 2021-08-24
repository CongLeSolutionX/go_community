// run -gcflags=-G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure we handle instantiated empty interfaces.

package main

type E[T any] interface {
}

//go:noinline
func f[T any](x E[T]) interface{} {
	return x
}

//go:noinline
func g[T any](x interface{}) E[T] {
	return x
}

func main() {
	f[int](0)
	g[int](1)
}
