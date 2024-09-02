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
		{0x04200000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},        // ADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZAND: {
		{0x041a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // AND <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04203000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // AND <Zd>.D, <Zn>.D, <Zm>.D
	},
	APANDS: {
		{0x25404000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ANDS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZANDV: {
		{0x041a2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // ANDV <V><d>, <Pg>, <Zn>.<T>
	},
	AZASR: {
		{0x04108000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZASRR: {
		{0x04148000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ASRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZBFCVT: {
		{0x658aa000, []int{F_ZdH_PgM_ZnS}, E_Pg_Zn_Zd}, // BFCVT <Zd>.H, <Pg>/M, <Zn>.S
	},
	AZBFCVTNT: {
		{0x648aa000, []int{F_ZdH_PgM_ZnS}, E_Pg_Zn_Zd}, // BFCVTNT <Zd>.H, <Pg>/M, <Zn>.S
	},
	AZBFDOT: {
		{0x64608000, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFDOT <Zda>.S, <Zn>.H, <Zm>.H
		{0x64604000, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i2}, // BFDOT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
	AZBFMLALB: {
		{0x64e08000, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFMLALB <Zda>.S, <Zn>.H, <Zm>.H
		{0x64e04000, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i3}, // BFMLALB <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
	AZBFMLALT: {
		{0x64e08400, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm},        // BFMLALT <Zda>.S, <Zn>.H, <Zm>.H
		{0x64e04400, []int{F_ZdaS_ZnH_ZmHidx}, E_Zda_Zn_Zm_i3}, // BFMLALT <Zda>.S, <Zn>.H, <Zm>.H[<imm>]
	},
	AZBFMMLA: {
		{0x6460e400, []int{F_ZdaS_ZnH_ZmH}, E_Rd_Rn_Rm}, // BFMMLA <Zda>.S, <Zn>.H, <Zm>.H
	},
	APBIC: {
		{0x25004010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BIC <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZBIC: {
		{0x041b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // BIC <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04e03000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // BIC <Zd>.D, <Zn>.D, <Zm>.D
	},
	APBICS: {
		{0x25404010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BICS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	ABRKA: {
		{0x25104000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKA <Pd>.B, <Pg>/Z, <Pn>.B
		{0x25104010, []int{F_PdB_PgM_PnB}, E_Pg_Pn_Pd}, // BRKA <Pd>.B, <Pg>/Z, <Pn>.B
	},
	ABRKAS: {
		{0x25504000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKAS <Pd>.B, <Pg>/Z, <Pn>.B
	},
	ABRKB: {
		{0x25904000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKB <Pd>.B, <Pg>/Z, <Pn>.B
		{0x25904010, []int{F_PdB_PgM_PnB}, E_Pg_Pn_Pd}, // BRKB <Pd>.B, <Pg>/Z, <Pn>.B
	},
	ABRKBS: {
		{0x25d04000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKBS <Pd>.B, <Pg>/Z, <Pn>.B
	},
	ABRKN: {
		{0x25184000, []int{F_PdB_PgZ_PnB_PmB}, E_Pg_Pn_Pdm}, // BRKN <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
	},
	ABRKNS: {
		{0x25584000, []int{F_PdB_PgZ_PnB_PmB}, E_Pg_Pn_Pdm}, // BRKNS <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
	},
	ABRKPA: {
		{0x2500c000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPA <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	ABRKPAS: {
		{0x2540c000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPAS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	ABRKPB: {
		{0x2500c010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPB <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	ABRKPBS: {
		{0x2540c010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPBS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZCLASTA: {
		{0x0530a000, FG_Rdn_Pg_Rdn_ZmT, E_size_Pg_Zm_Rdn}, // CLASTA <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
		{0x052a8000, FG_Vdn_Pg_Vdn_ZmT, E_size_Pg_Zm_Rdn}, // CLASTA <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
	},
	AZCLASTB: {
		{0x0531a000, FG_Rdn_Pg_Rdn_ZmT, E_size_Pg_Zm_Rdn}, // CLASTB <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
		{0x052b8000, FG_Vdn_Pg_Vdn_ZmT, E_size_Pg_Zm_Rdn}, // CLASTB <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
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
	ACTERMEQ: {
		{0x25e02000, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	ACTERMEQW: {
		{0x25a02000, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	ACTERMNE: {
		{0x25e02010, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	ACTERMNEW: {
		{0x25a02010, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	AZDECP: {
		{0x252d8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // DECP <Xdn>, <Pm>.<T>
	},
	AZEOR: {
		{0x04190000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // EOR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04a03000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // EOR <Zd>.D, <Zn>.D, <Zm>.D
	},
	APEOR: {
		{0x25004200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // EOR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APEORS: {
		{0x25404200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // EORS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZEORV: {
		{0x04192000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // EORV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFABD: {
		{0x65088000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABS: {
		{0x041ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFADD: {
		{0x65008000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000000, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},        // FADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFADDA: {
		{0x65182000, []int{F_Fdn_Pg_Fdn_ZmH, F_Fdn_Pg_Fdn_ZmS, F_Fdn_Pg_Fdn_ZmD}, E_size_Pg_Zm_Rdn}, // FADDA <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
	},
	AZFADDV: {
		{0x65002000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FADDV <V><d>, <Pg>, <Zn>.<T>
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
	AZFCVT: {
		{0x65c9a000, []int{F_ZdD_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVT <Zd>.D, <Pg>/M, <Zn>.H
		{0x65cba000, []int{F_ZdD_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVT <Zd>.D, <Pg>/M, <Zn>.S
		{0x65c8a000, []int{F_ZdH_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVT <Zd>.H, <Pg>/M, <Zn>.D
		{0x65caa000, []int{F_ZdS_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVT <Zd>.S, <Pg>/M, <Zn>.D
		{0x6589a000, []int{F_ZdS_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVT <Zd>.S, <Pg>/M, <Zn>.H
		{0x6588a000, []int{F_ZdH_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVT <Zd>.H, <Pg>/M, <Zn>.S
	},
	AZFCVTZS: {
		{0x65dea000, []int{F_ZdD_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.D, <Pg>/M, <Zn>.D
		{0x655ea000, []int{F_ZdD_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.D, <Pg>/M, <Zn>.H
		{0x65dca000, []int{F_ZdD_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.D, <Pg>/M, <Zn>.S
		{0x655aa000, []int{F_ZdH_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.H, <Pg>/M, <Zn>.H
		{0x65d8a000, []int{F_ZdS_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.S, <Pg>/M, <Zn>.D
		{0x655ca000, []int{F_ZdS_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.S, <Pg>/M, <Zn>.H
		{0x659ca000, []int{F_ZdS_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVTZS <Zd>.S, <Pg>/M, <Zn>.S
	},
	AZFCVTZU: {
		{0x65dfa000, []int{F_ZdD_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.D, <Pg>/M, <Zn>.D
		{0x655fa000, []int{F_ZdD_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.D, <Pg>/M, <Zn>.H
		{0x65dda000, []int{F_ZdD_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.D, <Pg>/M, <Zn>.S
		{0x655ba000, []int{F_ZdH_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.H, <Pg>/M, <Zn>.H
		{0x65d9a000, []int{F_ZdS_PgM_ZnD}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.S, <Pg>/M, <Zn>.D
		{0x655da000, []int{F_ZdS_PgM_ZnH}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.S, <Pg>/M, <Zn>.H
		{0x659da000, []int{F_ZdS_PgM_ZnS}, E_Pg_Zn_Zd}, // FCVTZU <Zd>.S, <Pg>/M, <Zn>.S
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
	AZFMAXNMV: {
		{0x65042000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMAXNMV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMAXV: {
		{0x65062000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMIN: {
		{0x65078000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMINNM: {
		{0x65058000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMINNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFMINNMV: {
		{0x65052000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMINNMV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMINV: {
		{0x65072000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMUL: {
		{0x65028000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000800, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},        // FMUL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFMULX: {
		{0x650a8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMULX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFNEG: {
		{0x041da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FNEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFRECPS: {
		{0x65001800, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FRECPS <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
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
	AZFRSQRTS: {
		{0x65001c00, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FRSQRTS <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFSCALE: {
		{0x65098000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSCALE <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSUB: {
		{0x65018000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000400, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},        // FSUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFSUBR: {
		{0x65038000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSQRT: {
		{0x650da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FSQRT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFTSMUL: {
		{0x65000c00, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FTSMUL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFTSSEL: {
		{0x0420b000, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FTSSEL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZINCP: {
		{0x252c8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // INCP <Xdn>, <Pm>.<T>
	},
	AZLASTA: {
		{0x0520a000, FG_Rd_Pg_ZnT, E_size_Pg_Zn_Rd}, // LASTA <R><d>, <Pg>, <Zn>.<T>
		{0x05228000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // LASTA <V><d>, <Pg>, <Zn>.<T>
	},
	AZLASTB: {
		{0x0521a000, FG_Rd_Pg_ZnT, E_size_Pg_Zn_Rd}, // LASTA <R><d>, <Pg>, <Zn>.<T>
		{0x05238000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // LASTB <V><d>, <Pg>, <Zn>.<T>
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
	APNAND: {
		{0x25804210, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NAND <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APNANDS: {
		{0x25c04210, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NANDS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZNEG: {
		{0x0417a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	APNOR: {
		{0x25804200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NOR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APNORS: {
		{0x25c04200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NORS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZNOT: {
		{0x041ea000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AORN: {
		{0x25804010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORN <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AORNS: {
		{0x25c04010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORNS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZORR: {
		{0x04180000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ORR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04603000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // ORR <Zd>.D, <Zn>.D, <Zm>.D
	},
	APORR: {
		{0x25804000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APORRS: {
		{0x25c04000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORRS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AZORV: {
		{0x04182000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // ORV <V><d>, <Pg>, <Zn>.<T>
	},
	APFALSE: {
		{0x2518e400, []int{F_PdB}, E_Pd}, // PFALSE <Pd>.B
	},
	APFIRST: {
		{0x2558c000, []int{F_PdnB_Pg_PdnB}, E_Pg_Pdn}, // PFIRST <Pdn>.B, <Pg>, <Pdn>.B
	},
	APNEXT: {
		{0x2519c400, FG_PdnT_Pv_PdnT, E_size_Pv_Pdn}, // PNEXT <Pdn>.<T>, <Pv>, <Pdn>.<T>
	},
	APTEST: {
		{0x2550c000, []int{F_Pg_PnB}, E_Pg_Pn}, // PTEST <Pg>, <Pn>.B
	},
	APUNPKHI: {
		{0x05314000, []int{F_PdH_PnB}, E_Pn_Pd},
	},
	APUNPKLO: {
		{0x05304000, []int{F_PdH_PnB}, E_Pn_Pd},
	},
	AZRBIT: {
		{0x05278000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // RBIT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	ARDFFR: {
		{0x2519f000, []int{F_PdB}, E_Pd},            // RDFFR <Pd>.B
		{0x2518f000, []int{F_PdB_PgZ}, E_Pg_8_5_Pd}, // RDFFR <Pd>.B, <Pg>/Z
	},
	ARDFFRS: {
		{0x2558f000, []int{F_PdB_PgZ}, E_Pg_8_5_Pd}, // RDFFRS <Pd>.B, <Pg>/Z
	},
	AZREV: {
		{0x05344000, FG_PdT_PnT, E_size_Pn_Pd}, // REV <Pd>.<T>, <Pn>.<T>
	},
	AZREVB: {
		{0x05248000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // REVB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVH: {
		{0x05258000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // REVH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZREVW: {
		{0x05268000, []int{F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // REVW <Zd>.D, <Pg>/M, <Zn>.D
	},
	AZSABD: {
		{0x040c0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSADDV: {
		{0x04002000, []int{F_Dd_Pg_ZnB, F_Dd_Pg_ZnH, F_Dd_Pg_ZnS}, E_size_Pg_Zn_Vd}, // SADDV <Dd>, <Pg>, <Zn>.<T>
	},
	AZSCVTF: {
		{0x65d6a000, []int{F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // SCVTF <Zd>.D, <Pg>/M, <Zn>.D
		{0x65d0a000, []int{F_ZdD_PgM_ZnS}, E_Pg_Zn_Zd},      // SCVTF <Zd>.D, <Pg>/M, <Zn>.S
		{0x6552a000, []int{F_ZdH_PgM_ZnH}, E_Pg_Zn_Zd},      // SCVTF <Zd>.H, <Pg>/M, <Zn>.H
		{0x6556a000, []int{F_ZdH_PgM_ZnD}, E_Pg_Zn_Zd},      // SCVTF <Zd>.H, <Pg>/M, <Zn>.D
		{0x65d4a000, []int{F_ZdS_PgM_ZnD}, E_Pg_Zn_Zd},      // SCVTF <Zd>.S, <Pg>/M, <Zn>.D
		{0x6594a000, []int{F_ZdS_PgM_ZnS}, E_Pg_Zn_Zd},      // SCVTF <Zd>.S, <Pg>/M, <Zn>.S
		{0x6554a000, []int{F_ZdH_PgM_ZnS}, E_Pg_Zn_Zd},      // SCVTF <Zd>.H, <Pg>/M, <Zn>.S
	},
	AZSDIV: {
		{0x04140000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIV <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSDIVR: {
		{0x04160000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SDIVR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	APSEL: {
		{0x25004210, []int{F_PdB_Pg_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // SEL <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
	},
	ASETFFR: {
		{0x252c9000, FG_none, E_none}, // SETFFR
	},
	AZSMAX: {
		{0x04080000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMAXV: {
		{0x04082000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // SMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZSMIN: {
		{0x040a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSMINV: {
		{0x040a2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // SMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZSMULH: {
		{0x04120000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSQADD: {
		{0x04201000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd}, // SQADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZSQSUB: {
		{0x04201800, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd}, // SQSUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZSTR: {
		{0xe5804000, []int{F_Zt_AddrXSP, F_Zt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe5800000, []int{F_Pt_AddrXSP, F_Pt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Pt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSUB: {
		{0x04010000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04200400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},        // SUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
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
	AZSXTW: {
		{0x0414a000, []int{F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // SXTW <Zd>.D, <Pg>/M, <Zn>.D
	},
	APTRN1: {
		{0x05205000, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // TRN1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZTRN1: {
		{0x05207000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // TRN1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a01800, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // TRN1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	APTRN2: {
		{0x05205400, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // TRN2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZTRN2: {
		{0x05207400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // TRN2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a01c00, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // TRN2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZUABD: {
		{0x040d0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUCVTF: {
		{0x65d7a000, []int{F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // UCVTF <Zd>.D, <Pg>/M, <Zn>.D
		{0x65d1a000, []int{F_ZdD_PgM_ZnS}, E_Pg_Zn_Zd},      // UCVTF <Zd>.D, <Pg>/M, <Zn>.S
		{0x6553a000, []int{F_ZdH_PgM_ZnH}, E_Pg_Zn_Zd},      // UCVTF <Zd>.H, <Pg>/M, <Zn>.H
		{0x6557a000, []int{F_ZdH_PgM_ZnD}, E_Pg_Zn_Zd},      // UCVTF <Zd>.H, <Pg>/M, <Zn>.D
		{0x65d5a000, []int{F_ZdS_PgM_ZnD}, E_Pg_Zn_Zd},      // UCVTF <Zd>.S, <Pg>/M, <Zn>.D
		{0x6595a000, []int{F_ZdS_PgM_ZnS}, E_Pg_Zn_Zd},      // UCVTF <Zd>.S, <Pg>/M, <Zn>.S
		{0x6555a000, []int{F_ZdH_PgM_ZnS}, E_Pg_Zn_Zd},      // UCVTF <Zd>.H, <Pg>/M, <Zn>.S
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
	AZUMAXV: {
		{0x04092000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // UMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZUMIN: {
		{0x040b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUMINV: {
		{0x040b2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // UMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZUMULH: {
		{0x04130000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUQADD: {
		{0x04201400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd}, // UQADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZUQSUB: {
		{0x04201c00, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd}, // UQADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZUXTB: {
		{0x0411a000, []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // UXTB <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUXTH: {
		{0x0413a000, []int{F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}, E_size0_Pg_Zn_Zd}, // UXTH <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZUXTW: {
		{0x0415a000, []int{F_ZdD_PgM_ZnD}, E_size_Pg_Zn_Zd}, // UXTW <Zd>.D, <Pg>/M, <Zn>.D
	},
	APUZP1: {
		{0x05204800, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // UZP1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZUZP1: {
		{0x05206800, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // UZP1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00800, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // UZP1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	APUZP2: {
		{0x05204c00, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // UZP2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZUZP2: {
		{0x05206c00, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // UZP2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00c00, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // UZP2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AWHILELE: {
		{0x25201410, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELE <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELEW: {
		{0x25200410, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELE <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELO: {
		{0x25201c00, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELOW: {
		{0x25200c00, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELS: {
		{0x25201c10, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELSW: {
		{0x25200c10, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELT: {
		{0x25201400, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWHILELTW: {
		{0x25200400, FG_PdT_Rn_Rm, E_size_Rm_Rn_Pd}, // WHILELO <Pd>.<T>, <R><n>, <R><m>
	},
	AWRFFR: {
		{0x25289000, []int{F_PnB}, E_Pn}, // WRFFR <Pn>.B
	},
	APZIP1: {
		{0x05204000, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // ZIP1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZZIP1: {
		{0x05206000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // ZIP1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00000, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // ZIP1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	APZIP2: {
		{0x05204400, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // ZIP2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	AZZIP2: {
		{0x05206400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // ZIP2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00400, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // ZIP2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
}

// Key into the format table.
const (
	F_none = iota
	F_ZdaS_ZnH_ZmH
	F_ZdaS_ZnH_ZmHidx
	F_ZdB_PgM_ZnB
	F_ZdH_PgM_ZnH
	F_ZdS_PgM_ZnS
	F_ZdD_PgM_ZnD
	F_ZdD_PgM_ZnH
	F_ZdD_PgM_ZnS
	F_ZdH_PgM_ZnD
	F_ZdS_PgM_ZnD
	F_ZdS_PgM_ZnH
	F_ZdH_PgM_ZnS
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
	F_PdB_PnB
	F_PdH_PnH
	F_PdS_PnS
	F_PdD_PnD
	F_PdB_PnB_PmB
	F_PdH_PnH_PmH
	F_PdS_PnS_PmS
	F_PdD_PnD_PmD
	F_PdB_Rn_Rm
	F_PdH_Rn_Rm
	F_PdS_Rn_Rm
	F_PdD_Rn_Rm
	F_PdB
	F_PdB_PgZ_PnB_PmB
	F_PdB_Pg_PnB_PmB
	F_PdB_PgZ_PnB
	F_PdB_PgM_PnB
	F_PdB_PgZ
	F_PdH_PnB
	F_PdnB_Pv_PdnB
	F_PdnH_Pv_PdnH
	F_PdnS_Pv_PdnS
	F_PdnD_Pv_PdnD
	F_PdnB_Pg_PdnB
	F_Pg_PnB
	F_PnB
	F_Rd_Pg_ZnB
	F_Rd_Pg_ZnH
	F_Rd_Pg_ZnS
	F_Rd_Pg_ZnD
	F_Rdn_Pg_Rdn_ZmB
	F_Rdn_Pg_Rdn_ZmH
	F_Rdn_Pg_Rdn_ZmS
	F_Rdn_Pg_Rdn_ZmD
	F_Rn_Rm
	F_Vd_Pg_ZnB
	F_Vd_Pg_ZnH
	F_Vd_Pg_ZnS
	F_Vd_Pg_ZnD
	F_Fd_Pg_ZnH
	F_Fd_Pg_ZnS
	F_Fd_Pg_ZnD
	F_ZdB_ZnB_ZmB
	F_ZdH_ZnH_ZmH
	F_ZdS_ZnS_ZmS
	F_ZdD_ZnD_ZmD
	F_ZdQ_ZnQ_ZmQ
	F_Vdn_Pg_Vdn_ZmB
	F_Vdn_Pg_Vdn_ZmH
	F_Vdn_Pg_Vdn_ZmS
	F_Vdn_Pg_Vdn_ZmD
	F_Fdn_Pg_Fdn_ZmH
	F_Fdn_Pg_Fdn_ZmS
	F_Fdn_Pg_Fdn_ZmD
)

// Format groups, common patterns of associated instruction formats. E.g. expansion of the <T> generic lane size.
var FG_none = []int{F_none} // No arguments.
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
var FG_PdT_PnT = []int{F_PdB_PnB, F_PdH_PnH, F_PdS_PnS, F_PdD_PnD}
var FG_PdT_PnT_PmT = []int{F_PdB_PnB_PmB, F_PdH_PnH_PmH, F_PdS_PnS_PmS, F_PdD_PnD_PmD}
var FG_PdT_Rn_Rm = []int{F_PdB_Rn_Rm, F_PdH_Rn_Rm, F_PdS_Rn_Rm, F_PdD_Rn_Rm}
var FG_PdnT_Pv_PdnT = []int{F_PdnB_Pv_PdnB, F_PdnH_Pv_PdnH, F_PdnS_Pv_PdnS, F_PdnD_Pv_PdnD}
var FG_Rd_Pg_ZnT = []int{F_Rd_Pg_ZnB, F_Rd_Pg_ZnH, F_Rd_Pg_ZnS, F_Rd_Pg_ZnD}
var FG_Rdn_Pg_Rdn_ZmT = []int{F_Rdn_Pg_Rdn_ZmB, F_Rdn_Pg_Rdn_ZmH, F_Rdn_Pg_Rdn_ZmS, F_Rdn_Pg_Rdn_ZmD}
var FG_Vd_Pg_ZnT = []int{F_Vd_Pg_ZnB, F_Vd_Pg_ZnH, F_Vd_Pg_ZnS, F_Vd_Pg_ZnD}
var FG_Vd_Pg_ZnT_FP = []int{F_Fd_Pg_ZnH, F_Fd_Pg_ZnS, F_Fd_Pg_ZnD}
var FG_ZdT_ZnT_ZmT = []int{F_ZdB_ZnB_ZmB, F_ZdH_ZnH_ZmH, F_ZdS_ZnS_ZmS, F_ZdD_ZnD_ZmD}
var FG_ZdT_ZnT_ZmT_FP = []int{F_ZdH_ZnH_ZmH, F_ZdS_ZnS_ZmS, F_ZdD_ZnD_ZmD}
var FG_Vdn_Pg_Vdn_ZmT = []int{F_Vdn_Pg_Vdn_ZmB, F_Vdn_Pg_Vdn_ZmH, F_Vdn_Pg_Vdn_ZmS, F_Vdn_Pg_Vdn_ZmD}

// The format table holds a representation of the operand syntax for an instruction.
var formats = map[int]format{
	F_none:               []int{},                                                                   // No arguments.
	F_ZdaS_ZnH_ZmH:       []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Zd>.S, <Zn>.H, <Zm>.H
	F_ZdaS_ZnH_ZmHidx:    []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                // <Zd>.S, <Zn>.H, <Zm>.H[<imm>]
	F_ZdB_PgM_ZnB:        []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B},                  // <Zd>.B, <Pg>/M, <Zn>.B
	F_ZdH_PgM_ZnH:        []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H},                  // <Zd>.H, <Pg>/M, <Zn>.H
	F_ZdS_PgM_ZnS:        []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S},                  // <Zd>.S, <Pg>/M, <Zn>.S
	F_ZdD_PgM_ZnD:        []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D},                  // <Zd>.D, <Pg>/M, <Zn>.D
	F_ZdD_PgM_ZnH:        []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_H},                  // <Zd>.D, <Pg>/M, <Zn>.H
	F_ZdD_PgM_ZnS:        []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_S},                  // <Zd>.D, <Pg>/M, <Zn>.S
	F_ZdH_PgM_ZnD:        []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_D},                  // <Zd>.H, <Pg>/M, <Zn>.D
	F_ZdS_PgM_ZnD:        []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_D},                  // <Zd>.S, <Pg>/M, <Zn>.D
	F_ZdS_PgM_ZnH:        []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_H},                  // <Zd>.S, <Pg>/M, <Zn>.H
	F_ZdH_PgM_ZnS:        []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_S},                  // <Zd>.S, <Pg>/M, <Zn>.H
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
	F_PdB_PnB:            []int{REG_P | EXT_B, REG_P | EXT_B},                                       // <Pd>.B, <Pn>.B
	F_PdH_PnH:            []int{REG_P | EXT_H, REG_P | EXT_H},                                       // <Pd>.H, <Pn>.H
	F_PdS_PnS:            []int{REG_P | EXT_S, REG_P | EXT_S},                                       // <Pd>.S, <Pn>.S
	F_PdD_PnD:            []int{REG_P | EXT_D, REG_P | EXT_D},                                       // <Pd>.D, <Pn>.D
	F_PdB_PnB_PmB:        []int{REG_P | EXT_B, REG_P | EXT_B, REG_P | EXT_B},                        // <Pd>.B, <Pn>.B, <Pm>.B
	F_PdH_PnH_PmH:        []int{REG_P | EXT_H, REG_P | EXT_H, REG_P | EXT_H},                        // <Pd>.H, <Pn>.H, <Pm>.H
	F_PdS_PnS_PmS:        []int{REG_P | EXT_S, REG_P | EXT_S, REG_P | EXT_S},                        // <Pd>.S, <Pn>.S, <Pm>.S
	F_PdD_PnD_PmD:        []int{REG_P | EXT_D, REG_P | EXT_D, REG_P | EXT_D},                        // <Pd>.D, <Pn>.D, <Pm>.D
	F_PdB_Rn_Rm:          []int{REG_P | EXT_B, REG_R, REG_R},                                        // <Pd>.B, <R><n>, <R><m>
	F_PdH_Rn_Rm:          []int{REG_P | EXT_H, REG_R, REG_R},                                        // <Pd>.B, <R><n>, <R><m>
	F_PdS_Rn_Rm:          []int{REG_P | EXT_S, REG_R, REG_R},                                        // <Pd>.B, <R><n>, <R><m>
	F_PdD_Rn_Rm:          []int{REG_P | EXT_D, REG_R, REG_R},                                        // <Pd>.B, <R><n>, <R><m>
	F_PdB:                []int{REG_P | EXT_B},                                                      // <Pd>.B
	F_PdB_PgZ_PnB_PmB:    []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_P | EXT_B, REG_P | EXT_B},   // <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	F_PdB_Pg_PnB_PmB:     []int{REG_P | EXT_B, REG_P, REG_P | EXT_B, REG_P | EXT_B},                 // <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
	F_PdB_PgZ_PnB:        []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_P | EXT_B},                  // <Pd>.B, <Pg>/Z, <Pn>.B
	F_PdB_PgM_PnB:        []int{REG_P | EXT_B, REG_P | EXT_MERGING, REG_P | EXT_B},                  // <Pd>.B, <Pg>/Z, <Pn>.B
	F_PdB_PgZ:            []int{REG_P | EXT_B, REG_P | EXT_ZEROING},                                 // <Pd>.B, <Pg>/Z
	F_PdH_PnB:            []int{REG_P | EXT_H, REG_P | EXT_B},                                       // <Pd>.H, <Pn>.B
	F_PdnB_Pv_PdnB:       []int{REG_P | EXT_B, REG_P, REG_P | EXT_B},                                // <Pdn>.B, <Pv>, <Pdn>.B
	F_PdnH_Pv_PdnH:       []int{REG_P | EXT_H, REG_P, REG_P | EXT_H},                                // <Pdn>.H, <Pv>, <Pdn>.H
	F_PdnS_Pv_PdnS:       []int{REG_P | EXT_S, REG_P, REG_P | EXT_S},                                // <Pdn>.S, <Pv>, <Pdn>.S
	F_PdnD_Pv_PdnD:       []int{REG_P | EXT_D, REG_P, REG_P | EXT_D},                                // <Pdn>.D, <Pv>, <Pdn>.D
	F_PdnB_Pg_PdnB:       []int{REG_P | EXT_B, REG_P, REG_P | EXT_B},                                // <Pdn>.B, <Pg>, <Pdn>.B
	F_Pg_PnB:             []int{REG_P, REG_P | EXT_B},                                               // <Pg>, <Pn>.B
	F_PnB:                []int{REG_P | EXT_B},                                                      // <Pn>.B
	F_Rd_Pg_ZnB:          []int{REG_R, REG_P, REG_Z | EXT_B},                                        // <Wd>, <Pg>, <Zn>.B
	F_Rd_Pg_ZnH:          []int{REG_R, REG_P, REG_Z | EXT_H},                                        // <Wd>, <Pg>, <Zn>.H
	F_Rd_Pg_ZnS:          []int{REG_R, REG_P, REG_Z | EXT_S},                                        // <Wd>, <Pg>, <Zn>.S
	F_Rd_Pg_ZnD:          []int{REG_R, REG_P, REG_Z | EXT_D},                                        // <Xd>, <Pg>, <Zn>.D
	F_Rdn_Pg_Rdn_ZmB:     []int{REG_R, REG_P, REG_R, REG_Z | EXT_B},                                 // <Wdn>, <Pg>, <Wdn>, <Zm>.B
	F_Rdn_Pg_Rdn_ZmH:     []int{REG_R, REG_P, REG_R, REG_Z | EXT_H},                                 // <Wdn>, <Pg>, <Wdn>, <Zm>.H
	F_Rdn_Pg_Rdn_ZmS:     []int{REG_R, REG_P, REG_R, REG_Z | EXT_S},                                 // <Wdn>, <Pg>, <Wdn>, <Zm>.S
	F_Rdn_Pg_Rdn_ZmD:     []int{REG_R, REG_P, REG_R, REG_Z | EXT_D},                                 // <Xdn>, <Pg>, <Xdn>, <Zm>.D
	F_Rn_Rm:              []int{REG_R, REG_R},                                                       // <R><n>, <R><m>
	F_Vd_Pg_ZnB:          []int{REG_V, REG_P, REG_Z | EXT_B},                                        // <V><d>, <Pg>, <Zn>.B
	F_Vd_Pg_ZnH:          []int{REG_V, REG_P, REG_Z | EXT_H},                                        // <V><d>, <Pg>, <Zn>.H
	F_Vd_Pg_ZnS:          []int{REG_V, REG_P, REG_Z | EXT_S},                                        // <V><d>, <Pg>, <Zn>.S
	F_Vd_Pg_ZnD:          []int{REG_V, REG_P, REG_Z | EXT_D},                                        // <V><d>, <Pg>, <Zn>.D
	F_Fd_Pg_ZnH:          []int{REG_F, REG_P, REG_Z | EXT_H},                                        // <V><d>, <Pg>, <Zn>.H
	F_Fd_Pg_ZnS:          []int{REG_F, REG_P, REG_Z | EXT_S},                                        // <V><d>, <Pg>, <Zn>.S
	F_Fd_Pg_ZnD:          []int{REG_F, REG_P, REG_Z | EXT_D},                                        // <V><d>, <Pg>, <Zn>.D
	F_ZdB_ZnB_ZmB:        []int{REG_Z | EXT_B, REG_Z | EXT_B, REG_Z | EXT_B},                        // <Zd>.B, <Zn>.B, <Zm>.B
	F_ZdH_ZnH_ZmH:        []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Zd>.H, <Zn>.H, <Zm>.H
	F_ZdS_ZnS_ZmS:        []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z | EXT_S},                        // <Zd>.S, <Zn>.S, <Zm>.S
	F_ZdD_ZnD_ZmD:        []int{REG_Z | EXT_D, REG_Z | EXT_D, REG_Z | EXT_D},                        // <Zd>.D, <Zn>.D, <Zm>.D
	F_ZdQ_ZnQ_ZmQ:        []int{REG_Z | EXT_Q, REG_Z | EXT_Q, REG_Z | EXT_Q},                        // <Zd>.Q, <Zn>.Q, <Zm>.Q
	F_Vdn_Pg_Vdn_ZmB:     []int{REG_V, REG_P, REG_V, REG_Z | EXT_B},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmH:     []int{REG_V, REG_P, REG_V, REG_Z | EXT_H},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmS:     []int{REG_V, REG_P, REG_V, REG_Z | EXT_S},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmD:     []int{REG_V, REG_P, REG_V, REG_Z | EXT_D},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Fdn_Pg_Fdn_ZmH:     []int{REG_F, REG_P, REG_F, REG_Z | EXT_H},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Fdn_Pg_Fdn_ZmS:     []int{REG_F, REG_P, REG_F, REG_Z | EXT_S},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Fdn_Pg_Fdn_ZmD:     []int{REG_F, REG_P, REG_F, REG_Z | EXT_D},                                 // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
}

// Key into the encoder table.
const (
	E_none = iota
	E_Rd_Rn_Rm
	E_Zda_Zn_Zm_i2
	E_size_Pg_Zn_Zd
	E_size_Pg_Zm_Zdn
	E_Zda_Zn_Zm_i3
	E_size_Pm_Rdn
	E_imm9h_imm9l_Rn_Zt
	E_size_Pg_Zn_Vd
	E_size_Pg_Zn_Pd_ImmFP0
	E_size_simm5_Pg_Zn_Pd
	E_size_uimm7_Pg_Zn_Pd
	E_size_Zm_Pg_Zn_Pd
	E_size_ZmD_Pg_Zn_Pd
	E_size_Pn_Pd
	E_size_Pm_Pn_Pd
	E_size_Rm_Rn_Pd
	E_Pd
	E_Pm_Pg_Pn_Pd
	E_Pg_Pn_Pd
	E_Pn_Pd
	E_Pg_Pn_Pdm
	E_size_Pv_Pdn
	E_Pg_Pdn
	E_Pg_Pn
	E_Pn
	E_size_Pg_Zn_Rd
	E_size_Pg_Zm_Rdn
	E_Rm_Rn
	E_size_Zm_Zn_Zd
	E_Pg_Rn_Rd

	// Equivalences
	E_size0_Pg_Zn_Zd = E_size_Pg_Zn_Zd
	E_Pg_8_5_Pd      = E_Pn_Pd
	E_Pg_Zn_Zd       = E_Pg_Rn_Rd
)

// The encoder table holds a list of encoding schemes for operands. Each scheme contains
// a list of rules and a list of indices that mark which operands need to be fed into the
// rule. Each rule produces a 32-bit number which should be OR'd with the base to create
// an instruction encoding.
var encoders = map[int]encoder{
	E_none:                 {},
	E_Rd_Rn_Rm:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_Zda_Zn_Zm_i2:         {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_size_Pg_Zn_Zd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}}},
	E_size_Pg_Zm_Zdn:       {[]rule{{[]int{0, 2}, Zdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
	E_Zda_Zn_Zm_i3:         {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi3}}},
	E_size_Pm_Rdn:          {[]rule{{[]int{0}, Rd}, {[]int{1}, Pm_8_5}, {[]int{1}, sveT}}},
	E_imm9h_imm9l_Rn_Zt:    {[]rule{{[]int{0}, Rt}, {[]int{1}, RnImm9MulVl}}},
	E_size_Pg_Zn_Vd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{2}, sveT}}},
	E_size_Pg_Zn_Pd_ImmFP0: {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, ImmFP0}}},
	E_size_simm5_Pg_Zn_Pd:  {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Imm5}}},
	E_size_uimm7_Pg_Zn_Pd:  {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Uimm7}}},
	E_size_Zm_Pg_Zn_Pd:     {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_ZmD_Pg_Zn_Pd:    {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2}, sveT}}},
	E_size_Pn_Pd:           {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}, {[]int{0, 1}, sveT}}},
	E_size_Pm_Pn_Pd:        {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}, {[]int{2}, Pm}, {[]int{0, 1, 2}, sveT}}},
	E_size_Rm_Rn_Pd:        {[]rule{{[]int{0}, Pd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0}, sveT}}},
	E_Pd:                   {[]rule{{[]int{0}, Pd}}},
	E_Pm_Pg_Pn_Pd:          {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Pn}, {[]int{3}, Pm}}},
	E_Pg_Pn_Pd:             {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Pn}}},
	E_Pn_Pd:                {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}}},
	E_Pg_Pn_Pdm:            {[]rule{{[]int{0, 3}, Pdm}, {[]int{1}, Pg}, {[]int{2}, Pn}}},
	E_size_Pv_Pdn:          {[]rule{{[]int{0, 2}, Pdn}, {[]int{1}, Pv}, {[]int{0, 2}, sveT}}},
	E_Pg_Pdn:               {[]rule{{[]int{0, 2}, Pdn}, {[]int{1}, Pg_8_5}}},
	E_Pg_Pn:                {[]rule{{[]int{0}, Pg}, {[]int{1}, Pn}}},
	E_Pn:                   {[]rule{{[]int{0}, Pn}}},
	E_size_Pg_Zn_Rd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{2}, sveT}}},
	E_size_Pg_Zm_Rdn:       {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{3}, sveT}}},
	E_Rm_Rn:                {[]rule{{[]int{0}, Rn}, {[]int{1}, Rm}}},
	E_size_Zm_Zn_Zd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0, 1, 2}, sveT}}},
	E_Pg_Rn_Rd:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Rn}}},
}
