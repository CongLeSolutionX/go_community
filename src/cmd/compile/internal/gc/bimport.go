// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Binary package import.
// Based loosely on x/tools/go/importer.

package gc

import (
	"cmd/compile/internal/big"
	"encoding/binary"
	"fmt"
)

// Import populates importpkg from the serialized package data.
func Import(data []byte) {
	// determine low-level encoding format
	read := 0
	var format byte = 'm' // missing format
	if len(data) > 0 {
		format = data[0]
		data = data[1:]
		read++
	}
	if format != 'c' && format != 'd' {
		Fatalf("invalid encoding format in export data: got %q; want 'c' or 'd'", format)
	}

	// --- generic export data ---

	p := importer{
		data:        data,
		debugFormat: format == 'd',
		read:        read,
	}

	if v := p.string(); v != exportVersion {
		Fatalf("unknown version: %s", v)
	}

	// populate typList with predeclared "known" types
	p.typList = append(p.typList, predeclared()...)

	// read package data
	p.pkg()
	if p.pkgList[0] != importpkg {
		Fatalf("imported package not found in pkgList[0]")
	}

	// read compiler-specific flags
	importpkg.Safe = p.string() == "safe"

	tcok := typecheckok
	typecheckok = true
	defercheckwidth()

	// read consts
	for i := p.int(); i > 0; i-- {
		sym := p.localname()
		typ := p.typ()
		val := p.value()
		importconst(sym, typ, nodlit(val))
	}

	// read vars
	for i := p.int(); i > 0; i-- {
		sym := p.localname()
		typ := p.typ()
		importvar(sym, typ)
	}

	// read funcs
	for i := p.int(); i > 0; i-- {
		// go.y:hidden_fndcl
		sym := p.localname()
		typ := p.typ()
		importsym(sym, ONAME)
		if sym.Def != nil && sym.Def.Op == ONAME && !Eqtype(typ, sym.Def.Type) {
			Fatalf("inconsistent definition for func %v during import\n\t%v\n\t%v", sym, sym.Def.Type, typ)
		}

		n := newfuncname(sym)
		n.Type = typ
		declare(n, PFUNC)
		funchdr(n)

		// go.y:rom hidden_import
		n.Func.Inl = nil
		funcbody(n)
		importlist = append(importlist, n) // TODO(gri) do this only if body is inlineable?
	}

	// read types
	for i := p.int(); i > 0; i-- {
		// name is parsed as part of named type
		p.typ()
	}

	// --- compiler-specific export data ---

	typecheckok = tcok
	resumecheckwidth()

	testdclstack() // debugging only
}

type importer struct {
	data    []byte
	pkgList []*Pkg
	typList []*Type

	debugFormat bool
	read        int // bytes read
}

func (p *importer) pkg() *Pkg {
	// if the package was seen before, i is its index (>= 0)
	i := p.tagOrIndex()
	if i >= 0 {
		return p.pkgList[i]
	}

	// otherwise, i is the package tag (< 0)
	if i != packageTag {
		Fatalf("unexpected package tag %d", i)
	}

	// read package data
	name := p.string()
	path := p.string()

	// we should never see an empty package name
	if name == "" {
		Fatalf("empty package name in import")
	}

	// we should never see a bad import path
	if isbadimport(path) {
		Fatalf("bad path in import: %q", path)
	}

	// an empty path denotes the package we are currently importing
	pkg := importpkg
	if path != "" {
		pkg = mkpkg(path)
	}
	if pkg.Name == "" {
		pkg.Name = name
	} else if pkg.Name != name {
		Fatalf("inconsistent package names: got %s; want %s (path = %s)", pkg.Name, name, path)
	}
	p.pkgList = append(p.pkgList, pkg)

	return pkg
}

func (p *importer) localname() *Sym {
	// go.y:hidden_importsym
	name := p.string()
	if name == "" {
		name = "?"
	}
	structpkg = importpkg // go.y:hidden_pkg_importsym
	return importpkg.Lookup(name)
}

func (p *importer) newtyp(etype int) *Type {
	t := typ(etype)
	p.typList = append(p.typList, t)
	return t
}

func (p *importer) typ() *Type {
	// if the type was seen before, i is its index (>= 0)
	i := p.tagOrIndex()
	if i >= 0 {
		return p.typList[i]
	}

	// otherwise, i is the type tag (< 0)
	var t *Type
	switch i {
	case namedTag:
		// go.y:hidden_importsym
		tsym := p.qualifiedName()

		// go.y:hidden_pkgtype
		t = pkgtype(tsym)
		importsym(tsym, OTYPE)
		p.typList = append(p.typList, t)

		// read underlying type
		// go.y:hidden_type
		t0 := p.typ()
		// TOOD(gri) the test below is done by importtype; pulled out so we can debug easily
		if t.Etype != TFORW && !Eqtype(t.Orig, t0) {
			fmt.Printf("importing %s (%s)\n", importpkg.Path, importpkg.Name)
			panic("types inconsistent")
		}
		importtype(t, t0) // go.y:hidden_import

		// interfaces don't have associated methods
		if t0.Etype == TINTER {
			break
		}

		// read associated methods
		for i := p.int(); i > 0; i-- {
			// go.y:hidden_fndcl
			name := p.string()
			recv := p.paramList() // TODO(gri) do we need a full param list for the receiver?
			params := p.paramList()
			result := p.paramList()

			// TODO(gri) is this the correct package to use?
			// (neither importpkg nor builtinpkg appear to be correct)
			pkg := localpkg
			if !exportname(name) {
				pkg = tsym.Pkg
			}
			sym := pkg.Lookup(name)

			n := methodname1(newname(sym), recv.N.Right)
			n.Type = functype(recv.N, params, result)
			checkwidth(n.Type)
			// addmethod uses the global variable structpkg to verify consistency
			{
				saved := structpkg
				structpkg = tsym.Pkg
				addmethod(sym, n.Type, false, nointerface)
				structpkg = saved
			}
			nointerface = false
			funchdr(n)

			// (comment from go.y)
			// inl.C's inlnode in on a dotmeth node expects to find the inlineable body as
			// (dotmeth's type).Nname.Inl, and dotmeth's type has been pulled
			// out by typecheck's lookdot as this $$.ttype.  So by providing
			// this back link here we avoid special casing there.
			n.Type.Nname = n

			// go.y:hidden_import
			n.Func.Inl = nil
			funcbody(n)
			importlist = append(importlist, n) // TODO(gri) do this only if body is inlineable?
		}

	case arrayTag, sliceTag:
		t = p.newtyp(TARRAY)
		t.Bound = -1
		if i == arrayTag {
			t.Bound = p.int64()
		}
		t.Type = p.typ()

	case dddTag:
		t = p.newtyp(T_old_DARRAY)
		t.Bound = -1
		t.Type = p.typ()

	case structTag:
		t = p.newtyp(TSTRUCT)
		tostruct0(t, p.fieldList())

	case pointerTag:
		t = p.newtyp(Tptr)
		t.Type = p.typ()

	case signatureTag:
		t = p.newtyp(TFUNC)
		params := p.paramList()
		result := p.paramList()
		functype0(t, nil, params, result)

	case interfaceTag:
		t = p.newtyp(TINTER)
		if p.int() != 0 {
			Fatalf("unexpected embedded interface")
		}
		tointerface0(t, p.methodList())

	case mapTag:
		t = p.newtyp(TMAP)
		t.Down = p.typ() // key
		t.Type = p.typ() // val

	case chanTag:
		t = p.newtyp(TCHAN)
		t.Chan = uint8(p.int())
		t.Type = p.typ()

	default:
		Fatalf("unexpected type (tag = %d)", i)
	}

	if t == nil {
		Fatalf("nil type (type tag = %d)", i)
	}

	return t
}

func (p *importer) qualifiedName() *Sym {
	name := p.string()
	pkg := p.pkg()
	return pkg.Lookup(name)
}

// go.y:hidden_structdcl_list
func (p *importer) fieldList() *NodeList {
	i := p.int()
	if i == 0 {
		return nil
	}
	n := list1(p.field())
	for i--; i > 0; i-- {
		n = list(n, p.field())
	}
	return n
}

// go.y:hidden_structdcl
func (p *importer) field() *Node {
	sym := p.fieldName()
	typ := p.typ()
	note := p.string()

	var n *Node
	if sym.Name != "" {
		n = Nod(ODCLFIELD, newname(sym), typenod(typ))
	} else {
		// anonymous field - typ must be T or *T and T must be a type name
		s := typ.Sym
		if s == nil && Isptr[typ.Etype] {
			s = typ.Type.Sym // deref
		}
		n = embedded(s, sym.Pkg)
		n.Right = typenod(typ)
	}

	if note != "" {
		n.SetVal(Val{U: note})
	}

	return n
}

// go.y:hidden_interfacedcl_list
func (p *importer) methodList() *NodeList {
	i := p.int()
	if i == 0 {
		return nil
	}
	n := list1(p.method())
	for i--; i > 0; i-- {
		n = list(n, p.method())
	}
	return n
}

// go.y:hidden_interfacedcl
func (p *importer) method() *Node {
	sym := p.fieldName()
	params := p.paramList()
	result := p.paramList()
	return Nod(ODCLFIELD, newname(sym), typenod(functype(fakethis(), params, result)))
}

// go.y:hidden_importsym
func (p *importer) fieldName() *Sym {
	name := p.string()
	// TODO(gri) Is this the correct package here?
	// (importpkg and builtinpkg don't work here)
	pkg := localpkg // anonymous and exported names assume local package
	if name != "" && !exportname(name) {
		pkg = p.pkg()
	}
	return pkg.Lookup(name)
}

// go.y:ohidden_funarg_list
func (p *importer) paramList() *NodeList {
	i := p.int()
	if i == 0 {
		return nil
	}
	// negative length indicates unnamed parameters
	named := true
	if i < 0 {
		i = -i
		named = false
	}
	// i > 0
	n := list1(p.param(named))
	i--
	for ; i > 0; i-- {
		n = list(n, p.param(named))
	}
	return n
}

// go.y:hidden_funarg
func (p *importer) param(named bool) *Node {
	typ := p.typ()

	isddd := false
	if typ.Etype == T_old_DARRAY {
		// T_old_DARRAY indicates ... type
		typ.Etype = TARRAY
		isddd = true
	}

	n := Nod(ODCLFIELD, nil, typenod(typ))
	n.Isddd = isddd

	if named {
		name := p.string()
		if name == "" {
			Fatalf("expected named parameter")
		}
		// TODO(gri) Is this the right package here?
		// Both localpkg and importpkg work.
		// see go.y:1181
		n.Left = newname(importpkg.Lookup(name))
	}

	return n
}

func (p *importer) value() (x Val) {
	switch tag := p.tagOrIndex(); tag {
	case falseTag:
		x.U = false
	case trueTag:
		x.U = true
	case int64Tag:
		u := new(Mpint)
		Mpmovecfix(u, p.int64())
		x.U = u
	case floatTag:
		u := newMpflt()
		p.float(u)
		x.U = u
	case complexTag:
		u := new(Mpcplx)
		p.float(&u.Real)
		p.float(&u.Imag)
		x.U = u
	case stringTag:
		x.U = p.string()
	default:
		Fatalf("unexpected value tag %d", tag)
	}
	return
}

func (p *importer) float(x *Mpflt) {
	sign := p.int()
	if sign == 0 {
		Mpmovecflt(x, 0)
		return
	}

	exp := p.int()
	mant := new(big.Int).SetBytes([]byte(p.string()))

	m := x.Val.SetInt(mant)
	m.SetMantExp(m, exp-mant.BitLen())
	if sign < 0 {
		m.Neg(m)
	}
}

// ----------------------------------------------------------------------------
// Low-level decoders

func (p *importer) tagOrIndex() int {
	if p.debugFormat {
		p.marker('t')
	}

	return int(p.rawInt64())
}

func (p *importer) int() int {
	x := p.int64()
	if int64(int(x)) != x {
		Fatalf("exported integer too large")
	}
	return int(x)
}

func (p *importer) int64() int64 {
	if p.debugFormat {
		p.marker('i')
	}

	return p.rawInt64()
}

func (p *importer) string() string {
	if p.debugFormat {
		p.marker('s')
	}

	var b []byte
	if n := int(p.rawInt64()); n > 0 {
		b = p.data[:n]
		p.data = p.data[n:]
		p.read += n
	}
	return string(b)
}

func (p *importer) marker(want byte) {
	if got := p.data[0]; got != want {
		Fatalf("incorrect marker: got %c; want %c (pos = %d)", got, want, p.read)
	}
	p.data = p.data[1:]
	p.read++

	pos := p.read
	if n := int(p.rawInt64()); n != pos {
		Fatalf("incorrect position: got %d; want %d", n, pos)
	}
}

// rawInt64 should only be used by low-level decoders
func (p *importer) rawInt64() int64 {
	i, n := binary.Varint(p.data)
	p.data = p.data[n:]
	p.read += n
	return i
}
