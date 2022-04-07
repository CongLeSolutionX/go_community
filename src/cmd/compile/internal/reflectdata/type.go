package reflectdata

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/types"
	"cmd/internal/dwarf"
	"cmd/internal/objabi"
	"strings"
)

type DwarfField struct {
	Field *types.Field
}

func (d DwarfField) Name() string {
	if d.Field.Sym == nil || len(d.Field.Sym.Name) == 0 {
		return types.TypeSymName(d.Field.Type)
	}
	return d.Field.Sym.Name
}

func (d DwarfField) Type() dwarf.Type {
	return DwarfType{d.Field.Type}
}

func (d DwarfField) IsDDD() bool {
	return d.Field.IsDDD()
}

func (d DwarfField) IsEmbed() bool {
	return d.Field.Embedded != 0
}

type DwarfType struct {
	Type *types.Type
}

func (d DwarfType) Size() int64 {
	return d.Type.Size()
}

func (d DwarfType) NumElem() int64 {
	return d.Type.NumElem()
}

func (d DwarfType) Name() string {
	name := types.TypeSymName(d.Type)
	return strings.Replace(name, `"".`, objabi.PathToPrefix(base.Ctxt.Pkgpath)+".", -1)

}

func (d DwarfType) Kind() objabi.SymKind {
	return objabi.SymKind(kinds[d.Type.Kind()])
}

func (d DwarfType) RuntimeType() dwarf.Sym {
	return types.TypeSymLookup(d.Name()).Linksym()
}

func (d DwarfType) Key() dwarf.Type {
	return DwarfType{d.Type.Key()}
}

func (d DwarfType) Elem() dwarf.Type {
	return DwarfType{d.Type.Elem()}
}

func (d DwarfType) Params() []dwarf.Field {
	tfields := d.Type.Params().FieldSlice()
	fields := make([]dwarf.Field, len(tfields))
	for i := 0; i < len(fields); i++ {
		fields[i] = DwarfField{tfields[i]}
	}
	return fields
}

func (d DwarfType) Results() []dwarf.Field {
	tfields := d.Type.Results().FieldSlice()
	fields := make([]dwarf.Field, len(tfields))
	for i := 0; i < len(fields); i++ {
		fields[i] = DwarfField{tfields[i]}
	}
	return fields
}

func (d DwarfType) Fields() []dwarf.Field {
	tfields := d.Type.Fields().Slice()
	fields := make([]dwarf.Field, len(tfields))
	for i := 0; i < len(fields); i++ {
		fields[i] = DwarfField{tfields[i]}
	}
	return fields
}

func (d DwarfType) IsEface() bool {
	return d.Type.AllMethods().Len() == 0
}

func (d DwarfType) FieldOff(i int) int64 {
	return d.Type.FieldOff(i)
}
