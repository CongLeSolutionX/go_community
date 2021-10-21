// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cformat

// This package provides apis for producing human-readable summaries
// of coverage data (e.g. a coverage percentage for a given package or
// set of packages) and for writing data in the legacy test format
// emitted by "go test -coverprofile=<outfile>".
//
// The model for using these apis is to create a Formatter object,
// then make a series of calls to SetPackage and AddUnit passing in
// data read from coverage meta-data and counter-data files. E.g.
//
//		myformatter := cformat.NewFormatter()
//		...
//		for each package P in meta-data file: {
//			myformatter.SetPackage(P)
//			for each function F in P: {
//				for each coverable unit U in F: {
//					myformatter.AddUnit(U)
//				}
//			}
//		}
//		myformatter.EmitPercent(os.Stdout, "")
//		myformatter.EmitTextual(somefile)
//
// These apis are linked into tests that are built with "-cover", and
// called at the end of test execution to produce text output or
// emit coverage percentages.

import (
	"fmt"
	"internal/coverage"
	"internal/coverage/cmerge"
	"io"
	"sort"
)

type Formatter struct {
	// Maps import path to package state.
	pm map[string]*pstate
	// Records current package being visited.
	pkg string
	// Pointer to current package state.
	p *pstate
	// Counter mode.
	cm coverage.CounterMode
}

type pstate struct {
	// Key here is coverable unit (with source file), value is
	// coverage count.
	unitTable map[filecu]uint32
}

// filecu is a specific coverage.CoverableUnit within a given source file.
type filecu struct {
	file string
	coverage.CoverableUnit
}

func NewFormatter(cm coverage.CounterMode) *Formatter {
	return &Formatter{
		pm: make(map[string]*pstate),
		cm: cm,
	}
}

// SetPackage tells the formatter that we're about to visit the
// coverage data for the package with the specified import path.
// Note that it's OK to call SetPackage more than once with the
// same import path; counter data values will be accumulated.
func (fm *Formatter) SetPackage(importpath string) {
	if importpath == fm.pkg {
		return
	}
	fm.pkg = importpath
	ps, ok := fm.pm[importpath]
	if !ok {
		ps = new(pstate)
		fm.pm[importpath] = ps
		ps.unitTable = make(map[filecu]uint32)

	}
	fm.p = ps
}

// AddUnit passes info on a single coverable unit (file, range of
// lines, and counter value) to the formatter. Counter values
// be accumulated where appropriate.
func (fm *Formatter) AddUnit(file string, unit coverage.CoverableUnit, count uint32) {
	key := filecu{file: file, CoverableUnit: unit}
	pcount := fm.p.unitTable[key]
	var result uint32
	if fm.cm == coverage.CtrModeSet {
		if count != 0 || pcount != 0 {
			result = 1
		}
	} else {
		// Use saturating arithmetic.
		result, _ = cmerge.SaturatingAdd(pcount, count)
	}
	fm.p.unitTable[key] = result
}

// EmitTextual writes the accumulated coverage data in the legacy text
// format to the writer 'w'. We sort the data items by importpath,
// source file, and line number before emitting (this is not
// explicitly mandated by the format, but seems like a good idea).
func (fm *Formatter) EmitTextual(w io.Writer) error {
	if fm.cm == coverage.CtrModeInvalid {
		panic("internal error, counter mode unset")
	}
	if _, err := fmt.Fprintf(w, "mode: %s\n", fm.cm.String()); err != nil {
		return err
	}
	pkgs := make([]string, 0, len(fm.pm))
	for importpath := range fm.pm {
		pkgs = append(pkgs, importpath)
	}
	sort.Strings(pkgs)
	for _, importpath := range pkgs {
		p := fm.pm[importpath]
		units := make([]filecu, 0, len(p.unitTable))
		for u := range p.unitTable {
			units = append(units, u)
		}
		sort.Slice(units, func(i, j int) bool {
			if units[i].file != units[j].file {
				return units[i].file < units[j].file
			}
			if units[i].StLine != units[j].StLine {
				return units[i].StLine < units[j].StLine
			}
			if units[i].EnLine != units[j].EnLine {
				return units[i].EnLine < units[j].EnLine
			}
			if units[i].StCol != units[j].StCol {
				return units[i].StCol < units[j].StCol
			}
			if units[i].EnCol != units[j].EnCol {
				return units[i].EnCol < units[j].EnCol
			}
			return units[i].NxStmts < units[j].NxStmts
		})
		for _, u := range units {
			count := p.unitTable[u]
			if _, err := fmt.Fprintf(w, "%s:%d.%d,%d.%d %d %d\n",
				u.file, u.StLine, u.StCol,
				u.EnLine, u.EnCol, u.NxStmts, count); err != nil {
				return err
			}
		}
	}
	return nil
}

// EmitPercent writes out a "percentage covered" string to the writer 'w'.
func (fm *Formatter) EmitPercent(w io.Writer, covpkgs string, noteEmpty bool) error {
	pkgs := make([]string, 0, len(fm.pm))
	for importpath := range fm.pm {
		pkgs = append(pkgs, importpath)
	}
	sort.Strings(pkgs)
	seenPkg := false
	for _, importpath := range pkgs {
		seenPkg = true
		p := fm.pm[importpath]
		var totalStmts, coveredStmts uint64
		for unit, count := range p.unitTable {
			nx := uint64(unit.NxStmts)
			totalStmts += nx
			if count != 0 {
				coveredStmts += nx
			}
		}
		if _, err := fmt.Fprintf(w, "\t%s\t", importpath); err != nil {
			return err
		}
		if totalStmts == 0 {
			if _, err := fmt.Fprintf(w, "coverage: [no statements]\n"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "coverage: %.1f%% of statements%s\n", 100*float64(coveredStmts)/float64(totalStmts), covpkgs); err != nil {
				return err
			}
		}
	}
	if noteEmpty && !seenPkg {
		if _, err := fmt.Fprintf(w, "coverage: [no statements]\n"); err != nil {
			return err
		}
	}

	return nil
}
