// run

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

//go:noinline
func mod3(x uint32) uint64 {
	return uint64(x % 3)
}
func main() {
	got := mod3(1<<32 - 1)
	want := uint64((1<<32 - 1) % 3)
	if got != want {
		panic(fmt.Sprintf("got %x want %x", got, want))
	}
}
