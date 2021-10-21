// Copyright 2021 The Go Authors. All rights reserved.
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
)

var argsflag *bool
var emitpercflag *bool
var emitdumpflag *bool
var emitlegacyflag *string
var legacyoutf *os.File
var totalStmts, coveredStmts int

func makeDumpOp() covOperation {
	emitpercflag = flag.Bool("emitperc", false, "Emit percentage of covered statements file.")
	emitdumpflag = flag.Bool("emitdump", false, "Dump out all data for debugging purposes.")
	emitlegacyflag = flag.String("emitlegacy", "", "Convert to legacy format and emit to selected file.")
	argsflag = flag.Bool("args", false, "Dump program args from counter data files")
	d := &dstate{}
	return d
}

type dstate struct {
	mm                map[pkfunc]decodecounter.FuncPayload
	pkm               map[uint32]uint32
	legacyModeEmitted bool
	preambleEmitted   bool
	pkgName           string
	pkgClass          string
	cmode             coverage.CounterMode
	batchCounterAlloc
}

func (d *dstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata dump -i=<directories>\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  go tool covdata dump -i=dir\n\n")
	Exit(2)
}

func (d *dstate) Setup() {
	if *indirsflag == "" {
		d.Usage("select input directories with '-i' option")
	}
	if *emitlegacyflag != "" {
		var err error
		legacyoutf, err = os.OpenFile(*emitlegacyflag, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			d.Usage(fmt.Sprintf("unable to open legacy output file %q", *emitlegacyflag))
		}
	}
}

func (d *dstate) BeginPod(p cov.Pod) {
	d.mm = make(map[pkfunc]decodecounter.FuncPayload)
}

func (d *dstate) EndPod(p cov.Pod) {
}

func (d *dstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visit counter data file %s dirIdx %d", cdf, dirIdx)
	if *argsflag {
		fmt.Printf("data file %s program args: %+v\n", cdf, cdr.OsArgs())
	}
}

func (d *dstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	if nf, ok := d.pkm[data.PkgIdx]; !ok || data.FuncIdx > nf {
		warn("func payload insanity: id [p=%d,f=%d] nf=%d len(ctrs)=%d in VisitFuncCounterData, ignored", data.PkgIdx, data.FuncIdx, nf, len(data.Counters))
		return
	}
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}
	val, found := d.mm[key]

	verb(5, "ctr visit pk=%d fid=%d found=%v len(val.ctrs)=%d len(data.ctrs)=%d", data.PkgIdx, data.FuncIdx, found, len(val.Counters), len(data.Counters))

	if len(val.Counters) < len(data.Counters) {
		t := val.Counters
		val.Counters = d.allocateCounters(len(data.Counters))
		copy(val.Counters, t)
	}
	for i := 0; i < len(data.Counters); i++ {
		if *sumcountsflag {
			val.Counters[i] += data.Counters[i]
		} else {
			if data.Counters[i] != 0 {
				val.Counters[i] = 1
			}
		}
	}
	d.mm[key] = val
}

func (d *dstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	if *emitdumpflag {
		fmt.Printf("\nModule: %s\n", mfr.ModuleName())
	}

	d.cmode = mfr.CounterMode()
	if !d.legacyModeEmitted {
		d.legacyModeEmitted = true
		fmt.Fprintf(legacyoutf, "mode: %s\n", d.cmode.String())
	}

	// To provide an additional layer of sanity checking when reading
	// counter data, walk the meta-data file to determine the set of
	// legal package/function combinations. This will help catch bugs
	// in the counter file reader.
	d.pkm = make(map[uint32]uint32)
	np := uint32(mfr.NumPackages())
	payload := []byte{}
	for pkIdx := uint32(0); pkIdx < np; pkIdx++ {
		var pd *decodemeta.CoverageMetaDataDecoder
		var err error
		pd, payload, err = mfr.GetPackageDecoder(pkIdx, payload)
		if err != nil {
			fatal("reading pkg %d from meta-file %s: %s", pkIdx, mdf, err)
		}
		d.pkm[pkIdx] = pd.NumFuncs()
	}
}

func (d *dstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	d.preambleEmitted = false
	d.pkgName = pd.PackagePath()
	d.pkgClass = pd.PackageClassification().String()
}

func emitLegacy(fd *coverage.FuncDesc, counters []uint32) {
	// Legacy format looks like:
	// <srcfile>:STLINE.STCOL,ENLINE.ENCOL NSTMTS EXE
	for i := 0; i < len(fd.Units); i++ {
		u := fd.Units[i]
		var count uint32
		if counters != nil {
			count = counters[i]
		}
		fmt.Fprintf(legacyoutf, "%s:%d.%d,%d.%d %d %d\n",
			fd.Srcfile, u.StLine, u.StCol,
			u.EnLine, u.EnCol, u.NxStmts, count)
	}
}

func (d *dstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	var counters []uint32
	key := pkfunc{pk: pkgIdx, fcn: fnIdx}
	v, haveCounters := d.mm[key]

	verb(5, "meta visit pk=%d fid=%d fname=%s file=%s found=%v len(val.ctrs)=%d", pkgIdx, fnIdx, fd.Funcname, fd.Srcfile, haveCounters, len(v.Counters))

	if haveCounters {
		counters = v.Counters
	} else if *liveflag {
		return
	}

	if legacyoutf != nil {
		emitLegacy(fd, counters)
	}
	if *emitdumpflag {
		if !d.preambleEmitted {
			fmt.Printf("\nPackage: %s\n", d.pkgName)
			fmt.Printf("\nClassification: %s\n", d.pkgClass)
			d.preambleEmitted = true
		}
		fmt.Printf("\nFunc: %s\n", fd.Funcname)
		fmt.Printf("Srcfile: %s\n", fd.Srcfile)
	}
	for i := 0; i < len(fd.Units); i++ {
		u := fd.Units[i]
		var count uint32
		if counters != nil {
			count = counters[i]
		}
		if *emitdumpflag {
			fmt.Printf("%d: L%d:C%d -- L%d:C%d NS=%d = %d\n",
				i, u.StLine, u.StCol, u.EnLine, u.EnCol, u.NxStmts, count)
		}
		totalStmts += int(u.NxStmts)
		if count != 0 {
			coveredStmts += int(u.NxStmts)
		}
	}
}

func (d *dstate) Finish() {
	if legacyoutf != nil {
		if err := legacyoutf.Close(); err != nil {
			fatal("closing legacy output file: %v", err)
		}
	}
	if *emitpercflag {
		if totalStmts == 0 {
			fmt.Println("coverage: [no statements]")
		} else {
			fmt.Printf("coverage: %.1f%% of statements\n", 100*float64(coveredStmts)/float64(totalStmts))
		}
	}
	if *emitdumpflag {
		fmt.Printf("totalStmts: %d coveredStmts: %d\n", totalStmts, coveredStmts)
	}
}
