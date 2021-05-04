// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types2

import (
	"cmd/compile/internal/syntax"
)

const UseInterface2 = true // exported so tests can use it

func noInterface2() {
	if UseInterface2 {
		panic("should not call this function when using Interface2")
	}
}

var emptyface Type

func init() {
	if UseInterface2 {
		emptyface = new(Interface2)
	} else {
		emptyface = &emptyInterface
	}
}

type Interface2 struct {
	methods []*Func // explicitly declared methods
	types   []Type  // explicitly embedded types
	obj     Object  // declaring object, or nil

	flattened *Interface2 // flattened interface, lazily computed
}

func NewInterface2(methods []*Func, types []Type) *Interface2 {
	sortMethods(methods)
	t := new(Interface2)
	t.methods = methods
	t.types = types
	return t
}

func (t *Interface2) Underlying() Type { return t }
func (t *Interface2) String() string   { return TypeString(t, nil) }

func (t *Interface2) Methods() []*Func {
	return t.methods
}

func (t *Interface2) Types() []Type {
	return t.types
}

func (t *Interface2) flat() *Interface2 {
	return t.flatten(nil)
}

func (t *Interface2) Flat() *Interface2 {
	return t.flat()
}

func (t *Interface2) empty() bool {
	return len(t.flat().methods) == 0 && len(t.flat().types) == 0
}

// hasType reports whether T is in the type set of interface t.
func (t *Interface2) hasType(T Type) bool {
	if len(t.types) == 0 {
		return true
	}
	for _, e := range t.types {
		if e, _ := e.(hasTyper); e != nil {
			if !e.hasType(T) {
				return false
			}
			continue
		}
		if !Identical(e, T) {
			return false
		}
	}
	return false
}

// IsComparable reports whether interface t is or embeds the predeclared interface "comparable".
func (ityp *Interface2) IsComparable() bool {
	return ityp.hasMethod(nil, "==")
}

func (ityp *Interface2) IsConstraint() bool {
	return ityp.IsComparable() || len(ityp.flat().types) > 0
}

// is reports whether interface ityp represents types that all satisfy pred.
func (ityp *Interface2) is(pred func(Type) bool) bool {
	for _, e := range ityp.types {
		if e, _ := e.(iser); e != nil {
			if e.is(pred) {
				return true
			}
			continue
		}
		if pred(e) {
			return true
		}
	}
	return false
}

func (ityp *Interface2) hasMethod(pkg *Package, name string) bool {
	_, m := ityp.lookupMethod(pkg, name)
	return m != nil
}

func (ityp *Interface2) lookupMethod(pkg *Package, name string) (int, *Func) {
	return lookupMethod(ityp.flat().methods, pkg, name)
}

func (ityp *Interface2) parse(check *Checker, expr *syntax.InterfaceType, def *Named) {
	var tlist []syntax.Expr
	var tname *syntax.Name // most recent "type" name

	for _, f := range expr.MethodList {
		if f.Name != nil {
			// We have a method with name f.Name, or a type
			// of a type list (f.Name.Value == "type").
			name := f.Name.Value
			if name == "_" {
				if check.conf.CompilerErrorMessages {
					check.error(f.Name, "methods must have a unique non-blank name")
				} else {
					check.error(f.Name, "invalid method name _")
				}
				continue // ignore
			}

			if name == "type" {
				// For now, collect all type list entries as if it
				// were a single union, where each union element is
				// of the form ~T.
				// TODO(gri) remove once we disallow type lists
				op := new(syntax.Operation)
				// We should also set the position (but there is no setter);
				// we don't care because this code will eventually go away.
				op.Op = syntax.Tilde
				op.X = f.Type
				tlist = append(tlist, op)
				if tname != nil && tname != f.Name {
					check.error(f.Name, "cannot have multiple type lists in an interface")
				}
				tname = f.Name
				continue
			}

			ftyp := check.typ(f.Type)
			sig, _ := ftyp.(*Signature)
			if sig == nil {
				if ftyp != Typ[Invalid] {
					check.errorf(f.Type, invalidAST+"%s is not a method signature", ftyp)
				}
				continue // ignore
			}

			// Always type-check method type parameters but complain if they are not enabled.
			// (This extra check is needed here because interface method signatures don't have
			// a receiver specification.)
			if sig.tparams != nil && !acceptMethodTypeParams {
				check.error(f.Type, "methods cannot have type parameters")
			}

			// use named receiver type if available (for better error messages)
			var recvTyp Type = ityp
			if def != nil {
				recvTyp = def
			}
			sig.recv = NewVar(f.Name.Pos(), check.pkg, "", recvTyp)

			m := NewFunc(f.Name.Pos(), check.pkg, name, sig)
			check.recordDef(f.Name, m)
			ityp.methods = append(ityp.methods, m)
			continue
		}

		// We have an embedded type (possibly a union of types).
		ityp.types = append(ityp.types, parseUnion(check, flattenUnion(nil, f.Type)))
		check.posMap2[ityp] = append(check.posMap2[ityp], f.Type.Pos())
	}

	// If we saw a type list, add it like an embedded union.
	if tlist != nil {
		ityp.types = append(ityp.types, parseUnion(check, tlist))
		// Types T in a type list are added as ~T expressions but we
		// don't have the position of the '~'. Use the type position
		// instead.
		check.posMap2[ityp] = append(check.posMap2[ityp], tlist[0].(*syntax.Operation).X.Pos())
	}

	// sort for API stability
	sortMethods(ityp.methods)
	sortTypes(ityp.types)

	check.later(func() {
		ityp.flatten(check)
	})
}

func flattenUnion(list []syntax.Expr, x syntax.Expr) []syntax.Expr {
	if o, _ := x.(*syntax.Operation); o != nil && o.Op == syntax.Or {
		list = flattenUnion(list, o.X)
		x = o.Y
	}
	return append(list, x)
}

func (ityp *Interface2) flatten(check *Checker) *Interface2 {
	// TODO(gri) We need to run with check if we haven't done so.

	if ityp.flattened != nil {
		return ityp.flattened
	}

	ftyp := new(Interface2)
	ityp.flattened = ftyp

	// Methods of embedded types are collected unchanged; i.e., the identity
	// of a method T.m's Func Object of an type T is the same as that of the
	// method m in an interface that embeds type T. On the other hand, if a
	// method is embedded via multiple overlapping embedded types, we don't
	// provide a guarantee which "original m" got chosen for the embedding
	// interface. See also issue #34421.
	//
	// If we don't care to provide this identity guarantee anymore, instead
	// of reusing the original method in embeddings, we can clone the method's
	// Func Object and give it the position of a corresponding embedded interface.
	// Then we can get rid of the mpos map below and simply use the cloned method's
	// position.

	var seen objset
	mpos := make(map[*Func]syntax.Pos) // method specification or method embedding position, for good error messages
	addMethod := func(pos syntax.Pos, m *Func, explicit bool) {
		switch other := seen.insert(m); {
		case other == nil:
			ftyp.methods = append(ftyp.methods, m)
			mpos[m] = pos
		case explicit:
			if check == nil {
				break
			}
			var err error_
			err.errorf(pos, "duplicate method %s", m.name)
			err.errorf(mpos[other.(*Func)], "other declaration of %s", m.name)
			check.report(&err)
		default:
			if check == nil {
				break
			}
			// We have a duplicate method name in an embedded (not explicitly declared) method.
			// If we're pre-go1.14 (overlapping embeddings are not permitted), report that
			// error here as well because it's the same error message.
			if !check.allowVersion(m.pkg, 1, 14) || !check.identical(m.typ, other.Type()) {
				var err error_
				err.errorf(pos, "duplicate method %s", m.name)
				err.errorf(mpos[other.(*Func)], "other declaration of %s", m.name)
				check.report(&err)
			}
		}
	}

	for _, m := range ityp.methods {
		addMethod(m.pos, m, true)
	}

	var posList []syntax.Pos
	if check != nil {
		posList = check.posMap2[ityp]
	}
	for i, typ := range ityp.types {
		var pos syntax.Pos
		if check != nil {
			pos = posList[i] // embedding position
		}
		utyp := under(typ)
		etyp := asInterface2(utyp)
		if etyp == nil {
			// Unions are the equivalent of type lists for now.
			// Accept them for now.
			// TODO(gri) this needs to be cleaned up.
			if UseInterface2 && asUnion(utyp) != nil {
				ftyp.types = append(ftyp.types, typ)
				continue
			}
			if check != nil && utyp != Typ[Invalid] && !check.allowVersion(check.pkg, 1, 18) {
				var format string
				if _, ok := utyp.(*TypeParam); ok {
					format = "%s is a type parameter, not an interface"
				} else {
					format = "%s is not an interface"
				}
				check.errorf(pos, format, typ)
			}
			continue
		}
		for _, m := range etyp.flat().methods {
			addMethod(pos, m, false) // use embedding position pos rather than m.pos
		}
		//ftyp.types = append(ftyp.types, typ)
	}

	sortMethods(ftyp.methods)
	sortTypes(ftyp.types)

	return ftyp
}
