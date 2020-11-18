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
	Copy() INode
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
	HasOpt() bool
	HasVal() bool
	Implicit() bool
	IndexMapLValue() bool
	Initorder() uint8
	Int64Val() int64
	Iota() int64
	IsAutoTmp() bool
	IsBlank() bool
	IsDDD() bool
	IsMethod() bool
	IsMethodExpression() bool
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
	SepCopy() INode
	SetBounded(b bool)
	SetClass(b Class)
	SetColas(b bool)
	SetDiag(b bool)
	SetEmbedded(b bool)
	SetEsc(x uint16)
	SetFunc(x *Func)
	SetHasBreak(b bool)
	SetHasCall(b bool)
	SetHasOpt(b bool)
	SetHasVal(b bool)
	SetImplicit(b bool)
	SetIndexMapLValue(b bool)
	SetInitorder(b uint8)
	SetIota(x int64)
	SetIsDDD(b bool)
	SetLeft(x INode)
	SetLikely(b bool)
	SetList(x Nodes)
	SetName(x *Name)
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
	Clear()
	CopyFrom(INode)
}

// PanicNode is an embeddable Node implementation
// that takes up no space and supplies any otherwise unimplemented methods
// needed to satisfy INode.
// The methods all panic.
type PanicNode struct{}

func (n *PanicNode) BoolVal() bool                       { panic("unavailable") }
func (n *PanicNode) Bounded() bool                       { panic("unavailable") }
func (n *PanicNode) CanBeAnSSASym()                      { panic("unavailable") }
func (n *PanicNode) CanInt64() bool                      { panic("unavailable") }
func (n *PanicNode) Class() Class                        { panic("unavailable") }
func (n *PanicNode) Colas() bool                         { panic("unavailable") }
func (n *PanicNode) Copy() INode                         { panic("unavailable") }
func (n *PanicNode) Diag() bool                          { panic("unavailable") }
func (n *PanicNode) Embedded() bool                      { panic("unavailable") }
func (n *PanicNode) Esc() uint16                         { panic("unavailable") }
func (n *PanicNode) Format(s fmt.State, verb rune)       { panic("unavailable") }
func (n *PanicNode) Func() *Func                         { panic("unavailable") }
func (n *PanicNode) Nbody() Nodes                        { panic("unavailable") }
func (n *PanicNode) Ninit() Nodes                        { panic("unavailable") }
func (n *PanicNode) Rlist() Nodes                        { panic("unavailable") }
func (n *PanicNode) HasBreak() bool                      { panic("unavailable") }
func (n *PanicNode) HasCall() bool                       { panic("unavailable") }
func (n *PanicNode) HasOpt() bool                        { panic("unavailable") }
func (n *PanicNode) HasVal() bool                        { panic("unavailable") }
func (n *PanicNode) Implicit() bool                      { panic("unavailable") }
func (n *PanicNode) IndexMapLValue() bool                { panic("unavailable") }
func (n *PanicNode) Initorder() uint8                    { panic("unavailable") }
func (n *PanicNode) Int64Val() int64                     { panic("unavailable") }
func (n *PanicNode) Iota() int64                         { panic("unavailable") }
func (n *PanicNode) IsAutoTmp() bool                     { panic("unavailable") }
func (n *PanicNode) IsBlank() bool                       { panic("unavailable") }
func (n *PanicNode) IsDDD() bool                         { panic("unavailable") }
func (n *PanicNode) IsMethod() bool                      { panic("unavailable") }
func (n *PanicNode) IsMethodExpression() bool            { panic("unavailable") }
func (n *PanicNode) IsNil() bool                         { panic("unavailable") }
func (n *PanicNode) IsSynthetic() bool                   { panic("unavailable") }
func (n *PanicNode) Left() INode                         { panic("unavailable") }
func (n *PanicNode) Likely() bool                        { panic("unavailable") }
func (n *PanicNode) Line() string                        { panic("unavailable") }
func (n *PanicNode) List() Nodes                         { panic("unavailable") }
func (n *PanicNode) MarkNonNil()                         { panic("unavailable") }
func (n *PanicNode) MarkReadonly()                       { panic("unavailable") }
func (n *PanicNode) MayBeShared() bool                   { panic("unavailable") }
func (n *PanicNode) Name() *Name                         { panic("unavailable") }
func (n *PanicNode) NoInline() bool                      { panic("unavailable") }
func (n *PanicNode) NonNil() bool                        { panic("unavailable") }
func (n *PanicNode) Op() Op                              { panic("unavailable") }
func (n *PanicNode) Opt() interface{}                    { panic("unavailable") }
func (n *PanicNode) Orig() INode                         { panic("unavailable") }
func (n *PanicNode) Pos() src.XPos                       { panic("unavailable") }
func (n *PanicNode) PtrList() *Nodes                     { panic("unavailable") }
func (n *PanicNode) PtrNbody() *Nodes                    { panic("unavailable") }
func (n *PanicNode) PtrNinit() *Nodes                    { panic("unavailable") }
func (n *PanicNode) PtrRlist() *Nodes                    { panic("unavailable") }
func (n *PanicNode) RawCopy() INode                      { panic("unavailable") }
func (n *PanicNode) ResetAux()                           { panic("unavailable") }
func (n *PanicNode) Right() INode                        { panic("unavailable") }
func (n *PanicNode) SepCopy() INode                      { panic("unavailable") }
func (n *PanicNode) SetBounded(b bool)                   { panic("unavailable") }
func (n *PanicNode) SetClass(b Class)                    { panic("unavailable") }
func (n *PanicNode) SetColas(b bool)                     { panic("unavailable") }
func (n *PanicNode) SetDiag(b bool)                      { panic("unavailable") }
func (n *PanicNode) SetEmbedded(b bool)                  { panic("unavailable") }
func (n *PanicNode) SetEsc(x uint16)                     { panic("unavailable") }
func (n *PanicNode) SetFunc(x *Func)                     { panic("unavailable") }
func (n *PanicNode) SetHasBreak(b bool)                  { panic("unavailable") }
func (n *PanicNode) SetHasCall(b bool)                   { panic("unavailable") }
func (n *PanicNode) SetHasOpt(b bool)                    { panic("unavailable") }
func (n *PanicNode) SetHasVal(b bool)                    { panic("unavailable") }
func (n *PanicNode) SetImplicit(b bool)                  { panic("unavailable") }
func (n *PanicNode) SetIndexMapLValue(b bool)            { panic("unavailable") }
func (n *PanicNode) SetInitorder(b uint8)                { panic("unavailable") }
func (n *PanicNode) SetIota(x int64)                     { panic("unavailable") }
func (n *PanicNode) SetIsDDD(b bool)                     { panic("unavailable") }
func (n *PanicNode) SetLeft(x *Node)                     { panic("unavailable") }
func (n *PanicNode) SetLikely(b bool)                    { panic("unavailable") }
func (n *PanicNode) SetList(x Nodes)                     { panic("unavailable") }
func (n *PanicNode) SetName(x *Name)                     { panic("unavailable") }
func (n *PanicNode) SetNbody(x Nodes)                    { panic("unavailable") }
func (n *PanicNode) SetNinit(x Nodes)                    { panic("unavailable") }
func (n *PanicNode) SetNoInline(b bool)                  { panic("unavailable") }
func (n *PanicNode) SetOp(x Op)                          { panic("unavailable") }
func (n *PanicNode) SetOpt(x interface{})                { panic("unavailable") }
func (n *PanicNode) SetOrig(x *Node)                     { panic("unavailable") }
func (n *PanicNode) SetPos(x src.XPos)                   { panic("unavailable") }
func (n *PanicNode) SetRight(x *Node)                    { panic("unavailable") }
func (n *PanicNode) SetRlist(x Nodes)                    { panic("unavailable") }
func (n *PanicNode) SetSliceBounds(low, high, max *Node) { panic("unavailable") }
func (n *PanicNode) SetSubOp(op Op)                      { panic("unavailable") }
func (n *PanicNode) SetSym(x *types.Sym)                 { panic("unavailable") }
func (n *PanicNode) SetTChanDir(dir types.ChanDir)       { panic("unavailable") }
func (n *PanicNode) SetTransient(b bool)                 { panic("unavailable") }
func (n *PanicNode) SetType(x *types.Type)               { panic("unavailable") }
func (n *PanicNode) SetTypecheck(b uint8)                { panic("unavailable") }
func (n *PanicNode) SetVal(v Val)                        { panic("unavailable") }
func (n *PanicNode) SetWalkdef(b uint8)                  { panic("unavailable") }
func (n *PanicNode) SetXoffset(x int64)                  { panic("unavailable") }
func (n *PanicNode) SliceBounds() (low, high, max *Node) { panic("unavailable") }
func (n *PanicNode) StorageClass() ssa.StorageClass      { panic("unavailable") }
func (n *PanicNode) String() string                      { panic("unavailable") }
func (n *PanicNode) StringVal() string                   { panic("unavailable") }
func (n *PanicNode) SubOp() Op                           { panic("unavailable") }
func (n *PanicNode) Sym() *types.Sym                     { panic("unavailable") }
func (n *PanicNode) TChanDir() types.ChanDir             { panic("unavailable") }
func (n *PanicNode) Transient() bool                     { panic("unavailable") }
func (n *PanicNode) Typ() *types.Type                    { panic("unavailable") }
func (n *PanicNode) Type() *types.Type                   { panic("unavailable") }
func (n *PanicNode) Typecheck() uint8                    { panic("unavailable") }
func (n *PanicNode) Val() Val                            { panic("unavailable") }
func (n *PanicNode) Walkdef() uint8                      { panic("unavailable") }
func (n *PanicNode) Xoffset() int64                      { panic("unavailable") }
