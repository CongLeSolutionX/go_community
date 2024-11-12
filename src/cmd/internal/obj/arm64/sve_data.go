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
	AORN: {
		{0x25804010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORN <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	AORNS: {
		{0x25c04010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORNS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APAND: {
		{0x25004000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // AND <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APANDS: {
		{0x25404000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ANDS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBIC: {
		{0x25004010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BIC <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBICS: {
		{0x25404010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BICS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBRKA: {
		{0x25104000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKA <Pd>.B, <Pg>/Z, <Pn>.B
		{0x25104010, []int{F_PdB_PgM_PnB}, E_Pg_Pn_Pd}, // BRKA <Pd>.B, <Pg>/Z, <Pn>.B
	},
	APBRKAS: {
		{0x25504000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKAS <Pd>.B, <Pg>/Z, <Pn>.B
	},
	APBRKB: {
		{0x25904000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKB <Pd>.B, <Pg>/Z, <Pn>.B
		{0x25904010, []int{F_PdB_PgM_PnB}, E_Pg_Pn_Pd}, // BRKB <Pd>.B, <Pg>/Z, <Pn>.B
	},
	APBRKBS: {
		{0x25d04000, []int{F_PdB_PgZ_PnB}, E_Pg_Pn_Pd}, // BRKBS <Pd>.B, <Pg>/Z, <Pn>.B
	},
	APBRKN: {
		{0x25184000, []int{F_PdB_PgZ_PnB_PmB}, E_Pg_Pn_Pdm}, // BRKN <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
	},
	APBRKNS: {
		{0x25584000, []int{F_PdB_PgZ_PnB_PmB}, E_Pg_Pn_Pdm}, // BRKNS <Pdm>.B, <Pg>/Z, <Pn>.B, <Pdm>.B
	},
	APBRKPA: {
		{0x2500c000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPA <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBRKPAS: {
		{0x2540c000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPAS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBRKPB: {
		{0x2500c010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPB <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APBRKPBS: {
		{0x2540c010, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // BRKPBS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APCNTP: {
		{0x25208000, FG_Xd_Pg_PnT, E_size_Pg_Pn_Rd}, // CNTP <Xd>, <Pg>, <Pn>.<T>
	},
	APDECP: {
		{0x252d8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // DECP <Xdn>, <Pm>.<T>
	},
	APEOR: {
		{0x25004200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // EOR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APEORS: {
		{0x25404200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // EORS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APFALSE: {
		{0x2518e400, []int{F_PdB}, E_Pd}, // PFALSE <Pd>.B
	},
	APFIRST: {
		{0x2558c000, []int{F_PdnB_Pg_PdnB}, E_Pg_Pdn}, // PFIRST <Pdn>.B, <Pg>, <Pdn>.B
	},
	APINCP: {
		{0x252c8800, FG_Xdn_PmT, E_size_Pm_Rdn}, // INCP <Xdn>, <Pm>.<T>
	},
	APLDR: {
		{0x85800000, []int{F_Pt_AddrXSP, F_Pt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // LDR <Pt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	APNAND: {
		{0x25804210, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NAND <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APNANDS: {
		{0x25c04210, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NANDS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APNEXT: {
		{0x2519c400, FG_PdnT_Pv_PdnT, E_size_Pv_Pdn}, // PNEXT <Pdn>.<T>, <Pv>, <Pdn>.<T>
	},
	APNOR: {
		{0x25804200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NOR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APNORS: {
		{0x25c04200, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // NORS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APORR: {
		{0x25804000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORR <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APORRS: {
		{0x25c04000, []int{F_PdB_PgZ_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // ORRS <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	},
	APRDFFR: {
		{0x2519f000, []int{F_PdB}, E_Pd},            // RDFFR <Pd>.B
		{0x2518f000, []int{F_PdB_PgZ}, E_Pg_8_5_Pd}, // RDFFR <Pd>.B, <Pg>/Z
	},
	APRDFFRS: {
		{0x2558f000, []int{F_PdB_PgZ}, E_Pg_8_5_Pd}, // RDFFRS <Pd>.B, <Pg>/Z
	},
	APREV: {
		{0x05344000, FG_PdT_PnT, E_size_Pn_Pd}, // REV <Pd>.<T>, <Pn>.<T>
	},
	APSEL: {
		{0x25004210, []int{F_PdB_Pg_PnB_PmB}, E_Pm_Pg_Pn_Pd}, // SEL <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
	},
	APSQDECP: {
		{0x252a8c00, FG_Xdn_PmT, E_size_Pm_Rdn}, // SQDECP <Xdn>, <Pm>.<T>
	},
	APSQDECPW: {
		{0x252a8800, FG_Xdn_PmT_Wdn, E_size_Pm_XWdn}, // SQDECP <Xdn>, <Pm>.<T>, <Wdn>
	},
	APSQINCP: {
		{0x25288c00, FG_Xdn_PmT, E_size_Pm_Rdn}, // SQINCP <Xdn>, <Pm>.<T>
	},
	APSQINCPW: {
		{0x25288800, FG_Xdn_PmT_Wdn, E_size_Pm_XWdn}, // SQINCP <Xdn>, <Pm>.<T>, <Wdn>
	},
	APSTR: {
		{0xe5800000, []int{F_Pt_AddrXSP, F_Pt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Pt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	APTEST: {
		{0x2550c000, []int{F_Pg_PnB}, E_Pg_Pn}, // PTEST <Pg>, <Pn>.B
	},
	APTRN1: {
		{0x05205000, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // TRN1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	APTRN2: {
		{0x05205400, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // TRN2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	APTRUE: {
		{0x2518e3e0, FG_PdT, E_size_Pd},                 // PTRUE <Pd>.<T>{, <pattern>}
		{0x2518e000, FG_PdT_pattern, E_size_pattern_Pd}, // PTRUE <Pd>.<T>{, <pattern>}
	},
	APTRUES: {
		{0x2519e3e0, FG_PdT, E_size_Pd},                 // PTRUES <Pd>.<T>{, <pattern>}
		{0x2519e000, FG_PdT_pattern, E_size_pattern_Pd}, // PTRUES <Pd>.<T>{, <pattern>}
	},
	APUNPKHI: {
		{0x05314000, []int{F_PdH_PnB}, E_Pn_Pd},
	},
	APUNPKLO: {
		{0x05304000, []int{F_PdH_PnB}, E_Pn_Pd},
	},
	APUQDECP: {
		{0x252b8c00, FG_Xdn_PmT, E_size_Pm_Rdn}, // UQDECP <Xdn>, <Pm>.<T>
	},
	APUQDECPW: {
		{0x252b8800, FG_Wdn_PmT, E_size_Pm_Rdn}, // UQDECP <Wdn>, <Pm>.<T>
	},
	APUQINCP: {
		{0x25298c00, FG_Xdn_PmT, E_size_Pm_Rdn}, // UQINCP <Xdn>, <Pm>.<T>
	},
	APUQINCPW: {
		{0x25298800, FG_Wdn_PmT, E_size_Pm_Rdn}, // UQINCP <Wdn>, <Pm>.<T>
	},
	APUZP1: {
		{0x05204800, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // UZP1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	APUZP2: {
		{0x05204c00, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // UZP2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	APWRFFR: {
		{0x25289000, []int{F_PnB}, E_Pn}, // WRFFR <Pn>.B
	},
	APZIP1: {
		{0x05204000, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // ZIP1 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
	},
	APZIP2: {
		{0x05204400, FG_PdT_PnT_PmT, E_size_Pm_Pn_Pd}, // ZIP2 <Pd>.<T>, <Pn>.<T>, <Pm>.<T>
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
	AZABS: {
		{0x0416a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // ABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZADD: {
		{0x04000000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},     // ADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04200000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // ADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2520c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // ADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2520c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // ADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZADDPL: {
		{0x04605000, []int{F_Xd_Xn_Imm}, E_Rn_imm6_Rd}, // ADDPL   <Xd|SP>, <Xn|SP>, #<imm>
	},
	AZADDVL: {
		{0x04205000, []int{F_Xd_Xn_Imm}, E_Rn_imm6_Rd}, // ADDVL   <Xd|SP>, <Xn|SP>, #<imm>
	},
	AZADR: {
		{0x04a0a000, []int{F_ZdS_AddrZnSZmS, F_ZdS_AddrZnSZmSLSL, F_ZdD_AddrZnDZmD, F_ZdD_AddrZnDZmDLSL}, E_sz_Zm_msz_Zn_Zd}, // ADR <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
		{0x0420a000, []int{F_ZdD_AddrZnDZmDSXTW}, E_Zm_msz_Zn_Zd},                                                            // ADR <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>, SXTW{ <amount>}]
		{0x0460a000, []int{F_ZdD_AddrZnDZmDUXTW}, E_Zm_msz_Zn_Zd},                                                            // ADR <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>, UXTW{ <amount>}]
	},
	AZAND: {
		{0x041a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // AND <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04203000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // AND <Zd>.D, <Zn>.D, <Zm>.D
		{0x05800000, FG_ZdnT_ZdnT_const, E_imm13_Zdn},        // AND <Zdn>.<T>, <Zdn>.<T>, #<const>
	},
	AZANDV: {
		{0x041a2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // ANDV <V><d>, <Pg>, <Zn>.<T>
	},
	AZASR: {
		{0x04108000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},                // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04209000, FG_ZdT_ZnT_const, E_tszh_tszl_imm3_bias1_Zn_Zd},        // ASR <Zd>.<T>, <Zn>.<T>, #<const>
		{0x04208000, FG_ZdT_ZnT_ZmD, E_size_Zm_ZnT_ZdT},                     // ASR <Zd>.<T>, <Zn>.<T>, <Zm>.D
		{0x04008000, FG_ZdnT_PgM_ZdnT_const, E_tszh_Pg_tszl_imm3_bias1_Zdn}, // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
		{0x04188000, FG_ZdnT_PgM_ZdnT_ZmD, E_size_Pg_ZmD_Zdn},               // ASR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	},
	AZASRD: {
		{0x04048000, FG_ZdnT_PgM_ZdnT_const, E_tszh_Pg_tszl_imm3_bias1_Zdn}, // ASRD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
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
	AZBIC: {
		{0x041b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // BIC <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04e03000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // BIC <Zd>.D, <Zn>.D, <Zm>.D
	},
	AZCLASTA: {
		{0x0530a000, FG_Rdn_Pg_Rdn_ZmT, E_size_Pg_Zm_Rdn},   // CLASTA <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
		{0x052a8000, FG_Vdn_Pg_Vdn_ZmT, E_size_Pg_Zm_Rdn},   // CLASTA <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
		{0x05288000, FG_ZdnT_Pg_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // CLASTA <Zdn>.<T>, <Pg>, <Zdn>.<T>, <Zm>.<T>
	},
	AZCLASTB: {
		{0x0531a000, FG_Rdn_Pg_Rdn_ZmT, E_size_Pg_Zm_Rdn},   // CLASTB <R><dn>, <Pg>, <R><dn>, <Zm>.<T>
		{0x052b8000, FG_Vdn_Pg_Vdn_ZmT, E_size_Pg_Zm_Rdn},   // CLASTB <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
		{0x05298000, FG_ZdnT_Pg_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // CLASTB <Zdn>.<T>, <Pg>, <Zdn>.<T>, <Zm>.<T>
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
	AZCNTB: {
		{0x042fe3e0, []int{F_Xd}, E_Rd},                          // CNTB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x042fe000, []int{F_Xd_pattern}, E_pattern_Rd},          // CNTB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0420e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // CNTB <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZCNTD: {
		{0x04efe3e0, []int{F_Xd}, E_Rd},                          // CNTD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04efe000, []int{F_Xd_pattern}, E_pattern_Rd},          // CNTD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04e0e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // CNTD <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZCNTH: {
		{0x046fe3e0, []int{F_Xd}, E_Rd},                          // CNTH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x046fe000, []int{F_Xd_pattern}, E_pattern_Rd},          // CNTH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0460e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // CNTH <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZCNTW: {
		{0x04afe3e0, []int{F_Xd}, E_Rd},                          // CNTW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04afe000, []int{F_Xd_pattern}, E_pattern_Rd},          // CNTW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04a0e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // CNTW <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZCOMPACT: {
		{0x05218000, []int{F_ZdS_Pg_ZnS, F_ZdD_Pg_ZnD}, E_size_Pg_Zn_Zd}, // COMPACT <Zd>.<T>, <Pg>, <Zn>.<T>
	},
	AZCPY: {
		{0x05104000, FG_ZdT_PgM_Imm, E_size_Pg_imm8_Zd},        // CPY <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
		{0x05104000, FG_ZdT_PgM_Imm_Imm, E_size_Pg_sh_imm8_Zd}, // CPY <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
		{0x05100000, FG_ZdT_PgZ_Imm, E_size_Pg_imm8_Zd},        // CPY <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
		{0x05100000, FG_ZdT_PgZ_Imm_Imm, E_size_Pg_sh_imm8_Zd}, // CPY <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
		{0x0528a000, FG_ZdT_PgM_WXn, E_size_Pg_Rn_Zd},          // CPY <Zd>.<T>, <Pg>/M, <R><n|SP>
		{0x05208000, FG_ZdT_PgM_Vn, E_size_Pg_Vn_Zd},           // CPY <Zd>.<T>, <Pg>/M, <V><n>
	},
	AZCTERMEQ: {
		{0x25e02000, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	AZCTERMEQW: {
		{0x25a02000, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	AZCTERMNE: {
		{0x25e02010, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	AZCTERMNEW: {
		{0x25a02010, []int{F_Rn_Rm}, E_Rm_Rn}, // CTERMEQ <R><n>, <R><m>
	},
	AZDECB: {
		{0x043fe7e0, []int{F_Xd}, E_Rd},                          // DECB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x043fe400, []int{F_Xd_pattern}, E_pattern_Rd},          // DECB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0430e400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // DECB <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZDECD: {
		{0x04ffe7e0, []int{F_Xd}, E_Rd},                           // DECD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04ffe400, []int{F_Xd_pattern}, E_pattern_Rd},           // DECD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04f0e400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // DECD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04ffc7e0, []int{F_ZdD}, E_Rd},                          // DECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04ffc400, []int{F_ZdD_pattern}, E_pattern_Rd},          // DECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04f0c400, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rd}, // DECD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZDECH: {
		{0x047fe7e0, []int{F_Xd}, E_Rd},                           // DECH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x047fe400, []int{F_Xd_pattern}, E_pattern_Rd},           // DECH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0470e400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // DECH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x047fc7e0, []int{F_ZdH}, E_Rd},                          // DECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x047fc400, []int{F_ZdH_pattern}, E_pattern_Rd},          // DECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0470c400, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rd}, // DECH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZDECP: {
		{0x252d8000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // DECP <Zdn>.<T>, <Pm>.<T>
	},
	AZDECW: {
		{0x04bfe7e0, []int{F_Xd}, E_Rd},                           // DECW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04bfe400, []int{F_Xd_pattern}, E_pattern_Rd},           // DECW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04b0e400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // DECW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04bfc7e0, []int{F_ZdS}, E_Rd},                          // DECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04bfc400, []int{F_ZdS_pattern}, E_pattern_Rd},          // DECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04b0c400, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rd}, // DECW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZDUP: {
		{0x2538c000, FG_ZdT_imm, E_size_imm8_Zd},          // DUP <Zd>.<T>, #<imm>{, <shift>}
		{0x2538c000, FG_ZdT_imm_shift, E_size_sh_imm8_Zd}, // DUP <Zd>.<T>, #<imm>{, <shift>}
		{0x05203800, FG_ZdT_RnSP, E_size_Rn_Zd},           // DUP <Zd>.<T>, <R><n|SP>
		{0x05202000, FG_ZdTq_ZnTqidx, E_dup_indexed},      // DUP <Zd>.<T>, <Zn>.<T>[<imm>]
	},
	AZDUPM: {
		{0x05c00000, FG_ZdT_const, E_dupm}, // DUPM <Zd>.<T>, #<const>
	},
	AZEOR: {
		{0x04190000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // EOR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04a03000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // EOR <Zd>.D, <Zn>.D, <Zm>.D
		{0x05400000, FG_ZdnT_ZdnT_const, E_imm13_Zdn},        // EOR <Zdn>.<T>, <Zdn>.<T>, #<const>
	},
	AZEORV: {
		{0x04192000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // EORV <V><d>, <Pg>, <Zn>.<T>
	},
	AZEXT: {
		{0x05200000, []int{F_ZdnB_ZdnB_ZmB_imm}, E_imm8h_imm8l_Zm_Zdn}, // EXT <Zdn>.B, <Zdn>.B, <Zm>.B, #<imm>
	},
	AZFABD: {
		{0x65088000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFABS: {
		{0x041ca000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FABS <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFACGE: {
		{0x6500c010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd}, // FACGE <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFACGT: {
		{0x6500e010, FG_PdT_PgZ_ZnT_ZmT_FP, E_size_Zm_Pg_Zn_Pd}, // FACGT <Pd>.<T>, <Pg>/Z, <Zn>.<T>, <Zm>.<T>
	},
	AZFADD: {
		{0x65008000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000000, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},                  // FADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x65188000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p5_1p0_Zdn}, // FADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFADDA: {
		{0x65182000, []int{F_Fdn_Pg_Fdn_ZmH, F_Fdn_Pg_Fdn_ZmS, F_Fdn_Pg_Fdn_ZmD}, E_size_Pg_Zm_Rdn}, // FADDA <V><dn>, <Pg>, <V><dn>, <Zm>.<T>
	},
	AZFADDV: {
		{0x65002000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FADDV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFCADD: {
		{0x64008000, FG_ZdnT_PgM_ZdnT_ZmT_const_FP, E_size_rot_Pg_Zm_Zdn}, // FCADD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>, <const>
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
	AZFCMLA: {
		{0x64000000, FG_ZdT_PgM_ZnT_ZmT_Imm_FP, E_size_Zm_rot_Pg_Zn_Zda}, // FCMLA <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>, <const>
		{0x64a01000, []int{F_ZdH_ZnH_ZmHidx_const}, E_Zmi2_rot_Zn_Zda},   // FCMLA <Zda>.H, <Zn>.H, <Zm>.H[<imm>], <const>
		{0x64e01000, []int{F_ZdS_ZnS_ZmSidx_const}, E_Zmi1_rot_Zn_Zda},   // FCMLA <Zda>.S, <Zn>.S, <Zm>.S[<imm>], <const>
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
	AZFCPY: {
		{0x0510c000, FG_ZdT_PgM_const_FP, E_size_Pg_imm8_Zd}, // FCPY <Zd>.<T>, <Pg>/M, #<const>
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
	AZFDUP: {
		{0x2539c000, FG_ZdT_const_FP, E_size_imm8_Zd_FP}, // FDUP <Zd>.<T>, #<const>
	},
	AZFEXPA: {
		{0x0420b800, FG_ZdT_ZnT_FP, E_size_Zn_Zd}, // FEXPA <Zd>.<T>, <Zn>.<T>
	},
	AZFMAD: {
		{0x65208000, FG_ZdnT_PgM_ZmT_ZaT_FP, E_size_Za_Pg_Zm_Zdn}, // FMAD <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZFMAX: {
		{0x65068000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x651e8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p0_1p0_Zdn}, // FMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFMAXNM: {
		{0x65048000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FMAXNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x651c8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p0_1p0_Zdn}, // FMAXNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFMAXNMV: {
		{0x65042000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMAXNMV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMAXV: {
		{0x65062000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMIN: {
		{0x65078000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x651f8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p0_1p0_Zdn}, // FMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFMINNM: {
		{0x65058000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FMINNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x651d8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p0_1p0_Zdn}, // FMINNM <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFMINNMV: {
		{0x65052000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMINNMV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMINV: {
		{0x65072000, FG_Vd_Pg_ZnT_FP, E_size_Pg_Zn_Vd}, // FMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZFMLA: {
		{0x65200000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Zm_Pg_Zn_Zda}, // FMLA <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
		{0x64200000, []int{F_ZdH_ZnH_ZmHidx}, E_i3h_i3l_Zm_Zn_Zd},  // FMLA <Zda>.H, <Zn>.H, <Zm>.H[<imm>]
		{0x64a00000, []int{F_ZdS_ZnS_ZmSidx}, E_Zmi2_Zn_Zda},       // FMLA <Zda>.S, <Zn>.S, <Zm>.S[<imm>]
		{0x64e00000, []int{F_ZdD_ZnD_ZmDidx}, E_Zmi1_Zn_Zda},       // FMLA <Zda>.D, <Zn>.D, <Zm>.D[<imm>]
	},
	AZFMLS: {
		{0x65202000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Zm_Pg_Zn_Zda}, // FMLS <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
		{0x64200400, []int{F_ZdH_ZnH_ZmHidx}, E_i3h_i3l_Zm_Zn_Zd},  // FMLS <Zda>.H, <Zn>.H, <Zm>.H[<imm>]
		{0x64a00400, []int{F_ZdS_ZnS_ZmSidx}, E_Zmi2_Zn_Zda},       // FMLS <Zda>.S, <Zn>.S, <Zm>.S[<imm>]
		{0x64e00400, []int{F_ZdD_ZnD_ZmDidx}, E_Zmi1_Zn_Zda},       // FMLS <Zda>.D, <Zn>.D, <Zm>.D[<imm>]
	},
	AZFMMLA: {
		{0x64a0e400, []int{F_ZdS_ZnS_ZmS}, E_Zm_Zn_Zda}, // FMMLA <Zda>.S, <Zn>.S, <Zm>.S (+FEAT_F32MM)
		{0x64e0e400, []int{F_ZdD_ZnD_ZmD}, E_Zm_Zn_Zda}, // FMMLA <Zda>.D, <Zn>.D, <Zm>.D (+FEAT_F64MM)
	},
	AZFMSB: {
		{0x6520a000, FG_ZdnT_PgM_ZmT_ZaT_FP, E_size_Za_Pg_Zm_Zdn}, // FMSB <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZFMUL: {
		{0x65028000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FMUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000800, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},                  // FMUL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x64202000, []int{F_ZdH_ZnH_ZmHidx}, E_i3h_i3l_Zm_Zn_Zd},         // FMUL <Zd>.H, <Zn>.H, <Zm>.H[<imm>]
		{0x64a02000, []int{F_ZdS_ZnS_ZmSidx}, E_i2_Zm_Zn_Zd},              // FMUL <Zd>.S, <Zn>.S, <Zm>.S[<imm>]
		{0x64e02000, []int{F_ZdD_ZnD_ZmDidx}, E_i1_Zm_Zn_Zd},              // FMUL <Zd>.D, <Zn>.D, <Zm>.D[<imm>]
		{0x651a8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p5_2p0_Zdn}, // FMUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFMULX: {
		{0x650a8000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FMULX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFNEG: {
		{0x041da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FNEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFNMAD: {
		{0x6520c000, FG_ZdnT_PgM_ZmT_ZaT_FP, E_size_Za_Pg_Zm_Zdn}, // FNMAD <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZFNMLA: {
		{0x65204000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Zm_Pg_Zn_Zda}, // FNMLA <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
	},
	AZFNMLS: {
		{0x65206000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Zm_Pg_Zn_Zda}, // FNMLS <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
	},
	AZFNMSB: {
		{0x6520e000, FG_ZdnT_PgM_ZmT_ZaT_FP, E_size_Za_Pg_Zm_Zdn}, // FNMSB <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZFRECPE: {
		{0x650e3000, FG_ZdT_ZnT_FP, E_size_Zn_Zd}, // FRECPE <Zd>.<T>, <Zn>.<T>
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
	AZFRSQRTE: {
		{0x650f3000, FG_ZdT_ZnT_FP, E_size_Zn_Zd}, // FRSQRTE <Zd>.<T>, <Zn>.<T>
	},
	AZFRSQRTS: {
		{0x65001c00, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FRSQRTS <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFSCALE: {
		{0x65098000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn}, // FSCALE <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZFSQRT: {
		{0x650da000, FG_ZdT_PgM_ZnT_FP, E_size_Pg_Zn_Zd}, // FSQRT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZFSUB: {
		{0x65018000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FSUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x65000400, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd},                  // FSUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x65198000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p5_1p0_Zdn}, // FSUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFSUBR: {
		{0x65038000, FG_ZdnT_PgM_ZdnT_ZmT_FP, E_size_Pg_Zm_Zdn},           // FSUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x651b8000, FG_ZdnT_PgM_ZdnT_const_FP, E_size_Pg_i1_0p5_1p0_Zdn}, // FSUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	},
	AZFTMAD: {
		{0x65108000, FG_ZdnT_ZdnT_ZmT_imm_FP, E_size_imm3_Zm_Zdn}, // FTMAD <Zdn>.<T>, <Zdn>.<T>, <Zm>.<T>, #<imm>
	},
	AZFTSMUL: {
		{0x65000c00, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FTSMUL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZFTSSEL: {
		{0x0420b000, FG_ZdT_ZnT_ZmT_FP, E_size_Zm_Zn_Zd}, // FTSSEL <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
	},
	AZINCB: {
		{0x043fe3e0, []int{F_Xd}, E_Rd},                          // INCB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x043fe000, []int{F_Xd_pattern}, E_pattern_Rd},          // INCB <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0430e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd}, // INCB <Xd>{, <pattern>{, MUL #<imm>}}
	},
	AZINCD: {
		{0x04ffe3e0, []int{F_Xd}, E_Rd},                           // INCD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04ffe000, []int{F_Xd_pattern}, E_pattern_Rd},           // INCD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04f0e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // INCD <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04ffc3e0, []int{F_ZdD}, E_Rd},                          // INCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04ffc000, []int{F_ZdD_pattern}, E_pattern_Rd},          // INCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04f0c000, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rd}, // INCD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZINCH: {
		{0x047fe3e0, []int{F_Xd}, E_Rd},                           // INCH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x047fe000, []int{F_Xd_pattern}, E_pattern_Rd},           // INCH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x0470e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // INCH <Xd>{, <pattern>{, MUL #<imm>}}
		{0x047fc3e0, []int{F_ZdH}, E_Rd},                          // INCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x047fc000, []int{F_ZdH_pattern}, E_pattern_Rd},          // INCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0470c000, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rd}, // INCH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZINCP: {
		{0x252c8000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // INCP <Zdn>.<T>, <Pm>.<T>
	},
	AZINCW: {
		{0x04bfe3e0, []int{F_Xd}, E_Rd},                           // INCW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04bfe000, []int{F_Xd_pattern}, E_pattern_Rd},           // INCW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04b0e000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rd},  // INCW <Xd>{, <pattern>{, MUL #<imm>}}
		{0x04bfc3e0, []int{F_ZdS}, E_Rd},                          // INCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04bfc000, []int{F_ZdS_pattern}, E_pattern_Rd},          // INCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04b0c000, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rd}, // INCW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZINDEX: {
		{0x04204000, FG_ZdT_Imm_Imm, E_size_imm5b_imm5_Zd}, // INDEX <Zd>.<T>, #<imm1>, #<imm2>
		{0x04204800, FG_ZdT_Imm_XWm, E_size_Rm_imm5_Zd},    // INDEX <Zd>.<T>, #<imm>, <R><m>
		{0x04204400, FG_ZdT_XWn_Imm, E_size_imm5_Rn_Zd},    // INDEX <Zd>.<T>, <R><n>, #<imm>
		{0x04204c00, FG_ZdT_XWn_XWm, E_size_Rm_Rn_Zd},      // INDEX <Zd>.<T>, <R><n>, <R><m>
	},
	AZINSR: {
		{0x05243800, FG_ZdnT_Rm, E_size_Rm_Zdn}, // INSR <Zdn>.<T>, <R><m>
		{0x05343800, FG_ZdnT_Vm, E_size_Rm_Zdn}, // INSR <Zdn>.<T>, <V><m>
	},
	AZLASTA: {
		{0x0520a000, FG_Rd_Pg_ZnT, E_size_Pg_Zn_Rd}, // LASTA <R><d>, <Pg>, <Zn>.<T>
		{0x05228000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // LASTA <V><d>, <Pg>, <Zn>.<T>
	},
	AZLASTB: {
		{0x0521a000, FG_Rd_Pg_ZnT, E_size_Pg_Zn_Rd}, // LASTA <R><d>, <Pg>, <Zn>.<T>
		{0x05238000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // LASTB <V><d>, <Pg>, <Zn>.<T>
	},
	AZLD1B: {
		{0xa4004000, []int{F_RlZtBc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa4204000, []int{F_RlZtHc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1B { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa4404000, []int{F_RlZtSc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa4604000, []int{F_RlZtDc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa400a000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa420a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1B { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa440a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa460a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc4004000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LD1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84004000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LD1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc440c000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                      // LD1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc420c000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_imm5_Pg_Zn_Zt},             // LD1B { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x8420c000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_imm5_Pg_Zn_Zt},             // LD1B { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLD1D: {
		{0xa5e04000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                                     // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5e0a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc5a04000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW3, F_RlZtDc1_PgZ_AddrXSPZmDUXTW3}, E_xs_Zm_Pg_Rn_Zt}, // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #3]
		{0xc5804000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0xc5e0c000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL3}, E_Zm_Pg_Rn_Zt},                                    // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #3]
		{0xc5c0c000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LD1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc5a0c000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_8imm5_Pg_Zn_Zt},              // LD1D { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	},
	AZLD1H: {
		{0xa4e04000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4c04000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LD1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4e0a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa4a0a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa4c0a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc4a04000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW1, F_RlZtDc1_PgZ_AddrXSPZmDUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #1]
		{0xc4804000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84a04000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW1, F_RlZtSc1_PgZ_AddrXSPZmSUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LD1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
		{0x84804000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc4e0c000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_Zt},                                    // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xc4c0c000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LD1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc4a0c000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_2imm5_Pg_Zn_Zt},              // LD1H { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x84a0c000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_2imm5_Pg_Zn_Zt},              // LD1H { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
		{0xa4a04000, []int{F_RlZtHc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LD1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
	},
	AZLD1RB: {
		{0x84408000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RB { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x8440e000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x8440a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x8440c000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RD: {
		{0x85c0e000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_8imm6_Pg_Rn_Zt}, // LD1RD { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RH: {
		{0x84c0e000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_2imm6_Pg_Rn_Zt}, // LD1RH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x84c0a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_2imm6_Pg_Rn_Zt}, // LD1RH { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x84c0c000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_2imm6_Pg_Rn_Zt}, // LD1RH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1ROB: {
		{0xa4200000, []int{F_RlZtBc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                  // LD1ROB { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa4202000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_32imm4_Pg_Rn_Zt}, // LD1ROB { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1ROD: {
		{0xa5a00000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                              // LD1ROD { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5a02000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_32imm4_Pg_Rn_Zt}, // LD1ROD { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1ROH: {
		{0xa4a00000, []int{F_RlZtHc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                              // LD1ROH { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4a02000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_32imm4_Pg_Rn_Zt}, // LD1ROH { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1ROW: {
		{0xa5200000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                              // LD1ROW { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa5202000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_32imm4_Pg_Rn_Zt}, // LD1ROW { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RQB: {
		{0xa4000000, []int{F_RlZtBc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                  // LD1RQB { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa4002000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_16imm4_Pg_Rn_Zt}, // LD1RQB { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RQD: {
		{0xa5800000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                              // LD1RQD { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5802000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_16imm4_Pg_Rn_Zt}, // LD1RQD { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RQH: {
		{0xa4800000, []int{F_RlZtHc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                              // LD1RQH { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4802000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_16imm4_Pg_Rn_Zt}, // LD1RQH { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RQW: {
		{0xa5000000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                              // LD1RQW { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa5002000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_16imm4_Pg_Rn_Zt}, // LD1RQW { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RSB: {
		{0x85c08000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RSB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x85c0c000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RSB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x85c0a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm6_Pg_Rn_Zt}, // LD1RSB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RSH: {
		{0x85408000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_2imm6_Pg_Rn_Zt}, // LD1RSH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x8540a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_2imm6_Pg_Rn_Zt}, // LD1RSH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RSW: {
		{0x84c08000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_4imm6_Pg_Rn_Zt}, // LD1RSW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1RW: {
		{0x8540e000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_4imm6_Pg_Rn_Zt}, // LD1RW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
		{0x8540c000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_4imm6_Pg_Rn_Zt}, // LD1RW { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>}]
	},
	AZLD1SB: {
		{0xa5c04000, []int{F_RlZtHc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa5804000, []int{F_RlZtDc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa5a04000, []int{F_RlZtSc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LD1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa5c0a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa580a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa5a0a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},        // LD1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc4000000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LD1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84000000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LD1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc4408000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                      // LD1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc4208000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_imm5_Pg_Zn_Zt},             // LD1SB { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x84208000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_imm5_Pg_Zn_Zt},             // LD1SB { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLD1SH: {
		{0xa5004000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa5204000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LD1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa500a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa520a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc4a00000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW1, F_RlZtDc1_PgZ_AddrXSPZmDUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #1]
		{0xc4800000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84a00000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW1, F_RlZtSc1_PgZ_AddrXSPZmSUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LD1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
		{0x84800000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc4e08000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_Zt},                                    // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xc4c08000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LD1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc4a08000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_2imm5_Pg_Zn_Zt},              // LD1SH { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x84a08000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_2imm5_Pg_Zn_Zt},              // LD1SH { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLD1SW: {
		{0xa4804000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa480a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc5200000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW2, F_RlZtDc1_PgZ_AddrXSPZmDUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #2]
		{0xc5000000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0xc5608000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_Zt},                                    // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xc5408000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LD1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc5208000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_4imm5_Pg_Zn_Zt},              // LD1SW { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	},
	AZLD1W: {
		{0xa5604000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa5404000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LD1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa560a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa540a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // LD1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc5204000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW2, F_RlZtDc1_PgZ_AddrXSPZmDUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #2]
		{0xc5004000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x85204000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW2, F_RlZtSc1_PgZ_AddrXSPZmSUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LD1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #2]
		{0x85004000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LD1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc560c000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_Zt},                                    // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xc540c000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LD1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc520c000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_4imm5_Pg_Zn_Zt},              // LD1W { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x8520c000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_4imm5_Pg_Zn_Zt},              // LD1W { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLD2B: {
		{0xa420c000, []int{F_RlZtBc2i1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                   // LD2B { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa420e000, []int{F_RlZtBc2i1_PgZ_AddrXSP, F_RlZtBc2i1_PgZ_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // LD2B { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD2D: {
		{0xa5a0c000, []int{F_RlZtDc2i1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                               // LD2D { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5a0e000, []int{F_RlZtDc2i1_PgZ_AddrXSP, F_RlZtDc2i1_PgZ_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // LD2D { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD2H: {
		{0xa4a0c000, []int{F_RlZtHc2i1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                               // LD2H { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4a0e000, []int{F_RlZtHc2i1_PgZ_AddrXSP, F_RlZtHc2i1_PgZ_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // LD2H { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD2W: {
		{0xa520c000, []int{F_RlZtSc2i1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                               // LD2W { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa520e000, []int{F_RlZtSc2i1_PgZ_AddrXSP, F_RlZtSc2i1_PgZ_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // LD2W { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD3B: {
		{0xa440c000, []int{F_RlZtBc3i1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                   // LD3B { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa440e000, []int{F_RlZtBc3i1_PgZ_AddrXSP, F_RlZtBc3i1_PgZ_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // LD3B { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD3D: {
		{0xa5c0c000, []int{F_RlZtDc3i1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                               // LD3D { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5c0e000, []int{F_RlZtDc3i1_PgZ_AddrXSP, F_RlZtDc3i1_PgZ_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // LD3D { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD3H: {
		{0xa4c0c000, []int{F_RlZtHc3i1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                               // LD3H { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4c0e000, []int{F_RlZtHc3i1_PgZ_AddrXSP, F_RlZtHc3i1_PgZ_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // LD3H { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD3W: {
		{0xa540c000, []int{F_RlZtSc3i1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                               // LD3W { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa540e000, []int{F_RlZtSc3i1_PgZ_AddrXSP, F_RlZtSc3i1_PgZ_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // LD3W { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD4B: {
		{0xa460c000, []int{F_RlZtBc4i1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                   // LD4B { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa460e000, []int{F_RlZtBc4i1_PgZ_AddrXSP, F_RlZtBc4i1_PgZ_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // LD4B { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD4D: {
		{0xa5e0c000, []int{F_RlZtDc4i1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                               // LD4D { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa5e0e000, []int{F_RlZtDc4i1_PgZ_AddrXSP, F_RlZtDc4i1_PgZ_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // LD4D { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD4H: {
		{0xa4e0c000, []int{F_RlZtHc4i1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                               // LD4H { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa4e0e000, []int{F_RlZtHc4i1_PgZ_AddrXSP, F_RlZtHc4i1_PgZ_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // LD4H { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLD4W: {
		{0xa560c000, []int{F_RlZtSc4i1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                               // LD4W { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa560e000, []int{F_RlZtSc4i1_PgZ_AddrXSP, F_RlZtSc4i1_PgZ_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // LD4W { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDFF1B: {
		{0xa41f6000, []int{F_RlZtBc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa4006000, []int{F_RlZtBc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa47f6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa4606000, []int{F_RlZtDc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa43f6000, []int{F_RlZtHc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1B { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa4206000, []int{F_RlZtHc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1B { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa45f6000, []int{F_RlZtSc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa4406000, []int{F_RlZtSc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xc4006000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84006000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc440e000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                      // LDFF1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xc420e000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_imm5_Pg_Zn_Zt},             // LDFF1B { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x8420e000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_imm5_Pg_Zn_Zt},             // LDFF1B { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLDFF1D: {
		{0xc5a06000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW3, F_RlZtDc1_PgZ_AddrXSPZmDUXTW3}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #3]
		{0xc5806000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0xc5e0e000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL3}, E_Zm_Pg_Rn_Zt},                                    // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #3]
		{0xc5c0e000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa5e06000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                                     // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #3}]
		{0xa5ff6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #3}]
		{0xc5a0e000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_8imm5_Pg_Zn_Zt},              // LDFF1D { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	},
	AZLDFF1H: {
		{0xc4a06000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW1, F_RlZtDc1_PgZ_AddrXSPZmDUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #1]
		{0xc4806000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84a06000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW1, F_RlZtSc1_PgZ_AddrXSPZmSUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
		{0x84806000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc4e0e000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_Zt},                                    // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xc4c0e000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa4e06000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa4ff6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa4a06000, []int{F_RlZtHc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LDFF1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa4bf6000, []int{F_RlZtHc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa4c06000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LDFF1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa4df6000, []int{F_RlZtSc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xc4a0e000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_2imm5_Pg_Zn_Zt},              // LDFF1H { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x84a0e000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_2imm5_Pg_Zn_Zt},              // LDFF1H { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLDFF1SB: {
		{0xc4002000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84002000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc440a000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                      // LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa59f6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa5806000, []int{F_RlZtDc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa5df6000, []int{F_RlZtHc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa5c06000, []int{F_RlZtHc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa5bf6000, []int{F_RlZtSc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                            // LDFF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xa5a06000, []int{F_RlZtSc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                       // LDFF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>}]
		{0xc420a000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_imm5_Pg_Zn_Zt},             // LDFF1SB { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x8420a000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_imm5_Pg_Zn_Zt},             // LDFF1SB { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLDFF1SH: {
		{0xc4a02000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW1, F_RlZtDc1_PgZ_AddrXSPZmDUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #1]
		{0xc4802000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x84a02000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW1, F_RlZtSc1_PgZ_AddrXSPZmSUXTW1}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
		{0x84802000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc4e0a000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_Zt},                                    // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xc4c0a000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa5006000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa51f6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa5206000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                                     // LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xa53f6000, []int{F_RlZtSc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #1}]
		{0xc4a0a000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_2imm5_Pg_Zn_Zt},              // LDFF1SH { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x84a0a000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_2imm5_Pg_Zn_Zt},              // LDFF1SH { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLDFF1SW: {
		{0xc5202000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW2, F_RlZtDc1_PgZ_AddrXSPZmDUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #2]
		{0xc5002000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0xc560a000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_Zt},                                    // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xc540a000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa4806000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xa49f6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xc520a000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_4imm5_Pg_Zn_Zt},              // LDFF1SW { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	},
	AZLDFF1W: {
		{0xc5206000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW2, F_RlZtDc1_PgZ_AddrXSPZmDUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod> #2]
		{0xc5006000, []int{F_RlZtDc1_PgZ_AddrXSPZmDSXTW, F_RlZtDc1_PgZ_AddrXSPZmDUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, <mod>]
		{0x85206000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW2, F_RlZtSc1_PgZ_AddrXSPZmSUXTW2}, E_xs_Zm_Pg_Rn_Zt}, // LDFF1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #2]
		{0x85006000, []int{F_RlZtSc1_PgZ_AddrXSPZmSSXTW, F_RlZtSc1_PgZ_AddrXSPZmSUXTW}, E_xs_Zm_Pg_Rn_Zt},   // LDFF1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
		{0xc560e000, []int{F_RlZtDc1_PgZ_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_Zt},                                    // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xc540e000, []int{F_RlZtDc1_PgZ_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                        // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Zm>.D]
		{0xa5606000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xa57f6000, []int{F_RlZtDc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xa5406000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                                     // LDFF1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xa55f6000, []int{F_RlZtSc1_PgZ_AddrXSP}, E_Pg_Rn_Zt},                                              // LDFF1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, <Xm>, LSL #2}]
		{0xc520e000, []int{F_RlZtDc1_PgZ_AddrZnD, F_RlZtDc1_PgZ_AddrZnDImm}, E_4imm5_Pg_Zn_Zt},              // LDFF1W { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
		{0x8520e000, []int{F_RlZtSc1_PgZ_AddrZnS, F_RlZtSc1_PgZ_AddrZnSImm}, E_4imm5_Pg_Zn_Zt},              // LDFF1W { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	},
	AZLDNF1B: {
		{0xa410a000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa470a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1B { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa430a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1B { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa450a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1B { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1D: {
		{0xa5f0a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1H: {
		{0xa4f0a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1H { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa4b0a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa4d0a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1H { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1SB: {
		{0xa590a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SB { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa5d0a000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SB { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa5b0a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SB { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1SH: {
		{0xa510a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SH { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa530a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SH { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1SW: {
		{0xa490a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1SW { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNF1W: {
		{0xa570a000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1W { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xa550a000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNF1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNT1B: {
		{0xa400c000, []int{F_RlZtBc1_PgZ_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                // LDNT1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
		{0xa400e000, []int{F_RlZtBc1_PgZ_AddrXSP, F_RlZtBc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNT1B { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNT1D: {
		{0xa580c000, []int{F_RlZtDc1_PgZ_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                            // LDNT1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
		{0xa580e000, []int{F_RlZtDc1_PgZ_AddrXSP, F_RlZtDc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNT1D { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNT1H: {
		{0xa480c000, []int{F_RlZtHc1_PgZ_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                            // LDNT1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
		{0xa480e000, []int{F_RlZtHc1_PgZ_AddrXSP, F_RlZtHc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNT1H { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDNT1W: {
		{0xa500c000, []int{F_RlZtSc1_PgZ_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                            // LDNT1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
		{0xa500e000, []int{F_RlZtSc1_PgZ_AddrXSP, F_RlZtSc1_PgZ_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // LDNT1W { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLDR: {
		{0x85804000, []int{F_Zt_AddrXSP, F_Zt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // LDR <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZLSL: {
		{0x04138000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},          // LSL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04209c00, FG_ZdT_ZnT_const, E_tszh_tszl_imm3_Zn_Zd},        // LSL <Zd>.<T>, <Zn>.<T>, #<const>
		{0x04208c00, FG_ZdT_ZnT_ZmD, E_size_Zm_ZnT_ZdT},               // LSL <Zd>.<T>, <Zn>.<T>, <Zm>.D
		{0x04038000, FG_ZdnT_PgM_ZdnT_const, E_tszh_Pg_tszl_imm3_Zdn}, // LSL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
		{0x041b8000, FG_ZdnT_PgM_ZdnT_ZmD, E_size_Pg_ZmD_Zdn},         // LSL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	},
	AZLSLR: {
		{0x04178000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSLR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZLSR: {
		{0x04118000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},                // LSR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04209400, FG_ZdT_ZnT_const, E_tszh_tszl_imm3_bias1_Zn_Zd},        // LSR <Zd>.<T>, <Zn>.<T>, #<const>
		{0x04208400, FG_ZdT_ZnT_ZmD, E_size_Zm_ZnT_ZdT},                     // LSR <Zd>.<T>, <Zn>.<T>, <Zm>.D
		{0x04018000, FG_ZdnT_PgM_ZdnT_const, E_tszh_Pg_tszl_imm3_bias1_Zdn}, // LSR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
		{0x04198000, FG_ZdnT_PgM_ZdnT_ZmD, E_size_Pg_ZmD_Zdn},               // LSR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	},
	AZLSRR: {
		{0x04158000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // LSRR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZMAD: {
		{0x0400c000, FG_ZdnT_PgM_ZmT_ZaT, E_size_Zm_Pg_Za_Zdn}, // MAD <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZMLA: {
		{0x04004000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Zm_Pg_Zn_Zda}, // MLA <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
	},
	AZMLS: {
		{0x04006000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Zm_Pg_Zn_Zda}, // MLS <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>
	},
	AZMOVPRFX: {
		{0x0420bc00, []int{F_Zd_Zn}, E_Rd_Rn},         // MOVPRFX <Zd>, <Zn>
		{0x04102000, FG_ZdT_PgZ_ZnT, E_size_Pg_Zn_Zd}, // MOVPRFX <Zd>.<T>, <Pg>/<ZM>, <Zn>.<T>
		{0x04112000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // MOVPRFX <Zd>.<T>, <Pg>/<ZM>, <Zn>.<T>
	},
	AZMSB: {
		{0x0400e000, FG_ZdnT_PgM_ZmT_ZaT, E_size_Zm_Pg_Za_Zdn}, // MSB <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	},
	AZMUL: {
		{0x04100000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // MUL <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x2530c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},    // MUL <Zdn>.<T>, <Zdn>.<T>, #<imm>
	},
	AZNEG: {
		{0x0417a000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NEG <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZNOT: {
		{0x041ea000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // NOT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZORR: {
		{0x04180000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // ORR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04603000, []int{F_ZdD_ZnD_ZmD}, E_Rd_Rn_Rm},       // ORR <Zd>.D, <Zn>.D, <Zm>.D
		{0x05000000, FG_ZdnT_ZdnT_const, E_imm13_Zdn},        // ORR <Zdn>.<T>, <Zdn>.<T>, #<const>
	},
	AZORV: {
		{0x04182000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // ORV <V><d>, <Pg>, <Zn>.<T>
	},
	AZPRFB: {
		{0x8400c000, []int{F_prfop_Pg_AddrXSPXm}, E_Rm_Pg_Rn_prfop},         // PRFB <prfop>, <Pg>, [<Xn|SP>, <Xm>]
		{0xc4608000, []int{F_prfop_Pg_AddrXSPZmD}, E_Zm_Pg_Rn_prfop},        // PRFB <prfop>, <Pg>, [<Xn|SP>, <Zm>.D]
		{0xc4600000, []int{F_prfop_Pg_AddrXSPZmDSXTW}, E_Zm_Pg_Rn_prfop},    // PRFB <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW]
		{0xc4200000, []int{F_prfop_Pg_AddrXSPZmDUXTW}, E_Zm_Pg_Rn_prfop},    // PRFB <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, UXTW]
		{0x85c00000, []int{F_prfop_Pg_AddrXSP}, E_Pg_Rn_prfop},              // PRFB <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0x85c00000, []int{F_prfop_Pg_AddrXSPImmMulVl}, E_imm6_Pg_Rn_prfop}, // PRFB <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc400e000, []int{F_prfop_Pg_AddrZnD}, E_Pg_Zn_prfop},              // PRFB <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0xc400e000, []int{F_prfop_Pg_AddrZnDImm}, E_imm5_Pg_Zn_prfop},      // PRFB <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0x84600000, []int{F_prfop_Pg_AddrXSPZmSSXTW}, E_Zm_Pg_Rn_prfop},    // PRFB <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW]
		{0x84200000, []int{F_prfop_Pg_AddrXSPZmSUXTW}, E_Zm_Pg_Rn_prfop},    // PRFB <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, UXTW]
		{0x8400e000, []int{F_prfop_Pg_AddrZnS}, E_Pg_Zn_prfop},              // PRFB <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
		{0x8400e000, []int{F_prfop_Pg_AddrZnSImm}, E_imm5_Pg_Zn_prfop},      // PRFB <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZPRFD: {
		{0x8580c000, []int{F_prfop_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_prfop},     // PRFD <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xc460e000, []int{F_prfop_Pg_AddrXSPZmDLSL3}, E_Zm_Pg_Rn_prfop},    // PRFD <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
		{0xc4606000, []int{F_prfop_Pg_AddrXSPZmDSXTW3}, E_Zm_Pg_Rn_prfop},   // PRFD <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #3]
		{0xc4206000, []int{F_prfop_Pg_AddrXSPZmDUXTW3}, E_Zm_Pg_Rn_prfop},   // PRFD <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, UXTW #3]
		{0x85c06000, []int{F_prfop_Pg_AddrXSP}, E_Pg_Rn_prfop},              // PRFD <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0x85c06000, []int{F_prfop_Pg_AddrXSPImmMulVl}, E_imm6_Pg_Rn_prfop}, // PRFD <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc580e000, []int{F_prfop_Pg_AddrZnD}, E_Pg_Zn_prfop},              // PRFD <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0xc580e000, []int{F_prfop_Pg_AddrZnDImm}, E_8imm5_Pg_Zn_prfop},     // PRFD <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0x84606000, []int{F_prfop_Pg_AddrXSPZmSSXTW3}, E_Zm_Pg_Rn_prfop},   // PRFD <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #3]
		{0x84206000, []int{F_prfop_Pg_AddrXSPZmSUXTW3}, E_Zm_Pg_Rn_prfop},   // PRFD <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, UXTW #3]
		{0x8580e000, []int{F_prfop_Pg_AddrZnS}, E_Pg_Zn_prfop},              // PRFD <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
		{0x8580e000, []int{F_prfop_Pg_AddrZnSImm}, E_8imm5_Pg_Zn_prfop},     // PRFD <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZPRFH: {
		{0x8480c000, []int{F_prfop_Pg_AddrXSPXmLSL1}, E_Rm_Pg_Rn_prfop},     // PRFH <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xc460a000, []int{F_prfop_Pg_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_prfop},    // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xc4602000, []int{F_prfop_Pg_AddrXSPZmDSXTW1}, E_Zm_Pg_Rn_prfop},   // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #1]
		{0xc4202000, []int{F_prfop_Pg_AddrXSPZmDUXTW1}, E_Zm_Pg_Rn_prfop},   // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, UXTW #1]
		{0x85c02000, []int{F_prfop_Pg_AddrXSP}, E_Pg_Rn_prfop},              // PRFH <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0x85c02000, []int{F_prfop_Pg_AddrXSPImmMulVl}, E_imm6_Pg_Rn_prfop}, // PRFH <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc480e000, []int{F_prfop_Pg_AddrZnD}, E_Pg_Zn_prfop},              // PRFH <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0xc480e000, []int{F_prfop_Pg_AddrZnDImm}, E_2imm5_Pg_Zn_prfop},     // PRFH <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0x84602000, []int{F_prfop_Pg_AddrXSPZmSSXTW1}, E_Zm_Pg_Rn_prfop},   // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #1]
		{0x84202000, []int{F_prfop_Pg_AddrXSPZmSUXTW1}, E_Zm_Pg_Rn_prfop},   // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, UXTW #1]
		{0x8480e000, []int{F_prfop_Pg_AddrZnS}, E_Pg_Zn_prfop},              // PRFH <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
		{0x8480e000, []int{F_prfop_Pg_AddrZnSImm}, E_2imm5_Pg_Zn_prfop},     // PRFH <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZPRFW: {
		{0x8500c000, []int{F_prfop_Pg_AddrXSPXmLSL2}, E_Rm_Pg_Rn_prfop},     // PRFW <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xc460c000, []int{F_prfop_Pg_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_prfop},    // PRFH <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xc4604000, []int{F_prfop_Pg_AddrXSPZmDSXTW2}, E_Zm_Pg_Rn_prfop},   // PRFW <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #2]
		{0xc4204000, []int{F_prfop_Pg_AddrXSPZmDUXTW2}, E_Zm_Pg_Rn_prfop},   // PRFW <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, UXTW #2]
		{0x85c04000, []int{F_prfop_Pg_AddrXSP}, E_Pg_Rn_prfop},              // PRFW <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0x85c04000, []int{F_prfop_Pg_AddrXSPImmMulVl}, E_imm6_Pg_Rn_prfop}, // PRFW <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xc500e000, []int{F_prfop_Pg_AddrZnD}, E_Pg_Zn_prfop},              // PRFH <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0xc500e000, []int{F_prfop_Pg_AddrZnDImm}, E_4imm5_Pg_Zn_prfop},     // PRFH <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
		{0x84604000, []int{F_prfop_Pg_AddrXSPZmSSXTW2}, E_Zm_Pg_Rn_prfop},   // PRFW <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #2]
		{0x84204000, []int{F_prfop_Pg_AddrXSPZmSUXTW2}, E_Zm_Pg_Rn_prfop},   // PRFW <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, UXTW #2]
		{0x8500e000, []int{F_prfop_Pg_AddrZnS}, E_Pg_Zn_prfop},              // PRFH <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
		{0x8500e000, []int{F_prfop_Pg_AddrZnSImm}, E_4imm5_Pg_Zn_prfop},     // PRFH <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZRBIT: {
		{0x05278000, FG_ZdT_PgM_ZnT, E_size_Pg_Zn_Zd}, // RBIT <Zd>.<T>, <Pg>/M, <Zn>.<T>
	},
	AZRDVL: {
		{0x04bf5000, []int{F_Xd_Imm}, E_imm6_Rd}, // RDVL <Xd>, #<imm>
	},
	AZREV: {
		{0x05383800, FG_ZdT_ZnT, E_size_Zn_Zd}, // REV <Zd>.<T>, <Zn>.<T>
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
	AZSDOT: {
		{0x44000000, FG_ZdT_ZnTb_ZmTb, E_size_Zm_Zn_Zda},     // SDOT <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
		{0x44a00000, []int{F_ZdS_ZnB_ZmBidx}, E_Zmi2_Zn_Zda}, // SDOT <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
		{0x44e00000, []int{F_ZdD_ZnH_ZmHidx}, E_Zmi1_Zn_Zda}, // SDOT <Zda>.D, <Zn>.H, <Zm>.H[<imm>]
	},
	AZSEL: {
		{0x0520c000, FG_ZdT_Pv_ZnT_ZmT, E_size_Zm_Pv_Zn_Zd}, // SEL <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
	},
	AZSETFFR: {
		{0x252c9000, FG_none, E_none}, // SETFFR
	},
	AZSMAX: {
		{0x04080000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x2528c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},    // SMAX <Zdn>.<T>, <Zdn>.<T>, #<imm>
	},
	AZSMAXV: {
		{0x04082000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // SMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZSMIN: {
		{0x040a0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x252ac000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},    // SMIN <Zdn>.<T>, <Zdn>.<T>, #<imm>
	},
	AZSMINV: {
		{0x040a2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // SMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZSMMLA: {
		{0x45009800, []int{F_ZdS_ZnB_ZmB}, E_Zm_Zn_Zda}, // SMMLA <Zda>.S, <Zn>.B, <Zm>.B (+FEAT_I8MM)
	},
	AZSMULH: {
		{0x04120000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // SMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZSPLICE: {
		{0x052c8000, FG_ZdnT_Pv_ZdnT_ZmT, E_size_Pv_Zm_Zdn}, // SPLICE <Zd>.<T>, <Pv>, { <Zn1>.<T>, <Zn2>.<T> }
	},
	AZSQADD: {
		{0x04201000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // SQADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2524c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // SQADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2524c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // SQADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZSQDECB: {
		{0x043ffbe0, []int{F_Xd}, E_Rdn},                          // SQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x043ff800, []int{F_Xd_pattern}, E_pattern_Rdn},          // SQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0430f800, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn}, // SQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECBW: {
		{0x042ffbe0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQDECB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x042ff800, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQDECB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0420f800, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQDECB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECD: {
		{0x04fffbe0, []int{F_Xd}, E_Rdn},                           // SQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04fff800, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04f0f800, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04efcbe0, []int{F_ZdD}, E_Rdn},                          // SQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04efc800, []int{F_ZdD_pattern}, E_pattern_Rdn},          // SQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04e0c800, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rdn}, // SQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECDW: {
		{0x04effbe0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQDECD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04eff800, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQDECD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04e0f800, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQDECD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECH: {
		{0x047ffbe0, []int{F_Xd}, E_Rdn},                           // SQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x047ff800, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0470f800, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x046fcbe0, []int{F_ZdH}, E_Rdn},                          // SQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x046fc800, []int{F_ZdH_pattern}, E_pattern_Rdn},          // SQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0460c800, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rdn}, // SQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECHW: {
		{0x046ffbe0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQDECH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x046ff800, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQDECH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0460f800, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQDECH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECP: {
		{0x252a8000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // SQDECP <Zdn>.<T>, <Pm>.<T>
	},
	AZSQDECW: {
		{0x04bffbe0, []int{F_Xd}, E_Rdn},                           // SQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04bff800, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04b0f800, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04afcbe0, []int{F_ZdS}, E_Rdn},                          // SQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04afc800, []int{F_ZdS_pattern}, E_pattern_Rdn},          // SQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04a0c800, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rdn}, // SQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZSQDECWW: {
		{0x04affbe0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQDECW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04aff800, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQDECW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04a0f800, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQDECW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCB: {
		{0x043ff3e0, []int{F_Xd}, E_Rdn},                          // SQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x043ff000, []int{F_Xd_pattern}, E_pattern_Rdn},          // SQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0430f000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn}, // SQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCBW: {
		{0x042ff3e0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQINCB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x042ff000, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQINCB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0420f000, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQINCB <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCD: {
		{0x04fff3e0, []int{F_Xd}, E_Rdn},                           // SQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04fff000, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04f0f000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04efc3e0, []int{F_ZdD}, E_Rdn},                          // SQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04efc000, []int{F_ZdD_pattern}, E_pattern_Rdn},          // SQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04e0c000, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rdn}, // SQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCDW: {
		{0x04eff3e0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQINCD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04eff000, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQINCD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04e0f000, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQINCD <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCH: {
		{0x047ff3e0, []int{F_Xd}, E_Rdn},                           // SQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x047ff000, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0470f000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x046fc3e0, []int{F_ZdH}, E_Rdn},                          // SQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x046fc000, []int{F_ZdH_pattern}, E_pattern_Rdn},          // SQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0460c000, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rdn}, // SQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCHW: {
		{0x046ff3e0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQINCH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x046ff000, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQINCH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0460f000, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQINCH <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCP: {
		{0x25288000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // SQINCP <Zdn>.<T>, <Pm>.<T>
	},
	AZSQINCW: {
		{0x04bff3e0, []int{F_Xd}, E_Rdn},                           // SQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04bff000, []int{F_Xd_pattern}, E_pattern_Rdn},           // SQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04b0f000, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // SQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04afc3e0, []int{F_ZdS}, E_Rdn},                          // SQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04afc000, []int{F_ZdS_pattern}, E_pattern_Rdn},          // SQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04a0c000, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rdn}, // SQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZSQINCWW: {
		{0x04aff3e0, []int{F_Xdn_Wdn}, E_XWdn},                          // SQINCW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04aff000, []int{F_Xdn_Wdn_pattern}, E_pattern_XWdn},          // SQINCW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04a0f000, []int{F_Xdn_Wdn_pattern_Imm}, E_imm4_pattern_XWdn}, // SQINCW <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZSQSUB: {
		{0x04201800, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // SQSUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2526c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // SQSUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2526c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // SQSUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZST1B: {
		{0xe4004000, FG_RlZtTc1_Pg_AddrXSPXm, E_size_Rm_Pg_Rn_Zt},                                       // ST1B { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
		{0xe400e000, FG_RlZtTc1_Pg_AddrXSPImmOpt_B, E_size_imm4_Pg_Rn_Zt},                               // ST1B { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe4008000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW, F_RlZtDc1_Pg_AddrXSPZmDUXTW}, E_Zm_xs_Pg_Rn_Zt}, // ST1B { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
		{0xe4408000, []int{F_RlZtSc1_Pg_AddrXSPZmSSXTW, F_RlZtSc1_Pg_AddrXSPZmSUXTW}, E_Zm_xs_Pg_Rn_Zt}, // ST1B { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod>]
		{0xe400a000, []int{F_RlZtDc1_Pg_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                     // ST1B { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
		{0xe440a000, []int{F_RlZtDc1_Pg_AddrZnD, F_RlZtDc1_Pg_AddrZnDImm}, E_imm5_Pg_Zn_Zt},             // ST1B { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
		{0xe460a000, []int{F_RlZtSc1_Pg_AddrZnS, F_RlZtSc1_Pg_AddrZnSImm}, E_imm5_Pg_Zn_Zt},             // ST1B { <Zt>.S }, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZST1D: {
		{0xe5e04000, []int{F_RlZtDc1_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                                    // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xe5e0e000, []int{F_RlZtDc1_Pg_AddrXSP, F_RlZtDc1_Pg_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt},          // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe5a08000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW3, F_RlZtDc1_Pg_AddrXSPZmDUXTW3}, E_Zm_xs_Pg_Rn_Zt}, // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #3]
		{0xe5808000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW, F_RlZtDc1_Pg_AddrXSPZmDUXTW}, E_Zm_xs_Pg_Rn_Zt},   // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
		{0xe5a0a000, []int{F_RlZtDc1_Pg_AddrXSPZmDLSL3}, E_Zm_Pg_Rn_Zt},                                   // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
		{0xe580a000, []int{F_RlZtDc1_Pg_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                       // ST1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
		{0xe5c0a000, []int{F_RlZtDc1_Pg_AddrZnD, F_RlZtDc1_Pg_AddrZnDImm}, E_8imm5_Pg_Zn_Zt},              // ST1D { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
	},
	AZST1H: {
		{0xe4804000, FG_RlZtTc1_Pg_AddrXSPXmLSL1, E_size_Rm_Pg_Rn_Zt},                                     // ST1H { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xe480e000, FG_RlZtTc1_Pg_AddrXSPImmOpt_H, E_size_imm4_Pg_Rn_Zt},                                 // ST1H { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe4a08000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW1, F_RlZtDc1_Pg_AddrXSPZmDUXTW1}, E_Zm_xs_Pg_Rn_Zt}, // ST1H { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #1]
		{0xe4808000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW, F_RlZtDc1_Pg_AddrXSPZmDUXTW}, E_Zm_xs_Pg_Rn_Zt},   // ST1H { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
		{0xe4e08000, []int{F_RlZtSc1_Pg_AddrXSPZmSSXTW1, F_RlZtSc1_Pg_AddrXSPZmSUXTW1}, E_Zm_xs_Pg_Rn_Zt}, // ST1H { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #1]
		{0xe4c08000, []int{F_RlZtSc1_Pg_AddrXSPZmSSXTW, F_RlZtSc1_Pg_AddrXSPZmSUXTW}, E_Zm_xs_Pg_Rn_Zt},   // ST1H { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod>]
		{0xe4a0a000, []int{F_RlZtDc1_Pg_AddrXSPZmDLSL1}, E_Zm_Pg_Rn_Zt},                                   // ST1H { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
		{0xe480a000, []int{F_RlZtDc1_Pg_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                       // ST1H { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
		{0xe4c0a000, []int{F_RlZtDc1_Pg_AddrZnD, F_RlZtDc1_Pg_AddrZnDImm}, E_2imm5_Pg_Zn_Zt},              // ST1H { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
		{0xe4e0a000, []int{F_RlZtSc1_Pg_AddrZnS, F_RlZtSc1_Pg_AddrZnSImm}, E_2imm5_Pg_Zn_Zt},              // ST1H { <Zt>.S }, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZST1W: {
		{0xe5404000, FG_RlZtTc1_Pg_AddrXSPXmLSL2, E_size_Rm_Pg_Rn_Zt},                                     // ST1W { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xe540e000, FG_RlZtTc1_Pg_AddrXSPImmOpt_W, E_size_imm4_Pg_Rn_Zt},                                 // ST1W { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
		{0xe5208000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW2, F_RlZtDc1_Pg_AddrXSPZmDUXTW2}, E_Zm_xs_Pg_Rn_Zt}, // ST1W { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #2]
		{0xe5008000, []int{F_RlZtDc1_Pg_AddrXSPZmDSXTW, F_RlZtDc1_Pg_AddrXSPZmDUXTW}, E_Zm_xs_Pg_Rn_Zt},   // ST1W { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
		{0xe5608000, []int{F_RlZtSc1_Pg_AddrXSPZmSSXTW2, F_RlZtSc1_Pg_AddrXSPZmSUXTW2}, E_Zm_xs_Pg_Rn_Zt}, // ST1W { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #2]
		{0xe5408000, []int{F_RlZtSc1_Pg_AddrXSPZmSSXTW, F_RlZtSc1_Pg_AddrXSPZmSUXTW}, E_Zm_xs_Pg_Rn_Zt},   // ST1W { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod>]
		{0xe520a000, []int{F_RlZtDc1_Pg_AddrXSPZmDLSL2}, E_Zm_Pg_Rn_Zt},                                   // ST1W { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
		{0xe500a000, []int{F_RlZtDc1_Pg_AddrXSPZmD}, E_Zm_Pg_Rn_Zt},                                       // ST1W { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
		{0xe540a000, []int{F_RlZtDc1_Pg_AddrZnD, F_RlZtDc1_Pg_AddrZnDImm}, E_4imm5_Pg_Zn_Zt},              // ST1W { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
		{0xe560a000, []int{F_RlZtSc1_Pg_AddrZnS, F_RlZtSc1_Pg_AddrZnSImm}, E_4imm5_Pg_Zn_Zt},              // ST1W { <Zt>.S }, <Pg>, [<Zn>.S{, #<imm>}]
	},
	AZST2B: {
		{0xe4206000, []int{F_RlZtBc2i1_Pg_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                  // ST2B { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>, <Xm>]
		{0xe430e000, []int{F_RlZtBc2i1_Pg_AddrXSP, F_RlZtBc2i1_Pg_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // ST2B { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST2D: {
		{0xe5a06000, []int{F_RlZtDc2i1_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                              // ST2D { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xe5b0e000, []int{F_RlZtDc2i1_Pg_AddrXSP, F_RlZtDc2i1_Pg_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // ST2D { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST2H: {
		{0xe4a06000, []int{F_RlZtHc2i1_Pg_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                              // ST2H { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xe4b0e000, []int{F_RlZtHc2i1_Pg_AddrXSP, F_RlZtHc2i1_Pg_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // ST2H { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST2W: {
		{0xe5206000, []int{F_RlZtSc2i1_Pg_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                              // ST2W { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xe530e000, []int{F_RlZtSc2i1_Pg_AddrXSP, F_RlZtSc2i1_Pg_AddrXSPImmMulVl}, E_2imm4_Pg_Rn_Zt}, // ST2W { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST3B: {
		{0xe4406000, []int{F_RlZtBc3i1_Pg_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                  // ST3B { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>, <Xm>]
		{0xe450e000, []int{F_RlZtBc3i1_Pg_AddrXSP, F_RlZtBc3i1_Pg_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // ST3B { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST3D: {
		{0xe5c06000, []int{F_RlZtDc3i1_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                              // ST3D { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xe5d0e000, []int{F_RlZtDc3i1_Pg_AddrXSP, F_RlZtDc3i1_Pg_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // ST3D { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST3H: {
		{0xe4c06000, []int{F_RlZtHc3i1_Pg_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                              // ST3H { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xe4d0e000, []int{F_RlZtHc3i1_Pg_AddrXSP, F_RlZtHc3i1_Pg_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // ST3H { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST3W: {
		{0xe5406000, []int{F_RlZtSc3i1_Pg_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                              // ST3W { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xe550e000, []int{F_RlZtSc3i1_Pg_AddrXSP, F_RlZtSc3i1_Pg_AddrXSPImmMulVl}, E_3imm4_Pg_Rn_Zt}, // ST3W { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST4B: {
		{0xe4606000, []int{F_RlZtBc4i1_Pg_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                                  // ST4B { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>, <Xm>]
		{0xe470e000, []int{F_RlZtBc4i1_Pg_AddrXSP, F_RlZtBc4i1_Pg_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // ST4B { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST4D: {
		{0xe5e06000, []int{F_RlZtDc4i1_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                              // ST4D { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xe5f0e000, []int{F_RlZtDc4i1_Pg_AddrXSP, F_RlZtDc4i1_Pg_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // ST4D { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST4H: {
		{0xe4e06000, []int{F_RlZtHc4i1_Pg_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                              // ST4H { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xe4f0e000, []int{F_RlZtHc4i1_Pg_AddrXSP, F_RlZtHc4i1_Pg_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // ST4H { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZST4W: {
		{0xe5606000, []int{F_RlZtSc4i1_Pg_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                              // ST4W { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xe570e000, []int{F_RlZtSc4i1_Pg_AddrXSP, F_RlZtSc4i1_Pg_AddrXSPImmMulVl}, E_4imm4_Pg_Rn_Zt}, // ST4W { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSTNT1B: {
		{0xe4006000, []int{F_RlZtBc1_Pg_AddrXSPXm}, E_Rm_Pg_Rn_Zt},                          // STNT1B { <Zt>.B }, <Pg>, [<Xn|SP>, <Xm>]
		{0xe410e000, []int{F_RlZtBc1_Pg_AddrXSP, F_RlZtBc1_Pg_AddrXSPImm}, E_imm4_Pg_Rn_Zt}, // STNT1B { <Zt>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSTNT1D: {
		{0xe5806000, []int{F_RlZtDc1_Pg_AddrXSPXmLSL3}, E_Rm_Pg_Rn_Zt},                           // STNT1D { <Zt>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
		{0xe590e000, []int{F_RlZtDc1_Pg_AddrXSP, F_RlZtDc1_Pg_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // STNT1D { <Zt>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSTNT1H: {
		{0xe4806000, []int{F_RlZtHc1_Pg_AddrXSPXmLSL1}, E_Rm_Pg_Rn_Zt},                           // STNT1H { <Zt>.H }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
		{0xe490e000, []int{F_RlZtHc1_Pg_AddrXSP, F_RlZtHc1_Pg_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // STNT1H { <Zt>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSTNT1W: {
		{0xe5006000, []int{F_RlZtSc1_Pg_AddrXSPXmLSL2}, E_Rm_Pg_Rn_Zt},                           // STNT1W { <Zt>.S }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
		{0xe510e000, []int{F_RlZtSc1_Pg_AddrXSP, F_RlZtSc1_Pg_AddrXSPImmMulVl}, E_imm4_Pg_Rn_Zt}, // STNT1W { <Zt>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSTR: {
		{0xe5804000, []int{F_Zt_AddrXSP, F_Zt_AddrXSPImmMulVl}, E_imm9h_imm9l_Rn_Zt}, // STR <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	},
	AZSUB: {
		{0x04010000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},     // SUB <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x04200400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // SUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2521c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // SUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2521c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // SUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZSUBR: {
		{0x04030000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn},     // SUBR <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x2523c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // SUBR <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2523c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // SUBR <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZSUDOT: {
		{0x44a01c00, []int{F_ZdS_ZnB_ZmBidx}, E_Zmi2_Zn_Zda}, // // SUDOT <Zda>.S, <Zn>.B, <Zm>.B[<imm>] (+FEAT_I8MM)
	},
	AZSUNPKHI: {
		{0x05313800, FG_ZdT_ZnTb, E_size_ZnTb_ZdT}, // SUNPKHI <Zd>.<T>, <Zn>.<Tb>
	},
	AZSUNPKLO: {
		{0x05303800, FG_ZdT_ZnTb, E_size_ZnTb_ZdT}, // SUNPKLO <Zd>.<T>, <Zn>.<Tb>
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
	AZTBL: {
		{0x05203000, []int{F_ZdB_RLZnB_ZmB, F_ZdH_RLZnH_ZmH, F_ZdS_RLZnS_ZmS, F_ZdD_RLZnD_ZmD}, E_size_Zm_RLZn_Zd}, // TBL <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
	},
	AZTRN1: {
		{0x05207000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // TRN1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a01800, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // TRN1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZTRN2: {
		{0x05207400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // TRN2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a01c00, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // TRN2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZUABD: {
		{0x040d0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UABD <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUADDV: {
		{0x04012000, []int{F_Dd_Pg_ZnB, F_Dd_Pg_ZnH, F_Dd_Pg_ZnS, F_Dd_Pg_ZnD}, E_size_Pg_Zn_Vd}, // UADDV <Dd>, <Pg>, <Zn>.<T>
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
	AZUDOT: {
		{0x44000400, FG_ZdT_ZnTb_ZmTb, E_size_Zm_Zn_Zda},     // UDOT <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
		{0x44a00400, []int{F_ZdS_ZnB_ZmBidx}, E_Zmi2_Zn_Zda}, // UDOT <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
		{0x44e00400, []int{F_ZdD_ZnH_ZmHidx}, E_Zmi1_Zn_Zda}, // UDOT <Zda>.D, <Zn>.H, <Zm>.H[<imm>]
	},
	AZUMAX: {
		{0x04090000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMAX <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x2529c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},    // UMAX <Zdn>.<T>, <Zdn>.<T>, #<imm>
	},
	AZUMAXV: {
		{0x04092000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // UMAXV <V><d>, <Pg>, <Zn>.<T>
	},
	AZUMIN: {
		{0x040b0000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMIN <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
		{0x252bc000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},    // UMIN <Zdn>.<T>, <Zdn>.<T>, #<imm>
	},
	AZUMINV: {
		{0x040b2000, FG_Vd_Pg_ZnT, E_size_Pg_Zn_Vd}, // UMINV <V><d>, <Pg>, <Zn>.<T>
	},
	AZUMMLA: {
		{0x45c09800, []int{F_ZdS_ZnB_ZmB}, E_Zm_Zn_Zda}, // UMMLA <Zda>.S, <Zn>.B, <Zm>.B (+FEAT_I8MM)
	},
	AZUMULH: {
		{0x04130000, FG_ZdnT_PgM_ZdnT_ZmT, E_size_Pg_Zm_Zdn}, // UMULH <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>
	},
	AZUQADD: {
		{0x04201400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // UQADD <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2525c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // UQADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2525c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // UQADD <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZUQDECB: {
		{0x043fffe0, []int{F_Xd}, E_Rdn},                          // UQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x043ffc00, []int{F_Xd_pattern}, E_pattern_Rdn},          // UQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0430fc00, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECB <Xdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECBW: {
		{0x042fffe0, []int{F_Wdn}, E_Rdn},                          // UQDECB <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x042ffc00, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQDECB <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0420fc00, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECB <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECD: {
		{0x04ffffe0, []int{F_Xd}, E_Rdn},                           // UQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04fffc00, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04f0fc00, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQDECD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04efcfe0, []int{F_ZdD}, E_Rdn},                          // UQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04efcc00, []int{F_ZdD_pattern}, E_pattern_Rdn},          // UQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04e0cc00, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECDW: {
		{0x04efffe0, []int{F_Wdn}, E_Rdn},                          // UQDECD <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04effc00, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQDECD <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04e0fc00, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECD <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECH: {
		{0x047fffe0, []int{F_Xd}, E_Rdn},                           // UQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x047ffc00, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0470fc00, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQDECH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x046fcfe0, []int{F_ZdH}, E_Rdn},                          // UQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x046fcc00, []int{F_ZdH_pattern}, E_pattern_Rdn},          // UQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0460cc00, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECHW: {
		{0x046fffe0, []int{F_Wdn}, E_Rdn},                          // UQDECH <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x046ffc00, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQDECH <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0460fc00, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECH <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECP: {
		{0x252b8000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // UQDECP <Zdn>.<T>, <Pm>.<T>
	},
	AZUQDECW: {
		{0x04bfffe0, []int{F_Xd}, E_Rdn},                           // UQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04bffc00, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04b0fc00, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQDECW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04afffe0, []int{F_ZdS}, E_Rdn},                          // UQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04affc00, []int{F_ZdS_pattern}, E_pattern_Rdn},          // UQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04a0cc00, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZUQDECWW: {
		{0x04afffe0, []int{F_Wdn}, E_Rdn},                          // UQDECW <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04affc00, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQDECW <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04a0fc00, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQDECW <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCB: {
		{0x043ff7e0, []int{F_Xd}, E_Rdn},                          // UQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x043ff400, []int{F_Xd_pattern}, E_pattern_Rdn},          // UQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0430f400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCB <Xdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCBW: {
		{0x042ff7e0, []int{F_Wdn}, E_Rdn},                          // UQINCB <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x042ff400, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQINCB <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0420f400, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCB <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCD: {
		{0x04fff7e0, []int{F_Xd}, E_Rdn},                           // UQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04fff400, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04f0f400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQINCD <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04efc7e0, []int{F_ZdD}, E_Rdn},                          // UQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04efc400, []int{F_ZdD_pattern}, E_pattern_Rdn},          // UQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
		{0x04e0c400, []int{F_ZdD_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCD <Zd>.D{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCDW: {
		{0x04eff7e0, []int{F_Wdn}, E_Rdn},                          // UQINCD <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04eff400, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQINCD <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04e0f400, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCD <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCH: {
		{0x047ff7e0, []int{F_Xd}, E_Rdn},                           // UQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x047ff400, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x0470f400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQINCH <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x046fc7e0, []int{F_ZdH}, E_Rdn},                          // UQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x046fc400, []int{F_ZdH_pattern}, E_pattern_Rdn},          // UQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
		{0x0460c400, []int{F_ZdH_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCH <Zd>.H{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCHW: {
		{0x046ff7e0, []int{F_Wdn}, E_Rdn},                          // UQINCH <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x046ff400, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQINCH <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x0460f400, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCH <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCP: {
		{0x25298000, FG_ZdnT_PmT, E_size_Pm_Zdn}, // UQINCP <Zdn>.<T>, <Pm>.<T>
	},
	AZUQINCW: {
		{0x04bff7e0, []int{F_Xd}, E_Rdn},                           // UQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04bff400, []int{F_Xd_pattern}, E_pattern_Rdn},           // UQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04b0f400, []int{F_Xd_pattern_Imm}, E_imm4_pattern_Rdn},  // UQINCW <Xdn>{, <pattern>{, MUL #<imm>}}
		{0x04afc7e0, []int{F_ZdS}, E_Rdn},                          // UQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04afc400, []int{F_ZdS_pattern}, E_pattern_Rdn},          // UQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
		{0x04a0c400, []int{F_ZdS_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCW <Zd>.S{, <pattern>{, MUL #<imm>}}
	},
	AZUQINCWW: {
		{0x04aff7e0, []int{F_Wdn}, E_Rdn},                          // UQINCW <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04aff400, []int{F_Wdn_pattern}, E_pattern_Rdn},          // UQINCW <Wdn>{, <pattern>{, MUL #<imm>}}
		{0x04a0f400, []int{F_Wdn_pattern_Imm}, E_imm4_pattern_Rdn}, // UQINCW <Wdn>{, <pattern>{, MUL #<imm>}}
	},
	AZUQSUB: {
		{0x04201c00, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},            // UQSUB <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x2527c000, FG_ZdnT_ZdnT_const, E_size_imm8_Zdn},        // UQSUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
		{0x2527c000, FG_ZdnT_ZdnT_imm_shift, E_size_sh_imm8_Zdn}, // UQSUB <Zdn>.<T>, <Zdn>.<T>, #imm{, <shift>}
	},
	AZUSDOT: {
		{0x44a01800, []int{F_ZdS_ZnB_ZmBidx}, E_Zmi2_Zn_Zda}, // USDOT <Zda>.S, <Zn>.B, <Zm>.B[<imm>] (+FEAT_I8MM)
		{0x44807800, []int{F_ZdS_ZnB_ZmB}, E_Zm_Zn_Zda},      // USDOT <Zda>.S, <Zn>.B, <Zm>.B (+FEAT_I8MM)
	},
	AZUSMMLA: {
		{0x45809800, []int{F_ZdS_ZnB_ZmB}, E_Zm_Zn_Zda}, // USMMLA <Zda>.S, <Zn>.B, <Zm>.B (+FEAT_I8MM)
	},
	AZUUNPKHI: {
		{0x05333800, FG_ZdT_ZnTb, E_size_ZnTb_ZdT}, // UUNPKHI <Zd>.<T>, <Zn>.<Tb>
	},
	AZUUNPKLO: {
		{0x05323800, FG_ZdT_ZnTb, E_size_ZnTb_ZdT}, // UUNPKLO <Zd>.<T>, <Zn>.<Tb>
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
	AZUZP1: {
		{0x05206800, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // UZP1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00800, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // UZP1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZUZP2: {
		{0x05206c00, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // UZP2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00c00, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // UZP2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZZIP1: {
		{0x05206000, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // ZIP1 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00000, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // ZIP1 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
	AZZIP2: {
		{0x05206400, FG_ZdT_ZnT_ZmT, E_size_Zm_Zn_Zd},  // ZIP2 <Zd>.<T>, <Zn>.<T>, <Zm>.<T>
		{0x05a00400, []int{F_ZdQ_ZnQ_ZmQ}, E_Rd_Rn_Rm}, // ZIP2 <Zd>.Q, <Zn>.Q, <Zm>.Q
	},
}

// Key into the format table.
const (
	F_none = iota
	F_Dd_Pg_ZnB
	F_Dd_Pg_ZnD
	F_Dd_Pg_ZnH
	F_Dd_Pg_ZnS
	F_Fd_Pg_ZnD
	F_Fd_Pg_ZnH
	F_Fd_Pg_ZnS
	F_Fdn_Pg_Fdn_ZmD
	F_Fdn_Pg_Fdn_ZmH
	F_Fdn_Pg_Fdn_ZmS
	F_PdB
	F_PdB_PgM_PnB
	F_PdB_PgZ
	F_PdB_PgZ_PnB
	F_PdB_PgZ_PnB_PmB
	F_PdB_PgZ_ZnB_Imm
	F_PdB_PgZ_ZnB_ZmB
	F_PdB_PgZ_ZnB_ZmD
	F_PdB_Pg_PnB_PmB
	F_PdB_PnB
	F_PdB_PnB_PmB
	F_PdB_Rn_Rm
	F_PdB_pattern
	F_PdD
	F_PdD_PgZ_ZnD_Imm
	F_PdD_PgZ_ZnD_ImmFP
	F_PdD_PgZ_ZnD_ZmD
	F_PdD_PnD
	F_PdD_PnD_PmD
	F_PdD_Rn_Rm
	F_PdD_pattern
	F_PdH
	F_PdH_PgZ_ZnH_Imm
	F_PdH_PgZ_ZnH_ImmFP
	F_PdH_PgZ_ZnH_ZmD
	F_PdH_PgZ_ZnH_ZmH
	F_PdH_PnB
	F_PdH_PnH
	F_PdH_PnH_PmH
	F_PdH_Rn_Rm
	F_PdH_pattern
	F_PdS
	F_PdS_PgZ_ZnS_Imm
	F_PdS_PgZ_ZnS_ImmFP
	F_PdS_PgZ_ZnS_ZmD
	F_PdS_PgZ_ZnS_ZmS
	F_PdS_PnS
	F_PdS_PnS_PmS
	F_PdS_Rn_Rm
	F_PdS_pattern
	F_PdnB_Pg_PdnB
	F_PdnB_Pv_PdnB
	F_PdnD_Pv_PdnD
	F_PdnH_Pv_PdnH
	F_PdnS_Pv_PdnS
	F_Pg_PnB
	F_PnB
	F_Pt_AddrXSP
	F_Pt_AddrXSPImmMulVl
	F_Rd_Pg_ZnB
	F_Rd_Pg_ZnD
	F_Rd_Pg_ZnH
	F_Rd_Pg_ZnS
	F_Rdn_Pg_Rdn_ZmB
	F_Rdn_Pg_Rdn_ZmD
	F_Rdn_Pg_Rdn_ZmH
	F_Rdn_Pg_Rdn_ZmS
	F_RlZtBc1_PgZ_AddrXSP
	F_RlZtBc1_PgZ_AddrXSPImmMulVl
	F_RlZtBc1_PgZ_AddrXSPXm
	F_RlZtBc1_Pg_AddrXSP
	F_RlZtBc1_Pg_AddrXSPImm
	F_RlZtBc1_Pg_AddrXSPXm
	F_RlZtBc1_Pg_AddrXSPXmLSL1
	F_RlZtBc1_Pg_AddrXSPXmLSL2
	F_RlZtBc2i1_PgZ_AddrXSP
	F_RlZtBc2i1_PgZ_AddrXSPImmMulVl
	F_RlZtBc2i1_PgZ_AddrXSPXm
	F_RlZtBc2i1_Pg_AddrXSP
	F_RlZtBc2i1_Pg_AddrXSPImmMulVl
	F_RlZtBc2i1_Pg_AddrXSPXm
	F_RlZtBc3i1_PgZ_AddrXSP
	F_RlZtBc3i1_PgZ_AddrXSPImmMulVl
	F_RlZtBc3i1_PgZ_AddrXSPXm
	F_RlZtBc3i1_Pg_AddrXSP
	F_RlZtBc3i1_Pg_AddrXSPImmMulVl
	F_RlZtBc3i1_Pg_AddrXSPXm
	F_RlZtBc4i1_PgZ_AddrXSP
	F_RlZtBc4i1_PgZ_AddrXSPImmMulVl
	F_RlZtBc4i1_PgZ_AddrXSPXm
	F_RlZtBc4i1_Pg_AddrXSP
	F_RlZtBc4i1_Pg_AddrXSPImmMulVl
	F_RlZtBc4i1_Pg_AddrXSPXm
	F_RlZtDc1_PgZ_AddrXSP
	F_RlZtDc1_PgZ_AddrXSPImmMulVl
	F_RlZtDc1_PgZ_AddrXSPXm
	F_RlZtDc1_PgZ_AddrXSPXmLSL1
	F_RlZtDc1_PgZ_AddrXSPXmLSL2
	F_RlZtDc1_PgZ_AddrXSPXmLSL3
	F_RlZtDc1_PgZ_AddrXSPZmD
	F_RlZtDc1_PgZ_AddrXSPZmDLSL1
	F_RlZtDc1_PgZ_AddrXSPZmDLSL2
	F_RlZtDc1_PgZ_AddrXSPZmDLSL3
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW1
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW2
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW3
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW1
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW2
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW3
	F_RlZtDc1_PgZ_AddrZnD
	F_RlZtDc1_PgZ_AddrZnDImm
	F_RlZtDc1_Pg_AddrXSP
	F_RlZtDc1_Pg_AddrXSPImm
	F_RlZtDc1_Pg_AddrXSPImmMulVl
	F_RlZtDc1_Pg_AddrXSPXm
	F_RlZtDc1_Pg_AddrXSPXmLSL1
	F_RlZtDc1_Pg_AddrXSPXmLSL2
	F_RlZtDc1_Pg_AddrXSPXmLSL3
	F_RlZtDc1_Pg_AddrXSPZmD
	F_RlZtDc1_Pg_AddrXSPZmDLSL1
	F_RlZtDc1_Pg_AddrXSPZmDLSL2
	F_RlZtDc1_Pg_AddrXSPZmDLSL3
	F_RlZtDc1_Pg_AddrXSPZmDSXTW
	F_RlZtDc1_Pg_AddrXSPZmDSXTW1
	F_RlZtDc1_Pg_AddrXSPZmDSXTW2
	F_RlZtDc1_Pg_AddrXSPZmDSXTW3
	F_RlZtDc1_Pg_AddrXSPZmDUXTW
	F_RlZtDc1_Pg_AddrXSPZmDUXTW1
	F_RlZtDc1_Pg_AddrXSPZmDUXTW2
	F_RlZtDc1_Pg_AddrXSPZmDUXTW3
	F_RlZtDc1_Pg_AddrZnD
	F_RlZtDc1_Pg_AddrZnDImm
	F_RlZtDc2i1_PgZ_AddrXSP
	F_RlZtDc2i1_PgZ_AddrXSPImmMulVl
	F_RlZtDc2i1_PgZ_AddrXSPXmLSL3
	F_RlZtDc2i1_Pg_AddrXSP
	F_RlZtDc2i1_Pg_AddrXSPImmMulVl
	F_RlZtDc2i1_Pg_AddrXSPXmLSL3
	F_RlZtDc3i1_PgZ_AddrXSP
	F_RlZtDc3i1_PgZ_AddrXSPImmMulVl
	F_RlZtDc3i1_PgZ_AddrXSPXmLSL3
	F_RlZtDc3i1_Pg_AddrXSP
	F_RlZtDc3i1_Pg_AddrXSPImmMulVl
	F_RlZtDc3i1_Pg_AddrXSPXmLSL3
	F_RlZtDc4i1_PgZ_AddrXSP
	F_RlZtDc4i1_PgZ_AddrXSPImmMulVl
	F_RlZtDc4i1_PgZ_AddrXSPXmLSL3
	F_RlZtDc4i1_Pg_AddrXSP
	F_RlZtDc4i1_Pg_AddrXSPImmMulVl
	F_RlZtDc4i1_Pg_AddrXSPXmLSL3
	F_RlZtHc1_PgZ_AddrXSP
	F_RlZtHc1_PgZ_AddrXSPImmMulVl
	F_RlZtHc1_PgZ_AddrXSPXm
	F_RlZtHc1_PgZ_AddrXSPXmLSL1
	F_RlZtHc1_Pg_AddrXSP
	F_RlZtHc1_Pg_AddrXSPImm
	F_RlZtHc1_Pg_AddrXSPImmMulVl
	F_RlZtHc1_Pg_AddrXSPXm
	F_RlZtHc1_Pg_AddrXSPXmLSL1
	F_RlZtHc1_Pg_AddrXSPXmLSL2
	F_RlZtHc2i1_PgZ_AddrXSP
	F_RlZtHc2i1_PgZ_AddrXSPImmMulVl
	F_RlZtHc2i1_PgZ_AddrXSPXmLSL1
	F_RlZtHc2i1_Pg_AddrXSP
	F_RlZtHc2i1_Pg_AddrXSPImmMulVl
	F_RlZtHc2i1_Pg_AddrXSPXmLSL1
	F_RlZtHc3i1_PgZ_AddrXSP
	F_RlZtHc3i1_PgZ_AddrXSPImmMulVl
	F_RlZtHc3i1_PgZ_AddrXSPXmLSL1
	F_RlZtHc3i1_Pg_AddrXSP
	F_RlZtHc3i1_Pg_AddrXSPImmMulVl
	F_RlZtHc3i1_Pg_AddrXSPXmLSL1
	F_RlZtHc4i1_PgZ_AddrXSP
	F_RlZtHc4i1_PgZ_AddrXSPImmMulVl
	F_RlZtHc4i1_PgZ_AddrXSPXmLSL1
	F_RlZtHc4i1_Pg_AddrXSP
	F_RlZtHc4i1_Pg_AddrXSPImmMulVl
	F_RlZtHc4i1_Pg_AddrXSPXmLSL1
	F_RlZtSc1_PgZ_AddrXSP
	F_RlZtSc1_PgZ_AddrXSPImmMulVl
	F_RlZtSc1_PgZ_AddrXSPXm
	F_RlZtSc1_PgZ_AddrXSPXmLSL1
	F_RlZtSc1_PgZ_AddrXSPXmLSL2
	F_RlZtSc1_PgZ_AddrXSPZmS
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW1
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW2
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW1
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW2
	F_RlZtSc1_PgZ_AddrZnS
	F_RlZtSc1_PgZ_AddrZnSImm
	F_RlZtSc1_Pg_AddrXSP
	F_RlZtSc1_Pg_AddrXSPImm
	F_RlZtSc1_Pg_AddrXSPImmMulVl
	F_RlZtSc1_Pg_AddrXSPXm
	F_RlZtSc1_Pg_AddrXSPXmLSL1
	F_RlZtSc1_Pg_AddrXSPXmLSL2
	F_RlZtSc1_Pg_AddrXSPZmS
	F_RlZtSc1_Pg_AddrXSPZmSSXTW
	F_RlZtSc1_Pg_AddrXSPZmSSXTW1
	F_RlZtSc1_Pg_AddrXSPZmSSXTW2
	F_RlZtSc1_Pg_AddrXSPZmSUXTW
	F_RlZtSc1_Pg_AddrXSPZmSUXTW1
	F_RlZtSc1_Pg_AddrXSPZmSUXTW2
	F_RlZtSc1_Pg_AddrZnS
	F_RlZtSc1_Pg_AddrZnSImm
	F_RlZtSc2i1_PgZ_AddrXSP
	F_RlZtSc2i1_PgZ_AddrXSPImmMulVl
	F_RlZtSc2i1_PgZ_AddrXSPXmLSL2
	F_RlZtSc2i1_Pg_AddrXSP
	F_RlZtSc2i1_Pg_AddrXSPImmMulVl
	F_RlZtSc2i1_Pg_AddrXSPXmLSL2
	F_RlZtSc3i1_PgZ_AddrXSP
	F_RlZtSc3i1_PgZ_AddrXSPImmMulVl
	F_RlZtSc3i1_PgZ_AddrXSPXmLSL2
	F_RlZtSc3i1_Pg_AddrXSP
	F_RlZtSc3i1_Pg_AddrXSPImmMulVl
	F_RlZtSc3i1_Pg_AddrXSPXmLSL2
	F_RlZtSc4i1_PgZ_AddrXSP
	F_RlZtSc4i1_PgZ_AddrXSPImmMulVl
	F_RlZtSc4i1_PgZ_AddrXSPXmLSL2
	F_RlZtSc4i1_Pg_AddrXSP
	F_RlZtSc4i1_Pg_AddrXSPImmMulVl
	F_RlZtSc4i1_Pg_AddrXSPXmLSL2
	F_Rn_Rm
	F_Vd_Pg_ZnB
	F_Vd_Pg_ZnD
	F_Vd_Pg_ZnH
	F_Vd_Pg_ZnS
	F_Vdn_Pg_Vdn_ZmB
	F_Vdn_Pg_Vdn_ZmD
	F_Vdn_Pg_Vdn_ZmH
	F_Vdn_Pg_Vdn_ZmS
	F_Wdn
	F_Wdn_pattern
	F_Wdn_pattern_Imm
	F_Xd_Imm
	F_Xd_Pg_PnB
	F_Xd_Pg_PnD
	F_Xd_Pg_PnH
	F_Xd_Pg_PnS
	F_Xd_Xn_Imm
	F_Xdn_PmB
	F_Xdn_PmB_Wdn
	F_Xdn_PmD
	F_Xdn_PmD_Wdn
	F_Xdn_PmH
	F_Xdn_PmH_Wdn
	F_Xdn_PmS
	F_Xdn_PmS_Wdn
	F_Xdn_Wdn
	F_Xdn_Wdn_pattern
	F_Xdn_Wdn_pattern_Imm
	F_ZdB_Imm
	F_ZdB_Imm_Imm
	F_ZdB_Imm_Wd
	F_ZdB_PgM_Imm
	F_ZdB_PgM_Imm_Imm
	F_ZdB_PgM_Vn
	F_ZdB_PgM_Wn
	F_ZdB_PgM_ZmB_ZaB
	F_ZdB_PgM_ZnB
	F_ZdB_PgZ_Imm
	F_ZdB_PgZ_Imm_Imm
	F_ZdB_PgZ_ZnB
	F_ZdB_Pv_ZnB_ZmB
	F_ZdB_RLZnB_ZmB
	F_ZdB_Rm
	F_ZdB_Vm
	F_ZdB_Wn
	F_ZdB_Wn_Imm
	F_ZdB_Wn_Wm
	F_ZdB_ZdB_const
	F_ZdB_ZdB_imm_shift
	F_ZdB_ZnB
	F_ZdB_ZnB_ZmB
	F_ZdB_ZnB_ZmD
	F_ZdB_ZnB_const
	F_ZdB_ZnBidx
	F_ZdB_const
	F_ZdD
	F_ZdD_AddrZnDZmD
	F_ZdD_AddrZnDZmDLSL
	F_ZdD_AddrZnDZmDSXTW
	F_ZdD_AddrZnDZmDUXTW
	F_ZdD_Imm
	F_ZdD_Imm_Imm
	F_ZdD_Imm_Xd
	F_ZdD_PgM_Imm
	F_ZdD_PgM_Imm_Imm
	F_ZdD_PgM_Vn
	F_ZdD_PgM_Xn
	F_ZdD_PgM_ZmD_ZaD
	F_ZdD_PgM_ZnD
	F_ZdD_PgM_ZnD_ZmD_Imm
	F_ZdD_PgM_ZnH
	F_ZdD_PgM_ZnS
	F_ZdD_PgM_const
	F_ZdD_PgZ_Imm
	F_ZdD_PgZ_Imm_Imm
	F_ZdD_PgZ_ZnD
	F_ZdD_Pg_ZnD
	F_ZdD_PmD
	F_ZdD_Pv_ZnD_ZmD
	F_ZdD_RLZnD_ZmD
	F_ZdD_Rm
	F_ZdD_Vm
	F_ZdD_Xn
	F_ZdD_Xn_Imm
	F_ZdD_Xn_Xm
	F_ZdD_ZdD_ZmD_imm
	F_ZdD_ZdD_const
	F_ZdD_ZdD_imm_shift
	F_ZdD_ZnD
	F_ZdD_ZnD_ZmD
	F_ZdD_ZnD_ZmDidx
	F_ZdD_ZnDidx
	F_ZdD_ZnH_ZmH
	F_ZdD_ZnH_ZmHidx
	F_ZdD_ZnS
	F_ZdD_Znd_const
	F_ZdD_const
	F_ZdD_const_FP
	F_ZdD_pattern
	F_ZdD_pattern_Imm
	F_ZdH
	F_ZdH_Imm
	F_ZdH_Imm_Imm
	F_ZdH_Imm_Wd
	F_ZdH_PgM_Imm
	F_ZdH_PgM_Imm_Imm
	F_ZdH_PgM_Vn
	F_ZdH_PgM_Wn
	F_ZdH_PgM_ZmH_ZaH
	F_ZdH_PgM_ZnD
	F_ZdH_PgM_ZnH
	F_ZdH_PgM_ZnH_ZmH_Imm
	F_ZdH_PgM_ZnS
	F_ZdH_PgM_const
	F_ZdH_PgZ_Imm
	F_ZdH_PgZ_Imm_Imm
	F_ZdH_PgZ_ZnH
	F_ZdH_PmH
	F_ZdH_Pv_ZnH_ZmH
	F_ZdH_RLZnH_ZmH
	F_ZdH_Rm
	F_ZdH_Vm
	F_ZdH_Wn
	F_ZdH_Wn_Imm
	F_ZdH_Wn_Wm
	F_ZdH_ZdH_ZmH_imm
	F_ZdH_ZdH_const
	F_ZdH_ZdH_imm_shift
	F_ZdH_ZnB
	F_ZdH_ZnH
	F_ZdH_ZnH_ZmD
	F_ZdH_ZnH_ZmH
	F_ZdH_ZnH_ZmHidx
	F_ZdH_ZnH_ZmHidx_const
	F_ZdH_ZnH_const
	F_ZdH_ZnHidx
	F_ZdH_const
	F_ZdH_const_FP
	F_ZdH_pattern
	F_ZdH_pattern_Imm
	F_ZdQ_ZnQ_ZmQ
	F_ZdQ_ZnQidx
	F_ZdS
	F_ZdS_AddrZnSZmS
	F_ZdS_AddrZnSZmSLSL
	F_ZdS_Imm
	F_ZdS_Imm_Imm
	F_ZdS_Imm_Wd
	F_ZdS_PgM_Imm
	F_ZdS_PgM_Imm_Imm
	F_ZdS_PgM_Vn
	F_ZdS_PgM_Wn
	F_ZdS_PgM_ZmS_ZaS
	F_ZdS_PgM_ZnD
	F_ZdS_PgM_ZnH
	F_ZdS_PgM_ZnS
	F_ZdS_PgM_ZnS_ZmS_Imm
	F_ZdS_PgM_const
	F_ZdS_PgZ_Imm
	F_ZdS_PgZ_Imm_Imm
	F_ZdS_PgZ_ZnS
	F_ZdS_Pg_ZnS
	F_ZdS_PmS
	F_ZdS_Pv_ZnS_ZmS
	F_ZdS_RLZnS_ZmS
	F_ZdS_Rm
	F_ZdS_Vm
	F_ZdS_Wn
	F_ZdS_Wn_Imm
	F_ZdS_Wn_Wm
	F_ZdS_ZdS_ZmS_imm
	F_ZdS_ZdS_const
	F_ZdS_ZdS_imm_shift
	F_ZdS_ZnB_ZmB
	F_ZdS_ZnB_ZmBidx
	F_ZdS_ZnH
	F_ZdS_ZnS
	F_ZdS_ZnS_ZmD
	F_ZdS_ZnS_ZmS
	F_ZdS_ZnS_ZmSidx
	F_ZdS_ZnS_ZmSidx_const
	F_ZdS_ZnS_const
	F_ZdS_ZnSidx
	F_ZdS_const
	F_ZdS_const_FP
	F_ZdS_pattern
	F_ZdS_pattern_Imm
	F_Zd_Zn
	F_ZdaS_ZnH_ZmH
	F_ZdaS_ZnH_ZmHidx
	F_ZdnB_PgM_ZdnB_ZmB
	F_ZdnB_PgM_ZdnB_ZmD
	F_ZdnB_PgM_ZdnB_const
	F_ZdnB_Pg_ZdnB_ZmB
	F_ZdnB_ZdnB_ZmB_imm
	F_ZdnD_PgM_ZdnD_ZmD
	F_ZdnD_PgM_ZdnD_ZmD_const
	F_ZdnD_PgM_ZdnD_const
	F_ZdnD_PgM_ZdnD_const_FP
	F_ZdnD_Pg_ZdnD_ZmD
	F_ZdnH_PgM_ZdnH_ZmD
	F_ZdnH_PgM_ZdnH_ZmH
	F_ZdnH_PgM_ZdnH_ZmH_const
	F_ZdnH_PgM_ZdnH_const
	F_ZdnH_PgM_ZdnH_const_FP
	F_ZdnH_Pg_ZdnH_ZmH
	F_ZdnS_PgM_ZdnS_ZmD
	F_ZdnS_PgM_ZdnS_ZmS
	F_ZdnS_PgM_ZdnS_ZmS_const
	F_ZdnS_PgM_ZdnS_const
	F_ZdnS_PgM_ZdnS_const_FP
	F_ZdnS_Pg_ZdnS_ZmS
	F_Zt_AddrXSP
	F_Zt_AddrXSPImmMulVl
	F_prfop_Pg_AddrXSP
	F_prfop_Pg_AddrXSPImmMulVl
	F_prfop_Pg_AddrXSPXm
	F_prfop_Pg_AddrXSPXmLSL1
	F_prfop_Pg_AddrXSPXmLSL2
	F_prfop_Pg_AddrXSPXmLSL3
	F_prfop_Pg_AddrXSPZmD
	F_prfop_Pg_AddrXSPZmDLSL1
	F_prfop_Pg_AddrXSPZmDLSL2
	F_prfop_Pg_AddrXSPZmDLSL3
	F_prfop_Pg_AddrXSPZmDSXTW
	F_prfop_Pg_AddrXSPZmDSXTW1
	F_prfop_Pg_AddrXSPZmDSXTW2
	F_prfop_Pg_AddrXSPZmDSXTW3
	F_prfop_Pg_AddrXSPZmDUXTW
	F_prfop_Pg_AddrXSPZmDUXTW1
	F_prfop_Pg_AddrXSPZmDUXTW2
	F_prfop_Pg_AddrXSPZmDUXTW3
	F_prfop_Pg_AddrXSPZmS
	F_prfop_Pg_AddrXSPZmSSXTW
	F_prfop_Pg_AddrXSPZmSSXTW1
	F_prfop_Pg_AddrXSPZmSSXTW2
	F_prfop_Pg_AddrXSPZmSSXTW3
	F_prfop_Pg_AddrXSPZmSUXTW
	F_prfop_Pg_AddrXSPZmSUXTW1
	F_prfop_Pg_AddrXSPZmSUXTW2
	F_prfop_Pg_AddrXSPZmSUXTW3
	F_prfop_Pg_AddrZnD
	F_prfop_Pg_AddrZnDImm
	F_prfop_Pg_AddrZnS
	F_prfop_Pg_AddrZnSImm

	// Format aliases
	F_Wdn_PmB        = F_Xdn_PmB
	F_Wdn_PmD        = F_Xdn_PmD
	F_Wdn_PmH        = F_Xdn_PmH
	F_Wdn_PmS        = F_Xdn_PmS
	F_Xd             = F_Wdn
	F_Xd_pattern     = F_Wdn_pattern
	F_Xd_pattern_Imm = F_Wdn_pattern_Imm
)

// Format groups, common patterns of associated instruction formats. E.g. expansion of the <T> generic lane size.
var FG_none = []int{F_none} // No arguments.
var FG_PdT = []int{F_PdB, F_PdH, F_PdS, F_PdD}
var FG_PdT_PgZ_ZnT_Imm = []int{F_PdB_PgZ_ZnB_Imm, F_PdH_PgZ_ZnH_Imm, F_PdS_PgZ_ZnS_Imm, F_PdD_PgZ_ZnD_Imm}
var FG_PdT_PgZ_ZnT_ImmFP = []int{F_PdH_PgZ_ZnH_ImmFP, F_PdS_PgZ_ZnS_ImmFP, F_PdD_PgZ_ZnD_ImmFP}
var FG_PdT_PgZ_ZnT_ZmD = []int{F_PdB_PgZ_ZnB_ZmD, F_PdH_PgZ_ZnH_ZmD, F_PdS_PgZ_ZnS_ZmD}
var FG_PdT_PgZ_ZnT_ZmT = []int{F_PdB_PgZ_ZnB_ZmB, F_PdH_PgZ_ZnH_ZmH, F_PdS_PgZ_ZnS_ZmS, F_PdD_PgZ_ZnD_ZmD}
var FG_PdT_PgZ_ZnT_ZmT_FP = []int{F_PdH_PgZ_ZnH_ZmH, F_PdS_PgZ_ZnS_ZmS, F_PdD_PgZ_ZnD_ZmD}
var FG_PdT_PnT = []int{F_PdB_PnB, F_PdH_PnH, F_PdS_PnS, F_PdD_PnD}
var FG_PdT_PnT_PmT = []int{F_PdB_PnB_PmB, F_PdH_PnH_PmH, F_PdS_PnS_PmS, F_PdD_PnD_PmD}
var FG_PdT_Rn_Rm = []int{F_PdB_Rn_Rm, F_PdH_Rn_Rm, F_PdS_Rn_Rm, F_PdD_Rn_Rm}
var FG_PdT_pattern = []int{F_PdB_pattern, F_PdH_pattern, F_PdS_pattern, F_PdD_pattern}
var FG_PdnT_Pv_PdnT = []int{F_PdnB_Pv_PdnB, F_PdnH_Pv_PdnH, F_PdnS_Pv_PdnS, F_PdnD_Pv_PdnD}
var FG_Rd_Pg_ZnT = []int{F_Rd_Pg_ZnB, F_Rd_Pg_ZnH, F_Rd_Pg_ZnS, F_Rd_Pg_ZnD}
var FG_Rdn_Pg_Rdn_ZmT = []int{F_Rdn_Pg_Rdn_ZmB, F_Rdn_Pg_Rdn_ZmH, F_Rdn_Pg_Rdn_ZmS, F_Rdn_Pg_Rdn_ZmD}
var FG_RlZtTc1_Pg_AddrXSPImmOpt_B = []int{F_RlZtBc1_Pg_AddrXSP, F_RlZtBc1_Pg_AddrXSPImm, F_RlZtHc1_Pg_AddrXSP, F_RlZtHc1_Pg_AddrXSPImm, F_RlZtSc1_Pg_AddrXSP, F_RlZtSc1_Pg_AddrXSPImm, F_RlZtDc1_Pg_AddrXSP, F_RlZtDc1_Pg_AddrXSPImm}
var FG_RlZtTc1_Pg_AddrXSPImmOpt_H = []int{F_RlZtHc1_Pg_AddrXSP, F_RlZtHc1_Pg_AddrXSPImm, F_RlZtSc1_Pg_AddrXSP, F_RlZtSc1_Pg_AddrXSPImm, F_RlZtDc1_Pg_AddrXSP, F_RlZtDc1_Pg_AddrXSPImm}
var FG_RlZtTc1_Pg_AddrXSPImmOpt_W = []int{F_RlZtSc1_Pg_AddrXSP, F_RlZtSc1_Pg_AddrXSPImm, F_RlZtDc1_Pg_AddrXSP, F_RlZtDc1_Pg_AddrXSPImm}
var FG_RlZtTc1_Pg_AddrXSPXm = []int{F_RlZtBc1_Pg_AddrXSPXm, F_RlZtHc1_Pg_AddrXSPXm, F_RlZtSc1_Pg_AddrXSPXm, F_RlZtDc1_Pg_AddrXSPXm}
var FG_RlZtTc1_Pg_AddrXSPXmLSL1 = []int{F_RlZtHc1_Pg_AddrXSPXmLSL1, F_RlZtSc1_Pg_AddrXSPXmLSL1, F_RlZtDc1_Pg_AddrXSPXmLSL1}
var FG_RlZtTc1_Pg_AddrXSPXmLSL2 = []int{F_RlZtSc1_Pg_AddrXSPXmLSL2, F_RlZtDc1_Pg_AddrXSPXmLSL2}
var FG_Vd_Pg_ZnT = []int{F_Vd_Pg_ZnB, F_Vd_Pg_ZnH, F_Vd_Pg_ZnS, F_Vd_Pg_ZnD}
var FG_Vd_Pg_ZnT_FP = []int{F_Fd_Pg_ZnH, F_Fd_Pg_ZnS, F_Fd_Pg_ZnD}
var FG_Vdn_Pg_Vdn_ZmT = []int{F_Vdn_Pg_Vdn_ZmB, F_Vdn_Pg_Vdn_ZmH, F_Vdn_Pg_Vdn_ZmS, F_Vdn_Pg_Vdn_ZmD}
var FG_Wdn_PmT = []int{F_Wdn_PmB, F_Wdn_PmH, F_Wdn_PmS, F_Wdn_PmD}
var FG_Xd_Pg_PnT = []int{F_Xd_Pg_PnB, F_Xd_Pg_PnH, F_Xd_Pg_PnS, F_Xd_Pg_PnD}
var FG_Xdn_PmT = []int{F_Xdn_PmB, F_Xdn_PmH, F_Xdn_PmS, F_Xdn_PmD}
var FG_Xdn_PmT_Wdn = []int{F_Xdn_PmB_Wdn, F_Xdn_PmH_Wdn, F_Xdn_PmS_Wdn, F_Xdn_PmD_Wdn}
var FG_ZdT_Imm_Imm = []int{F_ZdB_Imm_Imm, F_ZdH_Imm_Imm, F_ZdS_Imm_Imm, F_ZdD_Imm_Imm}
var FG_ZdT_Imm_XWm = []int{F_ZdB_Imm_Wd, F_ZdH_Imm_Wd, F_ZdS_Imm_Wd, F_ZdD_Imm_Xd}
var FG_ZdT_PgM_Imm = []int{F_ZdB_PgM_Imm, F_ZdH_PgM_Imm, F_ZdS_PgM_Imm, F_ZdD_PgM_Imm}
var FG_ZdT_PgM_Imm_Imm = []int{F_ZdB_PgM_Imm_Imm, F_ZdH_PgM_Imm_Imm, F_ZdS_PgM_Imm_Imm, F_ZdD_PgM_Imm_Imm}
var FG_ZdT_PgM_Vn = []int{F_ZdB_PgM_Vn, F_ZdH_PgM_Vn, F_ZdS_PgM_Vn, F_ZdD_PgM_Vn}
var FG_ZdT_PgM_WXn = []int{F_ZdB_PgM_Wn, F_ZdH_PgM_Wn, F_ZdS_PgM_Wn, F_ZdD_PgM_Xn}
var FG_ZdT_PgM_ZnT = []int{F_ZdB_PgM_ZnB, F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdT_PgM_ZnT_FP = []int{F_ZdH_PgM_ZnH, F_ZdS_PgM_ZnS, F_ZdD_PgM_ZnD}
var FG_ZdT_PgM_ZnT_ZmT_Imm_FP = []int{F_ZdH_PgM_ZnH_ZmH_Imm, F_ZdS_PgM_ZnS_ZmS_Imm, F_ZdD_PgM_ZnD_ZmD_Imm}
var FG_ZdT_PgM_const_FP = []int{F_ZdH_PgM_const, F_ZdS_PgM_const, F_ZdD_PgM_const}
var FG_ZdT_PgZ_Imm = []int{F_ZdB_PgZ_Imm, F_ZdH_PgZ_Imm, F_ZdS_PgZ_Imm, F_ZdD_PgZ_Imm}
var FG_ZdT_PgZ_Imm_Imm = []int{F_ZdB_PgZ_Imm_Imm, F_ZdH_PgZ_Imm_Imm, F_ZdS_PgZ_Imm_Imm, F_ZdD_PgZ_Imm_Imm}
var FG_ZdT_PgZ_ZnT = []int{F_ZdB_PgZ_ZnB, F_ZdH_PgZ_ZnH, F_ZdS_PgZ_ZnS, F_ZdD_PgZ_ZnD}
var FG_ZdT_Pv_ZnT_ZmT = []int{F_ZdB_Pv_ZnB_ZmB, F_ZdH_Pv_ZnH_ZmH, F_ZdS_Pv_ZnS_ZmS, F_ZdD_Pv_ZnD_ZmD}
var FG_ZdT_RnSP = []int{F_ZdB_Wn, F_ZdH_Wn, F_ZdS_Wn, F_ZdD_Xn}
var FG_ZdT_XWn_Imm = []int{F_ZdB_Wn_Imm, F_ZdH_Wn_Imm, F_ZdS_Wn_Imm, F_ZdD_Xn_Imm}
var FG_ZdT_XWn_XWm = []int{F_ZdB_Wn_Wm, F_ZdH_Wn_Wm, F_ZdS_Wn_Wm, F_ZdD_Xn_Xm}
var FG_ZdT_ZnT = []int{F_ZdB_ZnB, F_ZdH_ZnH, F_ZdS_ZnS, F_ZdD_ZnD}
var FG_ZdT_ZnT_FP = []int{F_ZdH_ZnH, F_ZdS_ZnS, F_ZdD_ZnD}
var FG_ZdT_ZnT_ZmD = []int{F_ZdB_ZnB_ZmD, F_ZdH_ZnH_ZmD, F_ZdS_ZnS_ZmD}
var FG_ZdT_ZnT_ZmT = []int{F_ZdB_ZnB_ZmB, F_ZdH_ZnH_ZmH, F_ZdS_ZnS_ZmS, F_ZdD_ZnD_ZmD}
var FG_ZdT_ZnT_ZmT_FP = []int{F_ZdH_ZnH_ZmH, F_ZdS_ZnS_ZmS, F_ZdD_ZnD_ZmD}
var FG_ZdT_ZnT_const = []int{F_ZdB_ZnB_const, F_ZdH_ZnH_const, F_ZdS_ZnS_const, F_ZdD_Znd_const}
var FG_ZdT_ZnTb = []int{F_ZdH_ZnB, F_ZdS_ZnH, F_ZdD_ZnS}
var FG_ZdT_ZnTb_ZmTb = []int{F_ZdS_ZnB_ZmB, F_ZdD_ZnH_ZmH}
var FG_ZdT_const = []int{F_ZdB_const, F_ZdH_const, F_ZdS_const, F_ZdD_const}
var FG_ZdT_const_FP = []int{F_ZdH_const_FP, F_ZdS_const_FP, F_ZdD_const_FP}
var FG_ZdT_imm = []int{F_ZdB_Imm, F_ZdH_Imm, F_ZdS_Imm, F_ZdD_Imm}
var FG_ZdT_imm_shift = []int{F_ZdB_Imm_Imm, F_ZdH_Imm_Imm, F_ZdS_Imm_Imm, F_ZdD_Imm_Imm}
var FG_ZdTq_ZnTqidx = []int{F_ZdB_ZnBidx, F_ZdH_ZnHidx, F_ZdS_ZnSidx, F_ZdD_ZnDidx, F_ZdQ_ZnQidx}
var FG_ZdnT_PgM_ZdnT_ZmD = []int{F_ZdnB_PgM_ZdnB_ZmD, F_ZdnH_PgM_ZdnH_ZmD, F_ZdnS_PgM_ZdnS_ZmD}
var FG_ZdnT_PgM_ZdnT_ZmT = []int{F_ZdnB_PgM_ZdnB_ZmB, F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}
var FG_ZdnT_PgM_ZdnT_ZmT_FP = []int{F_ZdnH_PgM_ZdnH_ZmH, F_ZdnS_PgM_ZdnS_ZmS, F_ZdnD_PgM_ZdnD_ZmD}
var FG_ZdnT_PgM_ZdnT_ZmT_const_FP = []int{F_ZdnH_PgM_ZdnH_ZmH_const, F_ZdnS_PgM_ZdnS_ZmS_const, F_ZdnD_PgM_ZdnD_ZmD_const}
var FG_ZdnT_PgM_ZdnT_const = []int{F_ZdnB_PgM_ZdnB_const, F_ZdnH_PgM_ZdnH_const, F_ZdnS_PgM_ZdnS_const, F_ZdnD_PgM_ZdnD_const}
var FG_ZdnT_PgM_ZdnT_const_FP = []int{F_ZdnH_PgM_ZdnH_const_FP, F_ZdnS_PgM_ZdnS_const_FP, F_ZdnD_PgM_ZdnD_const_FP}
var FG_ZdnT_PgM_ZmT_ZaT = []int{F_ZdB_PgM_ZmB_ZaB, F_ZdH_PgM_ZmH_ZaH, F_ZdS_PgM_ZmS_ZaS, F_ZdD_PgM_ZmD_ZaD}
var FG_ZdnT_PgM_ZmT_ZaT_FP = []int{F_ZdH_PgM_ZmH_ZaH, F_ZdS_PgM_ZmS_ZaS, F_ZdD_PgM_ZmD_ZaD}
var FG_ZdnT_Pg_ZdnT_ZmT = []int{F_ZdnB_Pg_ZdnB_ZmB, F_ZdnH_Pg_ZdnH_ZmH, F_ZdnS_Pg_ZdnS_ZmS, F_ZdnD_Pg_ZdnD_ZmD}
var FG_ZdnT_PmT = []int{F_ZdH_PmH, F_ZdS_PmS, F_ZdD_PmD}
var FG_ZdnT_Pv_ZdnT_ZmT = FG_ZdnT_Pg_ZdnT_ZmT
var FG_ZdnT_Rm = []int{F_ZdB_Rm, F_ZdH_Rm, F_ZdS_Rm, F_ZdD_Rm}
var FG_ZdnT_Vm = []int{F_ZdB_Vm, F_ZdH_Vm, F_ZdS_Vm, F_ZdD_Vm}
var FG_ZdnT_ZdnT_ZmT_imm_FP = []int{F_ZdH_ZdH_ZmH_imm, F_ZdS_ZdS_ZmS_imm, F_ZdD_ZdD_ZmD_imm}
var FG_ZdnT_ZdnT_const = []int{F_ZdB_ZdB_const, F_ZdH_ZdH_const, F_ZdS_ZdS_const, F_ZdD_ZdD_const}
var FG_ZdnT_ZdnT_imm_shift = []int{F_ZdB_ZdB_imm_shift, F_ZdH_ZdH_imm_shift, F_ZdS_ZdS_imm_shift, F_ZdD_ZdD_imm_shift}

// The format table holds a representation of the operand syntax for an instruction.
var formats = map[int]format{
	F_none:                          []int{},                                                                                        // No arguments.
	F_Dd_Pg_ZnB:                     []int{REG_V, REG_P, REG_Z | EXT_B},                                                             // <Dd>, <Pg>, <Zn>.B
	F_Dd_Pg_ZnD:                     []int{REG_V, REG_P, REG_Z | EXT_D},                                                             // <Dd>, <Pg>, <Zn>.D
	F_Dd_Pg_ZnH:                     []int{REG_V, REG_P, REG_Z | EXT_H},                                                             // <Dd>, <Pg>, <Zn>.H
	F_Dd_Pg_ZnS:                     []int{REG_V, REG_P, REG_Z | EXT_S},                                                             // <Dd>, <Pg>, <Zn>.S
	F_Fd_Pg_ZnD:                     []int{REG_F, REG_P, REG_Z | EXT_D},                                                             // <V><d>, <Pg>, <Zn>.D
	F_Fd_Pg_ZnH:                     []int{REG_F, REG_P, REG_Z | EXT_H},                                                             // <V><d>, <Pg>, <Zn>.H
	F_Fd_Pg_ZnS:                     []int{REG_F, REG_P, REG_Z | EXT_S},                                                             // <V><d>, <Pg>, <Zn>.S
	F_Fdn_Pg_Fdn_ZmD:                []int{REG_F, REG_P, REG_F, REG_Z | EXT_D},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Fdn_Pg_Fdn_ZmH:                []int{REG_F, REG_P, REG_F, REG_Z | EXT_H},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Fdn_Pg_Fdn_ZmS:                []int{REG_F, REG_P, REG_F, REG_Z | EXT_S},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_PdB:                           []int{REG_P | EXT_B},                                                                           // <Pd>.B
	F_PdB_PgM_PnB:                   []int{REG_P | EXT_B, REG_P | EXT_MERGING, REG_P | EXT_B},                                       // <Pd>.B, <Pg>/Z, <Pn>.B
	F_PdB_PgZ:                       []int{REG_P | EXT_B, REG_P | EXT_ZEROING},                                                      // <Pd>.B, <Pg>/Z
	F_PdB_PgZ_PnB:                   []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_P | EXT_B},                                       // <Pd>.B, <Pg>/Z, <Pn>.B
	F_PdB_PgZ_PnB_PmB:               []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_P | EXT_B, REG_P | EXT_B},                        // <Pd>.B, <Pg>/Z, <Pn>.B, <Pm>.B
	F_PdB_PgZ_ZnB_Imm:               []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, IMM | IMM_INT},                        // <Pd>.B, <Pg>/Z, <Zn>.B, #<imm>
	F_PdB_PgZ_ZnB_ZmB:               []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, REG_Z | EXT_B},                        // <Pd>.B, <Pg>/Z, <Zn>.B, <Zm>.B
	F_PdB_PgZ_ZnB_ZmD:               []int{REG_P | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B, REG_Z | EXT_D},                        // <Pd>.B, <Pg>/Z, <Zn>.B, <Zm>.D
	F_PdB_Pg_PnB_PmB:                []int{REG_P | EXT_B, REG_P, REG_P | EXT_B, REG_P | EXT_B},                                      // <Pd>.B, <Pg>, <Pn>.B, <Pm>.B
	F_PdB_PnB:                       []int{REG_P | EXT_B, REG_P | EXT_B},                                                            // <Pd>.B, <Pn>.B
	F_PdB_PnB_PmB:                   []int{REG_P | EXT_B, REG_P | EXT_B, REG_P | EXT_B},                                             // <Pd>.B, <Pn>.B, <Pm>.B
	F_PdB_Rn_Rm:                     []int{REG_P | EXT_B, REG_R, REG_R},                                                             // <Pd>.B, <R><n>, <R><m>
	F_PdB_pattern:                   []int{REG_P | EXT_B, PATTERN},                                                                  // <Pd>.B, <pattern>
	F_PdD:                           []int{REG_P | EXT_D},                                                                           // <Pd>.D
	F_PdD_PgZ_ZnD_Imm:               []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, IMM | IMM_INT},                        // <Pd>.D, <Pg>/Z, <Zn>.D, #<imm>
	F_PdD_PgZ_ZnD_ImmFP:             []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, IMM | IMM_FLOAT},                      // <Pd>.D, <Pg>/Z, <Zn>.D
	F_PdD_PgZ_ZnD_ZmD:               []int{REG_P | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D, REG_Z | EXT_D},                        // <Pd>.D, <Pg>/Z, <Zn>.D, <Zm>.D
	F_PdD_PnD:                       []int{REG_P | EXT_D, REG_P | EXT_D},                                                            // <Pd>.D, <Pn>.D
	F_PdD_PnD_PmD:                   []int{REG_P | EXT_D, REG_P | EXT_D, REG_P | EXT_D},                                             // <Pd>.D, <Pn>.D, <Pm>.D
	F_PdD_Rn_Rm:                     []int{REG_P | EXT_D, REG_R, REG_R},                                                             // <Pd>.B, <R><n>, <R><m>
	F_PdD_pattern:                   []int{REG_P | EXT_D, PATTERN},                                                                  // <Pd>.D, <pattern>
	F_PdH:                           []int{REG_P | EXT_H},                                                                           // <Pd>.H
	F_PdH_PgZ_ZnH_Imm:               []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, IMM | IMM_INT},                        // <Pd>.H, <Pg>/Z, <Zn>.H, #<imm>
	F_PdH_PgZ_ZnH_ImmFP:             []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, IMM | IMM_FLOAT},                      // <Pd>.H, <Pg>/Z, <Zn>.H
	F_PdH_PgZ_ZnH_ZmD:               []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, REG_Z | EXT_D},                        // <Pd>.H, <Pg>/Z, <Zn>.H, <Zm>.D
	F_PdH_PgZ_ZnH_ZmH:               []int{REG_P | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Pd>.H, <Pg>/Z, <Zn>.H, <Zm>.H
	F_PdH_PnB:                       []int{REG_P | EXT_H, REG_P | EXT_B},                                                            // <Pd>.H, <Pn>.B
	F_PdH_PnH:                       []int{REG_P | EXT_H, REG_P | EXT_H},                                                            // <Pd>.H, <Pn>.H
	F_PdH_PnH_PmH:                   []int{REG_P | EXT_H, REG_P | EXT_H, REG_P | EXT_H},                                             // <Pd>.H, <Pn>.H, <Pm>.H
	F_PdH_Rn_Rm:                     []int{REG_P | EXT_H, REG_R, REG_R},                                                             // <Pd>.B, <R><n>, <R><m>
	F_PdH_pattern:                   []int{REG_P | EXT_H, PATTERN},                                                                  // <Pd>.H, <pattern>
	F_PdS:                           []int{REG_P | EXT_S},                                                                           // <Pd>.S
	F_PdS_PgZ_ZnS_Imm:               []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, IMM | IMM_INT},                        // <Pd>.S, <Pg>/Z, <Zn>.S, #<imm>
	F_PdS_PgZ_ZnS_ImmFP:             []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, IMM | IMM_FLOAT},                      // <Pd>.S, <Pg>/Z, <Zn>.S
	F_PdS_PgZ_ZnS_ZmD:               []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, REG_Z | EXT_D},                        // <Pd>.S, <Pg>/Z, <Zn>.S, <Zm>.D
	F_PdS_PgZ_ZnS_ZmS:               []int{REG_P | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S, REG_Z | EXT_S},                        // <Pd>.S, <Pg>/Z, <Zn>.S, <Zm>.S
	F_PdS_PnS:                       []int{REG_P | EXT_S, REG_P | EXT_S},                                                            // <Pd>.S, <Pn>.S
	F_PdS_PnS_PmS:                   []int{REG_P | EXT_S, REG_P | EXT_S, REG_P | EXT_S},                                             // <Pd>.S, <Pn>.S, <Pm>.S
	F_PdS_Rn_Rm:                     []int{REG_P | EXT_S, REG_R, REG_R},                                                             // <Pd>.B, <R><n>, <R><m>
	F_PdS_pattern:                   []int{REG_P | EXT_S, PATTERN},                                                                  // <Pd>.S, <pattern>
	F_PdnB_Pg_PdnB:                  []int{REG_P | EXT_B, REG_P, REG_P | EXT_B},                                                     // <Pdn>.B, <Pg>, <Pdn>.B
	F_PdnB_Pv_PdnB:                  []int{REG_P | EXT_B, REG_P, REG_P | EXT_B},                                                     // <Pdn>.B, <Pv>, <Pdn>.B
	F_PdnD_Pv_PdnD:                  []int{REG_P | EXT_D, REG_P, REG_P | EXT_D},                                                     // <Pdn>.D, <Pv>, <Pdn>.D
	F_PdnH_Pv_PdnH:                  []int{REG_P | EXT_H, REG_P, REG_P | EXT_H},                                                     // <Pdn>.H, <Pv>, <Pdn>.H
	F_PdnS_Pv_PdnS:                  []int{REG_P | EXT_S, REG_P, REG_P | EXT_S},                                                     // <Pdn>.S, <Pv>, <Pdn>.S
	F_Pg_PnB:                        []int{REG_P, REG_P | EXT_B},                                                                    // <Pg>, <Pn>.B
	F_PnB:                           []int{REG_P | EXT_B},                                                                           // <Pn>.B
	F_Pt_AddrXSP:                    []int{REG_P, MEM_ADDR | MEM_RSP},                                                               // <Zt>, [<Xn|SP>]
	F_Pt_AddrXSPImmMulVl:            []int{REG_P, MEM_ADDR | MEM_RSP_IMM},                                                           // <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_Rd_Pg_ZnB:                     []int{REG_R, REG_P, REG_Z | EXT_B},                                                             // <Wd>, <Pg>, <Zn>.B
	F_Rd_Pg_ZnD:                     []int{REG_R, REG_P, REG_Z | EXT_D},                                                             // <Xd>, <Pg>, <Zn>.D
	F_Rd_Pg_ZnH:                     []int{REG_R, REG_P, REG_Z | EXT_H},                                                             // <Wd>, <Pg>, <Zn>.H
	F_Rd_Pg_ZnS:                     []int{REG_R, REG_P, REG_Z | EXT_S},                                                             // <Wd>, <Pg>, <Zn>.S
	F_Rdn_Pg_Rdn_ZmB:                []int{REG_R, REG_P, REG_R, REG_Z | EXT_B},                                                      // <Wdn>, <Pg>, <Wdn>, <Zm>.B
	F_Rdn_Pg_Rdn_ZmD:                []int{REG_R, REG_P, REG_R, REG_Z | EXT_D},                                                      // <Xdn>, <Pg>, <Xdn>, <Zm>.D
	F_Rdn_Pg_Rdn_ZmH:                []int{REG_R, REG_P, REG_R, REG_Z | EXT_H},                                                      // <Wdn>, <Pg>, <Wdn>, <Zm>.H
	F_Rdn_Pg_Rdn_ZmS:                []int{REG_R, REG_P, REG_R, REG_Z | EXT_S},                                                      // <Wdn>, <Pg>, <Wdn>, <Zm>.S
	F_RlZtBc1_PgZ_AddrXSP:           []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},                  // { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc1_PgZ_AddrXSPImmMulVl:   []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},              // { <Zt>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc1_PgZ_AddrXSPXm:         []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},                // { <Zt>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtBc1_Pg_AddrXSP:            []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP},                                // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc1_Pg_AddrXSPImm:         []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc1_Pg_AddrXSPXm:          []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R},                              // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtBc1_Pg_AddrXSPXmLSL1:      []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtBc1_Pg_AddrXSPXmLSL2:      []int{REGLIST_Z | EXT_B | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
	F_RlZtBc2i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc2i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc2i1_PgZ_AddrXSPXm:       []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},      // { <Zt1>.B, <Zt2>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtBc2i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc2i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc2i1_Pg_AddrXSPXm:        []int{REGLIST_Z | EXT_B | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R},                    // { <Zt1>.B, <Zt2>.B }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtBc3i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc3i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc3i1_PgZ_AddrXSPXm:       []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},      // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtBc3i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc3i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc3i1_Pg_AddrXSPXm:        []int{REGLIST_Z | EXT_B | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R},                    // { <Zt1>.B, <Zt2>.B, <Zt3>.B }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtBc4i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc4i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc4i1_PgZ_AddrXSPXm:       []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},      // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtBc4i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc4i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtBc4i1_Pg_AddrXSPXm:        []int{REGLIST_Z | EXT_B | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R},                    // { <Zt1>.B, <Zt2>.B, <Zt3>.B, <Zt4>.B }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtDc1_PgZ_AddrXSP:           []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},                  // { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc1_PgZ_AddrXSPImmMulVl:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},              // { <Zt>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc1_PgZ_AddrXSPXm:         []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},                // { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtDc1_PgZ_AddrXSPXmLSL1:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1},           // { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtDc1_PgZ_AddrXSPXmLSL2:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL2},           // { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #2]
	F_RlZtDc1_PgZ_AddrXSPXmLSL3:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL3},           // { <Zt>.D }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #3]
	F_RlZtDc1_PgZ_AddrXSPZmD:        []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD},               // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
	F_RlZtDc1_PgZ_AddrXSPZmDLSL1:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_LSL1},          // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
	F_RlZtDc1_PgZ_AddrXSPZmDLSL2:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_LSL2},          // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
	F_RlZtDc1_PgZ_AddrXSPZmDLSL3:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_LSL3},          // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW1:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_SXTW1},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #1]
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW2:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_SXTW2},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #2]
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW3:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_SXTW3},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #3]
	F_RlZtDc1_PgZ_AddrXSPZmDSXTW:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_SXTW},          // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW1:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_UXTW1},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #1]
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW2:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_UXTW2},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #2]
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW3:   []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_UXTW3},         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #3]
	F_RlZtDc1_PgZ_AddrXSPZmDUXTW:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZD_UXTW},          // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
	F_RlZtDc1_PgZ_AddrZnD:           []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_ZD},                   // { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	F_RlZtDc1_PgZ_AddrZnDImm:        []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_ZD_IMM},               // { <Zt>.D }, <Pg>/Z, [<Zn>.D{, #<imm>}]
	F_RlZtDc1_Pg_AddrXSP:            []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP},                                // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc1_Pg_AddrXSPImm:         []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc1_Pg_AddrXSPImmMulVl:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc1_Pg_AddrXSPXm:          []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R},                              // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtDc1_Pg_AddrXSPXmLSL1:      []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtDc1_Pg_AddrXSPXmLSL2:      []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
	F_RlZtDc1_Pg_AddrXSPXmLSL3:      []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL3},                         // { <Zt>.D }, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
	F_RlZtDc1_Pg_AddrXSPZmD:         []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD},                             // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D]
	F_RlZtDc1_Pg_AddrXSPZmDLSL1:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL1},                        // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
	F_RlZtDc1_Pg_AddrXSPZmDLSL2:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL2},                        // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
	F_RlZtDc1_Pg_AddrXSPZmDLSL3:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL3},                        // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
	F_RlZtDc1_Pg_AddrXSPZmDSXTW1:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW1},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #1]
	F_RlZtDc1_Pg_AddrXSPZmDSXTW2:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW2},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #2]
	F_RlZtDc1_Pg_AddrXSPZmDSXTW3:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW3},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #3]
	F_RlZtDc1_Pg_AddrXSPZmDSXTW:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW},                        // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
	F_RlZtDc1_Pg_AddrXSPZmDUXTW1:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW1},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #1]
	F_RlZtDc1_Pg_AddrXSPZmDUXTW2:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW2},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #2]
	F_RlZtDc1_Pg_AddrXSPZmDUXTW3:    []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW3},                       // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod> #3]
	F_RlZtDc1_Pg_AddrXSPZmDUXTW:     []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW},                        // { <Zt>.D }, <Pg>, [<Xn|SP>, <Zm>.D, <mod>]
	F_RlZtDc1_Pg_AddrZnD:            []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_ZD},                                 // { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
	F_RlZtDc1_Pg_AddrZnDImm:         []int{REGLIST_Z | EXT_D | RL_COUNT1, REG_P, MEM_ADDR | MEM_ZD_IMM},                             // { <Zt>.D }, <Pg>, [<Zn>.D{, #<imm>}]
	F_RlZtDc2i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc2i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc2i1_PgZ_AddrXSPXmLSL3:   []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL3}, // { <Zt1>.D, <Zt2>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtDc2i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc2i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc2i1_Pg_AddrXSPXmLSL3:    []int{REGLIST_Z | EXT_D | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL3},               // { <Zt1>.D, <Zt2>.D }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtDc3i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc3i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc3i1_PgZ_AddrXSPXmLSL3:   []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL3}, // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtDc3i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc3i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc3i1_Pg_AddrXSPXmLSL3:    []int{REGLIST_Z | EXT_D | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL3},               // { <Zt1>.D, <Zt2>.D, <Zt3>.D }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtDc4i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc4i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc4i1_PgZ_AddrXSPXmLSL3:   []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL3}, // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtDc4i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc4i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtDc4i1_Pg_AddrXSPXmLSL3:    []int{REGLIST_Z | EXT_D | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL3},               // { <Zt1>.D, <Zt2>.D, <Zt3>.D, <Zt4>.D }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtHc1_PgZ_AddrXSP:           []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},                  // { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc1_PgZ_AddrXSPImmMulVl:   []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},              // { <Zt>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc1_PgZ_AddrXSPXm:         []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},                // { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtHc1_PgZ_AddrXSPXmLSL1:     []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1},           // { <Zt>.H }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtHc1_Pg_AddrXSP:            []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP},                                // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc1_Pg_AddrXSPImm:         []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc1_Pg_AddrXSPImmMulVl:    []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc1_Pg_AddrXSPXm:          []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R},                              // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtHc1_Pg_AddrXSPXmLSL1:      []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtHc1_Pg_AddrXSPXmLSL2:      []int{REGLIST_Z | EXT_H | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
	F_RlZtHc2i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc2i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc2i1_PgZ_AddrXSPXmLSL1:   []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1}, // { <Zt1>.H, <Zt2>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtHc2i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc2i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc2i1_Pg_AddrXSPXmLSL1:    []int{REGLIST_Z | EXT_H | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},               // { <Zt1>.H, <Zt2>.H }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtHc3i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc3i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc3i1_PgZ_AddrXSPXmLSL1:   []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1}, // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtHc3i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc3i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc3i1_Pg_AddrXSPXmLSL1:    []int{REGLIST_Z | EXT_H | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},               // { <Zt1>.H, <Zt2>.H, <Zt3>.H }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtHc4i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc4i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc4i1_PgZ_AddrXSPXmLSL1:   []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1}, // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtHc4i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc4i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtHc4i1_Pg_AddrXSPXmLSL1:    []int{REGLIST_Z | EXT_H | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},               // { <Zt1>.H, <Zt2>.H, <Zt3>.H, <Zt4>.H }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtSc1_PgZ_AddrXSP:           []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},                  // { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc1_PgZ_AddrXSPImmMulVl:   []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},              // { <Zt>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc1_PgZ_AddrXSPXm:         []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R},                // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtSc1_PgZ_AddrXSPXmLSL1:     []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL1},           // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtSc1_PgZ_AddrXSPXmLSL2:     []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL2},           // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtSc1_PgZ_AddrXSPZmS:        []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS},               // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S]
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW1:   []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_SXTW1},         // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW2:   []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_SXTW2},         // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #2]
	F_RlZtSc1_PgZ_AddrXSPZmSSXTW:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_SXTW},          // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW1:   []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_UXTW1},         // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #1]
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW2:   []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_UXTW2},         // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod> #2]
	F_RlZtSc1_PgZ_AddrXSPZmSUXTW:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_ZS_UXTW},          // { <Zt>.S }, <Pg>/Z, [<Xn|SP>, <Zm>.S, <mod>]
	F_RlZtSc1_PgZ_AddrZnS:           []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_ZS},                   // { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	F_RlZtSc1_PgZ_AddrZnSImm:        []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_ZS_IMM},               // { <Zt>.S }, <Pg>/Z, [<Zn>.S{, #<imm>}]
	F_RlZtSc1_Pg_AddrXSP:            []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP},                                // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc1_Pg_AddrXSPImm:         []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.<T> }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc1_Pg_AddrXSPImmMulVl:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_IMM},                            // { <Zt>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc1_Pg_AddrXSPXm:          []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R},                              // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtSc1_Pg_AddrXSPXmLSL1:      []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
	F_RlZtSc1_Pg_AddrXSPXmLSL2:      []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},                         // { <Zt>.<T> }, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
	F_RlZtSc1_Pg_AddrXSPZmS:         []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS},                             // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S]
	F_RlZtSc1_Pg_AddrXSPZmSSXTW1:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW1},                       // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #1]
	F_RlZtSc1_Pg_AddrXSPZmSSXTW2:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW2},                       // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #2]
	F_RlZtSc1_Pg_AddrXSPZmSSXTW:     []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW},                        // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod>]
	F_RlZtSc1_Pg_AddrXSPZmSUXTW1:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW1},                       // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #1]
	F_RlZtSc1_Pg_AddrXSPZmSUXTW2:    []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW2},                       // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod> #2]
	F_RlZtSc1_Pg_AddrXSPZmSUXTW:     []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW},                        // { <Zt>.S }, <Pg>, [<Xn|SP>, <Zm>.S, <mod>]
	F_RlZtSc1_Pg_AddrZnS:            []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_ZS},                                 // { <Zt>.S }, <Pg>, [<Zn>.S{, #<imm>}]
	F_RlZtSc1_Pg_AddrZnSImm:         []int{REGLIST_Z | EXT_S | RL_COUNT1, REG_P, MEM_ADDR | MEM_ZS_IMM},                             // { <Zt>.S }, <Pg>, [<Zn>.S{, #<imm>}]
	F_RlZtSc2i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc2i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc2i1_PgZ_AddrXSPXmLSL2:   []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL2}, // { <Zt1>.S, <Zt2>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtSc2i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc2i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc2i1_Pg_AddrXSPXmLSL2:    []int{REGLIST_Z | EXT_S | RL_COUNT2 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},               // { <Zt1>.S, <Zt2>.S }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtSc3i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc3i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc3i1_PgZ_AddrXSPXmLSL2:   []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL2}, // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtSc3i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc3i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc3i1_Pg_AddrXSPXmLSL2:    []int{REGLIST_Z | EXT_S | RL_COUNT3 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},               // { <Zt1>.S, <Zt2>.S, <Zt3>.S }, <Pg>, [<Xn|SP>, <Xm>]
	F_RlZtSc4i1_PgZ_AddrXSP:         []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP},        // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc4i1_PgZ_AddrXSPImmMulVl: []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_IMM},    // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc4i1_PgZ_AddrXSPXmLSL2:   []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P | EXT_ZEROING, MEM_ADDR | MEM_RSP_R_LSL2}, // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>/Z, [<Xn|SP>, <Xm>]
	F_RlZtSc4i1_Pg_AddrXSP:          []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP},                      // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc4i1_Pg_AddrXSPImmMulVl:  []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_IMM},                  // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_RlZtSc4i1_Pg_AddrXSPXmLSL2:    []int{REGLIST_Z | EXT_S | RL_COUNT4 | RL_INC1, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},               // { <Zt1>.S, <Zt2>.S, <Zt3>.S, <Zt4>.S }, <Pg>, [<Xn|SP>, <Xm>]
	F_Rn_Rm:                         []int{REG_R, REG_R},                                                                            // <R><n>, <R><m>
	F_Vd_Pg_ZnB:                     []int{REG_V, REG_P, REG_Z | EXT_B},                                                             // <V><d>, <Pg>, <Zn>.B
	F_Vd_Pg_ZnD:                     []int{REG_V, REG_P, REG_Z | EXT_D},                                                             // <V><d>, <Pg>, <Zn>.D
	F_Vd_Pg_ZnH:                     []int{REG_V, REG_P, REG_Z | EXT_H},                                                             // <V><d>, <Pg>, <Zn>.H
	F_Vd_Pg_ZnS:                     []int{REG_V, REG_P, REG_Z | EXT_S},                                                             // <V><d>, <Pg>, <Zn>.S
	F_Vdn_Pg_Vdn_ZmB:                []int{REG_V, REG_P, REG_V, REG_Z | EXT_B},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmD:                []int{REG_V, REG_P, REG_V, REG_Z | EXT_D},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmH:                []int{REG_V, REG_P, REG_V, REG_Z | EXT_H},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Vdn_Pg_Vdn_ZmS:                []int{REG_V, REG_P, REG_V, REG_Z | EXT_S},                                                      // <V><d>, <Pg>, <V><dn>, <Zm>.<T>
	F_Wdn:                           []int{REG_R},                                                                                   // <Wdn>
	F_Wdn_pattern:                   []int{REG_R, PATTERN},                                                                          // <Wdn>{, <pattern>{, MUL #<imm>}}
	F_Wdn_pattern_Imm:               []int{REG_R, PATTERN, IMM | IMM_INT},                                                           // <Wdn>{, <pattern>{, MUL #<imm>}}
	F_Xd_Imm:                        []int{REG_R, IMM | IMM_INT},                                                                    // <Xd>, #<imm>
	F_Xd_Pg_PnB:                     []int{REG_R, REG_P, REG_P | EXT_B},                                                             // <Xd>, <Pg>, <Pn>.B
	F_Xd_Pg_PnD:                     []int{REG_R, REG_P, REG_P | EXT_D},                                                             // <Xd>, <Pg>, <Pn>.D
	F_Xd_Pg_PnH:                     []int{REG_R, REG_P, REG_P | EXT_H},                                                             // <Xd>, <Pg>, <Pn>.H
	F_Xd_Pg_PnS:                     []int{REG_R, REG_P, REG_P | EXT_S},                                                             // <Xd>, <Pg>, <Pn>.S
	F_Xd_Xn_Imm:                     []int{REG_R, REG_R, IMM | IMM_INT},                                                             // <Xd|SP>, <Xn|SP>, #<imm>
	F_Xdn_PmB:                       []int{REG_R, REG_P | EXT_B},                                                                    // <Xdn>, <Pm>.B
	F_Xdn_PmB_Wdn:                   []int{REG_R, REG_P | EXT_B, REG_R},                                                             // <Xdn>, <Pm>.B, <Wdn>
	F_Xdn_PmD:                       []int{REG_R, REG_P | EXT_D},                                                                    // <Xdn>, <Pm>.D
	F_Xdn_PmD_Wdn:                   []int{REG_R, REG_P | EXT_D, REG_R},                                                             // <Xdn>, <Pm>.D, <Wdn>
	F_Xdn_PmH:                       []int{REG_R, REG_P | EXT_H},                                                                    // <Xdn>, <Pm>.H
	F_Xdn_PmH_Wdn:                   []int{REG_R, REG_P | EXT_H, REG_R},                                                             // <Xdn>, <Pm>.H, <Wdn>
	F_Xdn_PmS:                       []int{REG_R, REG_P | EXT_S},                                                                    // <Xdn>, <Pm>.S
	F_Xdn_PmS_Wdn:                   []int{REG_R, REG_P | EXT_S, REG_R},                                                             // <Xdn>, <Pm>.S, <Wdn>
	F_Xdn_Wdn:                       []int{REG_R, REG_R},                                                                            // <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	F_Xdn_Wdn_pattern:               []int{REG_R, REG_R, PATTERN},                                                                   // <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	F_Xdn_Wdn_pattern_Imm:           []int{REG_R, REG_R, PATTERN, IMM | IMM_INT},                                                    // <Xdn>, <Wdn>{, <pattern>{, MUL #<imm>}}
	F_ZdB_Imm:                       []int{REG_Z | EXT_B, IMM | IMM_INT},                                                            // <Zd>.<T>, #<imm>{, <shift>}
	F_ZdB_Imm_Imm:                   []int{REG_Z | EXT_B, IMM | IMM_INT, IMM | IMM_INT},                                             // <Zd>.B, #<imm1>, #<imm2>
	F_ZdB_Imm_Wd:                    []int{REG_Z | EXT_B, IMM | IMM_INT, REG_R},                                                     // <Zd>.<T>, #<imm>, <R><m>
	F_ZdB_PgM_Imm:                   []int{REG_Z | EXT_B, REG_P | EXT_MERGING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdB_PgM_Imm_Imm:               []int{REG_Z | EXT_B, REG_P | EXT_MERGING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdB_PgM_Vn:                    []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_V},                                               // <Zd>.<T>, <Pg>/M, <V><n>
	F_ZdB_PgM_Wn:                    []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_R},                                               // <Zd>.<T>, <Pg>/M, <R><n|SP>
	F_ZdB_PgM_ZmB_ZaB:               []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, REG_Z | EXT_B},                        // <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	F_ZdB_PgM_ZnB:                   []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B},                                       // <Zd>.B, <Pg>/M, <Zn>.B
	F_ZdB_PgZ_Imm:                   []int{REG_Z | EXT_B, REG_P | EXT_ZEROING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdB_PgZ_Imm_Imm:               []int{REG_Z | EXT_B, REG_P | EXT_ZEROING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdB_PgZ_ZnB:                   []int{REG_Z | EXT_B, REG_P | EXT_ZEROING, REG_Z | EXT_B},                                       // <Zd>.<T>, <Pg>/Z, <Zn>.<T>
	F_ZdB_Pv_ZnB_ZmB:                []int{REG_Z | EXT_B, REG_P, REG_Z | EXT_B, REG_Z | EXT_B},                                      // <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
	F_ZdB_RLZnB_ZmB:                 []int{REG_Z | EXT_B, REGLIST_Z | EXT_B | RL_COUNT1, REG_Z | EXT_B},                             // <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
	F_ZdB_Rm:                        []int{REG_Z | EXT_B, REG_R},                                                                    // <Zdn>.<T>, <R><m>
	F_ZdB_Vm:                        []int{REG_Z | EXT_B, REG_V},                                                                    // <Zdn>.<T>, <V><m>
	F_ZdB_Wn:                        []int{REG_Z | EXT_B, REG_R},                                                                    // <Zd>.<T>, <R><n|SP>
	F_ZdB_Wn_Imm:                    []int{REG_Z | EXT_B, REG_R, IMM | IMM_INT},                                                     // <Zd>.<T>, <R><n>, #<imm>
	F_ZdB_Wn_Wm:                     []int{REG_Z | EXT_B, REG_R, REG_R},                                                             // <Zd>.<T>, <R><n>, <R><m>
	F_ZdB_ZdB_const:                 []int{REG_Z | EXT_B, REG_Z | EXT_B, IMM | IMM_INT},                                             // <Zdn>.<T>, <Zdn>.<T>, #<const>
	F_ZdB_ZdB_imm_shift:             []int{REG_Z | EXT_B, REG_Z | EXT_B, IMM | IMM_INT, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
	F_ZdB_ZnB:                       []int{REG_Z | EXT_B, REG_Z | EXT_B},                                                            // <Zd>.<T>, <Zn>.<T>
	F_ZdB_ZnB_ZmB:                   []int{REG_Z | EXT_B, REG_Z | EXT_B, REG_Z | EXT_B},                                             // <Zd>.B, <Zn>.B, <Zm>.B
	F_ZdB_ZnB_ZmD:                   []int{REG_Z | EXT_B, REG_Z | EXT_B, REG_Z | EXT_D},                                             // <Zd>.B, <Zn>.B, <Zm>.D
	F_ZdB_ZnB_const:                 []int{REG_Z | EXT_B, REG_Z | EXT_B, IMM | IMM_INT},                                             // <Zd>.<T>, <Zn>.<T>, #<const>
	F_ZdB_ZnBidx:                    []int{REG_Z | EXT_B, REG_Z_INDEXED | EXT_B},                                                    // <Zd>.<T>, <Zn>.<T>[<imm>]
	F_ZdB_const:                     []int{REG_Z | EXT_B, IMM | IMM_INT},                                                            // <Zd>.B, #<const>
	F_ZdD:                           []int{REG_Z | EXT_D},                                                                           // <Zdn>.D{, <pattern>{, MUL #<imm>}}
	F_ZdD_AddrZnDZmD:                []int{REG_Z | EXT_D, MEM_ADDR | MEM_ZD_ZD},                                                     // <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
	F_ZdD_AddrZnDZmDLSL:             []int{REG_Z | EXT_D, MEM_ADDR | MEM_ZD_ZD_LSL},                                                 // <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
	F_ZdD_AddrZnDZmDSXTW:            []int{REG_Z | EXT_D, MEM_ADDR | MEM_ZD_ZD_SXTW},                                                // <Zd>.D, [<Zn>.D, <Zm>.D, SXTW{ <amount>}]
	F_ZdD_AddrZnDZmDUXTW:            []int{REG_Z | EXT_D, MEM_ADDR | MEM_ZD_ZD_UXTW},                                                // <Zd>.D, [<Zn>.D, <Zm>.D, UXTW{ <amount>}]
	F_ZdD_Imm:                       []int{REG_Z | EXT_D, IMM | IMM_INT},                                                            // <Zd>.<T>, #<imm>{, <shift>}
	F_ZdD_Imm_Imm:                   []int{REG_Z | EXT_D, IMM | IMM_INT, IMM | IMM_INT},                                             // <Zd>.D, #<imm1>, #<imm2>
	F_ZdD_Imm_Xd:                    []int{REG_Z | EXT_D, IMM | IMM_INT, REG_R},                                                     // <Zd>.<T>, #<imm>, <R><m>
	F_ZdD_PgM_Imm:                   []int{REG_Z | EXT_D, REG_P | EXT_MERGING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdD_PgM_Imm_Imm:               []int{REG_Z | EXT_D, REG_P | EXT_MERGING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdD_PgM_Vn:                    []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_V},                                               // <Zd>.<T>, <Pg>/M, <V><n>
	F_ZdD_PgM_Xn:                    []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_R},                                               // <Zd>.<T>, <Pg>/M, <R><n|SP>
	F_ZdD_PgM_ZmD_ZaD:               []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D},                        // <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	F_ZdD_PgM_ZnD:                   []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D},                                       // <Zd>.D, <Pg>/M, <Zn>.D
	F_ZdD_PgM_ZnD_ZmD_Imm:           []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT},         // <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>, <const>
	F_ZdD_PgM_ZnH:                   []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_H},                                       // <Zd>.D, <Pg>/M, <Zn>.H
	F_ZdD_PgM_ZnS:                   []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_S},                                       // <Zd>.D, <Pg>/M, <Zn>.S
	F_ZdD_PgM_const:                 []int{REG_Z | EXT_D, REG_P | EXT_MERGING, IMM | IMM_FLOAT},                                     // <Zd>.<T>, <Pg>/M, #<const>
	F_ZdD_PgZ_Imm:                   []int{REG_Z | EXT_D, REG_P | EXT_ZEROING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdD_PgZ_Imm_Imm:               []int{REG_Z | EXT_D, REG_P | EXT_ZEROING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdD_PgZ_ZnD:                   []int{REG_Z | EXT_D, REG_P | EXT_ZEROING, REG_Z | EXT_D},                                       // <Zd>.<T>, <Pg>/Z, <Zn>.<T>
	F_ZdD_Pg_ZnD:                    []int{REG_Z | EXT_D, REG_P, REG_Z | EXT_D},                                                     // <Zd>.<T>, <Pg>, <Zd>.<T>
	F_ZdD_PmD:                       []int{REG_Z | EXT_D, REG_P | EXT_D},                                                            // <Zdn>.<T>, <Pm>.<T>
	F_ZdD_Pv_ZnD_ZmD:                []int{REG_Z | EXT_D, REG_P, REG_Z | EXT_D, REG_Z | EXT_D},                                      // <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
	F_ZdD_RLZnD_ZmD:                 []int{REG_Z | EXT_D, REGLIST_Z | EXT_D | RL_COUNT1, REG_Z | EXT_D},                             // <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
	F_ZdD_Rm:                        []int{REG_Z | EXT_D, REG_R},                                                                    // <Zdn>.<T>, <R><m>
	F_ZdD_Vm:                        []int{REG_Z | EXT_D, REG_V},                                                                    // <Zdn>.<T>, <V><m>
	F_ZdD_Xn:                        []int{REG_Z | EXT_D, REG_R},                                                                    // <Zd>.<T>, <R><n|SP>
	F_ZdD_Xn_Imm:                    []int{REG_Z | EXT_D, REG_R, IMM | IMM_INT},                                                     // <Zd>.<T>, <R><n>, #<imm>
	F_ZdD_Xn_Xm:                     []int{REG_Z | EXT_D, REG_R, REG_R},                                                             // <Zd>.<T>, <R><n>, <R><m>
	F_ZdD_ZdD_ZmD_imm:               []int{REG_Z | EXT_D, REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, <Zm>.<T>, #<imm>
	F_ZdD_ZdD_const:                 []int{REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT},                                             // <Zdn>.<T>, <Zdn>.<T>, #<const>
	F_ZdD_ZdD_imm_shift:             []int{REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
	F_ZdD_ZnD:                       []int{REG_Z | EXT_D, REG_Z | EXT_D},                                                            // <Zd>.<T>, <Zn>.<T>
	F_ZdD_ZnD_ZmD:                   []int{REG_Z | EXT_D, REG_Z | EXT_D, REG_Z | EXT_D},                                             // <Zd>.D, <Zn>.D, <Zm>.D
	F_ZdD_ZnD_ZmDidx:                []int{REG_Z | EXT_D, REG_Z | EXT_D, REG_Z_INDEXED | EXT_D},                                     // <Zd>.D, <Zn>.D, <Zm>.D[<imm>]
	F_ZdD_ZnDidx:                    []int{REG_Z | EXT_D, REG_Z_INDEXED | EXT_D},                                                    // <Zd>.<T>, <Zn>.<T>[<imm>]
	F_ZdD_ZnH_ZmH:                   []int{REG_Z | EXT_D, REG_Z | EXT_H, REG_Z | EXT_H},                                             // <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
	F_ZdD_ZnH_ZmHidx:                []int{REG_Z | EXT_D, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                                     // <Zda>.D, <Zn>.H, <Zm>.H[<imm>]
	F_ZdD_ZnS:                       []int{REG_Z | EXT_D, REG_Z | EXT_S},                                                            // <Zd>.<T>, <Zn>.<Tb>
	F_ZdD_Znd_const:                 []int{REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT},                                             // <Zd>.<T>, <Zn>.<T>, #<const>
	F_ZdD_const:                     []int{REG_Z | EXT_D, IMM | IMM_INT},                                                            // <Zd>.D, #<const>
	F_ZdD_const_FP:                  []int{REG_Z | EXT_D, IMM | IMM_FLOAT},                                                          // <Zd>.D, #<const>
	F_ZdD_pattern:                   []int{REG_Z | EXT_D, PATTERN},                                                                  // <Zdn>.D{, <pattern>{, MUL #<imm>}}
	F_ZdD_pattern_Imm:               []int{REG_Z | EXT_D, PATTERN, IMM | IMM_INT},                                                   // <Zdn>.D{, <pattern>{, MUL #<imm>}}
	F_ZdH:                           []int{REG_Z | EXT_H},                                                                           // <Zdn>.H{, <pattern>{, MUL #<imm>}}
	F_ZdH_Imm:                       []int{REG_Z | EXT_H, IMM | IMM_INT},                                                            // <Zd>.<T>, #<imm>{, <shift>}
	F_ZdH_Imm_Imm:                   []int{REG_Z | EXT_H, IMM | IMM_INT, IMM | IMM_INT},                                             // <Zd>.H, #<imm1>, #<imm2>
	F_ZdH_Imm_Wd:                    []int{REG_Z | EXT_H, IMM | IMM_INT, REG_R},                                                     // <Zd>.<T>, #<imm>, <R><m>
	F_ZdH_PgM_Imm:                   []int{REG_Z | EXT_H, REG_P | EXT_MERGING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdH_PgM_Imm_Imm:               []int{REG_Z | EXT_H, REG_P | EXT_MERGING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdH_PgM_Vn:                    []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_V},                                               // <Zd>.<T>, <Pg>/M, <V><n>
	F_ZdH_PgM_Wn:                    []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_R},                                               // <Zd>.<T>, <Pg>/M, <R><n|SP>
	F_ZdH_PgM_ZmH_ZaH:               []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	F_ZdH_PgM_ZnD:                   []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_D},                                       // <Zd>.H, <Pg>/M, <Zn>.D
	F_ZdH_PgM_ZnH:                   []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H},                                       // <Zd>.H, <Pg>/M, <Zn>.H
	F_ZdH_PgM_ZnH_ZmH_Imm:           []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT},         // <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>, <const>
	F_ZdH_PgM_ZnS:                   []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_S},                                       // <Zd>.S, <Pg>/M, <Zn>.H
	F_ZdH_PgM_const:                 []int{REG_Z | EXT_H, REG_P | EXT_MERGING, IMM | IMM_FLOAT},                                     // <Zd>.<T>, <Pg>/M, #<const>
	F_ZdH_PgZ_Imm:                   []int{REG_Z | EXT_H, REG_P | EXT_ZEROING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdH_PgZ_Imm_Imm:               []int{REG_Z | EXT_H, REG_P | EXT_ZEROING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdH_PgZ_ZnH:                   []int{REG_Z | EXT_H, REG_P | EXT_ZEROING, REG_Z | EXT_H},                                       // <Zd>.<T>, <Pg>/Z, <Zn>.<T>
	F_ZdH_PmH:                       []int{REG_Z | EXT_H, REG_P | EXT_H},                                                            // <Zdn>.<T>, <Pm>.<T>
	F_ZdH_Pv_ZnH_ZmH:                []int{REG_Z | EXT_H, REG_P, REG_Z | EXT_H, REG_Z | EXT_H},                                      // <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
	F_ZdH_RLZnH_ZmH:                 []int{REG_Z | EXT_H, REGLIST_Z | EXT_H | RL_COUNT1, REG_Z | EXT_H},                             // <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
	F_ZdH_Rm:                        []int{REG_Z | EXT_H, REG_R},                                                                    // <Zdn>.<T>, <R><m>
	F_ZdH_Vm:                        []int{REG_Z | EXT_H, REG_V},                                                                    // <Zdn>.<T>, <V><m>
	F_ZdH_Wn:                        []int{REG_Z | EXT_H, REG_R},                                                                    // <Zd>.<T>, <R><n|SP>
	F_ZdH_Wn_Imm:                    []int{REG_Z | EXT_H, REG_R, IMM | IMM_INT},                                                     // <Zd>.<T>, <R><n>, #<imm>
	F_ZdH_Wn_Wm:                     []int{REG_Z | EXT_H, REG_R, REG_R},                                                             // <Zd>.<T>, <R><n>, <R><m>
	F_ZdH_ZdH_ZmH_imm:               []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, <Zm>.<T>, #<imm>
	F_ZdH_ZdH_const:                 []int{REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT},                                             // <Zdn>.<T>, <Zdn>.<T>, #<const>
	F_ZdH_ZdH_imm_shift:             []int{REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
	F_ZdH_ZnB:                       []int{REG_Z | EXT_H, REG_Z | EXT_B},                                                            // <Zd>.<T>, <Zn>.<Tb>
	F_ZdH_ZnH:                       []int{REG_Z | EXT_H, REG_Z | EXT_H},                                                            // <Zd>.<T>, <Zn>.<T>
	F_ZdH_ZnH_ZmD:                   []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z | EXT_D},                                             // <Zd>.H, <Zn>.H, <Zm>.D
	F_ZdH_ZnH_ZmH:                   []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z | EXT_H},                                             // <Zd>.H, <Zn>.H, <Zm>.H
	F_ZdH_ZnH_ZmHidx:                []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                                     // <Zd>.D, <Zn>.D, <Zm>.D[<imm>]
	F_ZdH_ZnH_ZmHidx_const:          []int{REG_Z | EXT_H, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H, IMM | IMM_INT},                      // <Zda>.H, <Zn>.H, <Zm>.H[<imm>], <const>
	F_ZdH_ZnH_const:                 []int{REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT},                                             // <Zd>.<T>, <Zn>.<T>, #<const>
	F_ZdH_ZnHidx:                    []int{REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                                                    // <Zd>.<T>, <Zn>.<T>[<imm>]
	F_ZdH_const:                     []int{REG_Z | EXT_H, IMM | IMM_INT},                                                            // <Zd>.H, #<const>
	F_ZdH_const_FP:                  []int{REG_Z | EXT_H, IMM | IMM_FLOAT},                                                          // <Zd>.H, #<const>
	F_ZdH_pattern:                   []int{REG_Z | EXT_H, PATTERN},                                                                  // <Zdn>.H{, <pattern>{, MUL #<imm>}}
	F_ZdH_pattern_Imm:               []int{REG_Z | EXT_H, PATTERN, IMM | IMM_INT},                                                   // <Zdn>.H{, <pattern>{, MUL #<imm>}}
	F_ZdQ_ZnQ_ZmQ:                   []int{REG_Z | EXT_Q, REG_Z | EXT_Q, REG_Z | EXT_Q},                                             // <Zd>.Q, <Zn>.Q, <Zm>.Q
	F_ZdQ_ZnQidx:                    []int{REG_Z | EXT_Q, REG_Z_INDEXED | EXT_Q},                                                    // <Zd>.<T>, <Zn>.<T>[<imm>]
	F_ZdS:                           []int{REG_Z | EXT_S},                                                                           // <Zdn>.S{, <pattern>{, MUL #<imm>}}
	F_ZdS_AddrZnSZmS:                []int{REG_Z | EXT_S, MEM_ADDR | MEM_ZS_ZS},                                                     // <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
	F_ZdS_AddrZnSZmSLSL:             []int{REG_Z | EXT_S, MEM_ADDR | MEM_ZS_ZS_LSL},                                                 // <Zd>.<T>, [<Zn>.<T>, <Zm>.<T>{, <mod> <amount>}]
	F_ZdS_Imm:                       []int{REG_Z | EXT_S, IMM | IMM_INT},                                                            // <Zd>.<T>, #<imm>{, <shift>}
	F_ZdS_Imm_Imm:                   []int{REG_Z | EXT_S, IMM | IMM_INT, IMM | IMM_INT},                                             // <Zd>.S, #<imm1>, #<imm2>
	F_ZdS_Imm_Wd:                    []int{REG_Z | EXT_S, IMM | IMM_INT, REG_R},                                                     // <Zd>.<T>, #<imm>, <R><m>
	F_ZdS_PgM_Imm:                   []int{REG_Z | EXT_S, REG_P | EXT_MERGING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdS_PgM_Imm_Imm:               []int{REG_Z | EXT_S, REG_P | EXT_MERGING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/M, #<imm>{, <shift>}
	F_ZdS_PgM_Vn:                    []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_V},                                               // <Zd>.<T>, <Pg>/M, <V><n>
	F_ZdS_PgM_Wn:                    []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_R},                                               // <Zd>.<T>, <Pg>/M, <R><n|SP>
	F_ZdS_PgM_ZmS_ZaS:               []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S},                        // <Zdn>.<T>, <Pg>/M, <Zm>.<T>, <Za>.<T>
	F_ZdS_PgM_ZnD:                   []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_D},                                       // <Zd>.S, <Pg>/M, <Zn>.D
	F_ZdS_PgM_ZnH:                   []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_H},                                       // <Zd>.S, <Pg>/M, <Zn>.H
	F_ZdS_PgM_ZnS:                   []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S},                                       // <Zd>.S, <Pg>/M, <Zn>.S
	F_ZdS_PgM_ZnS_ZmS_Imm:           []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT},         // <Zda>.<T>, <Pg>/M, <Zn>.<T>, <Zm>.<T>, <const>
	F_ZdS_PgM_const:                 []int{REG_Z | EXT_S, REG_P | EXT_MERGING, IMM | IMM_FLOAT},                                     // <Zd>.<T>, <Pg>/M, #<const>
	F_ZdS_PgZ_Imm:                   []int{REG_Z | EXT_S, REG_P | EXT_ZEROING, IMM | IMM_INT},                                       // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdS_PgZ_Imm_Imm:               []int{REG_Z | EXT_S, REG_P | EXT_ZEROING, IMM | IMM_INT, IMM | IMM_INT},                        // <Zd>.<T>, <Pg>/Z, #<imm>{, <shift>}
	F_ZdS_PgZ_ZnS:                   []int{REG_Z | EXT_S, REG_P | EXT_ZEROING, REG_Z | EXT_S},                                       // <Zd>.<T>, <Pg>/Z, <Zn>.<T>
	F_ZdS_Pg_ZnS:                    []int{REG_Z | EXT_S, REG_P, REG_Z | EXT_S},                                                     // <Zd>.<T>, <Pg>, <Zd>.<T>
	F_ZdS_PmS:                       []int{REG_Z | EXT_S, REG_P | EXT_S},                                                            // <Zdn>.<T>, <Pm>.<T>
	F_ZdS_Pv_ZnS_ZmS:                []int{REG_Z | EXT_S, REG_P, REG_Z | EXT_S, REG_Z | EXT_S},                                      // <Zd>.<T>, <Pv>, <Zn>.<T>, <Zm>.<T>
	F_ZdS_RLZnS_ZmS:                 []int{REG_Z | EXT_S, REGLIST_Z | EXT_S | RL_COUNT1, REG_Z | EXT_S},                             // <Zd>.<T>, { <Zn>.<T> }, <Zm>.<T>
	F_ZdS_Rm:                        []int{REG_Z | EXT_S, REG_R},                                                                    // <Zdn>.<T>, <R><m>
	F_ZdS_Vm:                        []int{REG_Z | EXT_S, REG_V},                                                                    // <Zdn>.<T>, <V><m>
	F_ZdS_Wn:                        []int{REG_Z | EXT_S, REG_R},                                                                    // <Zd>.<T>, <R><n|SP>
	F_ZdS_Wn_Imm:                    []int{REG_Z | EXT_S, REG_R, IMM | IMM_INT},                                                     // <Zd>.<T>, <R><n>, #<imm>
	F_ZdS_Wn_Wm:                     []int{REG_Z | EXT_S, REG_R, REG_R},                                                             // <Zd>.<T>, <R><n>, <R><m>
	F_ZdS_ZdS_ZmS_imm:               []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, <Zm>.<T>, #<imm>
	F_ZdS_ZdS_const:                 []int{REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT},                                             // <Zdn>.<T>, <Zdn>.<T>, #<const>
	F_ZdS_ZdS_imm_shift:             []int{REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT, IMM | IMM_INT},                              // <Zdn>.<T>, <Zdn>.<T>, #<imm>{, <shift>}
	F_ZdS_ZnB_ZmB:                   []int{REG_Z | EXT_S, REG_Z | EXT_B, REG_Z | EXT_B},                                             // <Zda>.<T>, <Zn>.<Tb>, <Zm>.<Tb>
	F_ZdS_ZnB_ZmBidx:                []int{REG_Z | EXT_S, REG_Z | EXT_B, REG_Z_INDEXED | EXT_B},                                     // <Zda>.S, <Zn>.B, <Zm>.B[<imm>]
	F_ZdS_ZnH:                       []int{REG_Z | EXT_S, REG_Z | EXT_H},                                                            // <Zd>.<T>, <Zn>.<Tb>
	F_ZdS_ZnS:                       []int{REG_Z | EXT_S, REG_Z | EXT_S},                                                            // <Zd>.<T>, <Zn>.<T>
	F_ZdS_ZnS_ZmD:                   []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z | EXT_D},                                             // <Zd>.S, <Zn>.S, <Zm>.D
	F_ZdS_ZnS_ZmS:                   []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z | EXT_S},                                             // <Zd>.S, <Zn>.S, <Zm>.S
	F_ZdS_ZnS_ZmSidx:                []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z_INDEXED | EXT_S},                                     // <Zd>.D, <Zn>.D, <Zm>.D[<imm>]
	F_ZdS_ZnS_ZmSidx_const:          []int{REG_Z | EXT_S, REG_Z | EXT_S, REG_Z_INDEXED | EXT_S, IMM | IMM_INT},                      // <Zda>.S, <Zn>.S, <Zm>.S[<imm>], <const>
	F_ZdS_ZnS_const:                 []int{REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT},                                             // <Zd>.<T>, <Zn>.<T>, #<const>
	F_ZdS_ZnSidx:                    []int{REG_Z | EXT_S, REG_Z_INDEXED | EXT_S},                                                    // <Zd>.<T>, <Zn>.<T>[<imm>]
	F_ZdS_const:                     []int{REG_Z | EXT_S, IMM | IMM_INT},                                                            // <Zd>.S, #<const>
	F_ZdS_const_FP:                  []int{REG_Z | EXT_S, IMM | IMM_FLOAT},                                                          // <Zd>.S, #<const>
	F_ZdS_pattern:                   []int{REG_Z | EXT_S, PATTERN},                                                                  // <Zdn>.S{, <pattern>{, MUL #<imm>}}
	F_ZdS_pattern_Imm:               []int{REG_Z | EXT_S, PATTERN, IMM | IMM_INT},                                                   // <Zdn>.S{, <pattern>{, MUL #<imm>}}
	F_Zd_Zn:                         []int{REG_Z, REG_Z},                                                                            // <Zd>, <Zn>
	F_ZdaS_ZnH_ZmH:                  []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z | EXT_H},                                             // <Zd>.S, <Zn>.H, <Zm>.H
	F_ZdaS_ZnH_ZmHidx:               []int{REG_Z | EXT_S, REG_Z | EXT_H, REG_Z_INDEXED | EXT_H},                                     // <Zd>.S, <Zn>.H, <Zm>.H[<imm>]
	F_ZdnB_PgM_ZdnB_ZmB:             []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, REG_Z | EXT_B},                        // <Zdn>.B, <Pg>/M, <Zdn>.B, <Zm>.B
	F_ZdnB_PgM_ZdnB_ZmD:             []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, REG_Z | EXT_D},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	F_ZdnB_PgM_ZdnB_const:           []int{REG_Z | EXT_B, REG_P | EXT_MERGING, REG_Z | EXT_B, IMM | IMM_INT},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
	F_ZdnB_Pg_ZdnB_ZmB:              []int{REG_Z | EXT_B, REG_P, REG_Z | EXT_B, REG_Z | EXT_B},                                      // <Zdn>.B, <Pg>, <Zdn>.B, <Zm>.B
	F_ZdnB_ZdnB_ZmB_imm:             []int{REG_Z | EXT_B, REG_Z | EXT_B, REG_Z | EXT_B, IMM | IMM_INT},                              // <Zdn>.B, <Zdn>.B, <Zm>.B, #<imm>
	F_ZdnD_PgM_ZdnD_ZmD:             []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D},                        // <Zdn>.D, <Pg>/M, <Zdn>.D, <Zm>.D
	F_ZdnD_PgM_ZdnD_ZmD_const:       []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, REG_Z | EXT_D, IMM | IMM_INT},         // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>, <const>
	F_ZdnD_PgM_ZdnD_const:           []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, IMM | IMM_INT},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
	F_ZdnD_PgM_ZdnD_const_FP:        []int{REG_Z | EXT_D, REG_P | EXT_MERGING, REG_Z | EXT_D, IMM | IMM_FLOAT},                      // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	F_ZdnD_Pg_ZdnD_ZmD:              []int{REG_Z | EXT_D, REG_P, REG_Z | EXT_D, REG_Z | EXT_D},                                      // <Zdn>.D, <Pg>, <Zdn>.D, <Zm>.D
	F_ZdnH_PgM_ZdnH_ZmD:             []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_D},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	F_ZdnH_PgM_ZdnH_ZmH:             []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H},                        // <Zdn>.H, <Pg>/M, <Zdn>.H, <Zm>.H
	F_ZdnH_PgM_ZdnH_ZmH_const:       []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, REG_Z | EXT_H, IMM | IMM_INT},         // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>, <const>
	F_ZdnH_PgM_ZdnH_const:           []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, IMM | IMM_INT},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
	F_ZdnH_PgM_ZdnH_const_FP:        []int{REG_Z | EXT_H, REG_P | EXT_MERGING, REG_Z | EXT_H, IMM | IMM_FLOAT},                      // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	F_ZdnH_Pg_ZdnH_ZmH:              []int{REG_Z | EXT_H, REG_P, REG_Z | EXT_H, REG_Z | EXT_H},                                      // <Zdn>.H, <Pg>, <Zdn>.H, <Zm>.H
	F_ZdnS_PgM_ZdnS_ZmD:             []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_D},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.D
	F_ZdnS_PgM_ZdnS_ZmS:             []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S},                        // <Zdn>.S, <Pg>/M, <Zdn>.S, <Zm>.S
	F_ZdnS_PgM_ZdnS_ZmS_const:       []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, REG_Z | EXT_S, IMM | IMM_INT},         // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <Zm>.<T>, <const>
	F_ZdnS_PgM_ZdnS_const:           []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, IMM | IMM_INT},                        // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, #<const>
	F_ZdnS_PgM_ZdnS_const_FP:        []int{REG_Z | EXT_S, REG_P | EXT_MERGING, REG_Z | EXT_S, IMM | IMM_FLOAT},                      // <Zdn>.<T>, <Pg>/M, <Zdn>.<T>, <const>
	F_ZdnS_Pg_ZdnS_ZmS:              []int{REG_Z | EXT_S, REG_P, REG_Z | EXT_S, REG_Z | EXT_S},                                      // <Zdn>.S, <Pg>, <Zdn>.S, <Zm>.S
	F_Zt_AddrXSP:                    []int{REG_Z, MEM_ADDR | MEM_RSP},                                                               // <Zt>, [<Xn|SP>]
	F_Zt_AddrXSPImmMulVl:            []int{REG_Z, MEM_ADDR | MEM_RSP_IMM},                                                           // <Zt>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_prfop_Pg_AddrXSP:              []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP},                                                        // <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_prfop_Pg_AddrXSPImmMulVl:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_IMM},                                                    // <prfop>, <Pg>, [<Xn|SP>{, #<imm>, MUL VL}]
	F_prfop_Pg_AddrXSPXm:            []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_R},                                                      // <prfop>, <Pg>, [<Xn|SP>, <Xm>]
	F_prfop_Pg_AddrXSPXmLSL1:        []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_R_LSL1},                                                 // <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #1]
	F_prfop_Pg_AddrXSPXmLSL2:        []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_R_LSL2},                                                 // <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #2]
	F_prfop_Pg_AddrXSPXmLSL3:        []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_R_LSL3},                                                 // <prfop>, <Pg>, [<Xn|SP>, <Xm>, LSL #3]
	F_prfop_Pg_AddrXSPZmD:           []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD},                                                     // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D]
	F_prfop_Pg_AddrXSPZmDLSL1:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL1},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #1]
	F_prfop_Pg_AddrXSPZmDLSL2:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL2},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #2]
	F_prfop_Pg_AddrXSPZmDLSL3:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_LSL3},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, LSL #3]
	F_prfop_Pg_AddrXSPZmDSXTW1:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW1},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #1]
	F_prfop_Pg_AddrXSPZmDSXTW2:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW2},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #2]
	F_prfop_Pg_AddrXSPZmDSXTW3:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW3},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #3]
	F_prfop_Pg_AddrXSPZmDSXTW:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_SXTW},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW]
	F_prfop_Pg_AddrXSPZmDUXTW1:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW1},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #1]
	F_prfop_Pg_AddrXSPZmDUXTW2:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW2},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #2]
	F_prfop_Pg_AddrXSPZmDUXTW3:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW3},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, SXTW #3]
	F_prfop_Pg_AddrXSPZmDUXTW:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZD_UXTW},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.D, UXTW]
	F_prfop_Pg_AddrXSPZmS:           []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS},                                                     // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S]
	F_prfop_Pg_AddrXSPZmSSXTW1:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW1},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #1]
	F_prfop_Pg_AddrXSPZmSSXTW2:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW2},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #2]
	F_prfop_Pg_AddrXSPZmSSXTW3:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW3},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #3]
	F_prfop_Pg_AddrXSPZmSSXTW:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_SXTW},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW]
	F_prfop_Pg_AddrXSPZmSUXTW1:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW1},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #1]
	F_prfop_Pg_AddrXSPZmSUXTW2:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW2},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #2]
	F_prfop_Pg_AddrXSPZmSUXTW3:      []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW3},                                               // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, SXTW #3]
	F_prfop_Pg_AddrXSPZmSUXTW:       []int{PRFOP, REG_P, MEM_ADDR | MEM_RSP_ZS_UXTW},                                                // <prfop>, <Pg>, [<Xn|SP>, <Zm>.S, UXTW]
	F_prfop_Pg_AddrZnD:              []int{PRFOP, REG_P, MEM_ADDR | MEM_ZD},                                                         // <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
	F_prfop_Pg_AddrZnDImm:           []int{PRFOP, REG_P, MEM_ADDR | MEM_ZD_IMM},                                                     // <prfop>, <Pg>, [<Zn>.D{, #<imm>}]
	F_prfop_Pg_AddrZnS:              []int{PRFOP, REG_P, MEM_ADDR | MEM_ZS},                                                         // <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
	F_prfop_Pg_AddrZnSImm:           []int{PRFOP, REG_P, MEM_ADDR | MEM_ZS_IMM},                                                     // <prfop>, <Pg>, [<Zn>.S{, #<imm>}]
}

// Key into the encoder table.
const (
	E_none = iota
	E_16imm4_Pg_Rn_Zt
	E_2imm4_Pg_Rn_Zt
	E_2imm5_Pg_Zn_Zt
	E_2imm5_Pg_Zn_prfop
	E_2imm6_Pg_Rn_Zt
	E_32imm4_Pg_Rn_Zt
	E_3imm4_Pg_Rn_Zt
	E_4imm4_Pg_Rn_Zt
	E_4imm5_Pg_Zn_Zt
	E_4imm5_Pg_Zn_prfop
	E_4imm6_Pg_Rn_Zt
	E_8imm5_Pg_Zn_Zt
	E_8imm5_Pg_Zn_prfop
	E_8imm6_Pg_Rn_Zt
	E_Pd
	E_Pg_Pdn
	E_Pg_Pn
	E_Pg_Pn_Pd
	E_Pg_Pn_Pdm
	E_Pg_Rn_Rd
	E_Pg_Rn_Zt
	E_Pg_Rn_prfop
	E_Pg_Zn_prfop
	E_Pm_Pg_Pn_Pd
	E_Pn
	E_Pn_Pd
	E_Rd_Rn
	E_Rd_Rn_Rm
	E_Rdn
	E_Rm_Pg_Rn_Zt
	E_Rm_Pg_Rn_prfop
	E_Rm_Rn
	E_Rn_imm6_Rd
	E_XWdn
	E_Zda_Zn_Zm_i2
	E_Zda_Zn_Zm_i3
	E_Zm_Pg_Rn_Zt
	E_Zm_Pg_Rn_prfop
	E_Zm_msz_Zn_Zd
	E_Zm_xs_Pg_Rn_Zt
	E_Zmi1_Zn_Zda
	E_Zmi1_rot_Zn_Zda
	E_Zmi2_Zn_Zda
	E_Zmi2_rot_Zn_Zda
	E_dup_indexed
	E_dupm
	E_i1_Zm_Zn_Zd
	E_i2_Zm_Zn_Zd
	E_i3h_i3l_Zm_Zn_Zd
	E_imm13_Zdn
	E_imm4_Pg_Rn_Zt
	E_imm4_pattern_Rdn
	E_imm4_pattern_XWdn
	E_imm5_Pg_Zn_Zt
	E_imm5_Pg_Zn_prfop
	E_imm6_Pg_Rn_Zt
	E_imm6_Pg_Rn_prfop
	E_imm6_Rd
	E_imm8h_imm8l_Zm_Zdn
	E_imm9h_imm9l_Rn_Zt
	E_pattern_Rdn
	E_pattern_XWdn
	E_size_Pd
	E_size_Pg_Pn_Rd
	E_size_Pg_Rn_Zd
	E_size_Pg_ZmD_Zdn
	E_size_Pg_Zm_Rdn
	E_size_Pg_Zm_Zdn
	E_size_Pg_Zn_Pd_ImmFP0
	E_size_Pg_Zn_Rd
	E_size_Pg_Zn_Vd
	E_size_Pg_Zn_Zd
	E_size_Pg_i1_0p0_1p0_Zdn
	E_size_Pg_i1_0p5_1p0_Zdn
	E_size_Pg_i1_0p5_2p0_Zdn
	E_size_Pg_imm8_Zd
	E_size_Pg_sh_imm8_Zd
	E_size_Pm_Pn_Pd
	E_size_Pm_Rdn
	E_size_Pm_XWdn
	E_size_Pm_Zdn
	E_size_Pn_Pd
	E_size_Pv_Pdn
	E_size_Pv_Zm_Zdn
	E_size_Rm_Pg_Rn_Zt
	E_size_Rm_Rn_Pd
	E_size_Rm_Rn_Zd
	E_size_Rm_Zdn
	E_size_Rm_imm5_Zd
	E_size_Rn_Zd
	E_size_Za_Pg_Zm_Zdn
	E_size_ZmD_Pg_Zn_Pd
	E_size_Zm_Pg_Za_Zdn
	E_size_Zm_Pg_Zn_Pd
	E_size_Zm_Pg_Zn_Zda
	E_size_Zm_Pv_Zn_Zd
	E_size_Zm_RLZn_Zd
	E_size_Zm_ZnT_ZdT
	E_size_Zm_Zn_Zd
	E_size_Zm_Zn_Zda
	E_size_Zm_rot_Pg_Zn_Zda
	E_size_ZnTb_ZdT
	E_size_Zn_Zd
	E_size_imm3_Zm_Zdn
	E_size_imm4_Pg_Rn_Zt
	E_size_imm5_Rn_Zd
	E_size_imm5b_imm5_Zd
	E_size_imm8_Zd
	E_size_imm8_Zd_FP
	E_size_imm8_Zdn
	E_size_pattern_Pd
	E_size_rot_Pg_Zm_Zdn
	E_size_sh_imm8_Zd
	E_size_sh_imm8_Zdn
	E_size_simm5_Pg_Zn_Pd
	E_size_uimm7_Pg_Zn_Pd
	E_sz_Zm_msz_Zn_Zd
	E_tszh_Pg_tszl_imm3_Zdn
	E_tszh_Pg_tszl_imm3_bias1_Zdn
	E_tszh_tszl_imm3_Zn_Zd
	E_tszh_tszl_imm3_bias1_Zn_Zd
	E_xs_Zm_Pg_Rn_Zt

	// Equivalences
	E_Pg_8_5_Pd       = E_Pn_Pd
	E_Pg_Zn_Zd        = E_Pg_Rn_Rd
	E_Rd              = E_Rdn
	E_Zm_Zn_Zda       = E_Rd_Rn_Rm
	E_imm4_pattern_Rd = E_imm4_pattern_Rdn
	E_pattern_Rd      = E_pattern_Rdn
	E_size0_Pg_Zn_Zd  = E_size_Pg_Zn_Zd
	E_size_Pg_Vn_Zd   = E_size_Pg_Rn_Zd
)

// The encoder table holds a list of encoding schemes for operands. Each scheme contains
// a list of rules and a list of indices that mark which operands need to be fed into the
// rule. Each rule produces a 32-bit number which should be OR'd with the base to create
// an instruction encoding.
var encoders = map[int]encoder{
	E_none:                        {},
	E_16imm4_Pg_Rn_Zt:             {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset16_19_16}}},
	E_2imm4_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset2}}},
	E_2imm5_Pg_Zn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset2}}},
	E_2imm5_Pg_Zn_prfop:           {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset2}}},
	E_2imm6_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSP2imm6}}},
	E_32imm4_Pg_Rn_Zt:             {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset32_19_16}}},
	E_3imm4_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset3}}},
	E_4imm4_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset4}}},
	E_4imm5_Pg_Zn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset4}}},
	E_4imm5_Pg_Zn_prfop:           {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset4}}},
	E_4imm6_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSP4imm6}}},
	E_8imm5_Pg_Zn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset8}}},
	E_8imm5_Pg_Zn_prfop:           {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset8}}},
	E_8imm6_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSP8imm6}}},
	E_Pd:                          {[]rule{{[]int{0}, Pd}}},
	E_Pg_Pdn:                      {[]rule{{[]int{0, 2}, Pdn}, {[]int{1}, Pg_8_5}}},
	E_Pg_Pn:                       {[]rule{{[]int{0}, Pg}, {[]int{1}, Pn}}},
	E_Pg_Pn_Pd:                    {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Pn}}},
	E_Pg_Pn_Pdm:                   {[]rule{{[]int{0, 3}, Pdm}, {[]int{1}, Pg}, {[]int{2}, Pn}}},
	E_Pg_Rn_Rd:                    {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Rn}}},
	E_Pg_Rn_Zt:                    {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSP}}},
	E_Pg_Rn_prfop:                 {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSP}}},
	E_Pg_Zn_prfop:                 {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZn}}},
	E_Pm_Pg_Pn_Pd:                 {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg}, {[]int{2}, Pn}, {[]int{3}, Pm}}},
	E_Pn:                          {[]rule{{[]int{0}, Pn}}},
	E_Pn_Pd:                       {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}}},
	E_Rd_Rn:                       {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}}},
	E_Rd_Rn_Rm:                    {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_Rdn:                         {[]rule{{[]int{0}, Rd}}},
	E_Rm_Pg_Rn_Zt:                 {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPXmOffset}}},
	E_Rm_Pg_Rn_prfop:              {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPXmOffset}}},
	E_Rm_Rn:                       {[]rule{{[]int{0}, Rn}, {[]int{1}, Rm}}},
	E_Rn_imm6_Rd:                  {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn_20_16}, {[]int{2}, Imm6}}},
	E_XWdn:                        {[]rule{{[]int{0, 1}, Rdn}}},
	E_Zda_Zn_Zm_i2:                {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_Zda_Zn_Zm_i3:                {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi3}}},
	E_Zm_Pg_Rn_Zt:                 {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPZmOffset}}},
	E_Zm_Pg_Rn_prfop:              {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPZmOffset}}},
	E_Zm_msz_Zn_Zd:                {[]rule{{[]int{0}, Rd}, {[]int{1}, Zm_msz_Zn}}},
	E_Zm_xs_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPZmDmod}}},
	E_Zmi1_Zn_Zda:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi1}}},
	E_Zmi1_rot_Zn_Zda:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi1}, {[]int{3}, rot_11_10}}},
	E_Zmi2_Zn_Zda:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_Zmi2_rot_Zn_Zda:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}, {[]int{3}, rot_11_10}}},
	E_dup_indexed:                 {[]rule{{[]int{0, 1}, dup_indexed}}},
	E_dupm:                        {[]rule{{[]int{0, 1}, sveBitmask}}},
	E_i1_Zm_Zn_Zd:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi1}}},
	E_i2_Zm_Zn_Zd:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi2}}},
	E_i3h_i3l_Zm_Zn_Zd:            {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rmi3h_i3l}}},
	E_imm13_Zdn:                   {[]rule{{[]int{0, 1}, Rdn}, {[]int{0, 2}, sveBitmask}}},
	E_imm4_Pg_Rn_Zt:               {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset_19_16}}},
	E_imm4_pattern_Rdn:            {[]rule{{[]int{0}, Rd}, {[]int{1}, pattern}, {[]int{2}, Uimm4}}},
	E_imm4_pattern_XWdn:           {[]rule{{[]int{0, 1}, Rdn}, {[]int{2}, pattern}, {[]int{3}, Uimm4}}},
	E_imm5_Pg_Zn_Zt:               {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset}}},
	E_imm5_Pg_Zn_prfop:            {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemZnOffset}}},
	E_imm6_Pg_Rn_Zt:               {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPimm6}}},
	E_imm6_Pg_Rn_prfop:            {[]rule{{[]int{0}, svePrfop}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset}}},
	E_imm6_Rd:                     {[]rule{{[]int{0}, Rd}, {[]int{1}, Imm6}}},
	E_imm8h_imm8l_Zm_Zdn:          {[]rule{{[]int{0, 1}, Rdn}, {[]int{2}, Rn}, {[]int{3}, imm8h_imm8l}}},
	E_imm9h_imm9l_Rn_Zt:           {[]rule{{[]int{0}, Rt}, {[]int{1}, RnImm9MulVl}}},
	E_pattern_Rdn:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, pattern}}},
	E_pattern_XWdn:                {[]rule{{[]int{0, 1}, Rdn}, {[]int{2}, pattern}}},
	E_size_Pd:                     {[]rule{{[]int{0}, Pd}, {[]int{0}, sveT}}},
	E_size_Pg_Pn_Rd:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg}, {[]int{2}, Pn}, {[]int{2}, sveT}}},
	E_size_Pg_Rn_Zd:               {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}}},
	E_size_Pg_ZmD_Zdn:             {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{0, 2}, sveT}}},
	E_size_Pg_Zm_Rdn:              {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{3}, sveT}}},
	E_size_Pg_Zm_Zdn:              {[]rule{{[]int{0, 2}, Zdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
	E_size_Pg_Zn_Pd_ImmFP0:        {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, ImmFP0}}},
	E_size_Pg_Zn_Rd:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{2}, sveT}}},
	E_size_Pg_Zn_Vd:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{2}, sveT}}},
	E_size_Pg_Zn_Zd:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}}},
	E_size_Pg_i1_0p0_1p0_Zdn:      {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, i1_FP_0p0_1p0}, {[]int{0, 2}, sveT}}},
	E_size_Pg_i1_0p5_1p0_Zdn:      {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, i1_FP_0p5_1p0}, {[]int{0, 2}, sveT}}},
	E_size_Pg_i1_0p5_2p0_Zdn:      {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, i1_FP_0p5_2p0}, {[]int{0, 2}, sveT}}},
	E_size_Pg_imm8_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Pg_19_16}, {[]int{2}, imm8_FP}}},
	E_size_Pg_sh_imm8_Zd:          {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Pg_19_16}, {[]int{2}, imm8}, {[]int{3}, sh}}},
	E_size_Pm_Pn_Pd:               {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}, {[]int{2}, Pm}, {[]int{0, 1, 2}, sveT}}},
	E_size_Pm_Rdn:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Pm_8_5}, {[]int{1}, sveT}}},
	E_size_Pm_XWdn:                {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pm_8_5}, {[]int{1}, sveT}}},
	E_size_Pm_Zdn:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Pm_8_5}, {[]int{0, 1}, sveT}}},
	E_size_Pn_Pd:                  {[]rule{{[]int{0}, Pd}, {[]int{1}, Pn}, {[]int{0, 1}, sveT}}},
	E_size_Pv_Pdn:                 {[]rule{{[]int{0, 2}, Pdn}, {[]int{1}, Pv}, {[]int{0, 2}, sveT}}},
	E_size_Pv_Zm_Zdn:              {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pv_12_10}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
	E_size_Rm_Pg_Rn_Zt:            {[]rule{{[]int{0}, RLt}, {[]int{0}, sveT_22_21}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPXmOffset}}},
	E_size_Rm_Rn_Pd:               {[]rule{{[]int{0}, Pd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0}, sveT}}},
	E_size_Rm_Rn_Zd:               {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Rn}, {[]int{2}, Rm}}},
	E_size_Rm_Zdn:                 {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{0}, sveT}}},
	E_size_Rm_imm5_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, imm5}, {[]int{2}, Rm}}},
	E_size_Rn_Zd:                  {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Rn}}},
	E_size_Za_Pg_Zm_Zdn:           {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_ZmD_Pg_Zn_Pd:           {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2}, sveT}}},
	E_size_Zm_Pg_Za_Zdn:           {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rm}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}}},
	E_size_Zm_Pg_Zn_Pd:            {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_Zm_Pg_Zn_Zda:           {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_Zm_Pv_Zn_Zd:            {[]rule{{[]int{0}, Rd}, {[]int{1}, Pv_13_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{0, 2, 3}, sveT}}},
	E_size_Zm_RLZn_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{1}, RLn}, {[]int{2}, Rm}, {[]int{0, 1, 2}, sveT}}},
	E_size_Zm_ZnT_ZdT:             {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0, 1}, sveT}}},
	E_size_Zm_Zn_Zd:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0, 1, 2}, sveT}}},
	E_size_Zm_Zn_Zda:              {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{2}, Rm}, {[]int{0}, sveT}}},
	E_size_Zm_rot_Pg_Zn_Zda:       {[]rule{{[]int{0}, Rd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{3}, Rm}, {[]int{4}, rot}, {[]int{0, 2, 3}, sveT}}},
	E_size_ZnTb_ZdT:               {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{0}, sveT}}},
	E_size_Zn_Zd:                  {[]rule{{[]int{0}, Rd}, {[]int{0, 1}, sveT}, {[]int{1}, Rn}}},
	E_size_imm3_Zm_Zdn:            {[]rule{{[]int{0, 1}, Rdn}, {[]int{2}, Rn}, {[]int{3}, imm3}, {[]int{0, 1, 2}, sveT}}},
	E_size_imm4_Pg_Rn_Zt:          {[]rule{{[]int{0}, RLt}, {[]int{0}, sveT_22_21}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPOffset_19_16}}},
	E_size_imm5_Rn_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, Rn}, {[]int{2}, imm5b}}},
	E_size_imm5b_imm5_Zd:          {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, imm5}, {[]int{2}, imm5b}}},
	E_size_imm8_Zd:                {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, imm8}}},
	E_size_imm8_Zd_FP:             {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, imm8_FP}}},
	E_size_imm8_Zdn:               {[]rule{{[]int{0, 1}, Rdn}, {[]int{0, 1}, sveT}, {[]int{2}, imm8}}},
	E_size_pattern_Pd:             {[]rule{{[]int{0}, Pd}, {[]int{0}, sveT}, {[]int{1}, pattern}}},
	E_size_rot_Pg_Zm_Zdn:          {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{3}, Rn}, {[]int{0, 2, 3}, sveT}, {[]int{4}, rot_16_16}}},
	E_size_sh_imm8_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{0}, sveT}, {[]int{1}, imm8}, {[]int{2}, sh}}},
	E_size_sh_imm8_Zdn:            {[]rule{{[]int{0, 1}, Rdn}, {[]int{0, 1}, sveT}, {[]int{2}, imm8}, {[]int{3}, sh}}},
	E_size_simm5_Pg_Zn_Pd:         {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Imm5}}},
	E_size_uimm7_Pg_Zn_Pd:         {[]rule{{[]int{0}, Pd}, {[]int{1}, Pg_12_10}, {[]int{2}, Rn}, {[]int{0, 2}, sveT}, {[]int{3}, Uimm7}}},
	E_sz_Zm_msz_Zn_Zd:             {[]rule{{[]int{0}, Rd}, {[]int{0}, svesz}, {[]int{1}, Zm_msz_Zn}}},
	E_tszh_Pg_tszl_imm3_Zdn:       {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{0, 2, 3}, tszh_tszl_imm3_22_8_5}}},
	E_tszh_Pg_tszl_imm3_bias1_Zdn: {[]rule{{[]int{0, 2}, Rdn}, {[]int{1}, Pg_12_10}, {[]int{0, 2, 3}, tszh_tszl_imm3_bias1_22_8_5}}},
	E_tszh_tszl_imm3_Zn_Zd:        {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{0, 1, 2}, tszh_tszl_imm3}}},
	E_tszh_tszl_imm3_bias1_Zn_Zd:  {[]rule{{[]int{0}, Rd}, {[]int{1}, Rn}, {[]int{0, 1, 2}, tszh_tszl_imm3_bias1}}},
	E_xs_Zm_Pg_Rn_Zt:              {[]rule{{[]int{0}, RLt}, {[]int{1}, Pg_12_10}, {[]int{2}, MemXnSPZmDmod_22}}},
}
