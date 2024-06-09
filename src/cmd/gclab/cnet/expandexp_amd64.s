// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"


// 341 bits ([6]uint64) -> 1024 bits ([16]uint64)
TEXT ·expandX3(SB),NOSPLIT,$0-16
    MOVQ packed+0(FP), SI // Packed input bitmap pointer
    MOVQ unpacked+8(FP), DI // Expanded output bitmap pointer

    MOVQ $0b10010010_01001001_00100100_10010010_01001001_00100100_10010010_01001001, AX

    // CX = terminal value of DI
    LEAQ (15*8)(DI), CX

top:
    // Load the first 64 bits of the packed bitmap
    // XXX Maybe load the whole thing into Y regs and extract as a I go?
    MOVQ 0(SI), BX
    PDEPQ AX, BX, DX
    // TODO: AVX2 has VPMULLD, AVX512DQ has VPMULLQ
    IMULQ $0b111, DX
    MOVQ DX, 0(DI)
    SHRQ $21, BX // Keeps last bit, which we need more of. 43 bits remaining.

    CMPQ DI, CX
    JEQ end

    PDEPQ AX, BX, DX
    IMULQ $0b111, DX
    SARQ $1, DX
    MOVQ DX, 8(DI)
    SHRQ $21, BX // 22 bits remaining.

    PDEPQ AX, BX, DX
    IMULQ $0b111, DX
    SARQ $2, DX
    MOVQ DX, 16(DI)
    // Used remaining 22 bits.

    // Repeat. We turn each uint64 input into 3 uint64 outputs. Careful of the last one (or we'll produce 18 uint64 outputs.)
    ADDQ $8, SI
    ADDQ $24, DI
    JMP top

end:
    RET

// 170 bits ([3]uint64) -> 1024 bits ([16]uint64)
TEXT ·expandX6(SB),NOSPLIT,$0-16
    MOVQ packed+0(FP), SI // Packed input bitmap pointer
    MOVQ unpacked+8(FP), DI // Expanded output bitmap pointer

    // XXX It's going to be a problem once I need to SARQ by > 2.
    // I could switch to a left-aligned mask at that point, but that would mess up the low bits.
    MOVQ $0b00010000_01000001_00000100_00010000_01000001_00000100_00010000_01000001, AX

    // The rest of this hasn't been updated from X3.

    JMP abort(SB)

    // CX = terminal value of DI
    LEAQ (15*8)(DI), CX

top:
    // Load the first 64 bits of the packed bitmap
    // XXX Maybe load the whole thing into Y regs and extract as a I go?
    MOVQ 0(SI), BX
    PDEPQ AX, BX, DX
    // TODO: AVX2 has VPMULLD, AVX512DQ has VPMULLQ
    IMULQ $0b111, DX
    MOVQ DX, 0(DI)
    SHRQ $21, BX // Keeps last bit, which we need more of. 43 bits remaining.

    CMPQ DI, CX
    JEQ end

    PDEPQ AX, BX, DX
    IMULQ $0b111, DX
    SARQ $1, DX
    MOVQ DX, 8(DI)
    SHRQ $21, BX // 22 bits remaining.

    PDEPQ AX, BX, DX
    IMULQ $0b111, DX
    SARQ $2, DX
    MOVQ DX, 16(DI)
    // Used remaining 22 bits.

    // Repeat. We turn each uint64 input into 3 uint64 outputs. Careful of the last one (or we'll produce 18 uint64 outputs.)
    ADDQ $8, SI
    ADDQ $24, DI
    JMP top

end:
    RET

GLOBL gfExpand6<>(SB), RODATA, $0x40
DATA  gfExpand6<>+0x00(SB)/8, $0x0101010101010202
DATA  gfExpand6<>+0x08(SB)/8, $0x0202020204040404
DATA  gfExpand6<>+0x10(SB)/8, $0x0404080808080808
DATA  gfExpand6<>+0x18(SB)/8, $0x1010101010102020
DATA  gfExpand6<>+0x20(SB)/8, $0x2020202040404040
DATA  gfExpand6<>+0x28(SB)/8, $0x4040808080808080
DATA  gfExpand6<>+0x30(SB)/8, $0x0000000000000000
DATA  gfExpand6<>+0x38(SB)/8, $0x0000000000000000

GLOBL gfPerm6<>(SB), RODATA, $0x40
DATA  gfPerm6<>+0x00(SB)/1, $0x00
DATA  gfPerm6<>+0x01(SB)/1, $0x08
DATA  gfPerm6<>+0x02(SB)/1, $0x10
DATA  gfPerm6<>+0x03(SB)/1, $0x18
DATA  gfPerm6<>+0x04(SB)/1, $0x20
DATA  gfPerm6<>+0x05(SB)/1, $0x28
DATA  gfPerm6<>+0x06(SB)/1, $0x00
DATA  gfPerm6<>+0x07(SB)/1, $0x00
DATA  gfPerm6<>+0x08(SB)/1, $0x00
DATA  gfPerm6<>+0x09(SB)/1, $0x00
DATA  gfPerm6<>+0x0a(SB)/1, $0x00
DATA  gfPerm6<>+0x0b(SB)/1, $0x00
DATA  gfPerm6<>+0x0c(SB)/1, $0x00
DATA  gfPerm6<>+0x0d(SB)/1, $0x00
DATA  gfPerm6<>+0x0e(SB)/1, $0x00
DATA  gfPerm6<>+0x0f(SB)/1, $0x00
DATA  gfPerm6<>+0x10(SB)/1, $0x00
DATA  gfPerm6<>+0x11(SB)/1, $0x00
DATA  gfPerm6<>+0x12(SB)/1, $0x00
DATA  gfPerm6<>+0x13(SB)/1, $0x00
DATA  gfPerm6<>+0x14(SB)/1, $0x00
DATA  gfPerm6<>+0x15(SB)/1, $0x00
DATA  gfPerm6<>+0x16(SB)/1, $0x00
DATA  gfPerm6<>+0x17(SB)/1, $0x00
DATA  gfPerm6<>+0x18(SB)/1, $0x00
DATA  gfPerm6<>+0x19(SB)/1, $0x00
DATA  gfPerm6<>+0x1a(SB)/1, $0x00
DATA  gfPerm6<>+0x1b(SB)/1, $0x00
DATA  gfPerm6<>+0x1c(SB)/1, $0x00
DATA  gfPerm6<>+0x1d(SB)/1, $0x00
DATA  gfPerm6<>+0x1e(SB)/1, $0x00
DATA  gfPerm6<>+0x1f(SB)/1, $0x00
DATA  gfPerm6<>+0x20(SB)/1, $0x00
DATA  gfPerm6<>+0x21(SB)/1, $0x00
DATA  gfPerm6<>+0x22(SB)/1, $0x00
DATA  gfPerm6<>+0x23(SB)/1, $0x00
DATA  gfPerm6<>+0x24(SB)/1, $0x00
DATA  gfPerm6<>+0x25(SB)/1, $0x00
DATA  gfPerm6<>+0x26(SB)/1, $0x00
DATA  gfPerm6<>+0x27(SB)/1, $0x00
DATA  gfPerm6<>+0x28(SB)/1, $0x00
DATA  gfPerm6<>+0x29(SB)/1, $0x00
DATA  gfPerm6<>+0x2a(SB)/1, $0x00
DATA  gfPerm6<>+0x2b(SB)/1, $0x00
DATA  gfPerm6<>+0x2c(SB)/1, $0x00
DATA  gfPerm6<>+0x2d(SB)/1, $0x00
DATA  gfPerm6<>+0x2e(SB)/1, $0x00
DATA  gfPerm6<>+0x2f(SB)/1, $0x00
DATA  gfPerm6<>+0x30(SB)/1, $0x00
DATA  gfPerm6<>+0x31(SB)/1, $0x00
DATA  gfPerm6<>+0x32(SB)/1, $0x00
DATA  gfPerm6<>+0x33(SB)/1, $0x00
DATA  gfPerm6<>+0x34(SB)/1, $0x00
DATA  gfPerm6<>+0x35(SB)/1, $0x00
DATA  gfPerm6<>+0x36(SB)/1, $0x00
DATA  gfPerm6<>+0x37(SB)/1, $0x00
DATA  gfPerm6<>+0x38(SB)/1, $0x00
DATA  gfPerm6<>+0x39(SB)/1, $0x00
DATA  gfPerm6<>+0x3a(SB)/1, $0x00
DATA  gfPerm6<>+0x3b(SB)/1, $0x00
DATA  gfPerm6<>+0x3c(SB)/1, $0x00
DATA  gfPerm6<>+0x3d(SB)/1, $0x00
DATA  gfPerm6<>+0x3e(SB)/1, $0x00
DATA  gfPerm6<>+0x3f(SB)/1, $0x00

// 170 bits ([3]uint64) -> 1024 bits ([16]uint64)
TEXT ·expandX6AVX512(SB),NOSPLIT,$0-16
    MOVQ packed+0(FP), SI // Packed input bitmap pointer
    MOVQ unpacked+8(FP), DI // Expanded output bitmap pointer

    // Z1 [8]uint64 = expander matrixes
    VMOVDQU64 gfExpand6<>(SB), Z1
    // Z3 [64]uint8 = result shuffle
    VMOVDQU64 gfPerm6<>(SB), Z2

    // Z3 = broadcasted uint64 word of packed input
    VPBROADCASTQ 0(SI), Z3
    // Z3 = expand inputs
    VGF2P8AFFINEQB $0, Z1, Z3, Z3   // Requires GFNI
    // Z3 = shuffle output bytes into place
    VPERMB Z3, Z2, Z3               // Requires AVX512-VBMI
    // Write to output
    // XXX There's only 384 bits (48 bytes, 6 uint64s). The whole result will fit in two
    // Z registers, so I should arrange to use VALIGNQ to collect the results in two registers
    // and then write out at the end. Directly arranging for VALIGNQ would require
    // different permutations for each input. But I could use a mask.
    //
    // Or I could use a blend, which might be easier to think about,
    // but requires using a K register instead of just an immediate.
    //
    // 3 identical VPERMBs + 1 VALIGNQ + 2 masked VALIGNQs = 1 load + 3 * 3 + 3 * 3 = 1 load + 18
    //
    // 3 identical VPERMBs + 2 distinct VPERMT2D = 1 load + 3 * 3 + 2 loads + 2 * 3 = 3 loads + 15
    //
    // 2 distinct VPERMBs + 2 distinct masked VPERMBs = 2 loads + 2 * 3 + 2 loads + 2 * 5 = 4 loads + 16
    //
    // 2 distinct VPERMT2B = 2 loads + 2 * 5 = 2 loads + 10
    //
    //VMOVDQU64 Z3, 0(DI)

    // Z4 = expand packed[1] to 384 bits/48 bytes/6 uint64s.
    VPBROADCASTQ 8(SI), Z4
    VGF2P8AFFINEQB $0, Z1, Z4, Z4
    VPERMB Z4, Z2, Z4

    // Combine into Z3
    MOVQ $0b11000000, AX
    //KMOV AX, K1
    VALIGNQ $2, Z4, Z4, K1, Z3

    // Z4 = expand packed[2] to final 256 bits/32 bytes/4 uint64.
    VPBROADCASTQ 16(SI), Z4
    VGF2P8AFFINEQB $0, Z1, Z4, Z4
    VPERMB Z4, Z2, Z4


    VMOVDQU64 Z3, 0(DI)

    // XXX Repeat for each input word.
    RET
