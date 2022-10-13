// Code generated from _gen/PPC64latelower.rules; DO NOT EDIT.
// generated with: cd _gen; go run .

package ssa

import "cmd/compile/internal/types"

func rewriteValuePPC64latelower(v *Value) bool {
	switch v.Op {
	case OpPPC64CMPconst:
		return rewriteValuePPC64latelower_OpPPC64CMPconst(v)
	case OpPPC64ISEL:
		return rewriteValuePPC64latelower_OpPPC64ISEL(v)
	}
	return false
}
func rewriteValuePPC64latelower_OpPPC64CMPconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (CMPconst [0] a:(AND y z))
	// cond: a.Uses == 1
	// result: (Select1 <types.TypeFlags> (ANDCC y z ))
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		a := v_0
		if a.Op != OpPPC64AND {
			break
		}
		z := a.Args[1]
		y := a.Args[0]
		if !(a.Uses == 1) {
			break
		}
		v.reset(OpSelect1)
		v.Type = types.TypeFlags
		v0 := b.NewValue0(v.Pos, OpPPC64ANDCC, types.NewTuple(typ.Int64, types.TypeFlags))
		v0.AddArg2(y, z)
		v.AddArg(v0)
		return true
	}
	// match: (CMPconst [0] a:(XOR y z))
	// cond: a.Uses == 1
	// result: (Select1 <types.TypeFlags> (XORCC y z ))
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		a := v_0
		if a.Op != OpPPC64XOR {
			break
		}
		z := a.Args[1]
		y := a.Args[0]
		if !(a.Uses == 1) {
			break
		}
		v.reset(OpSelect1)
		v.Type = types.TypeFlags
		v0 := b.NewValue0(v.Pos, OpPPC64XORCC, types.NewTuple(typ.Int, types.TypeFlags))
		v0.AddArg2(y, z)
		v.AddArg(v0)
		return true
	}
	// match: (CMPconst [0] a:(OR y z))
	// cond: a.Uses == 1
	// result: (Select1 <types.TypeFlags> (ORCC y z ))
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		a := v_0
		if a.Op != OpPPC64OR {
			break
		}
		z := a.Args[1]
		y := a.Args[0]
		if !(a.Uses == 1) {
			break
		}
		v.reset(OpSelect1)
		v.Type = types.TypeFlags
		v0 := b.NewValue0(v.Pos, OpPPC64ORCC, types.NewTuple(typ.Int, types.TypeFlags))
		v0.AddArg2(y, z)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValuePPC64latelower_OpPPC64ISEL(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (ISEL [a] x (MOVDconst [0]) z)
	// result: (ISELZ [a] x z)
	for {
		a := auxIntToInt32(v.AuxInt)
		x := v_0
		if v_1.Op != OpPPC64MOVDconst || auxIntToInt64(v_1.AuxInt) != 0 {
			break
		}
		z := v_2
		v.reset(OpPPC64ISELZ)
		v.AuxInt = int32ToAuxInt(a)
		v.AddArg2(x, z)
		return true
	}
	// match: (ISEL [a] (MOVDconst [0]) y z)
	// result: (ISELZ [a^0x4] y z)
	for {
		a := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpPPC64MOVDconst || auxIntToInt64(v_0.AuxInt) != 0 {
			break
		}
		y := v_1
		z := v_2
		v.reset(OpPPC64ISELZ)
		v.AuxInt = int32ToAuxInt(a ^ 0x4)
		v.AddArg2(y, z)
		return true
	}
	return false
}
func rewriteBlockPPC64latelower(b *Block) bool {
	return false
}
