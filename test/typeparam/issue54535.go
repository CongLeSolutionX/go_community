// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type node[T any] struct {
	items    items[T]
	children items[*node[T]]
}

func (n *node[T]) f(i int, j int) bool {
	if len(n.children[i].items) < j {
		return false
	}
	return true
}

type items[T any] []T

func main() {
	_ = node[int]{}
}
