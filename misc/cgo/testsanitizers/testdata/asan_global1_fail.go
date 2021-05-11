// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

/*
#include <stdlib.h>
#include <stdio.h>

int test(int *a) {
	a[3] = 300;  // BOOM
	return a[3];
}
*/
import "C"

import "fmt"

var cIntArray [2]C.int

func main() {
	r := C.test(&cIntArray[0])
	fmt.Println("r value = ", r)
}
