// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lib

type FMap[K comparable, V comparable] map[K]V

//go:noinline
func (m FMap[K, V]) Flip() FMap[V, K] {
	out := make(FMap[V, K])
	return out
}

type MyType uint8

const (
	FIRST MyType = 0
)

var typeStrs = FMap[MyType, string]{
	FIRST: "FIRST",
}

func (self MyType) String() string {
	return typeStrs[self]
}
