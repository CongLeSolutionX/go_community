// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.aliastypeparams

package main

import (
	"fmt"

	"issue68526.dir/a"
)

func main() {
	var (
		astr a.A[string]
		aint a.A[int]
	)

	if any(astr) != any(struct{ F string }{}) || any(aint) != any(struct{ F int }{}) {
		panic("zero value of alias and concrete type not identical")
	}

	if any(astr) == any(aint) {
		panic("zero value of struct{ F string } and struct{ F int } are not distinct")
	}

	if got := fmt.Sprintf("%T", astr); got != "struct { F string }" {
		panic(got)
	}
}
