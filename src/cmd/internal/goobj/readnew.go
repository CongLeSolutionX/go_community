// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goobj

import (
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"fmt"
)

// Read object file in new format. For now we still fill
// the data to the current goobj API.
func (r *objReader) readNew() {
	start := uint32(r.offset)
	rr := obj.NewReader(r.f, start)
	if rr == nil {
		panic("cannot read object file")
	}

	// Imports
	pkglist := rr.Pkglist()
	r.p.Imports = pkglist

	abiToVer := func(abi uint16) int64 {
		var vers int64
		if abi == ^uint16(0) {
			// Static symbol
			vers = r.p.MaxVersion
		}
		return vers
	}

	resolveSymRef := func(s obj.OSymRef) SymID {
		var i int
		switch p := s.PkgIdx; p {
		case obj.PkgIdxInvalid:
			if s.SymIdx != 0 {
				panic("bad sym ref")
			}
			return SymID{}
		case obj.PkgIdxNone:
			i = int(s.SymIdx-1) + rr.NSym()
		case obj.PkgIdxSelf:
			i = int(s.SymIdx - 1)
		default:
			if p < obj.PkgIdxRefBase {
				panic("bad sym ref")
			}
			pkg := pkglist[p-obj.PkgIdxRefBase]
			return SymID{fmt.Sprintf("%s.<#%d>", pkg, s.SymIdx), 0}
		}
		sym := obj.OSym{}
		sym.Read(rr, rr.SymOff(i))
		return SymID{sym.Name, abiToVer(sym.ABI)}
	}

	// Read things for the current goobj API for now.

	// Symbols
	pcdataBase := start + rr.PcdataBase()
	for i, n, ndef := 0, rr.NSym()+rr.NNonpkgdef()+rr.NNonpkgref(), rr.NSym()+rr.NNonpkgdef(); i < n; i++ {
		osym := obj.OSym{}
		osym.Read(rr, rr.SymOff(i))
		if osym.Name == "" {
			continue // not a real symbol
		}
		symID := SymID{Name: osym.Name, Version: abiToVer(osym.ABI)}
		r.p.SymRefs = append(r.p.SymRefs, symID)

		if i >= ndef {
			continue // not a defined symbol from here
		}

		// Symbol data
		dataOff := rr.DataOff(i)
		siz := int64(rr.DataSize(i))

		sym := Sym{
			SymID: symID,
			Kind:  objabi.SymKind(osym.Type),
			DupOK: osym.Flag&1 != 0,
			Size:  int64(osym.Siz),
			Data:  Data{int64(start + dataOff), siz},
		}
		r.p.Syms = append(r.p.Syms, &sym)

		// Reloc
		nreloc := rr.NReloc(i)
		sym.Reloc = make([]Reloc, nreloc)
		for j := 0; j < nreloc; j++ {
			rel := obj.OReloc{}
			rel.Read(rr, rr.RelocOff(i, j))
			sym.Reloc[j] = Reloc{
				Offset: int64(rel.Off),
				Size:   int64(rel.Siz),
				Type:   objabi.RelocType(rel.Type),
				Add:    rel.Add,
				Sym:    resolveSymRef(rel.Sym),
			}
		}

		// Aux symbol info
		var isym uint32
		funcdata := make([]obj.OSymRef, 0, 4)
		naux := rr.NAux(i)
		for j := 0; j < naux; j++ {
			a := obj.OAux{}
			a.Read(rr, rr.AuxOff(i, j))
			switch a.Type {
			case obj.AuxGotype:
				sym.Type = resolveSymRef(a.Sym)
			case obj.AuxFuncInfo:
				if a.Sym.PkgIdx != obj.PkgIdxSelf {
					panic("funcinfo symbol not defined in current package")
				}
				isym = a.Sym.SymIdx
			case obj.AuxFuncdata:
				funcdata = append(funcdata, a.Sym)
			default:
				panic("unknown aux type")
			}
		}

		// Symbol Info
		if isym == 0 {
			continue
		}
		b := rr.BytesAt(rr.DataOff(int(isym-1)), rr.DataSize(int(isym-1)))
		info := obj.OFuncInfo{}
		info.Read(b)

		if objabi.SymKind(osym.Type) == objabi.STEXT {
			info.Pcdata = append(info.Pcdata, info.PcdataEnd) // for the ease of knowing where it ends
			f := &Func{
				Args:     int64(info.Args),
				Frame:    int64(info.Locals),
				NoSplit:  info.NoSplit != 0,
				Leaf:     info.Flags&(1<<0) != 0,
				TopFrame: info.Flags&(1<<4) != 0,
				PCSP:     Data{int64(pcdataBase + info.Pcsp), int64(info.Pcfile - info.Pcsp)},
				PCFile:   Data{int64(pcdataBase + info.Pcfile), int64(info.Pcline - info.Pcfile)},
				PCLine:   Data{int64(pcdataBase + info.Pcline), int64(info.Pcinline - info.Pcline)},
				PCInline: Data{int64(pcdataBase + info.Pcinline), int64(info.Pcdata[0] - info.Pcinline)},
				PCData:   make([]Data, len(info.Pcdata)-1), // -1 as we appended one above
				FuncData: make([]FuncData, len(info.Funcdataoff)),
				File:     make([]string, len(info.File)),
			}
			sym.Func = f
			for k := range f.PCData {
				f.PCData[k] = Data{int64(pcdataBase + info.Pcdata[k]), int64(info.Pcdata[k+1] - info.Pcdata[k])}
			}
			for k := range f.FuncData {
				symID := resolveSymRef(funcdata[k])
				f.FuncData[k] = FuncData{symID, int64(info.Funcdataoff[k])}
			}
			for k := range f.File {
				symID := resolveSymRef(info.File[k])
				f.File[k] = symID.Name
			}
		}
	}
}
