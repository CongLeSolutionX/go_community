// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ir"
	"cmd/compile/internal/staticinit"
	"cmd/compile/internal/typecheck"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/objabi"
	"cmd/internal/src"
	"crypto/md5"
	"fmt"
	"internal/coverage"
	"internal/coverage/encodemeta"
	"os"
	"path/filepath"
	"strings"
)

// pkgIdVar is a package-level temp var to hold pkg ID
var pkgIdVar *ir.Name

// mdbuilder is a helper object for building/encoding meta-data for the package.
var mdbuilder *encodemeta.CoverageMetaDataBuilder

// mdname is a package-level readonly var holding meta-data for the pkg
var mdname *ir.Name

// init function for package, recorded during the initial walk
// and then used when we're finalizing the package.
var initfn *ir.Func

// covmgr maintains state for the coverage phase/pass during the walk
// of a given function.
type covmgr struct {
	ind      int // debug trace output indent level
	unit     covRange
	units    []covRange
	counters *ir.Name
	fnPos    src.XPos
}

// Func visits a function and take actions needed for compiler-based
// code coverage instrumentation.
func Func(fn *ir.Func) {
	// FIXME: at the moment we're skipping all init functions. At some
	// point we'll want to change things so that we cover user-written
	// init functions but not compiler-generated init functions.
	fname := ir.FuncName(fn)
	if fname == "init" || strings.HasPrefix(fname, "init.") {
		initfn = fn
		return
	}

	if mdbuilder == nil {
		dummy := coverage.PkgNonMod // temp
		var err error
		mdbuilder, err = encodemeta.NewCoverageMetaDataBuilder(base.Ctxt.Pkgpath, dummy)
		if err != nil {
			panic(fmt.Sprintf("creating meta-data builder: %v", err))
		}

		// Create the coverage meta-data symbol. This will be a
		// package-level, read-only symbol that is exported (so as to
		// allow it to be referred to by inlined routine bodies in
		// other packages). Assign the symbol a dummy type of
		// "[1]uint8"; later on we'll fix up the type to "[K]uint"
		// where K is the actual length of the underlying LSym's data.
		// Mark the symbol as local so as to ensure that the linker
		// doesn't emit a DWARF type DIE for it (since it is not a
		// real Go variable).
		mdname = typecheck.NewName(typecheck.Lookup(".covmeta"))
		dummyLength := int64(1)
		metaType := types.NewArray(types.Types[types.TUINT8], dummyLength)
		mdname.SetType(metaType)
		mdname.SetPos(fn.Pos())
		typecheck.Declare(mdname, ir.PEXTERN)
		typecheck.Export(mdname)
		mdname.MarkReadonly()
		mdsym := mdname.Linksym()
		mdsym.Set(obj.AttrStatic, true)
		mdsym.Set(obj.AttrLocal, true)
	}

	c := newCovMgr()
	c.Func(fn)
}

func newCovMgr() *covmgr {

	// TODO: make sure this works with set/counter/atomic modes.

	// The counter variable for each instrumented function is a flat
	// array of uint32 values: a set of three prolog values (containing
	// the number of counters, pkg ID, and func ID) then the actual
	// counter values themselves.
	//
	// We won't know how many counters we have until we're done
	// instrumenting the function, so we can't construct an accurate
	// type at this point. Fudge things instead: just treat the
	// counter var as an array of uint32 with a dummy length, then
	// eventually update it once we know how long it will be.
	//
	dummyLength := int64(1)
	ctrType := types.NewArray(types.Types[types.TUINT32], dummyLength)
	counters := staticinit.StaticName(ctrType)
	counters.SetCoverageCounter(true)
	counters.Linksym().Type = objabi.SCOVERAGE_COUNTER

	// Create a static var to hold the pkg ID. This will be filled
	// as part of pkg init (in a subsequent patch).
	hcid := coverage.HardCodedPkgId(base.Ctxt.Pkgpath)
	if hcid == coverage.NotHardCoded && pkgIdVar == nil {
		pkgIdVar = staticinit.StaticName(types.Types[types.TUINT32])
	}

	return &covmgr{
		counters: counters,
	}
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
		// FIXME: the legacy cmd/cover implementation seems to count
		// empty blocks as executable statements, with the exception
		// of the outermost function block. Examples:
		//
		// func A() { }
		// func B() { { { { } } } }
		//
		// For function "A", cmd/cover creates a counter to record
		// whether the function ever executes, but considers the
		// function to have zero stmts. For function B, cmd/cover
		// treats it has having three executable statements. Not clear
		// whether we want to replicate this behavior. Might also be
		// worth trying other coverage tools to see how they treat
		// these cases.
		//
		// For now, empty blocks are treated as having no executable
		// statements.
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
// would like the ending source position for the statement in the
// meta-data to be line 14, not line 11. Figuring this our requires
// walking into sub-expressions to visit their positions.
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
	if len(fn.Body) == 0 {
		// don't process assembly functions.
		return
	}

	c.verb("\n\nfn %s.%v", base.Ctxt.Pkgpath, fn.Sym())

	c.stmts(&fn.Body)

	// In the case of an empty function, we want to make sure we get a
	// single coverable unit with with nxStmts set to 0. From a
	// statement coverage perspective, an empty function contains no
	// statements (which would argue for no counters), but we'd like
	// to be able to detect that the function was called. Hence a
	// single coverable unit (and counter) with zero exec stmts.
	if c.unit.nxStmts != 0 {
		c.units = append(c.units, c.unit)
	} else if len(c.units) == 0 {
		fn.Body = append(fn.Body, c.counterGen(fn.Pos())...)
		c.units = append(c.units, c.unit)
	}

	c.verb("finished fn %v", fn.Sym())
	c.verb("coverableUnits:")
	for k, u := range c.units {
		c.verb("%d: %s", k, u.String())
	}

	funcId := c.recordMetaData(fn)
	c.fixup(fn, uint32(funcId), uint32(len(c.units)))

	c.verb("final bod: %+v", fn.Body)
}

// fixup updates the type of the counter array (now that the number of
// counters is known) and inserts code into the entry basic block to
// record the meta-data symbol, func ID, and number of counters into
// the initial portion of the counter array.
func (c *covmgr) fixup(fn *ir.Func, funcId uint32, nCtrs uint32) {

	bpsave := base.Pos
	defer func() { base.Pos = bpsave }()
	base.Pos = c.fnPos

	if nCtrs == 0 {
		// something went wrong -- even for empty functions we should have
		// a coverable unit.
		panic("bad")
	}

	// We now know the length of the counter array, so update its type.
	// update the type of the counter array.
	t := types.NewArray(types.Types[types.TUINT32], int64(nCtrs)+coverage.FirstCtrOffset)
	c.counters.SetType(t)

	// emit: counters[numCtrsOffet] = <num counters>
	ixc := ir.NewIndexExpr(c.fnPos, c.counters, ir.NewInt(coverage.NumCtrsOffset))
	ixc.SetBounded(true)
	assign1 := typecheck.Stmt(ir.NewAssignStmt(c.fnPos, ixc, ir.NewInt(int64(nCtrs))))

	// emit: counters[pkgIdOffset] = pkgIdVar
	var val ir.Node = pkgIdVar
	pkid := coverage.HardCodedPkgId(base.Ctxt.Pkgpath)
	if pkid != coverage.NotHardCoded {
		val = ir.NewInt(int64(pkid))
	}
	ixp := ir.NewIndexExpr(c.fnPos, c.counters, ir.NewInt(coverage.PkgIdOffset))
	ixp.SetBounded(true)
	assign2 := typecheck.Stmt(ir.NewAssignStmt(c.fnPos, ixp, val))

	// emit: counters[funcidOffset] = <func id>
	ixf := ir.NewIndexExpr(c.fnPos, c.counters, ir.NewInt(coverage.FuncIdOffset))
	ixf.SetBounded(true)
	assign3 := typecheck.Stmt(ir.NewAssignStmt(c.fnPos, ixf, ir.NewInt(int64(funcId))))

	// prepend to function body.
	fn.Body = append([]ir.Node{assign1, assign2, assign3}, fn.Body...)
}

// recordMetaData registers the function we've just visited with the
// meta-data builder. The builder adds an entry for the function in
// its data structures, so as to include it in the final meta-data
// symbol for the package.
func (c *covmgr) recordMetaData(fn *ir.Func) uint {
	cunits := make([]coverage.CoverableUnit, 0, len(c.units))
	for _, u := range c.units {
		cu := coverage.CoverableUnit{
			StLine:  uint32(u.st.Line()),
			StCol:   uint32(u.st.Col()),
			EnLine:  uint32(u.en.Line()),
			EnCol:   uint32(u.en.Col()),
			NxStmts: uint32(u.nxStmts),
		}
		cunits = append(cunits, cu)
	}
	xp := fn.Pos()
	fnpos := base.Ctxt.PosTable.Pos(xp)
	fname := fn.Linksym().Name
	if strings.HasPrefix(fname, types.LocalPkg.Prefix+".") {
		fname = fname[3:]
	}
	fd := coverage.FuncDesc{
		Funcname: fname,
		Srcfile:  fnpos.Filename(),
		Units:    cunits,
	}

	// record what we've seen with the meta-data builder
	return mdbuilder.AddFunc(fd)
}

func (c *covmgr) counterGen(pos src.XPos) []ir.Node {
	var rval []ir.Node

	k := len(c.units)

	// Emit code to set the counter for this unit, e.g.
	//
	//   counter[k] = 1
	//
	ix := ir.NewIndexExpr(pos, c.counters, ir.NewInt(int64(k+coverage.FirstCtrOffset)))
	ix.SetBounded(true)
	update := ir.NewAssignStmt(pos, ix, ir.NewInt(1))
	rval = append(rval, typecheck.Stmt(update))

	return rval
}

// appendInstr takes a statement node ("n") and a (possibly nil)
// instrumentation sequence "instr" and appends them to an output
// list. It includes a special case for this case
//
//    mylabel:
//      for .. {
//      }
//
// For the code above, we want to avoid adding a counter update
// between the label and the "for", since if this happens it will
// appear that the label is associated with the counter update and not
// with the for loop.
func appendInstr(n ir.Node, instr []ir.Node, out []ir.Node) []ir.Node {
	if (n.Op() == ir.OFOR || n.Op() == ir.OFORUNTIL) &&
		len(out) != 0 && out[len(out)-1].Op() == ir.OLABEL {
		lab := out[len(out)-1]
		out = out[:len(out)-1]
		out = append(out, instr...)
		out = append(out, lab)
	} else {
		out = append(out, instr...)
	}
	out = append(out, n)
	return out
}

// stmts() visits a set of statements, collecting coverage meta
// data and inserting new statements with counter updates.
func (c *covmgr) stmts(nn *ir.Nodes) {
	c.ind++
	defer func() { c.ind-- }()
	out := []ir.Node{}
	for i, n := range *nn {
		c.verb("stmts: idx=%d op='%v' %v", i, n.Op(), n)
		c.pnode(n)
		nx := c.stmt(n)
		if len(nx) != 0 {
			out = appendInstr(n, nx, out)
		} else {
			out = append(out, n)
		}
	}
	*nn = out
}

// Visit examines the statement corresponding to node 'n', optionally
// returning a new node corresponding to a counter update, if we
// decided to add a counter up for this node.
func (c *covmgr) stmt(n ir.Node) []ir.Node {
	var ctr []ir.Node
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
		if c.unit.nxStmts == 0 {
			ctr = c.counterGen(n.Pos())
		}
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
	return ctr
}

type symbolWriteSeeker struct {
	ctxt *obj.Link
	sym  *obj.LSym
	off  int64
}

func (d *symbolWriteSeeker) Write(p []byte) (n int, err error) {
	amt := len(p)
	d.sym.WriteBytes(d.ctxt, d.off, p)
	d.off += int64(amt)
	return amt, nil
}

func (d *symbolWriteSeeker) Seek(offset int64, whence int) (int64, error) {
	if whence == os.SEEK_SET {
		d.off = offset
		return offset, nil
	}
	// other modes not supported
	panic("bad")
}

// FinishPackage is called once we've visited all of the functions
// in the package; it finalizes the meta-data symbol for the package.
func FinishPackage() {
	if mdbuilder == nil {
		return
	}
	if initfn == nil {
		panic("missing init function")
	}

	// Q: what pos should we be using for this init code? Is this OK?
	pos := mdname.Pos()
	if len(initfn.Body) != 0 {
		pos = initfn.Body[0].Pos()
	} else if !initfn.Pos().IsKnown() {
		initfn.SetPos(pos)
	}

	// Write accumulated content to the meta-data symbol.
	mdsym := mdname.Linksym()
	writer := &symbolWriteSeeker{
		sym:  mdsym,
		ctxt: base.Ctxt,
	}
	mdbuilder.Emit(writer)
	mdbuilder = nil

	if base.Debug.NewCovDebug != 0 {
		fmt.Fprintf(os.Stderr, "=-= pkg=%s mdlen=%d sum=%s 4b=%x\n", base.Ctxt.Pkgpath, len(mdsym.P), fmt.Sprintf("%x", md5.Sum(mdsym.P)), mdsym.P[0:4])
	}

	// Now that we know the length, update the type of the meta-data symbol
	// to reflect reality.
	mdname.SetType(types.NewArray(types.Types[types.TUINT8], int64(len(mdsym.P))))

	// Materalize expression corresponding to address of the meta-data symbol.
	mdax := typecheck.NodAddr(mdname)
	mdauspx := typecheck.ConvNop(mdax, types.Types[types.TUNSAFEPTR])

	// Materialize expression for length.
	len := uint32(len(mdsym.P))
	lenx := ir.NewInt(int64(len)) // untyped

	// Materialize expression for hash (an array literal)
	hash := md5.Sum(mdsym.P)
	elist := make([]ir.Node, 0, 16)
	for i := 0; i < 16; i++ {
		elem := ir.NewInt(int64(hash[i]))
		elist = append(elist, elem)
	}
	ht := types.NewArray(types.Types[types.TUINT8], 16)
	hashx := ir.NewCompLitExpr(pos, ir.OCOMPLIT, ir.TypeNode(ht), elist)

	// Generate a call to runtime.addcovmeta, e.g.
	//
	//   pkgIdVar = runtime.addcovmeta(&sym, len, hash, pkgpath, pkid, cmode)
	//
	fn := typecheck.LookupRuntime("addcovmeta")
	pkid := coverage.HardCodedPkgId(base.Ctxt.Pkgpath)
	pkIdNode := ir.NewInt(int64(pkid))
	cmodeNode := ir.NewInt(int64(1)) // hard-coded for now
	pkPathNode := ir.NewString(base.Ctxt.Pkgpath)
	callx := typecheck.Call(pos, fn, []ir.Node{mdauspx, lenx, hashx,
		pkPathNode, pkIdNode, cmodeNode}, false)
	assign := callx
	if pkid == coverage.NotHardCoded {
		assign = typecheck.Stmt(ir.NewAssignStmt(pos, pkgIdVar, callx))
	}

	// Tack the call onto the start of our init function. We do this
	// early in the init since it's possible that instrumented function
	// bodies (with counter updates) might be inlined into init.
	initfn.Body.Prepend(assign)
}
