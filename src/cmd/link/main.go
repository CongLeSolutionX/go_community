// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/obj"
	"fmt"
	"os"
)

var archMain = map[string]func(){}

func main() {
	if mainFn, ok := archMain[obj.GOARCH]; ok {
		mainFn()
	} else {
		fmt.Fprintf(os.Stderr, "link: unknown architecture %q\n", obj.GOARCH)
		os.Exit(2)
	}
}
