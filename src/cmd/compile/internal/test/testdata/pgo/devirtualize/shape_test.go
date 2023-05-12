// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WARNING: Please avoid updating this file. If this file needs to be updated,
// then a new shape.pprof file should be generated:
//
//	$ cd $GOROOT/src/cmd/compile/internal/test/testdata/pgo/devirtualize/
//	$ go mod init example.com/pgo/devirtualize
//	$ go test -bench=. -cpuprofile ./shape.pprof

package main

import (
	"testing"
)

func BenchmarkShape(b *testing.B) {
	var iface Shape = Square{10, 10}
	var iface2 Shape = Circle{10, 10}
	iter := 100000
	b.StartTimer()
	_, _ = Slow(iface, iface2, iter)
	b.StopTimer()
}
