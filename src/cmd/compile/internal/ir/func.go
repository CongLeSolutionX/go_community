// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/obj"
	"cmd/internal/src"
)

// A Func corresponds to a single function in a Go program
// (and vice versa: each function is denoted by exactly one *Func).
//
// There are multiple nodes that represent a Func in the IR.
//
// The ONAME node (Func.Name) is used for plain references to it.
// The ODCLFUNC node (Func.Decl) is used for its declaration code.
// The OCLOSURE node (Func.Closure) is used for a reference to a
// function literal.
//
// A Func for an imported function will have only an ONAME node.
// A declared top-level function or method has an ONAME and an ODCLFUNC.
// A function literal has an OCLOSURE and an ODCLFUNC; the latter
// is the compiled form that accessed the captured variables from
// a special data structure.
//
// A method declaration is represented like functions, except f.Sym
// will be the qualified method name (e.g., "T.m") and
// f.Func.Shortname is the bare method name (e.g., "m").
//
// A method expression (T.M) is represented as an OMETHEXPR node,
// in which n.Left and n.Right point to the type and method, respectively.
// Each distinct mention of a method expression in the source code
// constructs a fresh node.
//
// A method value (t.M) is represented by represented by
// ODOTMETH/ODOTINTER when it is called directly and by
// OCALLPART otherwise. These are like method expressions,
// except that for ODOTMETH/ODOTINTER, the method name
// is stored in Sym instead of Right.
//
// Each OCALLPART ends up being implemented as a new
// function, a bit like a closure, with its own ODCLFUNC.
// The OCALLPART has an n.Func, so the ODCLFUNC is
// n.Func.Decl.
type Func struct {
	Enter Nodes // setup at start of function

	Name          *Name    // ONAME node
	Decl          *DclFunc // ODCLFUNC node
	Closure       INode    // OCLOSURE node
	ClosureEnter  Nodes    // setup for closure body
	ClosureType   INode    // syntax for closure representation type
	ClosureCalled bool     // closure is only immediately called

	Shortname *types.Sym
	Exit      Nodes
	Cvars     Nodes   // closure params
	Dcl       []INode // autodcl for this func/closure

	// Parents records the parent scope of each scope within a
	// function. The root scope (0) has no parent, so the i'th
	// scope's parent is stored at Parents[i-1].
	Parents []ScopeID

	// Marks records scope boundary changes.
	Marks []Mark

	// Closgen tracks how many closures have been generated within
	// this function. Used by closurename for creating unique
	// function names.
	Closgen int

	FieldTrack map[*types.Sym]struct{}
	DebugInfo  *ssa.FuncDebug
	LSym       *obj.LSym

	Inl *Inline

	Label int32 // largest auto-generated label in this function

	Endlineno src.XPos
	WBPos     src.XPos // position of first write barrier; see SetWBPos

	Pragma PragmaFlag // go:xxx function annotations

	flags      bitset16
	NumDefers  int // number of defer calls in the function
	NumReturns int // number of explicit returns in the function

	// nwbrCalls records the LSyms of functions called by this
	// function for go:nowritebarrierrec analysis. Only filled in
	// if nowritebarrierrecCheck != nil.
	NWBRCalls *[]SymAndPos
}

// An Inline holds fields used for function bodies that can be inlined.
type Inline struct {
	Cost int32 // heuristic cost of inlining this function

	// Copies of Func.Dcl and Nbody for use during inlining.
	Dcl  []INode
	Body []INode
}

// A Mark represents a scope boundary.
type Mark struct {
	// Pos is the position of the token that marks the scope
	// change.
	Pos src.XPos

	// Scope identifies the innermost scope to the right of Pos.
	Scope ScopeID
}

type PragmaFlag int16

const (
	// Func pragmas.
	Nointerface    PragmaFlag = 1 << iota
	Noescape                  // func parameters don't escape
	Norace                    // func must not have race detector annotations
	Nosplit                   // func should not execute on separate stack
	Noinline                  // func should not be inlined
	NoCheckPtr                // func should not be instrumented by checkptr
	CgoUnsafeArgs             // treat a pointer to one arg as a pointer to them all
	UintptrEscapes            // pointers converted to uintptr escape

	// Runtime-only func pragmas.
	// See ../../../../runtime/README.md for detailed descriptions.
	Systemstack        // func must run on system stack
	Nowritebarrier     // emit compiler error instead of write barrier
	Nowritebarrierrec  // error on write barrier in this or recursive callees
	Yeswritebarrierrec // cancels Nowritebarrierrec in this function and callees

	// Runtime and cgo type pragmas
	NotInHeap // values of this type must not be heap allocated

	// Go command pragmas
	GoBuildPragma
)

type SymAndPos struct {
	Sym *obj.LSym // LSym of callee
	Pos src.XPos  // line of call
}

// A ScopeID represents a lexical scope within a function.
type ScopeID int32

const (
	funcDupok         = 1 << iota // duplicate definitions ok
	funcWrapper                   // is method wrapper
	funcNeedctxt                  // function uses context register (has closure variables)
	funcReflectMethod             // function calls reflect.Type.Method or MethodByName
	funcIsHiddenClosure
	funcHasDefer                 // contains a defer statement
	funcNilCheckDisabled         // disable nil checks when compiling this function
	funcInlinabilityChecked      // inliner has already determined whether the function is inlinable
	funcExportInline             // include inline body in export data
	funcInstrumentBody           // add race/msan instrumentation during SSA construction
	funcOpenCodedDeferDisallowed // can't do open-coded defers
)

func (f *Func) Dupok() bool                    { return f.flags&funcDupok != 0 }
func (f *Func) Wrapper() bool                  { return f.flags&funcWrapper != 0 }
func (f *Func) Needctxt() bool                 { return f.flags&funcNeedctxt != 0 }
func (f *Func) ReflectMethod() bool            { return f.flags&funcReflectMethod != 0 }
func (f *Func) IsHiddenClosure() bool          { return f.flags&funcIsHiddenClosure != 0 }
func (f *Func) HasDefer() bool                 { return f.flags&funcHasDefer != 0 }
func (f *Func) NilCheckDisabled() bool         { return f.flags&funcNilCheckDisabled != 0 }
func (f *Func) InlinabilityChecked() bool      { return f.flags&funcInlinabilityChecked != 0 }
func (f *Func) ExportInline() bool             { return f.flags&funcExportInline != 0 }
func (f *Func) InstrumentBody() bool           { return f.flags&funcInstrumentBody != 0 }
func (f *Func) OpenCodedDeferDisallowed() bool { return f.flags&funcOpenCodedDeferDisallowed != 0 }

func (f *Func) SetDupok(b bool)                    { f.flags.set(funcDupok, b) }
func (f *Func) SetWrapper(b bool)                  { f.flags.set(funcWrapper, b) }
func (f *Func) SetNeedctxt(b bool)                 { f.flags.set(funcNeedctxt, b) }
func (f *Func) SetReflectMethod(b bool)            { f.flags.set(funcReflectMethod, b) }
func (f *Func) SetIsHiddenClosure(b bool)          { f.flags.set(funcIsHiddenClosure, b) }
func (f *Func) SetHasDefer(b bool)                 { f.flags.set(funcHasDefer, b) }
func (f *Func) SetNilCheckDisabled(b bool)         { f.flags.set(funcNilCheckDisabled, b) }
func (f *Func) SetInlinabilityChecked(b bool)      { f.flags.set(funcInlinabilityChecked, b) }
func (f *Func) SetExportInline(b bool)             { f.flags.set(funcExportInline, b) }
func (f *Func) SetInstrumentBody(b bool)           { f.flags.set(funcInstrumentBody, b) }
func (f *Func) SetOpenCodedDeferDisallowed(b bool) { f.flags.set(funcOpenCodedDeferDisallowed, b) }

func (f *Func) SetWBPos(pos src.XPos) {
	if base.Debug.WB != 0 {
		base.WarnAt(pos, "write barrier")
	}
	if !f.WBPos.IsKnown() {
		f.WBPos = pos
	}
}

type DclFunc struct {
	defaultNode
	f     *Func
	nbody Nodes
	typ   *types.Type
	iota  int64
}

func newDclFunc() *DclFunc {
	d := new(DclFunc)
	d.f = new(Func)
	d.f.Decl = d
	return d
}

func (*DclFunc) Op() Op                  { return ODCLFUNC }
func (n *DclFunc) Func() *Func           { return n.f }
func (n *DclFunc) RawCopy() INode        { copy := *n; return &copy }
func (n *DclFunc) Nbody() Nodes          { return n.nbody }
func (n *DclFunc) PtrNbody() *Nodes      { return &n.nbody }
func (n *DclFunc) Type() *types.Type     { return n.typ }
func (n *DclFunc) SetType(x *types.Type) { n.typ = x }
func (n *DclFunc) Iota() int64           { return n.iota }
func (n *DclFunc) SetIota(x int64)       { n.iota = x }

// A Closure is the OCLOSURE node.
type Closure struct {
	defaultNode
	fn *Func
	nodeFieldOpt
	nodeFieldType
	nodeFieldTransient
}

func (*Closure) Op() Op            { return OCLOSURE }
func (c *Closure) RawCopy() INode  { copy := *c; return &copy }
func (c *Closure) Func() *Func     { return c.fn }
func (c *Closure) SetFunc(x *Func) { c.fn = x }

// A CallPart is the OCALLPART node.
type CallPart struct {
	defaultNode
	fn *Func
	nodeFieldLeft
	nodeFieldRight
	nodeFieldType
	nodeFieldOpt
	nodeFieldTransient
}

func (*CallPart) Op() Op            { return OCALLPART }
func (c *CallPart) RawCopy() INode  { copy := *c; return &copy }
func (c *CallPart) Func() *Func     { return c.fn }
func (c *CallPart) SetFunc(x *Func) { c.fn = x }

// FuncName returns the name (without the package) of the function n.
func FuncName(n INode) string {
	if n == nil || n.Func() == nil || n.Func().Name == nil {
		return "<nil>"
	}
	return n.Func().Name.Sym().Name
}

// PkgFuncName returns the name of the function referenced by n, with package prepended.
// This differs from the compiler's internal convention where local functions lack a package
// because the ultimate consumer of this is a human looking at an IDE; package is only empty
// if the compilation package is actually the empty string.
func PkgFuncName(n INode) string {
	var s *types.Sym
	if n == nil {
		return "<nil>"
	}
	if n.Op() == ONAME {
		s = n.Sym()
	} else {
		if n.Func() == nil || n.Func().Name == nil {
			return "<nil>"
		}
		s = n.Func().Name.Sym()
	}
	pkg := s.Pkg

	p := base.Ctxt.Pkgpath
	if pkg != nil && pkg.Path != "" {
		p = pkg.Path
	}
	if p == "" {
		return s.Name
	}
	return p + "." + s.Name
}
