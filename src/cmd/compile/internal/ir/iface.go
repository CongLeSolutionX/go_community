// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ir

import (
	"cmd/compile/internal/ssa"
	"cmd/compile/internal/types"
	"cmd/internal/src"
	"fmt"
)

// INode is the interface implemented by nodes.
// TODO(rsc): Reduce over time.
type INode interface {
	BoolVal() bool
	Bounded() bool
	CanBeAnSSASym()
	CanInt64() bool
	Class() Class
	Colas() bool
	Diag() bool
	Embedded() bool
	Esc() uint16
	Format(s fmt.State, verb rune)
	Func() *Func
	Nbody() Nodes
	Ninit() Nodes
	Rlist() Nodes
	HasBreak() bool
	HasCall() bool
	Implicit() bool
	IndexMapLValue() bool
	Initorder() uint8
	Int64Val() int64
	Iota() int64
	IsAutoTmp() bool
	IsBlank() bool
	IsDDD() bool
	IsMethod() bool
	IsNil() bool
	IsSynthetic() bool
	Left() INode
	Likely() bool
	Line() string
	List() Nodes
	MarkNonNil()
	MarkReadonly()
	MayBeShared() bool
	Name() *Name
	NoInline() bool
	NonNil() bool
	Op() Op
	Opt() interface{}
	Orig() INode
	Pos() src.XPos
	PtrList() *Nodes
	PtrNbody() *Nodes
	PtrNinit() *Nodes
	PtrRlist() *Nodes
	RawCopy() INode
	ResetAux()
	Right() INode
	SetBounded(b bool)
	SetClass(b Class)
	SetColas(b bool)
	SetDiag(b bool)
	SetEmbedded(b bool)
	SetEsc(x uint16)
	SetFunc(x *Func)
	SetHasBreak(b bool)
	SetHasCall(b bool)
	SetImplicit(b bool)
	SetIndexMapLValue(b bool)
	SetInitorder(b uint8)
	SetIota(x int64)
	SetIsDDD(b bool)
	SetLeft(x INode)
	SetLikely(b bool)
	SetList(x Nodes)
	SetNbody(x Nodes)
	SetNinit(x Nodes)
	SetNoInline(b bool)
	SetOp(x Op)
	SetOpt(x interface{})
	SetOrig(x INode)
	SetPos(x src.XPos)
	SetRight(x INode)
	SetRlist(x Nodes)
	SetSliceBounds(low, high, max INode)
	SetSubOp(op Op)
	SetSym(x *types.Sym)
	SetTChanDir(dir types.ChanDir)
	SetTransient(b bool)
	SetType(x *types.Type)
	SetTypecheck(b uint8)
	SetVal(v Val)
	SetWalkdef(b uint8)
	SetXoffset(x int64)
	SliceBounds() (low, high, max INode)
	StorageClass() ssa.StorageClass
	String() string
	StringVal() string
	SubOp() Op
	Sym() *types.Sym
	TChanDir() types.ChanDir
	Transient() bool
	Typ() *types.Type
	Type() *types.Type
	Typecheck() uint8
	Val() Val
	Walkdef() uint8
	Xoffset() int64
}

type defaultNode struct {
	minimalNode
}

// minimalNode is an embeddable Node implementation
// that takes up no space and supplies any otherwise unimplemented methods
// needed to satisfy INode.
// The methods mostly panic.
type minimalNode struct {
	pos       src.XPos
	orig      INode
	esc       uint16
	typecheck uint8
}

func (n *minimalNode) BoolVal() bool                 { panic("unavailable") }
func (n *minimalNode) Bounded() bool                 { return false }
func (n *minimalNode) CanBeAnSSASym()                { panic("unavailable") }
func (n *minimalNode) CanInt64() bool                { return false }
func (n *minimalNode) Class() Class                  { panic("unavailable") }
func (n *minimalNode) Colas() bool                   { return false }
func (n *minimalNode) CopyFrom(INode)                { panic("unavailable") }
func (n *minimalNode) Diag() bool                    { return false }
func (n *minimalNode) Embedded() bool                { return false }
func (n *minimalNode) Esc() uint16                   { return n.esc }
func (n *minimalNode) Format(s fmt.State, verb rune) { panic("unavailable") }
func (n *minimalNode) Func() *Func                   { panic("unavailable") }
func (n *minimalNode) Nbody() Nodes                  { return Nodes{} }
func (n *minimalNode) Ninit() Nodes                  { return Nodes{} }
func (n *minimalNode) Rlist() Nodes                  { return Nodes{} }
func (n *minimalNode) HasBreak() bool                { return false }
func (n *minimalNode) HasCall() bool                 { return false }
func (n *minimalNode) HasVal() bool                  { return false }
func (n *minimalNode) Implicit() bool                { return false }
func (n *minimalNode) IndexMapLValue() bool          { return false }
func (n *minimalNode) Initorder() uint8              { panic("unavailable") }
func (n *minimalNode) Int64Val() int64               { panic("unavailable") }
func (n *minimalNode) Iota() int64                   { panic("unavailable") }
func (n *minimalNode) IsAutoTmp() bool               { return false }
func (n *minimalNode) IsBlank() bool                 { return false }
func (n *minimalNode) IsDDD() bool                   { return false }
func (n *minimalNode) IsMethod() bool                { return false }
func (n *minimalNode) IsNil() bool                   { return false }
func (n *minimalNode) IsSynthetic() bool             { return false }
func (n *minimalNode) Left() INode                   { return nil }
func (n *minimalNode) Likely() bool                  { panic("unavailable") }
func (n *minimalNode) Line() string                  { panic("unavailable") }
func (n *minimalNode) List() Nodes                   { return Nodes{} }
func (n *minimalNode) MarkNonNil()                   { panic("unavailable") }
func (n *minimalNode) MarkReadonly()                 { panic("unavailable") }
func (n *minimalNode) MayBeShared() bool             { return false }
func (n *minimalNode) Name() *Name                   { return nil }
func (n *minimalNode) NoInline() bool                { panic("unavailable") }
func (n *minimalNode) NonNil() bool                  { panic("unavailable") }
func (n *minimalNode) Opt() interface{}              { panic("unavailable") }
func (n *minimalNode) Orig() INode                   { return n.orig }
func (n *minimalNode) Pos() src.XPos                 { return n.pos }
func (n *minimalNode) PtrList() *Nodes               { return nil }
func (n *minimalNode) PtrNbody() *Nodes              { return nil }
func (n *minimalNode) PtrNinit() *Nodes              { return nil }
func (n *minimalNode) PtrRlist() *Nodes              { return nil }
func (n *minimalNode) ResetAux()                     { panic("unavailable") }
func (n *minimalNode) Right() INode                  { return nil }
func (n *minimalNode) SetBounded(b bool)             { panic("unavailable") }
func (n *minimalNode) SetClass(b Class)              { panic("unavailable") }
func (n *minimalNode) SetColas(b bool)               { panic("unavailable") }
func (n *minimalNode) SetDiag(b bool)                { panic("unavailable") }
func (n *minimalNode) SetEmbedded(b bool)            { panic("unavailable") }
func (n *minimalNode) SetEsc(x uint16)               { n.esc = x }
func (n *minimalNode) SetFunc(x *Func)               { panic("unavailable") }
func (n *minimalNode) SetHasBreak(b bool)            { panic("unavailable") }
func (n *minimalNode) SetHasCall(b bool) {
	if b {
		panic("unavailable")
	}
}
func (n *minimalNode) SetHasVal(b bool)         { panic("unavailable") }
func (n *minimalNode) SetImplicit(b bool)       { panic("unavailable") }
func (n *minimalNode) SetIndexMapLValue(b bool) { panic("unavailable") }
func (n *minimalNode) SetInitorder(b uint8)     { panic("unavailable") }
func (n *minimalNode) SetIota(x int64)          { panic("unavailable") }
func (n *minimalNode) SetIsDDD(b bool)          { panic("unavailable") }
func (n *minimalNode) SetLeft(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *minimalNode) SetLikely(b bool)     { panic("unavailable") }
func (n *minimalNode) SetList(x Nodes)      { panic("unavailable") }
func (n *minimalNode) SetNbody(x Nodes)     { panic("unavailable") }
func (n *minimalNode) SetNinit(x Nodes)     { panic("unavailable") }
func (n *minimalNode) SetNoInline(b bool)   { panic("unavailable") }
func (n *minimalNode) SetOp(x Op)           { panic("unavailable") }
func (n *minimalNode) SetOpt(x interface{}) { panic("unavailable") }
func (n *minimalNode) SetOrig(x INode) {
	n.orig = x
}
func (n *minimalNode) SetPos(x src.XPos) { n.pos = x }
func (n *minimalNode) SetRight(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *minimalNode) SetRlist(x Nodes)                    { panic("unavailable") }
func (n *minimalNode) SetSliceBounds(low, high, max INode) { panic("unavailable") }
func (n *minimalNode) SetSubOp(op Op)                      { panic("unavailable: SetSubOp " + op.String()) }
func (n *minimalNode) SetSym(x *types.Sym)                 { panic("unavailable") }
func (n *minimalNode) SetTChanDir(dir types.ChanDir)       { panic("unavailable") }
func (n *minimalNode) SetTransient(b bool)                 { panic("unavailable") }
func (n *minimalNode) SetType(x *types.Type)               { panic("unavailable") }
func (n *minimalNode) SetTypecheck(b uint8)                { n.typecheck = b }
func (n *minimalNode) SetVal(v Val)                        { panic("unavailable") }
func (n *minimalNode) SetWalkdef(b uint8)                  { panic("unavailable") }
func (n *minimalNode) SetXoffset(x int64) {
	if x != types.BADWIDTH {
		panic("unavailable")
	}
}

func (n *minimalNode) SliceBounds() (low, high, max INode) { panic("unavailable") }
func (n *minimalNode) StorageClass() ssa.StorageClass      { panic("unavailable") }
func (n *minimalNode) String() string                      { panic("unavailable") }
func (n *minimalNode) StringVal() string                   { panic("unavailable") }
func (n *minimalNode) SubOp() Op                           { panic("unavailable") }
func (n *minimalNode) Sym() *types.Sym                     { panic("unavailable") }
func (n *minimalNode) TChanDir() types.ChanDir             { panic("unavailable") }
func (n *minimalNode) Transient() bool                     { panic("unavailable") }
func (n *minimalNode) Typ() *types.Type                    { return nil }
func (n *minimalNode) Type() *types.Type                   { return nil }
func (n *minimalNode) Typecheck() uint8                    { return n.typecheck }
func (n *minimalNode) Val() Val                            { panic("unavailable") }
func (n *minimalNode) Walkdef() uint8                      { panic("unavailable") }
func (n *minimalNode) Xoffset() int64                      { panic("unavailable") }

type nodeFieldOpt struct{ opt interface{} }

func (n *nodeFieldOpt) Opt() interface{}     { return n.opt }
func (n *nodeFieldOpt) SetOpt(x interface{}) { n.opt = x }

type nodeFieldType struct{ typ *types.Type }

func (n *nodeFieldType) Type() *types.Type     { return n.typ }
func (n *nodeFieldType) SetType(x *types.Type) { n.typ = x }

type nodeFieldTransient struct{ transient bool }

func (n *nodeFieldTransient) Transient() bool     { return n.transient }
func (n *nodeFieldTransient) SetTransient(x bool) { n.transient = x }
