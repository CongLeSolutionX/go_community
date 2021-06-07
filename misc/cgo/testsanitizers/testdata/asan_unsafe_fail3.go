// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"unsafe"
)

func main() {
	a := 1
	b := 2
	// The local variables.
	// Set -d=checkptr to 2 when ASan is enabled, unsafe.Pointer(&a) conversion is escaping.
	var p *int = (*int)(unsafe.Pointer(uintptr(unsafe.Pointer(&a)) + 1*unsafe.Sizeof(int(1)))) // BOOM
	*p = 20
	d := a + b
	fmt.Println(d)
}
