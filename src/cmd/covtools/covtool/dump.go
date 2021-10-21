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
var emitlegacyflag *string
var legacyoutf *os.File

func makeDumpOp() covOperation {
	emitlegacyflag = flag.String("emitlegacy", "", "Convert to legacy format and emit to selected file.")
	argsflag = flag.Bool("args", false, "Dump program args from counter data files")
	d := &dstate{}
	return d
}

type dstate struct {
	mm              map[pkfunc]decodecounter.FuncPayload
	preambleEmitted bool
	pkgName         string
	batchCounterAlloc
}

func (d *dstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: covtool dump -i=<directories>\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  covtool dump -i=dir\n\n")
	os.Exit(2)
}

func (d *dstate) Setup() {
	if *indirsflag == "" {
		usage("select input directories with '-i' option")
	}
	if *emitlegacyflag != "" {
		var err error
		legacyoutf, err = os.OpenFile(*emitlegacyflag, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		fmt.Fprintf(legacyoutf, "mode: set\n")
		if err != nil {
			d.Usage(fmt.Sprintf("unable to open legacy output file %q", *emitlegacyflag))
		}
	}
}

func (d *dstate) VisitIndir(idx uint32, indir string) {
}

func (d *dstate) BeginPod(p cov.Pod) {
	d.mm = make(map[pkfunc]decodecounter.FuncPayload)
}

func (d *dstate) EndPod(p cov.Pod) {
}

func (d *dstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader) {
	verb(2, "visit counter data file %s", cdf)
	if *argsflag {
		fmt.Printf("data file %s program args: %+v\n", cdf, cdr.Args())
	}
}

func (d *dstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}
	val := d.mm[key]

	if len(val.Counters) < len(data.Counters) {
		t := val.Counters
		val.Counters = d.allocateCounters(len(data.Counters))
		copy(val.Counters, t)
	}
	for i := 0; i < len(data.Counters); i++ {
		val.Counters[i] += data.Counters[i]
	}
	d.mm[key] = val
}

func (d *dstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	fmt.Printf("\nModule: %s\n", mfr.ModuleName())
}

func (d *dstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	d.preambleEmitted = false
	d.pkgName = pd.PackagePath()
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
	if haveCounters {
		counters = v.Counters
	} else if *liveflag {
		return
	}

	if legacyoutf != nil {
		emitLegacy(fd, counters)
	}
	if !d.preambleEmitted {
		fmt.Printf("\nPackage: %s\n", d.pkgName)
		d.preambleEmitted = true
	}
	fmt.Printf("Func: %s\n", fd.Funcname)
	fmt.Printf("Srcfile: %s\n", fd.Srcfile)
	for i := 0; i < len(fd.Units); i++ {
		u := fd.Units[i]
		var count uint32
		if counters != nil {
			count = counters[i]
		}
		fmt.Printf("%d: L%d:C%d -- L%d:C%d NS=%d = %d\n",
			i, u.StLine, u.StCol, u.EnLine, u.EnCol, u.NxStmts, count)
	}
}

func (d *dstate) Finish() {
	if legacyoutf != nil {
		if err := legacyoutf.Close(); err != nil {
			fatal("closing legacy output file: %v", err)
		}
	}
}
