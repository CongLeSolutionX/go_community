// Code generated from gen/ARM64latelower.rules; DO NOT EDIT.
// generated with: cd gen; go run *.go

package ssa

func rewriteValueARM64latelower(v *Value) bool {
	switch v.Op {
	case OpARM64ADDSconstflags:
		return rewriteValueARM64latelower_OpARM64ADDSconstflags(v)
	case OpARM64ADDconst:
		return rewriteValueARM64latelower_OpARM64ADDconst(v)
	case OpARM64ANDconst:
		return rewriteValueARM64latelower_OpARM64ANDconst(v)
	case OpARM64CMNWconst:
		return rewriteValueARM64latelower_OpARM64CMNWconst(v)
	case OpARM64CMNconst:
		return rewriteValueARM64latelower_OpARM64CMNconst(v)
	case OpARM64CMPWconst:
		return rewriteValueARM64latelower_OpARM64CMPWconst(v)
	case OpARM64CMPconst:
		return rewriteValueARM64latelower_OpARM64CMPconst(v)
	case OpARM64InvertFlags:
		return rewriteValueARM64latelower_OpARM64InvertFlags(v)
	case OpARM64ORconst:
		return rewriteValueARM64latelower_OpARM64ORconst(v)
	case OpARM64SUBconst:
		return rewriteValueARM64latelower_OpARM64SUBconst(v)
	case OpARM64TSTWconst:
		return rewriteValueARM64latelower_OpARM64TSTWconst(v)
	case OpARM64TSTconst:
		return rewriteValueARM64latelower_OpARM64TSTconst(v)
	case OpARM64XORconst:
		return rewriteValueARM64latelower_OpARM64XORconst(v)
	}
	return false
}
func rewriteValueARM64latelower_OpARM64ADDSconstflags(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ADDSconstflags [c] x)
	// cond: !isaddcon(c)
	// result: (ADDSflags x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64ADDSflags)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64ADDconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ADDconst [c] x)
	// cond: !isaddcon(c)
	// result: (ADD x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64ADD)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64ANDconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ANDconst [c] x)
	// cond: !isbitcon(uint64(c))
	// result: (AND x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isbitcon(uint64(c))) {
			break
		}
		v.reset(OpARM64AND)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64CMNWconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (CMNWconst [c] x)
	// cond: !isaddcon(int64(c))
	// result: (CMNW x (MOVDconst [int64(c)]))
	for {
		c := auxIntToInt32(v.AuxInt)
		x := v_0
		if !(!isaddcon(int64(c))) {
			break
		}
		v.reset(OpARM64CMNW)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(int64(c))
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64CMNconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (CMNconst [c] x)
	// cond: !isaddcon(c)
	// result: (CMN x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64CMN)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64CMPWconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (CMPWconst [c] x)
	// cond: !isaddcon(int64(c))
	// result: (CMPW x (MOVDconst [int64(c)]))
	for {
		c := auxIntToInt32(v.AuxInt)
		x := v_0
		if !(!isaddcon(int64(c))) {
			break
		}
		v.reset(OpARM64CMPW)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(int64(c))
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64CMPconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (CMPconst [c] x)
	// cond: !isaddcon(c)
	// result: (CMP x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64CMP)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64InvertFlags(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (InvertFlags (CMPconst [c] x))
	// cond: !isaddcon(c)
	// result: (CMP (MOVDconst [c]) x)
	for {
		if v_0.Op != OpARM64CMPconst {
			break
		}
		c := auxIntToInt64(v_0.AuxInt)
		x := v_0.Args[0]
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64CMP)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(v0, x)
		return true
	}
	// match: (InvertFlags (CMPWconst [c] x))
	// cond: !isaddcon(int64(c))
	// result: (CMPW (MOVDconst [int64(c)]) x)
	for {
		if v_0.Op != OpARM64CMPWconst {
			break
		}
		c := auxIntToInt32(v_0.AuxInt)
		x := v_0.Args[0]
		if !(!isaddcon(int64(c))) {
			break
		}
		v.reset(OpARM64CMPW)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(int64(c))
		v.AddArg2(v0, x)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64ORconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ORconst [c] x)
	// cond: !isbitcon(uint64(c))
	// result: (OR x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isbitcon(uint64(c))) {
			break
		}
		v.reset(OpARM64OR)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64SUBconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SUBconst [c] x)
	// cond: !isaddcon(c)
	// result: (SUB x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isaddcon(c)) {
			break
		}
		v.reset(OpARM64SUB)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64TSTWconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (TSTWconst [c] x)
	// cond: !isbitcon(uint64(c)|uint64(c)<<32)
	// result: (TSTW x (MOVDconst [int64(c)]))
	for {
		c := auxIntToInt32(v.AuxInt)
		x := v_0
		if !(!isbitcon(uint64(c) | uint64(c)<<32)) {
			break
		}
		v.reset(OpARM64TSTW)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(int64(c))
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64TSTconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (TSTconst [c] x)
	// cond: !isbitcon(uint64(c))
	// result: (TST x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isbitcon(uint64(c))) {
			break
		}
		v.reset(OpARM64TST)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteValueARM64latelower_OpARM64XORconst(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (XORconst [c] x)
	// cond: !isbitcon(uint64(c))
	// result: (XOR x (MOVDconst [c]))
	for {
		c := auxIntToInt64(v.AuxInt)
		x := v_0
		if !(!isbitcon(uint64(c))) {
			break
		}
		v.reset(OpARM64XOR)
		v0 := b.NewValue0(v.Pos, OpARM64MOVDconst, typ.UInt64)
		v0.AuxInt = int64ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	return false
}
func rewriteBlockARM64latelower(b *Block) bool {
	return false
}
