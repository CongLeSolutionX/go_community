// compile -G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Src[T any] func() Src[T]

func (s *Src[T]) Next() {
	*s = (*s)()
}

func main() {
	var src Src[int]
	src.Next()
}