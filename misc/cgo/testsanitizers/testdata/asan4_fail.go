// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <stdlib.h>
#include <stdio.h>

int test(int* a) {
	// Heap out of bounds.
	a[3] = 300;          // BOOM
	return a[3];
}*/
import "C"
import "fmt"

func main() {
	var cIntArray [2]C.int
	cIntArray[0] = 100
	cIntArray[1] = 200
	r := C.test(&cIntArray[0]) // cIntArray is moved to heap.
	fmt.Printf("r=%x\n", r)
}
