// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// “Abstract” syntax representation.

package ir

import (
	"fmt"
	"sort"
	"strings"

	"cmd/compile/internal/base"
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/src"
)

// A Node is a single node in the syntax tree.
// Actually the syntax tree is a syntax DAG, because there is only one
// node with Op=ONAME for a given instance of a variable x.
// The same is true for Op=OTYPE and Op=OLITERAL. See Node.mayBeShared.
type node struct {
	// Tree structure.
	// Generic recursive walks should follow these fields.
	left  INode
	right INode
	ninit Nodes
	nbody Nodes
	list  Nodes
	rlist Nodes

	// most nodes
	typ  *types.Type
	orig INode // original form, for printing, and tracking copies of ONAMEs

	// func
	fn *Func

	sym *types.Sym // various
	opt interface{}

	// Various. Usually an offset into a struct. For example:
	// - ONAME nodes that refer to local variables use it to identify their stack frame position.
	// - ODOT, ODOTPTR, and ORESULT use it to indicate offset relative to their base address.
	// - OSTRUCTKEY uses it to store the named field's offset.
	// - Named OLITERALs use it to store their ambient iota value.
	// - OINLMARK stores an index into the inlTree data structure.
	// - OCLOSURE uses it to store ambient iota value, if any.
	// Possibly still more uses. If you find any, document them.
	xoffset int64

	pos src.XPos

	flags bitset32

	esc uint16 // EscXXX

	op  Op
	aux uint8
}

var badForNode = [OEND]bool{
	OCONTINUE: true,
	ONONAME:   true,
	OLITERAL:  true,
	OPACK:     true,
	OLABEL:    true,
}

func (n *node) Format(s fmt.State, verb rune) { FmtNode(n, s, verb) }
func (n *node) String() string                { return fmt.Sprint(n) }

func (n *node) Left() INode           { return n.left }
func (n *node) SetLeft(x INode)       { n.left = x }
func (n *node) Right() INode          { return n.right }
func (n *node) SetRight(x INode)      { n.right = x }
func (n *node) Orig() INode           { return n.orig }
func (n *node) SetOrig(x INode)       { n.orig = x }
func (n *node) Type() *types.Type     { return n.typ }
func (n *node) SetType(x *types.Type) { n.typ = x }
func (n *node) Func() *Func           { return n.fn }
func (n *node) SetFunc(x *Func)       { n.fn = x }
func (n *node) Name() *Name           { return nil }
func (n *node) Sym() *types.Sym       { return n.sym }
func (n *node) SetSym(x *types.Sym)   { n.sym = x }
func (n *node) Pos() src.XPos         { return n.pos }
func (n *node) SetPos(x src.XPos)     { n.pos = x }
func (n *node) Xoffset() int64        { return n.xoffset }
func (n *node) SetXoffset(x int64)    { n.xoffset = x }
func (n *node) Esc() uint16           { return n.esc }
func (n *node) SetEsc(x uint16)       { n.esc = x }
func (n *node) Op() Op                { return n.op }
func (n *node) SetOp(op Op) {
	if badForNode[op] {
		panic("unavailable SetOp " + op.String())
	}
	n.op = op
}

func (n *node) Ninit() Nodes     { return n.ninit }
func (n *node) SetNinit(x Nodes) { n.ninit = x }
func (n *node) PtrNinit() *Nodes { return &n.ninit }
func (n *node) Nbody() Nodes     { return n.nbody }
func (n *node) SetNbody(x Nodes) { n.nbody = x }
func (n *node) PtrNbody() *Nodes { return &n.nbody }
func (n *node) List() Nodes      { return n.list }
func (n *node) SetList(x Nodes)  { n.list = x }
func (n *node) PtrList() *Nodes  { return &n.list }
func (n *node) Rlist() Nodes     { return n.rlist }
func (n *node) SetRlist(x Nodes) { n.rlist = x }
func (n *node) PtrRlist() *Nodes { return &n.rlist }

func (n *node) ResetAux() {
	n.aux = 0
}

func (n *node) SubOp() Op {
	switch n.Op() {
	case OASOP, ONAME:
	default:
		base.Fatal("unexpected op: %v", n.Op())
	}
	return Op(n.aux)
}

func (n *node) SetSubOp(op Op) {
	switch n.Op() {
	case OASOP, ONAME:
	default:
		base.Fatal("unexpected op: %v", n.Op())
	}
	n.aux = uint8(op)
}

func (n *node) IndexMapLValue() bool {
	if n.Op() != OINDEXMAP {
		base.Fatal("unexpected op: %v", n.Op())
	}
	return n.aux != 0
}

func (n *node) SetIndexMapLValue(b bool) {
	if n.Op() != OINDEXMAP {
		base.Fatal("unexpected op: %v", n.Op())
	}
	if b {
		n.aux = 1
	} else {
		n.aux = 0
	}
}

func (n *node) TChanDir() types.ChanDir {
	if n.Op() != OTCHAN {
		base.Fatal("unexpected op: %v", n.Op())
	}
	return types.ChanDir(n.aux)
}

func (n *node) SetTChanDir(dir types.ChanDir) {
	if n.Op() != OTCHAN {
		base.Fatal("unexpected op: %v", n.Op())
	}
	n.aux = uint8(dir)
}

func (n *node) IsSynthetic() bool {
	name := n.Sym().Name
	return name[0] == '.' || name[0] == '~'
}

// IsAutoTmp indicates if n was created by the compiler as a temporary,
// based on the setting of the .AutoTemp flag in n's Name.
func (n *node) IsAutoTmp() bool { return false }

const (
	nodeTypecheck, _ = iota, 1 << iota // tracks state during typechecking; 2 == loop detected; two bits
	_, _                               // second nodeTypecheck bit
	nodeInitorder, _                   // tracks state during init1; two bits
	_, _                               // second nodeInitorder bit
	_, nodeHasBreak
	_, nodeNoInline  // used internally by inliner to indicate that a function call should not be inlined; set for OCALLFUNC and OCALLMETH only
	_, nodeImplicit  // implicit OADDR or ODEREF; ++/-- statement represented as OASOP
	_, nodeIsDDD     // is the argument variadic
	_, nodeDiag      // already printed error about this
	_, nodeColas     // OAS resulting from :=
	_, nodeNonNil    // guaranteed to be non-nil
	_, nodeTransient // storage can be reused immediately after this statement
	_, nodeBounded   // bounds check unnecessary
	_, nodeHasCall   // expression contains a function call
	_, nodeLikely    // if statement condition likely
	_, nodeEmbedded  // ODCLFIELD embedded type
)

func (n *node) Class() Class     { panic("unavailable in " + n.Op().String()) }
func (n *node) Walkdef() uint8   { panic("unavailable in " + n.Op().String()) }
func (n *node) Typecheck() uint8 { return n.flags.get2(nodeTypecheck) }
func (n *node) Initorder() uint8 { return n.flags.get2(nodeInitorder) }

func (n *node) HasBreak() bool  { return n.flags&nodeHasBreak != 0 }
func (n *node) NoInline() bool  { return n.flags&nodeNoInline != 0 }
func (n *node) Implicit() bool  { return n.flags&nodeImplicit != 0 }
func (n *node) IsDDD() bool     { return n.flags&nodeIsDDD != 0 }
func (n *node) Diag() bool      { return n.flags&nodeDiag != 0 }
func (n *node) Colas() bool     { return n.flags&nodeColas != 0 }
func (n *node) NonNil() bool    { return n.flags&nodeNonNil != 0 }
func (n *node) Transient() bool { return n.flags&nodeTransient != 0 }
func (n *node) Bounded() bool   { return n.flags&nodeBounded != 0 }
func (n *node) HasCall() bool   { return n.flags&nodeHasCall != 0 }
func (n *node) Likely() bool    { return n.flags&nodeLikely != 0 }
func (n *node) Embedded() bool  { return n.flags&nodeEmbedded != 0 }

func (n *node) SetClass(b Class)     { panic("unavailable in " + n.Op().String()) }
func (n *node) SetWalkdef(b uint8)   { panic("unavailable in " + n.Op().String()) }
func (n *node) SetTypecheck(b uint8) { n.flags.set2(nodeTypecheck, b) }
func (n *node) SetInitorder(b uint8) { n.flags.set2(nodeInitorder, b) }

func (n *node) SetHasBreak(b bool)  { n.flags.set(nodeHasBreak, b) }
func (n *node) SetNoInline(b bool)  { n.flags.set(nodeNoInline, b) }
func (n *node) SetImplicit(b bool)  { n.flags.set(nodeImplicit, b) }
func (n *node) SetIsDDD(b bool)     { n.flags.set(nodeIsDDD, b) }
func (n *node) SetDiag(b bool)      { n.flags.set(nodeDiag, b) }
func (n *node) SetColas(b bool)     { n.flags.set(nodeColas, b) }
func (n *node) SetTransient(b bool) { n.flags.set(nodeTransient, b) }
func (n *node) SetHasCall(b bool)   { n.flags.set(nodeHasCall, b) }
func (n *node) SetLikely(b bool)    { n.flags.set(nodeLikely, b) }
func (n *node) SetEmbedded(b bool)  { n.flags.set(nodeEmbedded, b) }

// MarkNonNil marks a pointer n as being guaranteed non-nil,
// on all code paths, at all times.
// During conversion to SSA, non-nil pointers won't have nil checks
// inserted before dereferencing. See state.exprPtr.
func (n *node) MarkNonNil() {
	if !n.Type().IsPtr() && !n.Type().IsUnsafePtr() {
		base.Fatal("MarkNonNil(%v), type %v", n, n.Type())
	}
	n.flags.set(nodeNonNil, true)
}

// SetBounded indicates whether operation n does not need safety checks.
// When n is an index or slice operation, n does not need bounds checks.
// When n is a dereferencing operation, n does not need nil checks.
// When n is a makeslice+copy operation, n does not need length and cap checks.
func (n *node) SetBounded(b bool) {
	switch n.Op() {
	case OINDEX, OSLICE, OSLICEARR, OSLICE3, OSLICE3ARR, OSLICESTR:
		// No bounds checks needed.
	case ODOTPTR, ODEREF:
		// No nil check needed.
	case OMAKESLICECOPY:
		// No length and cap checks needed
		// since new slice and copied over slice data have same length.
	default:
		base.Fatal("SetBounded(%v)", n)
	}
	n.flags.set(nodeBounded, b)
}

func (n *node) MarkReadonly() { panic("unavailable") }
func (n *node) Val() Val      { panic("unavailable") }
func (n *node) SetVal(v Val)  { panic("unavailable") }

// Opt returns the optimizer data for the node.
func (n *node) Opt() interface{} {
	return n.opt
}

// SetOpt sets the optimizer data for the node.
func (n *node) SetOpt(x interface{}) {
	n.opt = x
}

func (n *node) Iota() int64 {
	return n.Xoffset()
}

func (n *node) SetIota(x int64) {
	n.SetXoffset(x)
}

// mayBeShared reports whether n may occur in multiple places in the AST.
// Extra care must be taken when mutating such a node.
func (n *node) MayBeShared() bool {
	return false
}

// TODO(rsc): make toplevel
// funcname returns the name (without the package) of the function n.
func FuncName(n INode) string {
	if n == nil || n.Func() == nil || n.Func().Nname == nil {
		return "<nil>"
	}
	return n.Func().Nname.Sym().Name
}

// TODO(rsc): make toplevel
// pkgFuncName returns the name of the function referenced by n, with package prepended.
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
		if n.Func() == nil || n.Func().Nname == nil {
			return "<nil>"
		}
		s = n.Func().Nname.Sym()
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

// The compiler needs *Node to be assignable to cmd/compile/internal/ssa.Sym.
func (n *node) CanBeAnSSASym() {
}

//go:generate stringer -type=Op -trimprefix=O

type Op uint8

// Node ops.
const (
	OXXX Op = iota

	// names
	ONAME    // var or func name
	ONONAME  // unnamed arg or return value: f(int, string) (int, error) { etc }
	OTYPE    // type name
	OPACK    // import
	OLITERAL // literal

	// expressions
	OADD          // Left + Right
	OSUB          // Left - Right
	OOR           // Left | Right
	OXOR          // Left ^ Right
	OADDSTR       // +{List} (string addition, list elements are strings)
	OADDR         // &Left
	OANDAND       // Left && Right
	OAPPEND       // append(List); after walk, Left may contain elem type descriptor
	OBYTES2STR    // Type(Left) (Type is string, Left is a []byte)
	OBYTES2STRTMP // Type(Left) (Type is string, Left is a []byte, ephemeral)
	ORUNES2STR    // Type(Left) (Type is string, Left is a []rune)
	OSTR2BYTES    // Type(Left) (Type is []byte, Left is a string)
	OSTR2BYTESTMP // Type(Left) (Type is []byte, Left is a string, ephemeral)
	OSTR2RUNES    // Type(Left) (Type is []rune, Left is a string)
	OAS           // Left = Right or (if Colas=true) Left := Right
	OAS2          // List = Rlist (x, y, z = a, b, c)
	OAS2DOTTYPE   // List = Right (x, ok = I.(int))
	OAS2FUNC      // List = Right (x, y = f())
	OAS2MAPR      // List = Right (x, ok = m["foo"])
	OAS2RECV      // List = Right (x, ok = <-c)
	OASOP         // Left Etype= Right (x += y)
	OCALL         // Left(List) (function call, method call or type conversion)

	// OCALLFUNC, OCALLMETH, and OCALLINTER have the same structure.
	// Prior to walk, they are: Left(List), where List is all regular arguments.
	// After walk, List is a series of assignments to temporaries,
	// and Rlist is an updated set of arguments.
	// Nbody is all OVARLIVE nodes that are attached to OCALLxxx.
	// TODO(josharian/khr): Use Ninit instead of List for the assignments to temporaries. See CL 114797.
	OCALLFUNC  // Left(List/Rlist) (function call f(args))
	OCALLMETH  // Left(List/Rlist) (direct method call x.Method(args))
	OCALLINTER // Left(List/Rlist) (interface method call x.Method(args))
	OCALLPART  // Left.Right (method expression x.Method, not called)
	OCAP       // cap(Left)
	OCLOSE     // close(Left)
	OCLOSURE   // func Type { Func.Closure.Nbody } (func literal)
	OCOMPLIT   // Right{List} (composite literal, not yet lowered to specific form)
	OMAPLIT    // Type{List} (composite literal, Type is map)
	OSTRUCTLIT // Type{List} (composite literal, Type is struct)
	OARRAYLIT  // Type{List} (composite literal, Type is array)
	OSLICELIT  // Type{List} (composite literal, Type is slice) Right.Int64() = slice length.
	OPTRLIT    // &Left (left is composite literal)
	OCONV      // Type(Left) (type conversion)
	OCONVIFACE // Type(Left) (type conversion, to interface)
	OCONVNOP   // Type(Left) (type conversion, no effect)
	OCOPY      // copy(Left, Right)
	ODCL       // var Left (declares Left of type Left.Type)

	// Used during parsing but don't last.
	ODCLFUNC  // func f() or func (r) f()
	ODCLFIELD // struct field, interface field, or func/method argument/return value.
	ODCLCONST // const pi = 3.14
	ODCLTYPE  // type Int int or type Int = int

	ODELETE        // delete(List)
	ODOT           // Left.Sym (Left is of struct type)
	ODOTPTR        // Left.Sym (Left is of pointer to struct type)
	ODOTMETH       // Left.Sym (Left is non-interface, Right is method name)
	ODOTINTER      // Left.Sym (Left is interface, Right is method name)
	OXDOT          // Left.Sym (before rewrite to one of the preceding)
	ODOTTYPE       // Left.Right or Left.Type (.Right during parsing, .Type once resolved); after walk, .Right contains address of interface type descriptor and .Right.Right contains address of concrete type descriptor
	ODOTTYPE2      // Left.Right or Left.Type (.Right during parsing, .Type once resolved; on rhs of OAS2DOTTYPE); after walk, .Right contains address of interface type descriptor
	OEQ            // Left == Right
	ONE            // Left != Right
	OLT            // Left < Right
	OLE            // Left <= Right
	OGE            // Left >= Right
	OGT            // Left > Right
	ODEREF         // *Left
	OINDEX         // Left[Right] (index of array or slice)
	OINDEXMAP      // Left[Right] (index of map)
	OKEY           // Left:Right (key:value in struct/array/map literal)
	OSTRUCTKEY     // Sym:Left (key:value in struct literal, after type checking)
	OLEN           // len(Left)
	OMAKE          // make(List) (before type checking converts to one of the following)
	OMAKECHAN      // make(Type, Left) (type is chan)
	OMAKEMAP       // make(Type, Left) (type is map)
	OMAKESLICE     // make(Type, Left, Right) (type is slice)
	OMAKESLICECOPY // makeslicecopy(Type, Left, Right) (type is slice; Left is length and Right is the copied from slice)
	// OMAKESLICECOPY is created by the order pass and corresponds to:
	//  s = make(Type, Left); copy(s, Right)
	//
	// Bounded can be set on the node when Left == len(Right) is known at compile time.
	//
	// This node is created so the walk pass can optimize this pattern which would
	// otherwise be hard to detect after the order pass.
	OMUL         // Left * Right
	ODIV         // Left / Right
	OMOD         // Left % Right
	OLSH         // Left << Right
	ORSH         // Left >> Right
	OAND         // Left & Right
	OANDNOT      // Left &^ Right
	ONEW         // new(Left); corresponds to calls to new in source code
	ONEWOBJ      // runtime.newobject(n.Type); introduced by walk; Left is type descriptor
	ONOT         // !Left
	OBITNOT      // ^Left
	OPLUS        // +Left
	ONEG         // -Left
	OOROR        // Left || Right
	OPANIC       // panic(Left)
	OPRINT       // print(List)
	OPRINTN      // println(List)
	OPAREN       // (Left)
	OSEND        // Left <- Right
	OSLICE       // Left[List[0] : List[1]] (Left is untypechecked or slice)
	OSLICEARR    // Left[List[0] : List[1]] (Left is array)
	OSLICESTR    // Left[List[0] : List[1]] (Left is string)
	OSLICE3      // Left[List[0] : List[1] : List[2]] (Left is untypedchecked or slice)
	OSLICE3ARR   // Left[List[0] : List[1] : List[2]] (Left is array)
	OSLICEHEADER // sliceheader{Left, List[0], List[1]} (Left is unsafe.Pointer, List[0] is length, List[1] is capacity)
	ORECOVER     // recover()
	ORECV        // <-Left
	ORUNESTR     // Type(Left) (Type is string, Left is rune)
	OSELRECV     // Left = <-Right.Left: (appears as .Left of OCASE; Right.Op == ORECV)
	OSELRECV2    // List = <-Right.Left: (appears as .Left of OCASE; count(List) == 2, Right.Op == ORECV)
	OIOTA        // iota
	OREAL        // real(Left)
	OIMAG        // imag(Left)
	OCOMPLEX     // complex(Left, Right) or complex(List[0]) where List[0] is a 2-result function call
	OALIGNOF     // unsafe.Alignof(Left)
	OOFFSETOF    // unsafe.Offsetof(Left)
	OSIZEOF      // unsafe.Sizeof(Left)
	OMETHEXPR    // method expression

	// statements
	OBLOCK // { List } (block of code)
	OBREAK // break [Sym]
	// OCASE:  case List: Nbody (List==nil means default)
	//   For OTYPESW, List is a OTYPE node for the specified type (or OLITERAL
	//   for nil), and, if a type-switch variable is specified, Rlist is an
	//   ONAME for the version of the type-switch variable with the specified
	//   type.
	OCASE
	OCONTINUE // continue [Sym]
	ODEFER    // defer Left (Left must be call)
	OEMPTY    // no-op (empty statement)
	OFALL     // fallthrough
	OFOR      // for Ninit; Left; Right { Nbody }
	// OFORUNTIL is like OFOR, but the test (Left) is applied after the body:
	// 	Ninit
	// 	top: { Nbody }   // Execute the body at least once
	// 	cont: Right
	// 	if Left {        // And then test the loop condition
	// 		List     // Before looping to top, execute List
	// 		goto top
	// 	}
	// OFORUNTIL is created by walk. There's no way to write this in Go code.
	OFORUNTIL
	OGOTO   // goto Sym
	OIF     // if Ninit; Left { Nbody } else { Rlist }
	OLABEL  // Sym:
	OGO     // go Left (Left must be call)
	ORANGE  // for List = range Right { Nbody }
	ORETURN // return List
	OSELECT // select { List } (List is list of OCASE)
	OSWITCH // switch Ninit; Left { List } (List is a list of OCASE)
	// OTYPESW:  Left := Right.(type) (appears as .Left of OSWITCH)
	//   Left is nil if there is no type-switch variable
	OTYPESW

	// types
	OTCHAN   // chan int
	OTMAP    // map[string]int
	OTSTRUCT // struct{}
	OTINTER  // interface{}
	// OTFUNC: func() - Left is receiver field, List is list of param fields, Rlist is
	// list of result fields.
	OTFUNC
	OTARRAY // []int, [8]int, [N]int or [...]int

	// misc
	ODDD        // func f(args ...int) or f(l...) or var a = [...]int{0, 1, 2}.
	OINLCALL    // intermediary representation of an inlined call.
	OEFACE      // itable and data words of an empty-interface value.
	OITAB       // itable word of an interface value.
	OIDATA      // data word of an interface value in Left
	OSPTR       // base pointer of a slice or string.
	OCLOSUREVAR // variable reference at beginning of closure function
	OCFUNC      // reference to c function pointer (not go func value)
	OCHECKNIL   // emit code to ensure pointer/interface not nil
	OVARDEF     // variable is about to be fully initialized
	OVARKILL    // variable is dead
	OVARLIVE    // variable is alive
	ORESULT     // result of a function call; Xoffset is stack offset
	OINLMARK    // start of an inlined body, with file/line of caller. Xoffset is an index into the inline tree.

	// arch-specific opcodes
	ORETJMP // return to other function
	OGETG   // runtime.getg() (read g pointer)

	OEND
)

// Nodes is a pointer to a slice of *Node.
// For fields that are not used in most nodes, this is used instead of
// a slice to save space.
type Nodes struct{ slice *[]INode }

// asNodes returns a slice of *Node as a Nodes value.
func AsNodes(s []INode) Nodes {
	return Nodes{&s}
}

// Slice returns the entries in Nodes as a slice.
// Changes to the slice entries (as in s[i] = n) will be reflected in
// the Nodes.
func (n Nodes) Slice() []INode {
	if n.slice == nil {
		return nil
	}
	return *n.slice
}

// Len returns the number of entries in Nodes.
func (n Nodes) Len() int {
	if n.slice == nil {
		return 0
	}
	return len(*n.slice)
}

// Index returns the i'th element of Nodes.
// It panics if n does not have at least i+1 elements.
func (n Nodes) Index(i int) INode {
	return (*n.slice)[i]
}

// First returns the first element of Nodes (same as n.Index(0)).
// It panics if n has no elements.
func (n Nodes) First() INode {
	return (*n.slice)[0]
}

// Second returns the second element of Nodes (same as n.Index(1)).
// It panics if n has fewer than two elements.
func (n Nodes) Second() INode {
	return (*n.slice)[1]
}

// Set sets n to a slice.
// This takes ownership of the slice.
func (n *Nodes) Set(s []INode) {
	if len(s) == 0 {
		if n == nil {
			return
		}
		n.slice = nil
	} else {
		// Copy s and take address of t rather than s to avoid
		// allocation in the case where len(s) == 0 (which is
		// over 3x more common, dynamically, for make.bash).
		t := s
		n.slice = &t
	}
}

// Set1 sets n to a slice containing a single node.
func (n *Nodes) Set1(n1 INode) {
	n.slice = &[]INode{n1}
}

// Set2 sets n to a slice containing two nodes.
func (n *Nodes) Set2(n1, n2 INode) {
	n.slice = &[]INode{n1, n2}
}

// Set3 sets n to a slice containing three nodes.
func (n *Nodes) Set3(n1, n2, n3 INode) {
	n.slice = &[]INode{n1, n2, n3}
}

// MoveNodes sets n to the contents of n2, then clears n2.
func (n *Nodes) MoveNodes(n2 *Nodes) {
	n.slice = n2.slice
	n2.slice = nil
}

// SetIndex sets the i'th element of Nodes to node.
// It panics if n does not have at least i+1 elements.
func (n Nodes) SetIndex(i int, node INode) {
	(*n.slice)[i] = node
}

// SetFirst sets the first element of Nodes to node.
// It panics if n does not have at least one elements.
func (n Nodes) SetFirst(node INode) {
	(*n.slice)[0] = node
}

// SetSecond sets the second element of Nodes to node.
// It panics if n does not have at least two elements.
func (n Nodes) SetSecond(node INode) {
	(*n.slice)[1] = node
}

// Addr returns the address of the i'th element of Nodes.
// It panics if n does not have at least i+1 elements.
func (n Nodes) Addr(i int) *INode {
	return &(*n.slice)[i]
}

// Append appends entries to Nodes.
func (n *Nodes) Append(a ...INode) {
	if len(a) == 0 {
		return
	}
	if n.slice == nil {
		s := make([]INode, len(a))
		copy(s, a)
		n.slice = &s
		return
	}
	*n.slice = append(*n.slice, a...)
}

// Prepend prepends entries to Nodes.
// If a slice is passed in, this will take ownership of it.
func (n *Nodes) Prepend(a ...INode) {
	if len(a) == 0 {
		return
	}
	if n.slice == nil {
		n.slice = &a
	} else {
		*n.slice = append(a, *n.slice...)
	}
}

// AppendNodes appends the contents of *n2 to n, then clears n2.
func (n *Nodes) AppendNodes(n2 *Nodes) {
	switch {
	case n2.slice == nil:
	case n.slice == nil:
		n.slice = n2.slice
	default:
		*n.slice = append(*n.slice, *n2.slice...)
	}
	n2.slice = nil
}

// inspect invokes f on each node in an AST in depth-first order.
// If f(n) returns false, inspect skips visiting n's children.
func Inspect(n INode, f func(INode) bool) {
	if n == nil || !f(n) {
		return
	}
	InspectList(n.Ninit(), f)
	Inspect(n.Left(), f)
	Inspect(n.Right(), f)
	InspectList(n.List(), f)
	InspectList(n.Nbody(), f)
	InspectList(n.Rlist(), f)
}

func InspectList(l Nodes, f func(INode) bool) {
	for _, n := range l.Slice() {
		Inspect(n, f)
	}
}

// nodeQueue is a FIFO queue of *Node. The zero value of nodeQueue is
// a ready-to-use empty queue.
type NodeQueue struct {
	ring       []INode
	head, tail int
}

// empty reports whether q contains no Nodes.
func (q *NodeQueue) Empty() bool {
	return q.head == q.tail
}

// pushRight appends n to the right of the queue.
func (q *NodeQueue) PushRight(n INode) {
	if len(q.ring) == 0 {
		q.ring = make([]INode, 16)
	} else if q.head+len(q.ring) == q.tail {
		// Grow the ring.
		nring := make([]INode, len(q.ring)*2)
		// Copy the old elements.
		part := q.ring[q.head%len(q.ring):]
		if q.tail-q.head <= len(part) {
			part = part[:q.tail-q.head]
			copy(nring, part)
		} else {
			pos := copy(nring, part)
			copy(nring[pos:], q.ring[:q.tail%len(q.ring)])
		}
		q.ring, q.head, q.tail = nring, 0, q.tail-q.head
	}

	q.ring[q.tail%len(q.ring)] = n
	q.tail++
}

// popLeft pops a node from the left of the queue. It panics if q is
// empty.
func (q *NodeQueue) PopLeft() INode {
	if q.Empty() {
		panic("dequeue empty")
	}
	n := q.ring[q.head%len(q.ring)]
	q.head++
	return n
}

// NodeSet is a set of Nodes.
type NodeSet map[INode]struct{}

// Has reports whether s contains n.
func (s NodeSet) Has(n INode) bool {
	_, isPresent := s[n]
	return isPresent
}

// Add adds n to s.
func (s *NodeSet) Add(n INode) {
	if *s == nil {
		*s = make(map[INode]struct{})
	}
	(*s)[n] = struct{}{}
}

// Sorted returns s sorted according to less.
func (s NodeSet) Sorted(less func(INode, INode) bool) []INode {
	var res []INode
	for n := range s {
		res = append(res, n)
	}
	sort.Slice(res, func(i, j int) bool { return less(res[i], res[j]) })
	return res
}

func AsNode(n types.IRNode) INode {
	if n == nil {
		return nil
	}
	return n.(INode)
}

func AsTypesNode(n INode) types.IRNode {
	return n
}

var BlankNode INode

// origSym returns the original symbol written by the user.
func OrigSym(s *types.Sym) *types.Sym {
	if s == nil {
		return nil
	}

	if len(s.Name) > 1 && s.Name[0] == '~' {
		switch s.Name[1] {
		case 'r': // originally an unnamed result
			return nil
		case 'b': // originally the blank identifier _
			// TODO(mdempsky): Does s.Pkg matter here?
			return BlankNode.Sym()
		}
		return s
	}

	if strings.HasPrefix(s.Name, ".anon") {
		// originally an unnamed or _ name (see subr.go: structargs)
		return nil
	}

	return s
}

// SliceBounds returns n's slice bounds: low, high, and max in expr[low:high:max].
// n must be a slice expression. max is nil if n is a simple slice expression.
func (n *node) SliceBounds() (low, high, max INode) {
	if n.list.Len() == 0 {
		return nil, nil, nil
	}

	switch n.Op() {
	case OSLICE, OSLICEARR, OSLICESTR:
		s := n.list.Slice()
		return s[0], s[1], nil
	case OSLICE3, OSLICE3ARR:
		s := n.list.Slice()
		return s[0], s[1], s[2]
	}
	base.Fatal("SliceBounds op %v: %v", n.Op(), n)
	return nil, nil, nil
}

// SetSliceBounds sets n's slice bounds, where n is a slice expression.
// n must be a slice expression. If max is non-nil, n must be a full slice expression.
func (n *node) SetSliceBounds(low, high, max INode) {
	switch n.Op() {
	case OSLICE, OSLICEARR, OSLICESTR:
		if max != nil {
			base.Fatal("SetSliceBounds %v given three bounds", n.Op())
		}
		s := n.list.Slice()
		if s == nil {
			if low == nil && high == nil {
				return
			}
			n.PtrList().Set2(low, high)
			return
		}
		s[0] = low
		s[1] = high
		return
	case OSLICE3, OSLICE3ARR:
		s := n.list.Slice()
		if s == nil {
			if low == nil && high == nil && max == nil {
				return
			}
			n.PtrList().Set3(low, high, max)
			return
		}
		s[0] = low
		s[1] = high
		s[2] = max
		return
	}
	base.Fatal("SetSliceBounds op %v: %v", n.Op(), n)
}

// IsSlice3 reports whether o is a slice3 op (OSLICE3, OSLICE3ARR).
// o must be a slicing op.
func (o Op) IsSlice3() bool {
	switch o {
	case OSLICE, OSLICEARR, OSLICESTR:
		return false
	case OSLICE3, OSLICE3ARR:
		return true
	}
	base.Fatal("IsSlice3 op %v", o)
	return false
}

func IsConst(n INode, ct Ctype) bool {
	t := ConstType(n)

	// If the caller is asking for CTINT, allow CTRUNE too.
	// Makes life easier for back ends.
	return t == ct || (ct == CTINT && t == CTRUNE)
}

func (n *node) CanInt64() bool    { return false }
func (n *node) Int64Val() int64   { panic("unavailable") }
func (n *node) BoolVal() bool     { panic("unavailable") }
func (n *node) StringVal() string { panic("unavailable") }

// rawcopy returns a shallow copy of n.
// Note: copy or sepcopy (rather than rawcopy) is usually the
//       correct choice (see comment with Node.copy, below).
func (n *node) RawCopy() INode {
	copy := *n
	return &copy
}

// SepCopy returns a separate shallow copy of n, with the copy's
// Orig pointing to itself.
func SepCopy(n INode) INode {
	copy := n.RawCopy()
	copy.SetOrig(copy)
	return copy
}

// Copy returns shallow copy of n and adjusts the copy's Orig if
// necessary: In general, if n.Orig points to itself, the copy's
// Orig should point to itself as well. Otherwise, if n is modified,
// the copy's Orig node appears modified, too, and then doesn't
// represent the original node anymore.
// (This caused the wrong complit Op to be used when printing error
// messages; see issues #26855, #27765).
func Copy(n INode) INode {
	copy := n.RawCopy()
	if n.Orig() == n {
		copy.SetOrig(copy)
	}
	return copy
}

// isNil reports whether n represents the universal untyped zero value "nil".
func (n *node) IsNil() bool {
	// Check n.Orig because constant propagation may produce typed nil constants,
	// which don't exist in the Go spec.
	return IsConst(n.Orig(), CTNIL)
}

func (n *node) IsBlank() bool {
	if n == nil {
		return false
	}
	if n.Sym().IsBlank() {
		panic("node is blank")
	}
	return n.Sym().IsBlank()
}

// IsMethod reports whether n is a method.
// n must be a function or a method.
func (n *node) IsMethod() bool {
	return n.Type().Recv() != nil
}

// Line returns n's position as a string. If n has been inlined,
// it uses the outermost position where n has been inlined.
func (n *node) Line() string {
	return base.FmtPos(n.Pos())
}

func (n *node) Typ() *types.Type {
	return n.Type()
}

func (n *node) StorageClass() ssa.StorageClass { panic("unavailable") }

func Nod(op Op, nleft, nright INode) INode {
	return NodAt(base.Pos, op, nleft, nright)
}

func NodAt(pos src.XPos, op Op, nleft, nright INode) INode {
	var n INode
	switch op {
	case ODCLFUNC:
		var x struct {
			n node
			f Func
		}
		n = &x.n
		n.SetOp(op)
		n.SetFunc(&x.f)
		n.Func().Decl = n
	case OCONTINUE:
		n = new(ContinueStmt)
	case ONONAME, OLITERAL, OTYPE:
		n = newName(op)
	default:
		n = new(node)
		n.SetOp(op)
	}
	n.SetLeft(nleft)
	n.SetRight(nright)
	n.SetPos(pos)
	n.SetXoffset(types.BADWIDTH)
	n.SetOrig(n)
	return n
}
