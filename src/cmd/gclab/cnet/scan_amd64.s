// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·expandAsm(SB), NOSPLIT, $0-24
    MOVQ sizeClass+0(FP), CX
    MOVQ packed+8(FP), AX

    // Call the expander for this size class
    LEAQ ·gcExpanders(SB), BX
    CALL (BX)(CX*8)

    MOVQ unpacked+16(FP), DI // Expanded output bitmap pointer
    VMOVDQU64 Z1, 0(DI)
    VMOVDQU64 Z2, 64(DI)
    RET

// If PPOPCNT is defined, popcount the scan mask in parallel. If not defined,
// popcount each field of the scan mask as we process it.
//
// It's slightly faster to do a parallel popcount.
#define PPOPCNT

// If COMPRESSLOOP is defined, compress the scan mask to just non-empty frames
// and loop sequentially over that. The idea behind this is to make the loop
// more predictable, but keeping track of which frames a in the compressed
// scan mask just makes it really complicated, so I didn't finish implementing
// this.
//
// If enabled, USESTACK must NOT be defined.
//
//#define COMPRESSLOOP

// If GATHERLOAD is defined, use a gather load to read each 64-byte frame, rather
// than a full VMOVDQU64.
//
// This is 4–8x slower (!).
//#define GATHERLOAD

// If USESTACK is defined, put the scan mask and popcount on the stack for the loop.
// Otherwise, consume them directly from registers.
//
// It's 30–90 ns faster to USESTACK (surprisingly)
#define USESTACK

// If FILTERNIL is defined, filter nil pointers before enqueueing.
//
// This is about 25% slower for in-cache scans and 5% slower for out-of-cache
// scans, even if all of the pointers get filtered. We probably still want to
// do this filtering, but maybe doing it in the scan loop is a bad idea.
//#define FILTERNIL

GLOBL byteIndexes<>(SB), RODATA, $0x80
DATA  byteIndexes<>+0x00(SB)/8, $0x0706050403020100
DATA  byteIndexes<>+0x08(SB)/8, $0x0f0e0d0c0b0a0908
DATA  byteIndexes<>+0x10(SB)/8, $0x1716151413121110
DATA  byteIndexes<>+0x18(SB)/8, $0x1f1e1d1c1b1a1918
DATA  byteIndexes<>+0x20(SB)/8, $0x2726252423222120
DATA  byteIndexes<>+0x28(SB)/8, $0x2f2e2d2c2b2a2928
DATA  byteIndexes<>+0x30(SB)/8, $0x3736353433323130
DATA  byteIndexes<>+0x38(SB)/8, $0x3f3e3d3c3b3a3938
DATA  byteIndexes<>+0x40(SB)/8, $0x4746454443424140
DATA  byteIndexes<>+0x48(SB)/8, $0x4f4e4d4c4b4a4948
DATA  byteIndexes<>+0x50(SB)/8, $0x5756555453525150
DATA  byteIndexes<>+0x58(SB)/8, $0x5f5e5d5c5b5a5958
DATA  byteIndexes<>+0x60(SB)/8, $0x6766656463626160
DATA  byteIndexes<>+0x68(SB)/8, $0x6f6e6d6c6b6a6968
DATA  byteIndexes<>+0x70(SB)/8, $0x7776757473727170
DATA  byteIndexes<>+0x78(SB)/8, $0x7f7e7d7c7b7a7978

GLOBL frameOffsets<>(SB), RODATA, $0x40
DATA  frameOffsets<>+0x00(SB)/8, $0x00
DATA  frameOffsets<>+0x08(SB)/8, $0x08
DATA  frameOffsets<>+0x10(SB)/8, $0x10
DATA  frameOffsets<>+0x18(SB)/8, $0x18
DATA  frameOffsets<>+0x20(SB)/8, $0x20
DATA  frameOffsets<>+0x28(SB)/8, $0x28
DATA  frameOffsets<>+0x30(SB)/8, $0x30
DATA  frameOffsets<>+0x38(SB)/8, $0x38

GLOBL shiftR1<>(SB), RODATA, $0x40
DATA  shiftR1<>+0x00(SB)/8, $0x0807060504030201
DATA  shiftR1<>+0x08(SB)/8, $0x100f0e0d0c0b0a09
DATA  shiftR1<>+0x10(SB)/8, $0x1817161514131211
DATA  shiftR1<>+0x18(SB)/8, $0x201f1e1d1c1b1a19
DATA  shiftR1<>+0x20(SB)/8, $0x2827262524232221
DATA  shiftR1<>+0x28(SB)/8, $0x302f2e2d2c2b2a29
DATA  shiftR1<>+0x30(SB)/8, $0x3837363534333231
DATA  shiftR1<>+0x38(SB)/8, $0x403f3e3d3c3b3a39

// TODO: If !USESTACK, set the frame size to 0. (Because vet doesn't understand
// the preprocessor, it complains about the USESTACK code if we do this.)
TEXT ·scanSpanPackedAVX512(SB), NOSPLIT, $256-44
    // Z1+Z2 = Expand the grey object mask into a grey mask
    MOVQ objDarts+16(FP), AX
    MOVQ sizeClass+24(FP), CX
    LEAQ ·gcExpanders(SB), BX
    CALL (BX)(CX*8)

    // Z3+Z4 = Load the pointer mask
    MOVQ ptrMask+32(FP), AX
    VMOVDQU64 0(AX), Z3
    VMOVDQU64 64(AX), Z4

    // Z1+Z2 = Combine the grey mask with the pointer mask to get the scan mask
    VPANDQ Z1, Z3, Z1
    VPANDQ Z2, Z4, Z2

#ifdef COMPRESSLOOP
    // Z3+Z4 [128]uint8 = Byte indexes
    VMOVDQU64 byteIndexes<>+0x00(SB), Z5
    VMOVDQU64 byteIndexes<>+0x40(SB), Z6

    // K1+K2 = Mask of non-zero elements in Z1+Z2
    VPTESTMB Z1, Z1, K1 // Requires AVX512BW
    VPTESTMB Z2, Z2, K2

    // Compress Z1+Z2 to just non-zero elements
    VPCOMPRESSB.Z Z1, K1, Z1  // Requires AVX512_VBMI2
    VPCOMPRESSB.Z Z2, K2, Z2

    // Compress Z5+Z6 indexes to the same subset
    VPCOMPRESSB.Z Z5, K1, Z5
    VPCOMPRESSB.Z Z6, K1, Z6
#endif

    // Now each bit of Z1+Z2 represents one word of the span.
    // Thus, each byte covers 64 bytes of memory, which is also how
    // much we can fix in a Z register.
    //
    // We do a load/compress for each 64 byte frame.
    //
    // Z3+Z4 [128]uint8 = Number of memory words to scan in each 64 byte frame
#ifdef PPOPCNT
    VPOPCNTB Z1, Z3 // Requires BITALG
    VPOPCNTB Z2, Z4
#endif

#ifdef USESTACK
    // Store the scan mask and word counts at 0(SP) and 128(SP).
    //
    // TODO: Is it better to read directly from the registers?
    VMOVDQU64 Z1, 0(SP)
    VMOVDQU64 Z2, 64(SP)
#ifdef PPOPCNT
    VMOVDQU64 Z3, 128(SP)
    VMOVDQU64 Z4, 192(SP)
#endif
#else // !USESTACK
    // Z8 [64]uint8 = Permutation to shift right one byte
    VMOVDQU64 shiftR1<>(SB), Z8
#endif

    // SI = Current address in span
    MOVQ mem+0(FP), SI
    // DI = Scan buffer base
    MOVQ buf+8(FP), DI
    // DX = Index in scan buffer, (DI)(DX*8) = Current position in scan buffer
    MOVQ $0, DX

#ifdef COMPRESSLOOP
    JMP loop1start
loop1body:
    KMOVB CX, K1
    // AX = frame index
    PEXTRB X5, $0, AX
    // AX = AX*(64*8) = AX<<9, byte offset in span
    SHLQ $9, AX
    // Z10 = load frame
    VMOVDQA64 (AX)(SI), Z10
    // Collect just the pointers from the greyed objects into the scan buffer.
    VPCOMPRESSQ Z10, K1, (DI)(DX*8)
    // Advance the scan buffer position by the number of pointers.
#ifdef PPOPCNT
    PEXTRB X3, $0, CX
#else
    POPCNTL CX, CX
#endif
    ADDQ CX, DX
    // XXX: Rotate Z1, Z3, Z5
#error not implemented
loop1start:
    // CX = K1 = scan mask
    PEXTRB X1, $0, CX
    TESTB CX, CX
    JNZ loop1body

    // XXX: Same loop over Z2/Z4/Z6

#else // !COMPRESSLOOP

#ifdef GATHERLOAD
    // Z7 [8]uint64 = Word offsets for gather load
    VMOVDQU64 frameOffsets<>(SB), Z7
#endif

#ifdef USESTACK
    // AX = address in scan mask, 128(AX) = address in popcount
    LEAQ 0(SP), AX

    // Loop over the 64 byte frames in this span.
    // BX = 1 past the end of the scan mask
    LEAQ 128(SP), BX

    PCALIGN $64
loop:
    // CX = Fetch the mask of words to load from this frame.
    MOVBQZX 0(AX), CX
    // Skip empty frames.
    TESTQ CX, CX
    JZ skip

    // Load the 64 byte frame.
    KMOVB CX, K1
#ifdef GATHERLOAD
    KMOVB CX, K2
    VPGATHERQQ (SI)(Z7*1), K2, Z1
#else // !GATHERLOAD
    VMOVDQA64 0(SI), Z1
#endif
#ifdef FILTERNIL
    // Filter out nil pointers
    VPCMPEQQ Z0, Z1, K2
    KANDB K1, K1, K2
#endif
    // Collect just the pointers from the greyed objects into the scan buffer.
    VPCOMPRESSQ Z1, K1, (DI)(DX*8)
    // Advance the scan buffer position by the number of pointers.
#ifdef PPOPCNT
    MOVBQZX 128(AX), CX
#else
    POPCNTL CX, CX
#endif
    ADDQ CX, DX

skip:
    ADDQ $64, SI
    ADDQ $1, AX
    CMPQ AX, BX
    JB loop
#else // !USESTACK
    CALL scan1<>(SB)
    VMOVDQA64 Z2, Z1
    VMOVDQA64 Z4, Z3
    CALL scan1<>(SB)
#endif // USESTACK
#endif // COMPRESSLOOP

end:
    MOVL DX, count+40(FP)
    VZEROUPPER
    RET

#ifndef USESTACK
// scan1 is a helper that scans 64x 64-byte frames (out of 128).
//
// Z1 [64]uint8 = scan mask
// Z3 [64]uint8 = pop counts
// Z7 [8]uint64 = (if GATHERLOAD) gather offsets
// Z8 [64]uint8 = permutation to shift right one byte
// SI unsafe.Pointer = starting memory location for scan
// DI *[...]uint64 = scan buffer base
// DX int = current index into scan buffer
//
// Z2 and Z4 are protected.
//
// On return, Z1 and Z3 are clobbered, SI points to the end of the scanned region,
// and DX is updated.
TEXT scan1<>(SB), NOSPLIT, $0-0
    // AX = final address
    LEAQ (64*64)(SI), AX
loop:
    // CX = Fetch the mask of words to load from this frame.
    PEXTRB $0, X1, CX
    // Skip empty frames.
    TESTQ CX, CX
    JZ skip

    // Load the 64 byte frame.
    KMOVB CX, K1
#ifdef GATHERLOAD
    KMOVB CX, K2
    VPGATHERQQ (SI)(Z7*1), K2, Z5
#else // !GATHERLOAD
    // TODO: VMOVNTDQA for non-temporal load?
VMOVDQA64 0(SI), Z5
#endif
#ifdef FILTERNIL
    // Filter out nil pointers
    VPCMPEQQ Z0, Z5, K2
    KANDB K1, K1, K2
#endif
    // Collect just the pointers from the greyed objects into the scan buffer.
    VPCOMPRESSQ Z5, K1, (DI)(DX*8)
    // Advance the scan buffer position by the number of pointers.
#ifdef PPOPCNT
    PEXTRB $0, X3, CX
#else
    POPCNTL CX, CX
#endif
    ADDQ CX, DX

skip:
    // Shift Z1 and Z3 by one byte
    VPERMB Z1, Z8, Z1
    VPERMB Z3, Z8, Z3
    ADDQ $64, SI
    CMPQ SI, AX
    JB loop

    RET
#endif

// If DOGATHERPREFETCH is defined, do a gather prefetch before the scan loop.
//
// This is weirdly slow, though it's suspected that gather prefetch actually
// just blocks.
//
// TODO: Try hitting each 64-byte cluster with a regular prefetch. Or maybe just
// the first one in a span, if we can hit it really early like as soon as we have
// the grey object bitmap?
//
//#define DOGATHERPREFETCH

GLOBL frameStarts<>(SB), RODATA, $0x40
DATA  frameStarts<>+0x00(SB)/8, $0x0000004000000000
DATA  frameStarts<>+0x08(SB)/8, $0x000000c000000080
DATA  frameStarts<>+0x10(SB)/8, $0x0000014000000100
DATA  frameStarts<>+0x18(SB)/8, $0x000001c000000180
DATA  frameStarts<>+0x20(SB)/8, $0x0000024000000200
DATA  frameStarts<>+0x28(SB)/8, $0x000002c000000280
DATA  frameStarts<>+0x30(SB)/8, $0x0000034000000300
DATA  frameStarts<>+0x38(SB)/8, $0x000003c000000380

// This assumes USESTACK, PPOPCNT, !COMPRESSLOOP, !GATHERLOAD
TEXT ·scanSpanPackedAVX512Lzcnt(SB), NOSPLIT, $0-44
    // Z1+Z2 = Expand the grey object mask into a grey mask
    MOVQ objDarts+16(FP), AX
    MOVQ sizeClass+24(FP), CX
    LEAQ ·gcExpanders(SB), BX
    CALL (BX)(CX*8)

    // Z3+Z4 = Load the pointer mask
    MOVQ ptrMask+32(FP), AX
    VMOVDQU64 0(AX), Z3
    VMOVDQU64 64(AX), Z4

    // Z1+Z2 = Combine the grey mask with the pointer mask to get the scan mask
    VPANDQ Z1, Z3, Z1
    VPANDQ Z2, Z4, Z2

    // Now each bit of Z1+Z2 represents one word of the span.
    // Thus, each byte covers 64 bytes of memory, which is also how
    // much we can fix in a Z register.
    //
    // We do a load/compress for each 64 byte frame.

    // SI = Current address in span
    MOVQ mem+0(FP), SI
    // DI = Scan buffer base
    MOVQ buf+8(FP), DI
    // DX = Index in scan buffer, (DI)(DX*8) = Current position in scan buffer
    MOVQ $0, DX

    CALL scanLzcnt<>(SB)
    VMOVDQA64 Z2, Z1
    ADDQ $(64*64), SI
    CALL scanLzcnt<>(SB)

    MOVL DX, count+40(FP)
    // We're done with our AVX-512 registers.
    VZEROUPPER
    RET

TEXT scanLzcnt<>(SB), NOSPLIT, $128-0
    // Store the scan mask at 0(SP)
    VMOVDQU64 Z1, 0(SP)

    // K1 = AX = Mask of non-empty 64-byte frames
    VPTESTMB Z1, Z1, K1 // Requires AVX512BW
    KMOVQ K1, AX

#ifdef DOGATHERPREFETCH
    VMOVDQU64 frameStarts<>(SB), Z6
    VGATHERPF0DPS K1, (SI)(Z6*1)
#endif

    // R8 = Loop count
    //
    // It's a decent bit faster to loop to this count rather than
    // testing the Z flag after the BSFQ.
    //
    // TODO: If this is above some threshold (~50% or 32 set bits),
    // it's faster to just hit all of the cluters. Though it only
    // really matters if it's in cache, so maybe it's better to
    // keep it simple.
    POPCNTQ AX, R8

    // Store the frame counts at 64(SP)
    VPOPCNTB Z1, Z3
    VMOVDQU64 Z3, 64(SP)

    JMP loophead

    PCALIGN $64
loopbody:
    // BX = Frame index
    BSFQ AX, BX

    // Clear that bit from the mask in AX
    // Using BLSRQ rather than BTRQ is slightly faster, probably
    // because it reduces dependencies.
    BLSRQ AX, AX

    // CX = K1 = Fetch the mask of words to load from this frame.
    MOVBQZX (SP)(BX*1), CX
    KMOVB CX, K1

    // CX = Number of pointers in this frame
    MOVBQZX 64(SP)(BX*1), CX

    // BX = 64 * BX = BX << 6 == memory offset of frame
    SHLQ $6, BX
    // Z1 = Load the 64 byte frame
    //
    // TODO: VMOVNTDQA for non-temporal load?
    VMOVDQA64 (SI)(BX*1), Z1

#ifdef FILTERNIL
    // Filter out nil pointers
    VPCMPEQQ Z0, Z1, K2
    KANDB K1, K1, K2
#endif

    // Collect just the pointers from the greyed objects into the scan buffer.
    VPCOMPRESSQ Z1, K1, (DI)(DX*8)
    // Advance scan buffer position
    ADDQ CX, DX

    // Decrement loop counter and loop if != 0.
    DECQ R8
loophead:
    // TESTQ and JNZ will macro-fuse
    TESTQ R8, R8
    JNZ loopbody

loopend:
    RET


// GLOBL offsets<>(SB), RODATA, $0x40
// DATA  offsets<>+0x00(SB)/8, $0x0000000100020003
// DATA  offsets<>+0x08(SB)/8, $0x0004000500060007
// DATA  offsets<>+0x10(SB)/8, $0x00080009000a000b
// DATA  offsets<>+0x18(SB)/8, $0x000c000d000e000f
// DATA  offsets<>+0x20(SB)/8, $0x0010001100120013
// DATA  offsets<>+0x28(SB)/8, $0x0014001500160017
// DATA  offsets<>+0x30(SB)/8, $0x00180019001a001b
// DATA  offsets<>+0x38(SB)/8, $0x001c001d001e001f
// 
// TEXT ·scanSpanPackedAVX512Gather(SB), 0, $2176-44
//     // Z1+Z2 = Expand the grey object mask into a grey mask
//     MOVQ objDarts+16(FP), AX
//     CALL expand3<>(SB) // XXX Use size class
// 
//     // Z3+Z4 = Load the pointer mask
//     MOVQ ptrMask+32(FP), AX
//     VMOVDQU64 0(AX), Z3
//     VMOVDQU64 64(AX), Z4
// 
//     // Z1+Z2 = Combine the grey mask with the pointer mask to get the scan mask
//     VPANDQ Z1, Z3, Z1
//     VPANDQ Z2, Z4, Z2
// 
//     // Store the scan mask at 0(SP).
//     VMOVDQU64 Z1, 0(SP)
//     VMOVDQU64 Z2, 64(SP)
// 
//     // Now each bit of Z1+Z2 represents one word of the span.
//     // Convert these bits into a set of offsets to load.
//     // DI = Scan buffer base
//     MOVQ buf+8(FP), DI
//     VMOVDQU64 offsets<>(SB), Z3
//     MOVQ $0x20, AX
//     VPBROADCASTW AX, Z4
// 
//     // Process the scan mask 32 bits at a time to construct the offsets list.
//     //
//     // TODO: Should we use PEXTRD/VEXTRACTI64X4 to get these straight from the registers?
//     MOVQ $0, CX
//     LEAQ 128(SP), DI // DI = pointer to offset buffer
//     MOVQ $0, DX      // DX = index in offset buffer
// loop:
//     // AX = K1 = Load 32 bit mask
//     MOVL (SP)(CX*4), AX
//     KMOVD AX, K1
//     // Collect 2-byte offsets corresponding to set bits
//     VPCOMPRESSW Z3, K1, (DI)(DX*2)  // Requires AVX512_VBMI2
//     // Update offsets
//     VPADDW Z3, Z4, Z3
//     // Advance offset buffer index by the number of entries we just added
//     POPCNTQ AX, BX
//     ADDQ BX, DX
//     // Advance to the next 32 bits of the scan mask.
//     ADDQ $1, CX
//     CMPQ CX, $32
//     JB loop
// 
//     // Now, DX = # of elements in offset buffer
// 
//     // Now use gather loads to load quadwords using the offset list into the scan buffer.
// 
//     // XXX Need to mask the tail of this loop
// 
//     // SI = Base address of span
//     MOVQ mem+0(FP), SI
//     // AX = Scan buffer base
//     MOVQ buf+8(FP), AX
//     // CX = Index in offset buffer
//     // (DI)(CX*2) = Current position in offset buffer
//     // (AX)(CX*8) = Current position in scan buffer
//     MOVQ $0, CX
// gather:
//     // Z1 [32]uint16 = Next 32 indexes
//     VMOVDQU64 (DI)(CX*2), Z1
// 
//     // Z2 [8]uint64 = Expand indexes Z1[0:8] to 64 bits
//     VPMOVZXWQ X1, Z2
//     // Z2 [8]uint64 = Gather load 8 indexes from span
//     KMOVQ K0, K1
//     VPGATHERQQ (SI)(Z2*8), K1, Z3
//     // Write scanned pointers into scan buffer
//     VMOVDQU64 Z3, (AX)(CX*8)
// 
//     // Z1 = Rotate Z1 right by 8 elements.
//     VSHUFI64X2 $0b00111001, Z1, Z1, Z1
//     // Process elements [8:16]
//     VPMOVZXWQ X1, Z2
//     KMOVQ K0, K1
//     VPGATHERQQ (SI)(Z2*8), K1, Z3
//     VMOVDQU64 Z3, 64(AX)(CX*8)
// 
//     VSHUFI64X2 $0b00111001, Z1, Z1, Z1
//     // Process elements [16:24]
//     VPMOVZXWQ X1, Z2
//     KMOVQ K0, K1
//     VPGATHERQQ (SI)(Z2*8), K1, Z3
//     VMOVDQU64 Z3, 128(AX)(CX*8)
// 
//     VSHUFI64X2 $0b00111001, Z1, Z1, Z1
//     // Process elements [24:32]
//     VPMOVZXWQ X1, Z2
//     KMOVQ K0, K1
//     VPGATHERQQ (SI)(Z2*8), K1, Z3
//     VMOVDQU64 Z3, 192(AX)(CX*8)
// 
//     ADDQ $32, CX
//     CMPQ CX, DX
//     JB gather
// 
//     MOVL DX, count+40(FP)
//     RET
