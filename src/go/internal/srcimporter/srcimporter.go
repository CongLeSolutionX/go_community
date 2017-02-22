// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package srcimporter implements importing directly
// from source files rather than installed packages.
package srcimporter // import "go/internal/srcimporter"

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
)

// An Importer provides the context for importing packages from source code.
type Importer struct {
	ctxt     build.Context
	sizes    types.Sizes
	fset     *token.FileSet
	packages map[string]*types.Package
}

// NewImporter returns a new Importer for the given context.
func NewImporter(ctxt build.Context) *Importer {
	// use same size computation as gc for type-checking
	wordSize := 8
	maxAlign := 8
	switch build.Default.GOARCH {
	case "386", "arm":
		wordSize = 4
		maxAlign = 4
		// add more cases as needed
	}

	return &Importer{
		ctxt:     ctxt,
		sizes:    &types.StdSizes{WordSize: int64(wordSize), MaxAlign: int64(maxAlign)},
		fset:     token.NewFileSet(),
		packages: make(map[string]*types.Package),
	}
}

// Importing is a sentinel taking the place in Importer.packages
// for a package that is in the process of being imported.
var importing types.Package

// Import(path) is a shortcut for ImportFrom(path, "", 0).
func (p *Importer) Import(path string) (*types.Package, error) {
	return p.ImportFrom(path, "" /* no vendoring */, 0)
}

// ImportFrom imports the package with the given import path resolved from the given srcDir.
// The import mode must be zero but is otherwise ignored.
func (p *Importer) ImportFrom(path, srcDir string, mode types.ImportMode) (pkg *types.Package, err error) {
	if mode != 0 {
		panic("non-zero import mode")
	}

	// determine package path (do vendor resolution)
	var bp *build.Package
	switch {
	default:
		if abs, err := filepath.Abs(srcDir); err == nil { // see issue 14282
			srcDir = abs
		}
		bp, err = p.ctxt.Import(path, srcDir, build.FindOnly)

	case build.IsLocalImport(path):
		// "./x" -> "srcDir/x"
		bp, err = p.ctxt.ImportDir(filepath.Join(srcDir, path), build.FindOnly)

	case filepath.IsAbs(path):
		return nil, fmt.Errorf("invalid absolute import path %q", path)
	}
	if err != nil {
		return // do not modify *build.NoGoError!
	}

	// package unsafe is known to the type checker
	if bp.ImportPath == "unsafe" {
		return types.Unsafe, nil
	}

	// no need to re-import if the package was imported completely before
	pkg = p.packages[bp.ImportPath]
	if pkg != nil {
		if pkg == &importing {
			return nil, fmt.Errorf("import cycle through package %q", bp.ImportPath)
		}
		if pkg.Complete() {
			return
		}
	} else {
		p.packages[bp.ImportPath] = &importing
	}

	// collect package files
	bp, err = p.ctxt.ImportDir(bp.Dir, 0)
	if err != nil {
		return // do not modify *build.NoGoError!
	}
	var filenames []string
	filenames = append(filenames, bp.GoFiles...)
	filenames = append(filenames, bp.CgoFiles...)

	// parse package files
	var files []*ast.File
	for _, filename := range filenames {
		f, err := parser.ParseFile(p.fset, filepath.Join(bp.Dir, filename), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parsing package file %s failed (%v)", filename, err)
		}
		files = append(files, f)
	}

	// type-check package files
	conf := types.Config{
		IgnoreFuncBodies: true,
		FakeImportC:      true,
		Importer:         p,
		Sizes:            p.sizes,
	}
	pkg, err = conf.Check(bp.ImportPath, p.fset, files, nil)
	if err != nil {
		return nil, fmt.Errorf("type-checking package %q failed (%v)", bp.ImportPath, err)
	}

	p.packages[bp.ImportPath] = pkg
	return
}
