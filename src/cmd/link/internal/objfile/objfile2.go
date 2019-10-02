package objfile

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"fmt"
	"log"
	"strings"
)

var _ = fmt.Print

// Load relocations for building the dependency graph in deadcode pass.
// For now, we load symbol types, relocations, gotype, and the contents
// of type symbols, which are needed in deadcode.
func LoadReloc(l *loader.Loader, r *obj.Reader, lib *sym.Library, localSymVersion int, libByPkg map[string]*sym.Library) {
	// PkgIdx
	pkglist := r.Pkglist()

	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."
	istart := l.StartIndex(r)

	resolveSymRef := func(s obj.OSymRef) *sym.Symbol {
		var rr *obj.Reader
		switch p := s.PkgIdx; p {
		case obj.PkgIdxInvalid:
			if s.SymIdx != 0 {
				panic("bad sym ref")
			}
			return nil
		case obj.PkgIdxNone:
			// Resolve by name
			i := int(s.SymIdx-1) + r.NSym()
			osym := obj.OSym{}
			osym.Read(r, r.SymOff(i))
			name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
			v := loader.AbiToVer(osym.ABI, localSymVersion)
			nv := loader.NameVer{name, v}
			i = l.SymsByName[nv]
			return l.Syms[i]
		case obj.PkgIdxSelf:
			rr = r
		default:
			if p < obj.PkgIdxRefBase {
				panic("bad sym ref")
			}
			pkg := pkglist[p-obj.PkgIdxRefBase]
			rr = libByPkg[pkg].Readers[0].Reader // typically Readers[0] is go object (others are asm)
		}
		i := l.ToGlobal(rr, int(s.SymIdx-1))
		return l.Syms[i]
	}

	for i, n := 0, r.NSym()+r.NNonpkgdef(); i < n; i++ {
		s := l.Syms[istart+i]
		if s == nil || s.Name == "" {
			continue
		}

		osym := obj.OSym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		if s.Name != name { // Sanity check. We can remove it in the final version.
			fmt.Println("name mismatch:", lib, i, s.Name, name)
			panic("name mismatch")
		}

		if s.Type != 0 && s.Type != sym.SXREF {
			fmt.Println("symbol already processed:", lib, i, s)
			panic("symbol already processed")
		}

		t := sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)]
		if t == sym.SXREF {
			log.Fatalf("bad sxref")
		}
		if t == 0 {
			log.Fatalf("missing type for %s in %s", s.Name, lib)
		}
		if t == sym.SBSS && (s.Type == sym.SRODATA || s.Type == sym.SNOPTRBSS) {
			t = s.Type
		}
		s.Type = t

		// Reloc
		nreloc := r.NReloc(i)
		s.R = make([]sym.Reloc, nreloc)
		for j := range s.R {
			rel := obj.OReloc{}
			rel.Read(r, r.RelocOff(i, j))
			s.R[j] = sym.Reloc{
				Off:  rel.Off,
				Siz:  rel.Siz,
				Type: objabi.RelocType(rel.Type),
				Add:  rel.Add,
				Sym:  resolveSymRef(rel.Sym),
			}
		}

		// XXX deadcode needs symbol data for type symbols. Read it now.
		if strings.HasPrefix(name, "type.") {
			s.P = r.BytesAt(r.DataOff(i), r.DataSize(i))
			s.Attr.Set(sym.AttrReadOnly, r.ReadOnly())
			s.Size = int64(osym.Siz)
		}

		// Aux symbol
		naux := r.NAux(i)
		for j := 0; j < naux; j++ {
			a := obj.OAux{}
			a.Read(r, r.AuxOff(i, j))
			switch a.Type {
			case obj.AuxGotype:
				typ := resolveSymRef(a.Sym)
				if typ != nil {
					s.Gotype = typ
				}
			case obj.AuxFuncdata:
				pc := s.FuncInfo
				if pc == nil {
					pc = &sym.FuncInfo{Funcdata: make([]*sym.Symbol, 0, 4)}
					s.FuncInfo = pc
				}
				pc.Funcdata = append(pc.Funcdata, resolveSymRef(a.Sym))
			}
		}

		if s.Type == sym.STEXT {
			dupok := osym.Flag&obj.SymFlagDupok != 0
			if !dupok {
				if s.Attr.OnList() {
					log.Fatalf("symbol %s listed multiple times", s.Name)
				}
				s.Attr |= sym.AttrOnList
				lib.Textp = append(lib.Textp, s)
			} else {
				// there may ba a dup in another package
				// put into a temp list and add to text later
				lib.DupTextSyms = append(lib.DupTextSyms, s)
			}
		}
	}
}

// Load full contents.
// TODO: For now, some contents are already load in LoadReloc. Maybe
// we should combine LoadReloc back into this, once we rewrite deadcode
// pass to use index directly.
func LoadFull(l *loader.Loader, r *obj.Reader, lib *sym.Library, localSymVersion int, libByPkg map[string]*sym.Library) {
	// PkgIdx
	pkglist := r.Pkglist()

	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."
	istart := l.StartIndex(r)

	resolveSymRef := func(s obj.OSymRef) *sym.Symbol {
		var rr *obj.Reader
		switch p := s.PkgIdx; p {
		case obj.PkgIdxInvalid:
			if s.SymIdx != 0 {
				panic("bad sym ref")
			}
			return nil
		case obj.PkgIdxNone:
			// Resolve by name
			i := int(s.SymIdx-1) + r.NSym()
			osym := obj.OSym{}
			osym.Read(r, r.SymOff(i))
			name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
			v := loader.AbiToVer(osym.ABI, localSymVersion)
			nv := loader.NameVer{name, v}
			i = l.SymsByName[nv]
			return l.Syms[i]
		case obj.PkgIdxSelf:
			rr = r
		default:
			if p < obj.PkgIdxRefBase {
				panic("bad sym ref")
			}
			pkg := pkglist[p-obj.PkgIdxRefBase]
			rr = libByPkg[pkg].Readers[0].Reader // typically Readers[0] is go object (others are asm)
		}
		i := l.ToGlobal(rr, int(s.SymIdx-1))
		return l.Syms[i]
	}

	pcdataBase := r.PcdataBase()
	for i, n := 0, r.NSym()+r.NNonpkgdef(); i < n; i++ {
		s := l.Syms[istart+i]
		if s == nil || s.Name == "" {
			continue
		}
		if !s.Attr.Reachable() && (s.Type < sym.SDWARFSECT || s.Type > sym.SDWARFLINES) {
			// No need to load unreachable symbols.
			// XXX DWARF symbols may be used but are not marked reachable.
			continue
		}

		osym := obj.OSym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		if s.Name != name { // Sanity check. We can remove it in the final version.
			fmt.Println("name mismatch:", lib, i, s.Name, name)
			panic("name mismatch")
		}

		dupok := osym.Flag&obj.SymFlagDupok != 0
		local := osym.Flag&obj.SymFlagLocal != 0
		makeTypelink := osym.Flag&obj.SymFlagTypelink != 0
		datasize := r.DataSize(i)
		size := osym.Siz

		// Symbol data
		s.P = r.BytesAt(r.DataOff(i), datasize)
		s.Attr.Set(sym.AttrReadOnly, r.ReadOnly())

		// Aux symbol info
		var isym uint32
		naux := r.NAux(i)
		for j := 0; j < naux; j++ {
			a := obj.OAux{}
			a.Read(r, r.AuxOff(i, j))
			switch a.Type {
			case obj.AuxGotype, obj.AuxFuncdata:
				// already loaded
			case obj.AuxFuncInfo:
				if a.Sym.PkgIdx != obj.PkgIdxSelf {
					panic("funcinfo symbol not defined in current package")
				}
				isym = a.Sym.SymIdx
			default:
				panic("unknown aux type")
			}
		}

		s.File = pkgprefix[:len(pkgprefix)-1]
		if dupok {
			s.Attr |= sym.AttrDuplicateOK
		}
		if s.Size < int64(size) {
			s.Size = int64(size)
		}
		s.Attr.Set(sym.AttrLocal, local)
		s.Attr.Set(sym.AttrMakeTypelink, makeTypelink)

		if s.Type != sym.STEXT {
			continue
		}

		// FuncInfo
		if isym == 0 {
			continue
		}
		b := r.BytesAt(r.DataOff(int(isym-1)), r.DataSize(int(isym-1)))
		info := obj.OFuncInfo{}
		info.Read(b)

		if info.NoSplit != 0 {
			s.Attr |= sym.AttrNoSplit
		}
		if info.Flags&obj.FuncFlagReflectMethod != 0 {
			s.Attr |= sym.AttrReflectMethod
		}
		if info.Flags&obj.FuncFlagShared != 0 {
			s.Attr |= sym.AttrShared
		}
		if info.Flags&obj.FuncFlagTopFrame != 0 {
			s.Attr |= sym.AttrTopFrame
		}

		info.Pcdata = append(info.Pcdata, info.PcdataEnd) // for the ease of knowing where it ends
		pc := s.FuncInfo
		if pc == nil {
			pc = &sym.FuncInfo{}
			s.FuncInfo = pc
		}
		pc.Args = int32(info.Args)
		pc.Locals = int32(info.Locals)
		pc.Pcdata = make([]sym.Pcdata, len(info.Pcdata)-1) // -1 as we appended one above
		pc.Funcdataoff = make([]int64, len(info.Funcdataoff))
		pc.File = make([]*sym.Symbol, len(info.File))
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
	}
}
