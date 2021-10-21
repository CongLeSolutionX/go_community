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
	"runtime"
	"runtime/pprof"
	"strings"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var hflag = flag.Bool("h", false, "Panic on fatal errors (for stack trace)")
var sumcountsflag = flag.Bool("sumcounts", false, "Sum coverage counts across input files.")
var whflag = flag.Bool("hw", false, "Panic on warnings (for stack trace)")
var indirsflag = flag.String("i", "", "Input dirs to examine (comma separated)")
var liveflag = flag.Bool("live", false, "Select only live (executed) funcs.")
var pkgreflag = flag.String("pkg", "", "Restrict output to package matching this regular expression.")
var cpuprofileflag = flag.String("cpuprofile", "", "Write CPU profile to specified file")
var memprofileflag = flag.String("memprofile", "", "Write CPU profile to specified file")
var memprofilerateflag = flag.Int("memprofilerate", 0, "Write CPU profile to specified file")

var pkgre *regexp.Regexp

var atExitFuncs []func()

func atExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

func Exit(code int) {
	for i := len(atExitFuncs) - 1; i >= 0; i-- {
		f := atExitFuncs[i]
		atExitFuncs = atExitFuncs[:i]
		f()
	}
	os.Exit(code)
}

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
	if *whflag {
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
	Exit(1)
}

func usage(msg string) {
	if len(msg) > 0 {
		fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	}
	fmt.Fprintf(os.Stderr, "usage: go tool covdata [command]\n\n")
	fmt.Fprintf(os.Stderr, "Commands are:\n\n")
	fmt.Fprintf(os.Stderr, "dump        dump coverage data in text format\n")
	fmt.Fprintf(os.Stderr, "merge       merge data files together\n")
	fmt.Fprintf(os.Stderr, "subtract    subtract one set of data files from another set\n")
	fmt.Fprintf(os.Stderr, "intersect   generate intersection of two sets of data files\n")
	Exit(2)
}

type covOperation interface {
	Setup()
	BeginPod(p cov.Pod)
	EndPod(p cov.Pod)
	VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int)
	VisitFuncCounterData(payload decodecounter.FuncPayload)
	VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader)
	VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32)
	VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc)
	Finish()
	Usage(mst string)
}

type batchCounterAlloc struct {
	pool []uint32
}

func (ca *batchCounterAlloc) allocateCounters(n int) []uint32 {
	const chunk = 8192
	if n > cap(ca.pool) {
		siz := chunk
		if n > chunk {
			siz = n
		}
		ca.pool = make([]uint32, siz)
	}
	rv := ca.pool[:n]
	ca.pool = ca.pool[n:]
	return rv
}

func processPackage(mfname string, pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32, op covOperation) {
	op.VisitPackage(pd, pkgIdx)
	if pkgre != nil {
		if !pkgre.Match([]byte(pd.PackagePath())) {
			return
		}
	}
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
	for k, cdf := range p.CounterDataFiles {
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
		op.VisitCounterDataFile(cdf, cdr, p.Origins[k])
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
		processPackage(p.MetaFile, pd, pkIdx, op)
	}

	op.EndPod(p)
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
	log.SetPrefix("covdata: ")

	// First argument should be mode/subcommand.
	if len(os.Args) < 2 {
		usage("missing command selector")
	}

	// Select mode
	var op covOperation
	cmd := os.Args[1]
	switch cmd {
	case "merge":
		op = makeMergeOp()
	case "dump":
		op = makeDumpOp()
	case "subtract":
		op = makeSubtractIntersectOp("subtract")
	case "intersect":
		op = makeSubtractIntersectOp("intersect")
	default:
		usage(fmt.Sprintf("unknown command selector %q", cmd))
	}

	// Edit out command selector, then parse flags.
	os.Args = append(os.Args[:1], os.Args[2:]...)
	flag.Usage = func() {
		op.Usage("")
	}
	flag.Parse()

	// Mode-independent flag setup
	verb(1, "starting mode-independent setup")
	if flag.NArg() != 0 {
		op.Usage("unknown extra arguments")
	}
	if *pkgreflag != "" {
		p, err := regexp.Compile(*pkgreflag)
		if err != nil {
			op.Usage(fmt.Sprintf("unable to compile -pkg argument %q", *pkgreflag))
		}
		pkgre = p
	}
	if *cpuprofileflag != "" {
		f, err := os.Create(*cpuprofileflag)
		if err != nil {
			fatal("%v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fatal("%v", err)
		}
		atExit(pprof.StopCPUProfile)
	}
	if *memprofileflag != "" {
		if *memprofilerateflag != 0 {
			runtime.MemProfileRate = *memprofilerateflag
		}
		f, err := os.Create(*memprofileflag)
		if err != nil {
			fatal("%v", err)
		}
		atExit(func() {
			runtime.GC()
			const writeLegacyFormat = 1
			if err := pprof.Lookup("heap").WriteTo(f, writeLegacyFormat); err != nil {
				fatal("%v", err)
			}
		})
	} else {
		// Not doing memory profiling; disable it entirely.
		runtime.MemProfileRate = 0
	}

	// Mode-dependent setup.
	op.Setup()

	// ... off and running now.
	verb(1, "starting perform")
	indirs := strings.Split(*indirsflag, ",")
	perform(indirs, op)
	verb(1, "leaving main")
	Exit(0)
}
