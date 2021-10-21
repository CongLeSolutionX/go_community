// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/cov"
	"crypto/md5"
	"flag"
	"fmt"
	"internal/coverage"
	"internal/coverage/decodecounter"
	"internal/coverage/decodemeta"
	"internal/coverage/encodecounter"
	"internal/coverage/encodemeta"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
	"unsafe"
)

var outdirflag *string
var pcombineflag *bool

func makeMergeOp() covOperation {
	outdirflag = flag.String("o", "", "Output directory to write")
	pcombineflag = flag.Bool("pcombine", false, "Combine profiles derived from distinct program executables")
	m := &mstate{
		pkm:   make(map[string]*pkstate),
		pktab: make(map[string]uint32),
		fntab: make(map[string]uint32),
	}
	if *pcombineflag {
		m.pktab = make(map[string]uint32)
		m.fntab = make(map[string]uint32)
	}
	return m
}

type mstate struct {
	pkm     map[string]*pkstate // maps package path to package state
	pkgs    []*pkstate
	indir   string
	p       *pkstate
	pod     *podstate
	gmm     map[pkfunc]decodecounter.FuncPayload
	pktab   map[string]uint32
	fntab   map[string]uint32
	pool    []uint32
	modname string
	batchCounterAlloc
}

type podstate struct {
	mm       map[pkfunc]decodecounter.FuncPayload
	mdf      string
	fileHash [16]byte
}

type pkstate struct {
	pkgIdx uint32
	ctab   map[uint32]decodecounter.FuncPayload
	// these below are filled in only in -pcombine mode.
	cmdb *encodemeta.CoverageMetaDataBuilder
	ftab map[[16]byte]uint32
}

type pkfunc struct {
	pk, fcn uint32
}

type batchCounterAlloc struct {
	pool []uint32
}

func (ca *batchCounterAlloc) allocateCounters(n int) []uint32 {
	const chunk = 8192
	if n > cap(ca.pool) {
		if n < chunk {
			n = chunk
		}
		ca.pool = make([]uint32, n)
	}
	rv := ca.pool[:n]
	ca.pool = ca.pool[n:]
	return rv
}

func (m *mstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: covtool merge -i=<directories> -o=<dir>\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  covtool merge -i=dir1,dir2,dir3 -o=outdir\n\n")
	fmt.Fprintf(os.Stderr, "  \tmerges all files in dir1/dir2/dir3\n")
	fmt.Fprintf(os.Stderr, "  \tinto output dir outdir\n")
	os.Exit(2)
}

func (m *mstate) Setup() {
	if *indirsflag == "" {
		usage("select input directories with '-i' option")
	}
	if *outdirflag == "" {
		usage("select output directory with '-o' option")
	}
}

func (m *mstate) VisitIndir(idx uint32, indir string) {
	m.indir = indir
}

func (m *mstate) BeginPod(p cov.Pod) {
	m.pod = &podstate{
		mm: make(map[pkfunc]decodecounter.FuncPayload),
	}
}

func copyFile(inpath, outpath string) {
	inf, err := os.Open(inpath)
	if err != nil {
		fatal("opening input meta-data file %s: %v", inpath, err)
	}
	defer inf.Close()

	fi, err := inf.Stat()
	if err != nil {
		fatal("accessing input meta-data file %s: %v", inpath, err)
	}

	outf, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
	if err != nil {
		fatal("opening output meta-data file %s: %v", outpath, err)
	}

	_, err = io.Copy(outf, inf)
	outf.Close()
	if err != nil {
		fatal("writing output meta-data file %s: %v", outpath, err)
	}
}

func (m *mstate) EndPod(p cov.Pod) {
	if !*pcombineflag {
		// Copy meta-data file for this pod to the output directory.
		inpath := m.pod.mdf
		mdfbase := filepath.Base(m.pod.mdf)
		outpath := filepath.Join(*outdirflag, mdfbase)
		copyFile(inpath, outpath)

		// Emit acccumulated counter data for this pod.
		m.emitCounters(*outdirflag, m.pod.fileHash)

		// Reset package state.
		m.pkm = make(map[string]*pkstate)
		m.pkgs = nil
	}
	m.pod = nil
}

func (m *mstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader) {
	verb(2, "visit counter data file %s", cdf)
}

func (m *mstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}
	val := m.pod.mm[key]
	// FIXME: in theory either A) len(val.Counters) is zero, or B)
	// the two lengths are equal. Assert if not? Of course, we could
	// see odd stuff if there is source file skew.
	verb(5, "visit fid=%d pk=%d len(counters)=%d", data.FuncIdx, data.PkgIdx,
		len(data.Counters))
	if len(val.Counters) < len(data.Counters) {
		t := val.Counters
		val.Counters = m.allocateCounters(len(data.Counters))
		copy(val.Counters, t)
	}
	for i := 0; i < len(data.Counters); i++ {
		val.Counters[i] += data.Counters[i]
	}
	m.pod.mm[key] = val
}

func (m *mstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	// Record module name.
	if m.modname == "" {
		m.modname = mfr.ModuleName()
	} else if m.modname != mfr.ModuleName() {
		m.modname = "<various>"
	}
	// Record meta-data file name.
	m.pod.mdf = mdf
	// Record file hash.
	m.pod.fileHash = mfr.FileHash()
}

func (m *mstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	// Install an entry for this package if needed.
	p, ok := m.pkm[pd.PackagePath()]
	if !ok {
		p = &pkstate{
			pkgIdx: uint32(len(m.pkgs)),
			cmdb:   encodemeta.NewCoverageMetaDataBuilder(pd.PackagePath()),
			ftab:   make(map[[16]byte]uint32),
			ctab:   make(map[uint32]decodecounter.FuncPayload),
		}
		m.pkgs = append(m.pkgs, p)
		m.pkm[pd.PackagePath()] = p
	}
	m.p = p
}

func counterMerge(dst, src []uint32) []uint32 {
	if len(src) != len(dst) {
		panic("bad")
	}
	for i := 0; i < len(src); i++ {
		// FIXME: handle different counter flavors
		// FIXME: handle overflow
		if src[i] != 0 {
			dst[i] = 1
		}
	}
	return dst
}

func (m *mstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	var counters []uint32
	key := pkfunc{pk: pkgIdx, fcn: fnIdx}
	v, haveCounters := m.pod.mm[key]
	if haveCounters {
		counters = v.Counters
	} else if *liveflag {
		return
	}

	verb(3, "visit pk=%d fid=%d func %s", pkgIdx, fnIdx, fd.Funcname)

	// If the merge is running in "combine programs" mode, then hash
	// the function and look it up in the package ftab to see if we've
	// encountered it before. If we haven't, then register it with the
	// meta-data builder.
	if *pcombineflag {
		fnhash := encodemeta.HashFuncDesc(fd)
		gfidx, ok := m.p.ftab[fnhash]
		if !ok {
			// We haven't seen this function before, need to add it to
			// the meta data.
			gfidx = uint32(m.p.cmdb.AddFunc(*fd))
			m.p.ftab[fnhash] = gfidx
			verb(3, "new meta entry for fn %s fid=%d", fd.Funcname, gfidx)
		}
		fnIdx = gfidx
	}
	if !haveCounters {
		return
	}

	// Install counters in package ctab.
	gfp, ok := m.p.ctab[fnIdx]
	if ok {
		verb(3, "counter merge for %s fidx=%d", fd.Funcname, fnIdx)
		// Merge.
		gfp.Counters = counterMerge(gfp.Counters, counters)
		m.p.ctab[fnIdx] = gfp
	} else {
		verb(3, "null merge for %s fidx %d", fd.Funcname, fnIdx)
		gfp := v
		gfp.PkgIdx = m.p.pkgIdx
		gfp.FuncIdx = fnIdx
		m.p.ctab[fnIdx] = gfp
	}
}

func (m *mstate) emitMeta(outdir string) [16]byte {
	// This implementation emits encoded meta-data to in-memory buffers,
	// which is very inefficient. It would be better to refactor things
	// so that as soon as we're done processing a package we write out
	// its encoded meta-data to a temp file.
	fh := md5.New()
	blobs := [][]byte{}
	tlen := uint64(unsafe.Sizeof(coverage.MetaFileHeader{}))
	for _, p := range m.pkgs {
		mdw := &mdWriteSeeker{}
		p.cmdb.Emit(mdw)
		blob := mdw.payload
		ph := md5.Sum(blob)
		blobs = append(blobs, blob)
		if _, err := fh.Write(ph[:]); err != nil {
			panic(fmt.Sprintf("internal error: md5 sum failed: %v", err))
		}
		tlen += uint64(len(blob))
	}
	var finalHash [16]byte
	fhh := fh.Sum(nil)
	copy(finalHash[:], fhh)

	// Open meta-file.
	fn := fmt.Sprintf("%s.%x", coverage.MetaFilePref, finalHash)
	fpath := filepath.Join(outdir, fn)
	mf, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fatal("unable to open output meta-data file %s: %v", fpath, err)
	}

	mfw := encodemeta.NewCoverageMetaFileWriter(fpath, mf)
	err = mfw.Write(finalHash, m.modname, blobs)
	if err != nil {
		fatal("error writing %s: %v\n", fpath, err)
	}
	return finalHash
}

func (m *mstate) counterVisitor(f encodecounter.CounterVisitorFcn) bool {
	// For each package, for each function, construct counter
	// array and then call "f" on it.
	for pidx, p := range m.pkgs {
		fids := make([]int, len(p.ctab))
		for fid := range p.ctab {
			fids = append(fids, int(fid))
		}
		sort.Ints(fids)
		for _, fid := range fids {
			fp := p.ctab[uint32(fid)]
			if !f(uint32(pidx), uint32(fid), fp.Counters) {
				return false
			}
		}
	}
	return true
}

func (m *mstate) emitCounters(outdir string, metaHash [16]byte) {

	// Open output file.
	fn := fmt.Sprintf("%s.%x.%d", coverage.CounterFilePref, metaHash, time.Now().UnixNano())
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
	args := []string{}
	cfw := encodecounter.NewCoverageDataFileWriter(fpath, cf, args)

	visitor := func(f encodecounter.CounterVisitorFcn) bool {
		return m.counterVisitor(f)
	}
	if !cfw.Write(metaHash, visitor) {
		fatal("counter file write failed")
	}
}

func (m *mstate) Finish() {
	if *pcombineflag {
		finalHash := m.emitMeta(*outdirflag)
		m.emitCounters(*outdirflag, finalHash)
	}
}
