// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements initialization and assignment checks.

package types

import (
	"errors"
	"go/token"
	itypes "internal/types"
)

// assignment reports whether x can be assigned to a variable of type T,
// if necessary by attempting to convert untyped values to the appropriate
// type. context describes the context in which the assignment takes place.
// Use T == nil to indicate assignment to an untyped blank identifier.
// x.mode is set to invalid if the assignment failed.
func (check *Checker) assignment(x *operand, T Type, context string) {
	check.singleValue(x)

	switch x.mode {
	case invalid:
		return // error reported before
	case constant_, variable, mapindex, value, commaok, commaerr:
		// ok
	default:
		unreachable()
	}

	if isUntyped(x.typ) {
		target := T
		// spec: "If an untyped constant is assigned to a variable of interface
		// type or the blank identifier, the constant is first converted to type
		// bool, rune, int, float64, complex128 or string respectively, depending
		// on whether the value is a boolean, rune, integer, floating-point,
		// complex, or string constant."
		if T == nil || IsInterface(T) {
			if T == nil && x.typ == Typ[UntypedNil] {
				check.errorf(x.pos(), "use of untyped nil in %s", context)
				x.mode = invalid
				return
			}
			target = Default(x.typ)
		}
		if err := check.canConvertUntyped(x, target); err != nil {
			var internalErr Error
			var msg string
			if errors.As(err, &internalErr) {
				msg = internalErr.Msg
			} else {
				msg = err.Error()
			}
			check.errorf(x.pos(), "cannot use %s as %s value in %s: %v", x, target, context, msg)
			x.mode = invalid
			return
		}
	}
	// x.typ is typed

	// spec: "If a left-hand side is the blank identifier, any typed or
	// non-constant value except for the predeclared identifier nil may
	// be assigned to it."
	if T == nil {
		return
	}

	if reason := ""; !x.assignableTo(check, T, &reason) {
		if reason != "" {
			check.errorf(x.pos(), "cannot use %s as %s value in %s: %s", x, T, context, reason)
		} else {
			check.errorf(x.pos(), "cannot use %s as %s value in %s", x, T, context)
		}
		x.mode = invalid
	}
}

func (check *Checker) initConst(lhs *Const, x *operand) {
	if x.mode == invalid || x.typ == Typ[Invalid] || lhs.typ == Typ[Invalid] {
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return
	}

	// rhs must be a constant
	if x.mode != constant_ {
		check.errorf(x.pos(), "%s is not constant", x)
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return
	}
	assert(isConstType(x.typ))

	// If the lhs doesn't have a type yet, use the type of x.
	if lhs.typ == nil {
		lhs.typ = x.typ
	}

	check.assignment(x, lhs.typ, "constant declaration")
	if x.mode == invalid {
		return
	}

	lhs.val = x.val
}

func (check *Checker) initVar(lhs *Var, x *operand, context string) Type {
	if x.mode == invalid || x.typ == Typ[Invalid] || lhs.typ == Typ[Invalid] {
		if lhs.typ == nil {
			lhs.typ = Typ[Invalid]
		}
		return nil
	}

	// If the lhs doesn't have a type yet, use the type of x.
	if lhs.typ == nil {
		typ := x.typ
		if isUntyped(typ) {
			// convert untyped types to default types
			if typ == Typ[UntypedNil] {
				check.errorf(x.pos(), "use of untyped nil in %s", context)
				lhs.typ = Typ[Invalid]
				return nil
			}
			typ = Default(typ)
		}
		lhs.typ = typ
	}

	check.assignment(x, lhs.typ, context)
	if x.mode == invalid {
		return nil
	}

	return x.typ
}

func (check *Checker) assignVar(lhs itypes.Expr, x *operand) Type {
	if x.mode == invalid || x.typ == Typ[Invalid] {
		return nil
	}

	// Determine if the lhs is a (possibly parenthesized) identifier.
	ident, _ := unparen(lhs).(itypes.Ident)

	// Don't evaluate lhs if it is the blank identifier.
	if ident != nil && ident.Name() == "_" {
		check.recordDef(ident, nil)
		check.assignment(x, nil, "assignment to _ identifier")
		if x.mode == invalid {
			return nil
		}
		return x.typ
	}

	// If the lhs is an identifier denoting a variable v, this assignment
	// is not a 'use' of v. Remember current value of v.used and restore
	// after evaluating the lhs via check.expr.
	var v *Var
	var vUsed bool
	if ident != nil {
		if obj := check.lookup(ident.Name()); obj != nil {
			// It's ok to mark non-local variables, but ignore variables
			// from other packages to avoid potential race conditions with
			// dot-imported variables.
			if w, _ := obj.(*Var); w != nil && w.pkg == check.pkg {
				v = w
				vUsed = v.used
			}
		}
	}

	var z operand
	check.expr(&z, lhs)
	if v != nil {
		v.used = vUsed // restore v.used
	}

	if z.mode == invalid || z.typ == Typ[Invalid] {
		return nil
	}

	// spec: "Each left-hand side operand must be addressable, a map index
	// expression, or the blank identifier. Operands may be parenthesized."
	switch z.mode {
	case invalid:
		return nil
	case variable, mapindex:
		// ok
	default:
		if sel, _ := z.expr.(itypes.SelectorExpr); sel != nil {
			var op operand
			check.expr(&op, sel.X())
			if op.mode == mapindex {
				check.errorf(z.pos(), "cannot assign to struct field %s in map", exprString(z.expr))
				return nil
			}
		}
		check.errorf(z.pos(), "cannot assign to %s", &z)
		return nil
	}

	check.assignment(x, z.typ, "assignment")
	if x.mode == invalid {
		return nil
	}

	return x.typ
}

// If returnPos is valid, initVars is called to type-check the assignment of
// return expressions, and returnPos is the position of the return statement.
func (check *Checker) initVars(lhs []*Var, rhs itypes.ExprList, returnPos itypes.Pos) {
	l := len(lhs)
	get, r, commaOk := unpack(func(x *operand, i int) { check.multiExpr(x, rhs.Expr(i)) }, rhs.Len(), l == 2 && !returnPos.IsValid())
	if get == nil || l != r {
		// invalidate lhs and use rhs
		for _, obj := range lhs {
			if obj.typ == nil {
				obj.typ = Typ[Invalid]
			}
		}
		if get == nil {
			return // error reported by unpack
		}
		check.useGetter(get, r)
		if returnPos.IsValid() {
			check.errorf(returnPos, "wrong number of return values (want %d, got %d)", l, r)
			return
		}
		check.errorf(rhs.Expr(0).Pos(), "cannot initialize %d variables with %d values", l, r)
		return
	}

	context := "assignment"
	if returnPos.IsValid() {
		context = "return statement"
	}

	x := operandPool.Get().(*operand)
	defer operandPool.Put(x)
	if commaOk {
		var a [2]Type
		for i := range a {
			get(x, i)
			a[i] = check.initVar(lhs[i], x, context)
		}
		check.recordCommaOkTypes(rhs.Expr(0), a)
		return
	}

	for i, lhs := range lhs {
		get(x, i)
		check.initVar(lhs, x, context)
	}
}

func (check *Checker) assignVars(stmt itypes.AssignStmt) {
	l := stmt.LhsLen()
	get, r, commaOk := unpack(func(x *operand, i int) { check.multiExpr(x, stmt.RhsExpr(i)) }, stmt.RhsLen(), l == 2)
	if get == nil {
		check.useLHS(stmt.Lhs())
		return // error reported by unpack
	}
	if l != r {
		check.useGetter(get, r)
		check.errorf(stmt.RhsExpr(0).Pos(), "cannot assign %d values to %d variables", r, l)
		return
	}

	x := operandPool.Get().(*operand)
	defer operandPool.Put(x)
	if commaOk {
		var a [2]Type
		for i := range a {
			get(x, i)
			a[i] = check.assignVar(stmt.LhsExpr(i), x)
		}
		check.recordCommaOkTypes(stmt.RhsExpr(0), a)
		return
	}

	for i := 0; i < stmt.LhsLen(); i++ {
		lhsx := stmt.LhsExpr(i)
		get(x, i)
		check.assignVar(lhsx, x)
	}
}

func (check *Checker) shortVarDecl(pos itypes.Pos, stmt itypes.AssignStmt) {
	top := len(check.delayed)
	scope := check.scope

	// collect lhs variables
	var newVars []*Var
	var lhsVars = make([]*Var, stmt.LhsLen())
	for i := 0; i < stmt.LhsLen(); i++ {
		lhsx := stmt.LhsExpr(i)
		var obj *Var
		if ident, _ := lhsx.(itypes.Ident); ident != nil {
			// Use the correct obj if the ident is redeclared. The
			// variable's scope starts after the declaration; so we
			// must use Scope.Lookup here and call Scope.Insert
			// (via check.declare) later.
			name := ident.Name()
			if alt := scope.Lookup(name); alt != nil {
				// redeclared object must be a variable
				if alt, _ := alt.(*Var); alt != nil {
					obj = alt
				} else {
					// TODO: confirm this error message is fixed.
					check.errorf(lhsx.Pos(), "cannot assign to %s", lhsx)
				}
				check.recordUse(ident, alt)
			} else {
				// declare new variable, possibly a blank (_) variable
				obj = newVar(ident.Pos(), check.pkg, name, nil)
				if name != "_" {
					newVars = append(newVars, obj)
				}
				check.recordDef(ident, obj)
			}
		} else {
			// TODO: this is leaky; we can't call the expr list wrapper here.
			check.useLHS(internalExprList{lhsx})
			// TODO: confirm this error message is fixed.
			check.errorf(lhsx.Pos(), "cannot declare %s", lhsx)
		}
		if obj == nil {
			obj = newVar(lhsx.Pos(), check.pkg, "_", nil) // dummy variable
		}
		lhsVars[i] = obj
	}

	check.initVars(lhsVars, stmt.Rhs(), token.NoPos)

	// process function literals in rhs expressions before scope changes
	check.processDelayed(top)

	// declare new variables
	if len(newVars) > 0 {
		// spec: "The scope of a constant or variable identifier declared inside
		// a function begins at the end of the ConstSpec or VarSpec (ShortVarDecl
		// for short variable declarations) and ends at the end of the innermost
		// containing block."
		scopePos := stmt.RhsExpr(stmt.RhsLen() - 1).End()
		for _, obj := range newVars {
			check.declare(scope, nil, obj, scopePos) // recordObject already called
		}
	} else {
		check.softErrorf(pos, "no new variables on left side of :=")
	}
}
