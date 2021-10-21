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
	"internal/coverage/encodecounter"
	"internal/coverage/encodemeta"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func makeSubtractIntersectOp(verb string) covOperation {
	outdirflag = flag.String("o", "", "Output directory to write")
	s := &sstate{
		pkm:   make(map[string]*pkstate),
		verb:  verb,
		inidx: -1,
	}
	return s
}

type sstate struct {
	cov.BatchCounterAlloc
	pkm   map[string]*pkstate // maps package path to package state
	pkgs  []*pkstate
	indir string
	inidx int
	p     *pkstate
	pod   *podstate
	verb  string
	imm   map[pkfunc]struct{}
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
	s.pod = &podstate{
		mm: make(map[pkfunc]decodecounter.FuncPayload),
	}
}

func (s *sstate) EndPod(p cov.Pod) {
	// Copy meta-data file for this pod to the output directory.
	inpath := s.pod.mdf
	mdfbase := filepath.Base(s.pod.mdf)
	outpath := filepath.Join(*outdirflag, mdfbase)
	copyFile(inpath, outpath)

	// Emit acccumulated counter data for this pod.
	s.emitCounters(*outdirflag, s.pod.fileHash)

	// Reset package state.
	s.pkm = make(map[string]*pkstate)
	s.pkgs = nil
	s.pod = nil
}

func (s *sstate) prune() {
	keys := make([]pkfunc, 0, len(s.pod.mm))
	for k := range s.pod.mm {
		keys = append(keys, k)
	}
	for _, k := range keys {
		if _, found := s.imm[k]; !found {
			delete(s.pod.mm, k)
		}
	}
}

func (s *sstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visiting counter data file %s diridx %d", cdf, dirIdx)
	if s.inidx != dirIdx {
		s.inidx = dirIdx
		// We're switching in-dirs.
		if dirIdx == 0 {
			s.imm = nil
		} else {
			if s.imm != nil {
				s.prune()
			}
			if s.verb == "intersect" {
				s.imm = make(map[pkfunc]struct{})
			}
		}
	}
}

func (s *sstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}

	if *verbflag >= 5 {
		fmt.Printf("ctr visit fid=%d pk=%d inidx=%d data.Counters=%+v\n", data.FuncIdx, data.PkgIdx, s.inidx, data.Counters)
	}

	// If we're processing counter data from the direct input
	// directory, then just install it into the counter data map
	// as usual.
	if s.inidx == 0 {
		val, found := s.pod.mm[key]
		if !found {
			val.Counters = s.AllocateCounters(len(data.Counters))
		} else if len(val.Counters) != len(data.Counters) {
			if len(val.Counters) < len(data.Counters) {
				t := val.Counters
				val.Counters = s.AllocateCounters(len(data.Counters))
				copy(val.Counters, t)
			}
		}
		for i := 0; i < len(data.Counters); i++ {
			val.Counters[i] += data.Counters[i]
		}
		s.pod.mm[key] = val
		return
	}
	// If we're looking at counter data from a dir other than
	// the first, then perform the intersect/subtract.
	if val, ok := s.pod.mm[key]; ok {
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
	if s.imm != nil {
		s.prune()
		s.imm = nil
	}

	// Record meta-data file name.
	s.pod.mdf = mdf
	// Record file hash.
	s.pod.fileHash = mfr.FileHash()
}

func (s *sstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	// Install an entry for this package if needed.
	p, ok := s.pkm[pd.PackagePath()]
	if !ok {
		cmdb, err := encodemeta.NewCoverageMetaDataBuilder(pd.PackagePath(), pd.PackageClassification())
		if err != nil {
			fatal("fatal error creating meta-data builder: %v", err)
		}
		p = &pkstate{
			pkgIdx: uint32(len(s.pkgs)),
			cmdb:   cmdb,
			ftab:   make(map[[16]byte]uint32),
			ctab:   make(map[uint32]decodecounter.FuncPayload),
		}
		s.pkgs = append(s.pkgs, p)
		s.pkm[pd.PackagePath()] = p
		if *verbflag > 1 {
			fmt.Printf("VisitPackage(%d) pkpath=%s\n", pkgIdx, pd.PackagePath())
		}
	}
	s.p = p
}

func (s *sstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	key := pkfunc{pk: pkgIdx, fcn: fnIdx}
	v, haveCounters := s.pod.mm[key]
	if !haveCounters {
		return
	}

	if *verbflag >= 5 {
		fmt.Printf("meta visit fid=%d pk=%d fname=%s file=%s found=%v len(val.ctrs)=%d\n", fnIdx, pkgIdx, fd.Funcname, fd.Srcfile, haveCounters, len(v.Counters))
	}

	// Install counters in package ctab.
	_, ok := s.p.ctab[fnIdx]
	if ok {
		panic("should never see this for merge/subtract")
	}
	s.p.ctab[fnIdx] = v
}

func (s *sstate) NumFuncs() (int, error) {
	rval := 0
	for _, p := range s.pkgs {
		rval += len(p.ctab)
	}
	return rval, nil
}

func (s *sstate) VisitFuncs(f encodecounter.CounterVisitorFn) error {
	// For each package, for each function, construct counter
	// array and then call "f" on it.
	for pidx, p := range s.pkgs {
		fids := make([]int, 0, len(p.ctab))
		for fid := range p.ctab {
			fids = append(fids, int(fid))
		}
		sort.Ints(fids)
		for _, fid := range fids {
			fp := p.ctab[uint32(fid)]
			if err := f(uint32(pidx), uint32(fid), fp.Counters); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *sstate) emitCounters(outdir string, metaHash [16]byte) {

	// FIXME: it might be better to preserve the process ID from the
	// original source as opposed using the process ID of the merge
	// tool here.
	processID := os.Getpid()

	// Open output file.
	fn := fmt.Sprintf(coverage.CounterFileTempl, coverage.CounterFilePref, metaHash, processID, time.Now().UnixNano())
	fpath := filepath.Join(outdir, fn)
	cf, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fatal("opening counter data file %s: %v", fpath, err)
	}
	defer func() {
		if err := cf.Close(); err != nil {
			fatal("error closing output meta-data file %s: %v", fpath, err)
		}
	}()

	// FIXME: do something with args
	var args map[string]string
	cfw := encodecounter.NewCoverageDataFileWriter(cf, coverage.CtrULeb128)
	if err := cfw.Write(metaHash, args, s); err != nil {
		fatal("counter file write failed: %v", err)
	}
}

func (s *sstate) Finish() {
}
