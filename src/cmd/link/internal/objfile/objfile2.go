package objfile

import (
	"cmd/internal/bio"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var _ = fmt.Print

// Preload a package: add autolibs, add symbols to the symbol table.
// Does not read symbol data yet.
func LoadNew(arch *sys.Arch, syms *sym.Symbols, f *bio.Reader, lib *sym.Library, length int64, pn string, flags int) {
	start := f.Offset()
	r := obj.NewReader(f.File(), uint32(start))
	localSymVersion := syms.IncVersion()
	lib.Readers = append(lib.Readers, struct {
		Reader  *obj.Reader
		Version int
	}{r, localSymVersion})

	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."

	// Autolib
	lib.ImportStrings = append(lib.ImportStrings, r.Autolib()...)

	ndef := r.NSym()
	if ndef > 0 {
		if lib.Syms != nil {
			panic("multiple objects have defined symbols in package " + lib.String())
		}
		lib.Syms = make([]*sym.Symbol, ndef+1)
		lib.Syms[0] = nil
	}

	for i, n := 0, r.NSym()+r.NNonpkgdef()+r.NNonpkgref(); i < n; i++ {
		osym := obj.OSym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
		s := syms.Lookup(name, abiToVer(osym.ABI, localSymVersion))
		preprocess(arch, s) // TODO: put this at a better place

		if i < ndef {
			lib.Syms[i+1] = s
		}
	}

	// The caller expects us consuming all the data
	f.MustSeek(length, os.SEEK_CUR)
}

func abiToVer(abi uint16, localSymVersion int) int {
	var v int
	if abi == ^uint16(0) {
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
	if s.Name[0] == '$' && len(s.Name) > 5 && s.Type == 0 && len(s.P) == 0 {
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
		s.Attr.Set(sym.AttrReachable, false)
	}
	if strings.HasPrefix(s.Name, "runtime.gcbits.") {
		s.Attr |= sym.AttrLocal
	}
}

func LoadFull(r *obj.Reader, lib *sym.Library, syms *sym.Symbols, localSymVersion int, libByPkg map[string]*sym.Library) {
	// PkgIdx
	pkglist := r.Pkglist()

	pkgprefix := objabi.PathToPrefix(lib.Pkg) + "."

	resolveSymRef := func(s obj.OSymRef) *sym.Symbol {
		switch p := s.PkgIdx; p {
		case obj.PkgIdxInvalid:
			if s.SymIdx != 0 {
				panic("bad sym ref")
			}
			return nil
		case obj.PkgIdxNone:
			i := int(s.SymIdx-1) + r.NSym()
			osym := obj.OSym{}
			osym.Read(r, r.SymOff(i))
			name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)
			return syms.Lookup(name, abiToVer(osym.ABI, localSymVersion))
		case obj.PkgIdxSelf:
			return lib.Syms[s.SymIdx]
		default:
			if p < obj.PkgIdxRefBase {
				panic("bad sym ref")
			}
			pkg := pkglist[p-obj.PkgIdxRefBase]
			return libByPkg[pkg].Syms[s.SymIdx]
		}
	}

	ndef := r.NSym()
	for i, n := 0, r.NSym()+r.NNonpkgdef(); i < n; i++ {
		osym := obj.OSym{}
		osym.Read(r, r.SymOff(i))
		name := strings.Replace(osym.Name, "\"\".", pkgprefix, -1)

		var s *sym.Symbol
		if i < ndef {
			s = lib.Syms[i+1]
			if s.Name != name {
				fmt.Println("XXX", lib, i, s.Name, name)
				panic("XXX")
			}
		} else {
			s = syms.Lookup(name, abiToVer(osym.ABI, localSymVersion))
		}

		// Symbol Info
		info := obj.OSymInfo{}
		info.Read(r, r.InfoOff(i), objabi.SymKind(osym.Type) == objabi.STEXT)

		dupok := osym.Flag&1 != 0
		local := osym.Flag&2 != 0
		makeTypelink := osym.Flag&4 != 0
		nreloc := r.NReloc(i)
		datasize := r.DataSize(i)
		size := info.Size
		typ := resolveSymRef(info.GoType)

		t := sym.AbiSymKindToSymKind[objabi.SymKind(osym.Type)]
		if s.Type != 0 && s.Type != sym.SXREF {
			if (t == sym.SDATA || t == sym.SBSS || t == sym.SNOPTRBSS) && datasize == 0 && nreloc == 0 {
				if s.Size < int64(size) {
					s.Size = int64(size)
				}
				if typ != nil && s.Gotype == nil {
					s.Gotype = typ
				}
				continue
			}

			if (s.Type == sym.SDATA || s.Type == sym.SBSS || s.Type == sym.SNOPTRBSS) && len(s.P) == 0 && len(s.R) == 0 {
				goto overwrite
			}
			if s.Type != sym.SBSS && s.Type != sym.SNOPTRBSS && !dupok && !s.Attr.DuplicateOK() {
				log.Fatalf("duplicate symbol %s (types %d and %d) in %s and %s", s.Name, s.Type, t, s.File, "<TODO>")
				continue
			}
			if len(s.P) > 0 {
				if s.Type == sym.STEXT {
					lib.DupTextSyms = append(lib.DupTextSyms, s)
				}
				continue // duplicated symbol, already read
			}
		}

	overwrite:
		// Symbol data
		s.P = r.BytesAt(r.DataOff(i), datasize)

		// Reloc
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

		s.File = pkgprefix[:len(pkgprefix)-1]
		s.Lib = lib
		if dupok {
			s.Attr |= sym.AttrDuplicateOK
		}
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
		if s.Size < int64(size) {
			s.Size = int64(size)
		}
		s.Attr.Set(sym.AttrLocal, local)
		s.Attr.Set(sym.AttrMakeTypelink, makeTypelink)
		if typ != nil {
			s.Gotype = typ
		}

		if s.Type != sym.STEXT {
			continue
		}

		if info.NoSplit != 0 {
			s.Attr |= sym.AttrNoSplit
		}
		if info.Flags&(1<<2) != 0 {
			s.Attr |= sym.AttrReflectMethod
		}
		if info.Flags&(1<<3) != 0 {
			s.Attr |= sym.AttrShared
		}
		if info.Flags&(1<<4) != 0 {
			s.Attr |= sym.AttrTopFrame
		}

		info.Pcdata = append(info.Pcdata, info.AuxDataEnd) // for the ease of knowing where it ends
		auxDataBase := r.AuxDataBase()
		pc := &sym.FuncInfo{
			Args:        int32(info.Args),
			Locals:      int32(info.Locals),
			Pcdata:      make([]sym.Pcdata, len(info.Pcdata)-1), // -1 as we appended one above
			Funcdata:    make([]*sym.Symbol, len(info.Funcdata)),
			Funcdataoff: make([]int64, len(info.Funcdataoff)),
			File:        make([]*sym.Symbol, len(info.File)),
		}
		s.FuncInfo = pc
		pc.Pcsp.P = r.BytesAt(auxDataBase+info.Pcsp, int(info.Pcfile-info.Pcsp))
		pc.Pcfile.P = r.BytesAt(auxDataBase+info.Pcfile, int(info.Pcline-info.Pcfile))
		pc.Pcline.P = r.BytesAt(auxDataBase+info.Pcline, int(info.Pcinline-info.Pcline))
		pc.Pcinline.P = r.BytesAt(auxDataBase+info.Pcinline, int(info.Pcdata[0]-info.Pcinline))
		for k := range pc.Pcdata {
			pc.Pcdata[k].P = r.BytesAt(auxDataBase+info.Pcdata[k], int(info.Pcdata[k+1]-info.Pcdata[k]))
		}
		for k := range pc.Funcdata {
			pc.Funcdata[k] = resolveSymRef(info.Funcdata[k])
			pc.Funcdataoff[k] = int64(info.Funcdataoff[k])
		}
		for k := range pc.File {
			pc.File[k] = syms.Lookup(info.File[k], 0) // TODO: why File is a symbol?
		}

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
