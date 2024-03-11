// Code generated from _gen/Wasm32.rules using 'go generate'; DO NOT EDIT.

package ssa

import "internal/buildcfg"
import "math"
import "cmd/compile/internal/types"

func rewriteValueWasm32(v *Value) bool {
	switch v.Op {
	case OpAbs:
		v.Op = OpWasm32F64Abs
		return true
	case OpAdd16:
		v.Op = OpWasm32I32Add
		return true
	case OpAdd32:
		v.Op = OpWasm32I32Add
		return true
	case OpAdd32F:
		v.Op = OpWasm32F32Add
		return true
	case OpAdd64:
		return rewriteValueWasm32_OpAdd64(v)
	case OpAdd64F:
		v.Op = OpWasm32F64Add
		return true
	case OpAdd8:
		v.Op = OpWasm32I32Add
		return true
	case OpAddPtr:
		v.Op = OpWasm32I32Add
		return true
	case OpAddr:
		return rewriteValueWasm32_OpAddr(v)
	case OpAnd16:
		v.Op = OpWasm32I32And
		return true
	case OpAnd32:
		v.Op = OpWasm32I32And
		return true
	case OpAnd64:
		return rewriteValueWasm32_OpAnd64(v)
	case OpAnd8:
		v.Op = OpWasm32I32And
		return true
	case OpAndB:
		v.Op = OpWasm32I32And
		return true
	case OpArg:
		return rewriteValueWasm32_OpArg(v)
	case OpBitLen32:
		return rewriteValueWasm32_OpBitLen32(v)
	case OpBitLen64:
		return rewriteValueWasm32_OpBitLen64(v)
	case OpBswap64:
		return rewriteValueWasm32_OpBswap64(v)
	case OpCeil:
		v.Op = OpWasm32F64Ceil
		return true
	case OpClosureCall:
		v.Op = OpWasm32LoweredClosureCall
		return true
	case OpCom16:
		return rewriteValueWasm32_OpCom16(v)
	case OpCom32:
		return rewriteValueWasm32_OpCom32(v)
	case OpCom64:
		return rewriteValueWasm32_OpCom64(v)
	case OpCom8:
		return rewriteValueWasm32_OpCom8(v)
	case OpCondSelect:
		v.Op = OpWasm32Select
		return true
	case OpConst16:
		return rewriteValueWasm32_OpConst16(v)
	case OpConst32:
		return rewriteValueWasm32_OpConst32(v)
	case OpConst32F:
		v.Op = OpWasm32F32Const
		return true
	case OpConst64:
		return rewriteValueWasm32_OpConst64(v)
	case OpConst64F:
		v.Op = OpWasm32F64Const
		return true
	case OpConst8:
		return rewriteValueWasm32_OpConst8(v)
	case OpConstBool:
		return rewriteValueWasm32_OpConstBool(v)
	case OpConstNil:
		return rewriteValueWasm32_OpConstNil(v)
	case OpConvert:
		v.Op = OpWasm32LoweredConvert
		return true
	case OpCopysign:
		v.Op = OpWasm32F64Copysign
		return true
	case OpCtz16:
		return rewriteValueWasm32_OpCtz16(v)
	case OpCtz16NonZero:
		v.Op = OpWasm32I32Ctz
		return true
	case OpCtz32:
		v.Op = OpWasm32I32Ctz
		return true
	case OpCtz32NonZero:
		v.Op = OpWasm32I32Ctz
		return true
	case OpCtz64:
		return rewriteValueWasm32_OpCtz64(v)
	case OpCtz64NonZero:
		v.Op = OpCtz64
		return true
	case OpCtz8:
		return rewriteValueWasm32_OpCtz8(v)
	case OpCtz8NonZero:
		v.Op = OpWasm32I32Ctz
		return true
	case OpCvt32Fto32:
		v.Op = OpWasm32I32TruncSatF32S
		return true
	case OpCvt32Fto32U:
		v.Op = OpWasm32I32TruncSatF32U
		return true
	case OpCvt32Fto64F:
		v.Op = OpWasm32F64PromoteF32
		return true
	case OpCvt32Uto32F:
		v.Op = OpWasm32F32ConvertI32U
		return true
	case OpCvt32Uto64F:
		return rewriteValueWasm32_OpCvt32Uto64F(v)
	case OpCvt32to32F:
		v.Op = OpWasm32F32ConvertI32S
		return true
	case OpCvt32to64F:
		return rewriteValueWasm32_OpCvt32to64F(v)
	case OpCvt64Fto32:
		v.Op = OpWasm32I32TruncSatF64S
		return true
	case OpCvt64Fto32F:
		v.Op = OpWasm32F32DemoteF64
		return true
	case OpCvt64Fto32U:
		v.Op = OpWasm32I64TruncSatF64U
		return true
	case OpCvtBoolToUint8:
		v.Op = OpCopy
		return true
	case OpDiv16:
		return rewriteValueWasm32_OpDiv16(v)
	case OpDiv16u:
		return rewriteValueWasm32_OpDiv16u(v)
	case OpDiv32:
		return rewriteValueWasm32_OpDiv32(v)
	case OpDiv32F:
		v.Op = OpWasm32F32Div
		return true
	case OpDiv32u:
		return rewriteValueWasm32_OpDiv32u(v)
	case OpDiv64:
		return rewriteValueWasm32_OpDiv64(v)
	case OpDiv64F:
		v.Op = OpWasm32F64Div
		return true
	case OpDiv64u:
		return rewriteValueWasm32_OpDiv64u(v)
	case OpDiv8:
		return rewriteValueWasm32_OpDiv8(v)
	case OpDiv8u:
		return rewriteValueWasm32_OpDiv8u(v)
	case OpEq16:
		return rewriteValueWasm32_OpEq16(v)
	case OpEq32:
		v.Op = OpWasm32I32Eq
		return true
	case OpEq32F:
		v.Op = OpWasm32F32Eq
		return true
	case OpEq64:
		return rewriteValueWasm32_OpEq64(v)
	case OpEq64F:
		v.Op = OpWasm32F64Eq
		return true
	case OpEq8:
		return rewriteValueWasm32_OpEq8(v)
	case OpEqB:
		v.Op = OpWasm32I32Eq
		return true
	case OpEqPtr:
		v.Op = OpWasm32I32Eq
		return true
	case OpFloor:
		v.Op = OpWasm32F64Floor
		return true
	case OpGetCallerPC:
		v.Op = OpWasm32LoweredGetCallerPC
		return true
	case OpGetCallerSP:
		v.Op = OpWasm32LoweredGetCallerSP
		return true
	case OpGetClosurePtr:
		v.Op = OpWasm32LoweredGetClosurePtr
		return true
	case OpInt64Hi:
		return rewriteValueWasm32_OpInt64Hi(v)
	case OpInt64Lo:
		return rewriteValueWasm32_OpInt64Lo(v)
	case OpInterCall:
		v.Op = OpWasm32LoweredInterCall
		return true
	case OpIsInBounds:
		v.Op = OpWasm32I32LtU
		return true
	case OpIsNonNil:
		return rewriteValueWasm32_OpIsNonNil(v)
	case OpIsSliceInBounds:
		v.Op = OpWasm32I32LeU
		return true
	case OpLeq16:
		return rewriteValueWasm32_OpLeq16(v)
	case OpLeq16U:
		return rewriteValueWasm32_OpLeq16U(v)
	case OpLeq32:
		v.Op = OpWasm32I32LeS
		return true
	case OpLeq32F:
		v.Op = OpWasm32F32Le
		return true
	case OpLeq32U:
		v.Op = OpWasm32I32LeU
		return true
	case OpLeq64:
		return rewriteValueWasm32_OpLeq64(v)
	case OpLeq64F:
		v.Op = OpWasm32F64Le
		return true
	case OpLeq64U:
		return rewriteValueWasm32_OpLeq64U(v)
	case OpLeq8:
		return rewriteValueWasm32_OpLeq8(v)
	case OpLeq8U:
		return rewriteValueWasm32_OpLeq8U(v)
	case OpLess16:
		return rewriteValueWasm32_OpLess16(v)
	case OpLess16U:
		return rewriteValueWasm32_OpLess16U(v)
	case OpLess32:
		v.Op = OpWasm32I32LtS
		return true
	case OpLess32F:
		v.Op = OpWasm32F32Lt
		return true
	case OpLess32U:
		v.Op = OpWasm32I32LtU
		return true
	case OpLess64:
		return rewriteValueWasm32_OpLess64(v)
	case OpLess64F:
		v.Op = OpWasm32F64Lt
		return true
	case OpLess64U:
		return rewriteValueWasm32_OpLess64U(v)
	case OpLess8:
		return rewriteValueWasm32_OpLess8(v)
	case OpLess8U:
		return rewriteValueWasm32_OpLess8U(v)
	case OpLoad:
		return rewriteValueWasm32_OpLoad(v)
	case OpLocalAddr:
		return rewriteValueWasm32_OpLocalAddr(v)
	case OpLsh16x16:
		return rewriteValueWasm32_OpLsh16x16(v)
	case OpLsh16x32:
		v.Op = OpLsh32x32
		return true
	case OpLsh16x64:
		return rewriteValueWasm32_OpLsh16x64(v)
	case OpLsh16x8:
		return rewriteValueWasm32_OpLsh16x8(v)
	case OpLsh32x16:
		return rewriteValueWasm32_OpLsh32x16(v)
	case OpLsh32x32:
		return rewriteValueWasm32_OpLsh32x32(v)
	case OpLsh32x64:
		return rewriteValueWasm32_OpLsh32x64(v)
	case OpLsh32x8:
		return rewriteValueWasm32_OpLsh32x8(v)
	case OpLsh64x16:
		return rewriteValueWasm32_OpLsh64x16(v)
	case OpLsh64x32:
		return rewriteValueWasm32_OpLsh64x32(v)
	case OpLsh64x64:
		return rewriteValueWasm32_OpLsh64x64(v)
	case OpLsh64x8:
		return rewriteValueWasm32_OpLsh64x8(v)
	case OpLsh8x16:
		return rewriteValueWasm32_OpLsh8x16(v)
	case OpLsh8x32:
		v.Op = OpLsh32x32
		return true
	case OpLsh8x64:
		return rewriteValueWasm32_OpLsh8x64(v)
	case OpLsh8x8:
		return rewriteValueWasm32_OpLsh8x8(v)
	case OpMod16:
		return rewriteValueWasm32_OpMod16(v)
	case OpMod16u:
		return rewriteValueWasm32_OpMod16u(v)
	case OpMod32:
		return rewriteValueWasm32_OpMod32(v)
	case OpMod32u:
		return rewriteValueWasm32_OpMod32u(v)
	case OpMod64:
		return rewriteValueWasm32_OpMod64(v)
	case OpMod64u:
		return rewriteValueWasm32_OpMod64u(v)
	case OpMod8:
		return rewriteValueWasm32_OpMod8(v)
	case OpMod8u:
		return rewriteValueWasm32_OpMod8u(v)
	case OpMove:
		return rewriteValueWasm32_OpMove(v)
	case OpMul16:
		v.Op = OpWasm32I32Mul
		return true
	case OpMul32:
		v.Op = OpWasm32I32Mul
		return true
	case OpMul32F:
		v.Op = OpWasm32F32Mul
		return true
	case OpMul64:
		return rewriteValueWasm32_OpMul64(v)
	case OpMul64F:
		v.Op = OpWasm32F64Mul
		return true
	case OpMul8:
		v.Op = OpWasm32I32Mul
		return true
	case OpNeg16:
		return rewriteValueWasm32_OpNeg16(v)
	case OpNeg32:
		return rewriteValueWasm32_OpNeg32(v)
	case OpNeg32F:
		v.Op = OpWasm32F32Neg
		return true
	case OpNeg64:
		return rewriteValueWasm32_OpNeg64(v)
	case OpNeg64F:
		v.Op = OpWasm32F64Neg
		return true
	case OpNeg8:
		return rewriteValueWasm32_OpNeg8(v)
	case OpNeq16:
		return rewriteValueWasm32_OpNeq16(v)
	case OpNeq32:
		return rewriteValueWasm32_OpNeq32(v)
	case OpNeq32F:
		v.Op = OpWasm32F32Ne
		return true
	case OpNeq64:
		return rewriteValueWasm32_OpNeq64(v)
	case OpNeq64F:
		v.Op = OpWasm32F64Ne
		return true
	case OpNeq8:
		return rewriteValueWasm32_OpNeq8(v)
	case OpNeqB:
		v.Op = OpWasm32I32Ne
		return true
	case OpNeqPtr:
		v.Op = OpWasm32I32Ne
		return true
	case OpNilCheck:
		v.Op = OpWasm32LoweredNilCheck
		return true
	case OpNot:
		v.Op = OpWasm32I32Eqz
		return true
	case OpOffPtr:
		return rewriteValueWasm32_OpOffPtr(v)
	case OpOr16:
		return rewriteValueWasm32_OpOr16(v)
	case OpOr32:
		return rewriteValueWasm32_OpOr32(v)
	case OpOr64:
		return rewriteValueWasm32_OpOr64(v)
	case OpOr8:
		return rewriteValueWasm32_OpOr8(v)
	case OpOrB:
		return rewriteValueWasm32_OpOrB(v)
	case OpPanicExtend:
		return rewriteValueWasm32_OpPanicExtend(v)
	case OpPopCount16:
		return rewriteValueWasm32_OpPopCount16(v)
	case OpPopCount32:
		return rewriteValueWasm32_OpPopCount32(v)
	case OpPopCount8:
		return rewriteValueWasm32_OpPopCount8(v)
	case OpRotateLeft16:
		return rewriteValueWasm32_OpRotateLeft16(v)
	case OpRotateLeft32:
		return rewriteValueWasm32_OpRotateLeft32(v)
	case OpRotateLeft64:
		return rewriteValueWasm32_OpRotateLeft64(v)
	case OpRotateLeft8:
		return rewriteValueWasm32_OpRotateLeft8(v)
	case OpRound32F:
		v.Op = OpCopy
		return true
	case OpRound64F:
		v.Op = OpCopy
		return true
	case OpRoundToEven:
		v.Op = OpWasm32F64Nearest
		return true
	case OpRsh16Ux16:
		return rewriteValueWasm32_OpRsh16Ux16(v)
	case OpRsh16Ux32:
		return rewriteValueWasm32_OpRsh16Ux32(v)
	case OpRsh16Ux64:
		return rewriteValueWasm32_OpRsh16Ux64(v)
	case OpRsh16Ux8:
		return rewriteValueWasm32_OpRsh16Ux8(v)
	case OpRsh16x16:
		return rewriteValueWasm32_OpRsh16x16(v)
	case OpRsh16x32:
		return rewriteValueWasm32_OpRsh16x32(v)
	case OpRsh16x64:
		return rewriteValueWasm32_OpRsh16x64(v)
	case OpRsh16x8:
		return rewriteValueWasm32_OpRsh16x8(v)
	case OpRsh32Ux16:
		return rewriteValueWasm32_OpRsh32Ux16(v)
	case OpRsh32Ux32:
		return rewriteValueWasm32_OpRsh32Ux32(v)
	case OpRsh32Ux64:
		return rewriteValueWasm32_OpRsh32Ux64(v)
	case OpRsh32Ux8:
		return rewriteValueWasm32_OpRsh32Ux8(v)
	case OpRsh32x16:
		return rewriteValueWasm32_OpRsh32x16(v)
	case OpRsh32x32:
		return rewriteValueWasm32_OpRsh32x32(v)
	case OpRsh32x64:
		return rewriteValueWasm32_OpRsh32x64(v)
	case OpRsh32x8:
		return rewriteValueWasm32_OpRsh32x8(v)
	case OpRsh64Ux16:
		return rewriteValueWasm32_OpRsh64Ux16(v)
	case OpRsh64Ux32:
		return rewriteValueWasm32_OpRsh64Ux32(v)
	case OpRsh64Ux64:
		return rewriteValueWasm32_OpRsh64Ux64(v)
	case OpRsh64Ux8:
		return rewriteValueWasm32_OpRsh64Ux8(v)
	case OpRsh64x16:
		return rewriteValueWasm32_OpRsh64x16(v)
	case OpRsh64x32:
		return rewriteValueWasm32_OpRsh64x32(v)
	case OpRsh64x64:
		return rewriteValueWasm32_OpRsh64x64(v)
	case OpRsh64x8:
		return rewriteValueWasm32_OpRsh64x8(v)
	case OpRsh8Ux16:
		return rewriteValueWasm32_OpRsh8Ux16(v)
	case OpRsh8Ux32:
		return rewriteValueWasm32_OpRsh8Ux32(v)
	case OpRsh8Ux64:
		return rewriteValueWasm32_OpRsh8Ux64(v)
	case OpRsh8Ux8:
		return rewriteValueWasm32_OpRsh8Ux8(v)
	case OpRsh8x16:
		return rewriteValueWasm32_OpRsh8x16(v)
	case OpRsh8x32:
		return rewriteValueWasm32_OpRsh8x32(v)
	case OpRsh8x64:
		return rewriteValueWasm32_OpRsh8x64(v)
	case OpRsh8x8:
		return rewriteValueWasm32_OpRsh8x8(v)
	case OpSignExt16to32:
		return rewriteValueWasm32_OpSignExt16to32(v)
	case OpSignExt16to64:
		return rewriteValueWasm32_OpSignExt16to64(v)
	case OpSignExt32to64:
		return rewriteValueWasm32_OpSignExt32to64(v)
	case OpSignExt8to16:
		return rewriteValueWasm32_OpSignExt8to16(v)
	case OpSignExt8to32:
		return rewriteValueWasm32_OpSignExt8to32(v)
	case OpSignExt8to64:
		return rewriteValueWasm32_OpSignExt8to64(v)
	case OpSignmask:
		return rewriteValueWasm32_OpSignmask(v)
	case OpSlicemask:
		return rewriteValueWasm32_OpSlicemask(v)
	case OpSqrt:
		v.Op = OpWasm32F64Sqrt
		return true
	case OpSqrt32:
		v.Op = OpWasm32F32Sqrt
		return true
	case OpStaticCall:
		v.Op = OpWasm32LoweredStaticCall
		return true
	case OpStore:
		return rewriteValueWasm32_OpStore(v)
	case OpSub16:
		v.Op = OpWasm32I32Sub
		return true
	case OpSub32:
		v.Op = OpWasm32I32Sub
		return true
	case OpSub32F:
		v.Op = OpWasm32F32Sub
		return true
	case OpSub64:
		return rewriteValueWasm32_OpSub64(v)
	case OpSub64F:
		v.Op = OpWasm32F64Sub
		return true
	case OpSub8:
		v.Op = OpWasm32I32Sub
		return true
	case OpSubPtr:
		v.Op = OpWasm32I32Sub
		return true
	case OpTailCall:
		v.Op = OpWasm32LoweredTailCall
		return true
	case OpTrunc:
		v.Op = OpWasm32F64Trunc
		return true
	case OpTrunc16to8:
		v.Op = OpCopy
		return true
	case OpTrunc32to16:
		v.Op = OpCopy
		return true
	case OpTrunc32to8:
		v.Op = OpCopy
		return true
	case OpTrunc64to16:
		return rewriteValueWasm32_OpTrunc64to16(v)
	case OpTrunc64to32:
		return rewriteValueWasm32_OpTrunc64to32(v)
	case OpTrunc64to8:
		return rewriteValueWasm32_OpTrunc64to8(v)
	case OpWB:
		v.Op = OpWasm32LoweredWB
		return true
	case OpWasm32F32Add:
		return rewriteValueWasm32_OpWasm32F32Add(v)
	case OpWasm32F32Mul:
		return rewriteValueWasm32_OpWasm32F32Mul(v)
	case OpWasm32F64Add:
		return rewriteValueWasm32_OpWasm32F64Add(v)
	case OpWasm32F64Mul:
		return rewriteValueWasm32_OpWasm32F64Mul(v)
	case OpWasm32I32Add:
		return rewriteValueWasm32_OpWasm32I32Add(v)
	case OpWasm32I32AddConst:
		return rewriteValueWasm32_OpWasm32I32AddConst(v)
	case OpWasm32I32And:
		return rewriteValueWasm32_OpWasm32I32And(v)
	case OpWasm32I32Eq:
		return rewriteValueWasm32_OpWasm32I32Eq(v)
	case OpWasm32I32Eqz:
		return rewriteValueWasm32_OpWasm32I32Eqz(v)
	case OpWasm32I32LeU:
		return rewriteValueWasm32_OpWasm32I32LeU(v)
	case OpWasm32I32Load:
		return rewriteValueWasm32_OpWasm32I32Load(v)
	case OpWasm32I32Load16S:
		return rewriteValueWasm32_OpWasm32I32Load16S(v)
	case OpWasm32I32Load16U:
		return rewriteValueWasm32_OpWasm32I32Load16U(v)
	case OpWasm32I32Load8S:
		return rewriteValueWasm32_OpWasm32I32Load8S(v)
	case OpWasm32I32Load8U:
		return rewriteValueWasm32_OpWasm32I32Load8U(v)
	case OpWasm32I32LtU:
		return rewriteValueWasm32_OpWasm32I32LtU(v)
	case OpWasm32I32Mul:
		return rewriteValueWasm32_OpWasm32I32Mul(v)
	case OpWasm32I32Ne:
		return rewriteValueWasm32_OpWasm32I32Ne(v)
	case OpWasm32I32Or:
		return rewriteValueWasm32_OpWasm32I32Or(v)
	case OpWasm32I32Shl:
		return rewriteValueWasm32_OpWasm32I32Shl(v)
	case OpWasm32I32ShrS:
		return rewriteValueWasm32_OpWasm32I32ShrS(v)
	case OpWasm32I32ShrU:
		return rewriteValueWasm32_OpWasm32I32ShrU(v)
	case OpWasm32I32Store:
		return rewriteValueWasm32_OpWasm32I32Store(v)
	case OpWasm32I32Store16:
		return rewriteValueWasm32_OpWasm32I32Store16(v)
	case OpWasm32I32Store8:
		return rewriteValueWasm32_OpWasm32I32Store8(v)
	case OpWasm32I32Xor:
		return rewriteValueWasm32_OpWasm32I32Xor(v)
	case OpWasm32I64Load:
		return rewriteValueWasm32_OpWasm32I64Load(v)
	case OpWasm32I64Load16S:
		return rewriteValueWasm32_OpWasm32I64Load16S(v)
	case OpWasm32I64Load16U:
		return rewriteValueWasm32_OpWasm32I64Load16U(v)
	case OpWasm32I64Load32S:
		return rewriteValueWasm32_OpWasm32I64Load32S(v)
	case OpWasm32I64Load32U:
		return rewriteValueWasm32_OpWasm32I64Load32U(v)
	case OpWasm32I64Load8S:
		return rewriteValueWasm32_OpWasm32I64Load8S(v)
	case OpWasm32I64Load8U:
		return rewriteValueWasm32_OpWasm32I64Load8U(v)
	case OpWasm32I64Store:
		return rewriteValueWasm32_OpWasm32I64Store(v)
	case OpWasm32I64Store16:
		return rewriteValueWasm32_OpWasm32I64Store16(v)
	case OpWasm32I64Store32:
		return rewriteValueWasm32_OpWasm32I64Store32(v)
	case OpWasm32I64Store8:
		return rewriteValueWasm32_OpWasm32I64Store8(v)
	case OpXor16:
		v.Op = OpWasm32I32Xor
		return true
	case OpXor32:
		v.Op = OpWasm32I32Xor
		return true
	case OpXor64:
		return rewriteValueWasm32_OpXor64(v)
	case OpXor8:
		v.Op = OpWasm32I32Xor
		return true
	case OpZero:
		return rewriteValueWasm32_OpZero(v)
	case OpZeroExt16to32:
		return rewriteValueWasm32_OpZeroExt16to32(v)
	case OpZeroExt16to64:
		return rewriteValueWasm32_OpZeroExt16to64(v)
	case OpZeroExt32to64:
		return rewriteValueWasm32_OpZeroExt32to64(v)
	case OpZeroExt8to16:
		return rewriteValueWasm32_OpZeroExt8to16(v)
	case OpZeroExt8to32:
		return rewriteValueWasm32_OpZeroExt8to32(v)
	case OpZeroExt8to64:
		return rewriteValueWasm32_OpZeroExt8to64(v)
	case OpZeromask:
		return rewriteValueWasm32_OpZeromask(v)
	}
	return false
}
func rewriteValueWasm32_OpAdd64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Add64 x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Add64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Add64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Add64Decomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpAddr(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Addr {sym} base)
	// result: (LoweredAddr {sym} [0] base)
	for {
		sym := auxToSym(v.Aux)
		base := v_0
		v.reset(OpWasm32LoweredAddr)
		v.AuxInt = int32ToAuxInt(0)
		v.Aux = symToAux(sym)
		v.AddArg(base)
		return true
	}
}
func rewriteValueWasm32_OpAnd64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (And64 x y)
	// result: (Int64Make (And32 <typ.UInt32> (Int64Hi x) (Int64Hi y)) (And32 <typ.UInt32> (Int64Lo x) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpArg(v *Value) bool {
	b := v.Block
	config := b.Func.Config
	typ := &b.Func.Config.Types
	// match: (Arg {n} [off])
	// cond: is64BitInt(v.Type) && !config.BigEndian && v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")
	// result: (Int64Make (Arg <typ.Int32> {n} [off+4]) (Arg <typ.UInt32> {n} [off]))
	for {
		off := auxIntToInt32(v.AuxInt)
		n := auxToSym(v.Aux)
		if !(is64BitInt(v.Type) && !config.BigEndian && v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpArg, typ.Int32)
		v0.AuxInt = int32ToAuxInt(off + 4)
		v0.Aux = symToAux(n)
		v1 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(off)
		v1.Aux = symToAux(n)
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: is64BitInt(v.Type) && !config.BigEndian && !v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")
	// result: (Int64Make (Arg <typ.UInt32> {n} [off+4]) (Arg <typ.UInt32> {n} [off]))
	for {
		off := auxIntToInt32(v.AuxInt)
		n := auxToSym(v.Aux)
		if !(is64BitInt(v.Type) && !config.BigEndian && !v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(off + 4)
		v0.Aux = symToAux(n)
		v1 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(off)
		v1.Aux = symToAux(n)
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: is64BitInt(v.Type) && config.BigEndian && v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")
	// result: (Int64Make (Arg <typ.Int32> {n} [off]) (Arg <typ.UInt32> {n} [off+4]))
	for {
		off := auxIntToInt32(v.AuxInt)
		n := auxToSym(v.Aux)
		if !(is64BitInt(v.Type) && config.BigEndian && v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpArg, typ.Int32)
		v0.AuxInt = int32ToAuxInt(off)
		v0.Aux = symToAux(n)
		v1 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(off + 4)
		v1.Aux = symToAux(n)
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Arg {n} [off])
	// cond: is64BitInt(v.Type) && config.BigEndian && !v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")
	// result: (Int64Make (Arg <typ.UInt32> {n} [off]) (Arg <typ.UInt32> {n} [off+4]))
	for {
		off := auxIntToInt32(v.AuxInt)
		n := auxToSym(v.Aux)
		if !(is64BitInt(v.Type) && config.BigEndian && !v.Type.IsSigned() && !(b.Func.pass.name == "decompose builtin")) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(off)
		v0.Aux = symToAux(n)
		v1 := b.NewValue0(v.Pos, OpArg, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(off + 4)
		v1.Aux = symToAux(n)
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpBitLen32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (BitLen32 x)
	// result: (I32Sub (I32Const [64]) (I32Clz x))
	for {
		x := v_0
		v.reset(OpWasm32I32Sub)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(64)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Clz, typ.Int32)
		v1.AddArg(x)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpBitLen64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (BitLen64 x)
	// result: (Add32 <typ.Int> (BitLen32 <typ.Int> (Int64Hi x)) (BitLen32 <typ.Int> (Or32 <typ.UInt32> (Int64Lo x) (Zeromask (Int64Hi x)))))
	for {
		x := v_0
		v.reset(OpAdd32)
		v.Type = typ.Int
		v0 := b.NewValue0(v.Pos, OpBitLen32, typ.Int)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpBitLen32, typ.Int)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v5.AddArg(v1)
		v3.AddArg2(v4, v5)
		v2.AddArg(v3)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpBswap64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Bswap64 x)
	// result: (Int64Make (Bswap32 <typ.UInt32> (Int64Lo x)) (Bswap32 <typ.UInt32> (Int64Hi x)))
	for {
		x := v_0
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpBswap32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpBswap32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v3.AddArg(x)
		v2.AddArg(v3)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpCom16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com16 x)
	// result: (I32Xor x (I32Const [-1]))
	for {
		x := v_0
		v.reset(OpWasm32I32Xor)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpCom32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com32 x)
	// result: (I32Xor x (I32Const [-1]))
	for {
		x := v_0
		v.reset(OpWasm32I32Xor)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpCom64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com64 x)
	// result: (Int64Make (Com32 <typ.UInt32> (Int64Hi x)) (Com32 <typ.UInt32> (Int64Lo x)))
	for {
		x := v_0
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpCom32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpCom32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v2.AddArg(v3)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpCom8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Com8 x)
	// result: (I32Xor x (I32Const [-1]))
	for {
		x := v_0
		v.reset(OpWasm32I32Xor)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(-1)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpConst16(v *Value) bool {
	// match: (Const16 [c])
	// result: (I32Const [int32(c)])
	for {
		c := auxIntToInt16(v.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(c))
		return true
	}
}
func rewriteValueWasm32_OpConst32(v *Value) bool {
	// match: (Const32 [c])
	// result: (I32Const [int32(c)])
	for {
		c := auxIntToInt32(v.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(c))
		return true
	}
}
func rewriteValueWasm32_OpConst64(v *Value) bool {
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Const64 <t> [c])
	// cond: t.IsSigned()
	// result: (Int64Make (Const32 <typ.Int32> [int32(c>>32)]) (Const32 <typ.UInt32> [int32(c)]))
	for {
		t := v.Type
		c := auxIntToInt64(v.AuxInt)
		if !(t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpConst32, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c >> 32))
		v1 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Const64 <t> [c])
	// cond: !t.IsSigned()
	// result: (Int64Make (Const32 <typ.UInt32> [int32(c>>32)]) (Const32 <typ.UInt32> [int32(c)]))
	for {
		t := v.Type
		c := auxIntToInt64(v.AuxInt)
		if !(!t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(int32(c >> 32))
		v1 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v1.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpConst8(v *Value) bool {
	// match: (Const8 [c])
	// result: (I32Const [int32(c)])
	for {
		c := auxIntToInt8(v.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(c))
		return true
	}
}
func rewriteValueWasm32_OpConstBool(v *Value) bool {
	// match: (ConstBool [c])
	// result: (I32Const [b2i32(c)])
	for {
		c := auxIntToBool(v.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(b2i32(c))
		return true
	}
}
func rewriteValueWasm32_OpConstNil(v *Value) bool {
	// match: (ConstNil)
	// result: (I32Const [0])
	for {
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
}
func rewriteValueWasm32_OpCtz16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Ctz16 x)
	// result: (I32Ctz (I32Or x (I32Const [0x10000])))
	for {
		x := v_0
		v.reset(OpWasm32I32Ctz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Or, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(0x10000)
		v0.AddArg2(x, v1)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpCtz64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Ctz64 x)
	// result: (Add32 <typ.UInt32> (Ctz32 <typ.UInt32> (Int64Lo x)) (And32 <typ.UInt32> (Com32 <typ.UInt32> (Zeromask (Int64Lo x))) (Ctz32 <typ.UInt32> (Int64Hi x))))
	for {
		x := v_0
		v.reset(OpAdd32)
		v.Type = typ.UInt32
		v0 := b.NewValue0(v.Pos, OpCtz32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg(v1)
		v2 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpCom32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v4.AddArg(v1)
		v3.AddArg(v4)
		v5 := b.NewValue0(v.Pos, OpCtz32, typ.UInt32)
		v6 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v6.AddArg(x)
		v5.AddArg(v6)
		v2.AddArg2(v3, v5)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpCtz8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Ctz8 x)
	// result: (I32Ctz (I32Or x (I32Const [0x100])))
	for {
		x := v_0
		v.reset(OpWasm32I32Ctz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Or, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(0x100)
		v0.AddArg2(x, v1)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpCvt32Uto64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32Uto64F x)
	// result: (F64ConvertI32U x)
	for {
		x := v_0
		v.reset(OpWasm32F64ConvertI32U)
		v.AddArg(x)
		return true
	}
}
func rewriteValueWasm32_OpCvt32to64F(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Cvt32to64F x)
	// result: (F64ConvertI32S x)
	for {
		x := v_0
		v.reset(OpWasm32F64ConvertI32S)
		v.AddArg(x)
		return true
	}
}
func rewriteValueWasm32_OpDiv16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div16 [false] x y)
	// result: (I32DivS (SignExt16to32 x) (SignExt16to32 y))
	for {
		if auxIntToBool(v.AuxInt) != false {
			break
		}
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivS)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpDiv16u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div16u x y)
	// result: (I32DivU (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivU)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpDiv32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Div32 [false] x y)
	// result: (I32DivS x y)
	for {
		if auxIntToBool(v.AuxInt) != false {
			break
		}
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivS)
		v.AddArg2(x, y)
		return true
	}
	return false
}
func rewriteValueWasm32_OpDiv32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Div32u x y)
	// result: (I32DivU x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivU)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpDiv64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div64 x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Div64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Div64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Div64Decomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpDiv64u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div64u x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Div64uDecomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Div64uDecomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Div64uDecomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpDiv8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div8 x y)
	// result: (I32DivS (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivS)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpDiv8u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Div8u x y)
	// result: (I32DivU (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32DivU)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpEq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Eq16 x y)
	// result: (I32Eq (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Eq)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpEq64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Eq64 x y)
	// result: (AndB (Eq32 (Int64Hi x) (Int64Hi y)) (Eq32 (Int64Lo x) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpAndB)
		v0 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpEq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Eq8 x y)
	// result: (I32Eq (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Eq)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpInt64Hi(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Int64Hi (Int64Make hi _))
	// result: hi
	for {
		if v_0.Op != OpInt64Make {
			break
		}
		hi := v_0.Args[0]
		v.copyOf(hi)
		return true
	}
	return false
}
func rewriteValueWasm32_OpInt64Lo(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Int64Lo (Int64Make _ lo))
	// result: lo
	for {
		if v_0.Op != OpInt64Make {
			break
		}
		lo := v_0.Args[1]
		v.copyOf(lo)
		return true
	}
	return false
}
func rewriteValueWasm32_OpIsNonNil(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (IsNonNil p)
	// result: (I32Eqz (I32Eqz p))
	for {
		p := v_0
		v.reset(OpWasm32I32Eqz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Eqz, typ.Bool)
		v0.AddArg(p)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpLeq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq16 x y)
	// result: (I32LeS (SignExt16to32 x) (SignExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LeS)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLeq16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq16U x y)
	// result: (I32LeU (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LeU)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLeq64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq64 x y)
	// result: (OrB (Less32 (Int64Hi x) (Int64Hi y)) (AndB (Eq32 (Int64Hi x) (Int64Hi y)) (Leq32U (Int64Lo x) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOrB)
		v0 := b.NewValue0(v.Pos, OpLess32, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpAndB, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v4.AddArg2(v1, v2)
		v5 := b.NewValue0(v.Pos, OpLeq32U, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v6.AddArg(x)
		v7 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v7.AddArg(y)
		v5.AddArg2(v6, v7)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpLeq64U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq64U x y)
	// result: (OrB (Less32U (Int64Hi x) (Int64Hi y)) (AndB (Eq32 (Int64Hi x) (Int64Hi y)) (Leq32U (Int64Lo x) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOrB)
		v0 := b.NewValue0(v.Pos, OpLess32U, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpAndB, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v4.AddArg2(v1, v2)
		v5 := b.NewValue0(v.Pos, OpLeq32U, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v6.AddArg(x)
		v7 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v7.AddArg(y)
		v5.AddArg2(v6, v7)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpLeq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq8 x y)
	// result: (I32LeS (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LeS)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLeq8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Leq8U x y)
	// result: (I32LeU (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LeU)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLess16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less16 x y)
	// result: (I32LtS (SignExt16to32 x) (SignExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LtS)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLess16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less16U x y)
	// result: (I32LtU (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LtU)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLess64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less64 x y)
	// result: (OrB (Less32 (Int64Hi x) (Int64Hi y)) (AndB (Eq32 (Int64Hi x) (Int64Hi y)) (Less32U (Int64Lo x) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOrB)
		v0 := b.NewValue0(v.Pos, OpLess32, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpAndB, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v4.AddArg2(v1, v2)
		v5 := b.NewValue0(v.Pos, OpLess32U, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v6.AddArg(x)
		v7 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v7.AddArg(y)
		v5.AddArg2(v6, v7)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpLess64U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less64U x y)
	// result: (OrB (Less32U (Int64Hi x) (Int64Hi y)) (AndB (Eq32 (Int64Hi x) (Int64Hi y)) (Less32U (Int64Lo x) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpOrB)
		v0 := b.NewValue0(v.Pos, OpLess32U, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpAndB, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpEq32, typ.Bool)
		v4.AddArg2(v1, v2)
		v5 := b.NewValue0(v.Pos, OpLess32U, typ.Bool)
		v6 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v6.AddArg(x)
		v7 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v7.AddArg(y)
		v5.AddArg2(v6, v7)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpLess8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less8 x y)
	// result: (I32LtS (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LtS)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLess8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Less8U x y)
	// result: (I32LtU (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32LtU)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpLoad(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	config := b.Func.Config
	typ := &b.Func.Config.Types
	// match: (Load <t> ptr mem)
	// cond: is32BitFloat(t)
	// result: (F32Load ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is32BitFloat(t)) {
			break
		}
		v.reset(OpWasm32F32Load)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitFloat(t)
	// result: (F64Load ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitFloat(t)) {
			break
		}
		v.reset(OpWasm32F64Load)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: t.Size() == 4
	// result: (I32Load ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.Size() == 4) {
			break
		}
		v.reset(OpWasm32I32Load)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: t.Size() == 2 && !t.IsSigned()
	// result: (I32Load16U ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.Size() == 2 && !t.IsSigned()) {
			break
		}
		v.reset(OpWasm32I32Load16U)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: t.Size() == 2 && t.IsSigned()
	// result: (I32Load16S ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.Size() == 2 && t.IsSigned()) {
			break
		}
		v.reset(OpWasm32I32Load16S)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: t.Size() == 1 && !t.IsSigned()
	// result: (I32Load8U ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.Size() == 1 && !t.IsSigned()) {
			break
		}
		v.reset(OpWasm32I32Load8U)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: t.Size() == 1 && t.IsSigned()
	// result: (I32Load8S ptr mem)
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(t.Size() == 1 && t.IsSigned()) {
			break
		}
		v.reset(OpWasm32I32Load8S)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitInt(t) && !config.BigEndian && t.IsSigned()
	// result: (Int64Make (Load <typ.Int32> (OffPtr <typ.Int32Ptr> [4] ptr) mem) (Load <typ.UInt32> ptr mem))
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitInt(t) && !config.BigEndian && t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpLoad, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpOffPtr, typ.Int32Ptr)
		v1.AuxInt = int64ToAuxInt(4)
		v1.AddArg(ptr)
		v0.AddArg2(v1, mem)
		v2 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v2.AddArg2(ptr, mem)
		v.AddArg2(v0, v2)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitInt(t) && !config.BigEndian && !t.IsSigned()
	// result: (Int64Make (Load <typ.UInt32> (OffPtr <typ.UInt32Ptr> [4] ptr) mem) (Load <typ.UInt32> ptr mem))
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitInt(t) && !config.BigEndian && !t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpOffPtr, typ.UInt32Ptr)
		v1.AuxInt = int64ToAuxInt(4)
		v1.AddArg(ptr)
		v0.AddArg2(v1, mem)
		v2 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v2.AddArg2(ptr, mem)
		v.AddArg2(v0, v2)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitInt(t) && config.BigEndian && t.IsSigned()
	// result: (Int64Make (Load <typ.Int32> ptr mem) (Load <typ.UInt32> (OffPtr <typ.UInt32Ptr> [4] ptr) mem))
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitInt(t) && config.BigEndian && t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpLoad, typ.Int32)
		v0.AddArg2(ptr, mem)
		v1 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpOffPtr, typ.UInt32Ptr)
		v2.AuxInt = int64ToAuxInt(4)
		v2.AddArg(ptr)
		v1.AddArg2(v2, mem)
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Load <t> ptr mem)
	// cond: is64BitInt(t) && config.BigEndian && !t.IsSigned()
	// result: (Int64Make (Load <typ.UInt32> ptr mem) (Load <typ.UInt32> (OffPtr <typ.UInt32Ptr> [4] ptr) mem))
	for {
		t := v.Type
		ptr := v_0
		mem := v_1
		if !(is64BitInt(t) && config.BigEndian && !t.IsSigned()) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v0.AddArg2(ptr, mem)
		v1 := b.NewValue0(v.Pos, OpLoad, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpOffPtr, typ.UInt32Ptr)
		v2.AuxInt = int64ToAuxInt(4)
		v2.AddArg(ptr)
		v1.AddArg2(v2, mem)
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpLocalAddr(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (LocalAddr <t> {sym} base mem)
	// cond: t.Elem().HasPointers()
	// result: (LoweredAddr {sym} (SPanchored base mem))
	for {
		t := v.Type
		sym := auxToSym(v.Aux)
		base := v_0
		mem := v_1
		if !(t.Elem().HasPointers()) {
			break
		}
		v.reset(OpWasm32LoweredAddr)
		v.Aux = symToAux(sym)
		v0 := b.NewValue0(v.Pos, OpSPanchored, typ.Uintptr)
		v0.AddArg2(base, mem)
		v.AddArg(v0)
		return true
	}
	// match: (LocalAddr <t> {sym} base _)
	// cond: !t.Elem().HasPointers()
	// result: (LoweredAddr {sym} base)
	for {
		t := v.Type
		sym := auxToSym(v.Aux)
		base := v_0
		if !(!t.Elem().HasPointers()) {
			break
		}
		v.reset(OpWasm32LoweredAddr)
		v.Aux = symToAux(sym)
		v.AddArg(base)
		return true
	}
	return false
}
func rewriteValueWasm32_OpLsh16x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x16 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh16x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (I32Shl x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpWasm32I32Shl)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh16x64 _ (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (Const16 [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpConst16)
		v.AuxInt = int16ToAuxInt(0)
		return true
	}
	// match: (Lsh16x64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh16x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Lsh16x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpLsh16x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Lsh16x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Lsh16x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpLsh16x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh16x64 x y)
	// result: (Lsh16x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh16x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh16x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh16x8 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh32x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x16 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x32 x y)
	// cond: shiftIsBounded(v)
	// result: (I32Shl x y)
	for {
		x := v_0
		y := v_1
		if !(shiftIsBounded(v)) {
			break
		}
		v.reset(OpWasm32I32Shl)
		v.AddArg2(x, y)
		return true
	}
	// match: (Lsh32x32 x (I32Const [c]))
	// cond: uint32(c) < 32
	// result: (I32Shl x (I32Const [c]))
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) < 32) {
			break
		}
		v.reset(OpWasm32I32Shl)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh32x32 x (I32Const [c]))
	// cond: uint32(c) >= 32
	// result: (I32Const [0])
	for {
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) >= 32) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh32x32 x y)
	// result: (Select (I32Shl x y) (I32Const [0]) (I32LtU y (I32Const [32])))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32Select)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Shl, typ.Int32)
		v0.AddArg2(x, y)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(0)
		v2 := b.NewValue0(v.Pos, OpWasm32I32LtU, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(32)
		v2.AddArg2(y, v3)
		v.AddArg3(v0, v1, v2)
		return true
	}
}
func rewriteValueWasm32_OpLsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (I32Shl x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpWasm32I32Shl)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh32x64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (Const32 [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh32x64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh32x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Lsh32x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Lsh32x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Lsh32x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpLsh32x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh32x64 x y)
	// result: (Lsh32x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh32x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh32x8 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh64x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh64x16 x s)
	// result: (Int64Make (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Lsh32x16 <typ.UInt32> (Int64Hi x) s) (Rsh32Ux16 <typ.UInt32> (Int64Lo x) (Sub16 <typ.UInt16> (Const16 <typ.UInt16> [32]) s))) (Lsh32x16 <typ.UInt32> (Int64Lo x) (Sub16 <typ.UInt16> s (Const16 <typ.UInt16> [32])))) (Lsh32x16 <typ.UInt32> (Int64Lo x) s))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpLsh32x16, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v3.AddArg(x)
		v2.AddArg2(v3, s)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux16, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v6 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v7 := b.NewValue0(v.Pos, OpConst16, typ.UInt16)
		v7.AuxInt = int16ToAuxInt(32)
		v6.AddArg2(v7, s)
		v4.AddArg2(v5, v6)
		v1.AddArg2(v2, v4)
		v8 := b.NewValue0(v.Pos, OpLsh32x16, typ.UInt32)
		v9 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v9.AddArg2(s, v7)
		v8.AddArg2(v5, v9)
		v0.AddArg2(v1, v8)
		v10 := b.NewValue0(v.Pos, OpLsh32x16, typ.UInt32)
		v10.AddArg2(v5, s)
		v.AddArg2(v0, v10)
		return true
	}
}
func rewriteValueWasm32_OpLsh64x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh64x32 x s)
	// result: (Int64Make (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Lsh32x32 <typ.UInt32> (Int64Hi x) s) (Rsh32Ux32 <typ.UInt32> (Int64Lo x) (Sub32 <typ.UInt32> (Const32 <typ.UInt32> [32]) s))) (Lsh32x32 <typ.UInt32> (Int64Lo x) (Sub32 <typ.UInt32> s (Const32 <typ.UInt32> [32])))) (Lsh32x32 <typ.UInt32> (Int64Lo x) s))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpLsh32x32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v3.AddArg(x)
		v2.AddArg2(v3, s)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v6 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v7.AuxInt = int32ToAuxInt(32)
		v6.AddArg2(v7, s)
		v4.AddArg2(v5, v6)
		v1.AddArg2(v2, v4)
		v8 := b.NewValue0(v.Pos, OpLsh32x32, typ.UInt32)
		v9 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v9.AddArg2(s, v7)
		v8.AddArg2(v5, v9)
		v0.AddArg2(v1, v8)
		v10 := b.NewValue0(v.Pos, OpLsh32x32, typ.UInt32)
		v10.AddArg2(v5, s)
		v.AddArg2(v0, v10)
		return true
	}
}
func rewriteValueWasm32_OpLsh64x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh64x64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const64 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst64)
		v.AuxInt = int64ToAuxInt(0)
		return true
	}
	// match: (Lsh64x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Lsh64x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpLsh64x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Lsh64x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Lsh64x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpLsh64x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh64x64 x y)
	// result: (Lsh64x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh64x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh64x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh64x8 x s)
	// result: (Int64Make (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Lsh32x8 <typ.UInt32> (Int64Hi x) s) (Rsh32Ux8 <typ.UInt32> (Int64Lo x) (Sub8 <typ.UInt8> (Const8 <typ.UInt8> [32]) s))) (Lsh32x8 <typ.UInt32> (Int64Lo x) (Sub8 <typ.UInt8> s (Const8 <typ.UInt8> [32])))) (Lsh32x8 <typ.UInt32> (Int64Lo x) s))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpLsh32x8, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v3.AddArg(x)
		v2.AddArg2(v3, s)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux8, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v6 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v7 := b.NewValue0(v.Pos, OpConst8, typ.UInt8)
		v7.AuxInt = int8ToAuxInt(32)
		v6.AddArg2(v7, s)
		v4.AddArg2(v5, v6)
		v1.AddArg2(v2, v4)
		v8 := b.NewValue0(v.Pos, OpLsh32x8, typ.UInt32)
		v9 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v9.AddArg2(s, v7)
		v8.AddArg2(v5, v9)
		v0.AddArg2(v1, v8)
		v10 := b.NewValue0(v.Pos, OpLsh32x8, typ.UInt32)
		v10.AddArg2(v5, s)
		v.AddArg2(v0, v10)
		return true
	}
}
func rewriteValueWasm32_OpLsh8x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x16 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh8x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x64 x (Const64 [c]))
	// cond: uint64(c) < 8
	// result: (I32Shl x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 8) {
			break
		}
		v.reset(OpWasm32I32Shl)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh8x64 _ (Const64 [c]))
	// cond: uint64(c) >= 8
	// result: (Const16 [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 8) {
			break
		}
		v.reset(OpConst16)
		v.AuxInt = int16ToAuxInt(0)
		return true
	}
	// match: (Lsh8x64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Lsh8x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Lsh8x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpLsh8x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Lsh8x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Lsh8x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpLsh8x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Lsh8x64 x y)
	// result: (Lsh8x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpLsh8x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpLsh8x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Lsh8x8 [c] x y)
	// result: (Lsh32x32 [c] x (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpLsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpMod16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod16 [false] x y)
	// result: (I32RemS (SignExt16to32 x) (SignExt16to32 y))
	for {
		if auxIntToBool(v.AuxInt) != false {
			break
		}
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemS)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpMod16u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod16u x y)
	// result: (I32RemU (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemU)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpMod32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mod32 [false] x y)
	// result: (I32RemS x y)
	for {
		if auxIntToBool(v.AuxInt) != false {
			break
		}
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemS)
		v.AddArg2(x, y)
		return true
	}
	return false
}
func rewriteValueWasm32_OpMod32u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Mod32u x y)
	// result: (I32RemU x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemU)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpMod64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod64 x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Mod64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Mod64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Mod64Decomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpMod64u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod64u x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Mod64uDecomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Mod64uDecomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Mod64uDecomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpMod8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod8 x y)
	// result: (I32RemS (SignExt8to32 x) (SignExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemS)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpMod8u(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mod8u x y)
	// result: (I32RemU (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32RemU)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpMove(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Move [0] _ _ mem)
	// result: mem
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		mem := v_2
		v.copyOf(mem)
		return true
	}
	// match: (Move [1] dst src mem)
	// result: (I32Store8 dst (I32Load8U src mem) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 1 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store8)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load8U, typ.UInt8)
		v0.AddArg2(src, mem)
		v.AddArg3(dst, v0, mem)
		return true
	}
	// match: (Move [2] dst src mem)
	// result: (I32Store16 dst (I32Load16U src mem) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 2 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store16)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load16U, typ.UInt16)
		v0.AddArg2(src, mem)
		v.AddArg3(dst, v0, mem)
		return true
	}
	// match: (Move [4] dst src mem)
	// result: (I32Store dst (I32Load src mem) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 4 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v0.AddArg2(src, mem)
		v.AddArg3(dst, v0, mem)
		return true
	}
	// match: (Move [3] dst src mem)
	// result: (I32Store8 [2] dst (I32Load8U [2] src mem) (I32Store16 dst (I32Load16U src mem) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 3 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store8)
		v.AuxInt = int32ToAuxInt(2)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load8U, typ.UInt8)
		v0.AuxInt = int32ToAuxInt(2)
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store16, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load16U, typ.UInt16)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [5] dst src mem)
	// result: (I32Store8 [4] dst (I32Load8U [4] src mem) (I32Store dst (I32Load src mem) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 5 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store8)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load8U, typ.UInt8)
		v0.AuxInt = int32ToAuxInt(4)
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [6] dst src mem)
	// result: (I32Store16 [4] dst (I32Load16U [4] src mem) (I32Store dst (I32Load src mem) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 6 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store16)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load16U, typ.UInt16)
		v0.AuxInt = int32ToAuxInt(4)
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [7] dst src mem)
	// result: (I32Store [3] dst (I32Load [3] src mem) (I32Store dst (I32Load src mem) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 7 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(3)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(3)
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [8] dst src mem)
	// result: (I32Store [4] dst (I32Load [4] src mem) (I32Store dst (I32Load src mem) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 8 {
			break
		}
		dst := v_0
		src := v_1
		mem := v_2
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(4)
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [s] dst src mem)
	// cond: s > 4 && s < 8
	// result: (I32Store [int32(s-4)] dst (I32Load [int32(s-4)] src mem) (I32Store dst (I32Load src mem) mem))
	for {
		s := auxIntToInt64(v.AuxInt)
		dst := v_0
		src := v_1
		mem := v_2
		if !(s > 4 && s < 8) {
			break
		}
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(int32(s - 4))
		v0 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(int32(s - 4))
		v0.AddArg2(src, mem)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Load, typ.UInt32)
		v2.AddArg2(src, mem)
		v1.AddArg3(dst, v2, mem)
		v.AddArg3(dst, v0, v1)
		return true
	}
	// match: (Move [s] dst src mem)
	// cond: logLargeCopy(v, s)
	// result: (LoweredMove [s] dst src mem)
	for {
		s := auxIntToInt64(v.AuxInt)
		dst := v_0
		src := v_1
		mem := v_2
		if !(logLargeCopy(v, s)) {
			break
		}
		v.reset(OpWasm32LoweredMove)
		v.AuxInt = int64ToAuxInt(s)
		v.AddArg3(dst, src, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpMul64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Mul64 x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Mul64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Mul64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Mul64Decomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpNeg16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neg16 x)
	// result: (I32Sub (I32Const [0]) x)
	for {
		x := v_0
		v.reset(OpWasm32I32Sub)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpNeg32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neg32 x)
	// result: (I32Sub (I32Const [0]) x)
	for {
		x := v_0
		v.reset(OpWasm32I32Sub)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpNeg64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	// match: (Neg64 <t> x)
	// result: (Sub64 (Const64 <t> [0]) x)
	for {
		t := v.Type
		x := v_0
		v.reset(OpSub64)
		v0 := b.NewValue0(v.Pos, OpConst64, t)
		v0.AuxInt = int64ToAuxInt(0)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpNeg8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neg8 x)
	// result: (I32Sub (I32Const [0]) x)
	for {
		x := v_0
		v.reset(OpWasm32I32Sub)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpNeq16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neq16 x y)
	// result: (I32Ne (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Ne)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpNeq32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Neq32 x y)
	// result: (I32Ne x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Ne)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpNeq64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neq64 x y)
	// result: (OrB (Neq32 (Int64Hi x) (Int64Hi y)) (Neq32 (Int64Lo x) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpOrB)
		v0 := b.NewValue0(v.Pos, OpNeq32, typ.Bool)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpNeq32, typ.Bool)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpNeq8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Neq8 x y)
	// result: (I32Ne (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Ne)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpOffPtr(v *Value) bool {
	v_0 := v.Args[0]
	// match: (OffPtr [off] ptr)
	// result: (I32AddConst [int32(off)] ptr)
	for {
		off := auxIntToInt64(v.AuxInt)
		ptr := v_0
		v.reset(OpWasm32I32AddConst)
		v.AuxInt = int32ToAuxInt(int32(off))
		v.AddArg(ptr)
		return true
	}
}
func rewriteValueWasm32_OpOr16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or16 x y)
	// result: (I32Or x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Or)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpOr32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or32 x y)
	// result: (I32Or x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Or)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpOr64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Or64 x y)
	// result: (Int64Make (Or32 <typ.UInt32> (Int64Hi x) (Int64Hi y)) (Or32 <typ.UInt32> (Int64Lo x) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpOr8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (Or8 x y)
	// result: (I32Or x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Or)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpOrB(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (OrB x y)
	// result: (I32Or x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Or)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpPanicExtend(v *Value) bool {
	v_3 := v.Args[3]
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (PanicExtend [kind] hi lo y mem)
	// cond: boundsABI(kind) == 0
	// result: (LoweredPanicExtendA [kind] hi lo y mem)
	for {
		kind := auxIntToInt64(v.AuxInt)
		hi := v_0
		lo := v_1
		y := v_2
		mem := v_3
		if !(boundsABI(kind) == 0) {
			break
		}
		v.reset(OpWasm32LoweredPanicExtendA)
		v.AuxInt = int64ToAuxInt(kind)
		v.AddArg4(hi, lo, y, mem)
		return true
	}
	// match: (PanicExtend [kind] hi lo y mem)
	// cond: boundsABI(kind) == 1
	// result: (LoweredPanicExtendB [kind] hi lo y mem)
	for {
		kind := auxIntToInt64(v.AuxInt)
		hi := v_0
		lo := v_1
		y := v_2
		mem := v_3
		if !(boundsABI(kind) == 1) {
			break
		}
		v.reset(OpWasm32LoweredPanicExtendB)
		v.AuxInt = int64ToAuxInt(kind)
		v.AddArg4(hi, lo, y, mem)
		return true
	}
	// match: (PanicExtend [kind] hi lo y mem)
	// cond: boundsABI(kind) == 2
	// result: (LoweredPanicExtendC [kind] hi lo y mem)
	for {
		kind := auxIntToInt64(v.AuxInt)
		hi := v_0
		lo := v_1
		y := v_2
		mem := v_3
		if !(boundsABI(kind) == 2) {
			break
		}
		v.reset(OpWasm32LoweredPanicExtendC)
		v.AuxInt = int64ToAuxInt(kind)
		v.AddArg4(hi, lo, y, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpPopCount16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (PopCount16 x)
	// result: (I32Popcnt (ZeroExt16to32 x))
	for {
		x := v_0
		v.reset(OpWasm32I32Popcnt)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpPopCount32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (PopCount32 x)
	// result: (I32Popcnt x)
	for {
		x := v_0
		v.reset(OpWasm32I32Popcnt)
		v.AddArg(x)
		return true
	}
}
func rewriteValueWasm32_OpPopCount8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (PopCount8 x)
	// result: (I32Popcnt (ZeroExt8to32 x))
	for {
		x := v_0
		v.reset(OpWasm32I32Popcnt)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpRotateLeft16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (RotateLeft16 <t> x (I32Const [c]))
	// result: (Or16 (Lsh16x32 <t> x (I32Const [c&15])) (Rsh16Ux32 <t> x (I32Const [-c&15])))
	for {
		t := v.Type
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		v.reset(OpOr16)
		v0 := b.NewValue0(v.Pos, OpLsh16x32, t)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(c & 15)
		v0.AddArg2(x, v1)
		v2 := b.NewValue0(v.Pos, OpRsh16Ux32, t)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(-c & 15)
		v2.AddArg2(x, v3)
		v.AddArg2(v0, v2)
		return true
	}
	// match: (RotateLeft16 x (Int64Make hi lo))
	// result: (RotateLeft16 x lo)
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v.reset(OpRotateLeft16)
		v.AddArg2(x, lo)
		return true
	}
	return false
}
func rewriteValueWasm32_OpRotateLeft32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (RotateLeft32 x (Int64Make hi lo))
	// result: (RotateLeft32 x lo)
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v.reset(OpRotateLeft32)
		v.AddArg2(x, lo)
		return true
	}
	// match: (RotateLeft32 x y)
	// result: (I32Rotl x y)
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32Rotl)
		v.AddArg2(x, y)
		return true
	}
}
func rewriteValueWasm32_OpRotateLeft64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (RotateLeft64 x (Int64Make hi lo))
	// result: (RotateLeft64 x lo)
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v.reset(OpRotateLeft64)
		v.AddArg2(x, lo)
		return true
	}
	return false
}
func rewriteValueWasm32_OpRotateLeft8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (RotateLeft8 <t> x (I32Const [c]))
	// result: (Or8 (Lsh8x32 <t> x (I32Const [c&7])) (Rsh8Ux32 <t> x (I32Const [-c&7])))
	for {
		t := v.Type
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		v.reset(OpOr8)
		v0 := b.NewValue0(v.Pos, OpLsh8x32, t)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(c & 7)
		v0.AddArg2(x, v1)
		v2 := b.NewValue0(v.Pos, OpRsh8Ux32, t)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(-c & 7)
		v2.AddArg2(x, v3)
		v.AddArg2(v0, v2)
		return true
	}
	// match: (RotateLeft8 x (Int64Make hi lo))
	// result: (RotateLeft8 x lo)
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v.reset(OpRotateLeft8)
		v.AddArg2(x, lo)
		return true
	}
	return false
}
func rewriteValueWasm32_OpRsh16Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux16 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt16to32 x) (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh16Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux32 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt16to32 x) y)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg2(v0, y)
		return true
	}
}
func rewriteValueWasm32_OpRsh16Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (I32ShrU x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpWasm32I32ShrU)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh16Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (I32Const [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh16Ux64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh16Ux64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh16Ux32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh16Ux32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh16Ux64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh16Ux32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh16Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh16Ux64 x y)
	// result: (Rsh16Ux32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh16Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh16Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16Ux8 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt16to32 x) (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh16x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x16 [c] x y)
	// result: (Rsh32x32 [c] (SignExt16to32 x) (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh16x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x32 [c] x y)
	// result: (Rsh32x32 [c] (SignExt16to32 x) y)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg2(v0, y)
		return true
	}
}
func rewriteValueWasm32_OpRsh16x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x64 x (Const64 [c]))
	// cond: uint64(c) < 16
	// result: (I32ShrS x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 16) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh16x64 x (Const64 [c]))
	// cond: uint64(c) >= 16
	// result: (I32ShrS x (I32Const [15]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 16) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(15)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh16x64 x (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Signmask (SignExt16to32 x))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpSignmask)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	// match: (Rsh16x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh16x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh16x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh16x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh16x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh16x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh16x64 x y)
	// result: (Rsh16x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh16x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh16x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh16x8 [c] x y)
	// result: (Rsh32x32 [c] (SignExt16to32 x) (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh32Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux16 [c] x y)
	// result: (Rsh32Ux32 [c] x (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux32 x y)
	// cond: shiftIsBounded(v)
	// result: (I32ShrU x y)
	for {
		x := v_0
		y := v_1
		if !(shiftIsBounded(v)) {
			break
		}
		v.reset(OpWasm32I32ShrU)
		v.AddArg2(x, y)
		return true
	}
	// match: (Rsh32Ux32 x (I32Const [c]))
	// cond: uint32(c) < 32
	// result: (I32ShrU x (I32Const [c]))
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) < 32) {
			break
		}
		v.reset(OpWasm32I32ShrU)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32Ux32 x (I32Const [c]))
	// cond: uint32(c) >= 32
	// result: (I32Const [0])
	for {
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) >= 32) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh32Ux32 x y)
	// result: (Select (I32ShrU x y) (I32Const [0]) (I32LtU y (I32Const [32])))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32Select)
		v0 := b.NewValue0(v.Pos, OpWasm32I32ShrU, typ.Int32)
		v0.AddArg2(x, y)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(0)
		v2 := b.NewValue0(v.Pos, OpWasm32I32LtU, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(32)
		v2.AddArg2(y, v3)
		v.AddArg3(v0, v1, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh32Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (I32ShrU x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpWasm32I32ShrU)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (Const32 [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh32Ux64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh32Ux64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh32Ux32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh32Ux64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh32Ux32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh32Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32Ux64 x y)
	// result: (Rsh32Ux32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32Ux8 [c] x y)
	// result: (Rsh32Ux32 [c] x (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x16 [c] x y)
	// result: (Rsh32x32 [c] x (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x32 x y)
	// cond: shiftIsBounded(v)
	// result: (I32ShrS x y)
	for {
		x := v_0
		y := v_1
		if !(shiftIsBounded(v)) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v.AddArg2(x, y)
		return true
	}
	// match: (Rsh32x32 x (I32Const [c]))
	// cond: uint32(c) < 32
	// result: (I32ShrS x (I32Const [c]))
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) < 32) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(c)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32x32 x (I32Const [c]))
	// cond: uint32(c) >= 32
	// result: (I32ShrS x (I32Const [31]))
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		c := auxIntToInt32(v_1.AuxInt)
		if !(uint32(c) >= 32) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(31)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32x32 x y)
	// result: (I32ShrS x (Select <typ.Int32> y (I32Const [31]) (I32LtU y (I32Const [32]))))
	for {
		x := v_0
		y := v_1
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32Select, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(31)
		v2 := b.NewValue0(v.Pos, OpWasm32I32LtU, typ.Bool)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(32)
		v2.AddArg2(y, v3)
		v0.AddArg3(y, v1, v2)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x64 x (Const64 [c]))
	// cond: uint64(c) < 32
	// result: (I32ShrS x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 32) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32x64 x (Const64 [c]))
	// cond: uint64(c) >= 32
	// result: (I32ShrS x (I32Const [31]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 32) {
			break
		}
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(31)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32x64 x (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Signmask x)
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpSignmask)
		v.AddArg(x)
		return true
	}
	// match: (Rsh32x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh32x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh32x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh32x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh32x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh32x64 x y)
	// result: (Rsh32x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh32x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh32x8 [c] x y)
	// result: (Rsh32x32 [c] x (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(y)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh64Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64Ux16 x s)
	// result: (Int64Make (Rsh32Ux16 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux16 <typ.UInt32> (Int64Lo x) s) (Lsh32x16 <typ.UInt32> (Int64Hi x) (Sub16 <typ.UInt16> (Const16 <typ.UInt16> [32]) s))) (Rsh32Ux16 <typ.UInt32> (Int64Hi x) (Sub16 <typ.UInt16> s (Const16 <typ.UInt16> [32])))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32Ux16, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux16, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x16, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v8 := b.NewValue0(v.Pos, OpConst16, typ.UInt16)
		v8.AuxInt = int16ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpRsh32Ux16, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v10.AddArg2(s, v8)
		v9.AddArg2(v1, v10)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh64Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64Ux32 x s)
	// result: (Int64Make (Rsh32Ux32 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux32 <typ.UInt32> (Int64Lo x) s) (Lsh32x32 <typ.UInt32> (Int64Hi x) (Sub32 <typ.UInt32> (Const32 <typ.UInt32> [32]) s))) (Rsh32Ux32 <typ.UInt32> (Int64Hi x) (Sub32 <typ.UInt32> s (Const32 <typ.UInt32> [32])))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x32, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v8 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v8.AuxInt = int32ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v10.AddArg2(s, v8)
		v9.AddArg2(v1, v10)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh64Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64Ux64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const64 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst64)
		v.AuxInt = int64ToAuxInt(0)
		return true
	}
	// match: (Rsh64Ux64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh64Ux32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh64Ux32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh64Ux64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh64Ux32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh64Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh64Ux64 x y)
	// result: (Rsh64Ux32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh64Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh64Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64Ux8 x s)
	// result: (Int64Make (Rsh32Ux8 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux8 <typ.UInt32> (Int64Lo x) s) (Lsh32x8 <typ.UInt32> (Int64Hi x) (Sub8 <typ.UInt8> (Const8 <typ.UInt8> [32]) s))) (Rsh32Ux8 <typ.UInt32> (Int64Hi x) (Sub8 <typ.UInt8> s (Const8 <typ.UInt8> [32])))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32Ux8, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux8, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x8, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v8 := b.NewValue0(v.Pos, OpConst8, typ.UInt8)
		v8.AuxInt = int8ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpRsh32Ux8, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v10.AddArg2(s, v8)
		v9.AddArg2(v1, v10)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh64x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64x16 x s)
	// result: (Int64Make (Rsh32x16 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux16 <typ.UInt32> (Int64Lo x) s) (Lsh32x16 <typ.UInt32> (Int64Hi x) (Sub16 <typ.UInt16> (Const16 <typ.UInt16> [32]) s))) (And32 <typ.UInt32> (Rsh32x16 <typ.UInt32> (Int64Hi x) (Sub16 <typ.UInt16> s (Const16 <typ.UInt16> [32]))) (Zeromask (ZeroExt16to32 (Rsh16Ux32 <typ.UInt16> s (Const32 <typ.UInt32> [5])))))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32x16, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux16, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x16, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v8 := b.NewValue0(v.Pos, OpConst16, typ.UInt16)
		v8.AuxInt = int16ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpRsh32x16, typ.UInt32)
		v11 := b.NewValue0(v.Pos, OpSub16, typ.UInt16)
		v11.AddArg2(s, v8)
		v10.AddArg2(v1, v11)
		v12 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v13 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v14 := b.NewValue0(v.Pos, OpRsh16Ux32, typ.UInt16)
		v15 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v15.AuxInt = int32ToAuxInt(5)
		v14.AddArg2(s, v15)
		v13.AddArg(v14)
		v12.AddArg(v13)
		v9.AddArg2(v10, v12)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh64x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64x32 x s)
	// result: (Int64Make (Rsh32x32 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux32 <typ.UInt32> (Int64Lo x) s) (Lsh32x32 <typ.UInt32> (Int64Hi x) (Sub32 <typ.UInt32> (Const32 <typ.UInt32> [32]) s))) (And32 <typ.UInt32> (Rsh32x32 <typ.UInt32> (Int64Hi x) (Sub32 <typ.UInt32> s (Const32 <typ.UInt32> [32]))) (Zeromask (Rsh32Ux32 <typ.UInt32> s (Const32 <typ.UInt32> [5]))))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32x32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x32, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v8 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v8.AuxInt = int32ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpRsh32x32, typ.UInt32)
		v11 := b.NewValue0(v.Pos, OpSub32, typ.UInt32)
		v11.AddArg2(s, v8)
		v10.AddArg2(v1, v11)
		v12 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v13 := b.NewValue0(v.Pos, OpRsh32Ux32, typ.UInt32)
		v14 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v14.AuxInt = int32ToAuxInt(5)
		v13.AddArg2(s, v14)
		v12.AddArg(v13)
		v9.AddArg2(v10, v12)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh64x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64x64 x (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Int64Make (Signmask (Int64Hi x)) (Signmask (Int64Hi x)))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSignmask, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg(v1)
		v.AddArg2(v0, v0)
		return true
	}
	// match: (Rsh64x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh64x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh64x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh64x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh64x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh64x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh64x64 x y)
	// result: (Rsh64x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh64x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh64x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh64x8 x s)
	// result: (Int64Make (Rsh32x8 <typ.UInt32> (Int64Hi x) s) (Or32 <typ.UInt32> (Or32 <typ.UInt32> (Rsh32Ux8 <typ.UInt32> (Int64Lo x) s) (Lsh32x8 <typ.UInt32> (Int64Hi x) (Sub8 <typ.UInt8> (Const8 <typ.UInt8> [32]) s))) (And32 <typ.UInt32> (Rsh32x8 <typ.UInt32> (Int64Hi x) (Sub8 <typ.UInt8> s (Const8 <typ.UInt8> [32]))) (Zeromask (ZeroExt8to32 (Rsh8Ux32 <typ.UInt8> s (Const32 <typ.UInt32> [5])))))))
	for {
		x := v_0
		s := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpRsh32x8, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v0.AddArg2(v1, s)
		v2 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v3 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpRsh32Ux8, typ.UInt32)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(x)
		v4.AddArg2(v5, s)
		v6 := b.NewValue0(v.Pos, OpLsh32x8, typ.UInt32)
		v7 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v8 := b.NewValue0(v.Pos, OpConst8, typ.UInt8)
		v8.AuxInt = int8ToAuxInt(32)
		v7.AddArg2(v8, s)
		v6.AddArg2(v1, v7)
		v3.AddArg2(v4, v6)
		v9 := b.NewValue0(v.Pos, OpAnd32, typ.UInt32)
		v10 := b.NewValue0(v.Pos, OpRsh32x8, typ.UInt32)
		v11 := b.NewValue0(v.Pos, OpSub8, typ.UInt8)
		v11.AddArg2(s, v8)
		v10.AddArg2(v1, v11)
		v12 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v13 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v14 := b.NewValue0(v.Pos, OpRsh8Ux32, typ.UInt8)
		v15 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v15.AuxInt = int32ToAuxInt(5)
		v14.AddArg2(s, v15)
		v13.AddArg(v14)
		v12.AddArg(v13)
		v9.AddArg2(v10, v12)
		v2.AddArg2(v3, v9)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpRsh8Ux16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux16 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt8to32 x) (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh8Ux32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux32 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt8to32 x) y)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg2(v0, y)
		return true
	}
}
func rewriteValueWasm32_OpRsh8Ux64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux64 x (Const64 [c]))
	// cond: uint64(c) < 8
	// result: (I32ShrU x (I32Const [int32(c)]))
	for {
		x := v_0
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) < 8) {
			break
		}
		v.reset(OpWasm32I32ShrU)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(int32(c))
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh8Ux64 _ (Const64 [c]))
	// cond: uint64(c) >= 8
	// result: (I32Const [0])
	for {
		if v_1.Op != OpConst64 {
			break
		}
		c := auxIntToInt64(v_1.AuxInt)
		if !(uint64(c) >= 8) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh8Ux64 _ (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Const32 [0])
	for {
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpConst32)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (Rsh8Ux64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh8Ux32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh8Ux32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh8Ux64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh8Ux32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh8Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh8Ux64 x y)
	// result: (Rsh8Ux32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh8Ux32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh8Ux8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8Ux8 [c] x y)
	// result: (Rsh32Ux32 [c] (ZeroExt8to32 x) (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32Ux32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh8x16(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x16 [c] x y)
	// result: (Rsh32x32 [c] (SignExt8to32 x) (ZeroExt16to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh8x32(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x32 [c] x y)
	// result: (Rsh32x32 [c] (SignExt8to32 x) (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpRsh8x64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x64 x (Int64Make (Const32 [c]) _))
	// cond: c != 0
	// result: (Signmask (SignExt8to32 x))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 {
			break
		}
		c := auxIntToInt32(v_1_0.AuxInt)
		if !(c != 0) {
			break
		}
		v.reset(OpSignmask)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	// match: (Rsh8x64 [c] x (Int64Make (Const32 [0]) lo))
	// result: (Rsh8x32 [c] x lo)
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		v_1_0 := v_1.Args[0]
		if v_1_0.Op != OpConst32 || auxIntToInt32(v_1_0.AuxInt) != 0 {
			break
		}
		v.reset(OpRsh8x32)
		v.AuxInt = boolToAuxInt(c)
		v.AddArg2(x, lo)
		return true
	}
	// match: (Rsh8x64 x (Int64Make hi lo))
	// cond: hi.Op != OpConst32
	// result: (Rsh8x32 x (Or32 <typ.UInt32> (Zeromask hi) lo))
	for {
		x := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		if !(hi.Op != OpConst32) {
			break
		}
		v.reset(OpRsh8x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v1.AddArg(hi)
		v0.AddArg2(v1, lo)
		v.AddArg2(x, v0)
		return true
	}
	// match: (Rsh8x64 x y)
	// result: (Rsh8x32 x (Or32 <typ.UInt32> (Zeromask (Int64Hi y)) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpRsh8x32)
		v0 := b.NewValue0(v.Pos, OpOr32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpZeromask, typ.UInt32)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v1.AddArg(v2)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(y)
		v0.AddArg2(v1, v3)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpRsh8x8(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Rsh8x8 [c] x y)
	// result: (Rsh32x32 [c] (SignExt8to32 x) (ZeroExt8to32 y))
	for {
		c := auxIntToBool(v.AuxInt)
		x := v_0
		y := v_1
		v.reset(OpRsh32x32)
		v.AuxInt = boolToAuxInt(c)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v1 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v1.AddArg(y)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpSignExt16to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt16to32 x:(I32Load16S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load16S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt16to32 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I32Extend16S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I32Extend16S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt16to32 x)
	// result: (I32ShrS (I32Shl x (I32Const [16])) (I32Const [16]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Shl, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(16)
		v0.AddArg2(x, v1)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpSignExt16to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt16to64 x:(I64Load16S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load16S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt16to64 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I64Extend16S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I64Extend16S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt16to64 x)
	// result: (SignExt32to64 (SignExt16to32 x))
	for {
		x := v_0
		v.reset(OpSignExt32to64)
		v0 := b.NewValue0(v.Pos, OpSignExt16to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpSignExt32to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt32to64 x:(I64Load32S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load32S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt32to64 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I64Extend32S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I64Extend32S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt32to64 x)
	// result: (Int64Make (Signmask x) x)
	for {
		x := v_0
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSignmask, typ.Int32)
		v0.AddArg(x)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpSignExt8to16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt8to16 x:(I32Load8S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load8S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt8to16 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I32Extend8S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I32Extend8S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt8to16 x)
	// result: (I32ShrS (I32Shl x (I32Const [24])) (I32Const [24]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Shl, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(24)
		v0.AddArg2(x, v1)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpSignExt8to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt8to32 x:(I32Load8S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load8S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt8to32 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I32Extend8S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I32Extend8S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt8to32 x)
	// result: (I32ShrS (I32Shl x (I32Const [24])) (I32Const [24]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Shl, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(24)
		v0.AddArg2(x, v1)
		v.AddArg2(v0, v1)
		return true
	}
}
func rewriteValueWasm32_OpSignExt8to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (SignExt8to64 x:(I64Load8S _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load8S {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (SignExt8to64 x)
	// cond: buildcfg.GOWASM.SignExt
	// result: (I64Extend8S x)
	for {
		x := v_0
		if !(buildcfg.GOWASM.SignExt) {
			break
		}
		v.reset(OpWasm32I64Extend8S)
		v.AddArg(x)
		return true
	}
	// match: (SignExt8to64 x)
	// result: (SignExt32to64 (SignExt8to32 x))
	for {
		x := v_0
		v.reset(OpSignExt32to64)
		v0 := b.NewValue0(v.Pos, OpSignExt8to32, typ.Int32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpSignmask(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Signmask x)
	// result: (I32ShrS x (I32Const [31]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(31)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpSlicemask(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Slicemask x)
	// result: (I32ShrS (I32Sub (I32Const [0]) x) (I32Const [63]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Sub, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v1.AuxInt = int32ToAuxInt(0)
		v0.AddArg2(v1, x)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v2.AuxInt = int32ToAuxInt(63)
		v.AddArg2(v0, v2)
		return true
	}
}
func rewriteValueWasm32_OpStore(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	config := b.Func.Config
	// match: (Store {t} ptr val mem)
	// cond: is64BitFloat(t)
	// result: (F64Store ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(is64BitFloat(t)) {
			break
		}
		v.reset(OpWasm32F64Store)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: is32BitFloat(t)
	// result: (F32Store ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(is32BitFloat(t)) {
			break
		}
		v.reset(OpWasm32F32Store)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 4
	// result: (I32Store ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 4) {
			break
		}
		v.reset(OpWasm32I32Store)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 2
	// result: (I32Store16 ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 2) {
			break
		}
		v.reset(OpWasm32I32Store16)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} ptr val mem)
	// cond: t.Size() == 1
	// result: (I32Store8 ptr val mem)
	for {
		t := auxToType(v.Aux)
		ptr := v_0
		val := v_1
		mem := v_2
		if !(t.Size() == 1) {
			break
		}
		v.reset(OpWasm32I32Store8)
		v.AddArg3(ptr, val, mem)
		return true
	}
	// match: (Store {t} dst (Int64Make hi lo) mem)
	// cond: t.Size() == 8 && !config.BigEndian
	// result: (Store {hi.Type} (OffPtr <hi.Type.PtrTo()> [4] dst) hi (Store {lo.Type} dst lo mem))
	for {
		t := auxToType(v.Aux)
		dst := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		mem := v_2
		if !(t.Size() == 8 && !config.BigEndian) {
			break
		}
		v.reset(OpStore)
		v.Aux = typeToAux(hi.Type)
		v0 := b.NewValue0(v.Pos, OpOffPtr, hi.Type.PtrTo())
		v0.AuxInt = int64ToAuxInt(4)
		v0.AddArg(dst)
		v1 := b.NewValue0(v.Pos, OpStore, types.TypeMem)
		v1.Aux = typeToAux(lo.Type)
		v1.AddArg3(dst, lo, mem)
		v.AddArg3(v0, hi, v1)
		return true
	}
	// match: (Store {t} dst (Int64Make hi lo) mem)
	// cond: t.Size() == 8 && config.BigEndian
	// result: (Store {lo.Type} (OffPtr <lo.Type.PtrTo()> [4] dst) lo (Store {hi.Type} dst hi mem))
	for {
		t := auxToType(v.Aux)
		dst := v_0
		if v_1.Op != OpInt64Make {
			break
		}
		lo := v_1.Args[1]
		hi := v_1.Args[0]
		mem := v_2
		if !(t.Size() == 8 && config.BigEndian) {
			break
		}
		v.reset(OpStore)
		v.Aux = typeToAux(lo.Type)
		v0 := b.NewValue0(v.Pos, OpOffPtr, lo.Type.PtrTo())
		v0.AuxInt = int64ToAuxInt(4)
		v0.AddArg(dst)
		v1 := b.NewValue0(v.Pos, OpStore, types.TypeMem)
		v1.Aux = typeToAux(hi.Type)
		v1.AddArg3(dst, hi, mem)
		v.AddArg3(v0, lo, v1)
		return true
	}
	return false
}
func rewriteValueWasm32_OpSub64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Sub64 x y)
	// result: (Int64Make (Select0 <typ.UInt32> (Sub64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))) (Select1 <typ.UInt32> (Sub64Decomp (Int64Hi x) (Int64Lo x) (Int64Hi y) (Int64Lo y))))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpSelect0, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpWasm32Sub64Decomp, types.NewTuple(typ.UInt32, typ.UInt32))
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(x)
		v3 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v3.AddArg(x)
		v4 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v4.AddArg(y)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v1.AddArg4(v2, v3, v4, v5)
		v0.AddArg(v1)
		v6 := b.NewValue0(v.Pos, OpSelect1, typ.UInt32)
		v6.AddArg(v1)
		v.AddArg2(v0, v6)
		return true
	}
}
func rewriteValueWasm32_OpTrunc64to16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Trunc64to16 (Int64Make _ lo))
	// result: (Trunc32to16 lo)
	for {
		if v_0.Op != OpInt64Make {
			break
		}
		lo := v_0.Args[1]
		v.reset(OpTrunc32to16)
		v.AddArg(lo)
		return true
	}
	// match: (Trunc64to16 x)
	// result: (Trunc32to16 (Int64Lo x))
	for {
		x := v_0
		v.reset(OpTrunc32to16)
		v0 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpTrunc64to32(v *Value) bool {
	v_0 := v.Args[0]
	// match: (Trunc64to32 (Int64Make _ lo))
	// result: lo
	for {
		if v_0.Op != OpInt64Make {
			break
		}
		lo := v_0.Args[1]
		v.copyOf(lo)
		return true
	}
	// match: (Trunc64to32 x)
	// result: (Int64Lo x)
	for {
		x := v_0
		v.reset(OpInt64Lo)
		v.AddArg(x)
		return true
	}
}
func rewriteValueWasm32_OpTrunc64to8(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Trunc64to8 (Int64Make _ lo))
	// result: (Trunc32to8 lo)
	for {
		if v_0.Op != OpInt64Make {
			break
		}
		lo := v_0.Args[1]
		v.reset(OpTrunc32to8)
		v.AddArg(lo)
		return true
	}
	// match: (Trunc64to8 x)
	// result: (Trunc32to8 (Int64Lo x))
	for {
		x := v_0
		v.reset(OpTrunc32to8)
		v0 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpWasm32F32Add(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (F32Add (F32Const [x]) y)
	// cond: y.Op != OpWasm32F32Const
	// result: (F32Add y (F32Const [x]))
	for {
		if v_0.Op != OpWasm32F32Const {
			break
		}
		x := auxIntToFloat32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32F32Const) {
			break
		}
		v.reset(OpWasm32F32Add)
		v0 := b.NewValue0(v.Pos, OpWasm32F32Const, typ.Float32)
		v0.AuxInt = float32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32F32Mul(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (F32Mul (F32Const [x]) y)
	// cond: y.Op != OpWasm32F32Const
	// result: (F32Mul y (F32Const [x]))
	for {
		if v_0.Op != OpWasm32F32Const {
			break
		}
		x := auxIntToFloat32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32F32Const) {
			break
		}
		v.reset(OpWasm32F32Mul)
		v0 := b.NewValue0(v.Pos, OpWasm32F32Const, typ.Float32)
		v0.AuxInt = float32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32F64Add(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (F64Add (F64Const [x]) (F64Const [y]))
	// result: (F64Const [x + y])
	for {
		if v_0.Op != OpWasm32F64Const {
			break
		}
		x := auxIntToFloat64(v_0.AuxInt)
		if v_1.Op != OpWasm32F64Const {
			break
		}
		y := auxIntToFloat64(v_1.AuxInt)
		v.reset(OpWasm32F64Const)
		v.AuxInt = float64ToAuxInt(x + y)
		return true
	}
	// match: (F64Add (F64Const [x]) y)
	// cond: y.Op != OpWasm32F64Const
	// result: (F64Add y (F64Const [x]))
	for {
		if v_0.Op != OpWasm32F64Const {
			break
		}
		x := auxIntToFloat64(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32F64Const) {
			break
		}
		v.reset(OpWasm32F64Add)
		v0 := b.NewValue0(v.Pos, OpWasm32F64Const, typ.Float64)
		v0.AuxInt = float64ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32F64Mul(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (F64Mul (F64Const [x]) (F64Const [y]))
	// cond: !math.IsNaN(x * y)
	// result: (F64Const [x * y])
	for {
		if v_0.Op != OpWasm32F64Const {
			break
		}
		x := auxIntToFloat64(v_0.AuxInt)
		if v_1.Op != OpWasm32F64Const {
			break
		}
		y := auxIntToFloat64(v_1.AuxInt)
		if !(!math.IsNaN(x * y)) {
			break
		}
		v.reset(OpWasm32F64Const)
		v.AuxInt = float64ToAuxInt(x * y)
		return true
	}
	// match: (F64Mul (F64Const [x]) y)
	// cond: y.Op != OpWasm32F64Const
	// result: (F64Mul y (F64Const [x]))
	for {
		if v_0.Op != OpWasm32F64Const {
			break
		}
		x := auxIntToFloat64(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32F64Const) {
			break
		}
		v.reset(OpWasm32F64Mul)
		v0 := b.NewValue0(v.Pos, OpWasm32F64Const, typ.Float64)
		v0.AuxInt = float64ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Add(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Add (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x + y])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x + y)
		return true
	}
	// match: (I32Add (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Add y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Add)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	// match: (I32Add x (I32Const <t> [y]))
	// cond: !t.IsPtr()
	// result: (I32AddConst [y] x)
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const {
			break
		}
		t := v_1.Type
		y := auxIntToInt32(v_1.AuxInt)
		if !(!t.IsPtr()) {
			break
		}
		v.reset(OpWasm32I32AddConst)
		v.AuxInt = int32ToAuxInt(y)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32AddConst(v *Value) bool {
	v_0 := v.Args[0]
	// match: (I32AddConst [0] x)
	// result: x
	for {
		if auxIntToInt32(v.AuxInt) != 0 {
			break
		}
		x := v_0
		v.copyOf(x)
		return true
	}
	// match: (I32AddConst [off] (LoweredAddr {sym} [off2] base))
	// cond: isU32Bit(int64(off+int32(off2)))
	// result: (LoweredAddr {sym} [int32(off)+off2] base)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32LoweredAddr {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		sym := auxToSym(v_0.Aux)
		base := v_0.Args[0]
		if !(isU32Bit(int64(off + int32(off2)))) {
			break
		}
		v.reset(OpWasm32LoweredAddr)
		v.AuxInt = int32ToAuxInt(int32(off) + off2)
		v.Aux = symToAux(sym)
		v.AddArg(base)
		return true
	}
	// match: (I32AddConst [off] x:(SP))
	// cond: isU32Bit(int64(off))
	// result: (LoweredAddr [int32(off)] x)
	for {
		off := auxIntToInt32(v.AuxInt)
		x := v_0
		if x.Op != OpSP || !(isU32Bit(int64(off))) {
			break
		}
		v.reset(OpWasm32LoweredAddr)
		v.AuxInt = int32ToAuxInt(int32(off))
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32And(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32And (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x & y])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x & y)
		return true
	}
	// match: (I32And (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32And y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32And)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Eq(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Eq (I32Const [x]) (I32Const [y]))
	// cond: x == y
	// result: (I32Const [1])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		if !(x == y) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (I32Eq (I32Const [x]) (I32Const [y]))
	// cond: x != y
	// result: (I32Const [0])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		if !(x != y) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (I32Eq (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Eq y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Eq)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	// match: (I32Eq x (I32Const [0]))
	// result: (I32Eqz x)
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const || auxIntToInt32(v_1.AuxInt) != 0 {
			break
		}
		v.reset(OpWasm32I32Eqz)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Eqz(v *Value) bool {
	v_0 := v.Args[0]
	// match: (I32Eqz (I32Eqz (I32Eqz x)))
	// result: (I32Eqz x)
	for {
		if v_0.Op != OpWasm32I32Eqz {
			break
		}
		v_0_0 := v_0.Args[0]
		if v_0_0.Op != OpWasm32I32Eqz {
			break
		}
		x := v_0_0.Args[0]
		v.reset(OpWasm32I32Eqz)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32LeU(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32LeU x (I32Const [0]))
	// result: (I32Eqz x)
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const || auxIntToInt32(v_1.AuxInt) != 0 {
			break
		}
		v.reset(OpWasm32I32Eqz)
		v.AddArg(x)
		return true
	}
	// match: (I32LeU (I32Const [1]) x)
	// result: (I32Eqz (I32Eqz x))
	for {
		if v_0.Op != OpWasm32I32Const || auxIntToInt32(v_0.AuxInt) != 1 {
			break
		}
		x := v_1
		v.reset(OpWasm32I32Eqz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Eqz, typ.Bool)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Load(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	config := b.Func.Config
	// match: (I32Load [off] (I32AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Load [off+off2] ptr mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Load)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (I32Load [off] (LoweredAddr {sym} [off2] (SB)) _)
	// cond: symIsRO(sym) && isU32Bit(int64(off+int32(off2)))
	// result: (I32Const [int32(read32(sym, int64(off+int32(off2)), config.ctxt.Arch.ByteOrder))])
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32LoweredAddr {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		sym := auxToSym(v_0.Aux)
		v_0_0 := v_0.Args[0]
		if v_0_0.Op != OpSB || !(symIsRO(sym) && isU32Bit(int64(off+int32(off2)))) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(read32(sym, int64(off+int32(off2)), config.ctxt.Arch.ByteOrder)))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Load16S(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Load16S [off] (I32AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Load16S [off+off2] ptr mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Load16S)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Load16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	config := b.Func.Config
	// match: (I32Load16U [off] (I32AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Load16U [off+off2] ptr mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Load16U)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (I32Load16U [off] (LoweredAddr {sym} [off2] (SB)) _)
	// cond: symIsRO(sym) && isU32Bit(int64(off+int32(off2)))
	// result: (I32Const [int32(read16(sym, int64(off+int32(off2)), config.ctxt.Arch.ByteOrder))])
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32LoweredAddr {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		sym := auxToSym(v_0.Aux)
		v_0_0 := v_0.Args[0]
		if v_0_0.Op != OpSB || !(symIsRO(sym) && isU32Bit(int64(off+int32(off2)))) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(read16(sym, int64(off+int32(off2)), config.ctxt.Arch.ByteOrder)))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Load8S(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Load8S [off] (I32AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Load8S [off+off2] ptr mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Load8S)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Load8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Load8U [off] (I32AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Load8U [off+off2] ptr mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Load8U)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	// match: (I32Load8U [off] (LoweredAddr {sym} [off2] (SB)) _)
	// cond: symIsRO(sym) && isU32Bit(int64(off+int32(off2)))
	// result: (I32Const [int32(read8(sym, int64(off+int32(off2))))])
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32LoweredAddr {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		sym := auxToSym(v_0.Aux)
		v_0_0 := v_0.Args[0]
		if v_0_0.Op != OpSB || !(symIsRO(sym) && isU32Bit(int64(off+int32(off2)))) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(read8(sym, int64(off+int32(off2)))))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32LtU(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32LtU (I32Const [0]) x)
	// result: (I32Eqz (I32Eqz x))
	for {
		if v_0.Op != OpWasm32I32Const || auxIntToInt32(v_0.AuxInt) != 0 {
			break
		}
		x := v_1
		v.reset(OpWasm32I32Eqz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Eqz, typ.Bool)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	// match: (I32LtU x (I32Const [1]))
	// result: (I32Eqz x)
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const || auxIntToInt32(v_1.AuxInt) != 1 {
			break
		}
		v.reset(OpWasm32I32Eqz)
		v.AddArg(x)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Mul(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Mul (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x * y])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x * y)
		return true
	}
	// match: (I32Mul (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Mul y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Mul)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Ne(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Ne (I32Const [x]) (I32Const [y]))
	// cond: x == y
	// result: (I32Const [0])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		if !(x == y) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(0)
		return true
	}
	// match: (I32Ne (I32Const [x]) (I32Const [y]))
	// cond: x != y
	// result: (I32Const [1])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		if !(x != y) {
			break
		}
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(1)
		return true
	}
	// match: (I32Ne (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Ne y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Ne)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	// match: (I32Ne x (I32Const [0]))
	// result: (I32Eqz (I32Eqz x))
	for {
		x := v_0
		if v_1.Op != OpWasm32I32Const || auxIntToInt32(v_1.AuxInt) != 0 {
			break
		}
		v.reset(OpWasm32I32Eqz)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Eqz, typ.Bool)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Or(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Or (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x | y])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x | y)
		return true
	}
	// match: (I32Or (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Or y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Or)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Shl(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Shl (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x << uint32(y)])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x << uint32(y))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32ShrS(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32ShrS (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x >> uint32(y)])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x >> uint32(y))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32ShrU(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32ShrU (I32Const [x]) (I32Const [y]))
	// result: (I32Const [int32(uint32(x) >> uint32(y))])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(int32(uint32(x) >> uint32(y)))
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Store(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Store [off] (I32AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Store [off+off2] ptr val mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Store16(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Store16 [off] (I32AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Store16 [off+off2] ptr val mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Store16)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Store8(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I32Store8 [off] (I32AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I32Store8 [off+off2] ptr val mem)
	for {
		off := auxIntToInt32(v.AuxInt)
		if v_0.Op != OpWasm32I32AddConst {
			break
		}
		off2 := auxIntToInt32(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I32Store8)
		v.AuxInt = int32ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I32Xor(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (I32Xor (I32Const [x]) (I32Const [y]))
	// result: (I32Const [x ^ y])
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		if v_1.Op != OpWasm32I32Const {
			break
		}
		y := auxIntToInt32(v_1.AuxInt)
		v.reset(OpWasm32I32Const)
		v.AuxInt = int32ToAuxInt(x ^ y)
		return true
	}
	// match: (I32Xor (I32Const [x]) y)
	// cond: y.Op != OpWasm32I32Const
	// result: (I32Xor y (I32Const [x]))
	for {
		if v_0.Op != OpWasm32I32Const {
			break
		}
		x := auxIntToInt32(v_0.AuxInt)
		y := v_1
		if !(y.Op != OpWasm32I32Const) {
			break
		}
		v.reset(OpWasm32I32Xor)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(x)
		v.AddArg2(y, v0)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load16S(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load16S [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load16S [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load16S)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load16U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load16U [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load16U [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load16U)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load32S(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load32S [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load32S [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load32S)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load32U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load32U [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load32U [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load32U)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load8S(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load8S [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load8S [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load8S)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Load8U(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Load8U [off] (I64AddConst [off2] ptr) mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Load8U [off+off2] ptr mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		mem := v_1
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Load8U)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg2(ptr, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Store(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Store [off] (I64AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Store [off+off2] ptr val mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Store)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Store16(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Store16 [off] (I64AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Store16 [off+off2] ptr val mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Store16)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Store32(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Store32 [off] (I64AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Store32 [off+off2] ptr val mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Store32)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpWasm32I64Store8(v *Value) bool {
	v_2 := v.Args[2]
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	// match: (I64Store8 [off] (I64AddConst [off2] ptr) val mem)
	// cond: isU32Bit(int64(off+off2))
	// result: (I64Store8 [off+off2] ptr val mem)
	for {
		off := auxIntToInt64(v.AuxInt)
		if v_0.Op != OpWasm32I64AddConst {
			break
		}
		off2 := auxIntToInt64(v_0.AuxInt)
		ptr := v_0.Args[0]
		val := v_1
		mem := v_2
		if !(isU32Bit(int64(off + off2))) {
			break
		}
		v.reset(OpWasm32I64Store8)
		v.AuxInt = int64ToAuxInt(off + off2)
		v.AddArg3(ptr, val, mem)
		return true
	}
	return false
}
func rewriteValueWasm32_OpXor64(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Xor64 x y)
	// result: (Int64Make (Xor32 <typ.UInt32> (Int64Hi x) (Int64Hi y)) (Xor32 <typ.UInt32> (Int64Lo x) (Int64Lo y)))
	for {
		x := v_0
		y := v_1
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpXor32, typ.UInt32)
		v1 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v1.AddArg(x)
		v2 := b.NewValue0(v.Pos, OpInt64Hi, typ.UInt32)
		v2.AddArg(y)
		v0.AddArg2(v1, v2)
		v3 := b.NewValue0(v.Pos, OpXor32, typ.UInt32)
		v4 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v4.AddArg(x)
		v5 := b.NewValue0(v.Pos, OpInt64Lo, typ.UInt32)
		v5.AddArg(y)
		v3.AddArg2(v4, v5)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteValueWasm32_OpZero(v *Value) bool {
	v_1 := v.Args[1]
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Zero [0] _ mem)
	// result: mem
	for {
		if auxIntToInt64(v.AuxInt) != 0 {
			break
		}
		mem := v_1
		v.copyOf(mem)
		return true
	}
	// match: (Zero [1] destptr mem)
	// result: (I32Store8 destptr (I32Const [0]) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 1 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store8)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg3(destptr, v0, mem)
		return true
	}
	// match: (Zero [2] destptr mem)
	// result: (I32Store16 destptr (I32Const [0]) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 2 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store16)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg3(destptr, v0, mem)
		return true
	}
	// match: (Zero [4] destptr mem)
	// result: (I32Store destptr (I32Const [0]) mem)
	for {
		if auxIntToInt64(v.AuxInt) != 4 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg3(destptr, v0, mem)
		return true
	}
	// match: (Zero [3] destptr mem)
	// result: (I32Store8 [2] destptr (I32Const [0]) (I32Store16 destptr (I32Const [0]) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 3 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store8)
		v.AuxInt = int32ToAuxInt(2)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store16, types.TypeMem)
		v1.AddArg3(destptr, v0, mem)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [5] destptr mem)
	// result: (I32Store8 [4] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 5 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store8)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AddArg3(destptr, v0, mem)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [6] destptr mem)
	// result: (I32Store16 [4] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 6 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store16)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AddArg3(destptr, v0, mem)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [7] destptr mem)
	// result: (I32Store [3] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 7 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(3)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AddArg3(destptr, v0, mem)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [8] destptr mem)
	// result: (I32Store [4] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem))
	for {
		if auxIntToInt64(v.AuxInt) != 8 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(4)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AddArg3(destptr, v0, mem)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [s] destptr mem)
	// cond: s%4 != 0 && s > 4 && s < 16
	// result: (Zero [s-s%4] (OffPtr <destptr.Type> destptr [s%4]) (I32Store destptr (I32Const [0]) mem))
	for {
		s := auxIntToInt64(v.AuxInt)
		destptr := v_0
		mem := v_1
		if !(s%4 != 0 && s > 4 && s < 16) {
			break
		}
		v.reset(OpZero)
		v.AuxInt = int64ToAuxInt(s - s%4)
		v0 := b.NewValue0(v.Pos, OpOffPtr, destptr.Type)
		v0.AuxInt = int64ToAuxInt(s % 4)
		v0.AddArg(destptr)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v2.AuxInt = int32ToAuxInt(0)
		v1.AddArg3(destptr, v2, mem)
		v.AddArg2(v0, v1)
		return true
	}
	// match: (Zero [12] destptr mem)
	// result: (I32Store [8] destptr (I32Const [0]) (I32Store [4] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem)))
	for {
		if auxIntToInt64(v.AuxInt) != 12 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(8)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AuxInt = int32ToAuxInt(4)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2.AddArg3(destptr, v0, mem)
		v1.AddArg3(destptr, v0, v2)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [16] destptr mem)
	// result: (I32Store [12] destptr (I32Const [0]) (I32Store [8] destptr (I32Const [0]) (I32Store [4] destptr (I32Const [0]) (I32Store destptr (I32Const [0]) mem))))
	for {
		if auxIntToInt64(v.AuxInt) != 16 {
			break
		}
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32I32Store)
		v.AuxInt = int32ToAuxInt(12)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v1.AuxInt = int32ToAuxInt(8)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v2.AuxInt = int32ToAuxInt(4)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Store, types.TypeMem)
		v3.AddArg3(destptr, v0, mem)
		v2.AddArg3(destptr, v0, v3)
		v1.AddArg3(destptr, v0, v2)
		v.AddArg3(destptr, v0, v1)
		return true
	}
	// match: (Zero [s] destptr mem)
	// result: (LoweredZero [s] destptr mem)
	for {
		s := auxIntToInt64(v.AuxInt)
		destptr := v_0
		mem := v_1
		v.reset(OpWasm32LoweredZero)
		v.AuxInt = int64ToAuxInt(s)
		v.AddArg2(destptr, mem)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt16to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt16to32 x:(I32Load16U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load16U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt16to32 x)
	// result: (I32And x (I32Const [0xffff]))
	for {
		x := v_0
		v.reset(OpWasm32I32And)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0xffff)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt16to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt16to64 x:(I64Load16U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load16U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt16to64 x)
	// result: (ZeroExt32to64 (ZeroExt16to32 x))
	for {
		x := v_0
		v.reset(OpZeroExt32to64)
		v0 := b.NewValue0(v.Pos, OpZeroExt16to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt32to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt32to64 x:(I64Load32U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load32U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt32to64 x)
	// result: (Int64Make (Const32 <typ.UInt32> [0]) x)
	for {
		x := v_0
		v.reset(OpInt64Make)
		v0 := b.NewValue0(v.Pos, OpConst32, typ.UInt32)
		v0.AuxInt = int32ToAuxInt(0)
		v.AddArg2(v0, x)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt8to16(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt8to16 x:(I32Load8U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load8U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt8to16 x)
	// result: (I32And x (I32Const [0xff]))
	for {
		x := v_0
		v.reset(OpWasm32I32And)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0xff)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt8to32(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt8to32 x:(I32Load8U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I32Load8U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt8to32 x)
	// result: (I32And x (I32Const [0xff]))
	for {
		x := v_0
		v.reset(OpWasm32I32And)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v0.AuxInt = int32ToAuxInt(0xff)
		v.AddArg2(x, v0)
		return true
	}
}
func rewriteValueWasm32_OpZeroExt8to64(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (ZeroExt8to64 x:(I64Load8U _ _))
	// result: x
	for {
		x := v_0
		if x.Op != OpWasm32I64Load8U {
			break
		}
		v.copyOf(x)
		return true
	}
	// match: (ZeroExt8to64 x)
	// result: (ZeroExt32to64 (ZeroExt8to32 x))
	for {
		x := v_0
		v.reset(OpZeroExt32to64)
		v0 := b.NewValue0(v.Pos, OpZeroExt8to32, typ.UInt32)
		v0.AddArg(x)
		v.AddArg(v0)
		return true
	}
}
func rewriteValueWasm32_OpZeromask(v *Value) bool {
	v_0 := v.Args[0]
	b := v.Block
	typ := &b.Func.Config.Types
	// match: (Zeromask x)
	// result: (I32ShrS (I32Sub (I32Shl x (I32Const [1])) x) (I32Const [31]))
	for {
		x := v_0
		v.reset(OpWasm32I32ShrS)
		v0 := b.NewValue0(v.Pos, OpWasm32I32Sub, typ.Int32)
		v1 := b.NewValue0(v.Pos, OpWasm32I32Shl, typ.Int32)
		v2 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v2.AuxInt = int32ToAuxInt(1)
		v1.AddArg2(x, v2)
		v0.AddArg2(v1, x)
		v3 := b.NewValue0(v.Pos, OpWasm32I32Const, typ.Int32)
		v3.AuxInt = int32ToAuxInt(31)
		v.AddArg2(v0, v3)
		return true
	}
}
func rewriteBlockWasm32(b *Block) bool {
	return false
}
