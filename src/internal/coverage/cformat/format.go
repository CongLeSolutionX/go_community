// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cformat

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
}

type pstate struct {
	um map[filecu]uint32
}

type filecu struct {
	file string
	coverage.CoverableUnit
}

func NewFormatter() *Formatter {
	return &Formatter{
		pm: make(map[string]*pstate),
	}
}

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

func (fm *Formatter) AddUnit(file string, unit coverage.CoverableUnit, count uint32) {
	key := filecu{file: file, CoverableUnit: unit}
	pcount := fm.p.um[key]
	p, c := uint64(pcount), uint64(count)
	sum := p + c
	if uint64(uint32(sum)) != sum {
		sum = math.MaxUint32
	}
	fm.p.um[key] = uint32(sum)
}

func (fm *Formatter) EmitTextual(w io.Writer, cm coverage.CounterMode) error {
	if cm == coverage.CtrModeInvalid {
		panic("internal error, counter mode unset")
	}
	if _, err := fmt.Fprintf(w, "mode: %s\n", cm.String()); err != nil {
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

func (fm *Formatter) EmitPercent(w io.Writer) error {
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
		if _, err := fmt.Fprintf(w, "coverage: %.1f%% of statements\n", 100*float64(coveredStmts)/float64(totalStmts)); err != nil {
			return err
		}
	}
	return nil
}
