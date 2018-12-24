// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements a simple printer performance benchmark:
// go test -bench=BenchmarkPrint

package printer

import (
	"bytes"
	"go/ast"
	"go/parser"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func testprint(out io.Writer, file *ast.File) {
	if err := (&Config{TabIndent | UseSpaces, 8, 0}).Fprint(out, fset, file); err != nil {
		log.Fatalf("print error: %s", err)
	}
}

// cannot initialize in init because (printer) Fprint launches goroutines.
func initialize(filename string) *ast.File {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("%s", err)
	}

	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		log.Fatalf("%s", err)
	}

	var buf bytes.Buffer
	testprint(&buf, file)
	if !bytes.Equal(buf.Bytes(), src) {
		log.Fatalf("print error: %s not idempotent", filename)
	}

	return file
}

func BenchmarkPrint(b *testing.B) {
	f := initialize("testdata/parser.go")
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testprint(ioutil.Discard, f)
	}
}

func makeUnbalanced() *os.File {
	f, err := ioutil.TempFile("testdata", "*.go")
	if err != nil {
		log.Fatalf("failed to create temp benchmark file")
	}

	f.WriteString("package p\n\nvar n = 1" + strings.Repeat(" + 1", 16000) + "\n")
	return f
}

func BenchmarkPrintUnbalanced(b *testing.B) {
	f := makeUnbalanced()
	file := initialize(f.Name())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		testprint(ioutil.Discard, file)
	}

	f.Close()
	err := os.Remove(f.Name())
	if err != nil {
		b.Fatalf("failed to removed temp file %v", f.Name())
	}
}
