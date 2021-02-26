// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

var sink *string

type toobig struct {
	a,b,c string
}

//go:registerparams
//go:noinline
func H(x toobig) string {
	return x.a + " " + x.b + " " + x.c
}

func main() {
	s := H(toobig{"Hello", "there,", "World"})
	gotVsWant(s, "Hello there, World")
	fmt.Println(s)
}

func gotVsWant(got, want string) {
	if got != want {
		fmt.Printf("FAIL, got %s, wanted %s\n", got, want)
	}
}