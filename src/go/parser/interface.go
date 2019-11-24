// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the exported entry points for invoking the parser.

package parser

import (
	"bytes"
	"errors"
	"go/ast"
	"go/scanner"
	"go/token"
	"internal/syntax"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// A Mode value is a set of flags (or 0).
// They control the amount of source code parsed and other optional
// parser functionality.
//
type Mode uint

const (
	PackageClauseOnly Mode             = 1 << iota // stop parsing after package clause
	ImportsOnly                                    // stop parsing after import declarations
	ParseComments                                  // parse comments and add them to AST
	Trace                                          // print a trace of parsed productions
	DeclarationErrors                              // report declaration errors
	SpuriousErrors                                 // same as AllErrors, for backward-compatibility
	AllErrors         = SpuriousErrors             // report all errors (not just the first 10 on different lines)
)

// ParseFile parses the source code of a single Go source file and returns
// the corresponding ast.File node. The source code may be provided via
// the filename of the source file, or via the src parameter.
//
// If src != nil, ParseFile parses the source from src and the filename is
// only used when recording position information. The type of the argument
// for the src parameter must be string, []byte, or io.Reader.
// If src == nil, ParseFile parses the file specified by filename.
//
// The mode parameter controls the amount of source text parsed and other
// optional parser functionality. Position information is recorded in the
// file set fset, which must not be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by source position.
//
func ParseFile(fset *token.FileSet, filename string, src interface{}, mode Mode) (f *ast.File, err error) {
	if fset == nil {
		panic("parser.ParseFile: no token.FileSet provided (fset == nil)")
	}
	var bs []byte
	switch x := src.(type) {
	case []byte:
		bs = x
	case string:
		bs = []byte(x)
	case io.Reader:
		var err error
		if bs, err = ioutil.ReadAll(x); err != nil {
			return nil, err
		}
	case nil:
		var err error
		if bs, err = ioutil.ReadFile(filename); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid source")
	}
	var errlist scanner.ErrorList
	errh := func(err error) {
		switch x := err.(type) {
		case syntax.Error:
			pos := token.Position{
				Line:   int(x.Pos.Line()),
				Column: int(x.Pos.Col()),
			}
			errlist.Add(pos, x.Msg)
		default:
			errlist.Add(token.Position{}, x.Error())
		}
	}
	r := bytes.NewReader(bs)
	sf, err := syntax.Parse(syntax.NewFileBase(filename), r, errh, nil, 0)
	if sf == nil {
		return nil, errlist
	}
	conv := converter{mode: mode, fset: fset, errh: errh}

	base := fset.Base()
	conv.tokfile = fset.AddFile(filename, base, len(bs))

	lines := make([]int, 1, sf.Lines)
	for i := 0; i < len(bs); i++ {
		c := bs[i]
		if c == '\n' {
			lines = append(lines, i)
		}
	}
	if !conv.tokfile.SetLines(lines) {
		//panic("could not set lines")
	}
	conv.lines = lines

	f = conv.file(sf)
	if errlist.Len() == 0 {
		return f, nil
	}
	return f, errlist
}

// ParseDir calls ParseFile for all files with names ending in ".go" in the
// directory specified by path and returns a map of package name -> package
// AST with all the packages found.
//
// If filter != nil, only the files with os.FileInfo entries passing through
// the filter (and ending in ".go") are considered. The mode bits are passed
// to ParseFile unchanged. Position information is recorded in fset, which
// must not be nil.
//
// If the directory couldn't be read, a nil map and the respective error are
// returned. If a parse error occurred, a non-nil but incomplete map and the
// first error encountered are returned.
//
func ParseDir(fset *token.FileSet, path string, filter func(os.FileInfo) bool, mode Mode) (pkgs map[string]*ast.Package, first error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	list, err := fd.Readdir(-1)
	if err != nil {
		return nil, err
	}

	pkgs = make(map[string]*ast.Package)
	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".go") && (filter == nil || filter(d)) {
			filename := filepath.Join(path, d.Name())
			if src, err := ParseFile(fset, filename, nil, mode); err == nil {
				name := src.Name.Name
				pkg, found := pkgs[name]
				if !found {
					pkg = &ast.Package{
						Name:  name,
						Files: make(map[string]*ast.File),
					}
					pkgs[name] = pkg
				}
				pkg.Files[filename] = src
			} else if first == nil {
				first = err
			}
		}
	}

	return
}

// ParseExprFrom is a convenience function for parsing an expression.
// The arguments have the same meaning as for ParseFile, but the source must
// be a valid Go (type or value) expression. Specifically, fset must not
// be nil.
//
// If the source couldn't be read, the returned AST is nil and the error
// indicates the specific failure. If the source was read but syntax
// errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors
// are returned via a scanner.ErrorList which is sorted by source position.
//
func ParseExprFrom(fset *token.FileSet, filename string, src interface{}, mode Mode) (expr ast.Expr, err error) {
	// TODO(mvdan): implement
	return nil, nil
}

// ParseExpr is a convenience function for obtaining the AST of an expression x.
// The position information recorded in the AST is undefined. The filename used
// in error messages is the empty string.
//
// If syntax errors were found, the result is a partial AST (with ast.Bad* nodes
// representing the fragments of erroneous source code). Multiple errors are
// returned via a scanner.ErrorList which is sorted by source position.
//
func ParseExpr(x string) (ast.Expr, error) {
	return ParseExprFrom(token.NewFileSet(), "", []byte(x), 0)
}
