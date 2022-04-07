// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO/NICETOHAVE:
//   - eliminate DW_CLS_ if not used
//   - package info in compilation units
//   - assign types to their packages
//   - gdb uses c syntax, meaning clumsy quoting is needed for go identifiers. eg
//     ptype struct '[]uint8' and qualifiers need to be quoted away
//   - file:line info for variables
//   - make strings a typedef so prettyprinters can see the underlying string type

package obj

import (
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
)

// dwctxt is a wrapper intended to satisfy the method set of
// dwarf.Context, so that functions like dwarf.PutAttrs will work with
// DIEs that use loader.Sym as opposed to *sym.Symbol. It is also
// being used as a place to store tables/maps that are useful as part
// of type conversion (this is just a convenience; it would be easy to
// split these things out into another type if need be).
type dwctxt struct {
	linkctxt    *Link
	arch        *LinkArch
	waitdefsets map[string]struct{}
	waitdef     []dwarf.Type
	// This maps type name string (e.g. "uintptr") to loader symbol for
	// the DWARF DIE for that type (e.g. "go.info.type.uintptr")

	tmap map[string]*LSym

	// This maps loader symbol for the DWARF DIE symbol generated for
	// a type (e.g. "go.info.uintptr") to the type symbol itself
	// ("type.uintptr").
	// FIXME: try converting this map (and the next one) to a single
	// array indexed by loader.Sym -- this may perform better.
	rtmap map[*LSym]*LSym

	// This maps Go type symbol (e.g. "type.XXX") to loader symbol for
	// the typedef DIE for that type (e.g. "go.info.XXX..def")
	tdmap map[*LSym]*LSym

	// Cache these type symbols, so as to avoid repeatedly looking them up
	typeRuntimeEface dwarf.Type
	typeRuntimeIface dwarf.Type
	uintptrInfoSym   *LSym

	// Used at various points in that parallel portion of DWARF gen to
	// protect against conflicting updates to globals (such as "gdbscript")
	dwmu          *sync.Mutex
	typeSymLookUp func(string) *LSym
}

var dwtypes dwarf.DWDie

// newattr attaches a new attribute to the specified DIE.
//
// FIXME: at the moment attributes are stored in a linked list in a
// fairly space-inefficient way -- it might be better to instead look
// up all attrs in a single large table, then store indices into the
// table in the DIE. This would allow us to common up storage for
// attributes that are shared by many DIEs (ex: byte size of N).
func newattr(die *dwarf.DWDie, attr uint16, cls int, value int64, data interface{}) {
	a := new(dwarf.DWAttr)
	a.Link = die.Attr
	die.Attr = a
	a.Atr = attr
	a.Cls = uint8(cls)
	a.Value = value
	a.Data = data
}

// Each DIE (except the root ones) has at least 1 attribute: its
// name. getattr moves the desired one to the front so
// frequently searched ones are found faster.
func getattr(die *dwarf.DWDie, attr uint16) *dwarf.DWAttr {
	if die.Attr.Atr == attr {
		return die.Attr
	}

	a := die.Attr
	b := a.Link
	for b != nil {
		if b.Atr == attr {
			a.Link = b.Link
			b.Link = die.Attr
			die.Attr = b
			return b
		}

		a = b
		b = b.Link
	}

	return nil
}

// Every DIE manufactured by the linker has at least an AT_name
// attribute (but it will only be written out if it is listed in the abbrev).
// The compiler does create nameless DWARF DIEs (ex: concrete subprogram
// instance).
// FIXME: it would be more efficient to bulk-allocate DIEs.
func (d *dwctxt) newdie(parent *dwarf.DWDie, abbrev int, name string) *dwarf.DWDie {
	die := new(dwarf.DWDie)
	die.Abbrev = abbrev
	die.Link = parent.Child
	parent.Child = die

	newattr(die, dwarf.DW_AT_name, dwarf.DW_CLS_STRING, int64(len(name)), name)

	// Sanity check: all DIEs created in the linker should be named.
	if name == "" {
		panic("nameless DWARF DIE")
	}

	var st objabi.SymKind
	switch abbrev {
	case dwarf.DW_ABRV_FUNCTYPEPARAM, dwarf.DW_ABRV_DOTDOTDOT, dwarf.DW_ABRV_STRUCTFIELD, dwarf.DW_ABRV_ARRAYRANGE:
		// There are no relocations against these dies, and their names
		// are not unique, so don't create a symbol.
		return die
	case dwarf.DW_ABRV_COMPUNIT, dwarf.DW_ABRV_COMPUNIT_TEXTLESS:
		// Avoid collisions with "real" symbol names.
		panic(`name = fmt.Sprintf(".pkg.%s.%d", name, len(d.linkctxt.compUnits))`)
		st = objabi.SDWARFCUINFO
	case dwarf.DW_ABRV_VARIABLE:
		st = objabi.SDWARFVAR
	default:
		// Everything else is assigned a type of SDWARFTYPE. that
		// this also includes loose ends such as STRUCT_FIELD.
		st = objabi.SDWARFTYPE
	}

	ds := d.linkctxt.Lookup(dwarf.InfoPrefix + name)
	ds.Set(AttrDuplicateOK, true)
	ds.Type = st
	die.Sym = ds
	if abbrev >= dwarf.DW_ABRV_NULLTYPE && abbrev <= dwarf.DW_ABRV_TYPEDECL {
		d.tmap[name] = ds
	}

	return die
}

func walktypedef(die *dwarf.DWDie) *dwarf.DWDie {
	if die == nil {
		return nil
	}
	// Resolve typedef if present.
	if die.Abbrev == dwarf.DW_ABRV_TYPEDECL {
		for attr := die.Attr; attr != nil; attr = attr.Link {
			if attr.Atr == dwarf.DW_AT_type && attr.Cls == dwarf.DW_CLS_REFERENCE && attr.Data != nil {
				return attr.Data.(*dwarf.DWDie)
			}
		}
	}

	return die
}

func (d *dwctxt) walksymtypedef(symIdx *LSym) *LSym {

	// We're being given the loader symbol for the type DIE, e.g.
	// "go.info.type.uintptr". Map that first to the type symbol (e.g.
	// "type.uintptr") and then to the typedef DIE for the type.
	// FIXME: this seems clunky, maybe there is a better way to do this.

	if ts, ok := d.rtmap[symIdx]; ok {
		if def, ok := d.tdmap[ts]; ok {
			return def
		}
		d.linkctxt.Diag("internal error: no entry for sym %d in tdmap\n", ts)
		return nil
	}
	d.linkctxt.Diag("internal error: no entry for sym %d in rtmap\n", symIdx)
	return nil
}

// Find child by AT_name using hashtable if available or linear scan
// if not.
func findchild(die *dwarf.DWDie, name string) *dwarf.DWDie {
	var prev *dwarf.DWDie
	for ; die != prev; prev, die = die, walktypedef(die) {
		for a := die.Child; a != nil; a = a.Link {
			if name == getattr(a, dwarf.DW_AT_name).Data {
				return a
			}
		}
		continue
	}
	return nil
}

// find looks up the loader symbol for the DWARF DIE generated for the
// type with the specified name.
func (d *dwctxt) find(name string) *LSym {
	return d.tmap[name]
}

func (d *dwctxt) mustFind(name string) *LSym {
	r := d.find(name)
	if r == nil {
		log.Fatalf("dwarf find: cannot find %s", name)
	}
	return r
}

func (d *dwctxt) newrefattr(die *dwarf.DWDie, attr uint16, ref interface{}) {
	if ref == nil {
		return
	}
	newattr(die, attr, dwarf.DW_CLS_REFERENCE, 0, ref)
}

func (d *dwctxt) dtolsym(s dwarf.Sym) *LSym {
	if s == nil {
		return nil
	}
	dws := s.(*LSym)
	return dws
}

func (d *dwctxt) putdie(syms []*LSym, die *dwarf.DWDie) []*LSym {
	s := d.dtolsym(die.Sym)
	if s == nil {
		s = syms[len(syms)-1]
	} else {
		syms = append(syms, s)
	}
	sDwsym := s
	dwarf.Uleb128put(dwCtxt{d.linkctxt}, sDwsym, int64(die.Abbrev))
	dwarf.PutAttrs(dwCtxt{d.linkctxt}, sDwsym, die.Abbrev, die.Attr)
	if dwarf.HasChildren(die) {
		for die := die.Child; die != nil; die = die.Link {
			syms = d.putdie(syms, die)
		}
		sym := syms[len(syms)-1]
		sym.WriteInt(d.linkctxt, sym.Size, 1, 0)
	}
	return syms
}

func reverselist(list **dwarf.DWDie) {
	curr := *list
	var prev *dwarf.DWDie
	for curr != nil {
		next := curr.Link
		curr.Link = prev
		prev = curr
		curr = next
	}

	*list = prev
}

func reversetree(list **dwarf.DWDie) {
	reverselist(list)
	for die := *list; die != nil; die = die.Link {
		if dwarf.HasChildren(die) {
			reversetree(&die.Child)
		}
	}
}

func newmemberoffsetattr(die *dwarf.DWDie, offs int32) {
	newattr(die, dwarf.DW_AT_data_member_location, dwarf.DW_CLS_CONSTANT, int64(offs), nil)
}

func (d *dwctxt) dotypedef(parent *dwarf.DWDie, name string, def *dwarf.DWDie) *dwarf.DWDie {
	// Only emit typedefs for real names.
	if strings.HasPrefix(name, "map[") {
		return nil
	}
	if strings.HasPrefix(name, "struct {") {
		return nil
	}
	// cmd/compile uses "noalg.struct {...}" as type name when hash and eq algorithm generation of
	// this struct type is suppressed.
	if strings.HasPrefix(name, "noalg.struct {") {
		return nil
	}
	if strings.HasPrefix(name, "chan ") {
		return nil
	}
	if name[0] == '[' || name[0] == '*' {
		return nil
	}
	if def == nil {
		d.linkctxt.Diag("dwarf: bad def in dotypedef")
	}

	// Create a new loader symbol for the typedef. We no longer
	// do lookups of typedef symbols by name, so this is going
	// to be an anonymous symbol (we want this for perf reasons).
	tds := &LSym{}
	tds.Name = dwarf.InfoPrefix + "typedef." + name
	tds.Type = objabi.SDWARFTYPE
	tds.Set(AttrDuplicateOK, true) // needed for shared linkage
	def.Sym = tds
	// The typedef entry must be created after the def,
	// so that future lookups will find the typedef instead
	// of the real definition. This hooks the typedef into any
	// circular definition loops, so that gdb can understand them.
	die := d.newdie(parent, dwarf.DW_ABRV_TYPEDECL, name)

	d.newrefattr(die, dwarf.DW_AT_type, tds)

	return die
}

// Define gotype, for composite ones recurse into constituents.
func Defgotype(t dwarf.Type) {
	typectxt.dwmu.Lock()
	defer typectxt.dwmu.Unlock()
	name := t.Name()
	if _, ok := typectxt.waitdefsets[name]; !ok {
		typectxt.waitdefsets[name] = struct{}{}
		typectxt.waitdef = append(typectxt.waitdef, t)
	}
}

// Define gotype, for composite ones recurse into constituents.
func (d *dwctxt) defgotype(t dwarf.Type) *LSym {
	if t == nil {
		return d.mustFind("<unspecified>")
	}

	// If we already have a tdmap entry for the gotype, return it.
	if ds, ok := d.tdmap[t.RuntimeType().(*LSym)]; ok {
		return ds
	}

	name := t.Name()
	//if !strings.HasPrefix(sn, "type.") {
	//	d.linkctxt.Errorf(gotype, "dwarf: type name doesn't start with \"type.\"")
	//	return d.mustFind("<unspecified>")
	//}
	//name := sn[5:] // could also decode from Type.string

	sdie := d.find(name)
	if sdie != nil {
		return sdie
	}

	gtdwSym := d.newtype(t)
	d.tdmap[t.RuntimeType().(*LSym)] = gtdwSym.Sym.(*LSym)
	return gtdwSym.Sym.(*LSym)
}

func (d *dwctxt) newtype(gotype dwarf.Type) *dwarf.DWDie {
	name := gotype.Name()
	bytesize := gotype.Size()
	var die, typedefdie *dwarf.DWDie
	switch gotype.Kind() {
	case objabi.KindBool:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BASETYPE, name)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_boolean, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindInt,
		objabi.KindInt8,
		objabi.KindInt16,
		objabi.KindInt32,
		objabi.KindInt64:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BASETYPE, name)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_signed, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindUint,
		objabi.KindUint8,
		objabi.KindUint16,
		objabi.KindUint32,
		objabi.KindUint64,
		objabi.KindUintptr:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BASETYPE, name)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_unsigned, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindFloat32,
		objabi.KindFloat64:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BASETYPE, name)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_float, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindComplex64,
		objabi.KindComplex128:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BASETYPE, name)
		newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_complex_float, 0)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindArray:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_ARRAYTYPE, name)
		typedefdie = d.dotypedef(&dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		s := gotype.Elem()
		d.newrefattr(die, dwarf.DW_AT_type, d.defgotype(s))
		fld := d.newdie(die, dwarf.DW_ABRV_ARRAYRANGE, "range")

		// use actual length not upper bound; correct for 0-length arrays.
		newattr(fld, dwarf.DW_AT_count, dwarf.DW_CLS_CONSTANT, gotype.NumElem(), 0)

		d.newrefattr(fld, dwarf.DW_AT_type, d.uintptrInfoSym)

	case objabi.KindChan:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_CHANTYPE, name)
		s := gotype.Elem()
		d.newrefattr(die, dwarf.DW_AT_go_elem, d.defgotype(s))
		// Save elem type for synthesizechantypes. We could synthesize here
		// but that would change the order of DIEs we output.
		d.newrefattr(die, dwarf.DW_AT_type, s)

	case objabi.KindFunc:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_FUNCTYPE, name)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)
		typedefdie = d.dotypedef(&dwtypes, name, die)

		params := gotype.Params()

		for i := 0; i < len(params); i++ {
			fld := d.newdie(die, dwarf.DW_ABRV_FUNCTYPEPARAM, params[i].Name())
			d.newrefattr(fld, dwarf.DW_AT_type, d.defgotype(params[i].Type()))
		}
		if len(params) > 0 {
			if params[len(params)-1].IsDDD() {
				d.newdie(die, dwarf.DW_ABRV_DOTDOTDOT, "...")
			}
		}

		results := gotype.Results()
		for i := 0; i < len(results); i++ {
			fld := d.newdie(die, dwarf.DW_ABRV_FUNCTYPEPARAM, results[i].Name())
			d.newrefattr(fld, dwarf.DW_AT_type, d.defptrto(d.defgotype(results[i].Type())))
		}

	case objabi.KindInterface:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_IFACETYPE, name)
		typedefdie = d.dotypedef(&dwtypes, name, die)

		if gotype.IsEface() {
			d.newrefattr(die, dwarf.DW_AT_type, d.defgotype(d.typeRuntimeEface))
		} else {
			d.newrefattr(die, dwarf.DW_AT_type, d.defgotype(d.typeRuntimeIface))
		}

	case objabi.KindMap:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_MAPTYPE, name)

		d.newrefattr(die, dwarf.DW_AT_go_key, d.defgotype(gotype.Key()))

		d.newrefattr(die, dwarf.DW_AT_go_elem, d.defgotype(gotype.Elem()))
		// Save gotype for use in synthesizemaptypes. We could synthesize here,
		// but that would change the order of the DIEs.
		d.newrefattr(die, dwarf.DW_AT_type, gotype)

	case objabi.KindPtr:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_PTRTYPE, name)
		typedefdie = d.dotypedef(&dwtypes, name, die)

		d.newrefattr(die, dwarf.DW_AT_type, d.defgotype(gotype.Elem()))

	case objabi.KindSlice:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_SLICETYPE, name)
		typedefdie = d.dotypedef(&dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

		elem := d.defgotype(gotype.Elem())
		d.newrefattr(die, dwarf.DW_AT_go_elem, elem)

	case objabi.KindString:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_STRINGTYPE, name)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

	case objabi.KindStruct:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_STRUCTTYPE, name)
		typedefdie = d.dotypedef(&dwtypes, name, die)
		newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, bytesize, 0)

		for i, field := range gotype.Fields() {
			f := field.Name()
			s := field.Type()
			if f == "" {
				f = s.Name()
			}
			fld := d.newdie(die, dwarf.DW_ABRV_STRUCTFIELD, f)
			d.newrefattr(fld, dwarf.DW_AT_type, d.defgotype(s))
			newmemberoffsetattr(fld, int32(gotype.FieldOff(i)))
			if field.IsEmbed() { // is embedded field
				newattr(fld, dwarf.DW_AT_go_embedded_field, dwarf.DW_CLS_FLAG, 1, 0)
			}
		}

	case objabi.KindUnsafePointer:
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_BARE_PTRTYPE, name)

	default:
		d.linkctxt.Diag("%v, dwarf: definition of unknown kind %d", gotype, gotype.Kind())
		die = d.newdie(&dwtypes, dwarf.DW_ABRV_TYPEDECL, name)
		d.newrefattr(die, dwarf.DW_AT_type, d.mustFind("<unspecified>"))
	}

	newattr(die, dwarf.DW_AT_go_kind, dwarf.DW_CLS_CONSTANT, int64(gotype.Kind()), 0)
	linksym := gotype.RuntimeType()
	//if d.ldr.AttrReachable(gotype) {
	newattr(die, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_GO_TYPEREF, 0, linksym)
	//}
	//fmt.Println("die:", die.Sym.(*LSym).Name, ", runtime_type:", linksym.(*LSym).Name)

	// Sanity check.
	if _, ok := d.rtmap[linksym.(*LSym)]; ok {
		log.Fatalf("internal error: rtmap entry already installed\n")
	}

	ds := die.Sym.(*LSym)
	if typedefdie != nil {
		ds = typedefdie.Sym.(*LSym)
	}
	d.rtmap[ds] = linksym.(*LSym)

	if _, ok := prototypedies[name]; ok {
		prototypedies[name] = die
	}

	if typedefdie != nil {
		return typedefdie
	}
	return die
}

func (d *dwctxt) nameFromDIESym(dwtypeDIESym *LSym) string {
	sn := dwtypeDIESym.Name
	return sn[len(dwarf.InfoPrefix):]
}

func (d *dwctxt) defptrto(dwtype *LSym) *LSym {

	// FIXME: it would be nice if the compiler attached an aux symbol
	// ref from the element type to the pointer type -- it would be
	// more efficient to do it this way as opposed to via name lookups.

	ptrname := "*" + d.nameFromDIESym(dwtype)
	if die := d.find(ptrname); die != nil {
		return die
	}

	pdie := d.newdie(&dwtypes, dwarf.DW_ABRV_PTRTYPE, ptrname)
	d.newrefattr(pdie, dwarf.DW_AT_type, dwtype)

	// The DWARF info synthesizes pointer types that don't exist at the
	// language level, like *hash<...> and *bucket<...>, and the data
	// pointers of slices. Link to the ones we can find.
	gts := d.typeSymLookUp(ptrname)
	if gts != nil {
		newattr(pdie, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_GO_TYPEREF, 0, gts)
	}

	if gts != nil {
		ds := pdie.Sym.(*LSym)
		d.rtmap[ds] = gts
		d.tdmap[gts] = ds
	}

	return d.dtolsym(pdie.Sym)
}

// Copies src's children into dst. Copies attributes by value.
// DWAttr.data is copied as pointer only. If except is one of
// the top-level children, it will not be copied.
func (d *dwctxt) copychildrenexcept(ctxt *Link, dst *dwarf.DWDie, src *dwarf.DWDie, except *dwarf.DWDie) {
	for src = src.Child; src != nil; src = src.Link {
		if src == except {
			continue
		}
		c := d.newdie(dst, src.Abbrev, getattr(src, dwarf.DW_AT_name).Data.(string))
		for a := src.Attr; a != nil; a = a.Link {
			newattr(c, a.Atr, int(a.Cls), a.Value, a.Data)
		}
		d.copychildrenexcept(ctxt, c, src, nil)
	}

	reverselist(&dst.Child)
}

func (d *dwctxt) copychildren(ctxt *Link, dst *dwarf.DWDie, src *dwarf.DWDie) {
	d.copychildrenexcept(ctxt, dst, src, nil)
}

// Search children (assumed to have TAG_member) for the one named
// field and set its AT_type to dwtype
func (d *dwctxt) substitutetype(structdie *dwarf.DWDie, field string, dwtype *LSym) {
	child := findchild(structdie, field)
	if child == nil {
		log.Fatalf("dwarf substitutetype: %s does not have member %s",
			getattr(structdie, dwarf.DW_AT_name).Data, field)
		return
	}

	a := getattr(child, dwarf.DW_AT_type)
	if a != nil {
		a.Data = dwtype
	} else {
		d.newrefattr(child, dwarf.DW_AT_type, dwtype)
	}
}

func (d *dwctxt) findprotodie(ctxt *Link, name string) *dwarf.DWDie {
	die, ok := prototypedies[name]
	if ok && die == nil {
		d.defgotype(StubTypes[name])
		die = prototypedies[name]
	}
	if die == nil {
		log.Fatalf("internal error: DIE generation failed for %s\n", name)
	}
	return die
}

func (d *dwctxt) synthesizestringtypes(ctxt *Link, die *dwarf.DWDie) {
	var once sync.Once
	var prototype *dwarf.DWDie

	for ; die != nil; die = die.Link {
		if die.Abbrev != dwarf.DW_ABRV_STRINGTYPE {
			continue
		}
		once.Do(func() {
			prototype = walktypedef(d.findprotodie(ctxt, "runtime.stringStructDWARF"))
		})
		if prototype == nil {
			return
		}
		d.copychildren(ctxt, die, prototype)
	}
}

func (d *dwctxt) synthesizeslicetypes(ctxt *Link, die *dwarf.DWDie) {
	var once sync.Once
	var prototype *dwarf.DWDie

	for ; die != nil; die = die.Link {
		if die.Abbrev != dwarf.DW_ABRV_SLICETYPE {
			continue
		}
		once.Do(func() {
			prototype = walktypedef(d.findprotodie(ctxt, "runtime.slice"))

		})
		if prototype == nil {
			return
		}
		d.copychildren(ctxt, die, prototype)
		elem := getattr(die, dwarf.DW_AT_go_elem).Data.(*LSym)
		d.substitutetype(die, "array", d.defptrto(elem))
	}
}

func mkinternaltypename(base string, arg1 string, arg2 string) string {
	if arg2 == "" {
		return fmt.Sprintf("%s<%s>", base, arg1)
	}
	return fmt.Sprintf("%s<%s,%s>", base, arg1, arg2)
}

// synthesizemaptypes is way too closely married to runtime/hashmap.c
const (
	MaxKeySize = 128
	MaxValSize = 128
	BucketSize = 8
)

func (d *dwctxt) mkinternaltype(ctxt *Link, abbrev int, typename, keyname, valname string, f func(*dwarf.DWDie)) *LSym {
	name := mkinternaltypename(typename, keyname, valname)
	symname := dwarf.InfoPrefix + name
	s := d.linkctxt.Lookup(symname)
	if s != nil && s.Type == objabi.SDWARFTYPE {
		return s
	}
	die := d.newdie(&dwtypes, abbrev, name)
	f(die)
	return d.dtolsym(die.Sym)
}

func (d *dwctxt) synthesizemaptypes(ctxt *Link, die *dwarf.DWDie) {
	var once sync.Once
	var hash *dwarf.DWDie
	var bucket *dwarf.DWDie

	for ; die != nil; die = die.Link {
		if die.Abbrev != dwarf.DW_ABRV_MAPTYPE {
			continue
		}
		once.Do(func() {
			hash = walktypedef(d.findprotodie(ctxt, "runtime.hmap"))
		})
		if hash == nil {
			return
		}
		gotype := getattr(die, dwarf.DW_AT_type).Data.(dwarf.Type)

		keysize, valsize := gotype.Key().Size(), gotype.Elem().Size()
		keytype, valtype := d.walksymtypedef(d.defgotype(gotype.Key())), d.walksymtypedef(d.defgotype(gotype.Elem()))

		// compute size info like hashmap.c does.
		indirectKey, indirectVal := false, false
		if keysize > MaxKeySize {
			keysize = int64(d.arch.PtrSize)
			indirectKey = true
		}
		if valsize > MaxValSize {
			valsize = int64(d.arch.PtrSize)
			indirectVal = true
		}

		// Construct type to represent an array of BucketSize keys
		keyname := d.nameFromDIESym(keytype)
		dwhks := d.mkinternaltype(ctxt, dwarf.DW_ABRV_ARRAYTYPE, "[]key", keyname, "", func(dwhk *dwarf.DWDie) {
			newattr(dwhk, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, BucketSize*keysize, 0)
			t := keytype
			if indirectKey {
				t = d.defptrto(keytype)
			}
			d.newrefattr(dwhk, dwarf.DW_AT_type, t)
			fld := d.newdie(dwhk, dwarf.DW_ABRV_ARRAYRANGE, "size")
			newattr(fld, dwarf.DW_AT_count, dwarf.DW_CLS_CONSTANT, BucketSize, 0)
			d.newrefattr(fld, dwarf.DW_AT_type, d.uintptrInfoSym)
		})

		// Construct type to represent an array of BucketSize values
		valname := d.nameFromDIESym(valtype)
		dwhvs := d.mkinternaltype(ctxt, dwarf.DW_ABRV_ARRAYTYPE, "[]val", valname, "", func(dwhv *dwarf.DWDie) {
			newattr(dwhv, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, BucketSize*valsize, 0)
			t := valtype
			if indirectVal {
				t = d.defptrto(valtype)
			}
			d.newrefattr(dwhv, dwarf.DW_AT_type, t)
			fld := d.newdie(dwhv, dwarf.DW_ABRV_ARRAYRANGE, "size")
			newattr(fld, dwarf.DW_AT_count, dwarf.DW_CLS_CONSTANT, BucketSize, 0)
			d.newrefattr(fld, dwarf.DW_AT_type, d.uintptrInfoSym)
		})

		// Construct bucket<K,V>
		dwhbs := d.mkinternaltype(ctxt, dwarf.DW_ABRV_STRUCTTYPE, "bucket", keyname, valname, func(dwhb *dwarf.DWDie) {
			if bucket == nil {
				bucket = walktypedef(d.findprotodie(ctxt, "runtime.bmap"))
				if bucket == nil {
					return
				}
			}
			// Copy over all fields except the field "data" from the generic
			// bucket. "data" will be replaced with keys/values below.
			d.copychildrenexcept(ctxt, dwhb, bucket, findchild(bucket, "data"))

			fld := d.newdie(dwhb, dwarf.DW_ABRV_STRUCTFIELD, "keys")
			d.newrefattr(fld, dwarf.DW_AT_type, dwhks)
			newmemberoffsetattr(fld, BucketSize)
			fld = d.newdie(dwhb, dwarf.DW_ABRV_STRUCTFIELD, "values")
			d.newrefattr(fld, dwarf.DW_AT_type, dwhvs)
			newmemberoffsetattr(fld, BucketSize+BucketSize*int32(keysize))
			fld = d.newdie(dwhb, dwarf.DW_ABRV_STRUCTFIELD, "overflow")
			d.newrefattr(fld, dwarf.DW_AT_type, d.defptrto(d.dtolsym(dwhb.Sym)))
			newmemberoffsetattr(fld, BucketSize+BucketSize*(int32(keysize)+int32(valsize)))
			if d.arch.RegSize > d.arch.PtrSize {
				fld = d.newdie(dwhb, dwarf.DW_ABRV_STRUCTFIELD, "pad")
				d.newrefattr(fld, dwarf.DW_AT_type, d.uintptrInfoSym)
				newmemberoffsetattr(fld, BucketSize+BucketSize*(int32(keysize)+int32(valsize))+int32(d.arch.PtrSize))
			}

			newattr(dwhb, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, BucketSize+BucketSize*keysize+BucketSize*valsize+int64(d.arch.RegSize), 0)
		})

		// Construct hash<K,V>
		dwhs := d.mkinternaltype(ctxt, dwarf.DW_ABRV_STRUCTTYPE, "hash", keyname, valname, func(dwh *dwarf.DWDie) {
			d.copychildren(ctxt, dwh, hash)
			d.substitutetype(dwh, "buckets", d.defptrto(dwhbs))
			d.substitutetype(dwh, "oldbuckets", d.defptrto(dwhbs))
			newattr(dwh, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, getattr(hash, dwarf.DW_AT_byte_size).Value, nil)
		})

		// make map type a pointer to hash<K,V>
		d.newrefattr(die, dwarf.DW_AT_type, d.defptrto(dwhs))
	}
}

func (d *dwctxt) synthesizechantypes(ctxt *Link, die *dwarf.DWDie) {
	var once sync.Once
	var sudog *dwarf.DWDie
	var waitq *dwarf.DWDie
	var hchan *dwarf.DWDie

	var sudogsize int

	for ; die != nil; die = die.Link {
		if die.Abbrev != dwarf.DW_ABRV_CHANTYPE {
			continue
		}

		once.Do(func() {
			sudog = walktypedef(d.findprotodie(ctxt, "runtime.sudog"))
			waitq = walktypedef(d.findprotodie(ctxt, "runtime.waitq"))
			hchan = walktypedef(d.findprotodie(ctxt, "runtime.hchan"))
			sudogsize = int(getattr(sudog, dwarf.DW_AT_byte_size).Value)
		})

		if sudog == nil || waitq == nil || hchan == nil {
			return
		}

		elemgotype := getattr(die, dwarf.DW_AT_type).Data.(dwarf.Type)
		elemname := elemgotype.Name()
		elemtype := d.walksymtypedef(d.defgotype(elemgotype))

		// sudog<T>
		dwss := d.mkinternaltype(ctxt, dwarf.DW_ABRV_STRUCTTYPE, "sudog", elemname, "", func(dws *dwarf.DWDie) {
			d.copychildren(ctxt, dws, sudog)
			d.substitutetype(dws, "elem", d.defptrto(elemtype))
			newattr(dws, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, int64(sudogsize), nil)
		})

		// waitq<T>
		dwws := d.mkinternaltype(ctxt, dwarf.DW_ABRV_STRUCTTYPE, "waitq", elemname, "", func(dww *dwarf.DWDie) {

			d.copychildren(ctxt, dww, waitq)
			d.substitutetype(dww, "first", d.defptrto(dwss))
			d.substitutetype(dww, "last", d.defptrto(dwss))
			newattr(dww, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, getattr(waitq, dwarf.DW_AT_byte_size).Value, nil)
		})

		// hchan<T>
		dwhs := d.mkinternaltype(ctxt, dwarf.DW_ABRV_STRUCTTYPE, "hchan", elemname, "", func(dwh *dwarf.DWDie) {
			d.copychildren(ctxt, dwh, hchan)
			d.substitutetype(dwh, "recvq", dwws)
			d.substitutetype(dwh, "sendq", dwws)
			newattr(dwh, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, getattr(hchan, dwarf.DW_AT_byte_size).Value, nil)
		})

		d.newrefattr(die, dwarf.DW_AT_type, d.defptrto(dwhs))
	}
}

// FIXME: might be worth looking replacing this map with a function
// that switches based on symbol instead.

var prototypedies map[string]*dwarf.DWDie

func dwarfEnabled(ctxt *Link) bool {
	return true
}

// mkBuiltinType populates the dwctxt2 sym lookup maps for the
// newly created builtin type DIE 'typeDie'.
func (d *dwctxt) mkBuiltinType(abrv int, tname string) *dwarf.DWDie {
	// create type DIE
	die := d.newdie(&dwtypes, abrv, tname)

	// Look up type symbol.
	gotype := d.typeSymLookUp(tname)

	// Map from die sym to type sym
	ds := die.Sym.(*LSym)
	d.rtmap[ds] = gotype

	// Map from type to def sym
	d.tdmap[gotype] = ds

	return die
}

var StubTypes = map[string]dwarf.Type{}

var typectxt *dwctxt

// dwarfGenerateDebugInfo generated debug info entries for all types,
// variables and functions in the program.
// Along with dwarfGenerateDebugSyms they are the two main entry points into
// dwarf generation: dwarfGenerateDebugInfo does all the work that should be
// done before symbol names are mangled while dwarfGenerateDebugSyms does
// all the work that can only be done after addresses have been assigned to
// text symbols.
func PrepareDwarfGenerateDebugInfo(ctxt *Link, typeSymLookUp func(string) *LSym) {
	if !dwarfEnabled(ctxt) {
		return
	}

	typectxt = &dwctxt{
		linkctxt:         ctxt,
		arch:             ctxt.Arch,
		waitdefsets:      make(map[string]struct{}),
		tmap:             make(map[string]*LSym),
		tdmap:            make(map[*LSym]*LSym),
		rtmap:            make(map[*LSym]*LSym),
		dwmu:             new(sync.Mutex),
		typeSymLookUp:    typeSymLookUp,
		typeRuntimeEface: StubTypes["runtime.eface"],
		typeRuntimeIface: StubTypes["runtime.iface"],
	}

	// For ctxt.Diagnostic messages.
	newattr(&dwtypes, dwarf.DW_AT_name, dwarf.DW_CLS_STRING, int64(len("dwtypes")), "dwtypes")

	// Unspecified type. There are no references to this in the symbol table.
	typectxt.newdie(&dwtypes, dwarf.DW_ABRV_NULLTYPE, "<unspecified>")

	// Some types that must exist to define other ones (uintptr in particular
	// is needed for array size)
	typectxt.mkBuiltinType(dwarf.DW_ABRV_BARE_PTRTYPE, "unsafe.Pointer")
	die := typectxt.mkBuiltinType(dwarf.DW_ABRV_BASETYPE, "uintptr")
	newattr(die, dwarf.DW_AT_encoding, dwarf.DW_CLS_CONSTANT, dwarf.DW_ATE_unsigned, 0)
	newattr(die, dwarf.DW_AT_byte_size, dwarf.DW_CLS_CONSTANT, int64(typectxt.arch.PtrSize), 0)
	newattr(die, dwarf.DW_AT_go_kind, dwarf.DW_CLS_CONSTANT, objabi.KindUintptr, 0)
	newattr(die, dwarf.DW_AT_go_runtime_type, dwarf.DW_CLS_ADDRESS, 0, typectxt.typeSymLookUp("uintptr"))

	typectxt.uintptrInfoSym = typectxt.mustFind("uintptr")

	// Prototypes needed for type synthesis.
	prototypedies = map[string]*dwarf.DWDie{
		"runtime.stringStructDWARF": nil,
		"runtime.slice":             nil,
		"runtime.hmap":              nil,
		"runtime.bmap":              nil,
		"runtime.sudog":             nil,
		"runtime.waitq":             nil,
		"runtime.hchan":             nil,
	}

	//todo: do prettyprinter code for interface inspection.
}
func SyntheSizeTypes() {
	sort.Slice(typectxt.waitdef, func(i, j int) bool {
		if typectxt.waitdef[i].Name() == typectxt.waitdef[j].Name() {
			panic("same name typedef: " + typectxt.waitdef[i].Name())
		}
		return typectxt.waitdef[i].Name() < typectxt.waitdef[j].Name()
	})
	//if typectxt.linkctxt.Pkgpath == "internal/goarch" {
	//	fmt.Println("compile unit:", "internal/goarch")
	//}

	for _, t := range typectxt.waitdef {
		//if typectxt.linkctxt.Pkgpath == "internal/goarch" {
		//	fmt.Println("def:", t.Name())
		//
		//}
		typectxt.defgotype(t)
	}
	typectxt.synthesizestringtypes(typectxt.linkctxt, dwtypes.Child)
	typectxt.synthesizeslicetypes(typectxt.linkctxt, dwtypes.Child)
	typectxt.synthesizemaptypes(typectxt.linkctxt, dwtypes.Child)
	typectxt.synthesizechantypes(typectxt.linkctxt, dwtypes.Child)
}

// dwarfGenerateDebugSyms constructs debug_line, debug_frame, and
// debug_loc. It also writes out the debug_info section using symbols
// generated in dwarfGenerateDebugInfo2.
func DwarfGenerateDebugSyms(ctxt *Link) {
	if !dwarfEnabled(ctxt) {
		return
	}

	typectxt.dwarfGenerateDebugSyms()
}

func (d *dwctxt) dwarfGenerateDebugSyms() {
	reversetree(&dwtypes.Child)
	for die := dwtypes.Child; die != nil; die = die.Link {
		d.linkctxt.Data = d.putdie(d.linkctxt.Data, die)
	}

}
