// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"internal/coverage"
	"internal/coverage/calloc"
	"internal/coverage/cformat"
	"internal/coverage/cmerge"
	"internal/coverage/decodecounter"
	"internal/coverage/decodemeta"
	"internal/coverage/pods"
	"os"
	"sort"
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
	d := &dstate{
		cmd:    cmd,
		cm:     &cmerge.Merger{},
		format: cformat.NewFormatter(),
	}
	if d.cmd == "pkglist" {
		d.pkgpaths = make(map[string]struct{})
	}
	return d
}

type dstate struct {
	calloc.BatchCounterAlloc
	cm                       *cmerge.Merger
	mm                       map[pkfunc]decodecounter.FuncPayload
	pkm                      map[uint32]uint32
	pkgpaths                 map[string]struct{}
	format                   *cformat.Formatter
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
	case "pkglist":
		fmt.Fprintf(os.Stderr, "  go tool covdata pkglist -i=dir1,dir2\n\n")
		fmt.Fprintf(os.Stderr, "  \treads coverage data files from dir1+dirs2\n")
		fmt.Fprintf(os.Stderr, "  \tand writes out a list of the import paths\n")
		fmt.Fprintf(os.Stderr, "  \tof all compiled packages.\n")
	case "textfmt":
		fmt.Fprintf(os.Stderr, "  go tool covdata textfmt -i=dir1,dir2 -o=out.txt\n\n")
		fmt.Fprintf(os.Stderr, "  \tmerges data from input directories dir1+dir2\n")
		fmt.Fprintf(os.Stderr, "  \tand emits text format into file 'out.txt'\n")
	case "percent":
		fmt.Fprintf(os.Stderr, "  go tool covdata percent -i=dir1,dir2\n\n")
		fmt.Fprintf(os.Stderr, "  \tmerges data from input directories dir1+dir2\n")
		fmt.Fprintf(os.Stderr, "  \tand emits percentage of statements covered\n\n")
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
		args := append([]string{os.Args[0]}, "debugdump")
		args = append(args, os.Args[1:]...)
		fmt.Printf(" *\t%s\n", strings.Join(args, " "))
		fmt.Printf(" */\n")
	}
}

func (d *dstate) BeginPod(p pods.Pod) {
	d.mm = make(map[pkfunc]decodecounter.FuncPayload)
}

func (d *dstate) EndPod(p pods.Pod) {
	if d.cmd == "debugdump" {
		d.cm.ResetModeAndGranularity()
	}
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
	err, overflow := d.cm.MergeCounters(val.Counters, data.Counters)
	if err != nil {
		fatal("%v", err)
	}
	if overflow {
		warn("uint32 overflow during counter merge")
	}
	d.mm[key] = val
}

func (d *dstate) EndCounters() {
}

func (d *dstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	if d.cmd == "debugdump" {
		fmt.Printf("\nModule: %s\n", mfr.ModuleName())
	}
	newgran := mfr.CounterGranularity()
	newmode := mfr.CounterMode()
	if err := d.cm.SetModeAndGranularity(mdf, newmode, newgran); err != nil {
		fatal("%v", err)
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
	if d.cmd == "pkglist" {
		d.pkgpaths[d.pkgName] = struct{}{}
	}
	d.format.SetPackage(pd.PackagePath())
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
		d.format.AddUnit(fd.Srcfile, u, count)
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
	if d.cmd == "percent" {
		d.format.EmitPercent(os.Stdout, "")
	}
	if d.textfmtoutf != nil {
		if err := d.format.EmitTextual(d.textfmtoutf, d.cm.Mode()); err != nil {
			fatal("writing to %s: %v", *textfmtoutflag, err)
		}
		if err := d.textfmtoutf.Close(); err != nil {
			fatal("closing textfmt output file %s: %v", *textfmtoutflag, err)
		}
	}
	if d.cmd == "debugdump" {
		fmt.Printf("totalStmts: %d coveredStmts: %d\n", d.totalStmts, d.coveredStmts)
	}
	if d.cmd == "pkglist" {
		pkgs := make([]string, 0, len(d.pkgpaths))
		for p := range d.pkgpaths {
			pkgs = append(pkgs, p)
		}
		sort.Strings(pkgs)
		for _, p := range pkgs {
			fmt.Printf("%s\n", p)
		}
	}
}
