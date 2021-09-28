// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/internal/src"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// covmgr maintains state for the coverage phase/pass during the walk
// of a given function.
type covmgr struct {
	ind   int
	unit  covRange
	units []covRange
}

// Func visits a function and take actions needed for compiler-based
// code coverage instrumentation.
func Func(fn *ir.Func) {
	// FIXME: at the moment we're skipping all init functions. At some
	// point we'll want to change things so that we cover user-written
	// init functions but not compiler-generated init functions.
	fname := ir.FuncName(fn)
	if fname == "init" || strings.HasPrefix(fname, "init.") {
		return
	}
	// TODO: figure out if there are parts of the runtime that are
	// unsafe to compile with coverage instrumentation.
	c := &covmgr{}
	c.Func(fn)
}

func (c *covmgr) verb(s string, a ...interface{}) {
	if base.Debug.NewCovDebug != 0 {
		fmt.Fprintf(os.Stderr, "=-= ")
		for i := 0; i < c.ind; i++ {
			fmt.Fprintf(os.Stderr, "  ")
		}
		fmt.Fprintf(os.Stderr, s, a...)
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// covRange corresponds roughly to a 'coverable unit', e.g. a basic-block
// like blob containing one or more executable statements.
type covRange struct {
	st      src.XPos // start position
	en      src.XPos // end position
	nxStmts uint     // number of executable statements within this range
}

func (r *covRange) String() string {
	return fmt.Sprintf("S:L%sC%s E:L%sC%s NX:%d",
		r.st.LineNumber(), r.st.ColumnNumber(),
		r.en.LineNumber(), r.en.ColumnNumber(), r.nxStmts)
}

//........................................................................

type stmtDisposition int

const (
	stmtIsExecutable = 1 << iota
	stmtHasStmtChildren
	stmtEndsBlock
)

func disposition(n ir.Node) stmtDisposition {
	switch n.Op() {
	case ir.ODCL, ir.ODCLCONST, ir.ODCLTYPE:
		return 0
	case ir.OAS, ir.OAS2, ir.OAS2DOTTYPE, ir.OAS2FUNC, ir.OAS2MAPR,
		ir.OAS2RECV, ir.OASOP, ir.ODEFER, ir.OGO, ir.OPRINT,
		ir.OPRINTN, ir.OCALLFUNC, ir.OCALLINTER, ir.OCLOSE, ir.ORECV,
		ir.ORECOVER, ir.ORECOVERFP, ir.OCOPY, ir.OSEND,
		ir.ODELETE:
		return stmtIsExecutable
	case ir.OIF, ir.OFOR, ir.OFORUNTIL, ir.OSELECT, ir.OSWITCH,
		ir.OTYPESW, ir.ORANGE:
		return stmtIsExecutable | stmtHasStmtChildren | stmtEndsBlock
	case ir.OBREAK, ir.OCONTINUE, ir.OFALL, ir.OGOTO, ir.ORETURN, ir.OPANIC:
		return stmtIsExecutable | stmtEndsBlock
	case ir.OLABEL:
		return stmtEndsBlock
	case ir.OBLOCK:
		return 0
	case ir.OFUNCINST:
		// Shouldn't see any of these given the current position in
		// the pipeline of the coverage phase.
		fallthrough
	default:
		panic(fmt.Sprintf("internal error: not handled: %+v\n", n))
	}
}

// appendPos incorporates the position of node 'n' into the range within
// the current coverable unit.
func (c *covmgr) appendPos(n ir.Node) {
	p := n.Pos()
	if c.unit.st == src.NoXPos {
		c.verb("update st pos for unit %d to %s", len(c.units), nodePosStr(n))
		c.unit.st = p
	}
	if p.After(c.unit.en) {
		c.verb("update en pos for unit %d to %s", len(c.units), nodePosStr(n))
		c.unit.en = p
	}
}

func posStr(xp src.XPos) string {
	p := base.Ctxt.PosTable.Pos(xp)
	b := filepath.Base(p.Filename())
	return fmt.Sprintf("{F=%s,L=%d,C=%d}", b, p.Line(), p.Col())
}

func nodePosStr(n ir.Node) string {
	return posStr(n.Pos())
}

// pnode dumps a node with src info for debugging purposes.
func (c *covmgr) pnode(n ir.Node) {
	if base.Debug.NewCovDebug != 0 {
		c.verb("pos: %s op='%v' n=%v", nodePosStr(n), n.Op(), n)
	}
}

// expr visits an expression tree, updating the current coverable
// unit's ending source position to reflect the source range of the
// expression. Example:
//
//  L10:  func foo() {
//  L11:     bar(101,
//  L12:         something(),
//  L13:         /* a comment */
//  L14:         q)
//  L15:  }
//
// The top level statement here is the CALLFUNC at line 11, however we
// would like the ending source position for the statement to be line
// 14, not line 11. Figuring this our requires walking into
// sub-expressions to visit their positions.
//
func (c *covmgr) expr(n ir.Node) {
	if n == nil {
		return
	}
	c.verb("expr(n) walk begin: pos: %s op='%v' n=%v", nodePosStr(n),
		n.Op(), n)
	c.appendPos(n)
	var vis func(cn ir.Node) bool
	vis = func(cn ir.Node) bool {
		if cn == nil || cn.Op() == ir.ONAME {
			return false
		}
		c.appendPos(cn)
		c.verb("expr walk pos: %s op='%v' n=%v", nodePosStr(cn), cn.Op(), cn)
		return ir.DoChildren(cn, vis)
	}
	ir.DoChildren(n, vis)
	c.verb("expr(n) walk end: op='%v' n=%v", n.Op(), n)
}

// exprList is similar to expr above, but handles a list of expressions.
func (c *covmgr) exprList(nn ir.Nodes) {
	for _, n := range nn {
		c.expr(n)
	}
}

// Func method visits a function for coverage instrumentation.
func (c *covmgr) Func(fn *ir.Func) {
	if base.Debug.NewCovDebug != 0 {
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	c.verb("fn %v", fn.Sym())

	if len(fn.Body) == 0 {
		c.verb("len(fn.Body) is 0")
		return
	}

	c.stmts(&fn.Body)

	if c.unit.nxStmts != 0 {
		c.units = append(c.units, c.unit)
	}

	c.verb("finished fn %v", fn.Sym())
	c.verb("coverableUnits:")
	for k, u := range c.units {
		c.verb("%d: %s", k, u.String())
	}
}

// stmts() visits a set of statements, collecting coverage meta
// data and inserting new statements with counter updates.
func (c *covmgr) stmts(nn *ir.Nodes) {
	c.ind++
	defer func() { c.ind-- }()

	for i, n := range *nn {
		c.verb("stmts: idx=%d op='%v' %v", i, n.Op(), n)
		c.pnode(n)
		c.stmt(n)
	}
}

// Visit examines the statement corresponding to node 'n',
func (c *covmgr) stmt(n ir.Node) {
	if len(n.Init()) != 0 {
		// At the moment it doesn't appear to be necessary to walk
		// the contents of an init from the perspective of picking
		// up additional statements, but we do want to consider the
		// contents of the init for its source position (if this
		// changes, see the remark below on "for" loops).
		c.exprList(n.Init())
	}
	disp := disposition(n)
	if disp&stmtIsExecutable != 0 {
		c.appendPos(n)
		c.unit.nxStmts++
	}
	if disp&stmtHasStmtChildren == 0 {
		c.expr(n)
	}
	endBlock := func() {
		if disp&stmtEndsBlock != 0 {
			if c.unit.nxStmts != 0 {
				c.units = append(c.units, c.unit)
			}
			c.unit = covRange{}
		}
	}
	switch n.Op() {
	case ir.ORANGE:
		n := n.(*ir.RangeStmt)
		c.expr(n.X)
		endBlock()
		c.stmts(&n.Body)
		endBlock()
	case ir.OFOR, ir.OFORUNTIL:
		unitIdx := len(c.units) - 1
		n := n.(*ir.ForStmt)
		c.expr(n.Cond)
		c.expr(n.Post)
		endBlock()
		c.stmts(&n.Body)
		if false {
			// If a 'for' loop has an initialized variable,
			// e.g. "for i := 0; ...", and the assignment forming the
			// initializer is counted as a separate statement (see init
			// above) then we will wind up with 2 executable statements
			// instead of 1. Undo the double counting here if needed.
			// [Maybe there is a less hacky way to handle this?]
			if len(n.Init()) != 0 {
				c.units[unitIdx].nxStmts--
			}
		}
		endBlock()
	case ir.OIF:
		n := n.(*ir.IfStmt)
		c.expr(n.Cond)
		endBlock()
		c.stmts(&n.Body)
		endBlock()
		c.stmts(&n.Else)
		endBlock()
	case ir.OSELECT:
		n := n.(*ir.SelectStmt)
		for _, cas := range n.Cases {
			endBlock()
			c.expr(cas.Comm)
			c.stmts(&cas.Body)
		}
		endBlock()
	case ir.OSWITCH, ir.OTYPESW:
		n := n.(*ir.SwitchStmt)
		for _, cas := range n.Cases {
			endBlock()
			c.exprList(cas.List)
			c.stmts(&cas.Body)
		}
		endBlock()
	case ir.OLABEL:
		// labels force the termination of the previous block.
		endBlock()
	case ir.OBREAK, ir.OFALL, ir.OGOTO, ir.ORETURN, ir.OPANIC, ir.OCONTINUE:
	default:
		if disp&stmtEndsBlock != 0 {
			fmt.Fprintf(os.Stderr, "endsblock not handled: %+v\n", n)
			panic("unhandled")
		}
	}
}
