// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/objabi"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	objabi.AddVersionFlag()

	unitchecker.Main(analyzers...)
}
