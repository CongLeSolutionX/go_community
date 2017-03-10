// compile

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Compiler rejected initialization of structs to composite literals
// in a non-static setting (e.g. in a function)
// when the struct contained a field named _.

package p

import "fmt"

type T struct {
	_ string
}

var y = T{"stare"}

func main() {
	var x = T{"check"}
	fmt.Printf("%#v\n", x)
}
