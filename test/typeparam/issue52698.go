// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/* if this type decl is moved to the position after decl of T15, it will compile successfully */
type T23 interface {
	~struct {
		Field0 T13[T15]
	}
}

type T1[P1 interface {
}] struct {
	Field2 P1
}

type T13[P2 interface {
}] struct {
	Field2 T1[P2]
}

type T15 struct {
	Field0 T13[string]
}

func main() {
}
