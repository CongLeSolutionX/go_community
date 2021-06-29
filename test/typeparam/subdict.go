// run -gcflags=-G=3

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test situations where functions/methods are not
// immediately called and we need to capture the dictionary
// required for later invocation.

package main

import (
	"fmt"
)

type value[T comparable] struct {
	val T
}

func (v *value[T]) test(def T) bool {
	return (v.val == def)
}

func (v *value[T]) get(def T) T {
	var c value[int]
	if c.test(32) {
		return def
	} else if v.test(def) {
		return def
	} else {
		return v.val
	}
}


func main() {
	var s value[string]
	if got, want := s.get("ab"), ""; got != want {
		panic(fmt.Sprintf("get() == %d, want %d", got, want))
	}
}
