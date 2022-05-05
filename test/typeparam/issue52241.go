// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func DoInOrder[
	Node interface{},
](n *Node, collect func(*Node)) {
}

type ValuesNode[T any] struct {
}

type Collector[T any] struct {
}

func (c *Collector[T]) Collect(tree *ValuesNode[T]) {
}

func TestInOrderIntTree() {
	root := &ValuesNode[int]{}
	collector := Collector[int]{}
	DoInOrder(root, collector.Collect)
}

func main() {
	TestInOrderIntTree()
}
