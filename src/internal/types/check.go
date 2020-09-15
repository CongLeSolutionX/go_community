// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package types is an abstract type checker.
package types

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"
)

const (
	trace = false
)

type Scope interface {
	Insert(obj Object) Object
	Lookup(name string) Object
}

type Package interface {
	Name() string
	Path() string
	Scope() Scope
	SetName(name string)
}

// RemoteReference is an internal object that is referenced by a remote
// package.
type RemoteReference interface {
	Reference() interface{}
	SetReference(interface{})
}

type reference struct {
	reference interface{}
}

func (r reference) Reference() interface{} {
	return r.reference
}

func (r *reference) SetReference(ref interface{}) {
	r.reference = ref
}

type Driver interface {
	Universe() Scope
	Errorf(pos Pos, format string, args ...interface{})

	// Example of recording type information.
	RecordDef(iid Ident, obj Object)

	// Capture objects.
	NewScope(parent Scope, start, end Pos) Scope

	// Constructors for objects out of our control. Since this is currently just
	// used for 'addInvalid', it's possible that we could use a different
	// concrete implementation of the Ident interface.
	NewIdent(name string, pos Pos) Ident
}

type Checker struct {
	driver   Driver
	pkg      Package
	files    []File
	firstErr error
	indent   int

	objMap  map[Object]*declInfo
	objPath []Object
	delayed []func()
}

func (check *Checker) trace(pos Pos, format string, args ...interface{}) {
	fmt.Printf("%s:\t%s%s\n",
		pos,
		strings.Repeat(".  ", check.indent),
		fmt.Sprintf(format, args...),
	)
}

func newChecker(driver Driver, pkg Package) *Checker {
	return &Checker{
		driver: driver,
		pkg:    pkg,
		objMap: make(map[Object]*declInfo),
	}
}

// A bailout panic is used for early termination.
type bailout struct{}

func (check *Checker) handleBailout(err *error) {
	switch p := recover().(type) {
	case nil, bailout:
		// normal return or early exit
		*err = check.firstErr
	default:
		// re-panic
		panic(p)
	}
}

func CheckFiles(files []File, driver Driver, pkg Package) (err error) {
	check := newChecker(driver, pkg)
	defer check.handleBailout(&err)

	check.initFiles(files)
	check.collectObjects()
	check.packageObjects()
	check.processDelayed(0)

	return
}

// processDelayed processes all delayed actions pushed after top.
func (check *Checker) processDelayed(top int) {
	// If each delayed action pushes a new action, the
	// stack will continue to grow during this loop.
	// However, it is only processing functions (which
	// are processed in a delayed fashion) that may
	// add more actions (such as nested functions), so
	// this is a sufficiently bounded process.
	for i := top; i < len(check.delayed); i++ {
		check.delayed[i]() // may append to check.delayed
	}
	assert(top <= len(check.delayed)) // stack must not have shrunk
	check.delayed = check.delayed[:top]
}

func (check *Checker) errorf(pos Pos, format string, args ...interface{}) {
	if check.firstErr == nil {
		check.firstErr = fmt.Errorf(format, args...)
	}
	check.driver.Errorf(pos, format, args)
}

func (check *Checker) invalidAST(pos Pos, format string, args ...interface{}) {
	check.errorf(pos, "invalid AST: "+format, args...)
}

func (check *Checker) initFiles(files []File) {
	for _, file := range files {
		fName := file.Name()
		switch name := fName.Name(); check.pkg.Name() {
		case "":
			if name != "_" {
				check.pkg.SetName(name)
			} else {
				check.errorf(fName.Pos(), "invalid package name _")
			}
			fallthrough

		case name:
			check.files = append(check.files, file)

		default:
			check.errorf(file.Package(), "package %s; expected %s", name, check.pkg.Name())
		}
	}
}

func colorFor(t Type) color {
	if t != nil {
		return black
	}
	return white
}

type declInfo struct {
	file Scope // scope of file containing this declaration
	typ  Expr  // type, or nil
}

func (check *Checker) collectObjects() {
	pkg := check.pkg

	for _, file := range check.files {
		// The package identifier denotes the current package,
		// but there is no corresponding package object.
		// TODO
		check.driver.RecordDef(file.Name(), nil)

		// Use the actual source file extent rather than itypes.File extent since the
		// latter doesn't include comments which appear at the start or end of the file.
		// Be conservative and use the itypes.File extent if we don't have a *token.File.
		pos, end := file.Pos(), file.End()
		// NOTE: Skipped calculating the actual file extent.
		fileScope := check.driver.NewScope(check.pkg.Scope(), pos, end)
		// TODO
		// check.recordScope(file, fileScope)

		for decli := 0; decli < file.DeclsLen(); decli++ {
			decl := file.Decl(decli)
			switch d := decl.(type) {
			case BadDecl:
				// ignore

			case GenDecl:
				for iota := 0; iota < d.SpecsLen(); iota++ {
					spec := d.Spec(iota)
					switch s := spec.(type) {
					case TypeSpec:
						obj := NewTypeName(s.Name().Pos(), pkg, s.Name().Name(), nil)
						// check.declarePkgObj(s.Name(), obj, &declInfo{file: fileScope, typ: s.Type(), alias: s.Assign().IsKnown()})
						check.declarePkgObj(s.Name(), obj, &declInfo{file: fileScope, typ: s.Type()})

					default:
						check.errorf(spec.Pos(), "unknown ast.Spec node %T", spec)
					}
				}

			default:
				check.errorf(decl.Pos(), "unknown ast.Decl node %T", decl)
			}
		}
	}
}

func assert(cond bool) {
	if !cond {
		panic("assertion failed")
	}
}

func unreachable() {
	panic("unreachable")
}

func (check *Checker) declarePkgObj(ident Ident, obj Object, d *declInfo) {
	assert(ident.Name() == obj.Name())

	// spec: "A package-scope or file-scope identifier with name init
	// may only be declared to be a function with this (func()) signature."
	if ident.Name() == "init" {
		check.errorf(ident.Pos(), "cannot declare init - must be func")
		return
	}

	// spec: "The main package must have package name main and declare
	// a function main that takes no arguments and returns no value."
	if ident.Name() == "main" && check.pkg.Name() == "main" {
		check.errorf(ident.Pos(), "cannot declare main - must be func")
		return
	}

	check.declare(check.pkg.Scope(), ident, obj, nil)
	// hmm
	check.objMap[obj] = d
	obj.setOrder(uint32(len(check.objMap)))
}

func (check *Checker) declare(scope Scope, id Ident, obj Object, pos Pos) {
	// spec: "The blank identifier, represented by the underscore
	// character _, may be used in a declaration like any other
	// identifier but the declaration does not introduce a new
	// binding."
	if obj.Name() != "_" {
		if alt := scope.Insert(obj); alt != nil {
			check.errorf(obj.Pos(), "%s redeclared in this block", obj.Name())
			// check.reportAltDecl(alt)
			return
		}
		// TODO: move this into the scope.
		// obj.setScopePos(pos)
	}
	if id != nil {
		check.driver.RecordDef(id, obj)
	}
}

func (check *Checker) packageObjects() {
	// process package objects in source order for reproducible results
	objList := make([]Object, len(check.objMap))
	i := 0
	for obj := range check.objMap {
		objList[i] = obj
		i++
	}

	// phase 1
	for _, obj := range objList {
		check.objDecl(obj, nil)
	}
}

// pathString returns a string of the form a->b-> ... ->g for a path [a, b, ... g].
func pathString(path []Object) string {
	var s string
	for i, p := range path {
		if i > 0 {
			s += "->"
		}
		s += p.Name()
	}
	return s
}

func (check *Checker) objDecl(obj Object, def *Named) {
	if trace {
		check.trace(obj.Pos(), "-- checking %s (%s, objPath = %s)", obj, obj.color(), pathString(check.objPath))
		check.indent++
		defer func() {
			check.indent--
			check.trace(obj.Pos(), "=> %s (%s)", obj, obj.color())
		}()
	}

	// Checking the declaration of obj means inferring its type
	// (and possibly its value, for constants).
	// An object's type (and thus the object) may be in one of
	// three states which are expressed by colors:
	//
	// - an object whose type is not yet known is painted white (initial color)
	// - an object whose type is in the process of being inferred is painted grey
	// - an object whose type is fully inferred is painted black
	//
	// During type inference, an object's color changes from white to grey
	// to black (pre-declared objects are painted black from the start).
	// A black object (i.e., its type) can only depend on (refer to) other black
	// ones. White and grey objects may depend on white and black objects.
	// A dependency on a grey object indicates a cycle which may or may not be
	// valid.
	//
	// When objects turn grey, they are pushed on the object path (a stack);
	// they are popped again when they turn black. Thus, if a grey object (a
	// cycle) is encountered, it is on the object path, and all the objects
	// it depends on are the remaining objects on that path. Color encoding
	// is such that the color value of a grey object indicates the index of
	// that object in the object path.

	// During type-checking, white objects may be assigned a type without
	// traversing through objDecl; e.g., when initializing constants and
	// variables. Update the colors of those objects here (rather than
	// everywhere where we set the type) to satisfy the color invariants.
	if obj.color() == white && obj.Type() != nil {
		obj.setColor(black)
		return
	}

	switch obj.color() {
	case white:
		assert(obj.Type() == nil)
		// All color values other than white and black are considered grey.
		// Because black and white are < grey, all values >= grey are grey.
		// Use those values to encode the object's index into the object path.
		obj.setColor(grey + color(check.push(obj)))
		defer func() {
			check.pop().setColor(black)
		}()

	case black:
		assert(obj.Type() != nil)
		return

	default:
		// Color values other than white or black are considered grey.
		fallthrough

	case grey:
		// We have a cycle.
		// In the existing code, this is marked by a non-nil type
		// for the object except for constants and variables whose
		// type may be non-nil (known), or nil if it depends on the
		// not-yet known initialization value.
		// In the former case, set the type to Typ[Invalid] because
		// we have an initialization cycle. The cycle error will be
		// reported later, when determining initialization order.
		// TODO(gri) Report cycle here and simplify initialization
		// order code.
		switch obj := obj.(type) {
		case *TypeName:
			if check.cycle(obj) {
				// break cycle
				// (without this, calling underlying()
				// below may lead to an endless loop
				// if we have a cycle for a defined
				// (*Named) type)
				obj.typ = Typ[Invalid]
			}

		default:
			unreachable()
		}
		assert(obj.Type() != nil)
		return
	}

	d := check.objMap[obj]
	if d == nil {
		unreachable()
	}

	// Const and var declarations must not have initialization
	// cycles. We track them by remembering the current declaration
	// in check.decl. Initialization expressions depending on other
	// consts, vars, or functions, add dependencies to the current
	// check.decl.
	switch obj := obj.(type) {
	case *TypeName:
		// invalid recursive types are detected via path
		check.typeDecl(obj, d.typ, def)
	default:
		panic(fmt.Sprintf("unsupported object %T", obj))
	}
}

// push pushes obj onto the object path and returns its index in the path.
func (check *Checker) push(obj Object) int {
	check.objPath = append(check.objPath, obj)
	return len(check.objPath) - 1
}

// pop pops and returns the topmost object from the object path.
func (check *Checker) pop() Object {
	i := len(check.objPath) - 1
	obj := check.objPath[i]
	check.objPath[i] = nil
	check.objPath = check.objPath[:i]
	return obj
}

// TODO: process later
func (check *Checker) later(f func()) {
	check.delayed = append(check.delayed, f)
}

func (check *Checker) typeDecl(obj *TypeName, typ Expr, def *Named) {
	assert(obj.typ == nil)

	check.later(func() {
		check.validType(obj.typ, nil)
	})

	named := &Named{obj: obj}
	def.SetUnderlying(named)
	obj.typ = named // make sure recursive type declarations terminate

	// determine underlying type of named
	named.orig = check.definedType(typ, named)

	// The underlying type of named may be itself a named type that is
	// incomplete:
	//
	//	type (
	//		A B
	//		B *C
	//		C A
	//	)
	//
	// The type of C is the (named) type of A which is incomplete,
	// and which has as its underlying type the named type B.
	// Determine the (final, unnamed) underlying type by resolving
	// any forward chain.
	named.underlying = check.underlying(named)
}

// validType verifies that the given type does not "expand" infinitely
// producing a cycle in the type graph. Cycles are detected by marking
// defined types.
// (Cycles involving alias types, as in "type A = [10]A" are detected
// earlier, via the objDecl cycle detection mechanism.)
func (check *Checker) validType(typ Type, path []Object) typeInfo {
	const (
		unknown typeInfo = iota
		marked
		valid
		invalid
	)

	switch t := typ.(type) {

	case *Struct:
		for _, f := range t.fields {
			if check.validType(f.typ, path) == invalid {
				return invalid
			}
		}

	case *Named:
		// don't touch the type if it is from a different package or the Universe scope
		// (doing so would lead to a race condition - was issue #35049)
		if t.obj.pkg != check.pkg {
			return valid
		}

		// don't report a 2nd error if we already know the type is invalid
		// (e.g., if a cycle was detected earlier, via Checker.underlying).
		if t.underlying == Typ[Invalid] {
			t.info = invalid
			return invalid
		}

		switch t.info {
		case unknown:
			t.info = marked
			t.info = check.validType(t.orig, append(path, t.obj)) // only types of current package added to path
		case marked:
			// cycle detected
			for i, tn := range path {
				if t.obj.pkg != check.pkg {
					panic("internal error: type cycle via package-external type")
				}
				if tn == t.obj {
					check.cycleError(path[i:])
					t.info = invalid
					t.underlying = Typ[Invalid]
					return t.info
				}
			}
			panic("internal error: cycle start not found")
		}
		return t.info
	}

	return valid
}

func (check *Checker) definedType(e Expr, def *Named) (T Type) {
	if trace {
		check.trace(e.Pos(), "typeexpr %s", e)
		check.indent++
		defer func() {
			check.indent--
			check.trace(e.Pos(), "=> typeexpr %s", T)
		}()
	}

	T = check.typInternal(e, def)
	// assert(isTyped(T))
	// check.recordTypeAndValue(e, typexpr, T, nil)

	return
}

func (check *Checker) typInternal(expr Expr, def *Named) Type {
	switch e := expr.(type) {
	case BadExpr:
		// ignore - error reported before

	case Ident:
		// Only support builtin types.
		obj := check.driver.Universe().Lookup(e.Name())
		if obj == nil {
			check.errorf(e.Pos(), "unknown builtin %s", e.Name())
		} else {
			return obj.Type()
		}

	// TODO: remove...
	case ParenExpr:
		return check.definedType(e.X(), def)

	case StructType:
		typ := new(Struct)
		def.SetUnderlying(typ)
		check.structType(typ, e)
		return typ

	default:
		check.errorf(e.Pos(), "%s is not a type", e)
	}

	typ := Typ[Invalid]
	def.SetUnderlying(typ)
	return typ
}

func (check *Checker) underlying(typ Type) Type {
	// If typ is not a defined type, its underlying type is itself.
	n0, _ := typ.(*Named)
	if n0 == nil {
		return typ // nothing to do
	}

	// If the underlying type of a defined type is not a defined
	// type, then that is the desired underlying type.
	typ = n0.underlying
	n, _ := typ.(*Named)
	if n == nil {
		return typ // common case
	}

	// Otherwise, follow the forward chain.
	seen := map[*Named]int{n0: 0}
	path := []Object{n0.obj}
	for {
		typ = n.underlying
		n1, _ := typ.(*Named)
		if n1 == nil {
			break // end of chain
		}

		seen[n] = len(seen)
		path = append(path, n.obj)
		n = n1

		if i, ok := seen[n]; ok {
			// cycle
			check.cycleError(path[i:])
			typ = Typ[Invalid]
			break
		}
	}

	for n := range seen {
		// We should never have to update the underlying type of an imported type;
		// those underlying types should have been resolved during the import.
		// Also, doing so would lead to a race condition (was issue #31749).
		if n.obj.pkg != check.pkg {
			panic("internal error: imported type with unresolved underlying type")
		}
		n.underlying = typ
	}

	return typ
}

func (check *Checker) structType(styp *Struct, e StructType) {
	list := e.Fields()
	if list == nil {
		return
	}

	// struct fields and tags
	var fields []*Var
	var tags []string

	// for double-declaration checks
	var fset objset

	// current field typ and tag
	var typ Type
	var tag string
	add := func(ident Ident, embedded bool, pos Pos) {
		if tag != "" && tags == nil {
			tags = make([]string, len(fields))
		}
		if tags != nil {
			tags = append(tags, tag)
		}

		name := ident.Name()
		fld := NewField(pos, check.pkg, name, typ, embedded)
		// spec: "Within a struct, non-blank field names must be unique."
		if name == "_" || check.declareInSet(&fset, pos, fld) {
			fields = append(fields, fld)
			check.driver.RecordDef(ident, fld)
		}
	}

	// addInvalid adds an embedded field of invalid type to the struct for
	// fields with errors; this keeps the number of struct fields in sync
	// with the source as long as the fields are _ or have different names
	// (issue #25627).
	addInvalid := func(ident Ident, pos Pos) {
		typ = Typ[Invalid]
		tag = ""
		add(ident, true, pos)
	}

	for fi := 0; fi < list.Len(); fi++ {
		f := list.Field(fi)
		fType := f.Type()
		typ = check.typ(fType)
		tag = check.tag(f.Tag())
		if f.NamesLen() > 0 {
			// named fields
			for ni := 0; ni < f.NamesLen(); ni++ {
				name := f.Name(ni)
				add(name, false, name.Pos())
			}
		} else {
			// embedded field
			// spec: "An embedded type must be specified as a type name T or as a pointer
			// to a non-interface type name *T, and T itself may not be a pointer type."
			pos := fType.Pos()
			name := embeddedFieldIdent(fType)
			if name == nil {
				check.invalidAST(pos, "embedded field type %s has no name", fType)
				name = check.driver.NewIdent("_", pos)
				addInvalid(name, pos)
				continue
			}
			// Because we have a name, typ must be of the form T or *T, where T is the name
			// of a (named or alias) type, and t (= deref(typ)) must be the type of T.
			switch t := typ.Underlying().(type) {
			case *Basic:
				if t == Typ[Invalid] {
					// error was reported before
					addInvalid(name, pos)
					continue
				}
			}
			add(name, true, pos)
		}
	}

	styp.fields = fields
	styp.tags = tags
}

func embeddedFieldIdent(expr Expr) Ident {
	switch e := expr.(type) {
	case Ident:
		return e
	case StarExpr:
		// *T is valid, but **T is not
		if xx, _ := e.X().(StarExpr); xx == nil {
			return embeddedFieldIdent(e.X())
		}
	case SelectorExpr:
		return e.Sel()
	}
	return nil // invalid embedded field
}

func (check *Checker) typ(e Expr) Type {
	return check.definedType(e, nil)
}

func (check *Checker) declareInSet(oset *objset, pos Pos, obj Object) bool {
	if alt := oset.insert(obj); alt != nil {
		check.errorf(pos, "%s redeclared", obj.Name())
		// TODO
		// check.reportAltDecl(alt)
		return false
	}
	return true
}

func (check *Checker) tag(t BasicLit) string {
	if t != nil {
		if t.Kind() == token.STRING {
			if val, err := strconv.Unquote(t.Value()); err == nil {
				return val
			}
		}
		check.invalidAST(t.Pos(), "incorrect tag syntax: %q", t.Value)
	}
	return ""
}

// cycle checks if the cycle starting with obj is valid and
// reports an error if it is not.
func (check *Checker) cycle(obj Object) (isCycle bool) {
	// The object map contains the package scope objects and the non-interface methods.

	// Count cycle objects.
	assert(obj.color() >= grey)
	start := obj.color() - grey // index of obj in objPath
	cycle := check.objPath[start:]
	nval := 0 // number of (constant or variable) values in the cycle
	ndef := 0 // number of type definitions in the cycle
	for _, obj := range cycle {
		switch obj.(type) {
		case *TypeName:
			ndef++
		default:
			unreachable()
		}
	}

	if trace {
		check.trace(obj.Pos(), "## cycle detected: objPath = %s->%s (len = %d)", pathString(cycle), obj.Name(), len(cycle))
		check.trace(obj.Pos(), "## cycle contains: %d values, %d type definitions", nval, ndef)
		defer func() {
			if isCycle {
				check.trace(obj.Pos(), "=> error: cycle is invalid")
			}
		}()
	}

	// A cycle involving only constants and variables is invalid but we
	// ignore them here because they are reported via the initialization
	// cycle check.
	if nval == len(cycle) {
		return false
	}

	// A cycle involving only types (and possibly functions) must have at least
	// one type definition to be permitted: If there is no type definition, we
	// have a sequence of alias type names which will expand ad infinitum.
	if nval == 0 && ndef > 0 {
		return false // cycle is permitted
	}

	check.cycleError(cycle)

	return true
}

func (check *Checker) cycleError(cycle []Object) {
	// TODO(gri) Should we start with the last (rather than the first) object in the cycle
	//           since that is the earliest point in the source where we start seeing the
	//           cycle? That would be more consistent with other error messages.
	i := firstInSrc(cycle)
	obj := cycle[i]
	check.errorf(obj.Pos(), "illegal cycle in declaration of %s", obj.Name())
	for range cycle {
		check.errorf(obj.Pos(), "\t%s refers to", obj.Name()) // secondary error, \t indented
		i++
		if i >= len(cycle) {
			i = 0
		}
		obj = cycle[i]
	}
	check.errorf(obj.Pos(), "\t%s", obj.Name())
}

// firstInSrc reports the index of the object with the "smallest"
// source position in path. path must not be empty.
func firstInSrc(path []Object) int {
	fst, pos := 0, path[0].Pos()
	for i, t := range path[1:] {
		if t.Pos().Before(pos) {
			fst, pos = i+1, t.Pos()
		}
	}
	return fst
}
