// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Binary package import.
// Based loosely on x/tools/go/importer.

package gc

import (
	"cmd/compile/internal/big"
	"encoding/binary"
)

// // mypkgMap and newpkg are work-arounds to avoid conflicts
// // since we are importing the package that we just compiled
// // for testing
// // TODO(gri) remove this in favor of calling mkpkg
// var mypkgMap = make(map[string]*Pkg)

// func newpkg(path, name string) *Pkg {
// 	if p := mypkgMap[path]; p != nil {
// 		if p.Name != name {
// 			Fatalf("package name inconsistency")
// 		}
// 		return p
// 	}

// 	p := new(Pkg)
// 	p.Path = path
// 	p.Name = name
// 	p.Prefix = pathtoprefix(path)
// 	p.Syms = make(map[string]*Sym)
// 	mypkgMap[path] = p
// 	return p
// }

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

	pkg := p.pkg()

	// sanity checks
	if p.pkgList[0] != pkg {
		Fatalf("imported package not found in pkgList[0]")
	}
	if pkg != importpkg {
		Fatalf("incorrect package setup")
	}

	tcok := typecheckok
	typecheckok = 1
	defercheckwidth()

	// read objects
	for p.obj() {
		// nothing to do here
	}

	// --- compiler-specific export data ---

	// read objects required for inlined functions/methods
	for p.obj() {
		// nothing to do here
	}

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

// (hidden_import in go.y)
func (p *importer) obj() bool {
	switch tag := p.tagOrIndex(); tag {
	case constTag:
		sym := p.localname()
		typ := p.typ()
		val := p.value()
		importconst(sym, typ, nodlit(val))

	case typeTag:
		// name is parsed as part of named type
		p.typ()

	case varTag:
		sym := p.localname()
		typ := p.typ()
		importvar(sym, typ)

	case funcTag:
		// (hidden_fndcl in go.y)
		sym := p.localname()
		typ := p.typ()
		importsym(sym, ONAME)
		if sym.Def != nil {
			Fatalf("function imported before?")
		}
		n := newfuncname(sym)
		n.Type = typ
		declare(n, PFUNC)
		funchdr(n)

		// from hidden_import in go.y
		n.Func.Inl = nil
		funcbody(n)
		importlist = list(importlist, n)

	case endTag:
		return false

	default:
		Fatalf("unexpected object tag %d", tag)
	}
	return true
}

func (p *importer) localname() *Sym {
	return p.pkgList[0].Lookup(p.string())
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
		if i < len(p.typList) {
			return p.typList[i]
		}
		Fatalf("typ index %d out of range (len = %d)", i, len(p.typList))
	}

	// otherwise, i is the type tag (< 0)
	var t *Type
	switch i {
	case namedTag:
		// (hidden_importsym in go.y)
		sym := p.qualifiedName()
		tpkg := sym.Pkg

		// (hidden_pkgtype in go.y)
		t = pkgtype(sym)
		importsym(sym, OTYPE)
		p.typList = append(p.typList, t)

		// read underlying type
		t0 := p.typ()
		importtype(t, t0)

		// interfaces don't have associated methods
		if t0.Etype == TINTER {
			break
		}

		// read associated methods
		for i := p.int(); i > 0; i-- {
			// (hidden_fndecl in go.y)
			sym := tpkg.Lookup(p.string())
			recv := p.paramList()
			params := p.paramList()
			result := p.paramList()

			n := methodname1(newname(sym), recv.N.Right)
			n.Type = functype(recv.N, params, result)
			checkwidth(n.Type)
			// addmethod uses the global variable (!!!) structpkg to verify consistency;
			// set it here so we know it's correctly set up.
			structpkg = tpkg                                       // set in hidden_pkg_importsym in go.y, used in addmethod in decl.go
			addmethod(sym, n.Type, false, false /* nointerface */) // TODO(gri) verify value for nointerface here
			funchdr(n)

			// inl.C's inlnode in on a dotmeth node expects to find the inlineable body as
			// (dotmeth's type).Nname.Inl, and dotmeth's type has been pulled
			// out by typecheck's lookdot as this $$.ttype.  So by providing
			// this back link here we avoid special casing there.
			n.Type.Nname = n

			// from hidden_import in go.y
			n.Func.Inl = nil
			funcbody(n)
			importlist = list(importlist, n)
		}

	case arrayTag, sliceTag:
		t = p.newtyp(TARRAY)
		t.Bound = -1
		if i == arrayTag {
			t.Bound = p.int64()
		}
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
		dir := p.int()
		switch dir {
		case 0:
		case 'r':
			dir = Crecv
		case 's':
			dir = Csend
		default:
			Fatalf("unexpected channel direction %d", dir)
		}
		t.Chan = uint8(dir)
		t.Type = p.typ()

	default:
		Fatalf("unexpected type (tag = %d)", i)
	}

	if t == nil {
		Fatalf("nil type (type tag = %d)", i)
	}

	return t
}

// (hidden_structdcl_list in go.y)
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

// (hidden_structdcl in go.y)
func (p *importer) field() *Node {
	sym := p.fieldName()
	typ := p.typ()
	note := p.string()

	var n *Node
	if sym.Name != "" {
		n = Nod(ODCLFIELD, newname(sym), typenod(typ))
	} else {
		// anonymous field
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

// (hidden_structdcl_list in go.y)
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

// (hidden_interfacedcl in go.y)
func (p *importer) method() *Node {
	sym := p.fieldName()
	params := p.paramList()
	result := p.paramList()
	return Nod(ODCLFIELD, newname(sym), typenod(functype(nil, params, result)))
}

func (p *importer) qualifiedName() *Sym {
	name := p.string()
	pkg := p.pkg()
	return pkg.Lookup(name)
}

// (hidden_importsym in go.y)
func (p *importer) fieldName() *Sym {
	name := p.string()
	pkg := p.pkgList[0] // anonymous and exported names assume current package
	if name != "" && !exportname(name) {
		pkg = p.pkg()
	}
	return pkg.Lookup(name)
}

// (ohidden_funarg_list in go.y)
func (p *importer) paramList() *NodeList {
	// negative length indicates variadic list
	i0 := p.int()
	if i0 == 0 {
		return nil
	}
	// i != 0
	i := i0
	if i < 0 {
		i = -i
	}
	// i > 0
	n := list1(p.param(i0 < 0 && i == 1))
	i--
	for ; i > 0; i-- {
		n = list(n, p.param(i0 < 0 && i == 1))
	}
	return n
}

// (hidden_funarg in go.y)
func (p *importer) param(variadic bool) *Node {
	t := p.typ()

	if variadic {
		st := typ(TARRAY)
		st.Bound = -1
		st.Type = t
		t = st
	}

	n := Nod(ODCLFIELD, nil, typenod(t))
	if name := p.string(); name != "" {
		n.Left = newname(nopkg.Lookup(name)) // TODO(gri) is nopkg the right package here?
	}
	n.Isddd = variadic

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

	// read mantissa bytes (big endian) and exponent
	mant := new(big.Int).SetBytes([]byte(p.string()))
	exp := p.int()

	// set mantissa from *big.Int, set exponent, sign
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
	return int(p.int64())
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
