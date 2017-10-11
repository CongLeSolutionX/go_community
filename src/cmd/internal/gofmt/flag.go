// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gofmt

import "flag"

// Flag is s struct that represents flags of gofmt.
type Flag struct {
	List        bool
	Write       bool
	RewriteRule string
	SimplifyAST bool
	DoDiff      bool
	AllErrors   bool
	Cpuprofile  string
}

// Args provides the contents of the gofmt flags in a format that is available on the command line.
func (f *Flag) Args() []string {
	var args []string

	if f.AllErrors {
		args = append(args, "-e")
	}
	if f.Cpuprofile != "" {
		args = append(args, "-cpuprofile", f.Cpuprofile)
	}
	if f.DoDiff {
		args = append(args, "-d")
	}
	if f.List {
		args = append(args, "-l")
	}
	if f.RewriteRule != "" {
		args = append(args, "-r", f.RewriteRule)
	}
	if f.SimplifyAST {
		args = append(args, "-s")
	}
	if f.Write {
		args = append(args, "-w")
	}

	return args
}

// Register gofmt flags to given FlagSet.
func (f *Flag) InitGofmtFlag(fs *flag.FlagSet) {
	// main operation modes
	fs.BoolVar(&f.List, "l", false, "list files whose formatting differs from gofmt's")
	fs.BoolVar(&f.Write, "w", false, "write result to (source) file instead of stdout")
	fs.StringVar(&f.RewriteRule, "r", "", "rewrite rule (e.g., 'a[b:len(a)] -> a[b:]')")
	fs.BoolVar(&f.SimplifyAST, "s", false, "simplify code")
	fs.BoolVar(&f.DoDiff, "d", false, "display diffs instead of rewriting files")
	fs.BoolVar(&f.AllErrors, "e", false, "report all errors (not just the first 10 on different lines)")

	// debugging
	fs.StringVar(&f.Cpuprofile, "cpuprofile", "", "write cpu profile to this file")
}
