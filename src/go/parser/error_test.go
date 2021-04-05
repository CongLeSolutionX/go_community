// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements a parser test harness. The files in the testdata
// directory are parsed and the errors reported are compared against the
// error messages expected in the test files. The test files must end in
// .src rather than .go so that they are not disturbed by gofmt runs.
//
// Expected errors are indicated in the test files by putting a comment
// of the form /* ERROR "rx" */ immediately following an offending token.
// The harness will verify that an error matching the regular expression
// rx is reported at that source position.
//
// For instance, the following test file indicates that a "not declared"
// error should be reported for the undeclared variable x:
//
//	package p
//	func f() {
//		_ = x /* ERROR "not declared" */ + 1
//	}

package parser

import (
	"bytes"
	"flag"
	"go/ast"
	"go/internal/typeparams"
	"go/printer"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const testdata = "testdata"

var updateGolden = flag.Bool("update_golden", false, "whether to update .golden files for recovery tests")

// getFile assumes that each filename occurs at most once
func getFile(fset *token.FileSet, filename string) (file *token.File) {
	fset.Iterate(func(f *token.File) bool {
		if f.Name() == filename {
			if file != nil {
				panic(filename + " used multiple times")
			}
			file = f
		}
		return true
	})
	return file
}

// ERROR comments must be of the form /* ERROR "rx" */ and rx is
// a regular expression that matches the expected error message.
// The special form /* ERROR HERE "rx" */ must be used for error
// messages that appear immediately after a token, rather than at
// a token's position.
//
var errRx = regexp.MustCompile(`^/\* *ERROR *(HERE)? *"([^"]*)" *\*/$`)

// expectedErrors collects the regular expressions of ERROR comments found
// in files and returns them as a map of error positions to error messages.
//
func expectedErrors(fdesc *token.File, src []byte) map[token.Pos]string {
	errors := make(map[token.Pos]string)

	var s scanner.Scanner
	// file was parsed already - do not add it again to the file
	// set otherwise the position information returned here will
	// not match the position information collected by the parser
	s.Init(fdesc, src, nil, scanner.ScanComments)
	var prev token.Pos // position of last non-comment, non-semicolon token
	var here token.Pos // position immediately after the token at position prev

	for {
		pos, tok, lit := s.Scan()
		switch tok {
		case token.EOF:
			return errors
		case token.COMMENT:
			s := errRx.FindStringSubmatch(lit)
			if len(s) == 3 {
				pos := prev
				if s[1] == "HERE" {
					pos = here
				}
				errors[pos] = string(s[2])
			}
		case token.SEMICOLON:
			// don't use the position of auto-inserted (invisible) semicolons
			if lit != ";" {
				break
			}
			fallthrough
		default:
			prev = pos
			var l int // token length
			if tok.IsLiteral() {
				l = len(lit)
			} else {
				l = len(tok.String())
			}
			here = prev + token.Pos(l)
		}
	}
}

// compareErrors compares the map of expected error messages with the list
// of found errors and reports discrepancies.
//
func compareErrors(t *testing.T, fdesc *token.File, expected map[token.Pos]string, found scanner.ErrorList) {
	t.Helper()
	for _, error := range found {
		// error.Pos is a token.Position, but we want
		// a token.Pos so we can do a map lookup
		pos := fdesc.Pos(error.Pos.Offset)
		if msg, found := expected[pos]; found {
			// we expect a message at pos; check if it matches
			rx, err := regexp.Compile(msg)
			if err != nil {
				t.Errorf("%s: %v", error.Pos, err)
				continue
			}
			if match := rx.MatchString(error.Msg); !match {
				t.Errorf("%s: %q does not match %q", error.Pos, error.Msg, msg)
				continue
			}
			// we have a match - eliminate this error
			delete(expected, pos)
		} else {
			// To keep in mind when analyzing failed test output:
			// If the same error position occurs multiple times in errors,
			// this message will be triggered (because the first error at
			// the position removes this position from the expected errors).
			t.Errorf("%s: unexpected error: %s", error.Pos, error.Msg)
		}
	}

	// there should be no expected errors left
	if len(expected) > 0 {
		t.Errorf("%d errors not reported:", len(expected))
		for pos, msg := range expected {
			t.Errorf("%s: %s\n", fdesc.Position(pos), msg)
		}
	}
}

func checkErrors(t *testing.T, filename string, input interface{}, mode Mode, expectErrors bool) {
	t.Helper()
	src, err := readSource(filename, input)
	if err != nil {
		t.Error(err)
		return
	}

	fset := token.NewFileSet()
	_, err = ParseFile(fset, filename, src, mode)
	compareWithExpected(t, err, getFile(fset, filename), src, expectErrors)
}

func compareWithExpected(t *testing.T, err error, fdesc *token.File, src []byte, expectErrors bool) {
	t.Helper()
	found, ok := err.(scanner.ErrorList)
	if err != nil && !ok {
		t.Error(err)
		return
	}
	found.RemoveMultiples()

	expected := map[token.Pos]string{}
	if expectErrors {
		// we are expecting the following errors
		// (collect these after parsing a file so that it is found in the file set)
		expected = expectedErrors(fdesc, src)
	}

	// verify errors returned by the parser
	compareErrors(t, fdesc, expected, found)
}

func TestErrors(t *testing.T) {
	list, err := os.ReadDir(testdata)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range list {
		name := d.Name()
		if !d.IsDir() && !strings.HasPrefix(name, ".") && (strings.HasSuffix(name, ".src") || strings.HasSuffix(name, ".go2")) {
			mode := DeclarationErrors | AllErrors
			if strings.HasSuffix(name, ".go2") {
				if !typeparams.Enabled {
					continue
				}
			} else {
				mode |= typeparams.DisallowParsing
			}
			checkErrors(t, filepath.Join(testdata, name), nil, mode, true)
		}
	}
}

func TestRecovery(t *testing.T) {
	dir := filepath.Join("testdata", "recovery")
	fis, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range fis {
		if !strings.HasSuffix(fi.Name(), ".input") {
			continue
		}
		name := strings.TrimSuffix(fi.Name(), ".input")
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join(dir, fi.Name())
			src, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			fset := token.NewFileSet()
			f, err := ParseFile(fset, filename, src, AllErrors)
			compareWithExpected(t, err, getFile(fset, filename), src, true)

			got := formattedFile(fset, f)
			goldenFile := filepath.Join(dir, name+".golden")
			if *updateGolden {
				os.WriteFile(goldenFile, got, 0644)
			} else {
				golden, err := os.ReadFile(goldenFile)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, golden) {
					t.Errorf("parsed file does not match golden:\nparsed:\n%s\ngolden:\n%s", string(got), string(golden))
				}
			}
		})
	}
}

func formattedFile(fset *token.FileSet, f *ast.File) []byte {
	var buf bytes.Buffer
	config := printer.Config{Mode: printer.RawFormat}
	config.Fprint(&buf, fset, f)
	return buf.Bytes()
}
