// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package format_test

import (
	"go/format"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
)

var sink []byte

func BenchmarkFormatUnicodeTables(b *testing.B) {
	tableFile := filepath.Join(runtime.GOROOT(), "src", "vendor", "golang_org", "x", "text", "unicode", "norm", "tables.go")
	data, err := ioutil.ReadFile(tableFile)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink, err = format.Source(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
