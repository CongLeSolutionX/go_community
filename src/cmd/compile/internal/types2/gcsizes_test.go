// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2_test

import (
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"testing"
)

type gcSizeTest struct {
	name string
	src  string
}

var gcSizesTests = []gcSizeTest{
	{
		"issue60431",
		`package main

	import "unsafe"

	// The foo struct size is expected to be rounded up to 16 bytes.
	type foo struct {
		a int64
		b bool
	}

	func main() {
		var _ [unsafe.Sizeof(foo{}) - 16]byte
	       println(unsafe.Sizeof(foo{}))
	}`,
	},
	{
		"issue60734",
		`package main

	import (
		"unsafe"
	)

	// The Data struct size is expected to be rounded up to 16 bytes.
	type Data struct {
		Value  uint32   // 4 bytes
		Label  [10]byte // 10 bytes
		Active bool     // 1 byte
		// padded with 1 byte to make it align
	}

	const (
		dataSize = unsafe.Sizeof(Data{})
		dataSizeLiteral = 16
	)

	func main() {
		_ = [16]byte{0, 132, 95, 237, 80, 104, 111, 110, 101, 0, 0, 0, 0, 0, 1, 0}
		_ = [dataSize]byte{0, 132, 95, 237, 80, 104, 111, 110, 101, 0, 0, 0, 0, 0, 1, 0}
	       _ = [dataSizeLiteral]byte{0, 132, 95, 237, 80, 104, 111, 110, 101, 0, 0, 0, 0, 0, 1, 0}
	}`,
	},
}

func TestGCSizes(t *testing.T) {
	for _, tc := range gcSizesTests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			conf := types2.Config{Importer: defaultImporter(), Sizes: types2.SizesFor("gc", "amd64")}
			if _, err := conf.Check("main.go", []*syntax.File{mustParse(tc.src)}, nil); err != nil {
				t.Fatal(err) // type error
			}
		})
	}
}
