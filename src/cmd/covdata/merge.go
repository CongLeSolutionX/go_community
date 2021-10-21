// Copyright 2022 The Go Authors. All rights reserved.
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
		pkm: make(map[string]*pkstate),
	}
	return m
}

type mstate struct {
	cov.BatchCounterAlloc
	counterMerge
	pkm     map[string]*pkstate // maps package path to package state
	pkgs    []*pkstate
	p       *pkstate
	pod     *podstate
	modname string
}

type podstate struct {
	mm       map[pkfunc]decodecounter.FuncPayload
	mdf      string
	fileHash [16]byte
}

type pkstate struct {
	pkgIdx uint32
	ctab   map[uint32]decodecounter.FuncPayload
	// these below are filled in only in -pcombine mode for merging.
	cmdb *encodemeta.CoverageMetaDataBuilder
	ftab map[[16]byte]uint32
}

type pkfunc struct {
	pk, fcn uint32
}

func (m *mstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata merge -i=<directories> -o=<dir>\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  go tool covdata merge -i=dir1,dir2,dir3 -o=outdir\n\n")
	fmt.Fprintf(os.Stderr, "  \tmerges all files in dir1/dir2/dir3\n")
	fmt.Fprintf(os.Stderr, "  \tinto output dir outdir\n")
	Exit(2)
}

func (m *mstate) Setup() {
	if *indirsflag == "" {
		m.Usage("select input directories with '-i' option")
	}
	if *outdirflag == "" {
		m.Usage("select output directory with '-o' option")
	}
}

func (m *mstate) BeginPod(p cov.Pod) {
	m.pod = &podstate{
		mm: make(map[pkfunc]decodecounter.FuncPayload),
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

func (m *mstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visit counter data file %s dirIdx %d", cdf, dirIdx)
}

func (m *mstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}
	val := m.pod.mm[key]
	// FIXME: in theory either A) len(val.Counters) is zero, or B)
	// the two lengths are equal. Assert if not? Of course, we could
	// see odd stuff if there is source file skew.
	if *verbflag > 4 {
		fmt.Printf("visit pk=%d fid=%d len(counters)=%d\n", data.PkgIdx, data.FuncIdx, len(data.Counters))
	}
	if len(val.Counters) < len(data.Counters) {
		t := val.Counters
		val.Counters = m.AllocateCounters(len(data.Counters))
		copy(val.Counters, t)
	}
	m.mergeCounters(val.Counters, data.Counters)
	m.pod.mm[key] = val
}

func (m *mstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	verb(2, "VisitMetaDataFile(mdf=%s)", mdf)
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
	// Keep track of counter mode.
	m.cmode = mfr.CounterMode()
}

func (m *mstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	verb(3, "VisitPackage(pk=%d path=%s)", pkgIdx, pd.PackagePath())
	// Install an entry for this package if needed.
	p, ok := m.pkm[pd.PackagePath()]
	if !ok {
		cmdb, err := encodemeta.NewCoverageMetaDataBuilder(pd.PackagePath(), pd.PackageClassification())
		if err != nil {
			fatal("fatal error creating meta-data builder: %v", err)
		}
		verb(2, "install new pkm entry for package %s pk=%d", pd.PackagePath(), pkgIdx)
		p = &pkstate{
			pkgIdx: uint32(len(m.pkgs)),
			cmdb:   cmdb,
			ftab:   make(map[[16]byte]uint32),
			ctab:   make(map[uint32]decodecounter.FuncPayload),
		}
		m.pkgs = append(m.pkgs, p)
		m.pkm[pd.PackagePath()] = p
	}
	m.p = p
}

func (m *mstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	var counters []uint32
	key := pkfunc{pk: pkgIdx, fcn: fnIdx}
	v, haveCounters := m.pod.mm[key]
	if haveCounters {
		counters = v.Counters
	}

	if *verbflag >= 3 {
		fmt.Printf("visit pk=%d fid=%d func %s\n", pkgIdx, fnIdx, fd.Funcname)
	}

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
			if *verbflag >= 3 {
				fmt.Printf("new meta entry for fn %s fid=%d\n", fd.Funcname, gfidx)
			}
		}
		fnIdx = gfidx
	}
	if !haveCounters {
		return
	}

	// Install counters in package ctab.
	gfp, ok := m.p.ctab[fnIdx]
	if ok {
		if *verbflag >= 3 {
			fmt.Printf("counter merge for %s fidx=%d\n", fd.Funcname, fnIdx)
		}
		// Merge.
		m.mergeCounters(gfp.Counters, counters)
		m.p.ctab[fnIdx] = gfp
	} else {
		if *verbflag >= 3 {
			fmt.Printf("null merge for %s fidx %d\n", fd.Funcname, fnIdx)
		}
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
	err = mfw.Write(finalHash, m.modname, blobs, m.cmode)
	if err != nil {
		fatal("error writing %s: %v\n", fpath, err)
	}
	return finalHash
}

func (m *mstate) NumFuncs() (int, error) {
	rval := 0
	for _, p := range m.pkgs {
		rval += len(p.ctab)
	}
	return rval, nil
}

func (m *mstate) VisitFuncs(f encodecounter.CounterVisitorFn) error {
	if *verbflag >= 4 {
		fmt.Printf("counterVisitor invoked\n")
	}
	// For each package, for each function, construct counter
	// array and then call "f" on it.
	for pidx, p := range m.pkgs {
		fids := make([]int, 0, len(p.ctab))
		for fid := range p.ctab {
			fids = append(fids, int(fid))
		}
		sort.Ints(fids)
		if *verbflag >= 4 {
			fmt.Printf("fids for pk=%d: %+v\n", pidx, fids)
		}
		for _, fid := range fids {
			fp := p.ctab[uint32(fid)]
			if *verbflag >= 4 {
				fmt.Printf("counter write for pk=%d fid=%d len(ctrs)=%d\n", pidx, fid, len(fp.Counters))
			}
			if err := f(uint32(pidx), uint32(fid), fp.Counters); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *mstate) emitCounters(outdir string, metaHash [16]byte) {

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
	if err := cfw.Write(metaHash, args, m); err != nil {
		fatal("counter file write failed: %v", err)
	}
}

func (m *mstate) Finish() {
	if *pcombineflag {
		finalHash := m.emitMeta(*outdirflag)
		m.emitCounters(*outdirflag, finalHash)
	}
}
