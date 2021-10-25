// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package generics

func F[P ~int]() {}

type myInt int

type T1[P myInt] int
type T2[P int|string] int
type T3[P int|~string] int

type I[P I[P]] interface{}

type Normal interface{
	m()
}
