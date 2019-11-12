// Code generated from gen/decArgs.rules; DO NOT EDIT.
// generated with: cd gen; go run *.go

package ssa

import "cmd/compile/internal/types"

func rewriteValuedecArgs(v *Value) bool {
	switch v.Op {
	case OpArg:
		return rewriteValuedecArgs_OpArg_0(v)
	}
	return false
}
func rewriteValuedecArgs_OpArg_0(v *Value) bool {
	b := v.Block
	config := b.Func.Config
	typ := &b.Func.Config.Types
	// match: (Arg {n} [off])
	// cond: v.Type.IsString()
	// result: (StringMake (Arg <typ.BytePtr> {n} [off]) (Arg <typ.Int> {n} [off+config.PtrSize]))
	for {
		off := v.AuxInt
		n := v.Aux
		if !(v.Type.IsString()) {
			break
		}
		v.reset(OpStringMake)
		v0 := b.NewValue0(v.Pos, OpArg, typ.BytePtr)
		v0.AuxInt = off
		v0.Aux = n
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpArg, typ.Int)
		v1.AuxInt = off + config.PtrSize
		v1.Aux = n
		v.AddArg(v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: v.Type.IsSlice()
	// result: (SliceMake (Arg <v.Type.Elem().PtrTo()> {n} [off]) (Arg <typ.Int> {n} [off+config.PtrSize]) (Arg <typ.Int> {n} [off+2*config.PtrSize]))
	for {
		off := v.AuxInt
		n := v.Aux
		if !(v.Type.IsSlice()) {
			break
		}
		v.reset(OpSliceMake)
		v0 := b.NewValue0(v.Pos, OpArg, v.Type.Elem().PtrTo())
		v0.AuxInt = off
		v0.Aux = n
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpArg, typ.Int)
		v1.AuxInt = off + config.PtrSize
		v1.Aux = n
		v.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpArg, typ.Int)
		v2.AuxInt = off + 2*config.PtrSize
		v2.Aux = n
		v.AddArg(v2)
		return true
	}
	// match: (Arg {n} [off])
	// cond: v.Type.IsInterface()
	// result: (IMake (Arg <typ.Uintptr> {n} [off]) (Arg <typ.BytePtr> {n} [off+config.PtrSize]))
	for {
		off := v.AuxInt
		n := v.Aux
		if !(v.Type.IsInterface()) {
			break
		}
		v.reset(OpIMake)
		v0 := b.NewValue0(v.Pos, OpArg, typ.Uintptr)
		v0.AuxInt = off
		v0.Aux = n
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpArg, typ.BytePtr)
		v1.AuxInt = off + config.PtrSize
		v1.Aux = n
		v.AddArg(v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: v.Type.IsComplex() && v.Type.Size() == 16
	// result: (ComplexMake (Arg <typ.Float64> {n} [off]) (Arg <typ.Float64> {n} [off+8]))
	for {
		off := v.AuxInt
		n := v.Aux
		if !(v.Type.IsComplex() && v.Type.Size() == 16) {
			break
		}
		v.reset(OpComplexMake)
		v0 := b.NewValue0(v.Pos, OpArg, typ.Float64)
		v0.AuxInt = off
		v0.Aux = n
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpArg, typ.Float64)
		v1.AuxInt = off + 8
		v1.Aux = n
		v.AddArg(v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: v.Type.IsComplex() && v.Type.Size() == 8
	// result: (ComplexMake (Arg <typ.Float32> {n} [off]) (Arg <typ.Float32> {n} [off+4]))
	for {
		off := v.AuxInt
		n := v.Aux
		if !(v.Type.IsComplex() && v.Type.Size() == 8) {
			break
		}
		v.reset(OpComplexMake)
		v0 := b.NewValue0(v.Pos, OpArg, typ.Float32)
		v0.AuxInt = off
		v0.Aux = n
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpArg, typ.Float32)
		v1.AuxInt = off + 4
		v1.Aux = n
		v.AddArg(v1)
		return true
	}
	// match: (Arg <t>)
	// cond: t.IsStruct()
	// result: { argStruct(v) }
	for {
		t := v.Type
		if !(t.IsStruct()) {
			break
		}
		argStruct(v)
		return true
	}
	// match: (Arg <t>)
	// cond: t.IsArray() && t.NumElem() == 0
	// result: (ArrayMake)
	for {
		t := v.Type
		if !(t.IsArray() && t.NumElem() == 0) {
			break
		}
		v.reset(OpArrayMake)
		return true
	}
	// match: (Arg <t> {n} [off])
	// cond: t.IsArray() && t.NumElem() == 1
	// result: (ArrayUpdate (ArrayMake <t>) (Const64 <types.Types[types.TINT]> [0]) (Arg <t.Elem()> {n} [off]))
	for {
		t := v.Type
		off := v.AuxInt
		n := v.Aux
		if !(t.IsArray() && t.NumElem() == 1) {
			break
		}
		v.reset(OpArrayUpdate)
		v0 := b.NewValue0(v.Pos, OpArrayMake, t)
		v.AddArg(v0)
		v1 := b.NewValue0(v.Pos, OpConst64, types.Types[types.TINT])
		v1.AuxInt = 0
		v.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpArg, t.Elem())
		v2.AuxInt = off
		v2.Aux = n
		v.AddArg(v2)
		return true
	}
	return false
}
func rewriteBlockdecArgs(b *Block) bool {
	switch b.Kind {
	}
	return false
}
