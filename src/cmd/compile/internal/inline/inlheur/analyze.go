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
	"path/filepath"
	"strings"
)

const debugTrace = 0
const (
	debugTraceFuncFlags = 1 << iota
)

// propAnalyzer interface is used for defining one or more
// analyzer helper objects, each tasked with computing some
// specific subset of the properties we're interested in.

type propAnalyzer interface {
	nodeVisit(n ir.Node, aux any)
}

type fnInlHeur struct {
	fname string
	file  string
	line  uint
	props *FuncProps
}

// for it a function "properties" object, to be used to drive inlining
// heuristics. See comments on the FuncProps type for more info.
func computeFuncProps(fn *ir.Func) *FuncProps {
	if debugTrace != 0 {
		fmt.Fprintf(os.Stderr, "=-= starting analysis of func %v:\n%+v\n",
			fn.Sym().Name, fn)
	}
	ffa := makeFuncFlagsAnalyzer(fn)
	analyzers := []propAnalyzer{ffa}
	runAnalyzersOnFunction(fn, analyzers)
	return &FuncProps{
		Flags: ffa.results(),
	}
}

func runAnalyzersOnFunction(fn *ir.Func, analyzers []propAnalyzer) {
	var doNode func(ir.Node) bool
	doNode = func(n ir.Node) bool {
		ir.DoChildren(n, doNode)
		for _, a := range analyzers {
			a.nodeVisit(n, nil)
		}
		return false
	}
	doNode(fn)
}

func interestingToCompare(fname string) bool {
	if strings.HasPrefix(fname, "init.") {
		return true
	}
	if strings.HasPrefix(fname, "T_") {
		return true
	}
	f := strings.Split(fname, ".")
	if len(f) == 2 && strings.HasPrefix(f[1], "T_") {
		return true
	}
	return false
}

func fnFileLine(fn *ir.Func) (string, uint) {
	p := base.Ctxt.InnermostPos(fn.Pos())
	return filepath.Base(p.Filename()), p.Line()
}

// DumpFuncProps computes and caches function properties for the func
// 'fn', or if fn is nil, writes out the cached set of properties to
// the file given in 'dumpfile'. Used only for unit testing.
func DumpFuncProps(fn *ir.Func, dumpfile string) {
	if fn != nil {
		fp := computeFuncProps(fn)
		file, line := fnFileLine(fn)
		entry := fnInlHeur{
			fname: fn.Sym().Name,
			file:  file,
			line:  line,
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
	dumpFilePreamble(outf)
	for _, entry := range dumpBuffer {
		if err := dumpFnPreamble(outf, &entry); err != nil {
			base.Fatalf("function props dump: %v\n", err)
		}
	}
	dumpBuffer = nil
}

func dumpFilePreamble(w io.Writer) {
	fmt.Fprintf(w, "// DO NOT EDIT (use 'go test -v -update-expected' instead.)\n")
	fmt.Fprintf(w, "// See cmd/compile/internal/inline/inlheur/testdata/props/README.txt\n")
	fmt.Fprintf(w, "// for more information on the format of this file.\n")
	fmt.Fprintf(w, "// %s\n", preambleDelimiter)
}

func dumpFnPreamble(w io.Writer, fih *fnInlHeur) error {
	fmt.Fprintf(w, "// %s %s %d\n", fih.file, fih.fname, fih.line)
	// emit props as comments, followed by delimiter
	fmt.Fprintf(w, "%s// %s\n", fih.props.ToString("// "), comDelimiter)
	data, err := json.Marshal(fih.props)
	if err != nil {
		return fmt.Errorf("marshall error %v\n", err)
	}
	fmt.Fprintf(w, "// %s\n// %s\n", string(data), fnDelimiter)
	return nil
}

const preambleDelimiter = "=^=^="
const fnDelimiter = "=-=-="
const comDelimiter = "====="

var dumpBuffer []fnInlHeur
