// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

import (
	"bytes"
	"cmd/internal/bio"
	"cmd/internal/dwarf"
	"cmd/internal/goobj2"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"fmt"
	"log"
	"math/bits"
	"os"
	"sort"
	"strconv"
	"strings"
)

var _ = fmt.Print

// Sym encapsulates a global symbol index, used to identify a specific
// Go symbol. The 0-valued Sym is corresponds to an invalid symbol.
type Sym int

// Relocs encapsulates the set of relocations on a given symbol; an
// instance of this type is returned by the Loader Relocs() method.
type Relocs struct {
	Count int // number of relocs

	li int      // local index of symbol whose relocs we're examining
	r  *oReader // object reader for containing package
	l  *Loader  // loader

	ext    *sym.Symbol // external symbol if not nil
	extIdx Sym         // index of external symbol we're examining or 0
}

// Reloc contains the payload for a specific relocation.
// TODO: replace this with sym.Reloc, once we change the
// relocation target from "*sym.Symbol" to "loader.Sym" in sym.Reloc.
type Reloc struct {
	Off  int32            // offset to rewrite
	Size uint8            // number of bytes to rewrite: 0, 1, 2, or 4
	Type objabi.RelocType // the relocation type
	Add  int64            // addend
	Sym  Sym              // global index of symbol the reloc addresses
}

// oReader is a wrapper type of obj.Reader, along with some
// extra information.
// TODO: rename to objReader once the old one is gone?
type oReader struct {
	*goobj2.Reader
	unit      *sym.CompilationUnit
	version   int    // version of static symbol
	flags     uint32 // read from object file
	pkgprefix string
	rcache    []Sym // cache mapping local PkgNone symbol to resolved Sym
}

type objIdx struct {
	r *oReader
	i Sym // start index
	e Sym // end index
}

type nameVer struct {
	name string
	v    int
}

type bitmap []uint32

// set the i-th bit.
func (bm bitmap) Set(i Sym) {
	n, r := uint(i)/32, uint(i)%32
	bm[n] |= 1 << r
}

// unset the i-th bit.
func (bm bitmap) Unset(i Sym) {
	n, r := uint(i)/32, uint(i)%32
	bm[n] &= ^(1 << r)
}

// whether the i-th bit is set.
func (bm bitmap) Has(i Sym) bool {
	n, r := uint(i)/32, uint(i)%32
	return bm[n]&(1<<r) != 0
}

// return current capacity of bitmap in bits.
func (bm bitmap) Cap() int {
	return len(bm) * 32
}

func makeBitmap(n int) bitmap {
	return make(bitmap, (n+31)/32)
}

// growBitmap insures that the specified bitmap has enough capacity,
// reallocating (doubling the size) if needed.
func growBitmap(reqCap int, b bitmap) bitmap {
	curCap := b.Cap()
	if reqCap > b.Cap() {
		newCap := curCap * 2
		if newCap < reqCap {
			newCap = reqCap
		}
		newb := makeBitmap(newCap)
		copy(newb[:curCap/32], b)
		b = newb
	}
	return b
}

// A Loader loads new object files and resolves indexed symbol references.
//
// Notes on the layout of global symbol index space:
//
// - Go object files are read before host object files; each Go object
//   read allocates a new chunk of global index space of size P + NP,
//   where P is the number of package defined symbols in the object and
//   NP is the number of non-package defined symbols.
//
// - In loader.LoadRefs(), the loader makes a sweep through all of the
//   non-package references in each object file and allocates sym indices
//   for any symbols that have not yet been defined (start of this space
//   is marked by loader.extStart).
//
// - Host object file loadering happens; the host object loader does
//   a name/version lookup for each symbol it finds; this can wind up
//   extending the external symbol index space range. The host object
//   loader stores symbol payloads in sym.Symbol objects, which get
//   handed off to the loader.
//
type Loader struct {
	start       map[*oReader]Sym // map from object file to its start index
	objs        []objIdx         // sorted by start index (i.e. objIdx.i)
	max         Sym              // current max index
	extStart    Sym              // index of first externally-defined symbol
	extSyms     []nameVer        // externally defined symbols
	builtinSyms []Sym            // global index of builtin symbols
	ocache      int              // index (into 'objs') of most recent lookup

	symsByName    [2]map[string]Sym // map symbol name to index, two maps are for ABI0 and ABIInternal
	extStaticSyms map[nameVer]Sym   // externally defined static symbols, keyed by name
	overwrite     map[Sym]Sym       // overwrite[i]=j if symbol j overwrites symbol i

	itablink map[Sym]struct{} // itablink[j] defined if j is go.itablink.*

	objByPkg map[string]*oReader // map package path to its Go object reader

	Syms []*sym.Symbol // indexed symbols. XXX we still make sym.Symbol for now.

	anonVersion int // most recently assigned ext static sym pseudo-version

	Reachable bitmap // bitmap of reachable symbols, indexed by global index

	// Used to implement field tracking; created during deadcode if
	// field tracking is enabled. Reachparent[K] contains the index of
	// the symbol that triggered the marking of symbol K as live.
	Reachparent []Sym

	relocBatch []sym.Reloc // for bulk allocation of relocations

	flags uint32

	strictDupMsgs int // number of strict-dup warning/errors, when FlagStrictDups is enabled

	payloads   []extSymPayload // contents of linker-materialized external syms
	hasPayload bitmap          // whether ext sym has payload, indexed by global index

	align map[Sym]int32 // stores alignment for symbols

	elfsetstring elfsetstringFunc
}

type elfsetstringFunc func(s *sym.Symbol, str string, off int)

// extSymPayload holds the payload (data + relocations) for linker-synthesized
// external symbols.
type extSymPayload struct {
	name   string // TODO: would this be better as offset into str table?
	size   int64
	value  int64
	ver    int
	kind   sym.SymKind
	relocs []Reloc
	data   []byte
}

const (
	// Loader.flags
	FlagStrictDups = 1 << iota
)

func NewLoader(flags uint32, elfsetstring elfsetstringFunc) *Loader {
	nbuiltin := goobj2.NBuiltin()
	return &Loader{
		start:         make(map[*oReader]Sym),
		objs:          []objIdx{{nil, 0, 0}},
		symsByName:    [2]map[string]Sym{make(map[string]Sym), make(map[string]Sym)},
		objByPkg:      make(map[string]*oReader),
		overwrite:     make(map[Sym]Sym),
		align:         make(map[Sym]int32),
		itablink:      make(map[Sym]struct{}),
		extStaticSyms: make(map[nameVer]Sym),
		builtinSyms:   make([]Sym, nbuiltin),
		flags:         flags,
		elfsetstring:  elfsetstring,
	}
}

// Return the start index in the global index space for a given object file.
func (l *Loader) startIndex(r *oReader) Sym {
	return l.start[r]
}

// Add object file r, return the start index.
func (l *Loader) addObj(pkg string, r *oReader) Sym {
	if _, ok := l.start[r]; ok {
		panic("already added")
	}
	pkg = objabi.PathToPrefix(pkg) // the object file contains escaped package path
	if _, ok := l.objByPkg[pkg]; !ok {
		l.objByPkg[pkg] = r
	}
	n := r.NSym() + r.NNonpkgdef()
	i := l.max + 1
	l.start[r] = i
	l.objs = append(l.objs, objIdx{r, i, i + Sym(n) - 1})
	l.max += Sym(n)
	return i
}

func (l *Loader) extStartSym() Sym {
	return l.extStart
}

func (l *Loader) extEndSym() Sym {
	return l.max
}

// Add a symbol with a given index, return if it is added.
func (l *Loader) AddSym(name string, ver int, i Sym, r *oReader, dupok bool, typ sym.SymKind) bool {
	if l.extStart != 0 {
		panic("AddSym called after AddExtSym is called")
	}
	if ver == r.version {
		// Static symbol. Add its global index but don't
		// add to name lookup table, as it cannot be
		// referenced by name.
		return true
	}
	if oldi, ok := l.symsByName[ver][name]; ok {
		if dupok {
			if l.flags&FlagStrictDups != 0 {
				l.checkdup(name, i, r, oldi)
			}
			return false
		}
		oldr, li := l.toLocal(oldi)
		oldsym := goobj2.Sym{}
		oldsym.Read(oldr.Reader, oldr.SymOff(li))
		if oldsym.Dupok() {
			return false
		}
		overwrite := r.DataSize(int(i-l.startIndex(r))) != 0
		if overwrite {
			// new symbol overwrites old symbol.
			oldtyp := sym.AbiSymKindToSymKind[objabi.SymKind(oldsym.Type)]
			if !oldtyp.IsData() && r.DataSize(li) == 0 {
				log.Fatalf("duplicated definition of symbol " + name)
			}
			l.overwrite[oldi] = i
		} else {
			// old symbol overwrites new symbol.
			if typ != sym.SDATA && typ != sym.SNOPTRDATA && typ != sym.SBSS && typ != sym.SNOPTRBSS { // only allow overwriting data symbol
				log.Fatalf("duplicated definition of symbol " + name)
			}
			l.overwrite[i] = oldi
			return false
		}
	}
	l.symsByName[ver][name] = i
	return true
}

// newExtSym creates a new external sym with the specified
// name/version.
func (l *Loader) newExtSym(name string, ver int) Sym {
	l.max++
	i := l.max
	if l.extStart == 0 {
		l.extStart = i
		l.initHasPayload()
	}
	l.extSyms = append(l.extSyms, nameVer{name, ver})
	l.growSyms(int(i))
	l.hasPayload.Set(i)
	pi := i - l.extStart
	l.payloads[pi].name = name
	l.payloads[pi].ver = ver
	return i
}

// Add an external symbol (without index). Return the index of newly added
// symbol, or 0 if not added.
func (l *Loader) AddExtSym(name string, ver int) Sym {
	i := l.Lookup(name, ver)
	if i != 0 {
		return 0
	}
	i = l.newExtSym(name, ver)
	static := ver >= sym.SymVerStatic || ver < 0
	if static {
		l.extStaticSyms[nameVer{name, ver}] = i
	} else {
		l.symsByName[ver][name] = i
	}
	return i
}

// GetExtSym looks up the symbol with the specified name/version,
// returning its Sym index. If the symbol can't be found, a new
// symbol will be created.
func (l *Loader) GetExtSym(name string, ver int) Sym {
	i := l.Lookup(name, ver)
	if i != 0 {
		return i
	}
	i = l.newExtSym(name, ver)
	static := ver >= sym.SymVerStatic || ver < 0
	if static {
		l.extStaticSyms[nameVer{name, ver}] = i
	} else {
		l.symsByName[ver][name] = i
	}
	return i
}

// isExternal returns true if this was an externally defined sym.
func (l *Loader) isExternal(i Sym) bool {
	return l.extStart != 0 && i >= l.extStart
}

// Ensure Syms slice and payload slice have enough space.
func (l *Loader) growSyms(i int) {
	n := len(l.Syms)
	if n > i {
		return
	}
	l.Syms = append(l.Syms, make([]*sym.Symbol, i+1-n)...)
	l.payloads = append(l.payloads, make([]extSymPayload, i+1-n)...)
	l.growHasPayload(i + 1)
}

// Convert a local index to a global index.
func (l *Loader) toGlobal(r *oReader, i int) Sym {
	g := l.startIndex(r) + Sym(i)
	if ov, ok := l.overwrite[g]; ok {
		return ov
	}
	return g
}

// Convert a global index to a local index.
func (l *Loader) toLocal(i Sym) (*oReader, int) {
	if ov, ok := l.overwrite[i]; ok {
		i = ov
	}
	if l.isExternal(i) {
		return nil, int(i - l.extStart)
	}
	oc := l.ocache
	if oc != 0 && i >= l.objs[oc].i && i <= l.objs[oc].e {
		return l.objs[oc].r, int(i - l.objs[oc].i)
	}
	// Search for the local object holding index i.
	// Below k is the first one that has its start index > i,
	// so k-1 is the one we want.
	k := sort.Search(len(l.objs), func(k int) bool {
		return l.objs[k].i > i
	})
	l.ocache = k - 1
	return l.objs[k-1].r, int(i - l.objs[k-1].i)
}

// rcacheGet checks for a valid entry for 's' in the readers cache,
// where 's' is a local PkgIdxNone ref or def, or zero if
// the cache is empty or doesn't contain a value for 's'.
func (or *oReader) rcacheGet(symIdx uint32) Sym {
	if len(or.rcache) > 0 {
		return or.rcache[symIdx]
	}
	return 0
}

// rcacheSet installs a new entry in the oReader's PkgNone
// resolver cache for the specified PkgIdxNone ref or def,
// allocating a new cache if needed.
func (or *oReader) rcacheSet(symIdx uint32, gsym Sym) {
	if len(or.rcache) == 0 {
		or.rcache = make([]Sym, or.NNonpkgdef()+or.NNonpkgref())
	}
	or.rcache[symIdx] = gsym
}

// Resolve a local symbol reference. Return global index.
func (l *Loader) resolve(r *oReader, s goobj2.SymRef) Sym {
	var rr *oReader
	switch p := s.PkgIdx; p {
	case goobj2.PkgIdxInvalid:
		if s.SymIdx != 0 {
			panic("bad sym ref")
		}
		return 0
	case goobj2.PkgIdxNone:
		// Check for cached version first
		if cached := r.rcacheGet(s.SymIdx); cached != 0 {
			return cached
		}
		// Resolve by name
		i := int(s.SymIdx) + r.NSym()
		osym := goobj2.Sym{}
		osym.Read(r.Reader, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
		v := abiToVer(osym.ABI, r.version)
		gsym := l.Lookup(name, v)
		// Add to cache, then return.
		r.rcacheSet(s.SymIdx, gsym)
		return gsym
	case goobj2.PkgIdxBuiltin:
		return l.builtinSyms[s.SymIdx]
	case goobj2.PkgIdxSelf:
		rr = r
	default:
		pkg := r.Pkg(int(p))
		var ok bool
		rr, ok = l.objByPkg[pkg]
		if !ok {
			log.Fatalf("reference of nonexisted package %s, from %v", pkg, r.unit.Lib)
		}
	}
	return l.toGlobal(rr, int(s.SymIdx))
}

// Look up a symbol by name, return global index, or 0 if not found.
// This is more like Syms.ROLookup than Lookup -- it doesn't create
// new symbol.
func (l *Loader) Lookup(name string, ver int) Sym {
	if ver >= sym.SymVerStatic || ver < 0 {
		return l.extStaticSyms[nameVer{name, ver}]
	}
	return l.symsByName[ver][name]
}

// Returns whether i is a dup of another symbol, and i is not
// "primary", i.e. Lookup i by name will not return i.
func (l *Loader) IsDup(i Sym) bool {
	if _, ok := l.overwrite[i]; ok {
		return true
	}
	if l.isExternal(i) {
		return false
	}
	r, li := l.toLocal(i)
	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(li))
	if !osym.Dupok() {
		return false
	}
	if osym.Name == "" {
		return false // Unnamed aux symbol cannot be dup.
	}
	if osym.ABI == goobj2.SymABIstatic {
		return false // Static symbol cannot be dup.
	}
	name := strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
	ver := abiToVer(osym.ABI, r.version)
	return l.symsByName[ver][name] != i
}

// Check that duplicate symbols have same contents.
func (l *Loader) checkdup(name string, i Sym, r *oReader, dup Sym) {
	li := int(i - l.startIndex(r))
	p := r.Data(li)
	if strings.HasPrefix(name, "go.info.") {
		p, _ = patchDWARFName1(p, r)
	}
	rdup, ldup := l.toLocal(dup)
	pdup := rdup.Data(ldup)
	if strings.HasPrefix(name, "go.info.") {
		pdup, _ = patchDWARFName1(pdup, rdup)
	}
	if bytes.Equal(p, pdup) {
		return
	}
	reason := "same length but different contents"
	if len(p) != len(pdup) {
		reason = fmt.Sprintf("new length %d != old length %d", len(p), len(pdup))
	}
	fmt.Fprintf(os.Stderr, "cmd/link: while reading object for '%v': duplicate symbol '%s', previous def at '%v', with mismatched payload: %s\n", r.unit.Lib, name, rdup.unit.Lib, reason)

	// For the moment, whitelist DWARF subprogram DIEs for
	// auto-generated wrapper functions. What seems to happen
	// here is that we get different line numbers on formal
	// params; I am guessing that the pos is being inherited
	// from the spot where the wrapper is needed.
	whitelist := strings.HasPrefix(name, "go.info.go.interface") ||
		strings.HasPrefix(name, "go.info.go.builtin") ||
		strings.HasPrefix(name, "go.debuglines")
	if !whitelist {
		l.strictDupMsgs++
	}
}

func (l *Loader) NStrictDupMsgs() int { return l.strictDupMsgs }

// Number of total symbols.
func (l *Loader) NSym() int {
	return int(l.max + 1)
}

// Number of defined Go symbols.
func (l *Loader) NDef() int {
	return int(l.extStart)
}

// Returns the raw (unpatched) name of the i-th symbol.
func (l *Loader) RawSymName(i Sym) string {
	if l.isExternal(i) {
		if s := l.Syms[i]; s != nil {
			return s.Name
		}
		if l.hasPayload.Has(i) {
			pi := i - l.extStart
			return l.payloads[pi].name
		}
		return ""
	}
	r, li := l.toLocal(i)
	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(li))
	return osym.Name
}

// Returns the (patched) name of the i-th symbol.
func (l *Loader) SymName(i Sym) string {
	if l.isExternal(i) {
		if s := l.Syms[i]; s != nil {
			return s.Name // external name should already be patched?
		}
		if l.hasPayload.Has(i) {
			pi := int(i - l.extStart)
			return l.payloads[pi].name
		}
		return ""
	}
	r, li := l.toLocal(i)
	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(li))
	return strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
}

// Returns the type of the i-th symbol.
func (l *Loader) SymType(i Sym) sym.SymKind {
	if l.isExternal(i) {
		if s := l.Syms[i]; s != nil {
			return s.Type
		}
		if l.hasPayload.Has(i) {
			pi := i - l.extStart
			return l.payloads[pi].kind
		}
		var empty sym.SymKind
		return empty
	}
	r, li := l.toLocal(i)
	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(li))
	return sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)]
}

// Set the type of an external symbol.
func (l *Loader) SetType(i Sym, kind sym.SymKind) {
	if !l.isExternal(i) {
		panic("tried to update sym type for non-external sym")
	}
	if s := l.Syms[i]; s != nil {
		s.Type = kind
		return
	}
	if !l.hasPayload.Has(i) {
		panic("try to set type for no-payload sym")
	}

	pi := i - l.extStart
	l.payloads[pi].kind = kind
}

// Returns the attributes of the i-th symbol.
func (l *Loader) SymAttr(i Sym) uint8 {
	if l.isExternal(i) {
		// TODO: do something? External symbols have different representation of attributes. For now, ReflectMethod is the only thing matters and it cannot be set by external symbol.
		return 0
	}
	r, li := l.toLocal(i)
	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(li))
	return osym.Flag
}

// Returns whether the i-th symbol has ReflectMethod attribute set.
func (l *Loader) IsReflectMethod(i Sym) bool {
	return l.SymAttr(i)&goobj2.SymFlagReflectMethod != 0
}

// Returns whether this is a Go type symbol.
func (l *Loader) IsGoType(i Sym) bool {
	return l.SymAttr(i)&goobj2.SymFlagGoType != 0
}

// Returns whether this is a "go.itablink.*" symbol.
func (l *Loader) IsItabLink(i Sym) bool {
	if _, ok := l.itablink[i]; ok {
		return true
	}
	return false
}

// Returns the symbol content of the i-th symbol. i is global index.
func (l *Loader) Data(i Sym) []byte {
	if l.isExternal(i) {
		if s := l.Syms[i]; s != nil {
			return s.P
		}
		pi := i - l.extStart
		return l.payloads[pi].data
	}
	r, li := l.toLocal(i)
	return r.Data(li)
}

// SymValue returns the value attribute for a symbol. Currently only
// makes sense for linker-materialized external symbols.
func (l *Loader) SymValue(i Sym) int64 {
	if !l.isExternal(i) {
		panic("loader.(*Loader).Value() called for non-external sym")
	}
	pi := i - l.extStart
	return l.payloads[pi].value
}

// Size returns the size for a symbol. Currently only makes sense for
// linker-materialized external symbols.
func (l *Loader) SymSize(i Sym) int64 {
	if !l.isExternal(i) {
		panic("loader.(*Loader).Size() called for non-external sym")
	}
	pi := i - l.extStart
	return l.payloads[pi].size
}

// SymAlign returns the alignment for a symbol.
func (l *Loader) SymAlign(i Sym) int32 {
	// If an alignment has been recorded, return that.
	if align, ok := l.align[i]; ok {
		return align
	}
	// TODO: would it make sense to return an arch-specific
	// alignment depending on section type? E.g. STEXT => 32,
	// SDATA => 1, etc?
	return 0
}

// SetSymAlign sets the alignment for a symbol.
func (l *Loader) SetSymAlign(i Sym, align int32) {
	// reject bad synbols
	if i < l.max {
		panic("bad symbol index in SetSymAlign")
	}
	// Reject nonsense alignments.
	// TODO: do we need this?
	if align < 0 {
		panic("bad alignment value")
	}
	if bits.OnesCount32(uint32(align)) != 1 {
		panic("bad alignment value")
	}
	l.align[i] = align
}

// Returns the number of aux symbols given a global index.
func (l *Loader) NAux(i Sym) int {
	if l.isExternal(i) {
		return 0
	}
	r, li := l.toLocal(i)
	return r.NAux(li)
}

// Returns the referred symbol of the j-th aux symbol of the i-th
// symbol.
func (l *Loader) AuxSym(i Sym, j int) Sym {
	if l.isExternal(i) {
		return 0
	}
	r, li := l.toLocal(i)
	a := goobj2.Aux{}
	a.Read(r.Reader, r.AuxOff(li, j))
	return l.resolve(r, a.Sym)
}

// ReadAuxSyms reads the aux symbol ids for the specified symbol into the
// slice passed as a parameter. If the slice capacity is not large enough, a new
// larger slice will be allocated. Final slice is returned.
func (l *Loader) ReadAuxSyms(symIdx Sym, dst []Sym) []Sym {
	if l.isExternal(symIdx) {
		return dst[:0]
	}
	naux := l.NAux(symIdx)
	if naux == 0 {
		return dst[:0]
	}

	if cap(dst) < naux {
		dst = make([]Sym, naux)
	}
	dst = dst[:0]

	r, li := l.toLocal(symIdx)
	for i := 0; i < naux; i++ {
		a := goobj2.Aux{}
		a.Read(r.Reader, r.AuxOff(li, i))
		dst = append(dst, l.resolve(r, a.Sym))
	}

	return dst
}

// OuterSym gets the outer symbol for host object loaded symbols.
func (l *Loader) OuterSym(i Sym) Sym {
	sym := l.Syms[i]
	if sym != nil && sym.Outer != nil {
		outer := sym.Outer
		return l.Lookup(outer.Name, int(outer.Version))
	}
	return 0
}

// SubSym gets the subsymbol for host object loaded symbols.
func (l *Loader) SubSym(i Sym) Sym {
	sym := l.Syms[i]
	if sym != nil && sym.Sub != nil {
		sub := sym.Sub
		return l.Lookup(sub.Name, int(sub.Version))
	}
	return 0
}

// Initialize Reachable bitmap for running deadcode pass.
func (l *Loader) InitReachable() {
	l.Reachable = makeBitmap(l.NSym())
}

// Insure that reachable bitmap has enough size.
func (l *Loader) growReachable(reqCap int) {
	if reqCap > l.Reachable.Cap() {
		l.Reachable = growBitmap(reqCap, l.Reachable)
	}
}

// Initialize hasPayload bitmap.
func (l *Loader) initHasPayload() {
	l.hasPayload = makeBitmap(l.NSym())
}

// Insure that hasPayload bitmap has enough size.
func (l *Loader) growHasPayload(reqCap int) {
	if reqCap > l.hasPayload.Cap() {
		l.hasPayload = growBitmap(reqCap, l.hasPayload)
	}
}

// At method returns the j-th reloc for a global symbol.
func (relocs *Relocs) At(j int) Reloc {
	if relocs.ext != nil {
		rel := &relocs.ext.R[j]
		return Reloc{
			Off:  rel.Off,
			Size: rel.Siz,
			Type: rel.Type,
			Add:  rel.Add,
			Sym:  relocs.l.Lookup(rel.Sym.Name, int(rel.Sym.Version)),
		}
	}

	if relocs.extIdx != 0 {
		pi := relocs.extIdx - relocs.l.extStart
		return relocs.l.payloads[pi].relocs[j]
	}

	rel := goobj2.Reloc{}
	rel.Read(relocs.r.Reader, relocs.r.RelocOff(relocs.li, j))
	target := relocs.l.resolve(relocs.r, rel.Sym)
	return Reloc{
		Off:  rel.Off,
		Size: rel.Siz,
		Type: objabi.RelocType(rel.Type),
		Add:  rel.Add,
		Sym:  target,
	}
}

// ReadAll method reads all relocations for a symbol into the
// specified slice. If the slice capacity is not large enough, a new
// larger slice will be allocated. Final slice is returned.
func (relocs *Relocs) ReadAll(dst []Reloc) []Reloc {
	if relocs.Count == 0 {
		return dst[:0]
	}

	if cap(dst) < relocs.Count {
		dst = make([]Reloc, relocs.Count)
	}
	dst = dst[:0]

	if relocs.ext != nil {
		for i := 0; i < relocs.Count; i++ {
			erel := &relocs.ext.R[i]
			rel := Reloc{
				Off:  erel.Off,
				Size: erel.Siz,
				Type: erel.Type,
				Add:  erel.Add,
				Sym:  relocs.l.Lookup(erel.Sym.Name, int(erel.Sym.Version)),
			}
			dst = append(dst, rel)
		}
		return dst
	}

	if relocs.extIdx != 0 {
		pi := relocs.extIdx - relocs.l.extStart
		dst = append(dst, relocs.l.payloads[pi].relocs...)
		return dst
	}

	off := relocs.r.RelocOff(relocs.li, 0)
	for i := 0; i < relocs.Count; i++ {
		rel := goobj2.Reloc{}
		rel.Read(relocs.r.Reader, off)
		off += uint32(rel.Size())
		target := relocs.l.resolve(relocs.r, rel.Sym)
		dst = append(dst, Reloc{
			Off:  rel.Off,
			Size: rel.Siz,
			Type: objabi.RelocType(rel.Type),
			Add:  rel.Add,
			Sym:  target,
		})
	}
	return dst
}

// Relocs returns a Relocs object for the given global sym.
func (l *Loader) Relocs(i Sym) Relocs {
	if l.isExternal(i) {
		if s := l.Syms[i]; s != nil {
			return Relocs{Count: len(s.R), l: l, ext: s}
		}
		pi := int(i - l.extStart)
		if pi < len(l.payloads) {
			return Relocs{Count: len(l.payloads[pi].relocs), l: l, extIdx: i}
		}
		return Relocs{}
	}
	r, li := l.toLocal(i)
	if r == nil {
		panic(fmt.Sprintf("trying to get oreader for %d\n%+v\n", i, *l))
	}
	return l.relocs(r, li)
}

// Relocs returns a Relocs object given a local sym index and reader.
func (l *Loader) relocs(r *oReader, li int) Relocs {
	return Relocs{
		Count: r.NReloc(li),
		li:    li,
		r:     r,
		l:     l,
	}
}

// MutableReloc returns a pointer to relocation 'ridx' on symbol 'symIdx'.
// This method will panic if applied to a non-external symbol (relocations
// on object file symbols are treated as read-only at the moment).
func (l *Loader) MutableReloc(symIdx Sym, ridx int) *Reloc {
	if !l.isExternal(symIdx) {
		panic("loader.(*Loader).MutableReloc() called for non-external sym")
	}
	if l.Syms[symIdx] != nil {
		panic("can't mutate reloc for sym backed by sym.Symbol")
	}
	pi := symIdx - l.extStart
	return &l.payloads[pi].relocs[ridx]
}

// Preload a package: add autolibs, add symbols to the symbol table.
// Does not read symbol data yet.
func (l *Loader) Preload(arch *sys.Arch, syms *sym.Symbols, f *bio.Reader, lib *sym.Library, unit *sym.CompilationUnit, length int64, pn string, flags int) {
	roObject, readonly, err := f.Slice(uint64(length))
	if err != nil {
		log.Fatal("cannot read object file:", err)
	}
	r := goobj2.NewReaderFromBytes(roObject, readonly)
	if r == nil {
		panic("cannot read object file")
	}
	localSymVersion := syms.IncVersion()
	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."
	or := &oReader{r, unit, localSymVersion, r.Flags(), pkgprefix, nil}

	// Autolib
	lib.ImportStrings = append(lib.ImportStrings, r.Autolib()...)

	// DWARF file table
	nfile := r.NDwarfFile()
	unit.DWARFFileTable = make([]string, nfile)
	for i := range unit.DWARFFileTable {
		unit.DWARFFileTable[i] = r.DwarfFile(i)
	}

	istart := l.addObj(lib.Pkg, or)

	ndef := r.NSym()
	nnonpkgdef := r.NNonpkgdef()
	for i, n := 0, ndef+nnonpkgdef; i < n; i++ {
		osym := goobj2.Sym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		if name == "" {
			continue // don't add unnamed aux symbol
		}
		v := abiToVer(osym.ABI, localSymVersion)
		dupok := osym.Dupok()
		added := l.AddSym(name, v, istart+Sym(i), or, dupok, sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)])
		if added && strings.HasPrefix(name, "go.itablink.") {
			l.itablink[istart+Sym(i)] = struct{}{}
		}
		if added && strings.HasPrefix(name, "runtime.") {
			if bi := goobj2.BuiltinIdx(name, v); bi != -1 {
				// This is a definition of a builtin symbol. Record where it is.
				l.builtinSyms[bi] = istart + Sym(i)
			}
		}
	}

	// The caller expects us consuming all the data
	f.MustSeek(length, os.SEEK_CUR)
}

// Make sure referenced symbols are added. Most of them should already be added.
// This should only be needed for referenced external symbols.
func (l *Loader) LoadRefs(arch *sys.Arch, syms *sym.Symbols) {
	for _, o := range l.objs[1:] {
		loadObjRefs(l, o.r, arch, syms)
	}
}

func loadObjRefs(l *Loader, r *oReader, arch *sys.Arch, syms *sym.Symbols) {
	ndef := r.NSym() + r.NNonpkgdef()
	for i, n := 0, r.NNonpkgref(); i < n; i++ {
		osym := goobj2.Sym{}
		osym.Read(r.Reader, r.SymOff(ndef+i))
		name := strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
		v := abiToVer(osym.ABI, r.version)
		l.AddExtSym(name, v)
	}
}

func abiToVer(abi uint16, localSymVersion int) int {
	var v int
	if abi == goobj2.SymABIstatic {
		// Static
		v = localSymVersion
	} else if abiver := sym.ABIToVersion(obj.ABI(abi)); abiver != -1 {
		// Note that data symbols are "ABI0", which maps to version 0.
		v = abiver
	} else {
		log.Fatalf("invalid symbol ABI: %d", abi)
	}
	return v
}

func preprocess(arch *sys.Arch, s *sym.Symbol) {
	if s.Name != "" && s.Name[0] == '$' && len(s.Name) > 5 && s.Type == 0 && len(s.P) == 0 {
		x, err := strconv.ParseUint(s.Name[5:], 16, 64)
		if err != nil {
			log.Panicf("failed to parse $-symbol %s: %v", s.Name, err)
		}
		s.Type = sym.SRODATA
		s.Attr |= sym.AttrLocal
		switch s.Name[:5] {
		case "$f32.":
			if uint64(uint32(x)) != x {
				log.Panicf("$-symbol %s too large: %d", s.Name, x)
			}
			s.AddUint32(arch, uint32(x))
		case "$f64.", "$i64.":
			s.AddUint64(arch, x)
		default:
			log.Panicf("unrecognized $-symbol: %s", s.Name)
		}
	}
}

// Load full contents.
func (l *Loader) LoadFull(arch *sys.Arch, syms *sym.Symbols) {
	// create all Symbols first.
	l.growSyms(l.NSym())

	nr := 0 // total number of sym.Reloc's we'll need
	for _, o := range l.objs[1:] {
		nr += loadObjSyms(l, syms, o.r)
	}

	// allocate a single large slab of relocations for all live symbols
	l.relocBatch = make([]sym.Reloc, nr)

	// external symbols
	es := l.extStartSym()
	ee := l.extEndSym()
	if es != 0 {
		for i := es; i <= ee; i++ {
			if s := l.Syms[i]; s != nil {
				s.Attr.Set(sym.AttrReachable, l.Reachable.Has(i))
				continue // already loaded from external object
			}
			nv := l.extSyms[i-l.extStart]
			if l.Reachable.Has(i) || strings.HasPrefix(nv.name, "gofile..") { // XXX file symbols are used but not marked
				s := syms.Newsym(nv.name, nv.v)
				if l.hasPayload.Has(i) {
					// Unpack payload into sym.
					// FIXME: add a real implementation here.
					ps := l.payloads[i-l.extStart]
					if len(ps.relocs) != 0 || len(ps.data) != 0 {
						panic("need to handle this")
					}
				}
				preprocess(arch, s)
				s.Attr.Set(sym.AttrReachable, l.Reachable.Has(i))
				l.Syms[i] = s
			}
		}
	}

	// load contents of defined symbols
	for _, o := range l.objs[1:] {
		loadObjFull(l, o.r)
	}

	// Resolve ABI aliases for external symbols. This is only
	// needed for internal cgo linking.
	// (The old code does this in deadcode, but deadcode2 doesn't
	// do this.)
	for i := l.extStart; i <= l.max; i++ {
		if s := l.Syms[i]; s != nil && s.Attr.Reachable() {
			for ri := range s.R {
				r := &s.R[ri]
				if r.Sym != nil && r.Sym.Type == sym.SABIALIAS {
					r.Sym = r.Sym.R[0].Sym
				}
			}
		}
	}
}

// ExtractSymbols grabs the symbols out of the loader for work that hasn't been
// ported to the new symbol type.
func (l *Loader) ExtractSymbols(syms *sym.Symbols) {
	// Nil out overwritten symbols.
	// Overwritten Go symbols aren't a problem (as they're lazy loaded), but
	// symbols loaded from host object loaders are fully loaded, and we might
	// have multiple symbols with the same name. This loop nils them out.
	for oldI := range l.overwrite {
		l.Syms[oldI] = nil
	}

	// Add symbols to the ctxt.Syms lookup table. This explicitly
	// skips things created via loader.Create (marked with versions
	// less than zero), since if we tried to add these we'd wind up
	// with collisions. Along the way, update the version from the
	// negative anon version to something larger than sym.SymVerStatic
	// (needed so that sym.symbol.IsFileLocal() works properly).
	anonVerReplacement := syms.IncVersion()
	for _, s := range l.Syms {
		if s == nil {
			continue
		}
		if s.Name != "" && s.Version >= 0 {
			syms.Add(s)
		}
		if s.Version < 0 {
			s.Version = int16(anonVerReplacement)
		}
	}
}

// addNewSym adds a new sym.Symbol to the i-th index in the list of symbols.
func (l *Loader) addNewSym(i Sym, syms *sym.Symbols, name string, ver int, unit *sym.CompilationUnit, t sym.SymKind) *sym.Symbol {
	s := syms.Newsym(name, ver)
	if s.Type != 0 && s.Type != sym.SXREF {
		fmt.Println("symbol already processed:", unit.Lib, i, s)
		panic("symbol already processed")
	}
	if t == sym.SBSS && (s.Type == sym.SRODATA || s.Type == sym.SNOPTRBSS) {
		t = s.Type
	}
	s.Type = t
	s.Unit = unit
	l.growSyms(int(i))
	if l.Syms[i] != nil {
		panic("sym already present in addNewSym")
	}
	l.InstallSym(i, s)
	return s
}

// InstallSym sets the underlying sym.Symbol for the specified sym index.
func (l *Loader) InstallSym(i Sym, s *sym.Symbol) {
	if l.hasPayload.Has(i) {
		l.hasPayload.Unset(i)
		// don't expect to see symbol content here
		pi := i - l.extStart
		if len(l.payloads[pi].relocs) != 0 || len(l.payloads[pi].data) != 0 {
			panic("expected empty payload")
		}
	}
	l.Syms[i] = s
}

// loadObjSyms creates sym.Symbol objects for the live Syms in the
// object corresponding to object reader "r". Return value is the
// number of sym.Reloc entries required for all the new symbols.
func loadObjSyms(l *Loader, syms *sym.Symbols, r *oReader) int {
	istart := l.startIndex(r)
	nr := 0

	for i, n := 0, r.NSym()+r.NNonpkgdef(); i < n; i++ {
		// If it's been previously loaded in host object loading, we don't need to do it again.
		if s := l.Syms[istart+Sym(i)]; s != nil {
			// Mark symbol as reachable as it wasn't marked as such before.
			s.Attr.Set(sym.AttrReachable, l.Reachable.Has(istart+Sym(i)))
			nr += r.NReloc(i)
			continue
		}
		osym := goobj2.Sym{}
		osym.Read(r.Reader, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
		if name == "" {
			continue
		}
		ver := abiToVer(osym.ABI, r.version)
		if osym.ABI != goobj2.SymABIstatic && l.symsByName[ver][name] != istart+Sym(i) {
			continue
		}

		t := sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)]
		if t == sym.SXREF {
			log.Fatalf("bad sxref")
		}
		if t == 0 {
			log.Fatalf("missing type for %s in %s", name, r.unit.Lib)
		}
		if !l.Reachable.Has(istart+Sym(i)) && !(t == sym.SRODATA && strings.HasPrefix(name, "type.")) && name != "runtime.addmoduledata" && name != "runtime.lastmoduledatap" {
			// No need to load unreachable symbols.
			// XXX some type symbol's content may be needed in DWARF code, but they are not marked.
			// XXX reference to runtime.addmoduledata may be generated later by the linker in plugin mode.
			continue
		}

		s := l.addNewSym(istart+Sym(i), syms, name, ver, r.unit, t)
		s.Attr.Set(sym.AttrReachable, l.Reachable.Has(istart+Sym(i)))
		nr += r.NReloc(i)
	}
	return nr
}

// funcInfoSym records the sym.Symbol for a function, along with a copy
// of the corresponding goobj2.Sym and the index of its FuncInfo aux sym.
// We use this to delay populating FuncInfo until we can batch-allocate
// slices for their sub-objects.
type funcInfoSym struct {
	s    *sym.Symbol // sym.Symbol for a live function
	osym goobj2.Sym  // object file symbol data for that function
	isym int         // global symbol index of FuncInfo aux sym for func
}

// funcAllocInfo records totals/counts for all functions in an objfile;
// used to help with bulk allocation of sym.Symbol sub-objects.
type funcAllocInfo struct {
	symPtr  uint32 // number of *sym.Symbol's needed in file slices
	inlCall uint32 // number of sym.InlinedCall's needed in inltree slices
	pcData  uint32 // number of sym.Pcdata's needed in pdata slices
	fdOff   uint32 // number of int64's needed in all Funcdataoff slices
}

// LoadSymbol loads a single symbol by name.
// This function should only be used by the host object loaders.
// NB: This function does NOT set the symbol as reachable.
func (l *Loader) LoadSymbol(name string, version int, syms *sym.Symbols) *sym.Symbol {
	global := l.Lookup(name, version)

	// If we're already loaded, bail.
	if global != 0 && int(global) < len(l.Syms) && l.Syms[global] != nil {
		return l.Syms[global]
	}

	// Read the symbol.
	r, i := l.toLocal(global)
	istart := l.startIndex(r)

	osym := goobj2.Sym{}
	osym.Read(r.Reader, r.SymOff(int(i)))
	if l.symsByName[version][name] != istart+Sym(i) {
		return nil
	}

	return l.addNewSym(istart+Sym(i), syms, name, version, r.unit, sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)])
}

// LookupOrCreate looks up a symbol by name, and creates one if not found.
// Either way, it will also create a sym.Symbol for it, if not already.
// This should only be called when interacting with parts of the linker
// that still works on sym.Symbols (i.e. internal cgo linking, for now).
func (l *Loader) LookupOrCreate(name string, version int, syms *sym.Symbols) *sym.Symbol {
	i := l.Lookup(name, version)
	if i != 0 {
		// symbol exists
		if int(i) < len(l.Syms) && l.Syms[i] != nil {
			return l.Syms[i] // already loaded
		}
		if l.isExternal(i) {
			panic("Can't load an external symbol.")
		}
		return l.LoadSymbol(name, version, syms)
	}
	i = l.AddExtSym(name, version)
	s := syms.Newsym(name, version)
	l.Syms[i] = s
	return s
}

// CreateExtSym creates a new external symbol with the specified name
// without adding it to any lookup tables, returning a Sym index for it.
func (l *Loader) CreateExtSym(name string) Sym {
	return l.newExtSym(name, sym.SymVerABI0)
}

// Create creates a symbol with the specified name, returning a
// sym.Symbol object for it. This method is intended for static/hidden
// symbols discovered while loading host objects. We can see more than
// one instance of a given static symbol with the same name/version,
// so we can't add them to the lookup tables "as is". Instead assign
// them fictitious (unique) versions, starting at -1 and decreasing by
// one for each newly created symbol, and record them in the
// extStaticSyms hash.
func (l *Loader) Create(name string, syms *sym.Symbols) *sym.Symbol {
	i := l.CreateExtSym(name)

	// Assign a new unique negative version -- this is to mark the
	// symbol so that it can be skipped when ExtractSymbols is adding
	// ext syms to the sym.Symbols hash.
	l.anonVersion--
	ver := l.anonVersion
	l.extSyms[len(l.extSyms)-1].v = ver
	s := syms.Newsym(name, ver)
	l.InstallSym(i, s)
	l.extStaticSyms[nameVer{name, ver}] = i

	return s
}

// AddReloc adds a relocation to a linker-materialized external symbol.
func (l *Loader) AddReloc(symIdx Sym, r Reloc) {
	if !l.isExternal(symIdx) {
		panic("trying to add reloc to non-external symbol")
	}
	if !l.hasPayload.Has(symIdx) {
		panic("adding reloc to ext sym with no payload marking")
	}
	pi := symIdx - l.extStart
	l.payloads[pi].relocs = append(l.payloads[pi].relocs, r)
}

// AddBytes adds the specified data slice to a linker-materialized ext symbol.
func (l *Loader) AddBytes(symIdx Sym, data []byte) {
	if !l.isExternal(symIdx) {
		panic("trying to add data to non-external symbol")
	}
	if !l.hasPayload.Has(symIdx) {
		panic("adding data to ext sym with no payload marking")
	}
	l.Reachable.Set(symIdx)
	pi := symIdx - l.extStart
	ps := &l.payloads[pi]
	if ps.kind == 0 {
		ps.kind = sym.SDATA
	}
	ps.data = append(ps.data, data...)
	ps.size = int64(len(ps.data))
}

func (ms *extSymPayload) Grow(siz int64) {
	if int64(int(siz)) != siz {
		log.Fatalf("symgrow size %d too long", siz)
	}
	if int64(len(ms.data)) >= siz {
		return
	}
	if cap(ms.data) < int(siz) {
		data := make([]byte, 2*(siz+1))
		ms.data = append(data[:0], ms.data...)
	}
	ms.data = ms.data[:siz]
}

func (l *Loader) payload(symIdx Sym) *extSymPayload {
	if !l.isExternal(symIdx) {
		panic("trying to mutate non-external symidx")
	}
	if !l.hasPayload.Has(symIdx) {
		panic("trying to mutate non-payload external sym")
	}
	pi := symIdx - l.extStart
	return &l.payloads[pi]
}

func (l *Loader) AddUint8(symIdx Sym, v uint8) int64 {
	sp := l.payload(symIdx)
	off := sp.size
	if sp.kind == 0 {
		sp.kind = sym.SDATA
	}
	l.Reachable.Set(symIdx)
	sp.size++
	sp.data = append(sp.data, v)

	return off
}

func (l *Loader) AddUintXX(symIdx Sym, arch *sys.Arch, v uint64, wid int) int64 {
	sp := l.payload(symIdx)
	off := sp.size
	l.Reachable.Set(symIdx)
	sp.setUintXX(arch, off, v, int64(wid))
	return off
}

func (sp *extSymPayload) setUintXX(arch *sys.Arch, off int64, v uint64, wid int64) int64 {
	if sp.kind == 0 {
		sp.kind = sym.SDATA
	}
	if sp.size < off+wid {
		sp.size = off + wid
		sp.Grow(sp.size)
	}

	switch wid {
	case 1:
		sp.data[off] = uint8(v)
	case 2:
		arch.ByteOrder.PutUint16(sp.data[off:], uint16(v))
	case 4:
		arch.ByteOrder.PutUint32(sp.data[off:], uint32(v))
	case 8:
		arch.ByteOrder.PutUint64(sp.data[off:], v)
	}

	return off + wid
}

func (l *Loader) AddUint16(symIdx Sym, arch *sys.Arch, v uint16) int64 {
	return l.AddUintXX(symIdx, arch, uint64(v), 2)
}

func (l *Loader) AddUint32(symIdx Sym, arch *sys.Arch, v uint32) int64 {
	return l.AddUintXX(symIdx, arch, uint64(v), 4)
}

func (l *Loader) AddUint64(symIdx Sym, arch *sys.Arch, v uint64) int64 {
	return l.AddUintXX(symIdx, arch, v, 8)
}

func (l *Loader) AddUint(symIdx Sym, arch *sys.Arch, v uint64) int64 {
	return l.AddUintXX(symIdx, arch, v, arch.PtrSize)
}

func (l *Loader) SetUint8(symIdx Sym, arch *sys.Arch, r int64, v uint8) int64 {
	sp := l.payload(symIdx)
	l.Reachable.Set(symIdx)
	return sp.setUintXX(arch, r, uint64(v), 1)
}

func (l *Loader) SetUint16(symIdx Sym, arch *sys.Arch, r int64, v uint16) int64 {
	sp := l.payload(symIdx)
	l.Reachable.Set(symIdx)
	return sp.setUintXX(arch, r, uint64(v), 2)
}

func (l *Loader) SetUint32(symIdx Sym, arch *sys.Arch, r int64, v uint32) int64 {
	sp := l.payload(symIdx)
	l.Reachable.Set(symIdx)
	return sp.setUintXX(arch, r, uint64(v), 4)
}

func (l *Loader) SetUint(symIdx Sym, arch *sys.Arch, r int64, v uint64) int64 {
	sp := l.payload(symIdx)
	l.Reachable.Set(symIdx)
	return sp.setUintXX(arch, r, v, int64(arch.PtrSize))
}

func (l *Loader) Addstring(symIdx Sym, str string) int64 {
	sp := l.payload(symIdx)
	if sp.kind == 0 {
		sp.kind = sym.SNOPTRDATA
	}
	l.Reachable.Set(symIdx)
	r := sp.size
	if sp.name == ".shstrtab" {
		// FIME: find a better mechanism for this
		l.elfsetstring(nil, str, int(r))
	}
	sp.data = append(sp.data, str...)
	sp.data = append(sp.data, 0)
	sp.size = int64(len(sp.data))
	return r
}

func (sp *extSymPayload) addRel() *Reloc {
	sp.relocs = append(sp.relocs, Reloc{})
	return &sp.relocs[len(sp.relocs)-1]
}

func (sp *extSymPayload) addAddrPlus(tgt Sym, add int64, typ objabi.RelocType, rsize int) int64 {
	if sp.kind == 0 {
		sp.kind = sym.SDATA
	}
	i := sp.size

	sp.size += int64(rsize)
	sp.Grow(sp.size)

	r := sp.addRel()
	r.Sym = tgt
	r.Off = int32(i)
	r.Size = uint8(rsize)
	r.Type = typ
	r.Add = add

	return i + int64(r.Size)
}

func (l *Loader) AddAddrPlus(symIdx Sym, arch *sys.Arch, tgt Sym, add int64) int64 {
	l.Reachable.Set(symIdx)
	sp := l.payload(symIdx)
	return sp.addAddrPlus(tgt, add, objabi.R_ADDR, arch.PtrSize)
}

func (l *Loader) AddAddrPlus4(symIdx Sym, arch *sys.Arch, tgt Sym, add int64) int64 {
	l.Reachable.Set(symIdx)
	sp := l.payload(symIdx)
	return sp.addAddrPlus(tgt, add, objabi.R_ADDR, 4)
}

func (l *Loader) AddCURelativeAddrPlus(symIdx Sym, arch *sys.Arch, t Sym, add int64) int64 {
	l.Reachable.Set(symIdx)
	sp := l.payload(symIdx)
	return sp.addAddrPlus(t, add, objabi.R_ADDRCUOFF, arch.PtrSize)
}

//........................................................................

func loadObjFull(l *Loader, r *oReader) {
	lib := r.unit.Lib
	istart := l.startIndex(r)

	resolveSymRef := func(s goobj2.SymRef) *sym.Symbol {
		i := l.resolve(r, s)
		return l.Syms[i]
	}

	funcs := []funcInfoSym{}
	fdsyms := []*sym.Symbol{}
	var funcAllocCounts funcAllocInfo
	pcdataBase := r.PcdataBase()
	rslice := []Reloc{}
	for i, n := 0, r.NSym()+r.NNonpkgdef(); i < n; i++ {
		osym := goobj2.Sym{}
		osym.Read(r.Reader, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", r.pkgprefix, -1)
		if name == "" {
			continue
		}
		ver := abiToVer(osym.ABI, r.version)
		dupok := osym.Dupok()
		if dupok {
			if dupsym := l.symsByName[ver][name]; dupsym != istart+Sym(i) {
				if l.Reachable.Has(dupsym) {
					// A dupok symbol is resolved to another package. We still need
					// to record its presence in the current package, as the trampoline
					// pass expects packages are laid out in dependency order.
					s := l.Syms[dupsym]
					if s.Type == sym.STEXT {
						lib.DupTextSyms = append(lib.DupTextSyms, s)
					}
				}
				continue
			}
		}

		s := l.Syms[istart+Sym(i)]
		if s == nil {
			continue
		}
		if s.Name != name { // Sanity check. We can remove it in the final version.
			fmt.Println("name mismatch:", lib, i, s.Name, name)
			panic("name mismatch")
		}

		local := osym.Local()
		makeTypelink := osym.Typelink()
		size := osym.Siz

		// Symbol data
		s.P = r.Data(i)
		s.Attr.Set(sym.AttrReadOnly, r.ReadOnly())

		// Relocs
		relocs := l.relocs(r, i)
		rslice = relocs.ReadAll(rslice)
		batch := l.relocBatch
		s.R = batch[:relocs.Count:relocs.Count]
		l.relocBatch = batch[relocs.Count:]
		for j := range s.R {
			r := rslice[j]
			rs := r.Sym
			sz := r.Size
			rt := r.Type
			if rt == objabi.R_METHODOFF {
				if l.Reachable.Has(rs) {
					rt = objabi.R_ADDROFF
				} else {
					sz = 0
					rs = 0
				}
			}
			if rt == objabi.R_WEAKADDROFF && !l.Reachable.Has(rs) {
				rs = 0
				sz = 0
			}
			if rs != 0 && l.Syms[rs] != nil && l.Syms[rs].Type == sym.SABIALIAS {
				rsrelocs := l.Relocs(rs)
				rs = rsrelocs.At(0).Sym
			}
			s.R[j] = sym.Reloc{
				Off:  r.Off,
				Siz:  sz,
				Type: rt,
				Add:  r.Add,
				Sym:  l.Syms[rs],
			}
		}

		// Aux symbol info
		isym := -1
		naux := r.NAux(i)
		for j := 0; j < naux; j++ {
			a := goobj2.Aux{}
			a.Read(r.Reader, r.AuxOff(i, j))
			switch a.Type {
			case goobj2.AuxGotype:
				typ := resolveSymRef(a.Sym)
				if typ != nil {
					s.Gotype = typ
				}
			case goobj2.AuxFuncdata:
				fdsyms = append(fdsyms, resolveSymRef(a.Sym))
			case goobj2.AuxFuncInfo:
				if a.Sym.PkgIdx != goobj2.PkgIdxSelf {
					panic("funcinfo symbol not defined in current package")
				}
				isym = int(a.Sym.SymIdx)
			case goobj2.AuxDwarfInfo, goobj2.AuxDwarfLoc, goobj2.AuxDwarfRanges, goobj2.AuxDwarfLines:
				// ignored for now
			default:
				panic("unknown aux type")
			}
		}

		s.File = r.pkgprefix[:len(r.pkgprefix)-1]
		if dupok {
			s.Attr |= sym.AttrDuplicateOK
		}
		if s.Size < int64(size) {
			s.Size = int64(size)
		}
		s.Attr.Set(sym.AttrLocal, local)
		s.Attr.Set(sym.AttrMakeTypelink, makeTypelink)

		if s.Type == sym.SDWARFINFO {
			// For DWARF symbols, replace `"".` to actual package prefix
			// in the symbol content.
			// TODO: maybe we should do this in the compiler and get rid
			// of this.
			patchDWARFName(s, r)
		}

		if s.Type != sym.STEXT {
			continue
		}

		if isym == -1 {
			continue
		}

		// Record function sym and associated info for additional
		// processing in the loop below.
		fwis := funcInfoSym{s: s, isym: isym, osym: osym}
		funcs = append(funcs, fwis)

		// Read the goobj2.FuncInfo for this text symbol so that we can
		// collect allocation counts. We'll read it again in the loop
		// below.
		b := r.Data(isym)
		info := goobj2.FuncInfo{}
		info.Read(b)
		funcAllocCounts.symPtr += uint32(len(info.File))
		funcAllocCounts.pcData += uint32(len(info.Pcdata))
		funcAllocCounts.inlCall += uint32(len(info.InlTree))
		funcAllocCounts.fdOff += uint32(len(info.Funcdataoff))
	}

	// At this point we can do batch allocation of the sym.FuncInfo's,
	// along with the slices of sub-objects they use.
	fiBatch := make([]sym.FuncInfo, len(funcs))
	inlCallBatch := make([]sym.InlinedCall, funcAllocCounts.inlCall)
	symPtrBatch := make([]*sym.Symbol, funcAllocCounts.symPtr)
	pcDataBatch := make([]sym.Pcdata, funcAllocCounts.pcData)
	fdOffBatch := make([]int64, funcAllocCounts.fdOff)

	// Populate FuncInfo contents for func symbols.
	for fi := 0; fi < len(funcs); fi++ {
		s := funcs[fi].s
		isym := funcs[fi].isym
		osym := funcs[fi].osym

		s.FuncInfo = &fiBatch[0]
		fiBatch = fiBatch[1:]

		b := r.Data(isym)
		info := goobj2.FuncInfo{}
		info.Read(b)

		if info.NoSplit != 0 {
			s.Attr |= sym.AttrNoSplit
		}
		if osym.ReflectMethod() {
			s.Attr |= sym.AttrReflectMethod
		}
		if r.Flags()&goobj2.ObjFlagShared != 0 {
			s.Attr |= sym.AttrShared
		}
		if osym.TopFrame() {
			s.Attr |= sym.AttrTopFrame
		}

		pc := s.FuncInfo

		if len(info.Funcdataoff) != 0 {
			nfd := len(info.Funcdataoff)
			pc.Funcdata = fdsyms[:nfd:nfd]
			fdsyms = fdsyms[nfd:]
		}

		info.Pcdata = append(info.Pcdata, info.PcdataEnd) // for the ease of knowing where it ends
		pc.Args = int32(info.Args)
		pc.Locals = int32(info.Locals)

		npc := len(info.Pcdata) - 1 // -1 as we appended one above
		pc.Pcdata = pcDataBatch[:npc:npc]
		pcDataBatch = pcDataBatch[npc:]

		nfd := len(info.Funcdataoff)
		pc.Funcdataoff = fdOffBatch[:nfd:nfd]
		fdOffBatch = fdOffBatch[nfd:]

		nsp := len(info.File)
		pc.File = symPtrBatch[:nsp:nsp]
		symPtrBatch = symPtrBatch[nsp:]

		nic := len(info.InlTree)
		pc.InlTree = inlCallBatch[:nic:nic]
		inlCallBatch = inlCallBatch[nic:]

		pc.Pcsp.P = r.BytesAt(pcdataBase+info.Pcsp, int(info.Pcfile-info.Pcsp))
		pc.Pcfile.P = r.BytesAt(pcdataBase+info.Pcfile, int(info.Pcline-info.Pcfile))
		pc.Pcline.P = r.BytesAt(pcdataBase+info.Pcline, int(info.Pcinline-info.Pcline))
		pc.Pcinline.P = r.BytesAt(pcdataBase+info.Pcinline, int(info.Pcdata[0]-info.Pcinline))
		for k := range pc.Pcdata {
			pc.Pcdata[k].P = r.BytesAt(pcdataBase+info.Pcdata[k], int(info.Pcdata[k+1]-info.Pcdata[k]))
		}
		for k := range pc.Funcdataoff {
			pc.Funcdataoff[k] = int64(info.Funcdataoff[k])
		}
		for k := range pc.File {
			pc.File[k] = resolveSymRef(info.File[k])
		}
		for k := range pc.InlTree {
			inl := &info.InlTree[k]
			pc.InlTree[k] = sym.InlinedCall{
				Parent:   inl.Parent,
				File:     resolveSymRef(inl.File),
				Line:     inl.Line,
				Func:     l.SymName(l.resolve(r, inl.Func)),
				ParentPC: inl.ParentPC,
			}
		}

		dupok := osym.Dupok()
		if !dupok {
			if s.Attr.OnList() {
				log.Fatalf("symbol %s listed multiple times", s.Name)
			}
			s.Attr.Set(sym.AttrOnList, true)
			lib.Textp = append(lib.Textp, s)
		} else {
			// there may be a dup in another package
			// put into a temp list and add to text later
			lib.DupTextSyms = append(lib.DupTextSyms, s)
		}
	}
}

var emptyPkg = []byte(`"".`)

func patchDWARFName1(p []byte, r *oReader) ([]byte, int) {
	// This is kind of ugly. Really the package name should not
	// even be included here.
	if len(p) < 1 || p[0] != dwarf.DW_ABRV_FUNCTION {
		return p, -1
	}
	e := bytes.IndexByte(p, 0)
	if e == -1 {
		return p, -1
	}
	if !bytes.Contains(p[:e], emptyPkg) {
		return p, -1
	}
	pkgprefix := []byte(r.pkgprefix)
	patched := bytes.Replace(p[:e], emptyPkg, pkgprefix, -1)
	return append(patched, p[e:]...), e
}

func patchDWARFName(s *sym.Symbol, r *oReader) {
	patched, e := patchDWARFName1(s.P, r)
	if e == -1 {
		return
	}
	s.P = patched
	s.Attr.Set(sym.AttrReadOnly, false)
	delta := int64(len(s.P)) - s.Size
	s.Size = int64(len(s.P))
	for i := range s.R {
		r := &s.R[i]
		if r.Off > int32(e) {
			r.Off += int32(delta)
		}
	}
}

// For debugging.
func (l *Loader) Dump() {
	fmt.Println("objs")
	for _, obj := range l.objs {
		if obj.r != nil {
			fmt.Println(obj.i, obj.r.unit.Lib)
		}
	}
	if l.extStart != 0 {
		fmt.Println("extStart:", l.extStart)
	}
	fmt.Println("max:", l.max)
	fmt.Printf("syms(len=%d)\n", len(l.Syms))
	for i, s := range l.Syms {
		if i == 0 {
			continue
		}
		if s != nil {
			fmt.Println(i, s, s.Type)
		} else {
			fmt.Println(i, l.SymName(Sym(i)), "<not loaded>")
		}
	}
	fmt.Println("overwrite:", l.overwrite)
	fmt.Println("symsByName")
	for name, i := range l.symsByName[0] {
		fmt.Println(i, name, 0)
	}
	for name, i := range l.symsByName[1] {
		fmt.Println(i, name, 1)
	}
}
