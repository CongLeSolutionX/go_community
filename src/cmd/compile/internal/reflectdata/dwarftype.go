// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"strings"
)

type dwarfType struct {
	typ *types.Type
}

func (d dwarfType) DwarfName(ctxt dwarf.Context) string {
	name := types.TypeSymName(d.typ)
	return strings.Replace(name, `"".`, objabi.PathToPrefix(base.Ctxt.Pkgpath)+".", -1)
}

func (d dwarfType) Name(ctxt dwarf.Context) string {
	return types.TypeSymName(d.typ)
}

func (d dwarfType) Size(ctxt dwarf.Context) int64 {
	return d.typ.Size()
}

func (d dwarfType) Kind(ctxt dwarf.Context) objabi.SymKind {
	return objabi.SymKind(kinds[d.typ.Kind()])
}

func (d dwarfType) RuntimeType(ctxt dwarf.Context) dwarf.Sym {
	return types.TypeSym(d.typ).Linksym()
}

func (d dwarfType) Key(ctxt dwarf.Context) dwarf.Type {
	return dwarfType{typ: d.typ.Key()}
}

func (d dwarfType) Elem(ctxt dwarf.Context) dwarf.Type {
	return dwarfType{d.typ.Elem()}
}

func (d dwarfType) NumElem(ctxt dwarf.Context) int64 {
	if d.typ.IsArray() {
		return d.typ.NumElem()
	}
	if d.typ.IsStruct() {
		return int64(d.typ.NumFields())
	}
	if d.typ.Kind() == types.TFUNC {
		return int64(d.typ.NumParams())
	}
	panic("unreachable")
}

func (d dwarfType) NumResult(ctxt dwarf.Context) int64 {
	return int64(d.typ.NumResults())
}

func (d dwarfType) IsDDD(ctxt dwarf.Context) bool {
	return d.typ.IsVariadic()
}

func (d dwarfType) FieldName(ctxt dwarf.Context, g dwarf.FieldsGroup, i int) string {
	switch g {
	case dwarf.GroupFields:
		return d.typ.FieldName(i)
	case dwarf.GroupParams:
		return dwarfType{d.typ.Params().FieldType(i)}.DwarfName(ctxt)
	case dwarf.GroupResults:
		return dwarfType{d.typ.Results().FieldType(i)}.DwarfName(ctxt)
	}
	panic("unreachable")
}

func (d dwarfType) FieldType(ctxt dwarf.Context, g dwarf.FieldsGroup, i int) dwarf.Type {
	switch g {
	case dwarf.GroupFields:
		return dwarfType{d.typ.FieldType(i)}
	case dwarf.GroupParams:
		return dwarfType{d.typ.Params().FieldType(i)}
	case dwarf.GroupResults:
		return dwarfType{d.typ.Results().FieldType(i)}
	}
	panic("unreachable")
}

func (d dwarfType) FieldIsEmbed(ctxt dwarf.Context, i int) bool {
	return d.typ.Field(i).Embedded != 0
}

func (d dwarfType) FieldOffset(ctxt dwarf.Context, i int) int64 {
	return d.typ.Field(i).Offset
}

func (d dwarfType) IsEface(ctxt dwarf.Context) bool {
	return d.typ.IsEmptyInterface()
}

func LookupDwPredefined(name string) dwarf.Type {
	t := typecheck.LookupRuntime(name[len("runtime."):])
	return dwarfType{typ: t.Type()}
}
