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

// TrivNode is an embeddable Node implementation
// that takes up no space and supplies any otherwise unimplemented methods
// needed to satisfy INode.
// The methods mostly panic.
type TrivNode struct {
	pos       src.XPos
	orig      INode
	typecheck uint8
}

func (n *TrivNode) BoolVal() bool                 { panic("unavailable") }
func (n *TrivNode) Bounded() bool                 { return false }
func (n *TrivNode) CanBeAnSSASym()                { panic("unavailable") }
func (n *TrivNode) CanInt64() bool                { return false }
func (n *TrivNode) Class() Class                  { panic("unavailable") }
func (n *TrivNode) Colas() bool                   { return false }
func (n *TrivNode) CopyFrom(INode)                { panic("unavailable") }
func (n *TrivNode) Diag() bool                    { return false }
func (n *TrivNode) Embedded() bool                { return false }
func (n *TrivNode) Esc() uint16                   { panic("unavailable") }
func (n *TrivNode) Format(s fmt.State, verb rune) { panic("unavailable") }
func (n *TrivNode) Func() *Func                   { panic("unavailable") }
func (n *TrivNode) Nbody() Nodes                  { return Nodes{} }
func (n *TrivNode) Ninit() Nodes                  { return Nodes{} }
func (n *TrivNode) Rlist() Nodes                  { return Nodes{} }
func (n *TrivNode) HasBreak() bool                { return false }
func (n *TrivNode) HasCall() bool                 { return false }
func (n *TrivNode) HasVal() bool                  { return false }
func (n *TrivNode) Implicit() bool                { return false }
func (n *TrivNode) IndexMapLValue() bool          { return false }
func (n *TrivNode) Initorder() uint8              { panic("unavailable") }
func (n *TrivNode) Int64Val() int64               { panic("unavailable") }
func (n *TrivNode) Iota() int64                   { panic("unavailable") }
func (n *TrivNode) IsAutoTmp() bool               { return false }
func (n *TrivNode) IsBlank() bool                 { return false }
func (n *TrivNode) IsDDD() bool                   { return false }
func (n *TrivNode) IsMethod() bool                { return false }
func (n *TrivNode) IsNil() bool                   { return false }
func (n *TrivNode) IsSynthetic() bool             { return false }
func (n *TrivNode) Left() INode                   { return nil }
func (n *TrivNode) Likely() bool                  { panic("unavailable") }
func (n *TrivNode) Line() string                  { panic("unavailable") }
func (n *TrivNode) List() Nodes                   { return Nodes{} }
func (n *TrivNode) MarkNonNil()                   { panic("unavailable") }
func (n *TrivNode) MarkReadonly()                 { panic("unavailable") }
func (n *TrivNode) MayBeShared() bool             { return false }
func (n *TrivNode) Name() *Name                   { return nil }
func (n *TrivNode) NoInline() bool                { panic("unavailable") }
func (n *TrivNode) NonNil() bool                  { panic("unavailable") }
func (n *TrivNode) Opt() interface{}              { panic("unavailable") }
func (n *TrivNode) Orig() INode                   { return n.orig }
func (n *TrivNode) Pos() src.XPos                 { return n.pos }
func (n *TrivNode) PtrList() *Nodes               { return nil }
func (n *TrivNode) PtrNbody() *Nodes              { return nil }
func (n *TrivNode) PtrNinit() *Nodes              { return nil }
func (n *TrivNode) PtrRlist() *Nodes              { return nil }
func (n *TrivNode) ResetAux()                     { panic("unavailable") }
func (n *TrivNode) Right() INode                  { return nil }
func (n *TrivNode) SetBounded(b bool)             { panic("unavailable") }
func (n *TrivNode) SetClass(b Class)              { panic("unavailable") }
func (n *TrivNode) SetColas(b bool)               { panic("unavailable") }
func (n *TrivNode) SetDiag(b bool)                { panic("unavailable") }
func (n *TrivNode) SetEmbedded(b bool)            { panic("unavailable") }
func (n *TrivNode) SetEsc(x uint16)               { panic("unavailable") }
func (n *TrivNode) SetFunc(x *Func)               { panic("unavailable") }
func (n *TrivNode) SetHasBreak(b bool)            { panic("unavailable") }
func (n *TrivNode) SetHasCall(b bool) {
	if b {
		panic("unavailable")
	}
}
func (n *TrivNode) SetHasVal(b bool)         { panic("unavailable") }
func (n *TrivNode) SetImplicit(b bool)       { panic("unavailable") }
func (n *TrivNode) SetIndexMapLValue(b bool) { panic("unavailable") }
func (n *TrivNode) SetInitorder(b uint8)     { panic("unavailable") }
func (n *TrivNode) SetIota(x int64)          { panic("unavailable") }
func (n *TrivNode) SetIsDDD(b bool)          { panic("unavailable") }
func (n *TrivNode) SetLeft(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *TrivNode) SetLikely(b bool)     { panic("unavailable") }
func (n *TrivNode) SetList(x Nodes)      { panic("unavailable") }
func (n *TrivNode) SetNbody(x Nodes)     { panic("unavailable") }
func (n *TrivNode) SetNinit(x Nodes)     { panic("unavailable") }
func (n *TrivNode) SetNoInline(b bool)   { panic("unavailable") }
func (n *TrivNode) SetOp(x Op)           { panic("unavailable") }
func (n *TrivNode) SetOpt(x interface{}) { panic("unavailable") }
func (n *TrivNode) SetOrig(x INode) {
	n.orig = x
}
func (n *TrivNode) SetPos(x src.XPos) { n.pos = x }
func (n *TrivNode) SetRight(x INode) {
	if x != nil {
		panic("unavailable")
	}
}
func (n *TrivNode) SetRlist(x Nodes)                    { panic("unavailable") }
func (n *TrivNode) SetSliceBounds(low, high, max INode) { panic("unavailable") }
func (n *TrivNode) SetSubOp(op Op)                      { panic("unavailable: SetSubOp " + op.String()) }
func (n *TrivNode) SetSym(x *types.Sym)                 { panic("unavailable") }
func (n *TrivNode) SetTChanDir(dir types.ChanDir)       { panic("unavailable") }
func (n *TrivNode) SetTransient(b bool)                 { panic("unavailable") }
func (n *TrivNode) SetType(x *types.Type)               { panic("unavailable") }
func (n *TrivNode) SetTypecheck(b uint8)                { n.typecheck = b }
func (n *TrivNode) SetVal(v Val)                        { panic("unavailable") }
func (n *TrivNode) SetWalkdef(b uint8)                  { panic("unavailable") }
func (n *TrivNode) SetXoffset(x int64) {
	if x != types.BADWIDTH {
		panic("unavailable")
	}
}

func (n *TrivNode) SliceBounds() (low, high, max INode) { panic("unavailable") }
func (n *TrivNode) StorageClass() ssa.StorageClass      { panic("unavailable") }
func (n *TrivNode) String() string                      { panic("unavailable") }
func (n *TrivNode) StringVal() string                   { panic("unavailable") }
func (n *TrivNode) SubOp() Op                           { panic("unavailable") }
func (n *TrivNode) Sym() *types.Sym                     { panic("unavailable") }
func (n *TrivNode) TChanDir() types.ChanDir             { panic("unavailable") }
func (n *TrivNode) Transient() bool                     { panic("unavailable") }
func (n *TrivNode) Typ() *types.Type                    { return nil }
func (n *TrivNode) Type() *types.Type                   { return nil }
func (n *TrivNode) Typecheck() uint8                    { return n.typecheck }
func (n *TrivNode) Val() Val                            { panic("unavailable") }
func (n *TrivNode) Walkdef() uint8                      { panic("unavailable") }
func (n *TrivNode) Xoffset() int64                      { panic("unavailable") }
