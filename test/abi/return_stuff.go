// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

//go:registerparams
//go:noinline
func F(a,b,c *int) int { 
	return *a + *b + *c
}

//go:registerparams
//go:noinline
func H(s, t string) string {
	return s + " " + t 
}

func main() {
	a,b,c := 1,4,16
	x := F(&a, &b, &c)
	fmt.Printf("x = %d\n", x)
	y := H("Hello","World!")
	fmt.Println("len(y) =", len(y))
	fmt.Println("y =", y)
}
