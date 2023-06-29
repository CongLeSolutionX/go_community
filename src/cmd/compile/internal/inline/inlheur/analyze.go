// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const debugTrace = 0

// computeFuncProps analyzes the specified function 'fn' and computes
// for it a function "properties" object, to be used to drive inlining
// heuristics. See comments on the FuncProps type for more info.
func computeFuncProps(fn *ir.Func) *FuncProps {
	if debugTrace != 0 {
		fmt.Fprintf(os.Stderr, "=-= starting analysis of func %v:\n%+v\n",
			fn.Sym().Name, fn)
	}
	// implementation stubbed out for now
	return &FuncProps{}
}

// DumpFuncProps computes and caches function properties for the func
// 'fn', or if fn is nil, writes out the cached set of properties to
// the file given in 'dumpfile'. Used only for unit testing.
func DumpFuncProps(fn *ir.Func, dumpfile string) {
	if fn != nil {
		fp := computeFuncProps(fn)
		entry := fnWithProps{
			fname: fn.Sym().Name,
			props: fp,
		}
		dumpBuffer = append(dumpBuffer, entry)
		return
	}
	outf, err := os.OpenFile(dumpfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		base.Fatalf("opening function props dump file %q: %v\n", dumpfile, err)
	}
	defer outf.Close()
	fmt.Fprintf(outf, "// DO NOT EDIT (use 'go test -v -update-expected' instead)\n")
	for _, entry := range dumpBuffer {
		if !strings.HasPrefix(entry.fname, "T_") {
			continue
		}
		if err := dumpFnPreamble(outf, entry.fname, entry.props); err != nil {
			base.Fatalf("function props dump: %v\n", err)
		}
	}
	dumpBuffer = nil
}

func dumpFnPreamble(w io.Writer, fname string, props *FuncProps) error {
	fmt.Fprintf(w, "// %s\n", fname)
	// emit props as comments, followed by delimiter
	fmt.Fprintf(w, "%s// %s\n", props.ToString("// "), comDelimiter)
	data, err := json.Marshal(props)
	if err != nil {
		return fmt.Errorf("marshall error %v\n", err)
	}
	fmt.Fprintf(w, "// %s\n// %s\n", string(data), fnDelimiter)
	return nil
}

const fnDelimiter = "=-=-="
const comDelimiter = "====="

type fnWithProps struct {
	fname string
	props *FuncProps
}

var dumpBuffer []fnWithProps
