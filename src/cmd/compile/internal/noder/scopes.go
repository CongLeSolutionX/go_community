// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package noder

import (
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/dwarfgen"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/syntax"
	"cmd/compile/internal/types2"
	"cmd/internal/src"
)

func (g *irgen) claimScopes(fn *ir.Func, sig *syntax.FuncType) {
	scope, ok := g.info.Scopes[sig]
	if !ok {
		base.FatalfAt(fn.Pos(), "missing scope for %v", fn)
	}

	g.parents = g.parents[:0]
	g.marks = g.marks[:0]

	// Sentinel mark so we can omit bounds checks later.
	g.marks = append(g.marks, ir.Mark{Pos: src.NoXPos, Scope: 0})

	for i, n := 0, scope.NumChildren(); i < n; i++ {
		g.walkScope(0, scope.Child(i))
	}

	fn.Parents = make([]ir.ScopeID, len(g.parents))
	copy(fn.Parents, g.parents)

	fn.Marks = make([]ir.Mark, len(g.marks[1:]))
	copy(fn.Marks, g.marks[1:])
}

func (g *irgen) walkScope(parent ir.ScopeID, scope *types2.Scope) bool {
	// go/types doesn't provide a real API for determining the
	// lexical element a scope represents, so we have to resort to
	// string matching. Conveniently, this allows us to skip both
	// function types and function literals.
	if strings.HasPrefix(scope.String(), "function scope ") {
		return false
	}

	g.parents = append(g.parents, parent)
	this := ir.ScopeID(len(g.parents))

	g.mark(scope, g.pos(scope), this)

	haveVars := false
	for _, name := range scope.Names() {
		if obj, ok := scope.Lookup(name).(*types2.Var); ok && obj.Name() != "_" {
			haveVars = true
			break
		}
	}

	for i, n := 0, scope.NumChildren(); i < n; i++ {
		if g.walkScope(this, scope.Child(i)) {
			haveVars = true
		}
	}

	if haveVars {
		g.mark(scope, g.end(scope), parent)
	} else {
		g.retract(parent, this)
	}

	return haveVars
}

// markScope records that we transition into a new scope at pos.
func (g *irgen) mark(scope *types2.Scope, pos src.XPos, next ir.ScopeID) {
	if !pos.IsKnown() {
		base.Fatalf("unknown scope boundary for %v", scope)
	}

	last := &g.marks[len(g.marks)-1]
	if last.Pos.IsKnown() && dwarfgen.XPosBefore(pos, last.Pos) {
		base.FatalfAt(pos, "non-monotonic scope progression from %v at %v", scope, base.FmtPos(last.Pos))
	}

	if last.Pos == pos {
		last.Scope = next
	} else {
		g.marks = append(g.marks, ir.Mark{Pos: pos, Scope: next})
	}
}

// retract undoes a previous scope mark that turned out to be
// unnecessary due to not containing any variables.
func (g *irgen) retract(parent, this ir.ScopeID) {
	// First, release our scope ID for reuse by clearing
	// it's entry from the parents array.
	if this != ir.ScopeID(len(g.parents)) || g.parents[this-1] != parent {
		base.Fatalf("scope tracking inconsistency")
	}
	g.parents = g.parents[:this-1]

	last := len(g.marks) - 1
	if this != g.marks[last].Scope {
		base.Fatalf("retracted scope isn't the most recent")
	}

	// If removing the last mark returns us to our parent scope,
	// then just do that. Otherwise, update it in place.
	if g.marks[last-1].Scope == parent {
		g.marks = g.marks[:last]
	} else {
		g.marks[last].Scope = parent
	}
}
