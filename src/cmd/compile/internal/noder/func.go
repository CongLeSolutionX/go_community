// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

func (g *irgen) funcBody(fn *ir.Func, recv *syntax.Field, sig *syntax.FuncType, block *syntax.BlockStmt) {
	outerfn := ir.CurFunc
	ir.CurFunc = fn

	// TODO(mdempsky): Remove this.
	typecheck.Target.Decls = append(typecheck.Target.Decls, fn)

	typ := fn.Type()
	if param := typ.Recv(); param != nil {
		g.defParam(param, recv, ir.PPARAM)
	}
	for i, param := range typ.Params().FieldSlice() {
		g.defParam(param, sig.ParamList[i], ir.PPARAM)
	}
	for i, result := range typ.Results().FieldSlice() {
		g.defParam(result, sig.ResultList[i], ir.PPARAMOUT)
	}

	// We may have type-checked a call to this function already and
	// calculated its size, including parameter offsets. Now that we've
	// created the parameter Names, force a recalculation to ensure
	// their offsets are correct.
	typ.Align = 0
	types.CalcSize(typ)

	if block != nil {
		typecheck.DeclContext = ir.PAUTO

		fn.Body = g.stmts(block.List)
		if fn.Body == nil {
			fn.Body = []ir.Node{ir.NewBlockStmt(src.NoXPos, nil)}
		}
		fn.Endlineno = g.makeXPos(block.Rbrace)

		if base.Flag.Dwarf {
			g.recordScopes(fn, sig)
		}
	}

	ir.CurFunc = outerfn
}

func (g *irgen) defParam(param *types.Field, decl *syntax.Field, class ir.Class) {
	typecheck.DeclContext = class

	var name *ir.Name
	if decl.Name != nil {
		name, _ = g.def(decl.Name)
	} else if class == ir.PPARAMOUT {
		name = g.obj(g.info.Implicits[decl])
	}

	if name != nil {
		param.Nname = name
		param.Sym = name.Sym() // in case it was renamed
	}
}
