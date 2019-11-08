package objstats

import (
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"cmd/link/internal/sym"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sort"
	"unsafe"
)

type DupInfo struct {
	kind   sym.SymKind // symkind
	count  uint64      // number of dup syms for this symkind
	size   uint64      // total datasize
	relocs uint64      // total relocs
}

// Information/stats that deal with reading of host object.
type HostObjStats struct {
	// number of host objects read in.
	HostObjects uint64

	// number of symbols created from host objects.
	HostObjectSymbols uint64

	// relocations on host object symbols.
	HostObjectRelocs uint64

	// bytes of data from host object symbols.
	HostObjectDataBytes uint64
}

// Statistics on what's happening with the loader.
type LoaderStats struct {
	TotalSyms    uint64
	ExternalSyms uint64
	ExtObjRefs   uint64
	Builtins     uint64

	ToLocalCalls  uint64
	ToGlobalCalls uint64
}

type DieInfo struct {
	Count   uint64
	NAttrs  uint64
	SizeEst uint64
	Tag     int
}

type AttrInfo struct {
	Count   uint64
	SizeEst uint64
	Attr    uint16
}

type HashedAttrBucket struct {
	Attr  uint16
	Count uint64
	x     []HashedAttr
}

type HashedAttr struct {
	Attr    *dwarf.DWAttr // reference to a representative attr
	Count   uint64        // how many times we emitted this specific attr
	SizeEst uint64
}

// DWARF gen stats
type DwarfStats struct {
	Dies    uint64
	DieSize uint64
	DieMem  uint64

	Attrs    uint64
	AttrSize uint64
	AttrMem  uint64

	Strings     uint64
	StringBytes uint64

	// Maps from DIE tag to stats on dies with that tag
	DieTagMap map[int]DieInfo

	// Maps from attr code to stats on those attrs
	AttrMap map[uint16]AttrInfo

	// This hashes attrs for equality.
	HashedAttrs map[uint64]HashedAttrBucket
}

func (ds *DwarfStats) makeMaps() {
	if ds.DieTagMap == nil {
		ds.DieTagMap = make(map[int]DieInfo)
		ds.AttrMap = make(map[uint16]AttrInfo)
		ds.HashedAttrs = make(map[uint64]HashedAttrBucket)
	}
}

func (ds *DwarfStats) MkDie(abbrev int) {
	ds.makeMaps()
	dsz := uint64(unsafe.Sizeof(dwarf.DWDie{}))
	ds.Dies++
	ds.DieSize = dsz
	ds.DieMem += dsz
	tag := dwarf.AbbrevTag(abbrev)
	di := ds.DieTagMap[tag]
	di.Tag = tag
	di.Count += 1
	di.NAttrs += 1
	atsz := uint64(unsafe.Sizeof(dwarf.DWAttr{}))
	ds.AttrSize = atsz
	di.SizeEst += dsz + atsz
	ds.DieTagMap[tag] = di
}

func (ds *DwarfStats) MkAttr(at *dwarf.DWAttr, die *dwarf.DWDie) {
	ds.makeMaps()

	atsz := uint64(unsafe.Sizeof(dwarf.DWAttr{}))
	ds.Attrs++
	ds.AttrMem += atsz

	tag := dwarf.AbbrevTag(die.Abbrev)
	di := ds.DieTagMap[tag]
	di.Tag = tag
	di.NAttrs += 1
	ds.DieTagMap[tag] = di

	ai := ds.AttrMap[at.Atr]
	ai.Attr = at.Atr
	ai.Count++
	ds.AttrMap[at.Atr] = ai

	hc := HashDwAttr(at)
	bucket := ds.HashedAttrs[hc]
	bucket.Count += 1
	for i := range bucket.x {
		ha := &bucket.x[i]
		if DwAttrEq(ha.Attr, at) {
			// hit
			ha.Count++
			ha.SizeEst += atsz
			ds.HashedAttrs[hc] = bucket
			return
		}
	}
	bucket.x = append(bucket.x, HashedAttr{Attr: at, Count: 1, SizeEst: atsz})
	bucket.Attr = at.Atr
	ds.HashedAttrs[hc] = bucket
}

func DwAttrEq(at1 *dwarf.DWAttr, at2 *dwarf.DWAttr) bool {
	if at1.Atr != at2.Atr {
		return false
	}
	if at1.Cls != at2.Cls {
		return false
	}
	if at1.Value != at2.Value {
		return false
	}

	dn1 := at1.Data == nil
	dn2 := at2.Data == nil
	if dn1 != dn2 {
		return false
	}
	if dn1 {
		return true
	}

	s1, ok1 := at1.Data.(string)
	s2, ok2 := at2.Data.(string)
	if ok1 != ok2 {
		return false
	}
	if ok1 {
		return s1 == s2
	}
	i1, ok1i := at1.Data.(int)
	i2, ok2i := at2.Data.(int)
	if ok1i != ok2i {
		return false
	}
	if ok1i {
		return i1 == i2
	}
	wi1, ok1wi := at1.Data.(int64)
	wi2, ok2wi := at2.Data.(int64)
	if ok1wi != ok2wi {
		return false
	}
	if ok1wi {
		return wi1 == wi2
	}
	ds1, ok1d := at1.Data.(dwarf.Sym)
	ds2, ok2d := at2.Data.(dwarf.Sym)
	if ok1d != ok2d {
		return false
	}
	if ok1d {
		return ds1.Eq(ds2)
	}
	b1, ok1b := at1.Data.([]byte)
	b2, ok2b := at2.Data.([]byte)
	if ok1b != ok2b {
		return false
	}
	if ok1b {
		if len(b1) != len(b2) {
			return false
		}
		for i := range b1 {
			if b1[i] != b2[i] {
				return false
			}
		}
		return true
	}
	panic(fmt.Sprintf("unknown attr data contents: t:%s %+v t:%s %+v ", reflect.TypeOf(at1.Data).String(), at1.Data, reflect.TypeOf(at2.Data).String(), at2.Data))
}

func HashDwAttr(at *dwarf.DWAttr) uint64 {
	h := uint64(at.Cls)<<16 | uint64(at.Atr)
	h = h ^ uint64(at.Value)
	sp := fmt.Sprintf("%v", at.Data)
	for _, c := range sp {
		h = (h << 4) + uint64(c)
		high := h & uint64(0xF0000000000000)
		if high != 0 {
			h ^= high >> 48
			h &= high
		}
	}
	return h
}

// Information/stats that deal with Go object file reading. Note
// that these stats only cover symbols generated as a result of
// the Loader.Preload method calls -- other symbols added in later
// on (ex: DWARF, or host object reading) will turn up only
// in the loader stats.
type ObjStats struct {
	// number of objects read in
	Objects int

	// number of successful object mmaps
	Mmaps int

	// total number of symbols defined, referenced
	PkgDefSyms   int
	NoPkgDefSyms int
	NoPkgRefSyms int

	// count of defined symbols (total and duplicate). this includes
	// symbols later determined to be dead.
	Symbols    int
	DupSymbols int

	// aux symbol count
	AuxSymbols int

	// sizes and counts of duplicate symbols by kind
	Duptab map[sym.SymKind]DupInfo

	// count of object file relocations
	Relocs int
}

type SymStats struct {
	Syms *sym.Symbols

	// Stats from Go objects
	Os ObjStats

	// Stats from host objects
	Hostos HostObjStats

	// Loader related stats
	Loader LoaderStats

	// DWARF stats
	Dw DwarfStats

	// total symbols
	Symbols int
	// symbols with reachable attr set.
	ReachableSymbols int
	// dwarf symbols
	Dwsymbols int
	// dwarf symbol relocations
	Dwrelocs int
	// function symbols
	Fcns int
	// symbols with non-nil gotype
	HasGoType int
	// total relocation count overall
	Relocations int
	// total datasize (sum of lengths of all s.P slices)
	DataSize uint64

	// stats on a per-kind basis. intended to help tell which
	// flavors of symbol have more/fewer relocations or data.
	Stab map[sym.SymKind]SymKindStats

	// Symbol counts by version.
	Vtab map[int16]SymVerStats

	// Symbol counts by alignment.
	Atab map[int32]SymAlignStats

	// relocation oracle
	relocOracle relocOracle

	// if true, analyze duplicate relocs
	Deepreloc bool
}

// Stats on symbols with a specific alignment.
type SymAlignStats struct {
	Count uint64
	Align int32
}

// Stats on symbols with a specific version. This is here mainly to see
// what fraction of symbols are "regular" (non-static), "abi internal" (V1)
// vs local.
type SymVerStats struct {
	Version int16
	Count   uint64
}

// For stats on lookups (which symbols actually get looked up the most).
type SymLookupStats struct {
	Name    string
	Version int
	Count   uint64
}

// Sym stats for a given kind of symbol. Include breakdown of
// stats by relocation, also size and count.
type SymKindStats struct {
	Rtab  map[objabi.RelocType]RelocStats
	Size  uint64
	Count int
	Kind  sym.SymKind
}

// Stats for relocations.
type RelocStats struct {
	Rtype      objabi.RelocType
	Count      int
	Hasaddend  int
	Hasxsym    int
	Hasvariant int
}

func (skstats *SymKindStats) RecordRelocs(sp *sym.Symbol, ro *relocOracle) {
	if skstats.Rtab == nil {
		skstats.Rtab = make(map[objabi.RelocType]RelocStats)
	}
	for _, rp := range sp.R {
		rv := skstats.Rtab[rp.Type]
		rv.Rtype = rp.Type
		rv.Count++
		if rp.Add != 0 {
			rv.Hasaddend++
		}
		if rp.HasExt() && rp.Xsym != nil {
			rv.Hasxsym++
		}
		if rp.HasExt() && rp.Variant != 0 {
			rv.Hasvariant++
		}
		skstats.Rtab[rp.Type] = rv
		if ro != nil {
			ro.Lookup(&rp, sp.Type)
		}
	}
}

func (skstats *SymKindStats) NumRelocs() int {
	tot := 0
	for _, v := range skstats.Rtab {
		tot += v.Count
	}
	return tot
}

func (stats *SymStats) RecordSym(sp *sym.Symbol, syms *sym.Symbols) {

	// Update stab first
	if stats.Stab == nil {
		stats.Stab = make(map[sym.SymKind]SymKindStats)
		if stats.Deepreloc {
			stats.relocOracle.relhash = make(map[sym.Reloc]uint64)
			stats.relocOracle.srckindmap = make(map[sym.SymKind]symKindHitMiss)
			stats.relocOracle.tgtkindmap = make(map[sym.SymKind]symKindHitMiss)
		}
		stats.Syms = syms
	}
	if stats.Vtab == nil {
		stats.Vtab = make(map[int16]SymVerStats)
	}
	if stats.Atab == nil {
		stats.Atab = make(map[int32]SymAlignStats)
	}
	sv := stats.Stab[sp.Type]
	sv.Kind = sp.Type
	sv.Count++
	sv.Size += uint64(sp.Size)
	var ro *relocOracle
	if stats.Deepreloc {
		ro = &stats.relocOracle
	}
	sv.RecordRelocs(sp, ro)
	stats.Stab[sp.Type] = sv

	// Now top-level stats
	stats.Symbols++
	vs := stats.Vtab[sp.Version]
	vs.Version = sp.Version
	vs.Count += 1
	stats.Vtab[sp.Version] = vs
	as := stats.Atab[sp.Align]
	as.Align = sp.Align
	as.Count += 1
	stats.Atab[sp.Align] = as
	if sp.Attr.Reachable() {
		stats.ReachableSymbols++
	}
	if sp.Type == sym.STEXT {
		stats.Fcns++
	}
	if sp.Gotype != nil {
		stats.HasGoType++
	}
	stats.DataSize += uint64(len(sp.P))
	stats.Relocations += len(sp.R)
	if sp.Type == sym.SDWARFSECT || sp.Type == sym.SDWARFINFO ||
		sp.Type == sym.SDWARFRANGE || sp.Type == sym.SDWARFLOC {
		stats.Dwsymbols++
		stats.Dwrelocs += len(sp.R)
	}
}

func (sks *SymKindStats) AccumulateRelocs(other SymKindStats) {
	for _, ovr := range other.Rtab {
		svr := sks.Rtab[ovr.Rtype]
		svr.Rtype = ovr.Rtype
		svr.Count += ovr.Count
		svr.Hasxsym += ovr.Hasxsym
		svr.Hasaddend += ovr.Hasaddend
		svr.Hasvariant += ovr.Hasvariant
		if sks.Rtab == nil {
			sks.Rtab = make(map[objabi.RelocType]RelocStats)
		}
		sks.Rtab[ovr.Rtype] = svr
	}
}

func (s *SymStats) RecordPreload(mmaps int, pkgDefSyms int, noPkgDefSyms int, noPkgRefSyms int, nAux int) {
	s.Os.Objects++
	s.Os.Mmaps += mmaps
	s.Os.Symbols += pkgDefSyms + noPkgDefSyms
	s.Os.AuxSymbols += nAux
	s.Os.PkgDefSyms += pkgDefSyms
	s.Os.NoPkgDefSyms += noPkgDefSyms
	s.Os.NoPkgRefSyms += noPkgRefSyms
}

func (ro *relocOracle) DumpTop10(wf io.Writer) {
	// Hack: hijack the Xadd field to hold count
	top10 := [10]sym.Reloc{}
	for r, c := range ro.relhash {
		// find minimum and replace with r
		top10[0].InitExt()
		mc := top10[0].Xadd
		mi := 0
		for idx := 1; idx < 10; idx++ {
			top10[idx].InitExt()
			if top10[idx].Xadd < mc {
				mi = idx
			}
		}
		if top10[mi].Xadd < int64(c) {
			r.InitExt()
			r.Xadd = int64(c)
			top10[mi] = r
		}
	}
	fmt.Fprintf(wf, "  top 10 relocs by hit count:\n")
	fmt.Fprintf(wf, "     %15s: %s\n", "Count", "Target")
	for idx := 0; idx < 10; idx++ {
		c := top10[idx].Xadd
		top10[idx].Xadd = 0
		fmt.Fprintf(wf, "     %15d: %v\n", c, top10[idx])
	}
}

func (ro *relocOracle) DumpHitMissByKind(wf io.Writer, km map[sym.SymKind]symKindHitMiss, tag string) {
	sl := []symKindHitMiss{}
	for _, v := range km {
		sl = append(sl, v)
	}
	sort.Sort(bySymKindHitMiss(sl))
	first := true
	for _, hm := range sl {
		if first {
			fmt.Fprintf(wf, "    reloc hit/miss breakdown by %s SymKind:\n", tag)
			fmt.Fprintf(wf, "       %15s: %11s %11s %11s\n", "Kind", "Count", "Hit", "Miss")
			first = false
		}
		fmt.Fprintf(wf, "        %15s: %11d %11d %11d\n",
			hm.kind.String(), hm.hit+hm.miss, hm.hit, hm.miss)
	}
}

func (ro *relocOracle) Dump(wf io.Writer) {
	if ro.miss != 0 {
		fmt.Fprintf(wf, "  reloc oracle:\n")
		fmt.Fprintf(wf, "    hits=%d\n", ro.hit)
		fmt.Fprintf(wf, "    misses=%d\n", ro.miss)
		fmt.Fprintf(wf, "    niltarg=%d\n", ro.niltarg)
		ro.DumpHitMissByKind(wf, ro.srckindmap, "source")
		ro.DumpHitMissByKind(wf, ro.tgtkindmap, "target")
		ro.DumpTop10(wf)
	}
}

func (stats *SymStats) RecordHostObject(hosyms []*sym.Symbol) {
	stats.Hostos.HostObjects += 1
	stats.Hostos.HostObjectSymbols += uint64(len(hosyms))
	for _, s := range hosyms {
		stats.Hostos.HostObjectDataBytes += uint64(len(s.P))
		stats.Hostos.HostObjectRelocs += uint64(len(s.R))
	}
}

func (os *ObjStats) RecordDupSym(k sym.SymKind, nrelocs int, dataSize int) {
	os.DupSymbols += 1
	if os.Duptab == nil {
		os.Duptab = make(map[sym.SymKind]DupInfo)
	}
	di := os.Duptab[k]
	di.kind = k
	di.count += 1
	di.size += uint64(dataSize)
	di.relocs += uint64(nrelocs)
	os.Duptab[k] = di
}

func (hos *HostObjStats) Dump(wf io.Writer) {
	if hos.HostObjects != 0 {
		fmt.Fprintf(wf, "  host objects: %d\n", hos.HostObjects)
		fmt.Fprintf(wf, "  symbols from host objects: %d\n",
			hos.HostObjectSymbols)
		fmt.Fprintf(wf, "  host objects symbol relocations: %d\n",
			hos.HostObjectRelocs)
		fmt.Fprintf(wf, "  host objects symbol data size: %d\n",
			hos.HostObjectDataBytes)
	}
}

func AtToStr(at uint16) string {
	switch at {
	case dwarf.DW_AT_count:
		return "AT_count"
	case dwarf.DW_AT_external:
		return "AT_external"
	case dwarf.DW_AT_location:
		return "AT_Location"
	case dwarf.DW_AT_name:
		return "AT_name"
	case dwarf.DW_AT_byte_size:
		return "AT_byte_size"
	case dwarf.DW_AT_low_pc:
		return "AT_low_pc"
	case dwarf.DW_AT_bit_size:
		return "AT_bit_size"
	case dwarf.DW_AT_type:
		return "AT_type"
	case dwarf.DW_AT_data_member_location:
		return "AT_data_member_location"
	case dwarf.DW_AT_go_kind:
		return "AT_go_kind"
	case dwarf.DW_AT_go_runtime_type:
		return "AT_go_runtime_type"
	case dwarf.DW_AT_comp_dir:
		return "AT_comp_dir"
	case dwarf.DW_AT_go_package_name:
		return "AT_go_package_name"
	case dwarf.DW_AT_producer:
		return "AT_producer"
	case dwarf.DW_AT_go_embedded_field:
		return "AT_go_embedded_field "
	}
	return fmt.Sprintf("attr %x", at)
}

func TagToStr(tag int) string {
	switch tag {
	case dwarf.DW_TAG_unspecified_type:
		return "UnspecifiedType"
	case dwarf.DW_TAG_subroutine_type:
		return "SubroutineType"
	case dwarf.DW_TAG_base_type:
		return "BaseType"
	case dwarf.DW_TAG_formal_parameter:
		return "FormalParam"
	case dwarf.DW_TAG_subrange_type:
		return "SubrangeType"
	case dwarf.DW_TAG_array_type:
		return "ArrayType"
	case dwarf.DW_TAG_structure_type:
		return "StructType"
	case dwarf.DW_TAG_compile_unit:
		return "CompileUnit"
	case dwarf.DW_TAG_variable:
		return "Variable"
	case dwarf.DW_TAG_member:
		return "Member"
	case dwarf.DW_TAG_pointer_type:
		return "PointerType"
	case dwarf.DW_TAG_typedef:
		return "Typedef"
	}
	return fmt.Sprintf("tag %d", tag)
}

type byDieInfo []DieInfo

func (s byDieInfo) Len() int           { return len(s) }
func (s byDieInfo) Less(i, j int) bool { return s[j].Count < s[i].Count }
func (s byDieInfo) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type byAttrInfo []AttrInfo

func (s byAttrInfo) Len() int           { return len(s) }
func (s byAttrInfo) Less(i, j int) bool { return s[j].Count < s[i].Count }
func (s byAttrInfo) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (ds *DwarfStats) Dump(wf io.Writer) {
	fmt.Fprintf(wf, "  DWARF stats:\n")
	fmt.Fprintf(wf, "    Dies: %d\n", ds.Dies)
	fmt.Fprintf(wf, "    DieSize: %d\n", ds.DieSize)
	fmt.Fprintf(wf, "    DieMem: %d\n", ds.DieMem)
	fmt.Fprintf(wf, "    Attrs: %d\n", ds.Attrs)
	fmt.Fprintf(wf, "    AttrSize: %d\n", ds.AttrSize)
	fmt.Fprintf(wf, "    AttrMem: %d\n", ds.AttrMem)
	sl := []DieInfo{}
	for _, d := range ds.DieTagMap {
		sl = append(sl, d)
	}
	sort.Sort(byDieInfo(sl))
	fmt.Fprintf(wf, "    Die breakdown:\n")
	for _, di := range sl {
		fmt.Fprintf(wf, "      %15s : %10d NAttrs:%8d Sz:%10d\n",
			TagToStr(di.Tag), di.Count, di.NAttrs, di.SizeEst)
	}
	fmt.Fprintf(wf, "    Attr breakdown:\n")
	repeats := make(map[uint16]uint64) // num repeats of this attr
	totals := make(map[uint16]uint64)  // total number of this attr
	for _, b := range ds.HashedAttrs {
		atr := b.Attr
		totals[atr] += b.Count
		for _, at := range b.x {
			if at.Count > 1 {
				repeats[b.Attr] += at.Count - 1
			}
		}
	}
	// Hijacking AttrInfo here; Count is # total and size is repeats
	sl2 := []AttrInfo{}
	for atr, count := range totals {
		sl2 = append(sl2, AttrInfo{Count: count, Attr: atr, SizeEst: repeats[atr]})
	}
	sort.Sort(byAttrInfo(sl2))
	at := uint64(0)
	rt := uint64(0)
	for _, b := range sl2 {
		fmt.Fprintf(wf, "      %30s : %10d total, %10d repeated\n",
			AtToStr(b.Attr), b.Count, b.SizeEst)
		at += b.Count
		rt += b.SizeEst
	}
	fmt.Fprintf(wf, "      %30s : %10d total, %10d repeated\n",
		"<total>", at, rt)
}

func (ls *LoaderStats) Dump(wf io.Writer) {
	if ls.ToLocalCalls == 0 {
		return
	}
	fmt.Fprintf(wf, "  Loader stats:\n")
	if ls.TotalSyms != 0 {
		fmt.Fprintf(wf, "    TotalSyms: %d\n", ls.TotalSyms)
		fmt.Fprintf(wf, "    ExternalSyms: %d\n", ls.ExternalSyms)
		fmt.Fprintf(wf, "    Builtins: %d\n", ls.Builtins)
		fmt.Fprintf(wf, "    ExtObjRefs: %d\n", ls.ExtObjRefs)
		fmt.Fprintf(wf, "    ToLocalCalls: %d\n", ls.ToLocalCalls)
		fmt.Fprintf(wf, "    ToGlobalCalls: %d\n", ls.ToGlobalCalls)
	}
}

func (s *ObjStats) Dump(wf io.Writer) {
	fmt.Fprintf(wf, "  objects: %d\n", s.Objects)
	fmt.Fprintf(wf, "    Mmaps: %d\n", s.Mmaps)
	fmt.Fprintf(wf, "    AuxSymbols: %d\n", s.AuxSymbols)
	fmt.Fprintf(wf, "    PkgDefSyms: %d\n", s.PkgDefSyms)
	fmt.Fprintf(wf, "    NoPkgDefSyms: %d\n", s.NoPkgDefSyms)
	fmt.Fprintf(wf, "    NoPkgRefSyms: %d\n", s.NoPkgRefSyms)
	fmt.Fprintf(wf, "    symbols: %d\n", s.Symbols)
	fmt.Fprintf(wf, "    dupSymbols: %d\n", s.DupSymbols)
	if s.Duptab != nil {
		sl := []DupInfo{}
		for _, v := range s.Duptab {
			sl = append(sl, v)
		}
		sort.Sort(byDupInfo(sl))
		first := true
		for _, ks := range sl {
			if ks.count == 0 {
				continue
			}
			if first {
				fmt.Fprintf(wf, "    duplicate symbol breakdown by SymKind:\n")
				fmt.Fprintf(wf, "     %15s: %11s %11s %11s\n", "Kind", "Count", "Relocs", "DataSize")
				first = false
			}
			fmt.Fprintf(wf, "      %15s: %11d %11d %11d\n",
				ks.kind.String(), ks.count, ks.relocs, ks.size)
		}
	}
}

func (s *SymStats) Dump(tag string, wf io.Writer) {
	fmt.Fprintf(wf, "\nSymStats at '%s':\n", tag)
	s.Loader.Dump(wf)
	if s.Os.Objects != 0 {
		s.Os.Dump(wf)
		s.Hostos.Dump(wf)
	}
	if s.Dw.Dies != 0 {
		s.Dw.Dump(wf)
	}
	if tag == "final" && s.Vtab != nil {
		vvtab := make(map[int16]SymVerStats)
		for _, e := range s.Vtab {
			vn := e.Version
			if vn != 0 && vn != sym.SymVerABIInternal {
				vn = 2
			}
			vve := vvtab[vn]
			vve.Version = vn
			vve.Count += e.Count
			vvtab[vn] = vve
		}
		sl := []SymVerStats{}
		for _, e := range vvtab {
			sl = append(sl, e)
		}
		sort.Sort(bySymVerStats(sl))
		first := true
		which := []string{"global", "abi-internal", "local"}
		for _, ks := range sl {
			if ks.Count == 0 {
				continue
			}
			if first {
				fmt.Fprintf(wf, "  symbol breakdown by version:\n")
				fmt.Fprintf(wf, "  %15s: %11s\n", "Version", "Count")
				first = false
			}
			fmt.Fprintf(wf, "  %15s: %11d\n", which[ks.Version], ks.Count)
		}
		// begin
		for _, e := range s.Vtab {
			vn := e.Version
			if vn != 0 && vn != sym.SymVerABIInternal {
				vn = 2
			}
			vve := vvtab[vn]
			vve.Version = vn
			vve.Count += e.Count
			vvtab[vn] = vve
		}
		sal := []SymAlignStats{}
		for _, e := range s.Atab {
			sal = append(sal, e)
		}
		sort.Sort(bySymAlignStats(sal))
		first = true
		for _, ks := range sal {
			if ks.Count == 0 {
				continue
			}
			if first {
				fmt.Fprintf(wf, "  symbol breakdown by alignment:\n")
				fmt.Fprintf(wf, "  %15s: %11s\n", "Align", "Count")
				first = false
			}
			fmt.Fprintf(wf, "  %15d: %11d\n", ks.Align, ks.Count)
		}
		fmt.Fprintf(wf, "  global RW lookups: %d\n", s.Syms.Globs.Lookup)
		fmt.Fprintf(wf, "  global RO lookups: %d\n", s.Syms.Globs.ROLookup)
		fmt.Fprintf(wf, "  ABI-I RW lookups: %d\n", s.Syms.ABIIs.Lookup)
		fmt.Fprintf(wf, "  ABI-I RO lookups: %d\n", s.Syms.ABIIs.ROLookup)
		fmt.Fprintf(wf, "  local RW lookups: %d\n", s.Syms.Locals.Lookup)
		fmt.Fprintf(wf, "  local RO lookups: %d\n", s.Syms.Locals.ROLookup)
		if s.Syms.Unique != nil {
			fmt.Fprintf(wf, "  total unique lookups: %d\n", len(s.Syms.Unique))
			sl := []SymLookupStats{}
			for k, v := range s.Syms.Unique {
				e := SymLookupStats{Name: k.Name, Version: k.Ver, Count: uint64(v)}
				sl = append(sl, e)
			}
			sort.Sort(bySymLookupStats(sl))
			first := true
			howmany := 0
			for _, ks := range sl {
				if first {
					fmt.Fprintf(wf, "  top 25 lookup counts by symbol name:\n")
					fmt.Fprintf(wf, "  %11s : %s\n", "Count", "Name:Version")
					first = false
				}
				fmt.Fprintf(wf, "  %11d : %s[%d]\n", ks.Count, ks.Name, ks.Version)
				howmany++
				if howmany > 25 {
					break
				}
			}
		}
		if s.Syms.LookupSites != nil {
			fmt.Fprintf(wf, "  total dynamic lookup callsites: %d\n", len(s.Syms.LookupSites))
			if s.Syms.LSCollide != 0 {
				fmt.Fprintf(wf, "     [ ** collisions=%d]\n", s.Syms.LSCollide)
			}
			sl := []sym.LookupStack{}
			for _, v := range s.Syms.LookupSites {
				sl = append(sl, v)
			}
			sort.Sort(bySymLookupStack(sl))
			first := true
			howmany := 0
			for _, ks := range sl {
				if first {
					fmt.Fprintf(wf, "  top 25 lookup callsites by count:\n")
					fmt.Fprintf(wf, "  %11s : %s\n", "Count", "Frame")
					first = false
				}
				fmt.Fprintf(wf, "  %11d : ", ks.Count)
				frames := runtime.CallersFrames(ks.Pcs)
				ffirst := true
				for {
					frame, more := frames.Next()
					if ffirst {
						fmt.Fprintf(wf, "%s\n", frame.Function)
						ffirst = false
					} else {
						fmt.Fprintf(wf, "  %11s   %s\n", "", frame.Function)
					}
					if !more {
						break
					}
					if frame.Function == "runtime.main" || frame.Function == "main.main" {
						break
					}
				}
				fmt.Fprintf(wf, "\n")
				howmany++
				if howmany > 25 {
					break
				}
			}
		}

	}
	if s.Symbols != 0 {
		fmt.Fprintf(wf, "  symbols defined: %d\n", s.Symbols)
		fmt.Fprintf(wf, "  reachable symbols: %d\n", s.ReachableSymbols)
		usss := unsafe.Sizeof(sym.Symbol{})
		fmt.Fprintf(wf, "  unsafe.Sizeof(sym.Symbol{}) = %d\n", usss)
		fmt.Fprintf(wf, "  total symbolsize: %d\n", s.Symbols*int(usss))
		fmt.Fprintf(wf, "  reachable symbolsize: %d\n", s.ReachableSymbols*int(usss))
		fmt.Fprintf(wf, "  total data size: %d\n", s.DataSize)
		fmt.Fprintf(wf, "  relocations: %d\n", s.Relocations)
		usrs := unsafe.Sizeof(sym.Reloc{})
		fmt.Fprintf(wf, "  unsafe.Sizeof(sym.Reloc{}) = %d\n", usrs)
		fmt.Fprintf(wf, "  total relocsize: %d\n", s.Relocations*int(usrs))
		if s.relocOracle.miss != 0 {
			s.relocOracle.Dump(wf)
		}
		fmt.Fprintf(wf, "  fcns: %d\n", s.Fcns)
		fmt.Fprintf(wf, "  hasGoType: %d\n", s.HasGoType)
		fmt.Fprintf(wf, "  DWARF symbols defined: %d\n", s.Dwsymbols)
		fmt.Fprintf(wf, "  DWARF sym relocs: %d\n", s.Dwrelocs)

		if s.Stab != nil {
			sl := []SymKindStats{}
			for _, v := range s.Stab {
				sl = append(sl, v)
			}
			sort.Sort(bySymKindStats(sl))
			first := true
			accumRelocs := SymKindStats{}
			totalDataSize := uint64(0)
			for _, ks := range sl {
				if ks.Count == 0 {
					continue
				}
				if first {
					fmt.Fprintf(wf, "  symbol breakdown by SymKind:\n")
					fmt.Fprintf(wf, "     %15s: %11s %11s %11s\n", "Kind", "Count", "Relocs", "DataSize")
					first = false
				}
				nr := ks.NumRelocs()
				fmt.Fprintf(wf, "      %15s: %11d %11d %11d\n",
					ks.Kind.String(), ks.Count, nr, ks.Size)
				accumRelocs.AccumulateRelocs(ks)
				totalDataSize += uint64(ks.Size)
			}
			fmt.Fprintf(wf, "  total data size: %d\n", totalDataSize)

			rl := []RelocStats{}
			for _, v := range accumRelocs.Rtab {
				rl = append(rl, v)
			}
			sort.Sort(byRelocStats(rl))
			first = true
			tcount := 0
			thasxsym := 0
			thasaddend := 0
			thasvariant := 0
			for _, rv := range rl {
				if rv.Count == 0 {
					continue
				}
				if first {
					fmt.Fprintf(wf, "  relocation breakdown by type:\n")
					fmt.Fprintf(wf, "     %15s: %11s %11s %11s %11s\n", "Type", "Count", "Hasxsym", "Hasvariant", "Hasaddend")
					first = false
				}
				fmt.Fprintf(wf, "     %15s: %11d %11d %11d %11d\n",
					rv.Rtype.String(), rv.Count, rv.Hasxsym, rv.Hasvariant, rv.Hasaddend)
				tcount += rv.Count
				thasxsym += rv.Hasxsym
				thasaddend += rv.Hasaddend
				thasvariant += rv.Hasvariant
			}
			fmt.Fprintf(wf, "     %15s: %11d %11d %11d %11d\n",
				"<TOTAL>", tcount, thasxsym, thasvariant, thasaddend)
		}
	}
}

type byDupInfo []DupInfo

func (s byDupInfo) Len() int           { return len(s) }
func (s byDupInfo) Less(i, j int) bool { return s[i].kind < s[j].kind }
func (s byDupInfo) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymKindStats []SymKindStats

func (s bySymKindStats) Len() int           { return len(s) }
func (s bySymKindStats) Less(i, j int) bool { return s[i].Kind < s[j].Kind }
func (s bySymKindStats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymKindHitMiss []symKindHitMiss

func (s bySymKindHitMiss) Len() int           { return len(s) }
func (s bySymKindHitMiss) Less(i, j int) bool { return s[i].kind < s[j].kind }
func (s bySymKindHitMiss) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymVerStats []SymVerStats

func (s bySymVerStats) Len() int           { return len(s) }
func (s bySymVerStats) Less(i, j int) bool { return s[i].Count > s[j].Count }
func (s bySymVerStats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymAlignStats []SymAlignStats

func (s bySymAlignStats) Len() int           { return len(s) }
func (s bySymAlignStats) Less(i, j int) bool { return s[i].Count > s[j].Count }
func (s bySymAlignStats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymLookupStats []SymLookupStats

func (s bySymLookupStats) Len() int           { return len(s) }
func (s bySymLookupStats) Less(i, j int) bool { return s[i].Count > s[j].Count }
func (s bySymLookupStats) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type bySymLookupStack []sym.LookupStack

func (s bySymLookupStack) Len() int           { return len(s) }
func (s bySymLookupStack) Less(i, j int) bool { return s[i].Count > s[j].Count }
func (s bySymLookupStack) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type byRelocStats []RelocStats

func (rs byRelocStats) Len() int           { return len(rs) }
func (rs byRelocStats) Less(i, j int) bool { return rs[i].Rtype < rs[j].Rtype }
func (rs byRelocStats) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

type symKindHitMiss struct {
	kind sym.SymKind
	hit  uint64
	miss uint64
}

type relocOracle struct {
	relhash    map[sym.Reloc]uint64
	srckindmap map[sym.SymKind]symKindHitMiss
	tgtkindmap map[sym.SymKind]symKindHitMiss
	hit        uint64
	miss       uint64
	niltarg    uint64
}

func (ro *relocOracle) Lookup(r *sym.Reloc, kind sym.SymKind) {

	// Determine hit/miss and update hitcount for reloc
	hit := false
	if val, ok := ro.relhash[*r]; ok {
		hit = true
		ro.relhash[*r] = val + 1
		ro.hit++
	} else {
		ro.relhash[*r] = 1
		ro.miss++
	}

	// Record hit/miss by source symbol kind
	sk := ro.srckindmap[kind]
	sk.kind = kind
	if hit {
		sk.hit++
	} else {
		sk.miss++
	}
	ro.srckindmap[kind] = sk

	// Record hit/miss by target symbol kind
	if r.Sym == nil {
		ro.niltarg++
	} else {
		tk := ro.tgtkindmap[r.Sym.Type]
		tk.kind = r.Sym.Type
		if hit {
			tk.hit++
		} else {
			tk.miss++
		}
		ro.tgtkindmap[r.Sym.Type] = tk
	}
}
