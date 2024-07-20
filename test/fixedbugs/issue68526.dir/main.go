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
		gstr a.G[string]

		aint a.A[int]
		gint a.G[int]
	)

	if any(astr) != any(gstr) || any(aint) != any(gint) {
		panic("zero value of alias and concrete not identical")
	}

	if any(astr) == any(aint) {
		panic("zero value of a.G[string] and a.G[int] are distinct types")
	}

	if got, want := fmt.Sprintf("%T", astr), "a.G[string]"; got != want {
		panic(got)
	}
}
