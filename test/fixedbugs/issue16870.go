// run

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"reflect"
)

func test(got, want interface{}) {
	if !reflect.DeepEqual(got, want) {
		log.Printf("got %v, want %v", got, want)
	}
}

func main() {
	var i int
	var ok interface{}

	// Channel receives.
	c := make(chan int, 1)

	c <- 42
	i, ok = <-c
	test(i, 42)
	test(ok, true)

	c <- 42
	i, _ = <-c
	test(i, 42)

	c <- 42
	_, ok = <-c
	test(ok, true)

	c <- 42
	_, _ = <-c

	close(c)
	i, ok = <-c
	test(i, 0)
	test(ok, false)

	i, _ = <-c
	test(i, 0)

	_, ok = <-c
	test(ok, false)

	_, _ = <-c

	// Map indexing.
	m := make(map[int]int)

	i, ok = m[0]
	test(i, 0)
	test(ok, false)

	i, _ = m[0]
	test(i, 0)

	_, ok = m[0]
	test(ok, false)

	_, _ = m[0]

	m[0] = 42
	i, ok = m[0]
	test(i, 42)
	test(ok, true)

	i, _ = m[0]
	test(i, 42)

	_, ok = m[0]
	test(ok, true)

	_, _ = m[0]

	// Type assertions.
	var u interface{}

	i, ok = u.(int)
	test(i, 0)
	test(ok, false)

	i, _ = u.(int)
	test(i, 0)

	_, ok = u.(int)
	test(ok, false)

	_, _ = u.(int)

	u = 42
	i, ok = u.(int)
	test(i, 42)
	test(ok, true)

	i, _ = u.(int)
	test(i, 42)

	_, ok = u.(int)
	test(ok, true)

	_, _ = u.(int)
}
