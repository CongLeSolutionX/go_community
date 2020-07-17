// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import (
	"cmd/link/internal/loader"
	"cmd/link/internal/sym"
	"sort"
)

type byTypeStr []typelinkSortKey

type typelinkSortKey struct {
	TypeStr string
	Type    loader.Sym
}

func (s byTypeStr) Less(i, j int) bool { return s[i].TypeStr < s[j].TypeStr }
func (s byTypeStr) Len() int           { return len(s) }
func (s byTypeStr) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// typelink generates the typelink table which is used by reflect.typelinks().
// Types that should be added to the typelinks table are marked with the
// MakeTypelink attribute by the compiler.
func (ctxt *Link) typelink() {
	ldr := ctxt.loader
	typelinks := byTypeStr{}
	for s := loader.Sym(1); s < loader.Sym(ldr.NSym()); s++ {
		if ldr.AttrReachable(s) && ldr.IsTypelink(s) {
			typelinks = append(typelinks, typelinkSortKey{decodetypeStr(ldr, ctxt.Arch, s), s})
		}
	}
	if len(typelinks) == 0 {
		return
	}
	sort.Sort(typelinks)

	genTypelink := func(ctxt *Link, tl loader.Sym) {
		ldr := ctxt.loader
		sb := ldr.MakeSymbolUpdater(tl)
		// We record the offsets for each type symbol from the section.
		// All type symbols are in the same section.
		base := int64(ldr.SymSect(typelinks[0].Type).Vaddr)
		for i, s := range typelinks {
			sb.SetUint32(ctxt.Arch, int64(i*4), uint32(ldr.SymValue(s.Type)-base))
		}
	}
	tl := ctxt.createGeneratorSymbol("runtime.typelink", 0, sym.STYPELINK, int64(4*len(typelinks)), genTypelink)
	ldr.SetAttrLocal(tl, true)
}
