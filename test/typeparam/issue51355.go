// run -gcflags=-G=3

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

type Cache[E comparable] struct {
	set   map[E]struct{}
	adder func(...E)
}

func New[E comparable]() *Cache[E] {
	c := &Cache[E]{set: map[E]struct{}{}}

	c.adder = func(elements ...E) {
		for _, value := range elements {
			value := value

			c.set[value] = struct{}{}

			asdf := make(chan struct{})
			go func() {
				fmt.Printf("Value: %v\n", value)

				<-asdf
			}()
		}
	}

	return c
}

func main() {
	c := New[string]()

	c.adder("test")
	time.Sleep(100 * time.Millisecond)
}
