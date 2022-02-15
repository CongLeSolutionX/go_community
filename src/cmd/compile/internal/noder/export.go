// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"bytes"
	"fmt"
	"io"

	"cmd/compile/internal/base"
	"cmd/compile/internal/typecheck"
	"cmd/internal/bio"
)

// writeNewExportFunc is a hook that can be added to append extra
// export data after the normal export data section. It allows
// experimenting with new export data format designs without requiring
// immediate support in the go/internal or x/tools importers.
var writeNewExportFunc func(out io.Writer)

func WriteExports(out *bio.Writer) {
	var data bytes.Buffer

	if writeNewExportFunc != nil {
		data.WriteByte('u')
		writeNewExportFunc(&data)
	} else {
		typecheck.WriteExports(&data, true)
	}

	// The linker also looks for the $$ marker - use char after $$ to distinguish format.
	out.WriteString("\n$$B\n") // indicate binary export format
	io.Copy(out, &data)
	out.WriteString("\n$$\n")

	if base.Debug.Export != 0 {
		fmt.Printf("BenchmarkExportSize:%s 1 %d bytes\n", base.Ctxt.Pkgpath, data.Len())
	}
}
