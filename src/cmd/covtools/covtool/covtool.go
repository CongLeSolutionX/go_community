// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/internal/bio"
	"cmd/internal/cov"
	"flag"
	"fmt"
	"internal/coverage"
	"internal/coverage/decodecounter"
	"internal/coverage/decodemeta"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var hflag = flag.Bool("h", false, "Panic on fatal errors (for stack trace)")
var indirsflag = flag.String("i", "", "Input dirs to examine (comma separated)")
var liveflag = flag.Bool("live", false, "Select only live (executed) funcs.")
var pkgreflag = flag.String("pkg", "", "Restrict output to package matching this regular expression.")

var pkgre *regexp.Regexp

func verb(vlevel int, s string, a ...interface{}) {
	if *verbflag >= vlevel {
		fmt.Printf(s, a...)
		fmt.Printf("\n")
	}
}

func warn(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "warning: ")
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
	if *hflag {
		panic("unexpected warning")
	}
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: ")
	fmt.Fprintf(os.Stderr, s, a...)
	fmt.Fprintf(os.Stderr, "\n")
	if *hflag {
		panic("fatal error")
	}
	os.Exit(1)
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: covmerge -i=<directories> -o<dir>\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  covdump -i=dir1,dir2,dir3 -o=outdir\n\n")
	fmt.Fprintf(os.Stderr, "  \tmerges all files in dir1/dir2/dir3\n")
	fmt.Fprintf(os.Stderr, "  \tinto output dir outdir\n")
	os.Exit(2)
}

type covOperation interface {
	Setup()
	VisitIndir(idx uint32, indir string)
	BeginPod(p cov.Pod)
	EndPod(p cov.Pod)
	VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader)
	VisitFuncCounterData(payload decodecounter.FuncPayload)
	VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader)
	VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32)
	VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc)
	Finish()
	Usage(mst string)
}

func processPackage(mfname string, pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32, op covOperation) {
	if pkgre != nil {
		if !pkgre.Match([]byte(pd.PackagePath())) {
			return
		}
	}
	op.VisitPackage(pd, pkgIdx)
	nf := pd.NumFuncs()
	var fd coverage.FuncDesc
	for fidx := uint32(0); fidx < nf; fidx++ {
		if err := pd.ReadFunc(fidx, &fd); err != nil {
			fatal("reading meta-data file %s: %v", mfname, err)
		}
		op.VisitFunc(pkgIdx, fidx, &fd)
	}
}

// visitPod examines a coverage data 'pod', that is, a meta-data file and
// zero or more counter data files that refer to that meta-data file.
func visitPod(p cov.Pod, op covOperation) {
	verb(1, "visiting pod: metafile %s with %d counter files",
		p.MetaFile, len(p.CounterDataFiles))
	op.BeginPod(p)

	// Open meta-file
	f, err := os.Open(p.MetaFile)
	if err != nil {
		fatal("unable to open meta-file %s", p.MetaFile)
	}
	br := bio.NewReader(f)
	fi, err := f.Stat()
	if err != nil {
		fatal("unable to stat metafile %s: %v", p.MetaFile, err)
	}
	fileView := br.SliceRO(uint64(fi.Size()))
	br.MustSeek(0, os.SEEK_SET)

	verb(1, "fileView for pod is length %d", len(fileView))

	var mfr *decodemeta.CoverageMetaFileReader
	mfr, err = decodemeta.NewCoverageMetaFileReader(f, fileView)
	if err != nil {
		fatal("decoding meta-file %s: %s", p.MetaFile, err)
	}
	op.VisitMetaDataFile(p.MetaFile, mfr)

	// Read counter data files.
	for _, cdf := range p.CounterDataFiles {
		cf, err := os.Open(cdf)
		if err != nil {
			fatal("opening counter data file %s: %s", cdf, err)
		}
		var mr *cov.MReader
		mr, err = cov.NewMreader(cf)
		if err != nil {
			fatal("creating reader for counter data file %s: %s", cdf, err)
		}
		var cdr *decodecounter.CounterDataReader
		cdr, err = decodecounter.NewCounterDataReader(cdf, cf, mr)
		if err != nil {
			fatal("reading counter data file %s: %s", cdf, err)
		}
		op.VisitCounterDataFile(cdf, cdr)
		var data decodecounter.FuncPayload
		for {
			err := cdr.NextFunc(&data)
			if err == io.EOF {
				break
			} else if err != nil {
				fatal("reading counter data file %s: %v", cdf, err)
			}
			op.VisitFuncCounterData(data)
		}
	}

	// NB: packages in the meta-file will be in dependency order (basically
	// the order in which init files execute). Do we want an additional sort
	// pass here, say by packagepath?
	np := uint32(mfr.NumPackages())
	payload := []byte{}
	for pkIdx := uint32(0); pkIdx < np; pkIdx++ {
		var pd *decodemeta.CoverageMetaDataDecoder
		pd, payload, err = mfr.GetPackageDecoder(pkIdx, payload)
		if err != nil {
			fatal("reading pkg %d from meta-file %s: %s", pkIdx, p.MetaFile, err)
		}
		op.VisitPackage(pd, pkIdx)
		processPackage(p.MetaFile, pd, pkIdx, op)
	}

	op.EndPod(p)
}

type mdWriteSeeker struct {
	payload []byte
	off     int64
}

func (d *mdWriteSeeker) Write(p []byte) (n int, err error) {
	amt := len(p)
	towrite := d.payload[d.off:]
	if len(towrite) < amt {
		d.payload = append(d.payload, make([]byte, amt-len(towrite))...)
		towrite = d.payload[d.off:]
	}
	copy(towrite, p)
	d.off += int64(amt)
	return amt, nil
}

func (d *mdWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == os.SEEK_SET {
		d.off = offset
		return offset, nil
	} else if whence == os.SEEK_CUR {
		d.off += offset
		return d.off, nil
	}
	// other modes not supported
	panic("bad")
}

func (m *mstate) emit(outdir string) {
	finalHash := m.emitMeta(outdir)
	m.emitCounters(outdir, finalHash)
}

func perform(indirs []string, op covOperation) {
	pods, err := cov.CollectPods(indirs, false)
	if err != nil {
		fatal("reading inputs: %v", err)
	}
	if len(pods) == 0 {
		warn("no applicable files found in input directories")
	}
	for _, p := range pods {
		visitPod(p, op)
	}
	op.Finish()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("covtool: ")
	verb(1, "entering main")

	// First argument should be mode.
	if len(os.Args) < 2 {
		usage("missing mode selector")
	}

	// Select mode
	var operation covOperation
	switch os.Args[1] {
	case "merge":
		operation = makeMergeOp()
	case "dump":
		operation = makeDumpOp()
	case "intersect":
		panic("intersect not implemented yet")
	case "subtract":
		panic("subtract not implemented yet")
	default:
		usage("unknown mode selector")
	}

	// Edit out mode selector, then parse flags.
	os.Args = append(os.Args[:1], os.Args[2:]...)
	flag.Parse()

	// Mode-independent flag setup
	if flag.NArg() != 0 {
		usage("unknown extra arguments")
	}
	if *pkgreflag != "" {
		p, err := regexp.Compile(*pkgreflag)
		if err != nil {
			usage(fmt.Sprintf("unable to compile -pkg argument %q", *pkgreflag))
		}
		pkgre = p
	}

	// Mode setup.
	operation.Setup()

	// ... off and running now.
	verb(1, "starting perform")
	indirs := strings.Split(*indirsflag, ",")
	perform(indirs, operation)
	verb(1, "leaving main")
}
