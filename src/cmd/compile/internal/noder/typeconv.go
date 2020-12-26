// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"fmt"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

var myexpr syntax.Expr
var myt2 types2.Type

// setTypes2Type tries to set the type of n based on the types2 type accessed via
// p.typ(expr).Type.  It also does a bunch of error checking.
func (p *noder) setTypes2Type(n ir.Node, expr syntax.Expr) {
	myexpr = expr
	t2 := p.typ(expr).Type
	myt2 = t2
	if t2 == nil {
		switch expr.(type) {
		case *syntax.Name:
			// Probably a package name or _ (blank)
			return
		case *syntax.KeyValueExpr:
			return
		case *syntax.TypeSwitchGuard:
			return
		case *syntax.SelectorExpr:
			println("Nil type for SelectorExpr, maybe OffsetOf?")
			return
		}
		panic("Unexpected nil type2 type for an expression")
	}
	if tp, ok := t2.(*types2.Tuple); ok {
		if tp == nil {
			// func call with no return values
			return
		}
	}
	// p.typeExpr2() may resolve n and set its type, if n is
	// an ONONAME. So, call typeExpr2() before looking at n.Type().
	nt := p.typeExpr2(t2, nil)
	ot := n.Type()

	if ot != nil && n.HasType2() {
		if nt != ot && !types.Identical(nt, ot) {
			// Because an untyped constant is declared, then used?
			panic(fmt.Sprintf("Setting types2 type twice not identical, op %s, %s -> %s", n.Op().String(), ot.String(), nt.String()))
		}
		// Can happen for NAME, NONAME, TYPE, or LITERAL, as they are
		// repeatedly referenced.
		//println("Setting types2 twice identically", n.Op().String(), nt.String())
		return
	}
	if ot != nil {
		if ot == nt {
			println("Setting types2 type fully equal to existing type", n.Op().String(), nt.String())
			// Can be LITERAL or TYPE. Must not make it types2 if
			// OTYPE, because we may have just imported this type
			// (converted from an ONONAME during typeExpr2()
			return
		}
		if n.Op() == ir.OLITERAL {
			if !ot.IsUntyped() {
				panic("Literal that is not untyped with mismatched type")
			}
			if typecheck.DefaultType(ot) == nt ||
				(ot == types.UntypedInt &&
					(nt.Kind() == types.TINT32 ||
						nt.Kind() == types.TINT64 ||
						nt.Kind() == types.TUINT32 ||
						nt.Kind() == types.TUINT64 ||
						nt.Kind() == types.TUINT ||
						nt.Kind() == types.TUINT8 ||
						nt.Kind() == types.TUINTPTR ||
						nt == types.ByteType)) ||
				(ot == types.UntypedFloat &&
					nt.Kind() == types.TFLOAT32) {
				println("Setting type2 type on literal that already has matching untyped type: ", ot.String(), "->", nt.String())
			} else {
				// If type is already set on a untyped literal,
				// don't change it, since we can make it
				// inconsistent (1e5 case).
				println("Not setting type2 type on literal that already has an untyped type: ", ot.String(), "->", nt.String())
				return
			}
		} else {
			panic(fmt.Sprintf("Double setting type for Op %s, %s -> %s",
				n.Op().String(), ot.String(), nt.String()))
		}
	}
	switch n.(type) {
	case *ir.FuncType, *ir.ArrayType, *ir.StructType, *ir.SliceType, *ir.InterfaceType, *ir.MapType, *ir.ChanType:
		// Don't bother setting on any of the type syntax nodes
	default:
		// Mark the type as having been derived from types2
		n.SetType2(nt)
	}
}

var etypes = [...]types.Kind{
	types2.Bool:          types.TBOOL,
	types2.Int:           types.TINT,
	types2.Int8:          types.TINT8,
	types2.Int16:         types.TINT16,
	types2.Int32:         types.TINT32,
	types2.Int64:         types.TINT64,
	types2.Uint:          types.TUINT,
	types2.Uint8:         types.TUINT8,
	types2.Uint16:        types.TUINT16,
	types2.Uint32:        types.TUINT32,
	types2.Uint64:        types.TUINT64,
	types2.Uintptr:       types.TUINTPTR,
	types2.Float32:       types.TFLOAT32,
	types2.Float64:       types.TFLOAT64,
	types2.Complex64:     types.TCOMPLEX64,
	types2.Complex128:    types.TCOMPLEX128,
	types2.String:        types.TSTRING,
	types2.UnsafePointer: types.TUNSAFEPTR,
}

// typeExpr2 converts a types2.Type to a types.Type. typ must be non-nil, and the
// return value will be non-nil. If forceRecv is non-nil, then we will insert that
// field for the receiver if we are converting a Signature.
func (p *noder) typeExpr2(typ types2.Type, forceRecv *types.Field) (r *types.Type) {
	defer func() {
		if r == nil {
			panic("Nil return from typeExpr2")
		}
	}()

	// Guaranteed to return non-nil
	fieldOf := func(v *types2.Var) *types.Field {
		vt := p.typeExpr2(v.Type(), nil)

		var s *types.Sym
		if name := v.Name(); name != "" {
			pkg := types.LocalPkg
			if v.Pkg().Name() != pkg.Name {
				pkg = types.NewPkg(v.Pkg().Path(), v.Pkg().Name())
			}
			s = pkg.Lookup(name)
		}
		pos := src.NoXPos
		// Pos may not be known for params/results for imported functions
		if v.Pos().IsKnown() {
			pos = p.makeXPos(v.Pos())
		}
		return types.NewField(pos, s, vt)
	}

	switch typ := typ.(type) {
	case *types2.Named:
		obj := typ.Obj()
		//println("named type", obj.Name())
		if obj.Pkg() == nil { /* universe */
			if obj.Name() == "error" {
				return types.ErrorType
			}
			panic(fmt.Sprintf("Unhandled universal named type: %s", obj.Name()))
		}

		if obj.Pkg().Name() == types.LocalPkg.Name { /* current package */
			def := oldname(typecheck.Lookup(obj.Name()))
			if def.Op() != ir.OTYPE {
				base.Fatalf("definition for %v is not a type: %v (%v)", obj, def, def.Op())
			}
			if def.Type() == nil {
				// TODO(danscales): If we don't have a type for
				// this local type name, we must be defining it,
				// hence it is using the type within the type def.
				// I have another change to deal with this.
				panic("Circular type")
			}
			return def.Type()
		}

		// Different package
		pkg := types.NewPkg(obj.Pkg().Path(), obj.Pkg().Name())
		sym := pkg.Lookup(obj.Name())
		def := ir.AsNode(sym.Def)
		if def == nil {
			// Pos will be filled in by Resolve()
			def = ir.NewIdent(src.NoXPos, sym)
			def = typecheck.Resolve(def)
		}
		if def.Op() != ir.OTYPE {
			base.Fatalf("definition for %v is not a type: %v (%v)", obj, def, def.Op())
		}
		if got := ir.AsNode(def.Type().Obj()); got != def {
			panic("Mismatch between def and def.Type()")
		}
		return def.Type()

	case *types2.Basic:
		if k := typ.Kind(); uint64(k) < uint64(len(etypes)) {
			if k == types2.Uint8 && typ.Name() == "byte" {
				// Distinguish byte from uint8 in displaying types
				return types.ByteType
			}
			if k == types2.Int32 && typ.Name() == "rune" {
				// Distinguish rune from int32 in displaying types
				return types.RuneType
			}
			if et := etypes[k]; et != 0 {
				return types.Types[et]
			}
		}
		switch typ.Kind() {
		case types2.UntypedInt:
			return types.UntypedInt
		case types2.UntypedRune:
			return types.UntypedRune
		case types2.UntypedFloat:
			return types.UntypedFloat
		case types2.UntypedComplex:
			return types.UntypedComplex
		case types2.UntypedBool:
			return types.UntypedBool
		case types2.UntypedString:
			return types.UntypedString
		case types2.UntypedNil:
			// TODO(danscales) - there is no UntypedNil, is TNIL fine?
			return types.Types[types.TNIL]
		case types2.Invalid:
			// This could be because this is part of an expression
			// that has already been evaluated to a constant (e.g.
			// len(a) where a is an array).
			println("Returning TNIL for an Invalid type from types2")
			return types.Types[types.TNIL]
		}
		base.Fatalf("DIDN'T HANDLE THIS BASIC TYPE: %v\n", typ)
	case *types2.Array:
		elem := p.typeExpr2(typ.Elem(), nil)
		return types.NewArray(elem, typ.Len())
	case *types2.Chan:
		elem := p.typeExpr2(typ.Elem(), nil)
		var dir types.ChanDir
		switch dir2 := typ.Dir(); dir2 {
		case types2.SendRecv:
			dir = types.Cboth
		case types2.SendOnly:
			dir = types.Csend
		case types2.RecvOnly:
			dir = types.Crecv
		default:
			base.Fatalf("unexpected dir2: %v", dir2)
		}
		return types.NewChan(elem, dir)
	case *types2.Signature:
		var recv *types.Field
		if forceRecv != nil {
			recv = forceRecv
		} else if r := typ.Recv(); r != nil {
			recv = fieldOf(r)
		}

		fieldsOf := func(tup *types2.Tuple) []*types.Field {
			fields := make([]*types.Field, tup.Len())
			for i := range fields {
				fields[i] = fieldOf(tup.At(i))
			}
			return fields
		}

		in := fieldsOf(typ.Params())
		out := fieldsOf(typ.Results())
		if typ.Variadic() {
			in[len(in)-1].SetIsDDD(true)
		}
		return types.NewSignature(types.LocalPkg, recv, in, out)
	case *types2.Map:
		key := p.typeExpr2(typ.Key(), nil)
		elem := p.typeExpr2(typ.Elem(), nil)
		return types.NewMap(key, elem)
	case *types2.Pointer:
		elem := p.typeExpr2(typ.Elem(), nil)
		return types.NewPtr(elem)
	case *types2.Slice:
		elem := p.typeExpr2(typ.Elem(), nil)
		return types.NewSlice(elem)
	case *types2.Struct:
		fields := make([]*types.Field, typ.NumFields())
		for i := range fields {
			v, tag := typ.Field(i), typ.Tag(i)
			f := fieldOf(v)
			f.Note = tag
			if v.Embedded() {
				f.Embedded = 1
			}
			fields[i] = f
		}
		t := types.NewStruct(types.LocalPkg, fields)
		types.CheckSize(t)
		return t

	case *types2.Interface:
		embeddeds := make([]*types.Field, typ.NumEmbeddeds())
		for i := range embeddeds {
			ft := p.typeExpr2(typ.EmbeddedType(i), nil)
			// TODO(mdempsky): Set f.Pos
			f := types.NewField(src.NoXPos, nil, ft)
			// TODO(mdempsky) need a sym?
			embeddeds[i] = f
		}

		methods := make([]*types.Field, typ.NumExplicitMethods())
		for i := range methods {
			fun := typ.ExplicitMethod(i)

			pos := p.makeXPos(fun.Pos())
			sym := typecheck.Lookup(fun.Name())
			ft := p.typeExpr2(fun.Type(),
				types.NewField(pos, nil, types.FakeRecvType()))
			f := types.NewField(pos, sym, ft)
			methods[i] = f
		}

		t := types.NewInterface(types.LocalPkg, append(embeddeds, methods...))

		// Ensure we expand the interface in the frontend (#25055).
		types.CheckSize(t)
		return t

	case *types2.Tuple:
		fields := make([]*types.Field, typ.Len())
		for i := range fields {
			v := typ.At(i)
			f := fieldOf(v)
			if v.Embedded() {
				f.Embedded = 1
			}
			fields[i] = f
		}
		t := types.NewStruct(types.LocalPkg, fields)
		types.CheckSize(t)
		// Can only set after doing the types.CheckSize()
		t.StructType().Funarg = types.FunargResults
		return t

	default:
		base.Fatalf("ANOTHER TYPE TO WORRY ABOUT: %T, %v\n", typ, typ)
	}
	panic("Missing a return in a case")
}
