// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/gcprog"
	"cmd/internal/objabi"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

// Temporary dumping around for sym.Symbol version of helper
// functions in dodata(), still being used for some archs/oses.
// FIXME: get rid of this file when dodata() is completely
// converted.

func (ctxt *Link) dodata() {
	// Give zeros sized symbols space if necessary.
	fixZeroSizedSymbols(ctxt)

	// Collect data symbols by type into data.
	state := dodataState{ctxt: ctxt}
	for _, s := range ctxt.loader.Syms {
		if s == nil {
			continue
		}
		if !s.Attr.Reachable() || s.Attr.Special() || s.Attr.SubSymbol() {
			continue
		}
		if s.Type <= sym.STEXT || s.Type >= sym.SXREF {
			continue
		}
		state.data[s.Type] = append(state.data[s.Type], s)
	}

	// Now that we have the data symbols, but before we start
	// to assign addresses, record all the necessary
	// dynamic relocations. These will grow the relocation
	// symbol, which is itself data.
	//
	// On darwin, we need the symbol table numbers for dynreloc.
	if ctxt.HeadType == objabi.Hdarwin {
		panic("not supported")
		//machosymorder(ctxt)
	}
	state.dynreloc(ctxt)

	// Move any RO data with relocations to a separate section.
	state.makeRelroForSharedLib(ctxt)

	// Temporary for debugging.
	symToIdx := make(map[*sym.Symbol]loader.Sym)
	for s := loader.Sym(1); s < loader.Sym(ctxt.loader.NSym()); s++ {
		sp := ctxt.loader.Syms[s]
		if sp != nil {
			symToIdx[sp] = s
		}
	}

	// Sort symbols.
	var wg sync.WaitGroup
	for symn := range state.data {
		symn := sym.SymKind(symn)
		wg.Add(1)
		go func() {
			state.data[symn], state.dataMaxAlign[symn] = dodataSect(ctxt, symn, state.data[symn], symToIdx)
			wg.Done()
		}()
	}
	wg.Wait()

	if ctxt.HeadType == objabi.Haix && ctxt.LinkMode == LinkExternal {
		// These symbols must have the same alignment as their section.
		// Otherwize, ld might change the layout of Go sections.
		ctxt.Syms.ROLookup("runtime.data", 0).Align = state.dataMaxAlign[sym.SDATA]
		ctxt.Syms.ROLookup("runtime.bss", 0).Align = state.dataMaxAlign[sym.SBSS]
	}

	// Create *sym.Section objects and assign symbols to sections for
	// data/rodata (and related) symbols.
	state.allocateDataSections(ctxt)

	// Create *sym.Section objects and assign symbols to sections for
	// DWARF symbols.
	state.allocateDwarfSections(ctxt)

	/* number the sections */
	n := int16(1)

	for _, sect := range Segtext.Sections {
		sect.Extnum = n
		n++
	}
	for _, sect := range Segrodata.Sections {
		sect.Extnum = n
		n++
	}
	for _, sect := range Segrelrodata.Sections {
		sect.Extnum = n
		n++
	}
	for _, sect := range Segdata.Sections {
		sect.Extnum = n
		n++
	}
	for _, sect := range Segdwarf.Sections {
		sect.Extnum = n
		n++
	}
}

// makeRelroForSharedLib creates a section of readonly data if necessary.
func (state *dodataState) makeRelroForSharedLib(target *Link) {
	if !target.UseRelro() {
		return
	}

	// "read only" data with relocations needs to go in its own section
	// when building a shared library. We do this by boosting objects of
	// type SXXX with relocations to type SXXXRELRO.
	for _, symnro := range sym.ReadOnly {
		symnrelro := sym.RelROMap[symnro]

		ro := []*sym.Symbol{}
		relro := state.data[symnrelro]

		for _, s := range state.data[symnro] {
			isRelro := len(s.R) > 0
			switch s.Type {
			case sym.STYPE, sym.STYPERELRO, sym.SGOFUNCRELRO:
				// Symbols are not sorted yet, so it is possible
				// that an Outer symbol has been changed to a
				// relro Type before it reaches here.
				isRelro = true
			case sym.SFUNCTAB:
				if target.IsAIX() && s.Name == "runtime.etypes" {
					// runtime.etypes must be at the end of
					// the relro datas.
					isRelro = true
				}
			}
			if isRelro {
				s.Type = symnrelro
				if s.Outer != nil {
					s.Outer.Type = s.Type
				}
				relro = append(relro, s)
			} else {
				ro = append(ro, s)
			}
		}

		// Check that we haven't made two symbols with the same .Outer into
		// different types (because references two symbols with non-nil Outer
		// become references to the outer symbol + offset it's vital that the
		// symbol and the outer end up in the same section).
		for _, s := range relro {
			if s.Outer != nil && s.Outer.Type != s.Type {
				Errorf(s, "inconsistent types for symbol and its Outer %s (%v != %v)",
					s.Outer.Name, s.Type, s.Outer.Type)
			}
		}

		state.data[symnro] = ro
		state.data[symnrelro] = relro
	}
}

func dynrelocsym(ctxt *Link, s *sym.Symbol) {
	target := &ctxt.Target
	ldr := ctxt.loader
	syms := &ctxt.ArchSyms
	for ri := range s.R {
		r := &s.R[ri]
		if ctxt.BuildMode == BuildModePIE && ctxt.LinkMode == LinkInternal {
			// It's expected that some relocations will be done
			// later by relocsym (R_TLS_LE, R_ADDROFF), so
			// don't worry if Adddynrel returns false.
			thearch.Adddynrel(target, ldr, syms, s, r)
			continue
		}

		if r.Sym != nil && r.Sym.Type == sym.SDYNIMPORT || r.Type >= objabi.ElfRelocOffset {
			if r.Sym != nil && !r.Sym.Attr.Reachable() {
				Errorf(s, "dynamic relocation to unreachable symbol %s", r.Sym.Name)
			}
			if !thearch.Adddynrel(target, ldr, syms, s, r) {
				Errorf(s, "unsupported dynamic relocation for symbol %s (type=%d (%s) stype=%d (%s))", r.Sym.Name, r.Type, sym.RelocName(ctxt.Arch, r.Type), r.Sym.Type, r.Sym.Type)
			}
		}
	}
}

func (state *dodataState) dynreloc(ctxt *Link) {
	if ctxt.HeadType == objabi.Hwindows {
		return
	}
	// -d suppresses dynamic loader format, so we may as well not
	// compute these sections or mark their symbols as reachable.
	if *FlagD {
		return
	}

	for _, s := range ctxt.Textp {
		dynrelocsym(ctxt, s)
	}
	for _, syms := range state.data {
		for _, s := range syms {
			dynrelocsym(ctxt, s)
		}
	}
	if ctxt.IsELF {
		elfdynhash(ctxt)
	}
}

func Addstring(s *sym.Symbol, str string) int64 {
	if s.Type == 0 {
		s.Type = sym.SNOPTRDATA
	}
	s.Attr |= sym.AttrReachable
	r := s.Size
	if s.Name == ".shstrtab" {
		elfsetstring(s, str, int(r))
	}
	s.P = append(s.P, str...)
	s.P = append(s.P, 0)
	s.Size = int64(len(s.P))
	return r
}

// symalign returns the required alignment for the given symbol s.
func symalign(s *sym.Symbol) int32 {
	min := int32(thearch.Minalign)
	if s.Align >= min {
		return s.Align
	} else if s.Align != 0 {
		return min
	}
	if strings.HasPrefix(s.Name, "go.string.") || strings.HasPrefix(s.Name, "type..namedata.") {
		// String data is just bytes.
		// If we align it, we waste a lot of space to padding.
		return min
	}
	align := int32(thearch.Maxalign)
	for int64(align) > s.Size && align > min {
		align >>= 1
	}
	s.Align = align
	return align
}

func aligndatsize(datsize int64, s *sym.Symbol) int64 {
	return Rnd(datsize, int64(symalign(s)))
}

type GCProg struct {
	ctxt *Link
	sym  *sym.Symbol
	w    gcprog.Writer
}

func (p *GCProg) Init(ctxt *Link, name string) {
	p.ctxt = ctxt
	p.sym = ctxt.Syms.Lookup(name, 0)
	p.w.Init(p.writeByte(ctxt))
	if debugGCProg {
		fmt.Fprintf(os.Stderr, "ld: start GCProg %s\n", name)
		p.w.Debug(os.Stderr)
	}
}

func (p *GCProg) writeByte(ctxt *Link) func(x byte) {
	return func(x byte) {
		p.sym.AddUint8(x)
	}
}

func (p *GCProg) End(size int64) {
	p.w.ZeroUntil(size / int64(p.ctxt.Arch.PtrSize))
	p.w.End()
	if debugGCProg {
		fmt.Fprintf(os.Stderr, "ld: end GCProg\n")
	}
}

func (p *GCProg) AddSym(s *sym.Symbol) {
	typ := s.Gotype
	// Things without pointers should be in sym.SNOPTRDATA or sym.SNOPTRBSS;
	// everything we see should have pointers and should therefore have a type.
	if typ == nil {
		switch s.Name {
		case "runtime.data", "runtime.edata", "runtime.bss", "runtime.ebss":
			// Ignore special symbols that are sometimes laid out
			// as real symbols. See comment about dyld on darwin in
			// the address function.
			return
		}
		Errorf(s, "missing Go type information for global symbol: size %d", s.Size)
		return
	}

	ptrsize := int64(p.ctxt.Arch.PtrSize)
	nptr := decodetypePtrdata(p.ctxt.Arch, typ.P) / ptrsize

	if debugGCProg {
		fmt.Fprintf(os.Stderr, "gcprog sym: %s at %d (ptr=%d+%d)\n", s.Name, s.Value, s.Value/ptrsize, nptr)
	}

	if decodetypeUsegcprog(p.ctxt.Arch, typ.P) == 0 {
		// Copy pointers from mask into program.
		mask := decodetypeGcmask(p.ctxt, typ)
		for i := int64(0); i < nptr; i++ {
			if (mask[i/8]>>uint(i%8))&1 != 0 {
				p.w.Ptr(s.Value/ptrsize + i)
			}
		}
		return
	}

	// Copy program.
	prog := decodetypeGcprog(p.ctxt, typ)
	p.w.ZeroUntil(s.Value / ptrsize)
	p.w.Append(prog[4:], nptr)
}

// dataSortKey is used to sort a slice of data symbol *sym.Symbol pointers.
// The sort keys are kept inline to improve cache behavior while sorting.
type dataSortKey struct {
	size   int64
	name   string
	sym    *sym.Symbol
	symIdx loader.Sym
}

type bySizeAndName []dataSortKey

func (d bySizeAndName) Len() int      { return len(d) }
func (d bySizeAndName) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d bySizeAndName) Less(i, j int) bool {
	s1, s2 := d[i], d[j]
	if s1.size != s2.size {
		return s1.size < s2.size
	}
	if s1.name != s2.name {
		return s1.name < s2.name
	}
	return s1.symIdx < s2.symIdx
}

// fixZeroSizedSymbols gives a few special symbols with zero size some space.
func fixZeroSizedSymbols(ctxt *Link) {
	// The values in moduledata are filled out by relocations
	// pointing to the addresses of these special symbols.
	// Typically these symbols have no size and are not laid
	// out with their matching section.
	//
	// However on darwin, dyld will find the special symbol
	// in the first loaded module, even though it is local.
	//
	// (An hypothesis, formed without looking in the dyld sources:
	// these special symbols have no size, so their address
	// matches a real symbol. The dynamic linker assumes we
	// want the normal symbol with the same address and finds
	// it in the other module.)
	//
	// To work around this we lay out the symbls whose
	// addresses are vital for multi-module programs to work
	// as normal symbols, and give them a little size.
	//
	// On AIX, as all DATA sections are merged together, ld might not put
	// these symbols at the beginning of their respective section if there
	// aren't real symbols, their alignment might not match the
	// first symbol alignment. Therefore, there are explicitly put at the
	// beginning of their section with the same alignment.
	if !(ctxt.DynlinkingGo() && ctxt.HeadType == objabi.Hdarwin) && !(ctxt.HeadType == objabi.Haix && ctxt.LinkMode == LinkExternal) {
		return
	}

	bss := ctxt.Syms.Lookup("runtime.bss", 0)
	bss.Size = 8
	bss.Attr.Set(sym.AttrSpecial, false)

	ctxt.Syms.Lookup("runtime.ebss", 0).Attr.Set(sym.AttrSpecial, false)

	data := ctxt.Syms.Lookup("runtime.data", 0)
	data.Size = 8
	data.Attr.Set(sym.AttrSpecial, false)

	edata := ctxt.Syms.Lookup("runtime.edata", 0)
	edata.Attr.Set(sym.AttrSpecial, false)
	if ctxt.HeadType == objabi.Haix {
		// XCOFFTOC symbols are part of .data section.
		edata.Type = sym.SXCOFFTOC
	}

	types := ctxt.Syms.Lookup("runtime.types", 0)
	types.Type = sym.STYPE
	types.Size = 8
	types.Attr.Set(sym.AttrSpecial, false)

	etypes := ctxt.Syms.Lookup("runtime.etypes", 0)
	etypes.Type = sym.SFUNCTAB
	etypes.Attr.Set(sym.AttrSpecial, false)

	if ctxt.HeadType == objabi.Haix {
		rodata := ctxt.Syms.Lookup("runtime.rodata", 0)
		rodata.Type = sym.SSTRING
		rodata.Size = 8
		rodata.Attr.Set(sym.AttrSpecial, false)

		ctxt.Syms.Lookup("runtime.erodata", 0).Attr.Set(sym.AttrSpecial, false)
	}
}

// allocateDataSectionForSym creates a new sym.Section into which a a
// single symbol will be placed. Here "seg" is the segment into which
// the section will go, "s" is the symbol to be placed into the new
// section, and "rwx" contains permissions for the section.
func (state *dodataState) allocateDataSectionForSym(seg *sym.Segment, s *sym.Symbol, rwx int) *sym.Section {
	sect := addsection(state.ctxt.loader, state.ctxt.Arch, seg, s.Name, rwx)
	sect.Align = symalign(s)
	state.datsize = Rnd(state.datsize, int64(sect.Align))
	sect.Vaddr = uint64(state.datsize)
	return sect
}

// assignDsymsToSection assigns a collection of data symbols to a
// newly created section. "sect" is the section into which to place
// the symbols, "syms" holds the list of symbols to assign,
// "forceType" (if non-zero) contains a new sym type to apply to each
// sym during the assignment, and "aligner" is a hook to call to
// handle alignment during the assignment process.
func (state *dodataState) assignDsymsToSection(sect *sym.Section, syms []*sym.Symbol, forceType sym.SymKind, aligner func(datsize int64, s *sym.Symbol) int64) {
	for _, s := range syms {
		state.datsize = aligner(state.datsize, s)
		s.Sect = sect
		if forceType != sym.Sxxx {
			s.Type = forceType
		}
		s.Value = int64(uint64(state.datsize) - sect.Vaddr)
		state.datsize += s.Size
	}
	sect.Length = uint64(state.datsize) - sect.Vaddr
}

func (state *dodataState) assignToSection(sect *sym.Section, symn sym.SymKind, forceType sym.SymKind) {
	state.assignDsymsToSection(sect, state.data[symn], forceType, aligndatsize)
	state.checkdatsize(symn)
}

// allocateSingleSymSections walks through the bucketed data symbols
// with type 'symn', creates a new section for each sym, and assigns
// the sym to a newly created section. Section name is set from the
// symbol name. "Seg" is the segment into which to place the new
// section, "forceType" is the new sym.SymKind to assign to the symbol
// within the section, and "rwx" holds section permissions.
func (state *dodataState) allocateSingleSymSections(seg *sym.Segment, symn sym.SymKind, forceType sym.SymKind, rwx int) {
	for _, s := range state.data[symn] {
		sect := state.allocateDataSectionForSym(seg, s, rwx)
		s.Sect = sect
		s.Type = forceType
		s.Value = int64(uint64(state.datsize) - sect.Vaddr)
		state.datsize += s.Size
		sect.Length = uint64(state.datsize) - sect.Vaddr
	}
	state.checkdatsize(symn)
}

// allocateNamedSectionAndAssignSyms creates a new section with the
// specified name, then walks through the bucketed data symbols with
// type 'symn' and assigns each of them to this new section. "Seg" is
// the segment into which to place the new section, "secName" is the
// name to give to the new section, "forceType" (if non-zero) contains
// a new sym type to apply to each sym during the assignment, and
// "rwx" holds section permissions.
func (state *dodataState) allocateNamedSectionAndAssignSyms(seg *sym.Segment, secName string, symn sym.SymKind, forceType sym.SymKind, rwx int) *sym.Section {

	sect := state.allocateNamedDataSection(seg, secName, []sym.SymKind{symn}, rwx)
	state.assignDsymsToSection(sect, state.data[symn], forceType, aligndatsize)
	return sect
}

// allocateDataSections allocates sym.Section objects for data/rodata
// (and related) symbols, and then assigns symbols to those sections.
func (state *dodataState) allocateDataSections(ctxt *Link) {
	// Allocate sections.
	// Data is processed before segtext, because we need
	// to see all symbols in the .data and .bss sections in order
	// to generate garbage collection information.

	// Writable data sections that do not need any specialized handling.
	writable := []sym.SymKind{
		sym.SBUILDINFO,
		sym.SELFSECT,
		sym.SMACHO,
		sym.SMACHOGOT,
		sym.SWINDOWS,
	}
	for _, symn := range writable {
		state.allocateSingleSymSections(&Segdata, symn, sym.SDATA, 06)
	}

	// .got (and .toc on ppc64)
	if len(state.data[sym.SELFGOT]) > 0 {
		sect := state.allocateNamedSectionAndAssignSyms(&Segdata, ".got", sym.SELFGOT, sym.SDATA, 06)
		if ctxt.IsPPC64() {
			for _, s := range state.data[sym.SELFGOT] {
				// Resolve .TOC. symbol for this object file (ppc64)
				toc := ctxt.Syms.ROLookup(".TOC.", int(s.Version))
				if toc != nil {
					toc.Sect = sect
					toc.Outer = s
					toc.Sub = s.Sub
					s.Sub = toc

					toc.Value = 0x8000
				}
			}
		}
	}

	/* pointer-free data */
	sect := state.allocateNamedSectionAndAssignSyms(&Segdata, ".noptrdata", sym.SNOPTRDATA, sym.SDATA, 06)
	ctxt.Syms.Lookup("runtime.noptrdata", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.enoptrdata", 0).Sect = sect

	hasinitarr := ctxt.linkShared

	/* shared library initializer */
	switch ctxt.BuildMode {
	case BuildModeCArchive, BuildModeCShared, BuildModeShared, BuildModePlugin:
		hasinitarr = true
	}

	if ctxt.HeadType == objabi.Haix {
		if len(state.data[sym.SINITARR]) > 0 {
			Errorf(nil, "XCOFF format doesn't allow .init_array section")
		}
	}

	if hasinitarr && len(state.data[sym.SINITARR]) > 0 {
		state.allocateNamedSectionAndAssignSyms(&Segdata, ".init_array", sym.SINITARR, sym.Sxxx, 06)
	}

	/* data */
	sect = state.allocateNamedSectionAndAssignSyms(&Segdata, ".data", sym.SDATA, sym.SDATA, 06)
	ctxt.Syms.Lookup("runtime.data", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.edata", 0).Sect = sect
	dataGcEnd := state.datsize - int64(sect.Vaddr)

	// On AIX, TOC entries must be the last of .data
	// These aren't part of gc as they won't change during the runtime.
	state.assignToSection(sect, sym.SXCOFFTOC, sym.SDATA)
	state.checkdatsize(sym.SDATA)
	sect.Length = uint64(state.datsize) - sect.Vaddr

	/* bss */
	sect = state.allocateNamedSectionAndAssignSyms(&Segdata, ".bss", sym.SBSS, sym.Sxxx, 06)
	ctxt.Syms.Lookup("runtime.bss", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.ebss", 0).Sect = sect
	bssGcEnd := state.datsize - int64(sect.Vaddr)

	// Emit gcdata for bcc symbols now that symbol values have been assigned.
	gcsToEmit := []struct {
		symName string
		symKind sym.SymKind
		gcEnd   int64
	}{
		{"runtime.gcdata", sym.SDATA, dataGcEnd},
		{"runtime.gcbss", sym.SBSS, bssGcEnd},
	}
	for _, g := range gcsToEmit {
		var gc GCProg
		gc.Init(ctxt, g.symName)
		for _, s := range state.data[g.symKind] {
			gc.AddSym(s)
		}
		gc.End(g.gcEnd)
	}

	/* pointer-free bss */
	sect = state.allocateNamedSectionAndAssignSyms(&Segdata, ".noptrbss", sym.SNOPTRBSS, sym.Sxxx, 06)
	ctxt.Syms.Lookup("runtime.noptrbss", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.enoptrbss", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.end", 0).Sect = sect

	// Coverage instrumentation counters for libfuzzer.
	if len(state.data[sym.SLIBFUZZER_EXTRA_COUNTER]) > 0 {
		state.allocateNamedSectionAndAssignSyms(&Segdata, "__libfuzzer_extra_counters", sym.SLIBFUZZER_EXTRA_COUNTER, sym.Sxxx, 06)
	}

	if len(state.data[sym.STLSBSS]) > 0 {
		var sect *sym.Section
		// FIXME: not clear why it is sometimes necessary to suppress .tbss section creation.
		if (ctxt.IsELF || ctxt.HeadType == objabi.Haix) && (ctxt.LinkMode == LinkExternal || !*FlagD) {
			sect = addsection(ctxt.loader, ctxt.Arch, &Segdata, ".tbss", 06)
			sect.Align = int32(ctxt.Arch.PtrSize)
			// FIXME: why does this need to be set to zero?
			sect.Vaddr = 0
		}
		state.datsize = 0

		for _, s := range state.data[sym.STLSBSS] {
			state.datsize = aligndatsize(state.datsize, s)
			s.Sect = sect
			s.Value = state.datsize
			state.datsize += s.Size
		}
		state.checkdatsize(sym.STLSBSS)

		if sect != nil {
			sect.Length = uint64(state.datsize)
		}
	}

	/*
	 * We finished data, begin read-only data.
	 * Not all systems support a separate read-only non-executable data section.
	 * ELF and Windows PE systems do.
	 * OS X and Plan 9 do not.
	 * And if we're using external linking mode, the point is moot,
	 * since it's not our decision; that code expects the sections in
	 * segtext.
	 */
	var segro *sym.Segment
	if ctxt.IsELF && ctxt.LinkMode == LinkInternal {
		segro = &Segrodata
	} else if ctxt.HeadType == objabi.Hwindows {
		segro = &Segrodata
	} else {
		segro = &Segtext
	}

	state.datsize = 0

	/* read-only executable ELF, Mach-O sections */
	if len(state.data[sym.STEXT]) != 0 {
		Errorf(nil, "dodata found an sym.STEXT symbol: %s", state.data[sym.STEXT][0].Name)
	}
	state.allocateSingleSymSections(&Segtext, sym.SELFRXSECT, sym.SRODATA, 04)

	/* read-only data */
	sect = state.allocateNamedDataSection(segro, ".rodata", sym.ReadOnly, 04)
	ctxt.Syms.Lookup("runtime.rodata", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.erodata", 0).Sect = sect
	if !ctxt.UseRelro() {
		ctxt.Syms.Lookup("runtime.types", 0).Sect = sect
		ctxt.Syms.Lookup("runtime.etypes", 0).Sect = sect
	}
	for _, symn := range sym.ReadOnly {
		symnStartValue := state.datsize
		state.assignToSection(sect, symn, sym.SRODATA)
		if ctxt.HeadType == objabi.Haix {
			// Read-only symbols might be wrapped inside their outer
			// symbol.
			// XCOFF symbol table needs to know the size of
			// these outer symbols.
			xcoffUpdateOuterSize(ctxt, state.datsize-symnStartValue, symn)
		}
	}

	/* read-only ELF, Mach-O sections */
	state.allocateSingleSymSections(segro, sym.SELFROSECT, sym.SRODATA, 04)
	state.allocateSingleSymSections(segro, sym.SMACHOPLT, sym.SRODATA, 04)

	// There is some data that are conceptually read-only but are written to by
	// relocations. On GNU systems, we can arrange for the dynamic linker to
	// mprotect sections after relocations are applied by giving them write
	// permissions in the object file and calling them ".data.rel.ro.FOO". We
	// divide the .rodata section between actual .rodata and .data.rel.ro.rodata,
	// but for the other sections that this applies to, we just write a read-only
	// .FOO section or a read-write .data.rel.ro.FOO section depending on the
	// situation.
	// TODO(mwhudson): It would make sense to do this more widely, but it makes
	// the system linker segfault on darwin.
	const relroPerm = 06
	const fallbackPerm = 04
	relroSecPerm := fallbackPerm
	genrelrosecname := func(suffix string) string {
		return suffix
	}
	seg := segro

	if ctxt.UseRelro() {
		segrelro := &Segrelrodata
		if ctxt.LinkMode == LinkExternal && ctxt.HeadType != objabi.Haix {
			// Using a separate segment with an external
			// linker results in some programs moving
			// their data sections unexpectedly, which
			// corrupts the moduledata. So we use the
			// rodata segment and let the external linker
			// sort out a rel.ro segment.
			segrelro = segro
		} else {
			// Reset datsize for new segment.
			state.datsize = 0
		}

		genrelrosecname = func(suffix string) string {
			return ".data.rel.ro" + suffix
		}
		relroReadOnly := []sym.SymKind{}
		for _, symnro := range sym.ReadOnly {
			symn := sym.RelROMap[symnro]
			relroReadOnly = append(relroReadOnly, symn)
		}
		seg = segrelro
		relroSecPerm = relroPerm

		/* data only written by relocations */
		sect = state.allocateNamedDataSection(segrelro, genrelrosecname(""), relroReadOnly, relroSecPerm)

		ctxt.Syms.Lookup("runtime.types", 0).Sect = sect
		ctxt.Syms.Lookup("runtime.etypes", 0).Sect = sect

		for i, symnro := range sym.ReadOnly {
			if i == 0 && symnro == sym.STYPE && ctxt.HeadType != objabi.Haix {
				// Skip forward so that no type
				// reference uses a zero offset.
				// This is unlikely but possible in small
				// programs with no other read-only data.
				state.datsize++
			}

			symn := sym.RelROMap[symnro]
			symnStartValue := state.datsize

			for _, s := range state.data[symn] {
				if s.Outer != nil && s.Outer.Sect != nil && s.Outer.Sect != sect {
					Errorf(s, "s.Outer (%s) in different section from s, %s != %s", s.Outer.Name, s.Outer.Sect.Name, sect.Name)
				}
			}
			state.assignToSection(sect, symn, sym.SRODATA)
			if ctxt.HeadType == objabi.Haix {
				// Read-only symbols might be wrapped inside their outer
				// symbol.
				// XCOFF symbol table needs to know the size of
				// these outer symbols.
				xcoffUpdateOuterSize(ctxt, state.datsize-symnStartValue, symn)
			}
		}

		sect.Length = uint64(state.datsize) - sect.Vaddr
	}

	/* typelink */
	sect = state.allocateNamedDataSection(seg, genrelrosecname(".typelink"), []sym.SymKind{sym.STYPELINK}, relroSecPerm)
	typelink := ctxt.Syms.Lookup("runtime.typelink", 0)
	typelink.Sect = sect
	typelink.Type = sym.SRODATA
	state.datsize += typelink.Size
	state.checkdatsize(sym.STYPELINK)
	sect.Length = uint64(state.datsize) - sect.Vaddr

	/* itablink */
	sect = state.allocateNamedSectionAndAssignSyms(seg, genrelrosecname(".itablink"), sym.SITABLINK, sym.Sxxx, relroSecPerm)
	ctxt.Syms.Lookup("runtime.itablink", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.eitablink", 0).Sect = sect
	if ctxt.HeadType == objabi.Haix {
		// Store .itablink size because its symbols are wrapped
		// under an outer symbol: runtime.itablink.
		xcoffUpdateOuterSize(ctxt, int64(sect.Length), sym.SITABLINK)
	}

	/* gosymtab */
	sect = state.allocateNamedSectionAndAssignSyms(seg, genrelrosecname(".gosymtab"), sym.SSYMTAB, sym.SRODATA, relroSecPerm)
	ctxt.Syms.Lookup("runtime.symtab", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.esymtab", 0).Sect = sect

	/* gopclntab */
	sect = state.allocateNamedSectionAndAssignSyms(seg, genrelrosecname(".gopclntab"), sym.SPCLNTAB, sym.SRODATA, relroSecPerm)
	ctxt.Syms.Lookup("runtime.pclntab", 0).Sect = sect
	ctxt.Syms.Lookup("runtime.epclntab", 0).Sect = sect

	// 6g uses 4-byte relocation offsets, so the entire segment must fit in 32 bits.
	if state.datsize != int64(uint32(state.datsize)) {
		Errorf(nil, "read-only data segment too large: %d", state.datsize)
	}

	for symn := sym.SELFRXSECT; symn < sym.SXREF; symn++ {
		ctxt.datap = append(ctxt.datap, state.data[symn]...)
	}
}

// allocateDwarfSections allocates sym.Section objects for DWARF
// symbols, and assigns symbols to sections.
func (state *dodataState) allocateDwarfSections(ctxt *Link) {

	alignOne := func(datsize int64, s *sym.Symbol) int64 { return datsize }

	for i := 0; i < len(dwarfp); i++ {
		// First the section symbol.
		s := dwarfp[i].secSym()
		sect := state.allocateNamedDataSection(&Segdwarf, s.Name, []sym.SymKind{}, 04)
		sect.Sym = s
		s.Sect = sect
		curType := s.Type
		s.Type = sym.SRODATA
		s.Value = int64(uint64(state.datsize) - sect.Vaddr)
		state.datsize += s.Size

		// Then any sub-symbols for the section symbol.
		subSyms := dwarfp[i].subSyms()
		state.assignDsymsToSection(sect, subSyms, sym.SRODATA, alignOne)

		for j := 0; j < len(subSyms); j++ {
			s := subSyms[j]
			if ctxt.HeadType == objabi.Haix && curType == sym.SDWARFLOC {
				// Update the size of .debug_loc for this symbol's
				// package.
				addDwsectCUSize(".debug_loc", s.File, uint64(s.Size))
			}
		}
		sect.Length = uint64(state.datsize) - sect.Vaddr
		state.checkdatsize(curType)
	}
}

func dodataSect(ctxt *Link, symn sym.SymKind, syms []*sym.Symbol, symToIdx map[*sym.Symbol]loader.Sym) (result []*sym.Symbol, maxAlign int32) {
	if ctxt.HeadType == objabi.Hdarwin {
		// Some symbols may no longer belong in syms
		// due to movement in machosymorder.
		newSyms := make([]*sym.Symbol, 0, len(syms))
		for _, s := range syms {
			if s.Type == symn {
				newSyms = append(newSyms, s)
			}
		}
		syms = newSyms
	}

	var head, tail *sym.Symbol
	symsSort := make([]dataSortKey, 0, len(syms))
	for _, s := range syms {
		if s.Attr.OnList() {
			log.Fatalf("symbol %s listed multiple times", s.Name)
		}
		s.Attr |= sym.AttrOnList
		switch {
		case s.Size < int64(len(s.P)):
			Errorf(s, "initialize bounds (%d < %d)", s.Size, len(s.P))
		case s.Size < 0:
			Errorf(s, "negative size (%d bytes)", s.Size)
		case s.Size > cutoff:
			Errorf(s, "symbol too large (%d bytes)", s.Size)
		}

		// If the usually-special section-marker symbols are being laid
		// out as regular symbols, put them either at the beginning or
		// end of their section.
		if (ctxt.DynlinkingGo() && ctxt.HeadType == objabi.Hdarwin) || (ctxt.HeadType == objabi.Haix && ctxt.LinkMode == LinkExternal) {
			switch s.Name {
			case "runtime.text", "runtime.bss", "runtime.data", "runtime.types", "runtime.rodata":
				head = s
				continue
			case "runtime.etext", "runtime.ebss", "runtime.edata", "runtime.etypes", "runtime.erodata":
				tail = s
				continue
			}
		}

		key := dataSortKey{
			size:   s.Size,
			name:   s.Name,
			sym:    s,
			symIdx: symToIdx[s],
		}

		switch s.Type {
		case sym.SELFGOT:
			// For ppc64, we want to interleave the .got and .toc sections
			// from input files. Both are type sym.SELFGOT, so in that case
			// we skip size comparison and fall through to the name
			// comparison (conveniently, .got sorts before .toc).
			key.size = 0
		}

		symsSort = append(symsSort, key)
	}

	sort.Sort(bySizeAndName(symsSort))

	off := 0
	if head != nil {
		syms[0] = head
		off++
	}
	for i, symSort := range symsSort {
		syms[i+off] = symSort.sym
		align := symalign(symSort.sym)
		if maxAlign < align {
			maxAlign = align
		}
	}
	if tail != nil {
		syms[len(syms)-1] = tail
	}

	if ctxt.IsELF && symn == sym.SELFROSECT {
		// Make .rela and .rela.plt contiguous, the ELF ABI requires this
		// and Solaris actually cares.
		reli, plti := -1, -1
		for i, s := range syms {
			switch s.Name {
			case ".rel.plt", ".rela.plt":
				plti = i
			case ".rel", ".rela":
				reli = i
			}
		}
		if reli >= 0 && plti >= 0 && plti != reli+1 {
			var first, second int
			if plti > reli {
				first, second = reli, plti
			} else {
				first, second = plti, reli
			}
			rel, plt := syms[reli], syms[plti]
			copy(syms[first+2:], syms[first+1:second])
			syms[first+0] = rel
			syms[first+1] = plt

			// Make sure alignment doesn't introduce a gap.
			// Setting the alignment explicitly prevents
			// symalign from basing it on the size and
			// getting it wrong.
			rel.Align = int32(ctxt.Arch.RegSize)
			plt.Align = int32(ctxt.Arch.RegSize)
		}
	}

	return syms, maxAlign
}

func relocsym2(target *Target, ldr *loader.Loader, err *ErrorReporter, syms *ArchSyms, s *sym.Symbol) {
	if len(s.R) == 0 {
		return
	}
	if target.IsWasm() && s.Attr.ReadOnly() {
		// The symbol's content is backed by read-only memory.
		// Copy it to writable memory to apply relocations.
		// Only need to do this on Wasm. On other platforms we
		// apply relocations to the output buffer, which is
		// always writeable.
		s.P = append([]byte(nil), s.P...)
		// No need to unset AttrReadOnly because it will not be used.
	}
	for ri := int32(0); ri < int32(len(s.R)); ri++ {
		r := &s.R[ri]
		if r.Done {
			// Relocation already processed by an earlier phase.
			continue
		}
		r.Done = true
		off := r.Off
		siz := int32(r.Siz)
		if off < 0 || off+siz > int32(len(s.P)) {
			rname := ""
			if r.Sym != nil {
				rname = r.Sym.Name
			}
			Errorf(s, "invalid relocation %s: %d+%d not in [%d,%d)", rname, off, siz, 0, len(s.P))
			continue
		}

		if r.Sym != nil && ((r.Sym.Type == sym.Sxxx && !r.Sym.Attr.VisibilityHidden()) || r.Sym.Type == sym.SXREF) {
			// When putting the runtime but not main into a shared library
			// these symbols are undefined and that's OK.
			if target.IsShared() || target.IsPlugin() {
				if r.Sym.Name == "main.main" || (!target.IsPlugin() && r.Sym.Name == "main..inittask") {
					r.Sym.Type = sym.SDYNIMPORT
				} else if strings.HasPrefix(r.Sym.Name, "go.info.") {
					// Skip go.info symbols. They are only needed to communicate
					// DWARF info between the compiler and linker.
					continue
				}
			} else {
				err.errorUnresolved2(s, r)
				continue
			}
		}

		if r.Type >= objabi.ElfRelocOffset {
			continue
		}
		if r.Siz == 0 { // informational relocation - no work to do
			continue
		}

		// We need to be able to reference dynimport symbols when linking against
		// shared libraries, and Solaris, Darwin and AIX need it always
		if !target.IsSolaris() && !target.IsDarwin() && !target.IsAIX() && r.Sym != nil && r.Sym.Type == sym.SDYNIMPORT && !target.IsDynlinkingGo() && !r.Sym.Attr.SubSymbol() {
			if !(target.IsPPC64() && target.IsExternal() && r.Sym.Name == ".TOC.") {
				Errorf(s, "unhandled relocation for %s (type %d (%s) rtype %d (%s))", r.Sym.Name, r.Sym.Type, r.Sym.Type, r.Type, sym.RelocName(target.Arch, r.Type))
			}
		}
		if r.Sym != nil && r.Sym.Type != sym.STLSBSS && r.Type != objabi.R_WEAKADDROFF && !r.Sym.Attr.Reachable() {
			Errorf(s, "unreachable sym in relocation: %s", r.Sym.Name)
		}

		if target.IsExternal() {
			r.InitExt()
		}

		// TODO(mundaym): remove this special case - see issue 14218.
		if target.IsS390X() {
			switch r.Type {
			case objabi.R_PCRELDBL:
				r.InitExt()
				r.Type = objabi.R_PCREL
				r.Variant = sym.RV_390_DBL
			case objabi.R_CALL:
				r.InitExt()
				r.Variant = sym.RV_390_DBL
			}
		}

		var o int64
		switch r.Type {
		default:
			switch siz {
			default:
				Errorf(s, "bad reloc size %#x for %s", uint32(siz), r.Sym.Name)
			case 1:
				o = int64(s.P[off])
			case 2:
				o = int64(target.Arch.ByteOrder.Uint16(s.P[off:]))
			case 4:
				o = int64(target.Arch.ByteOrder.Uint32(s.P[off:]))
			case 8:
				o = int64(target.Arch.ByteOrder.Uint64(s.P[off:]))
			}
			if offset, ok := thearch.Archreloc(target, syms, r, s, o); ok {
				o = offset
			} else {
				Errorf(s, "unknown reloc to %v: %d (%s)", r.Sym.Name, r.Type, sym.RelocName(target.Arch, r.Type))
			}
		case objabi.R_TLS_LE:
			if target.IsExternal() && target.IsElf() {
				r.Done = false
				if r.Sym == nil {
					r.Sym = syms.Tlsg
				}
				r.Xsym = r.Sym
				r.Xadd = r.Add
				o = 0
				if !target.IsAMD64() {
					o = r.Add
				}
				break
			}

			if target.IsElf() && target.IsARM() {
				// On ELF ARM, the thread pointer is 8 bytes before
				// the start of the thread-local data block, so add 8
				// to the actual TLS offset (r->sym->value).
				// This 8 seems to be a fundamental constant of
				// ELF on ARM (or maybe Glibc on ARM); it is not
				// related to the fact that our own TLS storage happens
				// to take up 8 bytes.
				o = 8 + r.Sym.Value
			} else if target.IsElf() || target.IsPlan9() || target.IsDarwin() {
				o = int64(syms.Tlsoffset) + r.Add
			} else if target.IsWindows() {
				o = r.Add
			} else {
				log.Fatalf("unexpected R_TLS_LE relocation for %v", target.HeadType)
			}
		case objabi.R_TLS_IE:
			if target.IsExternal() && target.IsElf() {
				r.Done = false
				if r.Sym == nil {
					r.Sym = syms.Tlsg
				}
				r.Xsym = r.Sym
				r.Xadd = r.Add
				o = 0
				if !target.IsAMD64() {
					o = r.Add
				}
				break
			}
			if target.IsPIE() && target.IsElf() {
				// We are linking the final executable, so we
				// can optimize any TLS IE relocation to LE.
				if thearch.TLSIEtoLE == nil {
					log.Fatalf("internal linking of TLS IE not supported on %v", target.Arch.Family)
				}
				thearch.TLSIEtoLE(s, int(off), int(r.Siz))
				o = int64(syms.Tlsoffset)
				// TODO: o += r.Add when !target.IsAmd64()?
				// Why do we treat r.Add differently on AMD64?
				// Is the external linker using Xadd at all?
			} else {
				log.Fatalf("cannot handle R_TLS_IE (sym %s) when linking internally", s.Name)
			}
		case objabi.R_ADDR:
			if target.IsExternal() && r.Sym.Type != sym.SCONST {
				r.Done = false

				// set up addend for eventual relocation via outer symbol.
				rs := r.Sym

				r.Xadd = r.Add
				for rs.Outer != nil {
					r.Xadd += Symaddr(rs) - Symaddr(rs.Outer)
					rs = rs.Outer
				}

				if rs.Type != sym.SHOSTOBJ && rs.Type != sym.SDYNIMPORT && rs.Type != sym.SUNDEFEXT && rs.Sect == nil {
					Errorf(s, "missing section for relocation target %s", rs.Name)
				}
				r.Xsym = rs

				o = r.Xadd
				if target.IsElf() {
					if target.IsAMD64() {
						o = 0
					}
				} else if target.IsDarwin() {
					if rs.Type != sym.SHOSTOBJ {
						o += Symaddr(rs)
					}
				} else if target.IsWindows() {
					// nothing to do
				} else if target.IsAIX() {
					o = Symaddr(r.Sym) + r.Add
				} else {
					Errorf(s, "unhandled pcrel relocation to %s on %v", rs.Name, target.HeadType)
				}

				break
			}

			// On AIX, a second relocation must be done by the loader,
			// as section addresses can change once loaded.
			// The "default" symbol address is still needed by the loader so
			// the current relocation can't be skipped.
			if target.IsAIX() && r.Sym.Type != sym.SDYNIMPORT {
				// It's not possible to make a loader relocation in a
				// symbol which is not inside .data section.
				// FIXME: It should be forbidden to have R_ADDR from a
				// symbol which isn't in .data. However, as .text has the
				// same address once loaded, this is possible.
				if s.Sect.Seg == &Segdata {
					Xcoffadddynrel(target, ldr, s, r)
				}
			}

			o = Symaddr(r.Sym) + r.Add

			// On amd64, 4-byte offsets will be sign-extended, so it is impossible to
			// access more than 2GB of static data; fail at link time is better than
			// fail at runtime. See https://golang.org/issue/7980.
			// Instead of special casing only amd64, we treat this as an error on all
			// 64-bit architectures so as to be future-proof.
			if int32(o) < 0 && target.Arch.PtrSize > 4 && siz == 4 {
				Errorf(s, "non-pc-relative relocation address for %s is too big: %#x (%#x + %#x)", r.Sym.Name, uint64(o), Symaddr(r.Sym), r.Add)
				errorexit()
			}
		case objabi.R_DWARFSECREF:
			if r.Sym.Sect == nil {
				Errorf(s, "missing DWARF section for relocation target %s", r.Sym.Name)
			}

			if target.IsExternal() {
				r.Done = false

				// On most platforms, the external linker needs to adjust DWARF references
				// as it combines DWARF sections. However, on Darwin, dsymutil does the
				// DWARF linking, and it understands how to follow section offsets.
				// Leaving in the relocation records confuses it (see
				// https://golang.org/issue/22068) so drop them for Darwin.
				if target.IsDarwin() {
					r.Done = true
				}

				// PE code emits IMAGE_REL_I386_SECREL and IMAGE_REL_AMD64_SECREL
				// for R_DWARFSECREF relocations, while R_ADDR is replaced with
				// IMAGE_REL_I386_DIR32, IMAGE_REL_AMD64_ADDR64 and IMAGE_REL_AMD64_ADDR32.
				// Do not replace R_DWARFSECREF with R_ADDR for windows -
				// let PE code emit correct relocations.
				if !target.IsWindows() {
					r.Type = objabi.R_ADDR
				}

				r.Xsym = r.Sym.Sect.Sym
				r.Xadd = r.Add + Symaddr(r.Sym) - int64(r.Sym.Sect.Vaddr)

				o = r.Xadd
				if target.IsElf() && target.IsAMD64() {
					o = 0
				}
				break
			}
			o = Symaddr(r.Sym) + r.Add - int64(r.Sym.Sect.Vaddr)
		case objabi.R_WEAKADDROFF:
			if !r.Sym.Attr.Reachable() {
				continue
			}
			fallthrough
		case objabi.R_ADDROFF:
			// The method offset tables using this relocation expect the offset to be relative
			// to the start of the first text section, even if there are multiple.
			if r.Sym.Sect.Name == ".text" {
				o = Symaddr(r.Sym) - int64(Segtext.Sections[0].Vaddr) + r.Add
			} else {
				o = Symaddr(r.Sym) - int64(r.Sym.Sect.Vaddr) + r.Add
			}

		case objabi.R_ADDRCUOFF:
			// debug_range and debug_loc elements use this relocation type to get an
			// offset from the start of the compile unit.
			o = Symaddr(r.Sym) + r.Add - Symaddr(ldr.Syms[r.Sym.Unit.Textp2[0]])

			// r->sym can be null when CALL $(constant) is transformed from absolute PC to relative PC call.
		case objabi.R_GOTPCREL:
			if target.IsDynlinkingGo() && target.IsDarwin() && r.Sym != nil && r.Sym.Type != sym.SCONST {
				r.Done = false
				r.Xadd = r.Add
				r.Xadd -= int64(r.Siz) // relative to address after the relocated chunk
				r.Xsym = r.Sym

				o = r.Xadd
				o += int64(r.Siz)
				break
			}
			fallthrough
		case objabi.R_CALL, objabi.R_PCREL:
			if target.IsExternal() && r.Sym != nil && r.Sym.Type == sym.SUNDEFEXT {
				// pass through to the external linker.
				r.Done = false
				r.Xadd = 0
				if target.IsElf() {
					r.Xadd -= int64(r.Siz)
				}
				r.Xsym = r.Sym
				o = 0
				break
			}
			if target.IsExternal() && r.Sym != nil && r.Sym.Type != sym.SCONST && (r.Sym.Sect != s.Sect || r.Type == objabi.R_GOTPCREL) {
				r.Done = false

				// set up addend for eventual relocation via outer symbol.
				rs := r.Sym

				r.Xadd = r.Add
				for rs.Outer != nil {
					r.Xadd += Symaddr(rs) - Symaddr(rs.Outer)
					rs = rs.Outer
				}

				r.Xadd -= int64(r.Siz) // relative to address after the relocated chunk
				if rs.Type != sym.SHOSTOBJ && rs.Type != sym.SDYNIMPORT && rs.Sect == nil {
					Errorf(s, "missing section for relocation target %s", rs.Name)
				}
				r.Xsym = rs

				o = r.Xadd
				if target.IsElf() {
					if target.IsAMD64() {
						o = 0
					}
				} else if target.IsDarwin() {
					if r.Type == objabi.R_CALL {
						if target.IsExternal() && rs.Type == sym.SDYNIMPORT {
							if target.IsAMD64() {
								// AMD64 dynamic relocations are relative to the end of the relocation.
								o += int64(r.Siz)
							}
						} else {
							if rs.Type != sym.SHOSTOBJ {
								o += int64(uint64(Symaddr(rs)) - rs.Sect.Vaddr)
							}
							o -= int64(r.Off) // relative to section offset, not symbol
						}
					} else {
						o += int64(r.Siz)
					}
				} else if target.IsWindows() && target.IsAMD64() { // only amd64 needs PCREL
					// PE/COFF's PC32 relocation uses the address after the relocated
					// bytes as the base. Compensate by skewing the addend.
					o += int64(r.Siz)
				} else {
					Errorf(s, "unhandled pcrel relocation to %s on %v", rs.Name, target.HeadType)
				}

				break
			}

			o = 0
			if r.Sym != nil {
				o += Symaddr(r.Sym)
			}

			o += r.Add - (s.Value + int64(r.Off) + int64(r.Siz))
		case objabi.R_SIZE:
			o = r.Sym.Size + r.Add

		case objabi.R_XCOFFREF:
			if !target.IsAIX() {
				Errorf(s, "find XCOFF R_REF on non-XCOFF files")
			}
			if !target.IsExternal() {
				Errorf(s, "find XCOFF R_REF with internal linking")
			}
			r.Xsym = r.Sym
			r.Xadd = r.Add
			r.Done = false

			// This isn't a real relocation so it must not update
			// its offset value.
			continue

		case objabi.R_DWARFFILEREF:
			// The final file index is saved in r.Add in dwarf.go:writelines.
			o = r.Add
		}

		if target.IsPPC64() || target.IsS390X() {
			r.InitExt()
			if r.Variant != sym.RV_NONE {
				o = thearch.Archrelocvariant(target, syms, r, s, o)
			}
		}

		if false {
			nam := "<nil>"
			var addr int64
			if r.Sym != nil {
				nam = r.Sym.Name
				addr = Symaddr(r.Sym)
			}
			xnam := "<nil>"
			if r.Xsym != nil {
				xnam = r.Xsym.Name
			}
			fmt.Printf("relocate %s %#x (%#x+%#x, size %d) => %s %#x +%#x (xsym: %s +%#x) [type %d (%s)/%d, %x]\n", s.Name, s.Value+int64(off), s.Value, r.Off, r.Siz, nam, addr, r.Add, xnam, r.Xadd, r.Type, sym.RelocName(target.Arch, r.Type), r.Variant, o)
		}
		switch siz {
		default:
			Errorf(s, "bad reloc size %#x for %s", uint32(siz), r.Sym.Name)
			fallthrough

			// TODO(rsc): Remove.
		case 1:
			s.P[off] = byte(int8(o))
		case 2:
			if o != int64(int16(o)) {
				Errorf(s, "relocation address for %s is too big: %#x", r.Sym.Name, o)
			}
			i16 := int16(o)
			target.Arch.ByteOrder.PutUint16(s.P[off:], uint16(i16))
		case 4:
			if r.Type == objabi.R_PCREL || r.Type == objabi.R_CALL {
				if o != int64(int32(o)) {
					Errorf(s, "pc-relative relocation address for %s is too big: %#x", r.Sym.Name, o)
				}
			} else {
				if o != int64(int32(o)) && o != int64(uint32(o)) {
					Errorf(s, "non-pc-relative relocation address for %s is too big: %#x", r.Sym.Name, uint64(o))
				}
			}

			fl := int32(o)
			target.Arch.ByteOrder.PutUint32(s.P[off:], uint32(fl))
		case 8:
			target.Arch.ByteOrder.PutUint64(s.P[off:], uint64(o))
		}
	}
}

func (ctxt *Link) reloc2() {
	var wg sync.WaitGroup
	target := &ctxt.Target
	ldr := ctxt.loader
	reporter := &ctxt.ErrorReporter
	syms := &ctxt.ArchSyms
	wg.Add(3)
	go func() {
		if !ctxt.IsWasm() { // On Wasm, text relocations are applied in Asmb2.
			for _, s := range ctxt.Textp {
				relocsym2(target, ldr, reporter, syms, s)
			}
		}
		wg.Done()
	}()
	go func() {
		for _, s := range ctxt.datap {
			relocsym2(target, ldr, reporter, syms, s)
		}
		wg.Done()
	}()
	go func() {
		for _, si := range dwarfp {
			for _, s := range si.syms {
				relocsym2(target, ldr, reporter, syms, s)
			}
		}
		wg.Done()
	}()
	wg.Wait()
}
