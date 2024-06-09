// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

DATA match<>+0x00(SB)/8, $0x0003000200010000
DATA match<>+0x08(SB)/8, $0x0007000600050004
DATA match<>+0x10(SB)/8, $0x000b000a00090008
DATA match<>+0x18(SB)/8, $0x000f000e000d000c
GLOBL match<>(SB),RODATA,$0x20

// This is no faster than the obvious Go code.
// func(src []LAddr32, shift uint) (counts [16]uint16)
TEXT count32oneAtATime<>(SB),NOSPLIT,$0-64
    // AX = source pointer
    MOVQ src+0(FP), AX
    // BX = source length
    MOVQ src_len+8(FP), BX
    // CX = shift
    MOVQ shift+24(FP), CX

    VZEROUPPER

    // Y0 = 16 x 16 digit comparison array
    VMOVDQU match<>(SB), Y0
    // Y1 = 16 x 16 count array
    VPXOR Y1, Y1, Y1

    CMPQ BX, $0
    JEQ end

loop:
    // Load source value
    MOVQ (AX), R8  // 2 32-bit values
    MOVQ 8(AX), R10  // 2 32-bit values
    // Extract digit
    SHRQ CX, R8
    MOVQ R8, R9
    ANDQ $0xf, R8
    // Y2 = Broadcast digit to all lanes
    VPBROADCASTW R8, Y2  // XXX AVX512
    // Y2 = Compare each lane of Y2 with the match array
    VPCMPEQW Y0, Y2, Y2

    SHRQ $32, R9
    ANDQ $0xf, R9
    VPBROADCASTW R9, Y3  // XXX AVX512
    VPCMPEQW Y0, Y3, Y3

    SHRQ CX, R10
    MOVQ R10, R11
    ANDQ $0xf, R10
    VPBROADCASTW R10, Y4  // XXX AVX512
    VPCMPEQW Y0, Y4, Y4

    SHRQ $32, R11
    ANDQ $0xf, R11
    VPBROADCASTW R11, Y5  // XXX AVX512
    VPCMPEQW Y0, Y5, Y5
    // Add matched count to Y1 count array.
    // The matched lane will be 0xFFFF, so we use SUB.
    VPSUBW Y2, Y1, Y1
    VPSUBW Y3, Y1, Y1
    VPSUBW Y4, Y1, Y1
    VPSUBW Y5, Y1, Y1

    ADDQ $16, AX
    SUBQ $4, BX
    JNZ loop

end:
    VMOVDQU Y1, counts+32(FP)
    RET

DATA matchB<>+0x00(SB)/8, $0x0302010003020100
DATA matchB<>+0x08(SB)/8, $0x0302010003020100
DATA matchB<>+0x10(SB)/8, $0x0b0a09080b0a0908
DATA matchB<>+0x18(SB)/8, $0x0b0a09080b0a0908

DATA matchB<>+0x20(SB)/8, $0x0706050407060504
DATA matchB<>+0x28(SB)/8, $0x0706050407060504
DATA matchB<>+0x30(SB)/8, $0x0f0e0d0c0f0e0d0c
DATA matchB<>+0x38(SB)/8, $0x0f0e0d0c0f0e0d0c
GLOBL matchB<>(SB),RODATA,$0x40

DATA maskB<>+0x00(SB)/4,  $0x0000000f
GLOBL maskB<>(SB),RODATA,$4

DATA shufB<>+0x00(SB)/8,  $0x0404040400000000
DATA shufB<>+0x08(SB)/8,  $0x0c0c0c0c08080808
DATA shufB<>+0x10(SB)/8,  $0x1414141410101010
DATA shufB<>+0x18(SB)/8,  $0x1c1c1c1c18181818
GLOBL shufB<>(SB),RODATA,$0x20

// VPSHUFB mask to expand 8 x 8 in the low 64 bits of each 128 bit lane to 8 x 16.
//
// This weirdness is necessary because VPSHUFB can only shuffle within 128 bit lanes.
DATA expandLo<>+0x00(SB)/8, $0xFF05FF01FF04FF00
DATA expandLo<>+0x08(SB)/8, $0xFF07FF03FF06FF02
DATA expandLo<>+0x10(SB)/8, $0xFF05FF01FF04FF00
DATA expandLo<>+0x18(SB)/8, $0xFF07FF03FF06FF02
GLOBL expandLo<>(SB),RODATA,$0x20

// VPSHUFB mask to expand 8 x 8 in the high 64 bits of each 128 bit lane to 8 x 16.
DATA expandHi<>+0x00(SB)/8, $0xFF0dFF09FF0cFF08
DATA expandHi<>+0x08(SB)/8, $0xFF0fFF0bFF0eFF0a
DATA expandHi<>+0x10(SB)/8, $0xFF0dFF09FF0cFF08
DATA expandHi<>+0x18(SB)/8, $0xFF0fFF0bFF0eFF0a
GLOBL expandHi<>(SB),RODATA,$0x20

// Detailed optimizations are done for my Tiger Lake laptop.
// Instruction latency, throughput, and port information are for this architecture.
// Execution ports are the same as Sunny Cove:
// https://en.wikichip.org/wiki/intel/microarchitectures/sunny_cove#Scheduler_Ports_.26_Execution_Units

TEXT ·count32AVX2Asm(SB),NOSPLIT,$0-40
    // TODO: The startup cost of this is pretty serious. For smaller buffers
    // (and the tail), consider the approach from count32oneAtATime.

    // AX = source pointer
    MOVQ src+0(FP), AX
    // BX = source length
    MOVQ src_len+8(FP), BX
    // BX = AX + 4*BX = source end
    LEAQ (AX)(BX*4), BX
    // DX = shift
    MOVQ shift+24(FP), DX

    VZEROUPPER

#define YRESULT Y14
    VPXOR YRESULT, YRESULT, YRESULT

    CMPQ AX, BX
    JGE end

    // Prepare match vectors. We compute 8 histograms in parallel, where each
    // bucket is 8 bits. We do this in clusters of 4 buckets, so that these
    // clusters correspond to the 32-bit lane width of the input.
    //
    // Count  Match  Histogram buckets
    //    Y4     Y0  ba98 ba98 ba98 ba98  3210 3210 3210 3210
    //    Y5     Y1  fedc fedc fedc fedc  7654 7654 7654 7654
    //    Y6     Y2  3210 3210 3210 3210  ba98 ba98 ba98 ba98
    //    Y7     Y3  7654 7654 7654 7654  fedc fedc fedc fedc
    // Counter # ->     7    6    5    4     3    2    1    0
    //
    // This somewhat odd pattern gets things in place for VPHADDW.
    VMOVDQU matchB<>+0x00(SB), Y0
    VMOVDQU matchB<>+0x20(SB), Y1
    VPERMQ $0b01001110, Y0, Y2 // Swap 128 bit lanes
    VPERMQ $0b01001110, Y1, Y3 // Swap 128 bit lanes

    // Prepare mask vector
#define YMASK Y8
    VPBROADCASTD maskB<>(SB), YMASK

    // Prepare shuffle vector
#define YSHUF Y9
    VMOVDQU shufB<>(SB), Y9

outer:
    // Because we're keeping only 8 bit counts in the inner loop, the inner loop
    // must execute at most 255 times before we spill to the larger count vector.
    MOVQ AX, CX
    ADDQ $(32*255), CX
    // CX = min(BX, CX)
    CMPQ BX, CX
    CMOVQCS BX, CX

    // Prepare 8-bit count vectors
    VPXOR Y4, Y4, Y4
    VPXOR Y5, Y5, Y5
    VPXOR Y6, Y6, Y6
    VPXOR Y7, Y7, Y7

    // X15 = shift
#define XSHIFT X15
    MOVQ DX, XSHIFT

loop:
    // Process 8 32-bit values at a time.
    VMOVDQU (AX), Y10           // [≤5;≤8]  0.50    1*p23

    // Shift and mask
    // TODO: Shift by an immediate is faster. Generate count functions for each shift?
    //VPSRLD $20, Y10, Y10      // 1        0.50    1*p01
    VPSRLD XSHIFT, Y10, Y10     // [1;4]    1.00    1*p01+1*p5
    VPAND YMASK, Y10, Y10       // 1        0.33    1*p015

    // Broadcast low byte of each 32-bit lane to the 4 bytes in that lane.
    VPSHUFB YSHUF, Y10, Y10     // 1        0.5     1*p15

    // Compare with match vectors and add to count vectors.
    VPCMPEQB Y0, Y10, Y11       // 1        0.5     1*p01
    VPSUBB Y11, Y4, Y4          // 1        0.33    1*p015
    VPCMPEQB Y1, Y10, Y11       // 1        0.5     1*p01
    VPSUBB Y11, Y5, Y5          // 1        0.33    1*p015
    VPCMPEQB Y2, Y10, Y11       // 1        0.5     1*p01
    VPSUBB Y11, Y6, Y6          // 1        0.33    1*p015
    VPCMPEQB Y3, Y10, Y11       // 1        0.5     1*p01
    VPSUBB Y11, Y7, Y7          // 1        0.33    1*p015

    ADDQ $32, AX
    CMPQ AX, CX
    JL loop

// Free up X/Y15 for use below
#undef XSHIFT

    // Note: Don't set a gdb breakpoint on the next instruction or step to it.
    // If you stop on this instruction, it will set Y12 wrong!
#define ELO Y12
#define EHI Y13
    VMOVDQU expandLo<>(SB), ELO
    VMOVDQU expandHi<>(SB), EHI

    // Add in the counters from Y4 and Y5
    VPSHUFB ELO, Y4, Y10
    VPSHUFB ELO, Y5, Y11
    VPHADDW Y11, Y10, Y15
    VPADDW YRESULT, Y15, YRESULT
    VPSHUFB EHI, Y4, Y10
    VPSHUFB EHI, Y5, Y11
    VPHADDW Y11, Y10, Y15
    VPADDW YRESULT, Y15, YRESULT

    // Add in the counters from Y6 and Y7
    VPSHUFB ELO, Y6, Y10
    VPSHUFB ELO, Y7, Y11
    VPHADDW Y11, Y10, Y15
    VPERMQ $0b01001110, Y15, Y15 // Swap 128 bit lanes
    VPADDW YRESULT, Y15, YRESULT
    VPSHUFB EHI, Y6, Y10
    VPSHUFB EHI, Y7, Y11
    VPHADDW Y11, Y10, Y15
    VPERMQ $0b01001110, Y15, Y15 // Swap 128 bit lanes
    VPADDW YRESULT, Y15, YRESULT

    CMPQ AX, BX
    JL outer

end:
    MOVQ counts+32(FP), AX
    VMOVDQU YRESULT, (AX)

    RET
