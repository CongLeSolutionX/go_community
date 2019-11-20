// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"strings"
)

type dwSym struct {
	s loader.Sym
	l *loader.Loader
}

func (s dwSym) Len() int64 {
	return int64(len(s.l.Data(s.s)))
}

type dwctxt2 struct {
	linkctxt *Link

	// This maps type string (e.g. "type.uintptr") to loader symbol for
	// the DWARF DIE for that type (e.g. "go.info.type.uintptr")
	tmap map[string]loader.Sym

	// This maps type name (e.g. "type.XXX") to loader symbol for
	// the typedef DIE for that type (e.g. "go.info.type.XXX..def")
	tdmap map[string]loader.Sym
}

func newdwctxt2(linkctxt *Link, forTypeGen bool) dwctxt2 {
	var tmap map[string]loader.Sym
	var tdmap map[string]loader.Sym
	if forTypeGen {
		tmap = make(map[string]loader.Sym)
		tdmap = make(map[string]loader.Sym)
	}
	return dwctxt2{
		linkctxt: linkctxt,
		tmap:     tmap,
		tdmap:    tdmap,
	}
}

func (c dwctxt2) PtrSize() int {
	return c.linkctxt.Arch.PtrSize
}

func (c dwctxt2) AddInt(s dwarf.Sym, size int, i int64) {
	ds := s.(*dwSym)
	ds.l.AddUintXX(ds.s, c.linkctxt.Arch, uint64(i), size)
}

func (c dwctxt2) AddBytes(s dwarf.Sym, b []byte) {
	ds := s.(*dwSym)
	ds.l.AddBytes(ds.s, b)
}

func (c dwctxt2) AddString(s dwarf.Sym, v string) {
	ds := s.(*dwSym)
	ds.l.Addstring(ds.s, v)
}

func (c dwctxt2) AddAddress(s dwarf.Sym, data interface{}, value int64) {
	ds := s.(*dwSym)
	if value != 0 {
		value -= ds.l.Value(ds.s)
	}
	tgtds := data.(*dwSym)
	ds.l.AddAddrPlus(ds.s, c.linkctxt.Arch, tgtds.s, value)
}

func (c dwctxt2) AddCURelativeAddress(s dwarf.Sym, data interface{}, value int64) {
	ds := s.(*dwSym)
	if value != 0 {
		value -= ds.l.Value(ds.s)
	}
	tgtds := data.(*dwSym)
	ds.l.AddCURelativeAddrPlus(ds.s, c.linkctxt.Arch, tgtds.s, value)
}

func (c dwctxt2) AddSectionOffset(s dwarf.Sym, size int, t interface{}, ofs int64) {
	ds := s.(*dwSym)
	tds := t.(*dwSym)
	switch size {
	default:
		c.linkctxt.Errorf(ds.s, "invalid size %d in adddwarfref\n", size)
		fallthrough
	case c.linkctxt.Arch.PtrSize:
		ds.l.AddAddrPlus(ds.s, c.linkctxt.Arch, tds.s, 0)
	case 4:
		ds.l.AddAddrPlus4(ds.s, c.linkctxt.Arch, tds.s, 0)
	}
	relocs := ds.l.Relocs(ds.s)
	r := ds.l.MutableReloc(ds.s, relocs.Count-1)
	r.Type = objabi.R_ADDROFF
	r.Add = ofs
}

func (c dwctxt2) AddDWARFAddrSectionOffset(s dwarf.Sym, t interface{}, ofs int64) {
	size := 4
	if isDwarf64(c.linkctxt) {
		size = 8
	}

	c.AddSectionOffset(s, size, t, ofs)

	ds := s.(*dwSym)
	relocs := ds.l.Relocs(ds.s)
	r := ds.l.MutableReloc(ds.s, relocs.Count-1)
	r.Type = objabi.R_DWARFSECREF
}

func (c dwctxt2) Logf(format string, args ...interface{}) {
	c.linkctxt.Logf(format, args...)
}

// At the moment these interfaces are only used in the compiler.

func (c dwctxt2) AddFileRef(s dwarf.Sym, f interface{}) {
	panic("should be used only in the compiler")
}

func (c dwctxt2) CurrentOffset(s dwarf.Sym) int64 {
	panic("should be used only in the compiler")
}

func (c dwctxt2) RecordDclReference(s dwarf.Sym, t dwarf.Sym, dclIdx int, inlIndex int) {
	panic("should be used only in the compiler")
}

func (c dwctxt2) RecordChildDieOffsets(s dwarf.Sym, vars []*dwarf.Var, offsets []int32) {
	panic("should be used only in the compiler")
}

var dwarfp2 []loader.Sym

func writeabbrev2(ctxt *Link) loader.Sym {
	panic("not yet implemented")
	return 0
}

var dwtypes2 dwarf.DWDie

func walksymtypedef2(ctxt *Link, symIdx loader.Sym) loader.Sym {
	panic("not yet implemented")
	return 0
}

func putdie2(linkctxt *Link, ctxt dwarf.Context, die *dwarf.DWDie) {
	panic("not yet implemented")
}

// GDB doesn't like FORM_addr for AT_location, so emit a
// location expression that evals to a const.
func newabslocexprattr2(die *dwarf.DWDie, addr int64, symIdx loader.Sym) {
	newattr(die, dwarf.DW_AT_location, dwarf.DW_CLS_ADDRESS, addr, symIdx)
	// below
}

// dwarfFuncSym looks up a DWARF metadata symbol for function symbol s.
// If the symbol does not exist, it creates it if create is true,
// or returns nil otherwise.
func dwarfFuncSym2(ctxt *Link, fs loader.Sym, meta string, create bool) loader.Sym {
	panic("not yet implemented")
	return 0
}

// Search children (assumed to have TAG_member) for the one named
// field and set its AT_type to dwtype
func substitutetype2(structdie *dwarf.DWDie, field string, dwtype loader.Sym) {
	panic("not yet implemented")
}

func findprotodie2(ctxt *Link, name string) *dwarf.DWDie {
	panic("not yet implemented")
	return nil
}

func synthesizestringtypes2(ctxt *Link, die *dwarf.DWDie) {
	panic("not yet implemented")
}

func synthesizeslicetypes2(ctxt *Link, die *dwarf.DWDie) {
	panic("not yet implemented")
}

func mkinternaltype2(ctxt *Link, abbrev int, typename, keyname, valname string, f func(*dwarf.DWDie)) loader.Sym {
	panic("not yet implemented")
	return 0
}

func synthesizemaptypes2(ctxt *Link, die *dwarf.DWDie) {
	panic("not yet implemented")
}

func synthesizechantypes2(ctxt *Link, die *dwarf.DWDie) {
	panic("not yet implemented")
}

func dwarfDefineGlobal2(ctxt *Link, symIdx loader.Sym, str string, v int64, gotype loader.Sym) {
	panic("not yet implemented")
}

// createUnitLength creates the initial length field with value v and update
// offset of unit_length if needed.
func createUnitLength2(ctxt *Link, symIdx loader.Sym, v uint64) {
	panic("not yet implemented")
}

// addDwarfAddrField adds a DWARF field in DWARF 64bits or 32bits.
func addDwarfAddrField2(ctxt *Link, symIdx loader.Sym, v uint64) {
	panic("not yet implemented")
}

// addDwarfAddrRef adds a DWARF pointer in DWARF 64bits or 32bits.
func addDwarfAddrRef2(ctxt *Link, symIdx loader.Sym, t *sym.Symbol) {
	panic("not yet implemented")
}

// calcCompUnitRanges calculates the PC ranges of the compilation units.
func calcCompUnitRanges2(ctxt *Link) {
	panic("not yet implemented")
}

// If the pcln table contains runtime/proc.go, use that to set gdbscript path.
func finddebugruntimepath2(symIdx loader.Sym) {
	panic("not yet implemented")
}

/*
 * Walk prog table, emit line program and build DIE tree.
 */

func importInfoSymbol2(ctxt *Link, dsym loader.Sym) {
	panic("not yet implemented")
}

func writelines2(ctxt *Link, unit *sym.CompilationUnit, ls loader.Sym) {
	panic("not yet implemented")
}

// writepcranges generates the DW_AT_ranges table for compilation unit cu.
func writepcranges2(ctxt *Link, unit *sym.CompilationUnit, base loader.Sym, pcs []dwarf.Range, ranges loader.Sym) {
	panic("not yet implemented")
}

/*
 *  Emit .debug_frame
 */

func writeframes2(ctxt *Link) []loader.Sym {
	panic("not yet implemented")
	return nil
}

/*
 *  Walk DWarfDebugInfoEntries, and emit .debug_info
 */

func writeinfo2(ctxt *Link, syms []loader.Sym, units []*sym.CompilationUnit, abbrevsym loader.Sym, pubNames, pubTypes *pubWriter) []loader.Sym {
	panic("not yet implemented")
	return nil
}

/*
 *  Emit .debug_pubnames/_types.  _info must have been written before,
 *  because we need die->offs and infoo/infosize;
 */

// NB: define new pubWriter2

func writegdbscript2(ctxt *Link, syms []loader.Sym) []loader.Sym {
	panic("not yet implemented")
	return nil
}

// dwarfGenerateDebugInfo generated debug info entries for all types,
// variables and functions in the program.
// Along with dwarfGenerateDebugSyms they are the two main entry points into
// dwarf generation: dwarfGenerateDebugInfo does all the work that should be
// done before symbol names are mangled while dwarfGenerateDebugSyms does
// all the work that can only be done after addresses have been assigned to
// text symbols.
func dwarfGenerateDebugInfo2(ctxt *Link) {
	if !dwarfEnabled(ctxt) {
		return
	}

	dwarfctxt := newdwctxt2(ctxt, true)

	if ctxt.HeadType == objabi.Haix {
		// Initial map used to store package size for each DWARF section.
		dwsectCUSize = make(map[string]uint64)
	}

	// Forctxt.Diagnostic messages.
	newattr(&dwtypes, dwarf.DW_AT_name, dwarf.DW_CLS_STRING, int64(len("dwtypes")), "dwtypes")

	// Some types that must exist to define other ones.
	dwarfctxt.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_NULLTYPE, "<unspecified>", 0)

	dwarfctxt.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_NULLTYPE, "void", 0)
	dwarfctxt.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BARE_PTRTYPE, "unsafe.Pointer", 0)

	die := dwarfctxt.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, "uintptr", 0) // needed for array size
	newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_unsigned, 0)
	newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, int64(ctxt.Arch.PtrSize), 0)
	newattr(die, dwarf.DW_AT_go_kind, dwarf.DW_CLS_CONSTANT, objabi.KindUintptr, 0)
	newattr(die, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_ADDRESS, 0, lookupOrDiag2(ctxt, "type.uintptr"))

	// Prototypes needed for type synthesis.
	prototypedies = map[string]*dwarf.DWDie{
		"type.runtime.stringStructDWARF": nil,
		"type.runtime.slice":             nil,
		"type.runtime.hmap":              nil,
		"type.runtime.bmap":              nil,
		"type.runtime.sudog":             nil,
		"type.runtime.waitq":             nil,
		"type.runtime.hchan":             nil,
	}

	// Needed by the prettyprinter code for interface inspection.
	for _, typ := range []string{
		"type.runtime._type",
		"type.runtime.arraytype",
		"type.runtime.chantype",
		"type.runtime.functype",
		"type.runtime.maptype",
		"type.runtime.ptrtype",
		"type.runtime.slicetype",
		"type.runtime.structtype",
		"type.runtime.interfacetype",
		"type.runtime.itab",
		"type.runtime.imethod"} {
		dwarfctxt.defgotype(ctxt, lookupOrDiag2(ctxt, typ))
	}

	// fake root DIE for compile unit DIEs
	var dwroot dwarf.DWDie
	flagVariants := make(map[string]bool)

	for _, lib := range ctxt.Library {

		consts := ctxt.loader.Lookup(dwarf.ConstInfoPrefix+lib.Pkg, 0)
		for _, unit := range lib.Units {
			// We drop the constants into the first CU.
			if consts != 0 {
				importInfoSymbol2(ctxt, consts)
				// FIXME:
				unit.Consts2 = sym.LoaderSym(consts)
				consts = 0
			}

			ctxt.compUnits = append(ctxt.compUnits, unit)

			// We need at least one runtime unit.
			if unit.Lib.Pkg == "runtime" {
				ctxt.runtimeCU = unit
			}

			unit.DWInfo = dwarfctxt.newdie(ctxt, &dwroot, dwarf.DW_ABRV_COMPUNIT, unit.Lib.Pkg, 0)
			newattr(unit.DWInfo, dwarf.DW_AT_language, dwarf.DW_CLS_CONSTANT, int64(dwarf.DW_LANG_Go), 0)
			// OS X linker requires compilation dir or absolute path in comp unit name to output debug info.
			compDir := getCompilationDir()
			// TODO: Make this be the actual compilation directory, not
			// the linker directory. If we move CU construction into the
			// compiler, this should happen naturally.
			newattr(unit.DWInfo, dwarf.DW_AT_comp_dir, dwarf.DW_CLS_STRING, int64(len(compDir)), compDir)

			var peData []byte
			if producerExtra := ctxt.loader.Lookup(dwarf.CUInfoPrefix+"producer."+unit.Lib.Pkg, 0); producerExtra != 0 {
				peData = ctxt.loader.Data(producerExtra)
			}
			producer := "Go cmd/compile " + objabi.Version
			if len(peData) > 0 {
				// We put a semicolon before the flags to clearly
				// separate them from the version, which can be long
				// and have lots of weird things in it in development
				// versions. We promise not to put a semicolon in the
				// version, so it should be safe for readers to scan
				// forward to the semicolon.
				producer += "; " + string(peData)
				flagVariants[string(peData)] = true
			} else {
				flagVariants[""] = true
			}

			newattr(unit.DWInfo, dwarf.DW_AT_producer, dwarf.DW_CLS_STRING, int64(len(producer)), producer)

			var pkgname string
			if pnSymIdx := ctxt.loader.Lookup(dwarf.CUInfoPrefix+"packagename."+unit.Lib.Pkg, 0); pnSymIdx != 0 {
				pnsData := ctxt.loader.Data(pnSymIdx)
				pkgname = string(pnsData)
			}
			newattr(unit.DWInfo, dwarf.DW_AT_go_package_name, dwarf.DW_CLS_STRING, int64(len(pkgname)), pkgname)

			if len(unit.Textp) == 0 {
				unit.DWInfo.Abbrev = dwarf.DW_ABRV_COMPUNIT_TEXTLESS
			}

			// Scan all functions in this compilation unit, create DIEs for all
			// referenced types, create the file table for debug_line, find all
			// referenced abstract functions.
			// Collect all debug_range symbols in unit.rangeSyms
			for _, s := range unit.Textp { // textp has been dead-code-eliminated already.
				dsym := dwarfFuncSym(ctxt, s, dwarf.InfoPrefix, false)
				dsym.Attr |= sym.AttrNotInSymbolTable | sym.AttrReachable
				dsym.Type = sym.SDWARFINFO
				unit.FuncDIEs = append(unit.FuncDIEs, dsym)

				rangeSym := dwarfFuncSym(ctxt, s, dwarf.RangePrefix, false)
				if rangeSym != nil && rangeSym.Size > 0 {
					rangeSym.Attr |= sym.AttrReachable | sym.AttrNotInSymbolTable
					rangeSym.Type = sym.SDWARFRANGE
					if ctxt.HeadType == objabi.Haix {
						addDwsectCUSize(".debug_ranges", unit.Lib.Pkg, uint64(rangeSym.Size))
					}
					unit.RangeSyms = append(unit.RangeSyms, rangeSym)
				}

				for ri := 0; ri < len(dsym.R); ri++ {
					r := &dsym.R[ri]
					if r.Type == objabi.R_DWARFSECREF {
						rsym := r.Sym
						if strings.HasPrefix(rsym.Name, dwarf.InfoPrefix) && strings.HasSuffix(rsym.Name, dwarf.AbstractFuncSuffix) && !rsym.Attr.OnList() {
							// abstract function
							rsym.Attr |= sym.AttrOnList
							unit.AbsFnDIEs = append(unit.AbsFnDIEs, rsym)
							importInfoSymbol(ctxt, rsym)
						} else if rsym.Size == 0 {
							// a type we do not have a DIE for
							n := nameFromDIESym(rsym)
							defgotype(ctxt, ctxt.Syms.Lookup("type."+n, 0))
						}
					}
				}
			}
		}
	}

	// Fix for 31034: if the objects feeding into this link were compiled
	// with different sets of flags, then don't issue an error if
	// the -strictdups checks fail.
	if checkStrictDups > 1 && len(flagVariants) > 1 {
		checkStrictDups = 1
	}

	// Create DIEs for global variables and the types they use.
	genasmsym(ctxt, defdwsymb)

	// Create DIEs for variable types indirectly referenced by function
	// autos (which may not appear directly as param/var DIEs).
	for _, lib := range ctxt.Library {
		for _, unit := range lib.Units {
			lists := [][]*sym.Symbol{unit.AbsFnDIEs, unit.FuncDIEs}
			for _, list := range lists {
				for _, s := range list {
					for i := 0; i < len(s.R); i++ {
						r := &s.R[i]
						if r.Type == objabi.R_USETYPE {
							defgotype(ctxt, r.Sym)
						}
					}
				}
			}
		}
	}

	synthesizestringtypes(ctxt, dwtypes.Child)
	synthesizeslicetypes(ctxt, dwtypes.Child)
	synthesizemaptypes(ctxt, dwtypes.Child)
	synthesizechantypes(ctxt, dwtypes.Child)
}

// dwarfGenerateDebugSyms constructs debug_line, debug_frame, debug_loc,
// debug_pubnames and debug_pubtypes. It also writes out the debug_info
// section using symbols generated in dwarfGenerateDebugInfo2.
func dwarfGenerateDebugSyms2(ctxt *Link) {
	if !dwarfEnabled(ctxt) {
		return
	}
	panic("not yet implemented")
}

func collectlocs2(ctxt *Link, syms []loader.Sym, units []*sym.CompilationUnit) []loader.Sym {
	panic("not yet implemented")
	return nil
}

/*
 *  Elf.
 */
func dwarfaddshstrings2(ctxt *Link, shstrtab loader.Sym) {
	panic("not yet implemented")
}

// Add section symbols for DWARF debug info.  This is called before
// dwarfaddelfheaders.
func dwarfaddelfsectionsyms2(ctxt *Link) {
	panic("not yet implemented")
}

// dwarfcompress compresses the DWARF sections. Relocations are applied
// on the fly. After this, dwarfp will contain a different (new) set of
// symbols, and sections may have been replaced.
func dwarfcompress2(ctxt *Link) {
	panic("not yet implemented")
}

//......

// Every DIE manufactured by the linker has at least an AT_name
// attribute (but it will only be written out if it is listed in the abbrev).
// The compiler does create nameless DWARF DIEs (ex: concrete subprogram
// instance).
func (d *dwctxt2) newdie(ctxt *Link, parent *dwarf.DWDie, abbrev int, name string, version int) *dwarf.DWDie {
	die := new(dwarf.DWDie)
	die.Abbrev = abbrev
	die.Link = parent.Child
	parent.Child = die

	newattr(die, dwarf.DW_AT_name, dwarf.DW_CLS_STRING, int64(len(name)), name)

	// Q: do we need version here? My understanding is that all these
	// symbols should be version 0.
	ds := ctxt.loader.AddExtSym(dwarf.InfoPrefix+name, version)
	ctxt.loader.SetType(ds, sym.SDWARFINFO)

	// FIXME: loader currently doesn't have a hook for marking
	// a materialized symbol as AttrNotInSymbolTable. We could
	// handle this when converting into sym.Symbol I suppose.

	if abbrev >= dwarf.DW_ABRV_NULLTYPE && abbrev <= dwarf.DW_ABRV_TYPEDECL {
		die.Sym = dwSym{s: ds, l: ctxt.loader}
		d.tmap[name] = ds
	}

	return die
}

func lookupOrDiag2(ctxt *Link, n string) loader.Sym {
	symIdx := ctxt.loader.Lookup(n, 0)
	if symIdx == 0 || len(ctxt.loader.Data(symIdx)) == 0 {
		Exitf("dwarf: missing type: %s", n)
	}

	return symIdx
}

// find looks up the loader symbol for the DWARF DIE generated for the
// type with the specified name.
func (d *dwctxt2) find(name string) loader.Sym {
	if symIdx, ok := d.tmap[name]; ok {
		return symIdx
	}
	return 0
}

func (d *dwctxt2) mustFind(name string) loader.Sym {
	r := d.find(name)
	if r == 0 {
		Exitf("dwarf find: cannot find %s", name)
	}
	return r
}

// Define gotype, for composite ones recurse into constituents.
func (d *dwctxt2) defgotype(ctxt *Link, gotype loader.Sym) loader.Sym {
	if gotype == 0 {
		return d.mustFind("<unspecified>")
	}

	sn := ctxt.loader.SymName(gotype)
	if !strings.HasPrefix(sn, "type.") {
		ctxt.Errorf(gotype, "dwarf: type name doesn't start with \"type.\"")
		return d.mustFind("<unspecified>")
	}
	name := sn[5:] // could also decode from Type.string

	sdie := d.find(name)

	if sdie != 0 {
		return sdie
	}

	gtdwSym := d.newtype(ctxt, dwSym{s: gotype, l: ctxt.loader})
	return gtdwSym.Sym.(dwSym).s
}

func (d *dwctxt2) newtype(ctxt *Link, gotype dwSym) *dwarf.DWDie {
	sn := ctxt.loader.SymName(gotype.s)
	name := sn[5:] // could also decode from Type.string
	tdata := ctxt.loader.Data(gotype.s)

	kind := decodetypeKind(ctxt.Arch, tdata)
	bytesize := decodetypeSize(ctxt.Arch, tdata)

	var die, typedefdie *dwarf.DWDie
	switch kind {
	case objabi.KindBool:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, name, 0)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_boolean, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindInt,
		objabi.KindInt8,
		objabi.KindInt16,
		objabi.KindInt32,
		objabi.KindInt64:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, name, 0)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_signed, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindUint,
		objabi.KindUint8,
		objabi.KindUint16,
		objabi.KindUint32,
		objabi.KindUint64,
		objabi.KindUintptr:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, name, 0)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_unsigned, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindFloat32,
		objabi.KindFloat64:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, name, 0)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_float, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindComplex64,
		objabi.KindComplex128:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BASETYPE, name, 0)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_complex_float, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindArray:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_ARRAYTYPE, name, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		s := decodetypeArrayElem2(ctxt.loader, ctxt.Arch, gotype.s)
		newrefattr2(die, dwarf.DW_AT_type, d.defgotype(ctxt, s))
		fld := d.newdie(ctxt, die, dwarf.DW_ABRV_ARRAYRANGE, "range", 0)

		// use actual length not upper bound; correct for 0-length arrays.
		newattr(fld, dwarf.DW_AT_count, dwarf.DW_CLS_CONSTANT, decodetypeArrayLen2(ctxt.loader, ctxt.Arch, gotype.s), 0)

		newrefattr2(fld, dwarf.DW_AT_type, d.mustFind("uintptr"))

	case objabi.KindChan:
		die = newdie(ctxt, &dwtypes, dwarf.DW_ABRV_CHANTYPE, name, 0)
		s := decodetypeChanElem2(ctxt.loader, ctxt.Arch, gotype.s)
		newrefattr2(die, dwarf.DW_AT_go_elem, d.defgotype(ctxt, s))
		// Save elem type for synthesizechantypes. We could synthesize here
		// but that would change the order of DIEs we output.
		newrefattr2(die, dwarf.DW_AT_type, s)

	case objabi.KindFunc:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_FUNCTYPE, name, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		data := ctxt.loader.Data(gotype.s)
		// FIXME: add caching or reuse reloc slice.
		relocs := ctxt.loader.Relocs(gotype.s)
		rslice := relocs.ReadAll(nil)
		nfields := decodetypeFuncInCount(ctxt.Arch, data)
		for i := 0; i < nfields; i++ {
			s := decodetypeFuncInType2(ctxt.loader, ctxt.Arch, gotype.s, rslice, i)
			sn := ctxt.loader.SymName(s)
			fld := d.newdie(ctxt, die, dwarf.DW_ABRV_FUNCTYPEPARAM, sn[5:], 0)
			newrefattr2(fld, dwarf.DW_AT_type, d.defgotype(ctxt, s))
		}

		if decodetypeFuncDotdotdot(ctxt.Arch, data) {
			d.newdie(ctxt, die, dwarf.DW_ABRV_DOTDOTDOT, "...", 0)
		}
		nfields = decodetypeFuncOutCount(ctxt.Arch, data)
		for i := 0; i < nfields; i++ {
			s := decodetypeFuncOutType2(ctxt.loader, ctxt.Arch, gotype.s, rslice, i)
			sn := ctxt.loader.SymName(s)
			fld := d.newdie(ctxt, die, dwarf.DW_ABRV_FUNCTYPEPARAM, sn[5:], 0)
			newrefattr2(fld, dwarf.DW_AT_type, d.defptrto(ctxt, d.defgotype(ctxt, s)))
		}

	case objabi.KindInterface:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_IFACETYPE, name, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		data := ctxt.loader.Data(gotype.s)
		nfields := int(decodetypeIfaceMethodCount(ctxt.Arch, data))
		var s loader.Sym
		if nfields == 0 {
			s = lookupOrDiag2(ctxt, "type.runtime.eface")
		} else {
			s = lookupOrDiag2(ctxt, "type.runtime.iface")
		}
		newrefattr2(die, dwarf.DW_AT_type, d.defgotype(ctxt, s))

	case objabi.KindMap:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_MAPTYPE, name, 0)
		s := decodetypeMapKey2(ctxt.loader, ctxt.Arch, gotype.s)
		newrefattr2(die, dwarf.DW_AT_go_key, d.defgotype(ctxt, s))
		s = decodetypeMapValue2(ctxt.loader, ctxt.Arch, gotype.s)
		newrefattr2(die, dwarf.DW_AT_go_elem, d.defgotype(ctxt, s))
		// Save gotype for use in synthesizemaptypes. We could synthesize here,
		// but that would change the order of the DIEs.
		newrefattr2(die, dwarf.DW_AT_type, gotype.s)

	case objabi.KindPtr:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_PTRTYPE, name, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		s := decodetypePtrElem2(ctxt.loader, ctxt.Arch, gotype.s)
		newrefattr2(die, dwarf.DW_AT_type, d.defgotype(ctxt, s))

	case objabi.KindSlice:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_SLICETYPE, name, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		s := decodetypeArrayElem2(ctxt.loader, ctxt.Arch, gotype.s)
		elem := d.defgotype(ctxt, s)
		newrefattr2(die, dwarf.DW_AT_go_elem, elem)

	case objabi.KindString:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_STRINGTYPE, name, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindStruct:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_STRUCTTYPE, name, 0)
		typedefdie = d.dotypedef(ctxt, &dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		nfields := decodetypeStructFieldCount2(ctxt.loader, ctxt.Arch, gotype.s)
		for i := 0; i < nfields; i++ {
			f := decodetypeStructFieldName2(ctxt.loader, ctxt.Arch, gotype.s, i)
			s := decodetypeStructFieldType2(ctxt.loader, ctxt.Arch, gotype.s, i)
			if f == "" {
				sn := ctxt.loader.SymName(s)
				f = sn[5:] // skip "type."
			}
			fld := d.newdie(ctxt, die, dwarf.DW_ABRV_STRUCTFIELD, f, 0)
			newrefattr2(fld, dwarf.DW_AT_type, d.defgotype(ctxt, s))
			offsetAnon := decodetypeStructFieldOffsAnon2(ctxt.loader, ctxt.Arch, gotype.s, i)
			newmemberoffsetattr(fld, int32(offsetAnon>>1))
			if offsetAnon&1 != 0 { // is embedded field
				newattr(fld, dwarf.DW_AT_go_embedded_field, dwarf.DW_CLS_FLAG, 1, 0)
			}
		}

	case objabi.KindUnsafePointer:
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_BARE_PTRTYPE, name, 0)

	default:
		ctxt.Errorf(gotype.s, "dwarf: definition of unknown kind %d", kind)
		die = d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_TYPEDECL, name, 0)
		newrefattr2(die, dwarf.DW_AT_type, d.mustFind("<unspecified>"))
	}

	newattr(die, dwarf.DW_AT_go_kind, dwarf.DW_CLS_CONSTANT, int64(kind), 0)

	if ctxt.loader.Reachable.Has(gotype.s) {
		newattr(die, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_GO_TYPEREF, 0, gotype)
	}

	if _, ok := prototypedies[name]; ok {
		prototypedies[name] = die
	}

	if typedefdie != nil {
		return typedefdie
	}
	return die
}

func newrefattr2(die *dwarf.DWDie, attr uint16, refIdx loader.Sym) *dwarf.DWAttr {
	if refIdx == 0 {
		return nil
	}
	return newattr(die, attr, dwarf.DW_CLS_REFERENCE, 0, refIdx)
}

func (d *dwctxt2) dotypedef(ctxt *Link, parent *dwarf.DWDie, name string, def *dwarf.DWDie) *dwarf.DWDie {
	// Only emit typedefs for real names.
	if strings.HasPrefix(name, "map[") {
		return nil
	}
	if strings.HasPrefix(name, "struct {") {
		return nil
	}
	if strings.HasPrefix(name, "chan ") {
		return nil
	}
	if name[0] == '[' || name[0] == '*' {
		return nil
	}
	if def == nil {
		Errorf(nil, "dwarf: bad def in dotypedef")
	}

	// Create a new loader symbol for the typedef. We no longer
	// do lookups of typedef symbols by name, so this is going
	// to be an anonymous symbol.
	tds := ctxt.loader.CreateExtSym("")
	ctxt.loader.SetType(tds, sym.SDWARFINFO)
	def.Sym = dwSym{s: tds, l: ctxt.loader}
	// s.Attr |= sym.AttrNotInSymbolTable

	// The typedef entry must be created after the def,
	// so that future lookups will find the typedef instead
	// of the real definition. This hooks the typedef into any
	// circular definition loops, so that gdb can understand them.
	die := d.newdie(ctxt, parent, dwarf.DW_ABRV_TYPEDECL, name, 0)

	newrefattr2(die, dwarf.DW_AT_type, tds)

	return die
}

func dtolsym2(s dwarf.Sym) loader.Sym {
	if s == nil {
		return 0
	}
	dws := s.(dwSym)
	return dws.s
}

func nameFromDIESym2(ctxt *Link, dwtypeDIESym loader.Sym) string {
	sn := ctxt.loader.SymName(dwtypeDIESym)
	return sn[len(dwarf.InfoPrefix):]
}

func (d *dwctxt2) defptrto(ctxt *Link, dwtype loader.Sym) loader.Sym {
	ptrname := "*" + nameFromDIESym2(ctxt, dwtype)
	if die := d.find(ptrname); die != 0 {
		return die
	}

	pdie := d.newdie(ctxt, &dwtypes, dwarf.DW_ABRV_PTRTYPE, ptrname, 0)
	newrefattr2(pdie, dwarf.DW_AT_type, dwtype)

	// The DWARF info synthesizes pointer types that don't exist at the
	// language level, like *hash<...> and *bucket<...>, and the data
	// pointers of slices. Link to the ones we can find.
	gts := ctxt.loader.Lookup("type."+ptrname, 0)
	if gts != 0 && ctxt.loader.Reachable.Has(gts) {
		newattr(pdie, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_GO_TYPEREF, 0, gts)
	}
	return dtolsym2(pdie.Sym)
}
