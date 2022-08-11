// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

import "cmd/internal/obj"

// inst represents an instruction.
type Inst struct {
	GoOp     obj.As    // Go opcode mnemonic
	ArmOp    A64Type   // Arm64 opcode mnemonic
	Feature  uint16    // such as "FEAT_LSE", "FEAT_CSSC"
	Skeleton uint32    // known bits
	Mask     uint32    // mask for disassembly, 1 for known bits, 0 for unknown bits
	Alias    bool      // whether it is an alias
	Args     []Operand // operands, in Go order
}

// Operand is the operand type of an instruction.
type Operand struct {
	Class AClass    // operand class, register, constant, memory operation etc.
	Elms  []ElmType // the elements that this operand includes
}

// A64Type is the Arm64 opcode type, an Arm64 opcode is prefixed with "A64",
// a Go opcode is defined with a constant and is prefixed with "A".
type A64Type uint16

// ElmType is the element type, an element represents a symbol of a specific encoding form,
// such as <Xn>, #<uimm4>, <T>.
type ElmType uint16

type Icmp []Inst

func (x Icmp) Len() int {
	return len(x)
}

func (x Icmp) Swap(i, j int) {
	x[i], x[j] = x[j], x[i]
}

func (x Icmp) Less(i, j int) bool {
	p1 := &x[i]
	p2 := &x[j]
	if p1.GoOp != p2.GoOp {
		return p1.GoOp < p2.GoOp
	}
	if len(p1.Args) != len(p2.Args) {
		return len(p1.Args) < len(p2.Args)
	}
	for k := 0; k < len(p1.Args); k++ {
		if p1.Args[k].Class != p2.Args[k].Class {
			return p1.Args[k].Class < p2.Args[k].Class
		}
	}
	if p1.Skeleton != p2.Skeleton {
		return p1.Skeleton < p2.Skeleton
	}
	if p1.Mask != p2.Mask {
		return p1.Mask < p2.Mask
	}
	return false
}

// These constants represent Arm a-profile architecture extensions. For details,
// please refer to the Arm a-profile architecture reference manual.
// Update this table if new extensions are found when parsing the XML files.
const (
	FEAT_NONE uint16 = iota
	// The Armv8.0 Cryptographic Extension provides instructions for the acceleration
	// of encryption and decryption, and includes the following features:
	// FEAT_AES, FEAT_PMULL, FEAT_SHA1 and FEAT_SHA256.
	FEAT_AES
	// FEAT_ASMv8p2 adds the BFC instruction to the A64 instruction set as an alias of BFM.
	// It also requires that the BFC instruction and the A64 pseudo-instruction REV64 are
	// implemented by assemblers.
	FEAT_ASMv8p2
	// FEAT_BF16 supports the BFloat16, or BF16, 16-bit floating-point storage format in AArch64 state.
	FEAT_BF16
	// FEAT_BRBE provides a Branch record buffer for capturing control path history.
	FEAT_BRBE
	// FEAT_BTI allows memory pages to be guarded against the execution of instructions that are not the
	// intended target of a branch.
	FEAT_BTI
	// Check feature status extension.
	FEAT_CHK
	FEAT_CLRBHB
	// Changes to CRC32 instructions.
	FEAT_CRC32
	FEAT_CSSC
	FEAT_D128
	// FEAT_D128 && FEAT_THE
	FEAT_D128__THE
	// FEAT_DGH adds the Data Gathering Hint instruction to the hint space.
	FEAT_DGH
	// FEAT_DotProd provides instructions to perform the dot product of two 32-bit vectors,
	// accumulating the result in a third 32-bit vector.
	FEAT_DotProd
	// FEAT_F32MM adds support for the SVE FP32 single-precision floating-point matrix
	// multiplication variant of the FMMLA instruction.
	FEAT_F32MM
	FEAT_F64MM
	// FEAT_FCMA introduces instructions for floating-point multiplication and addition of complex numbers.
	FEAT_FCMA
	// FEAT_FHM adds floating-point multiplication instructions.
	FEAT_FHM
	// Half-precision floating-point data processing.
	FEAT_FP16
	// FEAT_FRINTTS provides instructions that round a floating-point number to an integral valued
	// floating-point number that fits in a 32-bit or 64-bit integer number range.
	FEAT_FRINTTS
	// FEAT_FlagM provides instructions which manipulate the PSTATE.{N,Z,C,V} flags.
	FEAT_FlagM
	// Enhancements to flag manipulation instructions.
	FEAT_FlagM2
	// The Guarded control stack provides protection against use of procedure return instructions
	// to return anywhere other than the return address created when the procedure was called.
	FEAT_GCS
	// FEAT_HBC provides the BC.cond instruction to give a conditional branch with a hint to branch
	// prediction logic that this branch will consistently and is highly unlikely to change direction.
	FEAT_HBC
	// FEAT_I8MM introduces integer matrix multiply-accumulate instructions and mixed sign dot
	// product instructions.
	FEAT_I8MM
	FEAT_ITE
	// FEAT_JSCVT introduces instructions that perform a conversion from a double-precision floating
	// point value to a signed 32-bit integer, with rounding to zero.
	FEAT_JSCVT
	// FEAT_LOR provides a set of instructions with Acquire semantics for loads, and Release semantics
	// for stores that apply in relation to the defined LORegions.
	FEAT_LOR
	// FEAT_LRCPC introduces three instructions to support the weaker Release Consistency processor
	// consistent (RCpc) model that enables the reordering of a Store-Release followed by a Load-Acquire
	// to a different address.
	FEAT_LRCPC
	FEAT_LRCPC2
	FEAT_LRCPC3
	// FEAT_LS64 introduces support for atomic single-copy 64-byte loads and stores without return.
	FEAT_LS64
	FEAT_LS64_ACCDATA
	FEAT_LS64_V
	// FEAT_LSE introduces a set of atomic instructions.
	FEAT_LSE
	FEAT_LSE128
	// FEAT_MOPS provides instructions that perform a memory copy or memory set, and adds Memory
	// Copy and Memory Set exceptions.
	FEAT_MOPS
	// FEAT_MTE and FEAT_MTE2 provide architectural support for runtime, always-on detection of
	// various classes of memory error to aid with software debugging to eliminate vulnerabilities
	// arising from memory-unsafe languages.
	FEAT_MTE
	FEAT_MTE2
	// FEAT_PAuth adds functionality that supports address authentication of the contents of a register
	// before that register is used as the target of an indirect branch, or as a load.
	FEAT_PAuth
	// RAS Extension v1.1.
	FEAT_RAS
	// FEAT_RDM introduces Rounding Double Multiply Add/Subtract Advanced SIMD instructions.
	FEAT_RDM
	FEAT_RPRFM
	// FEAT_SB introduces a barrier to control speculation.
	FEAT_SB
	FEAT_SHA1
	FEAT_SHA256
	// FEAT_SHA3 adds Advanced SIMD instructions that support SHA3 functionality.
	FEAT_SHA3
	// FEAT_SHA512 adds Advanced SIMD instructions that support SHA2-512 functionality.
	FEAT_SHA512
	// FEAT_SM3 adds Advanced SIMD instructions that support the Chinese cryptography algorithm SM3.
	FEAT_SM3
	FEAT_SM4
	// Scalable Matrix Extension.
	FEAT_SME
	FEAT_SME2
	FEAT_SME2p1
	FEAT_SME_F16F16
	FEAT_SME_F64F64
	FEAT_SME_I16I64
	// FEAT_SPECRES adds the CFP RCTX, CPP RCTX, DVP RCTX, CFPRCTX, CPPRCTX, and DVPRCTX System instructions.
	FEAT_SPE
	FEAT_SPECRES
	FEAT_SPECRES2
	FEAT_SVE2p1
	// Scalable Vector AES instructions.
	FEAT_SVE_AES
	FEAT_SVE_B16B16
	// Scalable Vector Bit Permutes instructions.
	FEAT_SVE_BitPerm
	// Scalable Vector PMULL instructions.
	FEAT_SVE_PMULL128
	// Scalable Vector SHA3 instructions.
	FEAT_SVE_SHA3
	FEAT_SVE_SM4
	FEAT_SYSINSTR128
	FEAT_SYSREG128
	FEAT_THE
	// FEAT_TME introduces a set of instructions to support hardware transaction memory, which means
	// a group of instructions can appear to be collectively executed as a single atomic operation.
	FEAT_TME
	// FEAT_TRF adds controls of trace in a self-hosted system through System registers.
	FEAT_TRF
	// FEAT_WFxT introduces WFET and WFIT.
	FEAT_WFxT
	// FEAT_XS introduces the XS attribute for memory to indicate that an access could take a long time
	// to complete.
	FEAT_XS
)
