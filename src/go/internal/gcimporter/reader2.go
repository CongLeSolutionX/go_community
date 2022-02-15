// UNREVIEWED

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gcimporter

import (
	"fmt"
	"go/token"
	"go/types"
	"internal/pkgbits"
)

type pkgReader2 struct {
	pkgbits.PkgDecoder

	fake fakeFileSet

	fset    *token.FileSet
	ctxt    *types.Context
	imports map[string]*types.Package

	posBases []string
	pkgs     []*types.Package
	typs     []types.Type
}

func readPackage2(fset *token.FileSet, ctxt *types.Context, imports map[string]*types.Package, input pkgbits.PkgDecoder) *types.Package {
	pr := pkgReader2{
		PkgDecoder: input,

		fake: fakeFileSet{
			fset:  fset,
			files: make(map[string]*fileInfo),
		},

		fset:    fset,
		ctxt:    ctxt,
		imports: imports,

		posBases: make([]string, input.NumElems(pkgbits.RelocPosBase)),
		pkgs:     make([]*types.Package, input.NumElems(pkgbits.RelocPkg)),
		typs:     make([]types.Type, input.NumElems(pkgbits.RelocType)),
	}

	r := pr.newReader(pkgbits.RelocMeta, pkgbits.PublicRootIdx, pkgbits.SyncPublic)
	pkg := r.pkg()
	r.Bool() // has init

	for i, n := 0, r.Len(); i < n; i++ {
		// As if r.obj(), but avoiding the Scope.Lookup call,
		// to avoid eager loading of imports.
		r.Sync(pkgbits.SyncObject)
		assert(!r.Bool())
		r.p.objIdx(r.Reloc(pkgbits.RelocObj))
		assert(r.Len() == 0)
	}

	r.Sync(pkgbits.SyncEOF)

	pkg.MarkComplete()
	return pkg
}

type reader2 struct {
	pkgbits.Decoder

	p *pkgReader2

	dict *reader2Dict
}

type reader2Dict struct {
	bounds []typeInfo

	tparams []*types.TypeParam

	derived      []derivedInfo
	derivedTypes []types.Type
}

type reader2TypeBound struct {
	derived  bool
	boundIdx int
}

func (pr *pkgReader2) newReader(k pkgbits.RelocKind, idx int, marker pkgbits.SyncMarker) *reader2 {
	return &reader2{
		Decoder: pr.NewDecoder(k, idx, marker),
		p:       pr,
	}
}

// @@@ Positions

func (r *reader2) pos() token.Pos {
	r.Sync(pkgbits.SyncPos)
	if !r.Bool() {
		return token.NoPos
	}

	// TODO(mdempsky): Delta encoding.
	posBase := r.posBase()
	line := r.Uint()
	col := r.Uint()
	return r.p.fake.pos(posBase, int(line), int(col))
}

func (r *reader2) posBase() string {
	return r.p.posBaseIdx(r.Reloc(pkgbits.RelocPosBase))
}

func (pr *pkgReader2) posBaseIdx(idx int) string {
	if b := pr.posBases[idx]; b != "" {
		return b
	}

	r := pr.newReader(pkgbits.RelocPosBase, idx, pkgbits.SyncPosBase)

	filename := r.String()

	if r.Bool() {
		// b = token.NewTrimmedFileBase(filename, true)
	} else {
		pos := r.pos()
		line := r.Uint()
		col := r.Uint()

		// b = token.NewLineBase(pos, filename, true, line, col)
		_, _, _ = pos, line, col
	}

	b := filename
	pr.posBases[idx] = b
	return b
}

// @@@ Packages

func (r *reader2) pkg() *types.Package {
	r.Sync(pkgbits.SyncPkg)
	return r.p.pkgIdx(r.Reloc(pkgbits.RelocPkg))
}

func (pr *pkgReader2) pkgIdx(idx int) *types.Package {
	// TODO(mdempsky): Consider using some non-nil pointer to indicate
	// the universe scope, so we don't need to keep re-reading it.
	if pkg := pr.pkgs[idx]; pkg != nil {
		return pkg
	}

	pkg := pr.newReader(pkgbits.RelocPkg, idx, pkgbits.SyncPkgDef).doPkg()
	pr.pkgs[idx] = pkg
	return pkg
}

func (r *reader2) doPkg() *types.Package {
	path := r.String()
	if path == "builtin" {
		return nil // universe
	}
	if path == "unsafe" {
		return types.Unsafe
	}
	if path == "" {
		path = r.p.PkgPath()
	}

	if pkg := r.p.imports[path]; pkg != nil {
		return pkg
	}

	name := r.String()
	height := r.Len()

	// pkg := types.NewPackageHeight(path, name, height)
	pkg, _ := types.NewPackage(path, name), height
	r.p.imports[path] = pkg

	// TODO(mdempsky): The list of imported packages is important for
	// go/types, but we could probably skip populating it for types.
	imports := make([]*types.Package, r.Len())
	for i := range imports {
		imports[i] = r.pkg()
	}
	pkg.SetImports(imports)

	return pkg
}

// @@@ Types

func (r *reader2) typ() types.Type {
	return r.p.typIdx(r.typInfo(), r.dict)
}

func (r *reader2) typInfo() typeInfo {
	r.Sync(pkgbits.SyncType)
	if r.Bool() {
		return typeInfo{idx: r.Len(), derived: true}
	}
	return typeInfo{idx: r.Reloc(pkgbits.RelocType), derived: false}
}

func (pr *pkgReader2) typIdx(info typeInfo, dict *reader2Dict) types.Type {
	idx := info.idx
	var where *types.Type
	if info.derived {
		where = &dict.derivedTypes[idx]
		idx = dict.derived[idx].idx
	} else {
		where = &pr.typs[idx]
	}

	if typ := *where; typ != nil {
		return typ
	}

	r := pr.newReader(pkgbits.RelocType, idx, pkgbits.SyncTypeIdx)
	r.dict = dict

	typ := r.doTyp()
	assert(typ != nil)

	// See comment in pkgReader.typIdx explaining how this happens.
	if prev := *where; prev != nil {
		return prev
	}

	*where = typ
	return typ
}

func (r *reader2) doTyp() (res types.Type) {
	switch tag := pkgbits.CodeType(r.Code(pkgbits.SyncType)); tag {
	default:
		errorf("unhandled type tag: %v", tag)
		panic("unreachable")

	case pkgbits.TypeBasic:
		return types.Typ[r.Len()]

	case pkgbits.TypeNamed:
		obj, targs := r.obj()
		name := obj.(*types.TypeName)
		if len(targs) != 0 {
			t, _ := types.Instantiate(r.p.ctxt, name.Type(), targs, false)
			return t
		}
		return name.Type()

	case pkgbits.TypeTypeParam:
		return r.dict.tparams[r.Len()]

	case pkgbits.TypeArray:
		len := int64(r.Uint64())
		return types.NewArray(r.typ(), len)
	case pkgbits.TypeChan:
		dir := types.ChanDir(r.Len())
		return types.NewChan(dir, r.typ())
	case pkgbits.TypeMap:
		return types.NewMap(r.typ(), r.typ())
	case pkgbits.TypePointer:
		return types.NewPointer(r.typ())
	case pkgbits.TypeSignature:
		return r.signature(nil, nil, nil)
	case pkgbits.TypeSlice:
		return types.NewSlice(r.typ())
	case pkgbits.TypeStruct:
		return r.structType()
	case pkgbits.TypeInterface:
		return r.interfaceType()
	case pkgbits.TypeUnion:
		return r.unionType()
	}
}

func (r *reader2) structType() *types.Struct {
	fields := make([]*types.Var, r.Len())
	var tags []string
	for i := range fields {
		pos := r.pos()
		pkg, name := r.selector()
		ftyp := r.typ()
		tag := r.String()
		embedded := r.Bool()

		fields[i] = types.NewField(pos, pkg, name, ftyp, embedded)
		if tag != "" {
			for len(tags) < i {
				tags = append(tags, "")
			}
			tags = append(tags, tag)
		}
	}
	return types.NewStruct(fields, tags)
}

func (r *reader2) unionType() *types.Union {
	terms := make([]*types.Term, r.Len())
	for i := range terms {
		terms[i] = types.NewTerm(r.Bool(), r.typ())
	}
	return types.NewUnion(terms)
}

func (r *reader2) interfaceType() *types.Interface {
	methods := make([]*types.Func, r.Len())
	embeddeds := make([]types.Type, r.Len())
	implicit := len(methods) == 0 && len(embeddeds) == 1 && r.Bool()

	for i := range methods {
		pos := r.pos()
		pkg, name := r.selector()
		mtyp := r.signature(nil, nil, nil)
		methods[i] = types.NewFunc(pos, pkg, name, mtyp)
	}

	for i := range embeddeds {
		embeddeds[i] = r.typ()
	}

	iface := types.NewInterfaceType(methods, embeddeds)
	if implicit {
		iface.MarkImplicit()
	}
	return iface
}

func (r *reader2) signature(recv *types.Var, rtparams, tparams []*types.TypeParam) *types.Signature {
	r.Sync(pkgbits.SyncSignature)

	params := r.params()
	results := r.params()
	variadic := r.Bool()

	return types.NewSignatureType(recv, rtparams, tparams, params, results, variadic)
}

func (r *reader2) params() *types.Tuple {
	r.Sync(pkgbits.SyncParams)
	params := make([]*types.Var, r.Len())
	for i := range params {
		params[i] = r.param()
	}
	return types.NewTuple(params...)
}

func (r *reader2) param() *types.Var {
	r.Sync(pkgbits.SyncParam)

	pos := r.pos()
	pkg, name := r.localIdent()
	typ := r.typ()

	return types.NewParam(pos, pkg, name, typ)
}

// @@@ Objects

func (r *reader2) obj() (types.Object, []types.Type) {
	r.Sync(pkgbits.SyncObject)

	assert(!r.Bool())

	pkg, name := r.p.objIdx(r.Reloc(pkgbits.RelocObj))
	obj := pkg.Scope().Lookup(name)

	targs := make([]types.Type, r.Len())
	for i := range targs {
		targs[i] = r.typ()
	}

	return obj, targs
}

func (pr *pkgReader2) objIdx(idx int) (*types.Package, string) {
	rname := pr.newReader(pkgbits.RelocName, idx, pkgbits.SyncObject1)

	objPkg, objName := rname.qualifiedIdent()
	assert(objName != "")

	tag := pkgbits.CodeObj(rname.Code(pkgbits.SyncCodeObj))

	if tag == pkgbits.ObjStub {
		if objPkg != nil && objPkg != types.Unsafe {
			fmt.Printf("who dis? %v\n", objPkg)
		}
		assert(objPkg == nil || objPkg == types.Unsafe)
		return objPkg, objName
	}

	if objPkg.Scope().Lookup(objName) == nil {
		// fmt.Println(objPkg, objName)

		dict := pr.objDictIdx(idx)

		r := pr.newReader(pkgbits.RelocObj, idx, pkgbits.SyncObject1)
		r.dict = dict

		declare := func(obj types.Object) {
			objPkg.Scope().Insert(obj)
		}

		switch tag {
		default:
			panic("weird")

		case pkgbits.ObjAlias:
			pos := r.pos()
			typ := r.typ()
			declare(types.NewTypeName(pos, objPkg, objName, typ))

		case pkgbits.ObjConst:
			pos := r.pos()
			typ := r.typ()
			val := r.Value()
			declare(types.NewConst(pos, objPkg, objName, typ, val))

		case pkgbits.ObjFunc:
			pos := r.pos()
			tparams := r.typeParamNames()
			sig := r.signature(nil, nil, tparams)
			declare(types.NewFunc(pos, objPkg, objName, sig))

		case pkgbits.ObjType:
			pos := r.pos()

			obj := types.NewTypeName(pos, objPkg, objName, nil)
			named := types.NewNamed(obj, nil, nil)
			declare(obj)

			named.SetTypeParams(r.typeParamNames())

			// TODO(mdempsky): Rewrite receiver types to underlying is an
			// Interface? The go/types importer does this (I think because
			// unit tests expected that), but cmd/compile doesn't care
			// about it, so maybe we can avoid worrying about that here.
			named.SetUnderlying(r.typ().Underlying())

			for i, n := 0, r.Len(); i < n; i++ {
				named.AddMethod(r.method())
			}

		case pkgbits.ObjVar:
			pos := r.pos()
			typ := r.typ()
			declare(types.NewVar(pos, objPkg, objName, typ))
		}
	}

	return objPkg, objName
}

func (pr *pkgReader2) objDictIdx(idx int) *reader2Dict {
	r := pr.newReader(pkgbits.RelocObjDict, idx, pkgbits.SyncObject1)

	var dict reader2Dict

	if implicits := r.Len(); implicits != 0 {
		errorf("unexpected object with %v implicit type parameter(s)", implicits)
	}

	dict.bounds = make([]typeInfo, r.Len())
	for i := range dict.bounds {
		dict.bounds[i] = r.typInfo()
	}

	dict.derived = make([]derivedInfo, r.Len())
	dict.derivedTypes = make([]types.Type, len(dict.derived))
	for i := range dict.derived {
		dict.derived[i] = derivedInfo{r.Reloc(pkgbits.RelocType), r.Bool()}
	}

	// function references follow, but reader2 doesn't need those

	return &dict
}

func (r *reader2) typeParamNames() []*types.TypeParam {
	r.Sync(pkgbits.SyncTypeParamNames)

	// Note: This code assumes it only processes objects without
	// implement type parameters. This is currently fine, because
	// reader2 is only used to read in exported declarations, which are
	// always package scoped.

	if len(r.dict.bounds) == 0 {
		return nil
	}

	// Careful: Type parameter lists may have cycles. To allow for this,
	// we construct the type parameter list in two passes: first we
	// create all the TypeNames and TypeParams, then we construct and
	// set the bound type.

	r.dict.tparams = make([]*types.TypeParam, len(r.dict.bounds))
	for i := range r.dict.bounds {
		pos := r.pos()
		pkg, name := r.localIdent()

		tname := types.NewTypeName(pos, pkg, name, nil)
		r.dict.tparams[i] = types.NewTypeParam(tname, nil)
	}

	for i, bound := range r.dict.bounds {
		r.dict.tparams[i].SetConstraint(r.p.typIdx(bound, r.dict))
	}

	return r.dict.tparams
}

func (r *reader2) method() *types.Func {
	r.Sync(pkgbits.SyncMethod)
	pos := r.pos()
	pkg, name := r.selector()

	rparams := r.typeParamNames()
	sig := r.signature(r.param(), rparams, nil)

	_ = r.pos() // TODO(mdempsky): Remove; this is a hacker for linker.go.
	return types.NewFunc(pos, pkg, name, sig)
}

func (r *reader2) qualifiedIdent() (*types.Package, string) { return r.ident(pkgbits.SyncSym) }
func (r *reader2) localIdent() (*types.Package, string)     { return r.ident(pkgbits.SyncLocalIdent) }
func (r *reader2) selector() (*types.Package, string)       { return r.ident(pkgbits.SyncSelector) }

func (r *reader2) ident(marker pkgbits.SyncMarker) (*types.Package, string) {
	r.Sync(marker)
	return r.pkg(), r.String()
}
