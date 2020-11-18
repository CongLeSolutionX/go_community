// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/base"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/objabi"
	"cmd/internal/src"
	"fmt"
)

// A Name is a Node representing a const, type, or var (OLITERAL, OTYPE, ONAME).
// Names with op == ONONAME mean the name is as yet unbound.
type Name struct {
	op        Op
	sym       *types.Sym
	typ       *types.Type
	diag      bool
	hasCall   bool
	pos       src.XPos
	typecheck uint8
	orig      INode
	esc       uint16
	nonNil    bool

	opt      interface{}
	hasOpt   bool
	implicit bool
	readonly bool
	ninit    Nodes
	fn       *Func
	xoffset  int64
	class    Class
	subop    Op
	val      Val
	isDDD    bool
	iota     int64
	walkdef  uint8

	Pack      *PackNode  // real package for import . names
	Pkg       *types.Pkg // pkg for OPACK nodes
	Defn      INode      // initializing assignment
	Curfn     INode      // function for local variables
	Decldepth int32      // declaration loop depth, increased for every loop or label
	Vargen    int32      // unique name for ONAME within a function.  Function outputs are numbered starting at one.
	flags     bitset16
	pragma    PragmaFlag

	// ONAME //go:embed info,
	embed *[]string

	Ntype    INode
	Heapaddr INode // temp holding heap address of param

	// ONAME PAUTOHEAP
	Stackcopy INode // the PPARAM/PPARAMOUT on-stack slot (moved func params only)

	// ONAME closure linkage
	// Consider:
	//
	//	func f() {
	//		x := 1 // x1
	//		func() {
	//			use(x) // x2
	//			func() {
	//				use(x) // x3
	//				--- parser is here ---
	//			}()
	//		}()
	//	}
	//
	// There is an original declaration of x and then a chain of mentions of x
	// leading into the current function. Each time x is mentioned in a new closure,
	// we create a variable representing x for use in that specific closure,
	// since the way you get to x is different in each closure.
	//
	// Let's number the specific variables as shown in the code:
	// x1 is the original x, x2 is when mentioned in the closure,
	// and x3 is when mentioned in the closure in the closure.
	//
	// We keep these linked (assume N > 1):
	//
	//   - x1.Defn = original declaration statement for x (like most variables)
	//   - x1.Innermost = current innermost closure x (in this case x3), or nil for none
	//   - x1.IsClosureVar() = false
	//
	//   - xN.Defn = x1, N > 1
	//   - xN.IsClosureVar() = true, N > 1
	//   - x2.Outer = nil
	//   - xN.Outer = x(N-1), N > 2
	//
	//
	// When we look up x in the symbol table, we always get x1.
	// Then we can use x1.Innermost (if not nil) to get the x
	// for the innermost known closure function,
	// but the first reference in a closure will find either no x1.Innermost
	// or an x1.Innermost with .Funcdepth < Funcdepth.
	// In that case, a new xN must be created, linked in with:
	//
	//     xN.Defn = x1
	//     xN.Outer = x1.Innermost
	//     x1.Innermost = xN
	//
	// When we finish the function, we'll process its closure variables
	// and find xN and pop it off the list using:
	//
	//     x1 := xN.Defn
	//     x1.Innermost = xN.Outer
	//
	// We leave x1.Innermost set so that we can still get to the original
	// variable quickly. Not shown here, but once we're
	// done parsing a function and no longer need xN.Outer for the
	// lexical x reference links as described above, funcLit
	// recomputes xN.Outer as the semantic x reference link tree,
	// even filling in x in intermediate closures that might not
	// have mentioned it along the way to inner closures that did.
	// See funcLit for details.
	//
	// During the eventual compilation, then, for closure variables we have:
	//
	//     xN.Defn = original variable
	//     xN.Outer = variable captured in next outward scope
	//                to make closure where xN appears
	//
	// Because of the sharding of pieces of the node, x.Defn means x.Name.Defn
	// and x.Innermost/Outer means x.Name.Param.Innermost/Outer.
	Innermost INode
	Outer     INode
}

// The Class of a variable/function describes the "storage class"
// of a variable or function. During parsing, storage classes are
// called declaration contexts.
type Class uint8

//go:generate stringer -type=Class
const (
	Pxxx      Class = iota // no class; used during ssa conversion to indicate pseudo-variables
	PEXTERN                // global variables
	PAUTO                  // local variables
	PAUTOHEAP              // local variables or parameters moved to heap
	PPARAM                 // input arguments
	PPARAMOUT              // output results
	PFUNC                  // global functions

	// Careful: Class is stored in three bits in Node.flags.
	_ = uint((1 << 3) - iota) // static assert for iota <= (1 << 3)
)

func (n *Name) Op() Op                { return n.op }
func (n *Name) Sym() *types.Sym       { return n.sym }
func (n *Name) SetSym(x *types.Sym)   { n.sym = x }
func (n *Name) Type() *types.Type     { return n.typ }
func (n *Name) SetType(x *types.Type) { n.typ = x }
func (n *Name) SubOp() Op             { return n.subop }
func (n *Name) SetSubOp(x Op)         { n.subop = x }
func (n *Name) Val() Val              { return n.val }
func (n *Name) SetVal(x Val)          { n.val = x }
func (n *Name) Class() Class          { return n.class }
func (n *Name) SetClass(x Class)      { n.class = x }
func (n *Name) Func() *Func           { return n.fn }
func (n *Name) SetFunc(x *Func)       { n.fn = x }
func (n *Name) Xoffset() int64        { return n.xoffset }
func (n *Name) SetXoffset(x int64)    { n.xoffset = x }
func (n *Name) IsDDD() bool           { return n.isDDD }
func (n *Name) SetIsDDD(x bool)       { n.isDDD = x }
func (n *Name) Diag() bool            { return n.diag }
func (n *Name) SetDiag(x bool)        { n.diag = x }
func (n *Name) Iota() int64           { return n.iota }
func (n *Name) SetIota(x int64)       { n.iota = x }
func (n *Name) Walkdef() uint8        { return n.walkdef }
func (n *Name) SetWalkdef(x uint8)    { n.walkdef = x }
func (n *Name) Ninit() Nodes          { return n.ninit }
func (n *Name) PtrNinit() *Nodes      { return &n.ninit }
func (n *Name) HasCall() bool         { return n.hasCall }
func (n *Name) SetHasCall(x bool)     { n.hasCall = x }

func (n *Name) Name() *Name { return n }

func (n *Name) RawCopy() INode {
	m := *n
	return &m
}

// newnamel returns a new ONAME Node associated with symbol s at position pos.
// The caller is responsible for setting n.Name.Curfn.
func NewNameAt(pos src.XPos, s *types.Sym) *Name {
	if s == nil {
		base.Fatal("newnamel nil")
	}

	n := new(Name)
	n.SetOp(ONAME)
	n.SetPos(pos)
	n.SetOrig(n)
	n.SetSym(s)
	return n
}

func newName(op Op) INode {
	n := new(Name)
	n.SetOp(op)
	return n
}

func (n *Name) SetOp(op Op) {
	switch op {
	default:
		panic("cannot Name SetOp " + op.String())
	case OLITERAL, ONAME, ONONAME, OTYPE: // TODO: OTYPE should not be allowed
		// ok
		n.op = op
	}
}

func (n *Name) IsAutoTmp() bool {
	return n.op == ONAME && n.Name().AutoTemp()
}

// Int64Val returns n as an int64.
// n must be an integer or rune constant.
func (n *Name) Int64Val() int64 {
	if !IsConst(n, CTINT) {
		base.Fatal("Int64Val(%v)", n)
	}
	return n.Val().U.(*Int).Int64()
}

// CanInt64 reports whether it is safe to call Int64Val() on n.
func (n *Name) CanInt64() bool {
	if !IsConst(n, CTINT) {
		return false
	}

	// if the value inside n cannot be represented as an int64, the
	// return value of Int64 is undefined
	return n.Val().U.(*Int).CmpInt64(n.Int64Val()) == 0
}

// BoolVal returns n as a bool.
// n must be a boolean constant.
func (n *Name) BoolVal() bool {
	if !IsConst(n, CTBOOL) {
		base.Fatal("BoolVal(%v)", n)
	}
	return n.Val().U.(bool)
}

// StringVal returns the value of a literal string Node as a string.
// n must be a string constant.
func (n *Name) StringVal() string {
	if !IsConst(n, CTSTR) {
		base.Fatal("StringVal(%v)", n)
	}
	return n.Val().U.(string)
}

func (n *Name) Bounded() bool                 { panic("unavailable") }
func (n *Name) CanBeAnSSASym()                { panic("unavailable") }
func (n *Name) Colas() bool                   { panic("unavailable") }
func (n *Name) CopyFrom(INode)                { panic("unavailable") }
func (n *Name) Embedded() bool                { panic("unavailable") }
func (n *Name) Esc() uint16                   { return n.esc }
func (n *Name) Format(s fmt.State, verb rune) { panic("unavailable") }
func (n *Name) Nbody() Nodes                  { return Nodes{} }
func (n *Name) Rlist() Nodes                  { return Nodes{} }
func (n *Name) HasBreak() bool                { panic("unavailable") }
func (n *Name) Implicit() bool                { return n.implicit }
func (n *Name) IndexMapLValue() bool          { panic("unavailable") }
func (n *Name) Initorder() uint8              { panic("unavailable") }
func (n *Name) IsSynthetic() bool {
	name := n.Sym().Name
	return name[0] == '.' || name[0] == '~'
}

func (n *Name) Left() INode  { return nil }
func (n *Name) Likely() bool { panic("unavailable") }
func (n *Name) Line() string { panic("unavailable") }
func (n *Name) List() Nodes  { return Nodes{} }
func (n *Name) MarkNonNil() {
	if !n.Type().IsPtr() && !n.Type().IsUnsafePtr() {
		base.Fatal("MarkNonNil(%v), type %v", n, n.Type())
	}
	n.nonNil = true
}
func (n *Name) MarkReadonly() {
	if n.Op() != ONAME {
		base.Fatal("Node.MarkReadonly %v", n.Op())
	}
	n.Name().SetReadonly(true)
	// Mark the linksym as readonly immediately
	// so that the SSA backend can use this information.
	// It will be overridden later during dumpglobls.
	n.Sym().Linksym().Type = objabi.SRODATA
}
func (n *Name) MayBeShared() bool        { return true }
func (n *Name) NoInline() bool           { panic("unavailable") }
func (n *Name) NonNil() bool             { return n.nonNil }
func (n *Name) Opt() interface{}         { return n.opt }
func (n *Name) Orig() INode              { return n.orig }
func (n *Name) Pos() src.XPos            { return n.pos }
func (n *Name) PtrList() *Nodes          { return nil }
func (n *Name) PtrNbody() *Nodes         { return nil }
func (n *Name) PtrRlist() *Nodes         { return nil }
func (n *Name) ResetAux()                { panic("unavailable") }
func (n *Name) Right() INode             { return nil }
func (n *Name) SetBounded(b bool)        { panic("unavailable") }
func (n *Name) SetColas(b bool)          { panic("unavailable") }
func (n *Name) SetEmbedded(b bool)       { panic("unavailable") }
func (n *Name) SetEsc(x uint16)          { n.esc = x }
func (n *Name) SetHasBreak(b bool)       { panic("unavailable") }
func (n *Name) SetImplicit(b bool)       { n.implicit = b }
func (n *Name) SetIndexMapLValue(b bool) { panic("unavailable") }
func (n *Name) SetInitorder(b uint8)     { panic("unavailable") }
func (n *Name) SetLeft(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *Name) SetLikely(b bool)     { panic("unavailable") }
func (n *Name) SetList(x Nodes)      { panic("unavailable") }
func (n *Name) SetNbody(x Nodes)     { panic("unavailable") }
func (n *Name) SetNinit(x Nodes)     { panic("unavailable") }
func (n *Name) SetNoInline(b bool)   { panic("unavailable") }
func (n *Name) SetOpt(x interface{}) { n.opt = x; n.hasOpt = true }
func (n *Name) SetOrig(x INode)      { n.orig = x }
func (n *Name) SetPos(x src.XPos)    { n.pos = x }
func (n *Name) SetRight(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *Name) SetRlist(x Nodes)                    { panic("unavailable") }
func (n *Name) SetSliceBounds(low, high, max INode) { panic("unavailable") }
func (n *Name) SetTChanDir(dir types.ChanDir)       { panic("unavailable") }
func (n *Name) SetTransient(b bool)                 { panic("unavailable") }
func (n *Name) SetTypecheck(b uint8)                { n.typecheck = b }

func (n *Name) SliceBounds() (low, high, max INode) { panic("unavailable") }
func (n *Name) StorageClass() ssa.StorageClass {
	switch n.Class() {
	case PPARAM:
		return ssa.ClassParam
	case PPARAMOUT:
		return ssa.ClassParamOut
	case PAUTO:
		return ssa.ClassAuto
	default:
		base.Fatal("untranslatable storage class for %v: %s", n, n.Class())
		return 0
	}
}
func (n *Name) String() string          { panic("unavailable") }
func (n *Name) TChanDir() types.ChanDir { panic("unavailable") }
func (n *Name) Transient() bool         { panic("unavailable") }
func (n *Name) Typecheck() uint8        { return n.typecheck }

func (n *Name) IsBlank() bool {
	if n == nil {
		return false
	}
	return n.Sym().IsBlank()
}

// IsMethod reports whether n is a method.
// n must be a function or a method.
func (n *Name) IsMethod() bool {
	return n.Type().Recv() != nil
}

// isNil reports whether n represents the universal untyped zero value "nil".
func (n *Name) IsNil() bool {
	// Check n.Orig because constant propagation may produce typed nil constants,
	// which don't exist in the Go spec.
	return IsConst(n.Orig(), CTNIL)
}

func (n *Name) Typ() *types.Type {
	return n.Type()
}

const (
	nameCaptured = 1 << iota // is the variable captured by a closure
	nameReadonly
	nameByval                 // is the variable captured by value or by reference
	nameNeedzero              // if it contains pointers, needs to be zeroed on function entry
	nameAutoTemp              // is the variable a temporary (implies no dwarf info. reset if escapes to heap)
	nameUsed                  // for variable declared and not used error
	nameIsClosureVar          // PAUTOHEAP closure pseudo-variable; original at n.Name.Defn
	nameIsOutputParamHeapAddr // pointer to a result parameter's heap copy
	nameAssigned              // is the variable ever assigned to
	nameAddrtaken             // address taken, even if not moved to heap
	nameInlFormal             // PAUTO created by inliner, derived from callee formal
	nameInlLocal              // PAUTO created by inliner, derived from callee local
	nameOpenDeferSlot         // if temporary var storing info for open-coded defers
	nameLibfuzzerExtraCounter // if PEXTERN should be assigned to __libfuzzer_extra_counters section
	nameAlias                 // is a type alias
)

func (n *Name) Captured() bool              { return n.flags&nameCaptured != 0 }
func (n *Name) Readonly() bool              { return n.flags&nameReadonly != 0 }
func (n *Name) Byval() bool                 { return n.flags&nameByval != 0 }
func (n *Name) Needzero() bool              { return n.flags&nameNeedzero != 0 }
func (n *Name) AutoTemp() bool              { return n.flags&nameAutoTemp != 0 }
func (n *Name) Used() bool                  { return n.flags&nameUsed != 0 }
func (n *Name) IsClosureVar() bool          { return n.flags&nameIsClosureVar != 0 }
func (n *Name) IsOutputParamHeapAddr() bool { return n.flags&nameIsOutputParamHeapAddr != 0 }
func (n *Name) Assigned() bool              { return n.flags&nameAssigned != 0 }
func (n *Name) Addrtaken() bool             { return n.flags&nameAddrtaken != 0 }
func (n *Name) InlFormal() bool             { return n.flags&nameInlFormal != 0 }
func (n *Name) InlLocal() bool              { return n.flags&nameInlLocal != 0 }
func (n *Name) OpenDeferSlot() bool         { return n.flags&nameOpenDeferSlot != 0 }
func (n *Name) LibfuzzerExtraCounter() bool { return n.flags&nameLibfuzzerExtraCounter != 0 }

func (n *Name) SetCaptured(b bool)              { n.flags.set(nameCaptured, b) }
func (n *Name) SetReadonly(b bool)              { n.flags.set(nameReadonly, b) }
func (n *Name) SetByval(b bool)                 { n.flags.set(nameByval, b) }
func (n *Name) SetNeedzero(b bool)              { n.flags.set(nameNeedzero, b) }
func (n *Name) SetAutoTemp(b bool)              { n.flags.set(nameAutoTemp, b) }
func (n *Name) SetUsed(b bool)                  { n.flags.set(nameUsed, b) }
func (n *Name) SetIsClosureVar(b bool)          { n.flags.set(nameIsClosureVar, b) }
func (n *Name) SetIsOutputParamHeapAddr(b bool) { n.flags.set(nameIsOutputParamHeapAddr, b) }
func (n *Name) SetAssigned(b bool)              { n.flags.set(nameAssigned, b) }
func (n *Name) SetAddrtaken(b bool)             { n.flags.set(nameAddrtaken, b) }
func (n *Name) SetInlFormal(b bool)             { n.flags.set(nameInlFormal, b) }
func (n *Name) SetInlLocal(b bool)              { n.flags.set(nameInlLocal, b) }
func (n *Name) SetOpenDeferSlot(b bool)         { n.flags.set(nameOpenDeferSlot, b) }
func (n *Name) SetLibfuzzerExtraCounter(b bool) { n.flags.set(nameLibfuzzerExtraCounter, b) }

// Pragma returns the PragmaFlag for p, which must be for an OTYPE.
func (n *Name) Pragma() PragmaFlag {
	return n.pragma
}

// SetPragma sets the PragmaFlag for p, which must be for an OTYPE.
func (n *Name) SetPragma(flag PragmaFlag) {
	n.pragma = flag
}

// Alias reports whether p, which must be for an OTYPE, is a type alias.
func (n *Name) Alias() bool { return n.flags&nameAlias != 0 }

// SetAlias sets whether p, which must be for an OTYPE, is a type alias.
func (n *Name) SetAlias(alias bool) {
	n.flags.set(nameAlias, alias)
}

// EmbedFiles returns the list of embedded files for p,
// which must be for an ONAME var.
func (n *Name) EmbedFiles() []string {
	if n.embed == nil {
		return nil
	}
	return *n.embed
}

// SetEmbedFiles sets the list of embedded files for p,
// which must be for an ONAME var.
func (n *Name) SetEmbedFiles(list []string) {
	if n.embed == nil {
		if len(list) == 0 {
			return
		}
		n.embed = new([]string)
	}
	*n.embed = list
}

// A PackNode is an INode for the name of an imported package.
type PackNode struct {
	TrivNode
	sym  *types.Sym
	Used bool
	Pkg  *types.Pkg
}

func NewPackNode(pos src.XPos, sym *types.Sym, pkg *types.Pkg) *PackNode {
	p := &PackNode{sym: sym, Pkg: pkg}
	p.SetPos(pos)
	return p
}

func (*PackNode) Op() Op            { return OPACK }
func (p *PackNode) RawCopy() INode  { panic("can't copy PackNode") }
func (p *PackNode) Sym() *types.Sym { return p.sym }

// A LabelNode is an INode for the name of an imported package.
type LabelNode struct {
	TrivNode
	sym  *types.Sym
	Defn INode // statement being labeled
}

func NewLabelNode(pos src.XPos, sym *types.Sym) *LabelNode {
	l := &LabelNode{sym: sym}
	l.SetPos(pos)
	return l
}

func (*LabelNode) Op() Op                { return OLABEL }
func (l *LabelNode) RawCopy() INode      { copy := *l; return &copy }
func (l *LabelNode) Sym() *types.Sym     { return l.sym }
func (l *LabelNode) SetSym(x *types.Sym) { l.sym = x }
