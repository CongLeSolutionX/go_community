// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/cov"
	"flag"
	"fmt"
	"internal/coverage"
	"internal/coverage/decodecounter"
	"internal/coverage/decodemeta"
	"os"
	"strings"
)

func makeSubtractIntersectOp(verb string) covOperation {
	outdirflag = flag.String("o", "", "Output directory to write")
	s := &sstate{
		verb:  verb,
		inidx: -1,
	}
	s.initMetaMerge()
	return s
}

// sstate holds state needed to implement subtraction and intersection
// operations on code coverage data files.
type sstate struct {
	metaMerge
	indir string
	inidx int
	verb  string
	// Used only for intersection; keyed by pkg/fn ID, it keeps track of
	// just the set of functions for which we have data in the current
	// input directory.
	imm map[pkfunc]struct{}
}

func (s *sstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata %s -i=dir1,dir2 -o=<dir>\n\n", s.verb)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	op := "from"
	if s.verb == "intersect" {
		op = "with"
	}
	fmt.Fprintf(os.Stderr, "  go tool covdata %s -i=dir1,dir2 -o=outdir\n\n", s.verb)
	fmt.Fprintf(os.Stderr, "  \t%ss dir2 %s dir1, writing result\n", s.verb, op)
	fmt.Fprintf(os.Stderr, "  \tinto output dir outdir.\n")
	os.Exit(2)
}

func (s *sstate) Setup() {
	if *indirsflag == "" {
		usage("select input directories with '-i' option")
	}
	indirs := strings.Split(*indirsflag, ",")
	if s.verb == "subtract" && len(indirs) != 2 {
		usage("supply exactly two input dirs for subtract operation")
	}
	if *outdirflag == "" {
		usage("select output directory with '-o' option")
	}
}

func (s *sstate) BeginPod(p cov.Pod) {
	s.metaBeginPod()
}

func (s *sstate) EndPod(p cov.Pod) {
	const pcombine = false
	s.metaEndPod(pcombine)
}

func (s *sstate) EndCounters() {
	if s.imm != nil {
		s.pruneCounters()
	}
}

// pruneCounters performs a function-level partial intersection using the
// current POD counter data (s.pod.pmm) and the intersected data from
// PODs in previous dirs (s.imm).
func (s *sstate) pruneCounters() {
	pkeys := make([]pkfunc, 0, len(s.pod.pmm))
	for k := range s.pod.pmm {
		pkeys = append(pkeys, k)
	}
	// Remove anything from pmm not found in imm. We don't need to
	// go the other way (removing things from imm not found in pmm)
	// since we don't add anything to imm if there is no pmm entry.
	for _, k := range pkeys {
		if _, found := s.imm[k]; !found {
			delete(s.pod.pmm, k)
		}
	}
	s.imm = nil
}

func (s *sstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visiting counter data file %s diridx %d", cdf, dirIdx)
	if s.inidx != dirIdx {
		if s.inidx > dirIdx {
			// We're relying on having data files presented in
			// the order they appear in the inputs (e.g. first all
			// data files from input dir 0, then dir 1, etc).
			panic("decreasing dir index, internal error")
		}
		if dirIdx == 0 {
			// No need to keep track of the functions in the first
			// directory, since that info will be replicated in
			// s.pod.pmm.
			s.imm = nil
		} else {
			// We're now starting to visit the Nth directory, N != 0.
			if s.verb == "intersect" {
				if s.imm != nil {
					s.pruneCounters()
				}
				s.imm = make(map[pkfunc]struct{})
			}
		}
		s.inidx = dirIdx
	}
}

func (s *sstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}

	if *verbflag >= 5 {
		fmt.Printf("ctr visit fid=%d pk=%d inidx=%d data.Counters=%+v\n", data.FuncIdx, data.PkgIdx, s.inidx, data.Counters)
	}

	// If we're processing counter data from the initial (first) input
	// directory, then just install it into the counter data map
	// as usual.
	if s.inidx == 0 {
		s.metaVisitFuncCounterData(data)
		return
	}

	// If we're looking at counter data from a dir other than
	// the first, then perform the intersect/subtract.
	if val, ok := s.pod.pmm[key]; ok {
		if s.verb == "subtract" {
			for i := 0; i < len(data.Counters); i++ {
				if data.Counters[i] != 0 {
					val.Counters[i] = 0
				}
			}
		} else if s.verb == "intersect" {
			s.imm[key] = struct{}{}
			for i := 0; i < len(data.Counters); i++ {
				if data.Counters[i] == 0 {
					val.Counters[i] = 0
				}
			}
		}
	}
}

func (s *sstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	if s.verb == "intersect" {
		s.imm = make(map[pkfunc]struct{})
	}
	s.metaVisitMetaDataFile(mdf, mfr)
}

func (s *sstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	s.metaVisitPackage(pd, pkgIdx, false)
}

func (s *sstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	s.metaVisitFunc(pkgIdx, fnIdx, fd, s.verb, false)
}

func (s *sstate) Finish() {
}
