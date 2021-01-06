package noder

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
)

func check1(noders []*noder) {
	for _, p := range noders {
		p.node()
		p.file = nil // release memory
	}

	if base.SyntaxErrors() != 0 {
		base.ErrorExit()
	}
	types.CheckDclstack()

	for _, p := range noders {
		p.processPragmas()
	}

	// Typecheck.
	types.LocalPkg.Height = myheight
	typecheck.DeclareUniverse()
	typecheck.TypecheckAllowed = true

	// Process top-level declarations in phases.

	// Phase 1: const, type, and names and types of funcs.
	//   This will gather all the information about types
	//   and methods but doesn't depend on any of it.
	//
	//   We also defer type alias declarations until phase 2
	//   to avoid cycles like #18640.
	//   TODO(gri) Remove this again once we have a fix for #25838.

	// Don't use range--typecheck can add closures to Target.Decls.
	base.Timer.Start("fe", "typecheck", "top1")
	for i := 0; i < len(typecheck.Target.Decls); i++ {
		n := typecheck.Target.Decls[i]
		if op := n.Op(); op != ir.ODCL && op != ir.OAS && op != ir.OAS2 && (op != ir.ODCLTYPE || !n.(*ir.Decl).X.Alias()) {
			typecheck.Target.Decls[i] = typecheck.Stmt(n)
		}
	}

	// Phase 2: Variable assignments.
	//   To check interface assignments, depends on phase 1.

	// Don't use range--typecheck can add closures to Target.Decls.
	base.Timer.Start("fe", "typecheck", "top2")
	for i := 0; i < len(typecheck.Target.Decls); i++ {
		n := typecheck.Target.Decls[i]
		if op := n.Op(); op == ir.ODCL || op == ir.OAS || op == ir.OAS2 || op == ir.ODCLTYPE && n.(*ir.Decl).X.Alias() {
			typecheck.Target.Decls[i] = typecheck.Stmt(n)
		}
	}

	// Phase 3: Type check function bodies.
	// Don't use range--typecheck can add closures to Target.Decls.
	base.Timer.Start("fe", "typecheck", "func")
	var fcount int64
	for i := 0; i < len(typecheck.Target.Decls); i++ {
		n := typecheck.Target.Decls[i]
		if n.Op() == ir.ODCLFUNC {
			typecheck.FuncBody(n.(*ir.Func))
			fcount++
		}
	}

	// Phase 4: Check external declarations.
	// TODO(mdempsky): This should be handled when type checking their
	// corresponding ODCL nodes.
	base.Timer.Start("fe", "typecheck", "externdcls")
	for i, n := range typecheck.Target.Externs {
		if n.Op() == ir.ONAME {
			typecheck.Target.Externs[i] = typecheck.Expr(typecheck.Target.Externs[i])
		}
	}

	// Phase 5: With all user code type-checked, it's now safe to verify map keys.
	// With all user code typechecked, it's now safe to verify unused dot imports.
	typecheck.CheckMapKeys()
	CheckDotImports()
	base.ExitIfErrors()
}
