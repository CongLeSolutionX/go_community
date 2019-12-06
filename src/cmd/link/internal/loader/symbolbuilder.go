// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

import (
	"cmd/internal/objabi"
	"cmd/internal/sys"
	"cmd/link/internal/sym"
	"log"
	"math/bits"
)

// SymbolBuilder is a helper designed to help with the construction
// of new symbol contents.
type SymbolBuilder struct {
	*extSymPayload         // points to payload being updated
	symIdx         Sym     // index of symbol being updated/constructed
	l              *Loader // loader
}

// NewSymbolBuilder creates a symbol builder for use in constructing
// an entirely new symbol.
func (l *Loader) MakeSymbolBuilder(name string) *SymbolBuilder {
	// for now assume that any new sym is intended to be static
	symIdx := l.CreateExtSym(name)
	if l.Syms[symIdx] != nil {
		panic("can't build if sym.Symbol already present")
	}
	if l.hasPayload.Has(symIdx - l.extStart) {
		panic("can't build if no payload flag")
	}

	sb := &SymbolBuilder{l: l, symIdx: symIdx}
	sb.extSymPayload = &l.payloads[symIdx-l.extStart]
	return sb
}

// NewSymbolBuilder creates a symbol builder helper for an already-allocated
// external symbol 'symIdx'.
func (l *Loader) MakeSymbolUpdater(symIdx Sym) *SymbolBuilder {
	if !l.isExternal(symIdx) {
		panic("can't build on non-external sym")
	}
	if l.Syms[symIdx] != nil {
		panic("can't build if sym.Symbol already present")
	}
	if !l.hasPayload.Has(symIdx - l.extStart) {
		panic("can't build if no payload flag")
	}
	sb := &SymbolBuilder{l: l, symIdx: symIdx}
	sb.extSymPayload = &l.payloads[symIdx-l.extStart]
	return sb
}

// Sym returns the underlying sym index from a sym builder.
func (sb *SymbolBuilder) Sym() Sym {
	return sb.symIdx
}

func (sb *SymbolBuilder) Name() string {
	return sb.name
}

func (sb *SymbolBuilder) Version() int {
	return sb.ver
}

func (sb *SymbolBuilder) Type() sym.SymKind {
	return sb.kind
}
func (sb *SymbolBuilder) SetType(kind sym.SymKind) {
	sb.kind = kind
}

func (sb *SymbolBuilder) Size() int64 {
	return sb.size
}

func (sb *SymbolBuilder) SetSize(size int64) {
	sb.size = size
}

func (sb *SymbolBuilder) Value() int64 {
	return sb.value
}

func (sb *SymbolBuilder) SetValue(v int64) {
	sb.value = v
}

func (sb *SymbolBuilder) Align() int32 {
	// If an alignment has been recorded, return that.
	if align, ok := sb.l.align[sb.symIdx]; ok {
		return align
	}
	// TODO: would it make sense to return an arch-specific
	// alignment depending on section type? E.g. STEXT => 32,
	// SDATA => 1, etc?
	return 0
}

func (sb *SymbolBuilder) SetAlign(align int32) {
	// Reject nonsense alignments.
	// TODO: do we need this?
	if align < 0 {
		panic("bad alignment value")
	}
	if bits.OnesCount32(uint32(align)) != 1 {
		panic("bad alignment value")
	}
	if align == 0 {
		delete(sb.l.align, sb.symIdx)
	} else {
		sb.l.align[sb.symIdx] = align
	}
}

func (sb *SymbolBuilder) Data() []byte {
	return sb.data
}

func (sb *SymbolBuilder) SetData(data []byte) {
	sb.data = data
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

func (sb *SymbolBuilder) AddBytes(data []byte) {
	sb.setReachable()
	if sb.kind == 0 {
		sb.kind = sym.SDATA
	}
	sb.data = append(sb.data, data...)
	sb.size = int64(len(sb.data))
}

func (sb *SymbolBuilder) Relocs() []Reloc {
	return sb.relocs
}

func (sb *SymbolBuilder) AddReloc(r Reloc) {
	sb.relocs = append(sb.relocs, r)
}

func (sb *SymbolBuilder) Reachable() bool {
	return sb.l.attrReachable.Has(sb.symIdx)
}

func (sb *SymbolBuilder) setReachable() {
	sb.l.SetAttrReachable(sb.symIdx, true)
}

// PrependSub prepends 'sub' onto the sub list for a given outer symbol.
// Will panic if 'sub' already has an outer sym or sub sym.
func (sb *SymbolBuilder) PrependSub(sub Sym) {
	if !sb.l.hasPayload.Has(sub - sb.l.extStart) {
		panic("expected hasPayload in PrependSub")
	}
	outer := sb.symIdx
	if sb.l.OuterSym(outer) != 0 {
		panic("outer has outer itself")
	}
	if sb.l.SubSym(sub) != 0 {
		panic("sub set for subsym")
	}
	if sb.l.OuterSym(sub) != 0 {
		panic("outer already set for subsym")
	}
	sb.l.sub[sub] = sb.l.sub[outer]
	sb.l.sub[outer] = sub
	sb.l.outer[sub] = outer
}

func (sb *SymbolBuilder) AddUint8(v uint8) int64 {
	off := sb.size
	if sb.kind == 0 {
		sb.kind = sym.SDATA
	}
	sb.setReachable()
	sb.size++
	sb.data = append(sb.data, v)
	return off
}

func (sb *SymbolBuilder) AddUintXX(arch *sys.Arch, v uint64, wid int) int64 {
	off := sb.size
	sb.setReachable()
	sb.setUintXX(arch, off, v, int64(wid))
	return off
}

func (sb *SymbolBuilder) setUintXX(arch *sys.Arch, off int64, v uint64, wid int64) int64 {
	if sb.kind == 0 {
		sb.kind = sym.SDATA
	}
	if sb.size < off+wid {
		sb.size = off + wid
		sb.Grow(sb.size)
	}

	switch wid {
	case 1:
		sb.data[off] = uint8(v)
	case 2:
		arch.ByteOrder.PutUint16(sb.data[off:], uint16(v))
	case 4:
		arch.ByteOrder.PutUint32(sb.data[off:], uint32(v))
	case 8:
		arch.ByteOrder.PutUint64(sb.data[off:], v)
	}

	return off + wid
}

func (sb *SymbolBuilder) AddUint16(arch *sys.Arch, v uint16) int64 {
	return sb.AddUintXX(arch, uint64(v), 2)
}

func (sb *SymbolBuilder) AddUint32(arch *sys.Arch, v uint32) int64 {
	return sb.AddUintXX(arch, uint64(v), 4)
}

func (sb *SymbolBuilder) AddUint64(arch *sys.Arch, v uint64) int64 {
	return sb.AddUintXX(arch, v, 8)
}

func (sb *SymbolBuilder) AddUint(arch *sys.Arch, v uint64) int64 {
	return sb.AddUintXX(arch, v, arch.PtrSize)
}

func (sb *SymbolBuilder) SetUint8(arch *sys.Arch, r int64, v uint8) int64 {
	sb.setReachable()
	return sb.setUintXX(arch, r, uint64(v), 1)
}

func (sb *SymbolBuilder) SetUint16(arch *sys.Arch, r int64, v uint16) int64 {
	sb.setReachable()
	return sb.setUintXX(arch, r, uint64(v), 2)
}

func (sb *SymbolBuilder) SetUint32(arch *sys.Arch, r int64, v uint32) int64 {
	sb.setReachable()
	return sb.setUintXX(arch, r, uint64(v), 4)
}

func (sb *SymbolBuilder) SetUint(arch *sys.Arch, r int64, v uint64) int64 {
	sb.setReachable()
	return sb.setUintXX(arch, r, v, int64(arch.PtrSize))
}

func (sb *SymbolBuilder) Addstring(str string) int64 {
	sb.setReachable()
	if sb.kind == 0 {
		sb.kind = sym.SNOPTRDATA
	}
	r := sb.size
	if sb.name == ".shstrtab" {
		// FIXME: find a better mechanism for this
		sb.l.elfsetstring(nil, str, int(r))
	}
	sb.data = append(sb.data, str...)
	sb.data = append(sb.data, 0)
	sb.size = int64(len(sb.data))
	return r
}

func (sb *SymbolBuilder) addRel() *Reloc {
	sb.relocs = append(sb.relocs, Reloc{})
	return &sb.relocs[len(sb.relocs)-1]
}

func (sb *SymbolBuilder) addAddrPlus(tgt Sym, add int64, typ objabi.RelocType, rsize int) int64 {
	if sb.kind == 0 {
		sb.kind = sym.SDATA
	}
	i := sb.size

	sb.size += int64(rsize)
	sb.Grow(sb.size)

	r := sb.addRel()
	r.Sym = tgt
	r.Off = int32(i)
	r.Size = uint8(rsize)
	r.Type = typ
	r.Add = add

	return i + int64(r.Size)
}

func (sb *SymbolBuilder) AddAddrPlus(arch *sys.Arch, tgt Sym, add int64) int64 {
	sb.setReachable()
	return sb.addAddrPlus(tgt, add, objabi.R_ADDR, arch.PtrSize)
}

func (sb *SymbolBuilder) AddAddrPlus4(arch *sys.Arch, tgt Sym, add int64) int64 {
	sb.setReachable()
	return sb.addAddrPlus(tgt, add, objabi.R_ADDR, 4)
}

func (sb *SymbolBuilder) AddCURelativeAddrPlus(arch *sys.Arch, tgt Sym, add int64) int64 {
	sb.setReachable()
	return sb.addAddrPlus(tgt, add, objabi.R_ADDRCUOFF, arch.PtrSize)
}
