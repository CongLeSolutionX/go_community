// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/goobj"
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
//
// TODO: remove. Use ctxt.Errorf instead.
// All remaining calls use nil as first arg.
func Errorf(dummy *int, format string, args ...interface{}) {
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
		ctxt.loader.Errorf(s, format, args...)
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

// dumpSyms is a helper for the -dumpsymsat linker command line option.
type dumpSyms struct {
	phases    []string
	stypes    []string
	dorelocs  bool
	dofrelocs bool
	dosub     bool
	doaux     bool
	dopkg     bool
	ctxt      *Link
	which     string
}

type auxinfo struct {
	typ uint8
	par loader.Sym
}

// genAuxMap returns a table that maps aux sym by index to
// parent symidx and aux type.
func (ds *dumpSyms) genAuxMap() map[loader.Sym]auxinfo {
	am := make(map[loader.Sym]auxinfo)
	ldr := ds.ctxt.loader
	for s := loader.Sym(1); s < loader.Sym(ldr.NSym()); s++ {
		naux := ldr.NAux(s)
		for i := 0; i < naux; i++ {
			a := ldr.Aux(s, i)
			if a.Sym() == 0 {
				continue
			}
			am[a.Sym()] = auxinfo{typ: a.Type(), par: s}
		}
	}
	return am
}

func (ds *dumpSyms) auxTag(s loader.Sym, am map[loader.Sym]auxinfo) string {
	if !ds.doaux {
		return ""
	}
	if ai, ok := am[s]; ok {
		ts := ""
		switch ai.typ {
		case goobj.AuxGotype:
			ts = "AuxGoType"
		case goobj.AuxFuncInfo:
			ts = "AuxFuncInfo"
		case goobj.AuxFuncdata:
			ts = "AuxFuncData"
		case goobj.AuxDwarfInfo:
			ts = "AuxDwarfInfo"
		case goobj.AuxDwarfLoc:
			ts = "AuxDwarfLoc"
		case goobj.AuxDwarfRanges:
			ts = "AuxDwarfRanges"
		case goobj.AuxDwarfLines:
			ts = "AuxDwarfLines"
		}
		return fmt.Sprintf(" %s for %d", ts, ai.par)
	}
	return ""
}

func (ds *dumpSyms) dump(syms []loader.Sym) {
	l := ds.ctxt.loader
	am := ds.genAuxMap()
	if len(syms) == 0 {
		switch ds.which {
		case "all":
			for i := 1; i < l.NSym(); i++ {
				syms = append(syms, loader.Sym(i))
			}
		case "data":
			syms = ds.ctxt.datap
		case "text":
			syms = ds.ctxt.Textp
		}
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
			fmt.Fprintf(os.Stderr, " %d %s\n", i+1, sect.Name)
		}
	}
	mp := make(map[string]int)
	if ds.dopkg {
		pkgs := []string{""}
		mp[""] = 0
		for _, s := range syms {
			pkg := l.SymPkg(s)
			if _, ok := mp[pkg]; !ok {
				mp[pkg] = len(pkgs)
				pkgs = append(pkgs, pkg)
			}
		}
		fmt.Fprintf(os.Stderr, "Packages:\n")
		for i, pkg := range pkgs {
			fmt.Fprintf(os.Stderr, " %d %s\n", i, pkg)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}
	for _, s := range syms {
		sn := l.SymName(s)
		st := l.SymType(s)
		doit := ds.includeSym(st)
		if !doit {
			continue
		}
		rch := " "
		if l.AttrReachable(s) {
			rch = "%"
		}
		sct := ""
		if l.SymSect(s) != nil {
			sct = fmt.Sprintf(" sect=%d", secti[l.SymSect(s)])
		}
		subinfo := ""
		if ds.dosub {
			if l.SubSym(s) != 0 {
				subinfo += fmt.Sprintf(" sub=%d", l.SubSym(s))
			}
			if l.OuterSym(s) != 0 {
				subinfo += fmt.Sprintf(" outer=%d", l.OuterSym(s))
			}
		}
		pkinfo := ""
		if ds.dopkg {
			p := l.SymPkg(s)
			pkinfo = fmt.Sprintf(" PK=%d", mp[p])
		}
		sver := l.SymVersion(s)
		sval := l.SymValue(s)
		sz := l.SymSize(s)
		dlen := len(l.Data(s))
		fmt.Fprintf(os.Stderr, "%sS%d: %s<%d> t=%s sz=%d dlen=%d val=%d%s%s%s%s\n", rch, s, sn, sver, st.String(), sz, dlen, sval, sct, ds.auxTag(s, am), subinfo, pkinfo)
		if ds.dorelocs {
			relocs := l.Relocs(s)
			for ri := 0; ri < relocs.Count(); ri++ {
				r := relocs.At(ri)
				rsrs := "<nil>"
				rvs := ""
				if r.Sym() != 0 {
					rsrs = l.SymName(r.Sym())
					rvs = fmt.Sprintf("<%d>", l.SymVersion(r.Sym()))
				}
				rvr := ""
				if l.RelocVariant(s, ri) != 0 {
					rvr = fmt.Sprintf(" RV=%d", l.RelocVariant(s, ri))
				}
				frel := ""
				if ds.dofrelocs {
					frel = fmt.Sprintf("[S%d]", r.Sym())
				}
				fmt.Fprintf(os.Stderr, "  + R%d: %-9s o=%d a=%d tgt=%s%s%s%s\n", ri, r.Type().String(), r.Off(), r.Add(), rsrs, rvs, frel, rvr)
			}
		}
	}
}

func (ds *dumpSyms) includeSym(t sym.SymKind) bool {
	for i := range ds.stypes {
		st := ds.stypes[i]
		if st == t.String() || st == "all" {
			return true
		}
	}
	return false
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
	if ctxt.loader == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "\nDump %s symbols before '%s':\n", ds.which, which)
	ds.dump(nil)
}

const usageString = `
Argument to -dumpsymsat should be of the form

{aux}?{symtypes}{qualifier}?@{phases}

where

 {aux} is a clause of the form [^&%...] where

  ^ => enables dumping of relocations
  % => enables dumping of outer/sub relationships
  & => enables dumping of aux relationships
  # => enables dumping of relocations with tgt symbol names and indices
  $ => enables dumping of Pkg for each sym

 {symtypes} is a clause of the form symtype{,symtype}*

 with symtype specifying a sym.Kind, e.g. STEXT, SDATA, or
 the pseudo-type "all" indicate all symbol types

 {qualifier} is a clause of the form

 "text" indicating that only the contents of ctxt.Textp should be dumped
 "data" indicating that only the contents of ctxt.datap should be dumped

 {phases} is a clause of the form phase{,phase}*

 with phase specifying a linker phase (ex: "deadcode", "linksetup", "symtab")
 of the pseudo-phase "all" to dump at all phases.

 Examples:

   - dump text symbols (with relocs) and NOPTRBSS symbols at dodata + final:

       [^]STEXT,SNOPTRBSS@dodata,final

   - dump all symbols at pclntab:

	   all@pclntab

   - dump all ctxt.datap symbols at address:

	   all{data}@address

   - dump SELFROSECT symbols at all phases:

	   SELFROSECT@all

   - dump all symbols at deadcode with aux info and outer/sub info:

       [&%]all@deadcode
`

func dumpSymsError(tag string) {
	panic(fmt.Sprintf("malformed argument to -dumpsymsat: %s\n%s", tag, usageString))
}

func makeDumpSyms(ctxt *Link) *dumpSyms {
	if *dumpSymsFlag != "" {
		// Vett the flag value. See usage above for more on structure
		// of the argument.
		phases := []string{}
		stypes := []string{}
		rel := false
		frel := false
		sub := false
		pkg := false
		daux := false
		selector := "all"
		f := *dumpSymsFlag
		if f[0] == '[' {
			i := strings.Index(f, "]")
			if i == -1 {
				dumpSymsError("expected ']' to close ']")
			}
			for k := 1; k < i; k++ {
				c := f[k]
				if c == '^' {
					rel = true
				} else if c == '#' {
					frel = true
					rel = true
				} else if c == '%' {
					sub = true
				} else if c == '$' {
					pkg = true
				} else if c == '&' {
					daux = true
				} else {
					dumpSymsError(fmt.Sprintf("bad char %c in [] clause '%s'", c, f[1:i]))
				}
			}
			f = f[i+1:]
		}
		halves := strings.Split(f, "@")
		if len(halves) != 2 {
			dumpSymsError("expected exactly one '@' character")
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
			dumpSymsError("empty section types clause")
		}
		for _, c := range left {
			if len(c) == 0 {
				dumpSymsError("zero length type in section types clause")
			}
			typ := c
			stypes = append(stypes, typ)
		}
		// phases
		right := strings.Split(halves[1], ",")
		if len(right) == 0 {
			dumpSymsError("empty phases clause")
		}
		for _, p := range right {
			if len(p) == 0 {
				dumpSymsError("zero length phase in phases clause")
			}
			phases = append(phases, p)
		}
		return &dumpSyms{
			phases:    phases,
			dorelocs:  rel,
			dofrelocs: frel,
			dosub:     sub,
			dopkg:     pkg,
			doaux:     daux,
			stypes:    stypes,
			ctxt:      ctxt,
			which:     selector,
		}
	}
	return nil
}

// loaderDeltaInspector helps developers figure out what additions
// took place in the loader (e.g. symbols added) over a specific window
// of time, for example, an operation like reading a host object file.
// Typical usage is as follows:
//
//	ldi := mkLoaderDeltaInspector(ctxt)
//	<some operation that creates symbols>
//	ldi.Dump()
type loaderDeltaInspector struct {
	ds dumpSyms
	m  map[loader.Sym]struct{}
}

func mkLoaderDeltaInspector(ctxt *Link) *loaderDeltaInspector {
	m := make(map[loader.Sym]struct{})
	for i := 1; i < ctxt.loader.NSym(); i++ {
		m[loader.Sym(i)] = struct{}{}
	}
	return &loaderDeltaInspector{
		ds: dumpSyms{
			dorelocs:  true,
			dofrelocs: true,
			dopkg:     true,
			which:     "all",
			ctxt:      ctxt,
			stypes:    []string{"all"},
		},
		m: m,
	}
}

func (ldi *loaderDeltaInspector) Dump(tag string) {
	syms := []loader.Sym{}
	for i := 1; i < ldi.ds.ctxt.loader.NSym(); i++ {
		s := loader.Sym(i)
		if _, ok := ldi.m[s]; ok {
			continue
		}
		syms = append(syms, s)
	}
	olds := len(ldi.m)
	news := ldi.ds.ctxt.loader.NSym()
	ldi.ds.ctxt.Logf("%s: old=%d new=%d ns=%d:\n",
		tag, olds, news, len(syms))
	ldi.ds.dump(syms)
	ldi.ds.ctxt.Logf("end %s\n", tag)
}

// dumpLoaderSyms dumps all symbols and relocations; it is intended
// for use in situations where we need a symbol dump within a phase
// (as opposed to at the start of a phase).
func dumpLoaderSyms(ctxt *Link, tag string) {
	ctxt.Logf("begin loader symbol dump at %q:\n", tag)
	ds := &dumpSyms{
		dorelocs:  true,
		dofrelocs: true,
		dopkg:     true,
		which:     "all",
		ctxt:      ctxt,
		stypes:    []string{"all"},
	}
	ds.dump(nil)
	ctxt.Logf("end loader symbol dump at %q:\n", tag)
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
