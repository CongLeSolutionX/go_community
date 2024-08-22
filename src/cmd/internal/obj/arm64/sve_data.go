// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package arm64

import "cmd/internal/obj"

// Mapping of mnemonic -> list of encoding. Each instruction has multiple potential encodings,
// for each different format of operands possible with the instruction. For example, ADD has
// 3 different formats:
//   - ADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
//   - ADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
//   - ADD <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
//
// Generic lane sizes such as <T> may be expanded into further formats, e.g. we may expand
// <T> into {B,H,S,D} to make it easier for the assembler to find a match with user input.
var instructionTable = map[obj.As][]encoding{
	AZABS: {
		{0x0416a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // ABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZADD: {
		{0x04000000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZAND: {
		{0x041a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // AND <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZASR: {
		{0x04108000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZASRR: {
		{0x04148000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZBFDOT: {
		{0x64608000, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFDOT <Zda>.S, <Zn>.H, <Zm>.H
		{0x64604000, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i2}, // BFDOT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
	AZBIC: {
		{0x041b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // BIC <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZCLS: {
		{0x0418a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CLS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCLZ: {
		{0x0419a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CLZ <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCMPEQ: {
		{0x25008000, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPEQ <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x2400a000, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPEQ <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x24002000, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPEQ <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPGE: {
		{0x25000000, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24008000, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x24004000, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPGT: {
		{0x25000010, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24008010, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x24004010, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPHI: {
		{0x24200010, FG_PdT_PgZ_ZnT_Imm, E_size_uimm7_Pg_Zn_Pd}, // CMPHI <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24000010, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPHI <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x2400c010, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPHI <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPHS: {
		{0x24200000, FG_PdT_PgZ_ZnT_Imm, E_size_uimm7_Pg_Zn_Pd}, // CMPHS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24000000, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPHS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x2400c000, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPHS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPLE: {
		{0x25002010, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPLE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24006010, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPLE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPLO: {
		{0x24202000, FG_PdT_PgZ_ZnT_Imm, E_size_uimm7_Pg_Zn_Pd}, // CMPLO <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x2400e000, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPLO <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPLS: {
		{0x24202010, FG_PdT_PgZ_ZnT_Imm, E_size_uimm7_Pg_Zn_Pd}, // CMPLS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x2400e010, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPLS <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPLT: {
		{0x25002000, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPLT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x24006000, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPLT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCMPNE: {
		{0x25008010, FG_PdT_PgZ_ZnT_Imm, E_size_simm5_Pg_Zn_Pd}, // CMPNE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #<imm>
		{0x2400a010, FG_PdT_PgZ_ZnT_ZmT, E_size_Zm_Pg_Zn_Pd},    // CMPNE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
		{0x24002010, FG_PdT_PgZ_ZnT_ZmD, E_size_ZmD_Pg_Zn_Pd},   // CMPNE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.D
	},
	AZCNOT: {
		{0x041ba000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CNOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZCNT: {
		{0x041aa000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // CNT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZDECP: {
		{0x252d8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // DECP <Xdn>, <Pm>.<T>
	},
	AZEOR: {
		{0x04190000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // EOR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABD: {
		{0x65088000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABS: {
		{0x041ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFADD: {
		{0x65008000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFACGE: {
		{0x6500c010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd}, // FACGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFACGT: {
		{0x6500e010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd}, // FACGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFCMEQ: {
		{0x65122000, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMEQ <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
		{0x65006000, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd},    // FCMEQ <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFCMGE: {
		{0x65102000, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
		{0x65004000, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd},    // FCMGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFCMGT: {
		{0x65102010, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
		{0x65004010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd},    // FCMGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFCMLE: {
		{0x65112010, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMLE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
	},
	AZFCMLT: {
		{0x65112000, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMLT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
	},
	AZFCMNE: {
		{0x65132000, FG_PdT_PgZ_ZnT_ImmFP, E_size_Pg_Zn_Pd_ImmFP0}, // FCMNE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, #0.0
		{0x65006010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd},    // FCMNE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFCMUO: {
		{0x6500c000, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd}, // FCMUO <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFDIV: {
		{0x650d8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFDIVR: {
		{0x650c8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMAX: {
		{0x65068000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMAXNM: {
		{0x65048000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMAXNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMIN: {
		{0x65078000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMINNM: {
		{0x65058000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMINNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMUL: {
		{0x65028000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMULX: {
		{0x650a8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMULX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFNEG: {
		{0x041da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FNEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRECPX: {
		{0x650ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRECPX <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTA: {
		{0x6504a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTA <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTI: {
		{0x6507a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTI <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTM: {
		{0x6502a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTM <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTN: {
		{0x6500a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTN <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTP: {
		{0x6501a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTP <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTX: {
		{0x6506a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTX <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRINTZ: {
		{0x6503a000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FRINTZ <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFSCALE: {
		{0x65098000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSCALE <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSUB: {
		{0x65018000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSUBR: {
		{0x65038000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSQRT: {
		{0x650da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FSQRT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZINCP: {
		{0x252c8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // INCP <Xdn>, <Pm>.<T>
	},
	AZLDR: {
		{0x85804000, []int{F_Zt_AddrXSP, F_Zt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // LDR <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0x85800000, []int{F_Pt_AddrXSP, F_Pt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // LDR <Pt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLSL: {
		{0x04138000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSLR: {
		{0x04178000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSLR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSR: {
		{0x04118000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSRR: {
		{0x04158000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZMUL: {
		{0x04100000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // MUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZNEG: {
		{0x0417a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZNOT: {
		{0x041ea000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZORR: {
		{0x04180000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ORR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZRBIT: {
		{0x05278000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // RBIT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVB: {
		{0x05248000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // REVB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVH: {
		{0x05258000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // REVH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZSABD: {
		{0x040c0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSADDV: {
		{0x04002000, []int{F_Dd_Pg_ZnB, F_Dd_Pg_ZnH, F_Dd_Pg_ZnS}, E_size_Pg_Zn_Vd}, // SADDV <Dd>, <Pg>, <Zn>.<T>
	},
	AZSDIV: {
		{0x04140000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSDIVR: {
		{0x04160000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMAX: {
		{0x04080000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMIN: {
		{0x040a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMULH: {
		{0x04120000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSTR: {
		{0xe5804000, []int{F_Zt_AddrXSP, F_Zt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe5800000, []int{F_Pt_AddrXSP, F_Pt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Pt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSUB: {
		{0x04010000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSUBR: {
		{0x04030000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSXTB: {
		{0x0410a000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // SXTB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZSXTH: {
		{0x0412a000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // SXTH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUABD: {
		{0x040d0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUDIV: {
		{0x04150000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUDIVR: {
		{0x04170000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMAX: {
		{0x04090000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMIN: {
		{0x040b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMULH: {
		{0x04130000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUXTB: {
		{0x0411a000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // UXTB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUXTH: {
		{0x0413a000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // UXTH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
}

// Key into the format table.
const (
	F_ZdaS_ZnH_ZmH = iota
	F_ZdaS_ZnH_ZmHidx
	F_ZdB_PgM_ZnB
	F_ZdH_PgM_ZnH
	F_ZdS_PgM_ZnS
	F_ZdD_PgM_ZnD
	F_ZdnB_PgM_ZdnB_ZmB
	F_ZdnH_PgM_ZdnH_ZmH
	F_ZdnS_PgM_ZdnS_ZmS
	F_ZdnD_PgM_ZdnD_ZmD
	F_Xdn_PmB
	F_Xdn_PmH
	F_Xdn_PmS
	F_Xdn_PmD
	F_Zt_AddrXSP
	F_Zt_AddrXSPImmMulVl
	F_Pt_AddrXSP
	F_Pt_AddrXSPImmMulVl
	F_Dd_Pg_ZnB
	F_Dd_Pg_ZnH
	F_Dd_Pg_ZnS
	F_PdH_PgZ_ZnH_ImmFP
	F_PdS_PgZ_ZnS_ImmFP
	F_PdD_PgZ_ZnD_ImmFP
	F_PdB_PgZ_ZnB_Imm
	F_PdH_PgZ_ZnH_Imm
	F_PdS_PgZ_ZnS_Imm
	F_PdD_PgZ_ZnD_Imm
	F_PdB_PgZ_ZnB_ZmB
	F_PdH_PgZ_ZnH_ZmH
	F_PdS_PgZ_ZnS_ZmS
	F_PdD_PgZ_ZnD_ZmD
	F_PdB_PgZ_ZnB_ZmD
	F_PdH_PgZ_ZnH_ZmD
	F_PdS_PgZ_ZnS_ZmD
)

// Format groups, common patterns of associated instruction formats. E.g. expansion of the <T> generic lane size.
var FG_ZdT_PgM_ZnT = []int{F_ZdB_PgM_ZnB, F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdT_PgM_ZnT_FP = []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdnT_PgM_ZdnT_ZmT = []int{F_ZdnB_PgM_ZdnB_ZmB, F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}
var FG_ZdnT_PgM_ZdnT_ZmT_FP = []int{F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}
var FG_Xdn_PmT = []int{F_Xdn_PmB, F_Xdn_PmH, F_Xdn_PmS, F_Xdn_PmD}
var FG_PdT_PgZ_ZnT_ImmFP = []int{F_PdH_PgZ_ZnH_ImmFP, F_PdS_PgZ_ZnS_ImmFP, F_PdD_PgZ_ZnD_ImmFP}
var FG_PdT_PgZ_ZnT_Imm = []int{F_PdB_PgZ_ZnB_Imm, F_PdH_PgZ_ZnH_Imm, F_PdS_PgZ_ZnS_Imm, F_PdD_PgZ_ZnD_Imm}
var FG_PdT_PgZ_ZnT_ZmT = []int{F_PdB_PgZ_ZnB_ZmB, F_PdH_PgZ_ZnH_ZmH, F_PdS_PgZ_ZnS_ZmS, F_PdD_PgZ_ZnD_ZmD}
var FG_PdT_PgZ_ZnT_ZmT_FP = []int{F_PdH_PgZ_ZnH_ZmH, F_PdS_PgZ_ZnS_ZmS, F_PdD_PgZ_ZnD_ZmD}
var FG_PdT_PgZ_ZnT_ZmD = []int{F_PdB_PgZ_ZnB_ZmD, F_PdH_PgZ_ZnH_ZmD, F_PdS_PgZ_ZnS_ZmD}

// The format table holds a representation of the operand syntax for an instruction.
var formats = map[int]format{
	F_ZdaS_ZnH_ZmH:       []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Zd>.S, <Zn>.H, <Zm>.H
	F_ZdaS_ZnH_ZmHidx:    []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                // <Zd>.S, <Zn>.H, <Zm>.H[<imm>]
	F_ZdB_PgM_ZnB:        []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B},                  // <Zd>.B, <Pg>/M, <Zn>.B
	F_ZdH_PgM_ZnH:        []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H},                  // <Zd>.H, <Pg>/M, <Zn>.H
	F_ZdS_PgM_ZnS:        []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S},                  // <Zd>.S, <Pg>/M, <Zn>.S
	F_ZdD_PgM_ZnD:        []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D},                  // <Zd>.D, <Pg>/M, <Zn>.D
	F_ZdnB_PgM_ZdnB_ZmB:  []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, REG_Z | EXT_B},   // <Zdn>.B, <Pg>/M, <Zdn>.B, <Zm>.B
	F_ZdnH_PgM_ZdnH_ZmH:  []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H},   // <Zdn>.H, <Pg>/M, <Zdn>.H, <Zm>.H
	F_ZdnS_PgM_ZdnS_ZmS:  []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S},   // <Zdn>.S, <Pg>/M, <Zdn>.S, <Zm>.S
	F_ZdnD_PgM_ZdnD_ZmD:  []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D},   // <Zdn>.D, <Pg>/M, <Zdn>.D, <Zm>.D
	F_Xdn_PmB:            []int{REG_R, REG_P | EXT_B},                                               // <Xdn>, <Pm>.B
	F_Xdn_PmH:            []int{REG_R, REG_P | EXT_H},                                               // <Xdn>, <Pm>.H
	F_Xdn_PmS:            []int{REG_R, REG_P | EXT_S},                                               // <Xdn>, <Pm>.S
	F_Xdn_PmD:            []int{REG_R, REG_P | EXT_D},                                               // <Xdn>, <Pm>.D
	F_Zt_AddrXSP:         []int{REG_Z, MEM_ADDR | MEM_BASE},                                         // <Zt>, [<Xn|SP>]
	F_Zt_AddrXSPImmMulVl: []int{REG_Z, MEM_ADDR | MEM_OFFSET_IMM},                                   // <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_Pt_AddrXSP:         []int{REG_P, MEM_ADDR | MEM_BASE},                                         // <Zt>, [<Xn|SP>]
	F_Pt_AddrXSPImmMulVl: []int{REG_P, MEM_ADDR | MEM_OFFSET_IMM},                                   // <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_Dd_Pg_ZnB:          []int{REG_F, REG_P, REG_Z | EXT_B},                                        // <Dd>, <Pg>, <Zn>.B
	F_Dd_Pg_ZnH:          []int{REG_F, REG_P, REG_Z | EXT_H},                                        // <Dd>, <Pg>, <Zn>.H
	F_Dd_Pg_ZnS:          []int{REG_F, REG_P, REG_Z | EXT_S},                                        // <Dd>, <Pg>, <Zn>.S
	F_PdH_PgZ_ZnH_ImmFP:  []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, IMM | IMM_FLOAT}, // <Pd>.H, <Pg>/Z, <Zn>.H
	F_PdS_PgZ_ZnS_ImmFP:  []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, IMM | IMM_FLOAT}, // <Pd>.S, <Pg>/Z, <Zn>.S
	F_PdD_PgZ_ZnD_ImmFP:  []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, IMM | IMM_FLOAT}, // <Pd>.D, <Pg>/Z, <Zn>.D
	F_PdB_PgZ_ZnB_Imm:    []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, IMM | IMM_INT},   // <Pd>.B, <Pg>/Z, <Zn>.B, #<imm>
	F_PdH_PgZ_ZnH_Imm:    []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, IMM | IMM_INT},   // <Pd>.H, <Pg>/Z, <Zn>.H, #<imm>
	F_PdS_PgZ_ZnS_Imm:    []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, IMM | IMM_INT},   // <Pd>.S, <Pg>/Z, <Zn>.S, #<imm>
	F_PdD_PgZ_ZnD_Imm:    []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, IMM | IMM_INT},   // <Pd>.D, <Pg>/Z, <Zn>.D, #<imm>
	F_PdB_PgZ_ZnB_ZmB:    []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, REG_Z | EXT_B},   // <Pd>.B, <Pg>/Z, <Zn>.B, <Zm>.B
	F_PdH_PgZ_ZnH_ZmH:    []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, REG_Z | EXT_H},   // <Pd>.H, <Pg>/Z, <Zn>.H, <Zm>.H
	F_PdS_PgZ_ZnS_ZmS:    []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, REG_Z | EXT_S},   // <Pd>.S, <Pg>/Z, <Zn>.S, <Zm>.S
	F_PdD_PgZ_ZnD_ZmD:    []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, REG_Z | EXT_D},   // <Pd>.D, <Pg>/Z, <Zn>.D, <Zm>.D
	F_PdB_PgZ_ZnB_ZmD:    []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, REG_Z | EXT_D},   // <Pd>.B, <Pg>/Z, <Zn>.B, <Zm>.D
	F_PdH_PgZ_ZnH_ZmD:    []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, REG_Z | EXT_D},   // <Pd>.H, <Pg>/Z, <Zn>.H, <Zm>.D
	F_PdS_PgZ_ZnS_ZmD:    []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, REG_Z | EXT_D},   // <Pd>.S, <Pg>/Z, <Zn>.S, <Zm>.D
}

// Key into the encoder table.
const (
	E_Rd_Rn_Rm = iota
	E_Zda_Zn_Zm_i2
	E_size_Pg_Zn_Zd
	E_size_Pg_Zm_Zdn
	E_size_Pm_Rdn
	E_imm9h_imm9l_Rn_Zt
	E_size_Pg_Zn_Vd
	E_size_Pg_Zn_Pd_ImmFP0
	E_size_simm5_Pg_Zn_Pd
	E_size_uimm7_Pg_Zn_Pd
	E_size_Zm_Pg_Zn_Pd
	E_size_ZmD_Pg_Zn_Pd

	// Equivalences
	E_size0_Pg_Zn_Zd = E_size_Pg_Zn_Zd
)

// The encoder table holds a list of encoding schemes for operands. Each scheme contains
// a list of rules and a list of indices that mark which operands need to be fed into the
// rule. Each rule produces a 32-bit number which should be OR'd with the base to create
// an instruction encoding.
var encoders = map[int]encoder{
	E_Rd_Rn_Rm:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_Zda_Zn_Zm_i2:         {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_size_Pg_Zn_Zd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}}},
	E_size_Pg_Zm_Zdn:       {[]rule{{[]int{0, 2}, Zdn}, {[]int{1}, Pg}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
	E_size_Pm_Rdn:          {[]rule{{[]int{0}, Rd}, {[]int{1}, Pm}, {[]int{1}, sveT}}},
	E_imm9h_imm9l_Rn_Zt:    {[]rule{{[]int{0}, Rt}, {[]int{1}, RnImm9MulVl}}},
	E_size_Pg_Zn_Vd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{2}, sveT}}},
	E_size_Pg_Zn_Pd_ImmFP0: {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, ImmFP0}}},
	E_size_simm5_Pg_Zn_Pd:  {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Imm5}}},
	E_size_uimm7_Pg_Zn_Pd:  {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Uimm7}}},
	E_size_Zm_Pg_Zn_Pd:     {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_ZmD_Pg_Zn_Pd:    {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2}, sveT}}},
}
