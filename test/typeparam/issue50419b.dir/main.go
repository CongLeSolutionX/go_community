// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that type substitution works correctly even for a method of a generic type
// that has multiple blank type params.

package main

import (
	"a"
	"fmt"
)

func main() {
	foo := &a.Foo[string, int]{
		valueA: "i am a string",
		valueB: 123,
	}
	if got, want := fmt.Sprintln(foo), "i am a string 123\n"; got != want {
		panic(fmt.Sprintf("got %s, want %s", got, want))
	}
}
