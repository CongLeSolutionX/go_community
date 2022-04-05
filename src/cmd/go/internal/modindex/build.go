// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modindex

import (
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"sort"
	"strings"
)

var defaultToolTags, defaultReleaseTags []string

// NoGoError is the error used by Import to describe a directory
// containing no buildable Go source files. (It may still contain
// test files, files hidden by build tags, and so on.)
type NoGoError struct {
	Dir string
}

func (e *NoGoError) Error() string {
	return "no buildable Go source files in " + e.Dir
}

// MultiplePackageError describes a directory containing
// multiple buildable Go source files for multiple packages.
type MultiplePackageError struct {
	Dir      string   // directory containing files
	Packages []string // package names found
	Files    []string // corresponding files: Files[i] declares package Packages[i]
}

func (e *MultiplePackageError) Error() string {
	// Error string limited to two entries for compatibility.
	return fmt.Sprintf("found packages %s (%s) and %s (%s) in %s", e.Packages[0], e.Files[0], e.Packages[1], e.Files[1], e.Dir)
}

func nameExt(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return ""
	}
	return name[i:]
}

func fileListForExt(p *build.Package, ext string) *[]string {
	switch ext {
	case ".c":
		return &p.CFiles
	case ".cc", ".cpp", ".cxx":
		return &p.CXXFiles
	case ".m":
		return &p.MFiles
	case ".h", ".hh", ".hpp", ".hxx":
		return &p.HFiles
	case ".f", ".F", ".for", ".f90":
		return &p.FFiles
	case ".s", ".S", ".sx":
		return &p.SFiles
	case ".swig":
		return &p.SwigFiles
	case ".swigcxx":
		return &p.SwigCXXFiles
	case ".syso":
		return &p.SysoFiles
	}
	return nil
}

var errNoModules = errors.New("not using modules")

var (
	slashSlash = []byte("//")
	slashStar  = []byte("/*")
	starSlash  = []byte("*/")
	newline    = []byte("\n")
)

var dummyPkg build.Package

func cleanDecls(m map[string][]token.Position) ([]string, map[string][]token.Position) {
	all := make([]string, 0, len(m))
	for path := range m {
		all = append(all, path)
	}
	sort.Strings(all)
	return all, m
}

var (
	bSlashSlash = []byte(slashSlash)
	bStarSlash  = []byte(starSlash)
	bSlashStar  = []byte(slashStar)
	bPlusBuild  = []byte("+build")

	goBuildComment = []byte("//go:build")

	errGoBuildWithoutBuild = errors.New("//go:build comment without // +build comment")
	errMultipleGoBuild     = errors.New("multiple //go:build comments")
)
