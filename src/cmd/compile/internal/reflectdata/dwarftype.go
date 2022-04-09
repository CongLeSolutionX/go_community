// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"strings"
)

type dwarfType struct {
	typ *types.Type
}

func (d dwarfType) DwarfName(interface{}) string {
	name := types.TypeSymName(d.typ)
	return strings.Replace(name, `"".`, objabi.PathToPrefix(base.Ctxt.Pkgpath)+".", -1)
}

func (d dwarfType) Name(interface{}) string {
	return types.TypeSymName(d.typ)
}

func (d dwarfType) Size(interface{}) int64 {
	return d.typ.Size()
}

func (d dwarfType) Kind(interface{}) objabi.SymKind {
	return objabi.SymKind(kinds[d.typ.Kind()])
}

func (d dwarfType) RuntimeType(interface{}) dwarf.Sym {
	return types.TypeSym(d.typ).Linksym()
}

func (d dwarfType) Key(interface{}) dwarf.Type {
	return dwarfType{typ: d.typ.Key()}
}

func (d dwarfType) Elem(interface{}) dwarf.Type {
	return dwarfType{d.typ.Elem()}
}

func (d dwarfType) NumElem(interface{}) int64 {
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

func (d dwarfType) NumResult(interface{}) int64 {
	return int64(d.typ.NumResults())
}

func (d dwarfType) IsDDD(interface{}) bool {
	return d.typ.IsVariadic()
}

func (d dwarfType) FieldName(dwctxt interface{}, g dwarf.FieldsGroup, i int) string {
	switch g {
	case dwarf.GroupFields:
		return d.typ.FieldName(i)
	case dwarf.GroupParams:
		return dwarfType{d.typ.Params().FieldType(i)}.DwarfName(d)
	case dwarf.GroupResults:
		return dwarfType{d.typ.Results().FieldType(i)}.DwarfName(d)
	}
	panic("unreachable")
}

func (d dwarfType) FieldType(dwctxt interface{}, g dwarf.FieldsGroup, i int) dwarf.Type {
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

func (d dwarfType) FieldIsEmbed(dwctxt interface{}, i int) bool {
	return d.typ.Field(i).Embedded != 0
}

func (d dwarfType) FieldOffset(dwctxt interface{}, i int) int64 {
	return d.typ.Field(i).Offset
}

func (d dwarfType) IsEface(interface{}) bool {
	return d.typ.IsEmptyInterface()
}
