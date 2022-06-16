// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//
// func escapeStringGetLenAVX512(src string) int32
//
#define src             CX        // input string pointer
#define srcLen          R8        // input string length
#define lutPtr          R10       // esc character length lookup table, used to reduce branching
#define char            R11       // 8-byte working buffer
#define escLen          R12       // esc length accumulator / return result
#define tmp             R13       // 8-byte working buffer used for single character esc len lookups
#define x1              X1        // avx working registers for streaming string data in blocks of 16(x)/32(y)/64(z) bytes
#define y1              Y1        // ""
#define z1              Z1        // ""
#define z2              Z2        // ""
#define z3              Z3        // ""
#define z4              Z4        // ""
#define z5              Z5        // ""
#define z6              Z6        // ""
#define z7              Z7        // ""
#define z8              Z8        // z1-z8 = 8 lanes x 64 bytes for streaming strings >= 512 bytes in length
#define str1            Z9        // 64-byte string copy used for 64-byte SIMD parallel esc lookup ops in 2-segment lookup table 
#define escAcc64        Z0        // 8x64b accumulator; each 8-byte lane accumulates 8-bit SAD results to avoid overflow
#define escAcc64y       Y0        // used for horizontal add to reduce 8x64bit esc counts to 1x32bit result
#define escAcc64x       X0        // ""
#define zeroy           Y11       // all zeros in ymm, used for horizontal add to reduce 8x64bit esc counts to 1x32bit result
#define zero            Z11       // all zeros in zmm, used for SAD reduction, also for 2-segment esc lookup table
#define escLen64x8      Z24       // 64 parallel escape length lookup results, 64 lanes x 8 bits
#define escAcc8x        X25       // used for horizontal add to reduce 8x64bit esc counts
#define escAcc8         Z25       // 64x8b length accumulator
#define hreduce8x32LUT  Z27       // lookup table to reduce 64x8 zmm to 16x8 xmm, selects low 8b for each; M392 | M 56 M 48 M 40 M 32 M 24 M 16 M 8 M 0 (each a byte) 
#define escLen64        Z28       // esc len partial lut register, entries 0-63
#define escLen128       Z29       // 64-127
#define escLen192       Z30       // 128-191
#define escLen256       Z31       // 192-256  
#define hreduceMask     K1        // x0000000000005555, i.e., write to bottom 8 x 16 using low 8b for each
#define noMask          K2        // all ones K mask
#define blendMask       K3        // blend mask to merge 2x7b lut results into 1x8b result on 64x8b lanes using hi-order bit

// lut to pack 8x64B --> 8x32B
DATA hreduce8x32<>+0x00(SB)/8, $0x0b0a090803020100 
DATA hreduce8x32<>+0x08(SB)/8, $0x1b1a191813121110                                           
DATA hreduce8x32<>+0x10(SB)/8, $0x2b2a292823222120
DATA hreduce8x32<>+0x18(SB)/8, $0x3b3a393833323130
GLOBL hreduce8x32<>(SB), (NOPTR+RODATA), $64

// mask off all but LS 8-bits from each of zmm 8x64b --> ymm 8x32b
DATA kMask8x64to8x32<>+0x00(SB)/8, $0x00000000ffffffff 
GLOBL kMask8x64to8x32<>(SB), (NOPTR+RODATA), $8

// esc (incremental) length lut
// 34(") &#34; = 4  38(&) &amp; = 4  39(') &#39; = 4  
// 60(<) &lt;  = 3  62(>) &gt;  = 3
DATA lut<>+0x20(SB)/8, $0x0404000000040000 
DATA lut<>+0x38(SB)/8, $0x0003000300000000
GLOBL lut<>(SB), (NOPTR+RODATA), $256

// scalar count x1
#define countEscBytes(char) \
  MOVB                  char, tmp                  /* load next byte */             \
  MOVBQZX               (lutPtr)(tmp*1), tmp       /* look up esc len */            \
  SHRQ                  $8, char                   /* advance to next byte */       \
  ADDQ                  tmp, escLen                /* accumulate esc len */  

// vector count x64
#define countEscBytes64(str) \
  /* get escape byte counts via 8-bit lut */                                        \
  /* implemented using 2 x 7-bit with blend on hi bit  */                           \
  VPCMPGTB              str, zero, noMask, blendMask                                \
  VMOVDQU64             str, str1                                                   \
  VPERMI2B              escLen128, escLen64, noMask, str                            \
  VPERMI2B              escLen256, escLen192, noMask, str1                          \
  VPBLENDMB             str1, str, blendMask, escLen64x8                            \                              
  VPADDB                escLen64x8, escAcc8, noMask, escAcc8

#define load128(src,offset,dst0,dst64) \
  VMOVDQU64             (0+offset)(src), dst0                                       \
  VMOVDQU64             (64+offset)(src), dst64 

#define load256(src,offset,dst0,dst64,dst128,dst192) \
  load128               (src,0+offset,dst0,dst64)                                   \
  load128               (src,128+offset,dst128,dst192)                                            

#define load512(src,dst0,dst64,dst128,dst192,dst256,dst320,dst384,dst448) \
  load256               (src,0,dst0,dst64,dst128,dst192)                            \
  load256               (src,256,dst256,dst320,dst384,dst448)                                            

// func escapeStringGetLenAVX512(s string) int32
//
// find escape characters in the input string
// calculate string length for escaped output string
// to enable single-step memory allocation for EscStringAvx
//
TEXT ·escapeStringGetLenAVX512(SB),0,$0-20
  MOVQ						        s_len+8(FP), srcLen
  CMPQ                    srcLen, $0                              // do nothing for 0-length string
  JG                      init
  MOVL                    srcLen, ret+16(FP)
  RET

init:
  MOVQ						        s_base+0(FP), src                       // set up src and esc lut pointers
  LEAQ                    lut<>(SB), lutPtr
  XORQ                    escLen, escLen                          // zero escLen accumulator / final result and tmp
  XORQ                    tmp, tmp
  CMPQ                    srcLen, $16                             // skip to scalar for short input strings (len<16 bytes)
  JL                      block8

  // init masks and register constants for vector blocks (len>=16 bytes)
  KMOVQ 					        kMask8x64to8x32<>(SB), hreduceMask      // hreduceMask for 8x64 --> 8x32 reduction
	VMOVDQU64				        (64*0)(lutPtr), escLen64                // LUT 0-63
	VMOVDQU64				        (64*1)(lutPtr), escLen128               // LUT 64-127
	VMOVDQU64				        (64*2)(lutPtr), escLen192               // LUT 128-191
	VMOVDQU64				        (64*3)(lutPtr), escLen256               // LUT 192-256
  VMOVDQU64               hreduce8x32<>(SB), hreduce8x32LUT       // reduce 8x64 --> 8x16, selecting LS 16 bits for for each reduction
  VPXORQ                  zero, zero, zero                        // LUT switch constant to detect via CMPGT val(char) in [128, 255] (signed compare)
  KXNORQ 					        noMask, noMask, noMask                  // all ones, i.e., no masked bits
  VPXORQ                  escAcc64, escAcc64, escAcc64            // init outer loop 8x64b lane counts to 0
  CMPQ                    srcLen, $512                            // process long strings (>=512 bytes) in blocks of 512 bytes
  JL                      block256 
                              
  block512:
    SUBQ                  $512, srcLen
    load512               (src,z1,z2,z3,z4,z5,z6,z7,z8)           // get next 512 bytes
    VPXORQ                escAcc8, escAcc8, escAcc8               // init inner loop 64x8b lane counts to 0

    // count total incremental escape bytes on each block, i.e. additional bytes introduced by each escape character
    // 8-bit acc on 64 lanes is sufficient for 8 iterations, max count per lane would be 32 (<<255)
    countEscBytes64       (z1)
    countEscBytes64       (z2)
    countEscBytes64       (z3)
    countEscBytes64       (z4)
    countEscBytes64       (z5)
    countEscBytes64       (z6)
    countEscBytes64       (z7)
    countEscBytes64       (z8)

    // convert to 8x64b for accumulation over x8 loops to avoid overflow
    // use SAD wrt 0 on 64x8b --> 8x64b to compute hadd on the 8b esc lanes
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64

    // update pointers, loop
    ADDQ                  $512, src
    CMPQ                  srcLen, $512
    JGE                   block512

  // process next 256 bytes using same flow as block512, but with only 4 zmm 
  block256:
    CMPQ                  srcLen, $256
    JL                    block128
    load256               (src,0,z1,z2,z3,z4)
    VPXORQ                escAcc8, escAcc8, escAcc8
    countEscBytes64       (z1)
    countEscBytes64       (z2)
    countEscBytes64       (z3)
    countEscBytes64       (z4)
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64
    ADDQ                  $256, src
    SUBQ                  $256, srcLen
    JEQ                   reduce                                   // skip to reduction if processing is complete           

  // process next 128 bytes using same flow as block256, but with only 2 zmm 
  block128:
    CMPQ                  srcLen, $128
    JL                    block64
    load128               (src,0,z1,z2)
    VPXORQ                escAcc8, escAcc8, escAcc8
    countEscBytes64       (z1)
    countEscBytes64       (z2)
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64
    ADDQ                  $128, src
    SUBQ                  $128, srcLen
    JEQ                   reduce

  // process next 64 bytes using same flow as block128, but with only 1 zmm
  block64:
    CMPQ                  srcLen, $64
    JL                    block32
    VMOVDQU64             (src), z1
    VPXORQ                escAcc8, escAcc8, escAcc8
    countEscBytes64       (z1)
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64
    ADDQ                  $64, src
    SUBQ                  $64, srcLen
    JEQ                   reduce

  // process next 32 bytes using same flow as block64; clear upper 32 bytes of input zmm
  block32:
    CMPQ                  srcLen, $32
    JL                    block16
    VPXORQ                z1, z1, z1
    VMOVDQU               (32*0)(src), y1
    VPXORQ                escAcc8, escAcc8, escAcc8
    countEscBytes64       (z1)
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64
    ADDQ                  $32, src
    SUBQ                  $32, srcLen
    JEQ                   reduce

  // process next 16 bytes using same flow as block32
  block16:
    CMPQ                  srcLen, $16
    JL                    reduce
    VPXORQ                z1, z1, z1
    VMOVDQU               (16*0)(src), x1
    VPXORQ                escAcc8, escAcc8, escAcc8
    countEscBytes64       (z1)
    VPSADBW               escAcc8, zero, escAcc8
    VPADDQ                escAcc8, escAcc64, escAcc64
    ADDQ                  $16, src
    SUBQ                  $16, srcLen

  // reduce from vector to scalar accumulator for blocks < 16
  reduce:
    // horizontal add 64x8b individual escape counts into a final result
    VPERMB.Z              escAcc64, hreduce8x32LUT, hreduceMask, escAcc64   // pack 8x64b -> 8x32b in ymm  [ s7 s6 s5 s4 s3 s2 s1 s0 ]  
    VPHADDD               zeroy, escAcc64y, escAcc64y                       // hadd 8x16b -> 4x32b in ymm  [ 0 0 s7+s6 s5+s4 0 0 s3+s2 s1+s0 ]
    VPHADDD               zeroy, escAcc64y, escAcc64y                       // hadd 4x16b -> 2x32b in ymm  [ 0 0 s7+s6+s5+s4 s7+s6+s5+s4 s3+s2+s1+s0 s3+s2+s1+s0 ]
    VEXTRACTI32X4         $1, escAcc64y, noMask, escAcc8x                   // align for last add in 2 xmms [ - - - s7+s6+s5+s4 ] + [ - - - s3+s2+s1+s0 ]
    VPADDD                escAcc8x, escAcc64x, escAcc64x                    // add last 2 partial sums in xmm [ - - - s7+s6+s5+s4+s3+s2+s1+s0 ]
    VMOVD                 escAcc64x, escLen                                 // copy reduction result to return accumulator

  // use scalar ops for block lengths 8, 4, 2, and 1
  // could also implement using 16-byte vectors with countEscBytes64
  // and reduction at the end
  block8:
    CMPQ                  srcLen, $8
    JL                    block4
    MOVQ                  (src), char                 // load next 8 
    countEscBytes         (char)                      // len(0)
    countEscBytes         (char)                      // 1
    countEscBytes         (char)                      // 2
    countEscBytes         (char)                      // 3
    countEscBytes         (char)                      // 4
    countEscBytes         (char)                      // 5
    countEscBytes         (char)                      // 6
    MOVB                  char, tmp                   // 7
    MOVBQZX               (lutPtr)(tmp*1), tmp        
    ADDQ                  $8, src
    ADDQ                  tmp, escLen
    SUBQ                  $8, srcLen
    JEQ                   done

  block4:
    CMPQ                  srcLen, $4
    JL                    block2
    MOVL                  (src), char                 // load next 4 
    countEscBytes         (char)                      // len(0)
    countEscBytes         (char)                      // 1
    countEscBytes         (char)                      // 2
    MOVB                  char, tmp                   // 3
    MOVBQZX               (lutPtr)(tmp*1), tmp        
    ADDQ                  $4, src
    ADDQ                  tmp, escLen
    SUBQ                  $4, srcLen
    JEQ                   done

  block2:
    CMPQ                  srcLen, $2
    JL                    block1
    MOVWQZX               (src), char                 // load next 2
    MOVB                  char, tmp
    MOVBQZX               (lutPtr)(tmp*1), tmp        // len(0)
    SHRQ                  $8, char
    ADDQ                  tmp, escLen                  
    MOVBQZX               (lutPtr)(char*1), tmp       // 1
    ADDQ                  $2, src
    ADDQ                  tmp, escLen
    SUBQ                  $2, srcLen
    JEQ                   done

  block1:
    CMPQ                  srcLen, $1
    JL                    done
    MOVBQZX               (src), char                 // load last char
    MOVBQZX               (lutPtr)(char*1), tmp       // len(0)
    ADDQ                  tmp, escLen

  done:
    ADDQ                  s_len+8(FP), escLen         // add input length to accumulated esc bytes
    MOVL                  escLen, ret+16(FP)
    RET

#undef src
#undef srcLen
#undef lutPtr
#undef char
#undef escLen
#undef tmp
#undef x1
#undef y1
#undef z1
#undef z2
#undef z3
#undef z4
#undef z5
#undef z6
#undef z7
#undef z8
#undef str1
#undef escAcc64
#undef escAcc64y
#undef escAcc64x
#undef zeroy
#undef zero
#undef escLen64x8
#undef escAcc8x
#undef escAcc8
#undef hreduce8x32LUT
#undef escLen64
#undef escLen128
#undef escLen192
#undef escLen256
#undef hreduceMask
#undef noMask
#undef blendMask

//
// func escapeStringAVX512(s string) string
//
#define tmp             AX        // 64-bit temp
#define src             CX        // input string pointer
#define escTab          BX        // esc LUT pointer
#define srcLen          R8        // input string length
#define dst             R9        // output string pointer
#define char80          R10       // 8 byte working buf0
#define char81          R11       // 8 byte working buf1
#define esc             R12       // escape working buf
#define x1              X1        // x16 input string load buf0
#define y1              Y1        // x32   "     "     " 
#define z1              Z1        // x64   "     "     "
#define x2              X2        // x16 input string load buf1
#define y2              Y2        // x32   "     "     "
#define z2              Z2        // x64   "     "     "
#define VSHR256         Z3        // zmm >> 256 control register
#define nomask          K1        // k mask, all ones

// escExpTab
// escape expansion lookup table (sparse)
//
// contains 256 x 4 byte entries, only 5 non-zero
// eliminates branching from the string walk
// for i=0:255,
//   lut[i], bytes [0:3] -- first 4 bytes of the esc expansion
//   lut[i] >> 24 && 0x1 -- 0 for 4-byte expansions, 1 for 5-byte expansions 
//                          used increment the dst pointer 
//                          instead of branching on expansion length
//
DATA escExpTab<>+136(SB)/8, $0x0000000034332326     // 34(") - &#34;
DATA escExpTab<>+152(SB)/8, $0x39332326706D6126     // 38(&) - &amp;  39(') - &#39;
DATA escExpTab<>+240(SB)/8, $0x0000000000746C26     // 60(<) - &lt;
DATA escExpTab<>+248(SB)/8, $0x0000000000746726     // 62(>) - &gt;
GLOBL escExpTab<>(SB), (RODATA|NOPTR), $1024

// vshr256
// zmm shift right 256 control register
//
// contains VPERMQ extract control bits
// to implement z[511:256] --> y[255:0]
//
DATA vshr256<>+0x00(SB)/8, $0x0000000000000004 
DATA vshr256<>+0x08(SB)/8, $0x0000000000000005                                           
DATA vshr256<>+0x10(SB)/8, $0x0000000000000006
DATA vshr256<>+0x18(SB)/8, $0x0000000000000007
DATA vshr256<>+0x20(SB)/8, $0x0000000000000000 
DATA vshr256<>+0x28(SB)/8, $0x0000000000000000
DATA vshr256<>+0x30(SB)/8, $0x0000000000000000 
DATA vshr256<>+0x38(SB)/8, $0x0000000000000000
GLOBL vshr256<>(SB), (NOPTR+RODATA), $64

// write escape expansion for one byte
// lut address arithmetic is used to minimize branching
#define writeEsc(esc,dst) \
  MOVL                  esc, (dst)                                          \
  SHRQ                  $28, esc                                            \
  ANDQ                  $1, esc                                             \
  ADDQ                  esc, dst                                            \
  MOVB                  $';', (3)(dst)                                      \ 
  ADDQ                  $4, dst        

// process one input byte using esc lut
// expand if escape sequence otherwise pass thru
// lut is used to minimize branching
#define processChar(src,dst,charProc,escProc) \
charProc:                                                                   \
  MOVBQZX               src, esc                                            \
  MOVLQZX               (escTab)(esc*4), esc                                \
  CMPB                  esc, $'&'                                           \
  JEQ                   escProc                                             \
  MOVB                  src, (dst)                                          \
  SHRQ                  $8, src                                             \
  ADDQ                  $1, dst 

// process one escape input byte
#define processEsc(src,dst,esc,nextCharProc,escProc) \
escProc:                                                                    \
  SHRQ                  $8, src                                             \
  writeEsc              (esc,dst)                                           \
  JMP                   nextCharProc

// process eight input bytes
// label macro parameters en, n=0,1,2,... and cn, n=0,1,2... 
// are needed to enable use of labels in macro expansions for unrolled processing
// c labels are for character processing; e labels are for escape processing
#define proc8(src,e0,c0,e1,c1,e2,c2,e3,c3,e4,c4,e5,c5,e6,c6,e7,c7,end) \
  processChar           (src,dst,c0,e0)                                     \
  processChar           (src,dst,c1,e1)                                     \
  processChar           (src,dst,c2,e2)                                     \
  processChar           (src,dst,c3,e3)                                     \
  processChar           (src,dst,c4,e4)                                     \
  processChar           (src,dst,c5,e5)                                     \
  processChar           (src,dst,c6,e6)                                     \
  processChar           (src,dst,c7,e7)                                     \
  JMP                   end                                                 \
  processEsc            (src,dst,esc,c1,e0)                                 \
  processEsc            (src,dst,esc,c2,e1)                                 \
  processEsc            (src,dst,esc,c3,e2)                                 \
  processEsc            (src,dst,esc,c4,e3)                                 \
  processEsc            (src,dst,esc,c5,e4)                                 \
  processEsc            (src,dst,esc,c6,e5)                                 \
  processEsc            (src,dst,esc,c7,e6)                                 \
e7:                                                                         \
  writeEsc              (esc,dst)                                           \
end:

// func escapeStringAVX512(src string) string
TEXT ·escapeStringAVX512(SB),0,$0-56
  MOVQ						      src_base+0(FP), src             // get input string pointer
  MOVQ						      src_len+8(FP), srcLen           // get input string length
  MOVQ                  dst_base+16(FP), dst            // get output string pointer
  MOVQ                  dst_len+24(FP), tmp             // get output string length
  MOVQ                  dst, ret_base+40(FP)            // return output string pointer
  MOVQ                  tmp, ret_len+48(FP)             // return output string length
  VMOVDQU64             vshr256<>(SB), VSHR256          // init zmm >> 256 control register
  MOVQ                  $escExpTab<>(SB), escTab        // init esc lut
  KXNORQ                nomask, nomask, nomask          // init all ones K mask
  CMPQ                  srcLen, $64                     // begin processing x64 if possible
  JL                    block32             

// for input strings longer than 64 bytes, process blocks of 64 at a time
block64:
  VMOVDQU64             (src), z1                                                                                         // load next 64 
  VPERMQ                z1, VSHR256, nomask, z2                                                                           // process 8x8
  VMOVQ                 x1, char80 
  ADDQ                  $64, src                    
  SUBQ                  $64, srcLen      
  VPEXTRQ               $1, x1, char81
  VPERMQ                $0xe, y1, y1
  proc8                 (char80,e0,c0,e1,c1,e2,c2,e3,c3,e4,c4,e5,c5,e6,c6,e7,c7,end0)                                     // 0..7
  VMOVQ                 x1, char80
  proc8                 (char81,e8,c8,e9,c9,e10,c10,e11,c11,e12,c12,e13,c13,e14,c14,e15,c15,end1)                         // 8..15
  VPEXTRQ               $1, x1, char81
  proc8                 (char80,e16,c16,e17,c17,e18,c18,e19,c19,e20,c20,e21,c21,e22,c22,e23,c23,end2)                     // 16..23
  VMOVQ                 x2, char80 
  proc8                 (char81,e24,c24,e25,c25,e26,c26,e27,c27,e28,c28,e29,c29,e30,c30,e31,c31,end3)                     // 24..31
  VPEXTRQ               $1, x2, char81
  VPERMQ                $0xe, y2, y2
  proc8                 (char80,e32,c32,e33,c33,e34,c34,e35,c35,e36,c36,e37,c37,e38,c38,e39,c39,end4)                     // 32-39       
  VMOVQ                 x2, char80
  proc8                 (char81,e40,c40,e41,c41,e42,c42,e43,c43,e44,c44,e45,c45,e46,c46,e47,c47,end5)                     // 40-47
  VPEXTRQ               $1, x2, char81
  proc8                 (char80,e48,c48,e49,c49,e50,c50,e51,c51,e52,c52,e53,c53,e54,c54,e55,c55,end6)                     // 48-55
  proc8                 (char81,e56,c56,e57,c57,e58,c58,e59,c59,e60,c60,e61,c61,e62,c62,e63,c63,end7)                     // 56-63
  CMPQ                  srcLen, $64                
  JGE                   block64

// process the remaining 32 bytes with x32 load
block32:
  CMPQ                  srcLen, $32
  JL                    block16
  VMOVDQU32             (src), y1                                                                                         // load next 32
  VMOVQ                 x1, char80
  ADDQ                  $32, src                    
  SUBQ                  $32, srcLen      
  VPEXTRQ               $1, x1, char81
  VPERMQ                $0xe, y1, y1
  proc8                 (char80,e64,c64,e65,c65,e66,c66,e67,c67,e68,c68,e69,c69,e70,c70,e71,c71,end8)                     // 0..7
  VMOVQ                 x1, char80
  proc8                 (char81,e72,c72,e73,c73,e74,c74,e75,c75,e76,c76,e77,c77,e78,c78,e79,c79,end9)                     // 8..15
  VPEXTRQ               $1, x1, char81
  proc8                 (char80,e80,c80,e81,c81,e82,c82,e83,c83,e84,c84,e85,c85,e86,c86,e87,c87,end10)                    // 16..23
  proc8                 (char81,e88,c88,e89,c89,e90,c90,e91,c91,e92,c92,e93,c93,e94,c94,e95,c95,end11)                    // 24..31

// process remaining 16 bytes, x16 load
block16:
  CMPQ                  srcLen, $16
  JL                    block8
  VMOVDQU16             (src), x1                                                                                         // load next 16 
  VMOVQ                 x1, char80 
  ADDQ                  $16, src                    
  SUBQ                  $16, srcLen      
  VPEXTRQ               $1, x1, char81
  proc8                 (char80,e96,c96,e97,c97,e98,c98,e99,c99,e100,c100,e101,c101,e102,c102,e103,c103,end12)            // 0..7
  proc8                 (char81,e104,c104,e105,c105,e106,c106,e107,c107,e108,c108,e109,c109,e110,c110,e111,c111,end13)    // 8..15

// process remaining 8 bytes, x8 load
block8:
  CMPQ                  srcLen, $8
  JL                    block4
  MOVQ                  (src), char80 
  ADDQ                  $8, src                    
  SUBQ                  $8, srcLen      
  proc8                 (char80,e112,c112,e113,c113,e114,c114,e115,c115,e116,c116,e117,c117,e118,c118,e119,c119,end14)    // 0..7

// process remaining 4 bytes, x4 load
block4:
  CMPQ                  srcLen, $4
  JL                    block2
  MOVL                  (src), char80
  ADDQ                  $4, src 
  SUBQ                  $4, srcLen
  processChar           (char80,dst,c120,e120)
  processChar           (char80,dst,c121,e121)
  processChar           (char80,dst,c122,e122)
  processChar           (char80,dst,c123,e123)
  JMP                   block2
  processEsc            (char80,dst,esc,c121,e120)
  processEsc            (char80,dst,esc,c122,e121)
  processEsc            (char80,dst,esc,c123,e122)
e123:
  writeEsc              (esc,dst) 

// process remaining 2 bytes, x2 load
block2:
  CMPQ                  srcLen, $2
  JL                    block1
  MOVW                  (src), char80
  ADDQ                  $2, src
  SUBQ                  $2, srcLen
  processChar           (char80,dst,c124,e124)
  processChar           (char80,dst,c125,e125)
  JMP                   block1
  processEsc            (char80,dst,esc,c125,e124)
e125:
  writeEsc              (esc,dst) 

// process last byte
block1:
  CMPQ                  srcLen, $1
  JL                    done
  MOVB                  (src), char80
  MOVBQZX               char80, esc 
  MOVLQZX               (escTab)(esc*4), esc
  CMPB                  esc, $'&'
  JEQ                   escProc 
  MOVB                  char80, (dst)  
  RET

escProc: 
  writeEsc              (esc,dst) 

done:
  RET
  