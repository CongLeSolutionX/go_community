// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// int a;
import "C"

func main() {
	i := 5
	C.a + i //ERROR HERE: :12:6:
}
