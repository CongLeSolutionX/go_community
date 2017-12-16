// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"

	"linkname"
)

func main() {
	i, j := linkname.Test()
	if i != 1 {
		log.Fatalf("linknamed function failed to increment linknamed variable (expected 1, got %d)\n", i)
	}
	if j != 43 {
		log.Fatalf("linknamed method failed to return correct result (expected 43, got %d)\n", j)
	}
	k := linkname.TestReflect()
	if k != 43 {
		log.Fatalf("linknamed method failed to return correct result (expected 43, got %d)\n", k)
	}
}
