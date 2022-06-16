// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type Interface[K any] interface {
	Name() string
}

type interfaceImpl[K any] struct {
	name string
}

func (i interfaceImpl[K]) Name() string {
	return i.name
}

type Type[T any, K any] struct {
	interfaceImpl[K]
}

func main() {
	a := Type[int, int]{
		interfaceImpl[int]{name: "hello"},
	}
	printName(a)
}

func printName[T any](k Type[T, int]) {
	fmt.Println(k.Name())
}
