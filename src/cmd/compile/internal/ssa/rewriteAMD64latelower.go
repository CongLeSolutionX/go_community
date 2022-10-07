// Code generated from gen/AMD64latelower.rules; DO NOT EDIT.
// generated with: cd gen; go run *.go

package ssa

func rewriteValueAMD64latelower(v *Value) bool {
	switch v.Op {
	case OpAMD64LEAL1:
		return rewriteValueAMD64latelower_OpAMD64LEAL1(v)
	case OpAMD64LEAL4:
		return rewriteValueAMD64latelower_OpAMD64LEAL4(v)
	case OpAMD64LEAL8:
		return rewriteValueAMD64latelower_OpAMD64LEAL8(v)
	case OpAMD64LEAQ1:
		return rewriteValueAMD64latelower_OpAMD64LEAQ1(v)
	case OpAMD64LEAQ4:
		return rewriteValueAMD64latelower_OpAMD64LEAQ4(v)
	case OpAMD64LEAQ8:
		return rewriteValueAMD64latelower_OpAMD64LEAQ8(v)
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL1(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL1 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDLconst [c] (LEAL1 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDLconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL1, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL4(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL4 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDLconst [c] (LEAL4 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDLconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL4, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAL8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAL8 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDLconst [c] (LEAL8 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDLconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAL8, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ1(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ1 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDQconst [c] (LEAQ1 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDQconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ1, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ4(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ4 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDQconst [c] (LEAQ4 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDQconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ4, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueAMD64latelower_OpAMD64LEAQ8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	// match: (LEAQ8 <t> [c] {s} x y)
	// cond: c != 0
	// result: (ADDQconst [c] (LEAQ8 <t> [0] {s} x y))
	for {
		t := v.Type
		c := auxIntToInt32(v.AuxInt)
		s := auxToSym(v.Aux)
		x := v_0
		y := v_1
		if !(c != 0) {
			break
		}
		v.reset(OpAMD64ADDQconst)
		v.AuxInt = int32ToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpAMD64LEAQ8, t)
		v0.AuxInt = int32ToAuxInt(0)
		v0.Aux = symToAux(s)
		v0.AddArg2(x, y)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteBlockAMD64latelower(b *Block) bool {
	return false
}
