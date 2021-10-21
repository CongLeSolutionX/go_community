// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cformat

// This package provides apis emit coverage data formatted for human
// consumption (a coverage percentage for a given package or set of
// packages) and in the legacy test format emitted by "go test
// -coverprofile=<outfile>".
//
// The model for using these apis is to create a Formatter object,
// then make a series of calls to SetPackage and AddUnit passing
// in data read from coverage meta-data and counter-data files.
// E.g.
//
//    myformatter := cformat.NewFormatter()
//    ...
//    for each package P in meta-data file: {
//       myformatter.SetPackage(P)
//       for each function F in P: {
//          for each coverable unit U in F: {
//             myformatter.AddUnit(U)
//          }
//       }
//    }
//    myformatter.EmitPercent(os.Stdout, "")
//    myformatter.EmitTextual(somefile)
//
// These apis are linked into tests that are built with "-cover", and
// called at the end of test execution to produce text output or
// emit coverage percentages.

import (
	"fmt"
	"internal/coverage"
	"io"
	"math"
	"sort"
)

type Formatter struct {
	pkg string
	pm  map[string]*pstate
	p   *pstate
	cm  coverage.CounterMode
}

type pstate struct {
	um map[filecu]uint32
}

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
// same import path; it will accumulated
func (fm *Formatter) SetPackage(importpath string) {
	if importpath == fm.pkg {
		return
	}
	fm.pkg = importpath
	ps, ok := fm.pm[importpath]
	if !ok {
		ps = new(pstate)
		fm.pm[importpath] = ps
		ps.um = make(map[filecu]uint32)

	}
	fm.p = ps
}

// AddUnit passes info on a single coverable unit (file, range of
// lines, and counter value) to the formatter. Counter values
// be accumulated where appropriate.
func (fm *Formatter) AddUnit(file string, unit coverage.CoverableUnit, count uint32) {
	key := filecu{file: file, CoverableUnit: unit}
	pcount := fm.p.um[key]
	var result uint32
	if fm.cm == coverage.CtrModeSet {
		if count != 0 || pcount != 0 {
			result = 1
		}
	} else {
		// Use saturating arithmetic.
		p, c := uint64(pcount), uint64(count)
		sum := p + c
		if uint64(uint32(sum)) != sum {
			sum = math.MaxUint32
		}
		result = uint32(sum)
	}
	fm.p.um[key] = result
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
		units := make([]filecu, 0, len(p.um))
		for u := range p.um {
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
			count := p.um[u]
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
func (fm *Formatter) EmitPercent(w io.Writer, covpkgs string) error {
	var totalStmts, coveredStmts uint64
	for _, pk := range fm.pm {
		for unit, count := range pk.um {
			nx := uint64(unit.NxStmts)
			totalStmts += nx
			if count != 0 {
				coveredStmts += nx
			}
		}
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
	return nil
}
