// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements type-checking of contracts.

package types

import (
	"go/ast"
	"go/token"
	"sort"
)

// TODO(gri) Handling a contract like a type is problematic because it
// won't exclude a contract where we only permit a type. Investigate.

func (check *Checker) contractType(contr *Contract, e *ast.ContractType) {
	scope := NewScope(check.scope, token.NoPos, token.NoPos, "contract type parameters")
	check.scope = scope
	defer check.closeScope()
	check.recordScope(e, scope)

	// collect type parameters
	for index, name := range e.TParams {
		tpar := NewTypeName(name.Pos(), check.pkg, name.Name, nil)
		NewTypeParam(tpar, index) // assigns type to tpar as a side-effect
		check.declare(scope, name, tpar, scope.pos)
		contr.TParams = append(contr.TParams, tpar)
	}

	addMethod := func(tpar *TypeName, m *Func) {
		cs := contr.insert(tpar)
		iface := cs.Iface
		if iface == nil {
			iface = new(Interface)
			cs.Iface = iface
		}
		iface.methods = append(iface.methods, m)
	}
	_ = addMethod

	addType := func(tpar *TypeName, typ Type) {
		cs := contr.insert(tpar)
		// TODO(gri) should we complain about duplicate types?
		cs.Types = append(cs.Types, typ)
	}
	_ = addType

	// collect constraints
	for _, c := range e.Constraints {
		if c.Param != nil {
			// TODO(gri) update this code
			/*
				// If a type name is present, it must be one of the contract's type parameters.
				pos := c.Param.Pos()
				obj := scope.Lookup(c.Param.Name)
				if obj == nil {
					check.errorf(pos, "%s not declared by contract", c.Param.Name)
					continue
				}
				if c.Type == nil {
					check.invalidAST(pos, "missing method or type constraint")
					continue
				}
				tpar := obj.(*TypeName) // scope holds only *TypeNames
				typ := check.typ(c.Type)
				if c.MName != nil {
					// If a method name is present, it must be unique for the respective
					// type parameter, and c.Type is a method signature (guaranteed by AST).
					sig, _ := typ.(*Signature)
					if sig == nil {
						check.invalidAST(c.Type.Pos(), "invalid method type %s", typ)
					}
					// add receiver to signture (TODO(gri) do we need this? what's the "correct" receiver?)
					assert(sig.recv == nil)
					recvTyp := tpar.typ
					sig.recv = NewVar(pos, check.pkg, "", recvTyp)
					// make a method
					m := NewFunc(c.MName.Pos(), check.pkg, c.MName.Name, sig)
					addMethod(tpar, m)
				} else {
					// no method name => we have a type constraint
					var why string
					if !check.typeConstraint(typ, &why) {
						check.errorf(c.Type.Pos(), "invalid type constraint %s (%s)", typ, why)
						continue
					}
					addType(tpar, typ)
				}
			*/
		} else { // c.Param == nil
			// no type name => we have an embedded contract
			// A correct AST will have no method name and a single type that is an *ast.CallExpr in this case.
			if len(c.MNames) != 0 {
				check.invalidAST(c.MNames[0].Pos(), "no method (%s) expected with embedded contract declaration", c.MNames[0].Name)
				// ignore and continue
			}
			if len(c.Types) != 1 {
				check.invalidAST(e.Pos(), "contract contains incorrect (possibly embedded contract) entry")
				continue
			}
			// TODO(gri) we can probably get away w/o checking this (even if the AST is broken)
			econtr, _ := c.Types[0].(*ast.CallExpr)
			if econtr == nil {
				check.invalidAST(c.Types[0].Pos(), "invalid embedded contract %s", econtr)
			}
			etyp := check.typ(c.Types[0])
			_ = etyp
			// TODO(gri) complete this
			check.errorf(c.Types[0].Pos(), "%s: contract embedding not yet implemented", c.Types[0])
		}
	}

	// cleanup/complete interfaces
	// TODO(gri) should check for duplicate entries in first pass (no need for this extra pass)
	for tpar, cs := range contr.CMap {
		iface := cs.Iface
		if iface == nil {
			cs.Iface = &emptyInterface
		} else {
			var mset objset
			for _, m := range iface.methods {
				if m0 := mset.insert(m); m0 != nil {
					// A method with the same name exists already.
					// Complain if the signatures are different
					// but leave it in the method set.
					// TODO(gri) should we remove it from the set?
					// TODO(gri) factor out this functionality
					if !Identical(m0.Type(), m.Type()) {
						check.errorf(m.Pos(), "method %s already declared with different signature for %s", m.name, tpar.name)
						check.reportAltDecl(m0)
					}
				}
			}
			sort.Sort(byUniqueMethodName(iface.methods))
			iface.Complete()
		}
	}
}

// TODO(gri) does this simply check for the absence of defined types?
//           (if so, should choose a better name)
func (check *Checker) typeConstraint(typ Type, why *string) bool {
	switch t := typ.(type) {
	case *Basic:
		// ok
	case *Array:
		return check.typeConstraint(t.elem, why)
	case *Slice:
		return check.typeConstraint(t.elem, why)
	case *Struct:
		for _, f := range t.fields {
			if !check.typeConstraint(f.typ, why) {
				return false
			}
		}
	case *Pointer:
		return check.typeConstraint(t.base, why)
	case *Tuple:
		if t == nil {
			return true
		}
		for _, v := range t.vars {
			if !check.typeConstraint(v.typ, why) {
				return false
			}
		}
	case *Signature:
		if len(t.tparams) != 0 {
			panic("type parameter in function type")
		}
		return (t.recv == nil || check.typeConstraint(t.recv.typ, why)) &&
			check.typeConstraint(t.params, why) &&
			check.typeConstraint(t.results, why)
	case *Interface:
		for _, m := range t.allMethods {
			if !check.typeConstraint(m.typ, why) {
				return false
			}
		}
	case *Map:
		return check.typeConstraint(t.key, why) && check.typeConstraint(t.elem, why)
	case *Chan:
		return check.typeConstraint(t.elem, why)
	case *Named:
		*why = check.sprintf("%s is not a type literal", t)
		return false
	case *Contract:
		// TODO(gri) we shouldn't reach here
		*why = check.sprintf("%s is not a type", t)
		return false
	case *TypeParam:
		// ok
	default:
		unreachable()
	}
	return true
}
