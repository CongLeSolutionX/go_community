// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/importer"
)

func main() {
	imp := importer.For("source", nil)
	pkg, err := imp.Import("example.com/testdata/testpkg")
	if err != nil {
		panic(err)
	}
	if got, want := pkg.Path(), "example.com/testdata/testpkg"; got != want {
		panic(fmt.Errorf("got %q; want %q", got, want))
	}
}
