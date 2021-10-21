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
	"internal/coverage/decodemeta"
	"io"
	"log"
	"os"
	"strings"
)

var verbflag = flag.Int("v", 0, "Verbose trace output level")
var hflag = flag.Bool("h", false, "Panic on fatal errors (for stack trace)")
var dirflag = flag.String("dir", "", "Input directory to examine")
var filesflag = flag.String("files", "", "Files to examine (comma separated)")

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
	fmt.Fprintf(os.Stderr, "usage: covdump [flags]\n\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n\n")
	fmt.Fprintf(os.Stderr, "  covdump -dir somedir\n\n")
	fmt.Fprintf(os.Stderr, "  \tdumps all files in somedir\n\n")
	fmt.Fprintf(os.Stderr, "  covdump -files X,Y,Z\n\n")
	fmt.Fprintf(os.Stderr, "  \tdumps only files X,Y,Z\n")
	os.Exit(2)
}

type pkfunc struct {
	pk, fcn uint32
}

func processPackage(mfname string, pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32, mm map[pkfunc]cov.FuncPayload) {
	fmt.Printf("\nPackage: %s\n", pd.PackagePath())
	nf := pd.NumFuncs()
	var fd coverage.FuncDesc
	for fidx := uint32(0); fidx < nf; fidx++ {
		if err := pd.ReadFunc(fidx, &fd); err != nil {
			fatal("reading meta-data file %s: %v", mfname, err)
		}
		var counters []uint32
		key := pkfunc{pk: pkgIdx, fcn: fidx}
		if v, ok := mm[key]; ok {
			counters = v.Counters
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
}

func visitPod(p cov.Pod) {
	verb(1, "visiting pod: metafile %s with %d counter files",
		p.MetaFile, len(p.CounterDataFiles))
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

	const chunk = 8192
	var pool []uint32
	batchAlloc := func(n int) []uint32 {
		if n > cap(pool) {
			if n < chunk {
				n = chunk
			}
			pool = make([]uint32, n)
		}
		rv := pool[:n]
		pool = pool[n:]
		return rv
	}

	// Read all counter data files here, effectively merging their
	// counter values.
	mm := make(map[pkfunc]cov.FuncPayload)
	for _, cdf := range p.CounterDataFiles {
		cdr, err := cov.NewCounterDataReader(cdf)
		if err != nil {
			fatal("reading counter data file %s: %s", cdf, err)
		}
		var data cov.FuncPayload
		for {
			err := cdr.NextFunc(&data)
			if err == io.EOF {
				break
			} else if err != nil {
				fatal("reading counter data file %s: %v", cdf, err)
			}
			key := pkfunc{pk: data.PkgIdx, fcn: data.FuncIdx}
			val := mm[key]
			if cap(val.Counters) < len(data.Counters) {
				val.Counters = batchAlloc(len(data.Counters))
			}
			for i := 0; i < len(data.Counters); i++ {
				val.Counters[i] += data.Counters[i]
			}
			mm[key] = val
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
		processPackage(p.MetaFile, pd, pkIdx, mm)
	}
}

func performdir(dirpath string) {
	pods, err := cov.CollectPods(dirpath, true)
	if err != nil {
		fatal("unable to read directory %s: %v", dirpath, err)
	}
	for _, p := range pods {
		visitPod(p)
	}
}

func performfiles(files []string) {
	pods := cov.CollectPodsFromFiles(files, true)
	for _, p := range pods {
		visitPod(p)
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("covdump: ")
	flag.Parse()
	verb(1, "entering main")
	if *dirflag == "" && *filesflag == "" {
		usage("select input directory or files with -dir or -files")
	}
	if *dirflag != "" && *filesflag != "" {
		usage("select only one of -files / -dir")
	}
	verb(1, "in main verblevel=%d", *verbflag)
	if flag.NArg() != 0 {
		usage("unknown extra arguments")
	}
	if *dirflag != "" {
		performdir(*dirflag)
	} else {
		files := strings.Split(*filesflag, ",")
		performfiles(files)
	}

	verb(1, "leaving main")
}
