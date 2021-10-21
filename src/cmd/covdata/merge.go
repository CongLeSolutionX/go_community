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
)

var outdirflag *string
var pcombineflag *bool

func makeMergeOp() covOperation {
	outdirflag = flag.String("o", "", "Output directory to write")
	pcombineflag = flag.Bool("pcombine", false, "Combine profiles derived from distinct program executables")
	m := &mstate{}
	m.initMetaMerge()
	return m
}

type mstate struct {
	metaMerge
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
	m.metaBeginPod()
}

func (m *mstate) EndPod(p cov.Pod) {
	m.metaEndPod(*pcombineflag)
}

func (m *mstate) VisitCounterDataFile(cdf string, cdr *decodecounter.CounterDataReader, dirIdx int) {
	verb(2, "visit counter data file %s dirIdx %d", cdf, dirIdx)
}

func (m *mstate) VisitFuncCounterData(data decodecounter.FuncPayload) {
	m.metaVisitFuncCounterData(data)
}

func (m *mstate) EndCounters() {
}

func (m *mstate) VisitMetaDataFile(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	m.metaVisitMetaDataFile(mdf, mfr)
}

func (m *mstate) VisitPackage(pd *decodemeta.CoverageMetaDataDecoder, pkgIdx uint32) {
	verb(3, "VisitPackage(pk=%d path=%s)", pkgIdx, pd.PackagePath())

	m.metaVisitPackage(pd, pkgIdx, *pcombineflag)
}

func (m *mstate) VisitFunc(pkgIdx uint32, fnIdx uint32, fd *coverage.FuncDesc) {
	m.metaVisitFunc(pkgIdx, fnIdx, fd, "merge", *pcombineflag)
}

func (m *mstate) Finish() {
	if *pcombineflag {
		finalHash := m.emitMeta(*outdirflag, true)
		m.emitCounters(*outdirflag, finalHash)
	}
}
