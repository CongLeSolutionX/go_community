// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/link/internal/benchmark"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

var atExitFuncs []func()

func AtExit(f func()) {
	atExitFuncs = append(atExitFuncs, f)
}

// runAtExitFuncs runs the queued set of AtExit functions.
func runAtExitFuncs() {
	for i := len(atExitFuncs) - 1; i >= 0; i-- {
		atExitFuncs[i]()
	}
	atExitFuncs = nil
}

// Exit exits with code after executing all atExitFuncs.
func Exit(code int) {
	runAtExitFuncs()
	os.Exit(code)
}

// Exitf logs an error message then calls Exit(2).
func Exitf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, os.Args[0]+": "+format+"\n", a...)
	nerrors++
	Exit(2)
}

// afterErrorAction updates 'nerrors' on error and invokes exit or
// panics in the proper circumstances.
func afterErrorAction() {
	nerrors++
	if *flagH {
		panic("error")
	}
	if nerrors > 20 {
		Exitf("too many errors")
	}
}

// Errorf logs an error message.
//
// If more than 20 errors have been printed, exit with an error.
//
// Logging an error means that on exit cmd/link will delete any
// output file and return a non-zero error code.
func Errorf(s *sym.Symbol, format string, args ...interface{}) {
	if s != nil {
		format = s.Name + ": " + format
	}
	format += "\n"
	fmt.Fprintf(os.Stderr, format, args...)
	afterErrorAction()
}

// Errorf method logs an error message.
//
// If more than 20 errors have been printed, exit with an error.
//
// Logging an error means that on exit cmd/link will delete any
// output file and return a non-zero error code.
func (ctxt *Link) Errorf(s loader.Sym, format string, args ...interface{}) {
	if ctxt.loader != nil {
		ctxt.loader.Errorf(s, format, args)
		return
	}
	// Note: this is not expected to happen very often.
	format = fmt.Sprintf("sym %d: %s", s, format)
	format += "\n"
	fmt.Fprintf(os.Stderr, format, args...)
	afterErrorAction()
}

func artrim(x []byte) string {
	i := 0
	j := len(x)
	for i < len(x) && x[i] == ' ' {
		i++
	}
	for j > i && x[j-1] == ' ' {
		j--
	}
	return string(x[i:j])
}

func stringtouint32(x []uint32, s string) {
	for i := 0; len(s) > 0; i++ {
		var buf [4]byte
		s = s[copy(buf[:], s):]
		x[i] = binary.LittleEndian.Uint32(buf[:])
	}
}

// contains reports whether v is in s.
func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// implements sort.Interface, for sorting symbols by name.
type byName []*sym.Symbol

func (s byName) Len() int           { return len(s) }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byName) Less(i, j int) bool { return s[i].Name < s[j].Name }

// dumpSyms is a helper for the -dumpsymsat linker command line option.
type dumpSyms struct {
	phases   []string
	stypes   []string
	dorelocs []bool
	ctxt     *Link
	which    string
}

func (ds *dumpSyms) dumpBeforeLoadlibfull() {
	l := ds.ctxt.loader
	syms := []loader.Sym{}
	switch ds.which {
	case "all":
		for i := 1; i < l.NSym(); i++ {
			syms = append(syms, loader.Sym(i))
		}
	case "text":
		syms = ds.ctxt.Textp2
	}
	secti := make(map[*sym.Section]int)
	sections := []*sym.Section{nil}
	for _, s := range syms {
		sect := l.SymSect(s)
		if sect != nil {
			if secti[sect] == 0 {
				secti[sect] = len(sections)
				sections = append(sections, sect)
			}
		}
	}
	if len(secti) > 0 {
		fmt.Fprintf(os.Stderr, "Sections:\n")
		for i, sect := range sections[1:] {
			fmt.Fprintf(os.Stderr, " %d %s\n", i, sect.Name)
		}
	}
	for _, s := range syms {
		sn := l.SymName(s)
		st := l.SymType(s)
		doit, dorelocs := ds.includeSym(st)
		if !doit {
			continue
		}
		rch := " "
		if l.AttrReachable(s) {
			rch = "%"
		}
		sver := l.SymVersion(s)
		sval := l.SymValue(s)
		sz := l.SymSize(s)
		dlen := len(l.Data(s))
		fmt.Fprintf(os.Stderr, "%sS%d: %s<%d> t=%s sz=%d dlen=%d val=%d\n", rch, s, sn, sver, st.String(), sz, dlen, sval)
		if dorelocs {
			relocs := l.Relocs(s)
			for ri := 0; ri < relocs.Count(); ri++ {
				r := relocs.At2(ri)
				rsrs := "<nil>"
				if r.Sym() != 0 {
					rsrs = l.SymName(r.Sym())
				}
				fmt.Fprintf(os.Stderr, "  + R%d: %-9s o=%d a=%d tgt=%s\n", ri, r.Type().String(), r.Off(), r.Add(), rsrs)
			}
		}
	}
}

func (ds *dumpSyms) dumpAfterLoadlibfull() {
	secti := make(map[*sym.Section]int)
	sections := []*sym.Section{nil}
	si := make(map[*sym.Symbol]int)
	l := ds.ctxt.loader
	for i := 1; i < len(l.Syms); i++ {
		s := l.Syms[i]
		if s == nil {
			continue
		}
		si[s] = i
		if s.Sect != nil {
			if secti[s.Sect] == 0 {
				secti[s.Sect] = len(sections)
				sections = append(sections, s.Sect)
			}
		}
	}

	syms := ds.ctxt.Syms.Allsym
	switch ds.which {
	case "data":
		syms = ds.ctxt.datap
	case "text":
		syms = ds.ctxt.Textp
	}
	if len(secti) > 0 {
		fmt.Fprintf(os.Stderr, "Sections:\n")
		for i, sect := range sections[1:] {
			fmt.Fprintf(os.Stderr, " %d %s\n", i, sect.Name)
		}
	}
	for _, s := range syms {
		foundType := false
		dorelocs := false
		for i := range ds.stypes {
			st := ds.stypes[i]
			if st == s.Type.String() || st == "all" {
				foundType = true
				dorelocs = ds.dorelocs[i]
				break
			}
		}
		if !foundType {
			continue
		}
		rch := " "
		if s.Attr.Reachable() {
			rch = "%"
		}
		fmt.Fprintf(os.Stderr, "%sS%d: %s<%d> t=%s sz=%d len(s.P)=%d val=%d sect=%d\n", rch, si[s], s.Name, s.Version, s.Type.String(), s.Size, len(s.P), s.Value, secti[s.Sect])
		if dorelocs {
			for ri := 0; ri < len(s.R); ri++ {
				r := &s.R[ri]
				rt := r.Type
				rsrs := "<nil>"
				if r.Sym != nil {
					rsrs = r.Sym.String()
				}
				fmt.Fprintf(os.Stderr, "  + R%d: %-9s o=%d a=%d tgt=%s\n", ri, rt.String(), r.Off, r.Add, rsrs)
			}
		}
	}
}

func (ds *dumpSyms) includeSym(t sym.SymKind) (bool, bool) {
	for i := range ds.stypes {
		st := ds.stypes[i]
		if st == t.String() || st == "all" {
			return true, ds.dorelocs[i]
		}
	}
	return false, false
}

func (ds *dumpSyms) Start(which string) {
	foundPhase := false
	for _, p := range ds.phases {
		if which == p || p == "all" {
			foundPhase = true
			break
		}
	}
	if !foundPhase {
		return
	}
	ctxt := ds.ctxt
	fmt.Fprintf(os.Stderr, "\nSymbols dump before '%s':\n", which)
	if ctxt.loader == nil || len(ctxt.loader.Syms) == 0 {
		return
	}
	// HACK: peek at the context to detect whether we're dumping before
	// or after loadlibfull.
	if ctxt.lookup == nil {
		ds.dumpBeforeLoadlibfull()
	} else {
		ds.dumpAfterLoadlibfull()
	}
}

func dumpSymsError() {
	fmt.Fprintf(os.Stderr, "malformed argument to -dumpsymsat; should be of form '^?stype[,stype]*[|selector]@phase[,phase]*'\n")
	usage()
}

func makeDumpSyms(ctxt *Link) *dumpSyms {
	if *dumpSymsFlag != "" {
		// Vett the flag value. Flag argument is of the form
		//
		// ^?stype[,stype]*{selector}?@phase[,phase]*
		//
		// where 'stype' is either a symbol type (ex: STEXT) or a list
		// such as "data"/"text", the '^' is a flag requesting dump of
		// relocations, and 'phase' is a linker phase (ex: "pclntab",
		// or "dodata"). The 'selector' clause is optional, if present
		// it can be set to 'data' or 'text' to visit symbols on
		// ctxt.data or ctxt.Textp instead of all syms. The keyword 'all' can be
		// substituted for a symbol type or a phase name. Examples:
		//
		// Dump text symbols (with relocs) and NOPTRBSS symbols at dodata + final:
		//
		//    ^STEXT,SNOPTRBSS@dodata,final
		//
		// Dump all symbols at pclntab:
		//
		//    all@pclntab
		//
		// Dump all ctxt.datap symbols at address:
		//
		//    all{data}@address
		//
		// Dump SELFROSECT symbols at all phases:
		//
		//    SELFROSECT@all
		//
		phases := []string{}
		stypes := []string{}
		dorelocs := []bool{}
		selector := "all"
		halves := strings.Split(*dumpSymsFlag, "@")
		if len(halves) != 2 {
			dumpSymsError()
		}
		// symbol types
		stypesec := halves[0]
		switch {
		case strings.HasSuffix(stypesec, "{data}"):
			selector = "data"
			stypesec = stypesec[:len(stypesec)-len("{data}")]
		case strings.HasSuffix(stypesec, "{text}"):
			selector = "text"
			stypesec = stypesec[:len(stypesec)-len("{text}")]
		case strings.HasSuffix(stypesec, "{all}"):
			selector = "all"
			stypesec = stypesec[:len(stypesec)-len("{all}")]
		}
		left := strings.Split(stypesec, ",")
		if len(left) == 0 {
			dumpSymsError()
		}
		for _, c := range left {
			if len(c) == 0 {
				dumpSymsError()
			}
			typ := c
			rel := false
			if c[0] == '^' {
				typ = c[1:]
				rel = true
			}
			stypes = append(stypes, typ)
			dorelocs = append(dorelocs, rel)
		}
		// phases
		right := strings.Split(halves[1], ",")
		if len(right) == 0 {
			dumpSymsError()
		}
		for _, p := range right {
			if len(p) == 0 {
				dumpSymsError()
			}
			phases = append(phases, p)
		}
		return &dumpSyms{
			phases:   phases,
			dorelocs: dorelocs,
			stypes:   stypes,
			ctxt:     ctxt,
			which:    selector,
		}
	}
	return nil
}

func makeBenchMetrics() *benchmark.Metrics {
	if *benchmarkFlag != "" {
		// enable benchmarking
		if *benchmarkFlag == "mem" {
			return benchmark.New(benchmark.GC, *benchmarkFileFlag)
		} else if *benchmarkFlag == "cpu" {
			return benchmark.New(benchmark.NoGC, *benchmarkFileFlag)
		} else {
			Errorf(nil, "unknown benchmark flag: %q", *benchmarkFlag)
			usage()
		}
	}
	return nil
}

// atPhaseStart keeps track of things to do at the start of
// each linker phase (benchmarking, symbol dumping).
type atPhaseStart struct {
	bench   *benchmark.Metrics
	sdumper *dumpSyms
}

func makeAtPhaseStart(ctxt *Link) *atPhaseStart {
	return &atPhaseStart{
		bench:   makeBenchMetrics(),
		sdumper: makeDumpSyms(ctxt),
	}
}

func (aps *atPhaseStart) Start(which string) {
	if aps.bench != nil {
		aps.bench.Start(which)
	}
	if aps.sdumper != nil {
		aps.sdumper.Start(which)
	}
}

func (aps *atPhaseStart) Report(f *os.File) {
	if aps.bench != nil {
		aps.bench.Report(f)
	}
	if aps.sdumper != nil {
		aps.sdumper.Start("final")
	}
}
