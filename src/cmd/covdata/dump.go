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

var textfmtoutflag *string
var liveflag *bool

func makeDumpOp(cmd string) covOperation {
	if cmd == "textfmt" || cmd == "percent" {
		textfmtoutflag = flag.String("o", "", "Output text format to file")
	}
	if cmd == "debugdump" {
		liveflag = flag.Bool("live", false, "Select only live (executed) functions for dump output.")
	}
	d := &dstate{cmd: cmd}
	return d
}

type dstate struct {
	cov.BatchCounterAlloc
	counterMerge
	mm                       map[pkfunc]decodecounter.FuncPayload
	pkm                      map[uint32]uint32
	textfmtModeEmitted       bool
	preambleEmitted          bool
	pkgName                  string
	pkgClass                 string
	cmd                      string
	textfmtoutf              *os.File
	totalStmts, coveredStmts int
}

func (d *dstate) Usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata %s -i=<directories>\n\n", d.cmd)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	switch d.cmd {
	case "textfmt":
		fmt.Fprintf(os.Stderr, "  go tool covdata textfmt -i=dir1,dir2 -o=out.txt\n\n")
		fmt.Fprintf(os.Stderr, "  \tmerges data from input directories dir1+dirs2\n")
		fmt.Fprintf(os.Stderr, "  \tand emits text format into file 'out.txt'\n")
	case "percent":
		fmt.Fprintf(os.Stderr, "  go tool covdata percent -i=dir1,dir2\n\n")
		fmt.Fprintf(os.Stderr, "  \tmerges data from input directories dir1+dirs2\n")
		fmt.Fprintf(os.Stderr, "  \tand emits percentage of statements covered\n\n")
		fmt.Fprintf(os.Stderr, "  go tool covdata percent -pkgre='^main$' -i=dir1\n\n")
		fmt.Fprintf(os.Stderr, "  \tcomputes and displays percentage of statements\n")
		fmt.Fprintf(os.Stderr, "  \tcovered for package 'main' from data in dir 'dir1'.\n")
	case "debugdump":
		fmt.Fprintf(os.Stderr, "  go tool covdata debugdump [flags] -i=dir1,dir2\n\n")
		fmt.Fprintf(os.Stderr, "  \treads coverage data from dir1+dir2 and dumps\n")
		fmt.Fprintf(os.Stderr, "  \tcontents in human-readable form to stdout, for\n")
		fmt.Fprintf(os.Stderr, "  \tdebugging purposes.\n")
	default:
		panic("unexpected")
	}
	describeClassification()
	Exit(2)
}

func (d *dstate) Setup() {
	if *indirsflag == "" {
		d.Usage("select input directories with '-i' option")
	}
	if d.cmd == "textfmt" || (d.cmd == "percent" && *textfmtoutflag != "") {
		if *textfmtoutflag == "" {
			d.Usage("select output file name with '-o' option")
		}
		var err error
		d.textfmtoutf, err = os.Create(*textfmtoutflag)
		if err != nil {
			d.Usage(fmt.Sprintf("unable to open textfmt output file %q: %v", *textfmtoutflag, err))
		}
	}
	if d.cmd == "debugdump" {
		fmt.Printf("/* WARNING: the format of this dump is not stable and is\n")
		fmt.Printf(" * expected to change from one Go release to the next.\n")
		fmt.Printf(" *\n")
		fmt.Printf(" * produced by:\n")
		fmt.Printf(" *\t%s\n", strings.Join(os.Args, " "))
		fmt.Printf(" */\n")
	}
}

func (d *dstate) BeginPod(p cov.Pod) {
	d.mm = make(map[pkfunc]decodecounter.FuncPayload)
}

func (d *dstate) EndPod(p cov.Pod) {
}

func (d *dstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visit counter data file %s dirIdx %d", cdf, dirIdx)
	if d.cmd == "debugdump" {
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
		val.Counters = d.AllocateCounters(len(data.Counters))
		copy(val.Counters, t)
	}
	d.mergeCounters(val.Counters, data.Counters)
	d.mm[key] = val
}

func (d *dstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	if d.cmd == "debugdump" {
		fmt.Printf("\nModule: %s\n", mfr.ModuleName())
	}
	d.cmode = mfr.CounterMode()
	if d.textfmtoutf != nil && !d.textfmtModeEmitted {
		d.textfmtModeEmitted = true
		fmt.Fprintf(d.textfmtoutf, "mode: %s\n", d.cmode.String())
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

func (d *dstate) emitLegacy(fd *coverage.FuncDesc, counters []uint32) {
	// Legacy text format looks like:
	// <srcfile>:STLINE.STCOL,ENLINE.ENCOL NSTMTS EXE
	for i := 0; i < len(fd.Units); i++ {
		u := fd.Units[i]
		// Skip units with non-zero parent (no way to represent
		// these in the existing format).
		if u.Parent != 0 {
			continue
		}
		var count uint32
		if counters != nil {
			count = counters[i]
		}
		fmt.Fprintf(d.textfmtoutf, "%s:%d.%d,%d.%d %d %d\n",
			fd.Srcfile, u.StLine, u.StCol,
			u.EnLine, u.EnCol, u.NxStmts, count)
	}
}

func (d *dstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	var counters []uint32
	key := pkfunc{pk: pkgIdx, fcn: fnIdx}
	v, haveCounters := d.mm[key]

	verb(5, "meta visit pk=%d fid=%d fname=%s file=%s found=%v len(val.ctrs)=%d", pkgIdx, fnIdx, fd.Funcname, fd.Srcfile, haveCounters, len(v.Counters))

	suppressOutput := false
	if haveCounters {
		counters = v.Counters
	} else if d.cmd == "debugdump" && *liveflag {
		suppressOutput = true
	}

	if d.textfmtoutf != nil {
		d.emitLegacy(fd, counters)
	}
	if d.cmd == "debugdump" && !suppressOutput {
		if !d.preambleEmitted {
			fmt.Printf("\nPackage path: %s\n", d.pkgName)
			fmt.Printf("Package classification: %s\n", d.pkgClass)
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
		if d.cmd == "debugdump" && !suppressOutput {
			fmt.Printf("%d: L%d:C%d -- L%d:C%d ",
				i, u.StLine, u.StCol, u.EnLine, u.EnCol)
			if u.Parent != 0 {
				fmt.Printf("Parent:%d = %d\n", u.Parent, count)
			} else {
				fmt.Printf("NS=%d = %d\n", u.NxStmts, count)
			}
		}
		d.totalStmts += int(u.NxStmts)
		if count != 0 {
			d.coveredStmts += int(u.NxStmts)
		}
	}
}

func (d *dstate) Finish() {
	if d.textfmtoutf != nil {
		if err := d.textfmtoutf.Close(); err != nil {
			fatal("closing textfmt output file: %v", err)
		}
	}
	if d.cmd == "percent" {
		if d.totalStmts == 0 {
			fmt.Println("coverage: [no statements]")
		} else {
			fmt.Printf("coverage: %.1f%% of statements\n", 100*float64(d.coveredStmts)/float64(d.totalStmts))
		}
	}
	if d.cmd == "debugdump" {
		fmt.Printf("totalStmts: %d coveredStmts: %d\n", d.totalStmts, d.coveredStmts)
	}
}
