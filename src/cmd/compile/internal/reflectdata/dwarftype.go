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

type DwarfType struct {
	Typ *types.Type
}

func (d DwarfType) DwarfName(ctxt dwarf.Context) string {
	name := types.TypeSymName(d.Typ)
	return strings.Replace(name, `"".`, objabi.PathToPrefix(base.Ctxt.Pkgpath)+".", -1)
}

func (d DwarfType) Name(ctxt dwarf.Context) string {
	return types.TypeSymName(d.Typ)
}

func (d DwarfType) Size(ctxt dwarf.Context) int64 {
	return d.Typ.Size()
}

func (d DwarfType) Kind(ctxt dwarf.Context) objabi.SymKind {
	return objabi.SymKind(kinds[d.Typ.Kind()])
}

func (d DwarfType) RuntimeType(ctxt dwarf.Context) dwarf.Sym {
	return types.TypeSym(d.Typ).Linksym()
}

func (d DwarfType) Key(ctxt dwarf.Context) dwarf.Type {
	return DwarfType{Typ: d.Typ.Key()}
}

func (d DwarfType) Elem(ctxt dwarf.Context) dwarf.Type {
	return DwarfType{d.Typ.Elem()}
}

func (d DwarfType) NumElem(ctxt dwarf.Context) int64 {
	if d.Typ.IsArray() {
		return d.Typ.NumElem()
	}
	if d.Typ.IsStruct() {
		return int64(d.Typ.NumFields())
	}
	if d.Typ.Kind() == types.TFUNC {
		return int64(d.Typ.NumParams())
	}
	panic("unreachable")
}

func (d DwarfType) NumResult(ctxt dwarf.Context) int64 {
	return int64(d.Typ.NumResults())
}

func (d DwarfType) IsDDD(ctxt dwarf.Context) bool {
	return d.Typ.IsVariadic()
}

func (d DwarfType) FieldName(ctxt dwarf.Context, g dwarf.FieldsGroup, i int) string {
	switch g {
	case dwarf.GroupFields:
		return d.Typ.FieldName(i)
	case dwarf.GroupParams:
		return DwarfType{d.Typ.Params().FieldType(i)}.DwarfName(ctxt)
	case dwarf.GroupResults:
		return DwarfType{d.Typ.Results().FieldType(i)}.DwarfName(ctxt)
	}
	panic("unreachable")
}

func (d DwarfType) FieldType(ctxt dwarf.Context, g dwarf.FieldsGroup, i int) dwarf.Type {
	switch g {
	case dwarf.GroupFields:
		return DwarfType{d.Typ.FieldType(i)}
	case dwarf.GroupParams:
		return DwarfType{d.Typ.Params().FieldType(i)}
	case dwarf.GroupResults:
		return DwarfType{d.Typ.Results().FieldType(i)}
	}
	panic("unreachable")
}

func (d DwarfType) FieldIsEmbed(ctxt dwarf.Context, i int) bool {
	return d.Typ.Field(i).Embedded != 0
}

func (d DwarfType) FieldOffset(ctxt dwarf.Context, i int) int64 {
	return d.Typ.Field(i).Offset
}

func (d DwarfType) IsEface(ctxt dwarf.Context) bool {
	return d.Typ.IsEmptyInterface()
}

func LookupDwPredefined(name string) dwarf.Type {
	t := typecheck.LookupRuntime(name[len("runtime."):])
	return DwarfType{Typ: t.Type()}
}
