// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"
#include "entityTabs.dat"             // mapN() hash tables: string tables, associated value tables, and av constants

////////////////////////////
// general-purpose registers
////////////////////////////
#define src             AX            // input string base pointer
#define dst             BX            // output string base pointer
#define dstLen          BX            // output string len (BX reuse)
#define shift           CX            // byte index in quadword string compare buffer, current entity byte
#define t0              CX            // temp for encode rune ops
#define avPtr           SI            // hash function associated values base ptr for mapN(), N=2,3,..,32
#define s2nPtr          SI            // string to decimal value conversion LUT base ptr for hex or decimal numeric escapes     
#define delimPtr        SI            // escape sequence delimiter LUT base ptr; LUT(a-f,A-F,0-9)=1, LUT(any other byte)=0
#define stPtr           DI            // hash function string table base ptr
#define mapPtr          BP            // map function jump table base ptr
#define utfWriter       DX            // utfWriter-N jump table; write N bytes to (dst), N=0,1,..,8; return to findAmpersand
#define writer          DX            // write-N jump table; write N bytes to (dst), N=0,1,..,8; return to unescapeMainLoop
#define srcLen          R8            // input string length
#define c               R9            // constant val for '&' detection
#define escLen          R9            // entity length, used to index mapN jump table
#define hash            R9            // hash value for current entity
#define val             R9            // numeric entity value accumulator
#define hexChar         R9            // hex numeric entity hex character, used to copy hex character for invalid/partial hex entity
#define idx             R10           // ampersand seek block-local '&' byte index (equivalent to go strings.Index())
#define nextChar        R10           // next input string character to be processed
#define hashStr         R10           // hash string from hash string table for mapN hash validation, for mapN, N=2,3,..,8
#define result          R11           // amersand seek (strings.Index) return value, index of '&'
#define writeMask       R11           // write mask for partial xmm write to (dst)
#define readMask        R11           // hash string compare read mask for hash string mapN validation reads from string table
#define msb             R11           // most significant byte select control register for accumulating mapN bytes in entity32, N>=8
#define stringMatch     R11           // entity string comparison result byte mask for strcmp(hash,input) in mapN, N>=8
#define buf8            R11           // quadword (8-byte) write buffer
#define buf4            R11           // double word (4-byte) write buffer
#define utf8            R11           // utf8-encoded entity representation
#define assocVal        R11           // hash associated values
#define digit           R11           // numeric entity digit
#define t1              R11           // temp for rune encode, also used for buffering MS quadword in current string
#define inv             R12           // complemented block for quadword '&' detection step during ampersand search
#define inputStr        R12           // input entity string accumulator (entity bytes are scanned bytewise)
#define quadSelect      R13           // initializer, value=1, for quadword select; used to index next quadword during entity32 accumulation
#define x               R13           // temp for quadword '&' detection step during ampersand search
#define numUtfBytes     R14           // number of utf bytes used to represent the current entity, retrieved from hash string table
#define t2              R14           // temp used for encode rune
#define isAlphaNumeric  R14           // entity accumulator loop LUT result for current character, LUT(a-f,A-F,0-9)=1, LUT(other)=0
#define dstBase         R15           // used to compute return string base and length
#define strMask         R15           // mask to select string bytes during string table lookups in hashN/mapN, N=2,3,..,8

///////////////////
// vector registers
///////////////////
#define noMask          K1            // constant all ones; no lanes masked for avx512 ops
#define e0              K2            // byte 0-63 vector esc search result, byteN=1 for '&', byteN=0 any other character, N=0,1,2,...63
#define writeMaskK      K2            // write mask for bytewise vector writes
#define readMaskK       K2            // read mask for bytewise vector reads
#define cmpMaskK        K2            // compare mask for hash string validation in mapN, N>8
#define msbK            K2            // mask for copying most significant entity bits from current quadword to entity32 vector
#define stringMatchK    K3            // entity string comparison result byte mask for strcmp(hash,input) in mapN, N>=8
#define e1              K3            // byte 64-127 vector esc search result for long block vector '&' compares, same as e0
#define e2              K4            // byte 128-191 "                "              " 
#define e3              K5            // byte 192-255 "                "              " 
#define ex              K6            // ex, ey: escape detection accumulators for multi-block compare
#define ey              K7            //         for 64 < len(input block) <= 256                      
#define str             X0            // entity string buffer for len(entity) <= 16 
#define str32           Y0            // entity string buffer for 16 < len(entity) <= 32
//#define str64           Z0
#define satMin16        X1            // associated value constant for hash arithmetic using compressed table and saturating ops
#define satMin          Y1            // associated value constant for hash arithmetic using compressed table and saturating ops
#define satMax16        X2            // x16 or x32 lanes used as needed depending on entity length
#define satMax          Y2
#define amp64           Z3            // amp64/32/16: 16/32/64-lane esc constants for '&' search
#define amp32           Y3
#define amp16           X3
#define buf0            Z4            // byte 0-63 input buffer for long block '&' search
#define buf0y           Y4            // byte 0-31
#define buf0x           X4            // byte 0-15
#define strMSB          Y4            // entity32 string buffer most significant quadword shuffle control mask
#define xones           X5            // load constant value=1 into xmm to initialize MS quadword select swtich (=0,1,2,3)
#define ones            Y5            // constant value 1 to increment MS quadword select to buffer each entity quadword in the entity loop
#define buf1            Z5            // byte 64-127 input buffer for long block '&' search
#define buf2            Z6            // byte 128-191 "                "              "
#define buf3            Z7            // byte 192-255 "                "              "
#define entity16        X8            // for unterminated entities causing invalid mapN out, copy back 6 lowest entity bytes from this buffer
#define entity32        Y8            // x32 entity buffer for hash string vs. input entity mapN validation string compares, N>8

//////////////////
// encodeRune defs
//////////////////
#define	rune1Max        0x0000007f    // 1<<7 - 1
#define	rune2Max        0x000007ff    // 1<<11 - 1
#define	rune3Max        0x0000ffff    // 1<<16 - 1
#define maxRune         0x0010ffff
#define	surrogateMin    0x0000d800
#define surrogateMax    0x0000dfff
#define runeError       0x0000fffd
#define maskx           0x0000003f
#define ampHash         0x2326

// encodeRune()
// compute utf-8 representation for the 
// input rune, r, and write to (dst)
// advance dst pointer as needed
// this could replace the utf-8 lookups
// in the entity string table if online 
// which would save the utf-8 byte length
// and encoded utf-8 representation storage
// space in each of the map tables
// currently used only for numeric unescapes
#define encodeRune(r,dst,p,t,r2) \
    CMPL                    r, $rune1Max            \
    JG                      rune_2                  \
    MOVB                    r, (dst)                \
    ADDQ                    $1, dst                 \
    JMP                     encodeRune_exit         \
  rune_2:                                           \
    MOVQ                    r, r2                   \
    CMPL                    r, $rune2Max            \
    JG                      rune_error_check        \
    MOVW                    $0x80c0, p              \
    MOVLQSX                 r, t                    \
    SHRQ                    $6, t                   \
    ORB                     t, p                    \
    ANDQ                    $maskx, r               \
    SHLQ                    $8, r                   \
    ORW                     r, p                    \
    MOVW                    p , (dst)               \
    ADDQ                    $2, dst                 \
    JMP                     encodeRune_exit         \
  rune_error_check:                                 \
    CMPL                    r, $maxRune             \
    JLE                     rune3                   \                     
    CMPL                    r, $surrogateMin        \
    JL                      rune3                   \
    CMPL                    r, $surrogateMax        \
    JG                      rune3                   \
    MOVQ                    $runeError, r2          \
  rune3:                                            \
    CMPL                    r, $rune3Max            \
    JG                      default                 \
    MOVQ                    $0x8080e0, p            \
    MOVLQSX                 r2, t                   \
    SHRQ                    $12, t                  \
    ORB                     t, p                    \
    MOVLQSX                 r2, t                   \                    
    SHRQ                    $6, t                   \
    ANDQ                    $maskx, t               \
    SHLQ                    $8, t                   \
    ORW                     t, p                    \
    MOVW                    p, (dst)                \
    ANDQ                    $maskx, r2              \
    ORB                     $0x80, r2               \
    MOVB                    r2, (2)(dst)            \
    ADDQ                    $3, dst                 \
    JMP                     encodeRune_exit         \
  default:                                          \
    MOVQ                    $0x808080f0, p          \
    MOVLQSX                 r2, t                   \
    SHRQ                    $18, t                  \
    ORB                     t, p                    \
    MOVLQSX                 r2, t                   \                    
    SHRQ                    $12, t                  \
    ANDQ                    $maskx, t               \
    SHLQ                    $8, t                   \
    ORW                     t, p                    \
    MOVLQSX                 r2, t                   \                    
    SHRQ                    $6, t                   \
    ANDQ                    $maskx, t               \
    SHLQ                    $16, t                  \
    ORL                     t, p                    \
    ANDQ                    $maskx, r2              \
    SHLQ                    $24, r2                 \
    ORL                     r2, p                   \
    MOVL                    p, (dst)                \
    ADDQ                    $4, dst                 \
  encodeRune_exit:

// findAmpAndCopy() constant '&' table
// for up to 64-lane byte-wide compares during '&' search
DATA amp<>+0x00(SB)/8, $0x2626262626262626
DATA amp<>+0x08(SB)/8, $0x2626262626262626
DATA amp<>+0x10(SB)/8, $0x2626262626262626
DATA amp<>+0x18(SB)/8, $0x2626262626262626
DATA amp<>+0x20(SB)/8, $0x2626262626262626
DATA amp<>+0x28(SB)/8, $0x2626262626262626
DATA amp<>+0x30(SB)/8, $0x2626262626262626
DATA amp<>+0x38(SB)/8, $0x2626262626262626
GLOBL amp<>(SB), (NOPTR+RODATA), $64

// writeN() jump table function entries 
// with parametric return for N=0,1,2,..,8 bytes
// used for string copy between &entities and for 
// writing utf-8 encoded unescaped entities
#define write0(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    JMP                   return     

#define write1(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVB                  buf8, (dst)             \
    ADDQ                  $1, dst                 \
    JMP                   return     

#define write2(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVW                  buf8, (dst)             \
    ADDQ                  $2, dst                 \
    JMP                   return

#define write3(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVW                  buf8, (dst)             \
    SHRQ                  $16, buf8               \
    MOVB                  buf8, (2)(dst)          \
    ADDQ                  $3, dst                 \
    JMP                   return

#define write4(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVL                  buf8, (dst)             \
    ADDQ                  $4, dst                 \
    JMP                   return

#define write5(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVL                  buf8, (dst)             \
    SHRQ                  $32, buf8               \
    MOVB                  buf8, (4)(dst)          \
    ADDQ                  $5, dst                 \
    JMP                   return

#define write6(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVL                  buf8, (dst)             \
    SHRQ                  $32, buf8               \
    MOVW                  buf8, (4)(dst)          \
    ADDQ                  $6, dst                 \
    JMP                   return

#define write7(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVL                  buf8, (dst)             \
    SHRQ                  $32, buf8               \
    MOVW                  buf8, (4)(dst)          \
    SHRQ                  $16, buf8               \
    MOVB                  buf8, (6)(dst)          \
    ADDQ                  $7, dst                 \
    JMP                   return

#define write8(name,return) \
  TEXT name(SB), NOSPLIT, $0-0                    \
    MOVQ                  buf8, (dst)             \
    ADDQ                  $8, dst                 \
    JMP                   return

// copy N-bytes remaining before next '&' for N=0,1,..,8
// then jump to unescape entity loop to encode found entity
// implemented as a 9-entry jump table
write0( Str0, unescapeMain(SB) )
write1( Str1, unescapeMain(SB) )
write2( Str2, unescapeMain(SB) )
write3( Str3, unescapeMain(SB) )
write4( Str4, unescapeMain(SB) )
write5( Str5, unescapeMain(SB) )
write6( Str6, unescapeMain(SB) )
write7( Str7, unescapeMain(SB) )
write8( Str8, unescapeMain(SB) )

// write N-byte utf-encoded entity data to (dst) for N=0,1,..,8
// then jump to ampersand scanner to find next entity ('&...')
// implemented as a 9-entry jump table
write0( Utf0,findAmpAndCopy(SB) )
write1( Utf1,findAmpAndCopy(SB) )
write2( Utf2,findAmpAndCopy(SB) )
write3( Utf3,findAmpAndCopy(SB) )
write4( Utf4,findAmpAndCopy(SB) )
write5( Utf5,findAmpAndCopy(SB) )
write6( Utf6,findAmpAndCopy(SB) )
write7( Utf7,findAmpAndCopy(SB) )
write8( Utf8,findAmpAndCopy(SB) )

// jump table for StrN and UtfN 
// StrN: copy N bytes string text found bewteen & entities, 
//       and then jump to entity loop to unescape current entity
// UtfN: write N bytes for utf-8 encoded unescape entity outtput, 
//       and then jump to findAmersand processing to seek next & 
DATA write<>+0x00(SB)/8, $Str0(SB)
DATA write<>+0x08(SB)/8, $Str1(SB)
DATA write<>+0x10(SB)/8, $Str2(SB)
DATA write<>+0x18(SB)/8, $Str3(SB)
DATA write<>+0x20(SB)/8, $Str4(SB)
DATA write<>+0x28(SB)/8, $Str5(SB)
DATA write<>+0x30(SB)/8, $Str6(SB)
DATA write<>+0x38(SB)/8, $Str7(SB)
DATA write<>+0x40(SB)/8, $Str8(SB)
DATA write<>+0x48(SB)/8, $Utf0(SB)
DATA write<>+0x50(SB)/8, $Utf1(SB)
DATA write<>+0x58(SB)/8, $Utf2(SB)
DATA write<>+0x60(SB)/8, $Utf3(SB)
DATA write<>+0x68(SB)/8, $Utf4(SB)
DATA write<>+0x70(SB)/8, $Utf5(SB)
DATA write<>+0x78(SB)/8, $Utf6(SB)
DATA write<>+0x80(SB)/8, $Utf7(SB)
DATA write<>+0x88(SB)/8, $Utf8(SB)
DATA write<>+0x90(SB)/8, $findAmpAndCopy(SB)
GLOBL write<>(SB), (RODATA|NOPTR), $152

// scalar scan for '&' in quadword from (src)
// if found return the index in idx and copy to (dst) 
// the input string content up to the found '&' 
// otherwise copy full quadword and fall thru
#define findAmp(ampMask,maskLen,mov,buf) \
  TZCNTQ                  ampMask, idx                  \ 
  SHRQ                    $3, idx                       \
  CMPQ                    ampMask, $0                   \
  JNE                     writePartial842               \
  mov                     buf, (dst)                    \
  SUBQ                    $maskLen, srcLen              \
  ADDQ                    $maskLen, src                 \
  ADDQ                    $maskLen, dst 

// vector scan for one or more '&' in (src), either 
// x16, x32, x64 bytes. if found return index in idx 
// and copy to (dst) the input string content up to the 
// found '&', otherwise copy full block and fall thru
#define vFindAmp(ampMask,numBytes,load,buf,writePartialBlock,mov) \
  mov                     (64*0)(src), buf              \
  VPCMPEQB                buf, ampMask, noMask, e0      \
  KMOVQ                   e0, idx                       \
  TZCNTQ                  idx, idx                      \
  CMPQ                    idx, $64                      \          
  JNE                     writePartialBlock             \
  mov                     buf, (64*0)(dst)              \
  ADDQ                    $numBytes, src                \
  SUBQ                    $numBytes, srcLen             \
  ADDQ                    $numBytes, dst                

// vector partial block write for less than 8/16/32 bytes
// idx gives location of & entity; all bytes up to & are
// written to (dst), then jump to entity processing loop
#define writePartial(buf) \
  MOVQ                    idx, shift                    \
  MOVQ                    $-1, writeMask                \
  SHLQ                    shift, writeMask              \
  NEGQ                    writeMask                     \
  SUBQ                    $1, writeMask                 \
  KMOVQ                   writeMask, writeMaskK         \
  VMOVDQU8                buf, writeMaskK, (0*64)(dst)  \  
  ADDQ                    idx, dst                      \
  ADDQ                    idx, src                      \
  SUBQ                    idx, srcLen                   \
  JMP                     unescapeMain(SB)

// detect whether '&' occurs in a quad word, then
// use scalar findAmp to compute the index
#define search8() \
  /* blocks of size 8/4/2 use the expression */       \
  /* (x-0x01...) & ~x & 0x80... to detect esc */      \
  /* instead of 8/4/2x shift/cmp/branch */            \
  /* to minimize branching */                         \
  CMPQ                    srcLen, $8                  \
  JL                      scan4                       \
  MOVQ                    (src), x                    \
  MOVQ                    x, buf8                     \
  MOVQ                    $0x2626262626262626, c      \
  XORQ                    c, x                        \
  MOVQ                    $0x0101010101010101, c      \
  MOVQ                    x, inv                      \
  SUBQ                    c, x                        \
  NOTQ                    inv                         \
  ANDQ                    inv, x                      \
  MOVQ                    $0x8080808080808080, c      \
  ANDQ                    c, x                        \
  findAmp                 (x,8,MOVQ,buf8) 

// detect whether '&' occurs in a long word, then
// use scalar findAmp to compute index
#define search4() \                                                          
  CMPQ                    srcLen, $4                  \
  JL                      scan2                       \
  MOVLQZX                 (src), x                    \
  MOVQ                    x, buf4                     \
  XORL                    $0x26262626, x              \
  MOVQ                    x, inv                      \
  SUBL                    $0x01010101, x              \
  NOTL                    inv                         \
  ANDL                    inv, x                      \
  ANDL                    $0x80808080, x              \
  findAmp                 (x,4,MOVL,buf4) 

//*****************************************
// findAmpAndCopy()
//
// find next occurrence of '&' in (src)
// copy characers preceding '&' to (dst),
// then jump to the entity loop
//
//*****************************************
TEXT findAmpAndCopy(SB),0,$0-0
  search8()                                                 // start with small block for minimal dense prefetch
  CMPQ                    srcLen, $256                      // then process in descending 2^N blocks, N=8,7,6,...,0
  JL                      scan128

// scan for esc flag via decreasing dyadic-length
// blocks: 256*N, 128, ..., 1, where N = len(s) div 256
scan256:                                                         
  VMOVDQU64               (64*0)(src), buf0                 // load next block
  VMOVDQU64               (64*1)(src), buf1
  VMOVDQU64               (64*2)(src), buf2
  VMOVDQU64               (64*3)(src), buf3
  VPCMPEQB                buf0, amp64, noMask, e0           // check for esc
  VPCMPEQB                buf1, amp64, noMask, e1           // on 256 bytes = 4x64 
  VPCMPEQB                buf2, amp64, noMask, e2           
  VPCMPEQB                buf3, amp64, noMask, e3           
  KORQ                    e0, e1, ex                        // reduce 4x64 to 1x64, any byte>0 
  KORQ                    e2, e3, ey                        // detects potential esc 
  KORQ                    ex, ey, ex
  KMOVQ                   ex, idx
  CMPQ                    idx, $0                           // if any byte==0x26 idx=1
  JNE                     return256                         // write partial block and jump to unescape entity loop
  VMOVDQU64               buf0, (64*0)(dst)                 // otherwise if no '&' copy block and continue scan loop
  VMOVDQU64               buf1, (64*1)(dst)
  VMOVDQU64               buf2, (64*2)(dst)
  VMOVDQU64               buf3, (64*3)(dst)
  ADDQ                    $256, src
  SUBQ                    $256, srcLen
  ADDQ                    $256, dst
  CMPQ                    srcLen, $256                      // continue scan loop for >= 256 bytes remaining
  JGE                     scan256
// scan past 256*N using same process for residual 
// blocks of sizes 128, 64, 32, 16, 8, 4, 2, and 1  
scan128:
  CMPQ                    srcLen, $128                          
  JL                      scan64                    
  VMOVDQU64               (64*0)(src), buf2                                     
  VMOVDQU64               (64*1)(src), buf3                                     
  VPCMPEQB                buf2, amp64, noMask, e2              
  VPCMPEQB                buf3, amp64, noMask, e3
  KORQ                    e2, e3, e1                            
  KMOVQ                   e1, idx
  CMPQ                    idx, $0                            
  JNE                     return128                         // if '&' found write partial block and jump to unescape entity loop
  VMOVDQU64               buf2, (64*0)(dst)
  VMOVDQU64               buf3, (64*1)(dst)
  ADDQ                    $128, src
  SUBQ                    $128, srcLen
  ADDQ                    $128, dst
scan64:
  CMPQ                    srcLen, $64                           
  JL                      scan32                                                         
  vFindAmp                (amp64,64,VMOVQDQU64,buf0,writePartial64,VMOVDQU64)
scan32:
  CMPQ                    srcLen, $32                           
  JL                      scan16
  vFindAmp                (amp32,32,VMOVQDQU32,buf0y,writePartial32,VMOVDQU32)
scan16:
  CMPQ                    srcLen, $16                           
  JL                      scan8                                                         
  vFindAmp                (amp16,16,VMOVQDQU16,buf0x,writePartial16,VMOVDQU16)
scan8:                                                          
  search8()
scan4:                                      
  search4()                    
scan2:                                                          
  CMPQ                    srcLen, $2
  JL                      scan1
  MOVWQZX                 (src), x
  MOVQ                    x, buf4
  XORW                    $0x2626, x
  MOVQ                    x, inv
  SUBW                    $0x0101, x
  NOTW                    inv
  ANDW                    inv, x
  ANDW                    $0x8080, x
  findAmp                 (x,2,MOVW,buf4) 
scan1:                                                          
  CMPQ                    srcLen, $1
  JL                      return
  SUBQ                    $1, srcLen
  MOVBQZX                 (src), buf4
  MOVB                    buf4, (dst)
  ADDQ                    $1, dst
return:
  MOVQ                    src_base+0(FP), dstBase
  MOVQ                    dstBase, ret+24(FP)
  SUBQ                    dstBase, dstLen
  MOVQ                    dstLen, ret+32(FP)
  RET

// copy to found '&', update pointers, jump to unescape entity loop
// below are return targets for each residual block size 256, 128, ..., 2
return256:
  KMOVQ                   e0, idx
  TZCNTQ                  idx, idx
  CMPQ                    idx, $64
  JNE                     return2560                    // '&' found in bytes 0..63, do partial write and jump to entity loop
  VMOVDQU64               buf0, (0*64)(dst)             // otherwise not found in 0..63, write full block and continue
  KMOVQ                   e1, idx
  TZCNTQ                  idx, idx
  ADDQ                    $64, dst
  ADDQ                    $64, src 
  SUBQ                    $64, srcLen 
  CMPQ                    idx, $64
  JNE                     return2561                    // '&' found in bytes 64..127, do partial write and jump to entity loop
  VMOVDQU64               buf1, (0*64)(dst)             // otherwise write full block and continue
  ADDQ                    $64, dst
  ADDQ                    $64, src 
  SUBQ                    $64, srcLen 
return128:
  KMOVQ                   e2, idx
  TZCNTQ                  idx, idx
  CMPQ                    idx, $64
  JNE                     return2562                    // '&' in 128..191 (block256) or 0..63 (block128), write and jump
  VMOVDQU64               buf2, (0*64)(dst)
  ADDQ                    $64, dst
  ADDQ                    $64, src 
  SUBQ                    $64, srcLen 
  KMOVQ                   e3, idx
  TZCNTQ                  idx, idx
  writePartial            (buf3)                        // write partial for block of 64 bytes containing the '&'
return2560:
  writePartial            (buf0)                        // in either buf0, buf1, buf2, or buf3
return2561:
  writePartial            (buf1)
return2562:
  writePartial            (buf2)
writePartial64:                                      
  writePartial            (buf0)   
writePartial32:                                      
  writePartial            (buf0y)   
writePartial16:                                      
  writePartial            (buf0x)   
writePartial842:                                        // write partial block for input blocks of size 8, 4, or 2
  ADDQ                    idx, src
  SUBQ                    idx, srcLen
  JMP                     (writer)(idx*8)

// hashInitLong()
// initialize hash computation for mapN() on long entities, i.e., 8 < len(entity) <= 32
// initializes hash tables and associated value saturating arithmetic for compressed av tables
// computes offset associated value indexes as needed using saturating arithmetic and table compression
// returns offset associated value indices in nextChar
#define hashInitLong(avTab,sTab,sMin,sMax,msbMask) \
  VMOVQ                   inputStr, str                 \   /* copy most significant entity quadword to str32 (last 8 bytes) */
  MOVQ                    $msbMask, msb                 \   /* copy most significant entity quadword to entity32 */
  KMOVQ                   msb, msbK                     \
  VPERMQ                  str32, strMSB, msbK, entity32 \   /* move full entity string to entity32 */
  VMOVDQA32               entity32, str32               \   /* copy entity32 to str32 to compute offset assocVal indices */
  VMOVDQU                 sMin<>(SB), satMin            \   /* load compressed assocVal offsets */                         
  VMOVDQU                 sMax<>(SB), satMax            \   /* note hash function offsets are embedded in sat constants */ 
  LEAQ                    avTab<>(SB), avPtr            \   /* load assocVal table base ptr */ 
  LEAQ                    sTab<>(SB), stPtr             \   /* load string table (hash table) base ptr */            
  VPSUBUSB                satMin, str32, str32          \   /* add offsets to compressed assocVal ranges */
  VPADDUSB                satMax, str32, str32          \    
  VPSUBUSB                satMax, str32, str32          \
  VMOVQ                   str, nextChar                     /* copy offset av indices to nextChar to use in av calc */ 

// hashInit()
// initialize hash computation for mapN() on short entities, i.e., len(entity) <= 8
// initializes hash tables and associated value saturating arithmetic for compressed av tables
// computes offset associated value indexes as needed using saturating arithmetic and table compression
// returns offset associated value indices in nextChar
#define hashInit(avTab,sTab,sMin,sMax,sMask) \
  VMOVQ                   inputStr, str                 \   /* copy entity to str (xmm) */
  VMOVDQU16               sMin<>(SB), satMin16          \   /* load compressed assocVal offsets */
  VMOVDQU16               sMax<>(SB), satMax16          \   /* note hash function offsets are embedded in sat constants */
  MOVQ                    sMask<>(SB), strMask          \   /* get mask for strcmp(hash,input entity) */
  LEAQ                    avTab<>(SB), avPtr            \   /* load assocVal table base ptr */
  LEAQ                    sTab<>(SB), stPtr             \   /* load string table (hash table) base ptr */ 
  PSUBUSB                 satMin16, str                 \   /* add offsets to compressed assocVal ranges */ 
  PADDUSB                 satMax16, str                 \ 
  PSUBUSB                 satMax16, str                 \          
  VMOVQ                   str, nextChar                     /* copy offset av indices to nextchar to use in av calc */ 

// computeAV() 
// compute one associated value term in the hash summation
#define computeAV( hash, assocVal, nextChar, op ) \
  MOVBQZX                 nextChar, assocVal                      \
  MOVWQZX                 (avPtr)(assocVal*2), assocVal           \
  SHRQ                    $8, nextChar                            \
  op                      assocVal, hash

// hashToUtf8() 
// compute the final associated value summation term for a short entity, i.e., len(entity)<=8,           
// generate hash, do string lookup, validate against input,
// then do utf-8 lookup and write to (dst)
// handle exceptions on either invalid hash or invalid string 
#define hashToUtf8( hash, assocVal, nextChar, strLen, bytesPerEntry, maxHashVal, strMask ) \
  MOVBQZX                 nextChar, assocVal                      \
  MOVWQZX                 (avPtr)(assocVal*2), assocVal           \
  ADDQ                    assocVal, hash                          \
  /* check for unrecognized input */                              \
  CMPQ                    hash, $maxHashVal                       \
  JG                      unknownString                           \
  /* check for unknown string with valid hash */                  \
  IMUL3Q                  $bytesPerEntry, hash, hash              \
  MOVQ                    (stPtr)(hash*1), hashStr                \
  ANDQ                    strMask, hashStr                        \
  /* get number of utf8 bytes */                                  \
  MOVBQZX                 (strLen)(stPtr)(hash*1), numUtfBytes    \
  CMPQ                    hashStr, inputStr                       \
  JNE                     unknownString                           \    /* backtrack and check for unterminated entity */
  /* get utf8-encoding, then write back and return num bytes */   \
  MOVQ                    (strLen+1)(stPtr)(hash*1), utf8         \
  JMP                     (72)(utfWriter)(numUtfBytes*8)

// hashToUtf8Long() 
// compute the final associated value summation term for a long entity, i.e., 8 < len(entity) <= 32,           
// generate hash, do string lookup, validate against input,
// then do utf-8 lookup, write to (dst), and jump to findAmpersand loop
// handle exceptions on either invalid hash or invalid string 
#define hashToUtf8Long( hash, assocVal, nextChar, strLen, bytesPerEntry, maxHashVal, strMask ) \
  MOVBQZX                 nextChar, assocVal                      \
  MOVWQZX                 (avPtr)(assocVal*2), assocVal           \
  ADDQ                    assocVal, hash                          \
  /* check for unrecognized input */                              \ 
  CMPQ                    hash, $maxHashVal                       \
  JG                      unknownString                           \
  /* check for unknown string with valid hash */                  \
  IMUL3Q                  $bytesPerEntry, hash, hash              \
  MOVQ                    $strMask, readMask                      \
  KMOVQ                   readMask, readMaskK                     \
  VMOVDQU32.Z             (stPtr)(hash*1), readMaskK, str32       \
  VPCMPEQB                str32, entity32, cmpMaskK, stringMatchK \
  KMOVQ                   stringMatchK, stringMatch               \
  CMPL                    stringMatch, $strMask                   \
  JNE                     unknownString                           \     /* backtrack and check for unterminated entity */
  /* get number of utf8 bytes */                                  \
  MOVBQZX                 (strLen)(stPtr)(hash*1), numUtfBytes    \
  /* get utf8-encoding, then write back and return num bytes */   \
  MOVQ                    (strLen+1)(stPtr)(hash*1), utf8         \
  JMP                     (72)(utfWriter)(numUtfBytes*8)

// nop()
// nop for entity bytes occuring in between entity32 buffer updates,
// i.e., bytes on non-quadword boundaries
#define nop() 

// bufferQuadword()
// update entity32, the entity string ymm accumulator, for len(entity) >= 8 bytes
// copy most recent 8-byte entity sub-block into the 8-byte block of entity32
// inexed by strMSB, where strMSB=0,1,2,3; after update, set strMSB = strMSB + 1
#define bufferQuadword() \
  MOVQ                    $1, t1                                  \  /* create constant value of one in xones / ones */
  VMOVQ                   t1, xones                               \ 
  VMOVQ                   inputStr, str                           \  /* copy current 8-byte block to xmm */ 
  VPERMQ                  str32, strMSB, writeMaskK, entity32     \  /* copy MS 8 entity bytes to entity32 */
  KSHIFTLB                $1, writeMaskK, writeMaskK              \  /* update copy mask for next 8 bytes */
  VPADDQ                  ones, strMSB, strMSB                    \  /* strMSB = strMSB + 1 for next 8 bytes */
  MOVQ                    $0, shift                               \  /* clear inputStr shift to align with new accumulator */ 
  XORQ                    inputStr, inputStr                         /* clean inputStr to accumulate next 8 bytes */ 

// getEntityByte()
// get next entity byte executes one step in an 8-byte loop
// that accumulates 8-byte blocks from (src) to the inputStr quadword,
// buffering up to 4 full quadwords to 32-byte accumulator entity32
// stops accumulating and exits to entity map on either: 
//  a) non alpha-numeric character,
//  b) semicolon
//  c) len(entity) > 32 (longest entity is 32 bytes)
//  d) end of input string
#define getEntityByte(updateBuffer,isSemicolon) \
  MOVBQZX                 (delimPtr)(nextChar*1), isAlphaNumeric  \  
  CMPB                    isAlphaNumeric, $0                      \
  JEQ                     isSemicolon                             \  /* exit on non alpha-numeric */
  updateBuffer            ()                                      \  /* update buffer on quadword boundaries */
  SHLQ                    shift, nextChar                         \
  ADDQ                    $1, escLen                              \
  CMPQ                    escLen, $32                             \
  JG                      invalidEntity                           \  /* exit on len(entity) > 32 */
  ADDQ                    $8, shift                               \
  ORQ                     nextChar, inputStr                      \
  ADDQ                    $1, src                                 \
  SUBQ                    $1, srcLen                              \
  CMPQ                    srcLen, $0                              \
  JEQ                     map                                     \
  MOVBQZX                 (src), nextChar        

// checkSemicolon()
// on each new received entity byte
// check for terminating semicolon, then map to utf-8
// on bytes 8, 16, 24, and 32: updateBuffer = bufferQuadword()
// on all other bytes: updateBuffer = nop() 
#define checkSemicolon(updateBuffer,getSemicolon) \
getSemicolon:                                                     \                                      
  CMPB                    nextChar, $';'                          \
  JNE                     map                                     \
  updateBuffer()                                                  \
  SHLQ                    shift, nextChar                         \
  ADDQ                    $1, escLen                              \
  ORQ                     nextChar, inputStr                      \
  ADDQ                    $1, src                                 \
  SUBQ                    $1, srcLen                              \
  JMP                     (mapPtr)(escLen*8)                        

// mapEntity()
// map buffered entity to utf-8
// number of entity bytes (esclen) indexes mapN() jump table
// to minimze branching
#define mapEntity() \
  checkSemicolon          (nop,semi0)                             \  /* semi0: no buffering required */
  JMP                     (mapPtr)(escLen*8)                      \                
  checkSemicolon          (bufferQuadword,semi1)                  \  /* semi1: buffer MS 8 on quadword boundaries */
map:                                                              \ 
  JMP                     (mapPtr)(escLen*8)   

// getEntityFirstByte()
// get next byte in the input stream from (src), presumed
// to be a valid first entity byte following a found '&'
#define getFirstByte() \
  getEntityByte           (nop,semi0)

// getQuadword()
// get next up to 8 entity bytes
// if loop completes 8 iterations then 
// copy resulting quadword to the ymm buffer entity32 and repeat,
// otherwise on occurrence of any exit conditon semi0 and semi1 are 
// non-buffered and buffered exit branch targets, respectively
#define getQuadword() \
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (nop,semi0)                             \ 
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (nop,semi0)                             \
  getEntityByte           (bufferQuadword,semi1)

//********************************************
// unescapeStringAVX512()
//
// html.unescapeString() external entry point
// process input string bytewise, converting 
// &<escape entities> to utf-8
// encoded representations
//
//********************************************
TEXT ·unescapeStringAVX512(SB),0,$0-40
  MOVQ						        s_base+0(FP), src                     // input string base and length
  MOVQ                    s_len+8(FP), srcLen
  MOVQ                    src, dst                              // in-place op; dst base ptr == src base ptr
  KXNORQ 					        noMask, noMask, noMask                // all 1 / no masked bits for any avx512 op
  VMOVDQU64               amp<>(SB), amp64                      // load '&' (escape delimiter) into 64 x 1 byte lanes
  LEAQ                    write<>(SB), writer                   // utf-8 and partial block writer jump table base ptr
  LEAQ                    map<>(SB), mapPtr                     // string table (hash table) 
  JMP                     findAmpAndCopy(SB)                    // seek to next '&'

// unescapeMain()
// internal entity loop entry point
// control transfers to here from 
// findAmpAndCopy() after a found '&' 
TEXT unescapeMain(SB),0,$0-64           
  CMPQ                    srcLen, $0                            // exit on end of string
  JEQ                     done

entityType:                                                     // discern &entity type, either numeric or string
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  CMPQ                    srcLen, $0                            // exit early if insufficient data after '&'
  JEQ                     earlyExit                 
  MOVBQZX                 (src), nextChar
  CMPB                    nextChar, $'#'                        // branch to numeric entity processing if '&#'
  JEQ                     entityNumeric
  JMP                     esc                                   // otherwise branch to string entity processing

entityNumeric:                                                  // process numeric entity
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  CMPQ                    srcLen, $0
  JEQ                     ampHashProc                           // exit early if string ends after '&#' 
  MOVBQZX                 (src), nextChar
  CMPB                    nextChar, $'x'                        // discern whether hex or decimal entity
  JEQ                     entityHex
  CMPB                    nextChar, $'X'
  JEQ                     entityHex                             // branch to hex processing on '&#x' or '&#X'
  XORQ                    val, val                              // init value accumulator
  LEAQ                    dec2dec<>(SB), s2nPtr                 // init ascii to decimal value conversion LUT

entityDec:                                                      // decimal entity decoder loop init
  MOVBQZX                 (s2nPtr)(nextChar*1), digit           // check first character after '&#'
  CMPB                    digit, $0xff
  JEQ                     ampHashProc                           // exit early if not a digit
  ADDQ                    digit, val                            // otherwise accumulate value of the first digit

entityDecLoop:                                                  // decimal entity decode loop
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  CMPQ                    srcLen, $0
  JEQ                     encodeValToUtf8                       // exit and encode val to utf-8 if out of data
  MOVBQZX                 (src), nextChar                       
  MOVBQZX                 (s2nPtr)(nextChar*1), digit           // accumlate sum(digit*10^n), n=0,1,..,len
  CMPB                    digit, $0xff                          // using LUT(next byte) 
  JEQ                     encodeValToUtf8                       // exit and encode val to utf-8 on non-digit
  IMUL3Q                  $10, val, val                         // on each iteration increment power of 10
  ADDQ                    digit, val                            // accumulate result in val
  JMP                     entityDecLoop

entityHex:                                                      // hex entity decode loop init
  MOVQ                    nextChar, hexChar
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  CMPQ                    srcLen, $0
  JEQ                     ampHashxProc                          // exit early if insufficient data after '&#x' or '&#X'
  MOVBQZX                 (src), nextChar
  LEAQ                    hex2dec<>(SB), s2nPtr                 // load ascii hex digit to decimal value LUT
  MOVBQZX                 (s2nPtr)(nextChar*1), digit           // load first hex digit
  CMPB                    digit, $0xff
  JEQ                     ampHashxNonDigitProc                  // exit early if nott a hex digit
  XORQ                    val, val                              // init value accumulator

entityHexLoop:                                                  // hex entity decoder loop
  ADDQ                    digit, val                            // accumulate value of the current digit
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  CMPQ                    srcLen, $0
  JEQ                     encodeValToUtf8                       // exit and encode val to utf-8 if out of data
  MOVBQZX                 (src), nextChar
  MOVBQZX                 (s2nPtr)(nextChar*1), digit
  CMPB                    digit, $0xff
  JEQ                     encodeValToUtf8                       // exit and encode val to utf-8 if non-digit encountered
  SHLQ                    $4, val                               // otherwise increment power of 16 and loop for next digit
  JMP                     entityHexLoop

encodeValToUtf8:                                                // utf-8 encoder entry point for dec or hex entity
  CMPQ                    val, $0x0
  JEQ                     replaceZero
  CMPQ                    val, $0x80                            // branch to replacement table for 0x80 < val < 0x9f
  JL                      encode                                // otherwise apply utf-8 rune encoding
  CMPQ                    val, $0x9f
  JG                      encode

replace:                                                        // use replacement table for 0x80 < val < 0x9f
  LEAQ                    replacementTable<>(SB), s2nPtr
  SUBQ                    $128, val
  MOVQ                    (s2nPtr)(val*8), digit
  MOVL                    digit, (dst)
  SHRQ                    $32, digit
  ADDQ                    digit, dst 
  CMPB                    nextChar, $';'
  JNE                     findNextEntity                        // skip ';' if terminated numeric entity
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
findNextEntity:
  JMP                     (144)(utfWriter)                      // jump to findAmpAndCopy() to find next entity

replaceZero:
  MOVW                    $0xfffd, val                          // replace invalid w/ 0xfffd

encode:                                                         // encode rune to utf-8 and write to (dst)
  encodeRune              (val,dst,t0,t1,t2)
  CMPB                    nextChar, $';'
  JNE                     findNextEntity
  ADDQ                    $1, src 
  SUBQ                    $1, srcLen
  JMP                     (144)(utfWriter)                      // goto findAmpAndCopy() to find next entity

ampHashProc:                                                    // "&#" exception, goto findAmpAndCopy() 
  MOVW                    $ampHash, (dst)
  ADDQ                    $2, dst
  JMP                     (144)(utfWriter)

ampHashxProc:                                                   // "&#x" or "&#X" exception, goto findAmpAndCopy() 
  MOVW                    $ampHash, (dst)
  MOVB                    nextChar, (2)(dst)
  ADDQ                    $3, dst
  JMP                     (144)(utfWriter)

ampHashxNonDigitProc:                                           // "&#x|X<nonDigit>" exception, goto findAmpAndCopy()
  MOVW                    $ampHash, (dst)
  MOVB                    hexChar, (2)(dst)
  ADDQ                    $3, dst
  JMP                     (144)(utfWriter)

earlyExit:                                                      // exit early if insufficient data after '&'
  MOVB                    nextChar, (dst)
  ADDQ                    $1, dst
  JMP                     done

esc:                                                            // non-numeric entity found
  LEAQ                    delim<>(SB), delimPtr                 // init string entity processing loop
  XORQ                    escLen, escLen                        // clear control variables and string buffers
  XORQ                    shift, shift                           
  XORQ                    inputStr, inputStr
  MOVQ                    $1, quadSelect                        // point to first quadword in entity32 buffer
  KMOVQ                   quadSelect, writeMaskK
  VPXORQ                  strMSB, strMSB, strMSB 
  VPXORQ                  str32, str32, str32
  getFirstByte()

// main non-numeric entity loop, maps N-byte string entity to utf-8, where N=2,3,..,32, and writes utf-8 to (dst)
// in the loop, buffer entity bytes from [0..9, a..f, A..F] until
// non entity byte or semicolon occurs; buffer blocks of 8 in entity32 (ymm) up to 32 bytes
// exit loop on either ';', non-alpha, or len>32
// upon loop exit map entity to utf-8 and write to (dst)
entityLoop:
  getQuadword()                                                 // accumulate up to 8 bytes in inputStr and entity32
  JMP                     entityLoop                            // repeat for up to 32 bytes until exit condition occurs
  mapEntity()                                                   // map entity to utf-8 on exit from accumulator loop and write to (dst)

// entity loop exit for len(entity) > 32
// check for unterminated entity with len(entity) == 6,5,...,2
// backtrack and try in descending order map6(), map5(), ..., map2()
invalidEntity: 
  VMOVQ                   entity16, inputStr 
  MOVQ                    $0xffffffffffff, strMask 
  ANDQ                    strMask, inputStr 
  SUBQ                    $26, src 
  ADDQ                    $26, srcLen 
  MOVQ                    $6, escLen 
  JMP                     (mapPtr)(escLen*8) 

// entity loop exit branch target for early termination on incomplete entity or 0-length string
done:
  MOVQ                    src_base+0(FP), dstBase
  MOVQ                    dstBase, ret+24(FP)
  SUBQ                    dstBase, dstLen
  MOVQ                    dstLen, ret+32(FP)
  RET

// unknownStringShort()
// mapN unknown string exit for N <=8
// backtracks by one character and calls mapN-1 to test for len N-1 unterminated entity
// unknownStringShort is called for an unknown &<entity>, which is reflected in one of 
// two conditions, either: a) invalid hash b) valid hash but unknown string
// for N=len(entity)<=6 descending maps of lengths N-1, N-2, N-3, ..., 2 
// are tried for a match to cover unterminated entities of len <= 6
// before flushing unrecognized characters to output string
#define unknownStringShort(srcOffset,mask,map) \
unknownString:                                          \
  SUBQ                    $srcOffset, src               \   /* backtrack 1 character */
  ADDQ                    $srcOffset, srcLen            \
  MOVQ                    $mask, strMask                \   /* adjust strcmp() mask for map N-1 */
  ANDQ                    strMask, inputStr             \
  JMP                     map(SB)                           /* call mapN-1 */

//************************************************
// map2(), map3(), ... map17to32()
//
// mapN converts length N entity to utf-8
// and writes the output to (dst) 
// mapN use gperf-derived perfect hash functions
// and an associated string table to check
// whether the received length N entity candidate
// string is a valid entity. The hash values
// are computed as a sum of associated value 
// terms indexed by offset versions of 
// selected input entity characters.
// The tables are sparse given that gperf 
// generates a perfect, non-minimal hash.
// Each string table entry contains a reference 
// string for validation and a utf-8 encoded
// output. Online computation could be used
// to eliminate the pre-computed utf-8 entries
// and length fields, which would compute for 
// memory. All hash tables and hash constants 
// are contained in the auto-generated
// file entityTabs.dat.  
// The hash generator allows arbitrarily 
// partitioned entity tables; the choice
// of partition used here was found to provide
// a reasonable performance vs. memory tradeoff
//
// mapN() are called via jump table to minimize
// comparisons and branching, especially for 
// dense escape strings
//************************************************

// map2()
// utf-8 encode length 2 entity using hash LUT and write utf-8 result to (dst)
// flush and return to '&' scan on either invalid hash or unknown string
// same process is used below for "short" entities via mapN, N=2,3,4,5,..,8
// the only differences are in the number and indices of associated value terms used to compute the hash
// for map2() there are two AV terms, one in computeAV() and one in hashToUtf8()
TEXT map2(SB), NOSPLIT, $0-0
  hashInit                (assocVal_2_to_3,stringTable_2_to_3,satMin_2_to_3,satMax_2_to_3,strMask_2_to_3)
  computeAV               (hash,assocVal,nextChar,MOVQ)                             // hash input string 
  hashToUtf8              (hash,assocVal,nextChar,3,bpe2_to_3,maxH2_to_3,strMask)   // look up, compare, map to utf-8
unknownString:                                                                      // for unknown string
  MOVB                    $'&', (dst)                                               // caused by invalid hash or valid unknown
  MOVW                    inputStr, (1)(dst)                                        // flush unmapped characters to (dst)
  ADDQ                    $3, dst                                                   // and scan for next '&'
  JMP                     findAmpAndCopy(SB)

// map3()
// utf-8 encode length 3 entity using hash LUT and write utf-8 result to (dst)
// shares associated value and string tables with map2()
// map3() has 3 AV terms in the hash summation
TEXT map3(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_2_to_3,stringTable_2_to_3,satMin_2_to_3,satMax_2_to_3,strMask_2_to_3)
  computeAV               (hash,assocVal,nextChar,MOVQ)                             // first AV term uses MOVQ to hash
  computeAV               (hash,assocVal,nextChar,ADDQ)                             // subsequent AV terms are added
  hashToUtf8              (hash,assocVal,nextChar,3,bpe2_to_3,maxH2_to_3,strMask)   // this patter is repeated for all mapN
  unknownStringShort      (1,0xffff,map2)                                           // test for unterminated len 2

// map4()
// utf-8 encode length 4 entity using hash LUT and write utf-8 result to (dst)
TEXT map4(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_4,stringTable_4,satMin_4,satMax_4,strMask_4)    // hash=sum(av[str[0..3]])
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  hashToUtf8              (hash,assocVal,nextChar,4,bpe4,maxH4,strMask)   
  unknownStringShort      (1,0xffffff,map3)

// map5()
// utf-8 encode length 5 entity using hash LUT and write utf-8 result to (dst)
TEXT map5(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_5,stringTable_5,satMin_5,satMax_5,strMask_5)    // hash=sum(av[str[0..4]])
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  hashToUtf8              (hash,assocVal,nextChar,5,bpe5,maxH5,strMask)   
  unknownStringShort      (1,0xffffffff,map4)

// map6()
// utf-8 encode length 6 entity using hash LUT and write utf-8 result to (dst)
TEXT map6(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_6,stringTable_6,satMin_6,satMax_6,strMask_6)    // hash=sum(av[str[0..5]])
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  hashToUtf8              (hash,assocVal,nextChar,6,bpe6,maxH6,strMask)   
  unknownStringShort      (1,0xffffffffff,map5)

// map7()
// utf-8 encode length 7 entity using hash LUT and write utf-8 result to (dst)
TEXT map7(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_7,stringTable_7,satMin_7,satMax_7,strMask_7)    // hash=sum(av[str[0..6]])
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  hashToUtf8              (hash,assocVal,nextChar,7,bpe7,maxH7,strMask)   
  unknownStringShort      (1,0xffffffffffff,map6)

// map8()
// utf-8 encode length 8 entity using hash LUT and write utf-8 result to (dst)
TEXT map8(SB), NOSPLIT, $0-0 
  hashInit                (assocVal_8,stringTable_8,satMin_8,satMax_8,strMask_8)    // hash=sum(av[str[0..3,5]])
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  SHRQ                    $8, nextChar                                              // skip str[4] in av sum
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  hashToUtf8              (hash,assocVal,nextChar,8,bpe8,maxH8,strMask)   
  unknownStringShort      (2,0xffffffffffff,map6)                                   // unknown entities longer than 6 backtrack to 6,5,4,3,...

// unknownStringLong()
// mapN unknown string exit for 8 < N <= 32
// backtracks to N=6, calls map6 to test for len 6 unterminated entity
// unknownStringLong is called for an unknown &<entity>, which is reflected in one of 
// two conditions, either: a) invalid hash b) valid hash but unknown string;
// if map6 returns unknownString then descending maps of lengths N-1, N-2, N-3, ..., 2 
// are tried for a match to cover unterminated entities of len <= 6
// before flushing unrecognized characters to output string
#define unknownStringLong(srcOffset) \
unknownString:                                                \
  SUBQ                    $srcOffset, src                     \
  ADDQ                    $srcOffset, srcLen                  \ 
  VMOVQ                   entity16, inputStr                  \
  MOVQ                    $0xffffffffffff, strMask            \
  ANDQ                    strMask, inputStr                   \
  JMP                     map6(SB) 

// map9()
// utf-8 encode length 9 entity using hash LUT and write utf-8 result to (dst)
TEXT map9(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_9,stringTable_9,satMin_9,satMax_9,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0..4,7]])
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  SHRQ                    $16, nextChar                                           // skip str[5] and str[6]
  hashToUtf8Long          (hash,assocVal,nextChar,9,bpe9,maxH9,0x1ff)   
  unknownStringLong       (3)

// map10()
// utf-8 encode length 10 entity using hash LUT and write utf-8 result to (dst)
TEXT map10(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_10,stringTable_10,satMin_10,satMax_10,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0..4,6]])
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  SHRQ                    $8, nextChar                                            // skip str[5]
  hashToUtf8Long          (hash,assocVal,nextChar,10,bpe10,maxH10,0x3ff)   
  unknownStringLong       (4)

// map11()
// utf-8 encode length 11 entity using hash LUT and write utf-8 result to (dst)
TEXT map11(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_11,stringTable_11,satMin_11,satMax_11,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0..6]])
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)                   
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  computeAV               (hash,assocVal,nextChar,ADDQ)    
  hashToUtf8Long          (hash,assocVal,nextChar,11,bpe11,maxH11,0x7ff)   
  unknownStringLong       (5)

// map12()
// utf-8 encode length 12 entity using hash LUT and write utf-8 result to (dst)
TEXT map12(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_12,stringTable_12,satMin_12,satMax_12,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0,1,6,8]])
  computeAV               (hash,assocVal,nextChar,ADDQ)
  SHRQ                    $32, nextChar                                           // skip str[2..5]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  VPEXTRQ                 $1, str, nextChar                                       // get str[8] from entity32 buffer
  hashToUtf8Long          (hash,assocVal,nextChar,12,bpe12,maxH12,0xfff)   
  unknownStringLong       (6)

// map13()
// utf-8 encode length 13 entity using hash LUT and write utf-8 result to (dst)
TEXT map13(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_13,stringTable_13,satMin_13,satMax_13,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0,5,9]])
  SHRQ                    $32, nextChar                                           // skip str[1..4]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  VPEXTRQ                 $1, str, nextChar                                       // get str[9] from entity32 buffer
  SHRQ                    $8, nextChar
  hashToUtf8Long          (hash,assocVal,nextChar,13,bpe13,maxH13,0x1fff)   
  unknownStringLong       (7)

// map14()
// utf-8 encode length 14 entity using hash LUT and write utf-8 result to (dst)
TEXT map14(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_14,stringTable_14,satMin_14,satMax_14,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0,1,5,8]])
  computeAV               (hash,assocVal,nextChar,ADDQ)
  SHRQ                    $24, nextChar                                           // skip str[2..4]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  VPEXTRQ                 $1, str, nextChar                                       // get str[8] from entity32 buffer
  hashToUtf8Long          (hash,assocVal,nextChar,14,bpe14,maxH14,0x3fff)   
  unknownStringLong       (8)

// map15()
// utf-8 encode length 15 entity using hash LUT and write utf-8 result to (dst)
TEXT map15(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_15,stringTable_15,satMin_15,satMax_15,0x2)
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[0,2,5,9]])    
  SHRQ                    $8, nextChar                                            // skip str[1]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  SHRQ                    $16, nextChar                                           // skip str[3,4]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  VPEXTRQ                 $1, str, nextChar                                       // get str[9] from entity32 buffer
  SHRQ                    $8, nextChar
  hashToUtf8Long          (hash,assocVal,nextChar,15,bpe15,maxH15,0x7fff)   
  unknownStringLong       (9)

// map16()
// utf-8 encode length 16 entity using hash LUT and write utf-8 result to (dst)
TEXT map16(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_16,stringTable_16,satMin_16,satMax_16,0x2)
  SHRQ                    $8, nextChar                                            // skip str[0]
  computeAV               (hash,assocVal,nextChar,MOVQ)                           // hash=sum(av[str[1,7,10]])    
  SHRQ                    $40, nextChar                                           // skip str[2..6]
  computeAV               (hash,assocVal,nextChar,ADDQ)
  VPEXTRQ                 $1, str, nextChar                                       // get str[10] from entity32 buffer
  SHRQ                    $16, nextChar
  hashToUtf8Long          (hash,assocVal,nextChar,16,bpe16,maxH16,0xffff)   
  unknownStringLong       (10)

// map17()
// utf-8 encode length 17 entity using hash LUT and write utf-8 result to (dst)
// map17 shares tables with map18..32, but uses truncated av sum
TEXT map17(SB), NOSPLIT, $0-0 
  hashInitLong            (assocVal_17_to_32,stringTable_17_to_32,satMin_17_to_32,satMax_17_to_32,0x4)
  computeAV               (hash,assocVal,nextChar,MOVQ)                   
  VPEXTRQ                 $1, str, nextChar                                       // hash=sum(av[str[0,8,13]])
  computeAV               (hash,assocVal,nextChar,ADDQ)
  SHRQ                    $32, nextChar
  hashToUtf8Long          (hash,assocVal,nextChar,32,bpe17_to_32,maxH17_to_32,0x1ffff)   
  unknownStringLong       (11)

// mapParametric()
// map18..map32 use a common table and the same associated value summation
// expressed in mapParametric. Parameters wrt N are as follows: 
//  msbSelect : indexes the high order quadword in the entity32 buffer
//  strMask   : mask for strcmp(hashStr,inputStr) 
//  offset    : backtrack parameter for unknown strings (#bytes to rewind input for map6..2 unterminated entity search)
// hash function is identical to map17 except for the addition of on associated value term on str[17], i.e.,
// hash = sum(av[str[0,8,13,17]])
#define mapParametric(msbSelect,strMask,offset) \
  hashInitLong            (assocVal_17_to_32,stringTable_17_to_32,satMin_17_to_32,satMax_17_to_32,msbSelect)  \
  computeAV               (hash,assocVal,nextChar,MOVQ)                                                       \
  VPEXTRQ                 $1, str, nextChar                                                                   \ 
  computeAV               (hash,assocVal,nextChar,ADDQ)                                                       \
  SHRQ                    $32, nextChar                                                                       \
  computeAV               (hash,assocVal,nextChar,ADDQ)                                                       \
  VPERMQ                  $0x2, str32, str32                                                                  \
  VMOVQ                   str, nextChar                                                                       \
  SHRQ                    $8, nextChar                                                                        \
  hashToUtf8Long          (hash,assocVal,nextChar,32,bpe17_to_32,maxH17_to_32,strMask)                        \ 
  unknownStringLong       (offset)

// map18(), map19(), ..., map32()
// utf-8 encode length N entity using hash LUT and write utf-8 result to (dst), N=18,19,..,32
TEXT map18(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0x3ffff,12)

TEXT map19(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0x7ffff,13)

TEXT map20(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0xfffff,14)

TEXT map21(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0x1fffff,15)

TEXT map22(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0x3fffff,16)

TEXT map23(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0x7fffff,17)

TEXT map24(SB), NOSPLIT, $0-0 
  mapParametric           (0x4,0xffffff,18)

TEXT map25(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x1ffffff,19)

TEXT map26(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x3ffffff,20)

TEXT map27(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x7ffffff,21)

TEXT map28(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0xfffffff,22)

TEXT map29(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x1fffffff,23)

TEXT map30(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x3fffffff,24)

TEXT map31(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0x7fffffff,25)

TEXT map32(SB), NOSPLIT, $0-0 
  mapParametric           (0x8,0xffffffff,26)

// mapN() jump table as a function of entity length, in bytes
DATA map<>+0x10(SB)/8, $map2(SB)
DATA map<>+0x18(SB)/8, $map3(SB)
DATA map<>+0x20(SB)/8, $map4(SB)
DATA map<>+0x28(SB)/8, $map5(SB)
DATA map<>+0x30(SB)/8, $map6(SB)
DATA map<>+0x38(SB)/8, $map7(SB)
DATA map<>+0x40(SB)/8, $map8(SB)
DATA map<>+0x48(SB)/8, $map9(SB)
DATA map<>+0x50(SB)/8, $map10(SB)
DATA map<>+0x58(SB)/8, $map11(SB)
DATA map<>+0x60(SB)/8, $map12(SB)
DATA map<>+0x68(SB)/8, $map13(SB)
DATA map<>+0x70(SB)/8, $map14(SB)
DATA map<>+0x78(SB)/8, $map15(SB)
DATA map<>+0x80(SB)/8, $map16(SB)
DATA map<>+0x88(SB)/8, $map17(SB)
DATA map<>+0x90(SB)/8, $map18(SB)
DATA map<>+0x98(SB)/8, $map19(SB)
DATA map<>+0xa0(SB)/8, $map20(SB)
DATA map<>+0xa8(SB)/8, $map21(SB)
DATA map<>+0xb0(SB)/8, $map22(SB)
DATA map<>+0xb8(SB)/8, $map23(SB)
DATA map<>+0xc0(SB)/8, $map24(SB)
DATA map<>+0xc8(SB)/8, $map25(SB)
DATA map<>+0xd0(SB)/8, $map26(SB)
DATA map<>+0xd8(SB)/8, $map27(SB)
DATA map<>+0xe0(SB)/8, $map28(SB)
DATA map<>+0xe8(SB)/8, $map29(SB)
DATA map<>+0xf0(SB)/8, $map30(SB)
DATA map<>+0xf8(SB)/8, $map31(SB)
DATA map<>+0x100(SB)/8, $map32(SB)
GLOBL map<>(SB), (RODATA|NOPTR), $264

// delimiter LUT for entity processing loop
// LUT(ascii[A-F],ascii[a-f],ascii[0-9]) = 1
// LUT(all other bytes) = 0
// used for conditional exit from the entity loop
DATA delim<>+0x30(SB)/8, $0x0101010101010101
DATA delim<>+0x38(SB)/8, $0x0000000000000101
DATA delim<>+0x40(SB)/8, $0x0101010101010100
DATA delim<>+0x48(SB)/8, $0x0101010101010101
DATA delim<>+0x50(SB)/8, $0x0101010101010101
DATA delim<>+0x58(SB)/8, $0x0000000000010101
DATA delim<>+0x60(SB)/8, $0x0101010101010100
DATA delim<>+0x68(SB)/8, $0x0101010101010101
DATA delim<>+0x70(SB)/8, $0x0101010101010101
DATA delim<>+0x78(SB)/8, $0x0000000000010101
GLOBL delim<>(SB), (RODATA|NOPTR), $256

// definitions to build hex and decimal 
// digit LUTs for the decimal and hex numeric entity unescape loops
#define non_digit 0xffffffffffffffff
#define n non_digit
#define digits_07 0x0706050403020100
#define digits_89 0xffffffffffff0908
#define digits_af 0xff0f0e0d0c0b0a00
#define makeTableEntry(name,addr,val) \
  DATA name<>+addr(SB)/8, $val
#define finalTableEntry(name,len) \
  GLOBL name<>(SB), (RODATA|NOPTR), $len
#define h2d(addr,val) \
  makeTableEntry(hex2dec,addr,val)
#define h2dFinal \
  finalTableEntry(hex2dec,256)
#define d2d(addr,val) \
  makeTableEntry(dec2dec,addr,val)
#define d2dFinal \
  finalTableEntry(dec2dec,256)

// hex entity to decimal value single digit conversion LUT
// used by the numeric entity loop to unescape '&#x' or '&#X' hex numeric entities
// input = ascii [0-9], [a-f], [A-F]
// output = decimal 0-9, 10-15, 10-15
h2d(0x00,n)           h2d(0x08,n) 
h2d(0x10,n)           h2d(0x18,n) 
h2d(0x20,n)           h2d(0x28,n)
h2d(0x30,digits_07)   h2d(0x38,digits_89)
h2d(0x40,digits_af)   h2d(0x48,n)
h2d(0x50,n)           h2d(0x58,n) 
h2d(0x60,digits_af)   h2d(0x68,n)
h2d(0x70,n)           h2d(0x78,n) 
h2d(0x80,n)           h2d(0x88,n) 
h2d(0x90,n)           h2d(0x98,n) 
h2d(0xa0,n)           h2d(0xa8,n) 
h2d(0xb0,n)           h2d(0xb8,n) 
h2d(0xc0,n)           h2d(0xc8,n) 
h2d(0xd0,n)           h2d(0xd8,n) 
h2d(0xe0,n)           h2d(0xe8,n) 
h2d(0xf0,n)           h2d(0xf8,n) 
h2dFinal

// decimal entity to decimal value single digit conversion LUT
// used by the numeric entity loop to unescape '&#' decimal numeric entities
// input = ascii [0-9]
// output = decimal 0-9
d2d(0x00,n)           d2d(0x08,n) 
d2d(0x10,n)           d2d(0x18,n) 
d2d(0x20,n)           d2d(0x28,n)
d2d(0x30,digits_07)   d2d(0x38,digits_89)
d2d(0x40,n)           d2d(0x48,n)
d2d(0x50,n)           d2d(0x58,n) 
d2d(0x60,n)           d2d(0x68,n)
d2d(0x70,n)           d2d(0x78,n) 
d2d(0x80,n)           d2d(0x88,n) 
d2d(0x90,n)           d2d(0x98,n) 
d2d(0xa0,n)           d2d(0xa8,n) 
d2d(0xb0,n)           d2d(0xb8,n) 
d2d(0xc0,n)           d2d(0xc8,n) 
d2d(0xd0,n)           d2d(0xd8,n) 
d2d(0xe0,n)           d2d(0xe8,n) 
d2d(0xf0,n)           d2d(0xf8,n) 
d2dFinal

//***********************************************
// unescapeScanAVX512()
//
// html.unescapeScanAVX512() external entry point
// locate next '&' occurrence in a string
// no memcpy() for fast unescapeNone()
//
//***********************************************
#undef srcLen
#undef x 
#undef inv 
#undef amp64                                            // esc16/32/64: 16/32/64-lane esc constants for avx byte compare
#undef amp32           
#undef amp16      
#undef findAmp
#undef vFindAmp 
#define srcLen          BX                              // input string length
#define x               CX                              // block of input characters, either 2, 4, or 8
#define inv             DX                              // complemented block for & detection step
#define amp64           Z0                              // esc16/32/64: 16/32/64-lane esc constants for byte compare
#define amp32           Y0
#define amp16           X0

// checks for scalar found '&' in ampMask and returns index in idx; 
// otherwise falls thru
#define findAmp(ampMask,maskLen) \
  TZCNTQ                  ampMask, idx                  \ 
  SHRQ                    $3, idx                       \
  CMPQ                    ampMask, $0                   \
  JNE                     return                        \
  SUBQ                    $maskLen, srcLen              \
  ADDQ                    $maskLen, src

// checks for one or more '&' in (src), mask in e0,
// then returns index in idx, othereise falls thru
#define vFindAmp(ampMask,numBytes) \
  VPCMPEQB                (src), ampMask, noMask, e0    \
  SUBQ                    $numBytes, srcLen             \
  KMOVQ                   e0, idx                       \
  TZCNTQ                  idx, idx                      \
  CMPQ                    idx, $64                      \          
  JNE                     return                        \
  ADDQ                    $numBytes, src

TEXT ·unescapeScanAVX512(SB),0,$0-24
  MOVQ						        s_base+0(FP), src
  MOVQ                    s_len+8(FP), srcLen
  KXNORQ 					        noMask, noMask, noMask                // all 1 / no masked bits for any avx512 op
  VMOVDQU64               amp<>(SB), amp64                      // load '&' (escape delimiter) into 64 x 1 byte lanes
  XORQ                    result, result                        // clear result accumulator
  CMPQ                    srcLen, $256                          // process in decending 2^N-sized blocks, N=8,7,6,...,0
  JL                      scan128

// scan for esc flag via decreasing dyadic-length
// blocks: 256*N, 128, ..., 1, where N = len(s) div 256
scan256:                                                         
  VPCMPEQB                (src), amp64, noMask, e0              // check next block for esc
  VPCMPEQB                (64)(src), amp64, noMask, e1          // on 256 bytes = 4x64 
  VPCMPEQB                (128)(src), amp64, noMask, e2           
  VPCMPEQB                (192)(src), amp64, noMask, e3           
  SUBQ                    $256, srcLen
  KORQ                    e0, e1, ex                            // reduce 4x64 to 1x64, any byte>0 
  KORQ                    e2, e3, ey                            // detects potential esc 
  KORQ                    ex, ey, ex
  KMOVQ                   ex, idx
  CMPQ                    idx, $0                               // if any byte==0x26 idx=1
  JNE                     return256                             // get index and return
  ADDQ                    $256, src
  CMPQ                    srcLen, $256                          // loop for >= 256 bytes remaining
  JGE                     scan256
// scan beyond after 256*N using same process as for 256 
scan128:
  CMPQ                    srcLen, $128                          
  JL                      scan64                                                         
  VPCMPEQB                (src), amp64, noMask, e2              
  VPCMPEQB                (64)(src), amp64, noMask, e3
  SUBQ                    $128, srcLen
  KORQ                    e2, e3, e1                            
  KMOVQ                   e1, idx
  CMPQ                    idx, $0                            
  JNE                     return128
  ADDQ                    $128, src
scan64:
  CMPQ                    srcLen, $64                           
  JL                      scan32                                                         
  vFindAmp                (amp64,64)
scan32:
  CMPQ                    srcLen, $32                           
  JL                      scan16
  vFindAmp                (amp32,32)
scan16:
  CMPQ                    srcLen, $16                           
  JL                      scan8                                                         
  vFindAmp                (amp16,16)
scan8:                                                          // blocks of size 8,4,2 use
  CMPQ                    srcLen, $8                            // (x-0x01...) & ~x & 0x80... to detect esc
  JL                      scan4                                 // instead of 8/4/2x shift/cmp/branch
  MOVQ                    (src), x                              // to minimize branching
  MOVQ                    $0x2626262626262626, c
  XORQ                    c, x
  MOVQ                    $0x0101010101010101, c
  MOVQ                    x, inv
  SUBQ                    c, x
  NOTQ                    inv
  ANDQ                    inv, x
  MOVQ                    $0x8080808080808080, c
  ANDQ                    c, x
  findAmp                 (x,8) 
scan4:                                                          
  CMPQ                    srcLen, $4
  JL                      scan2
  MOVLQZX                 (src), x
  XORL                    $0x26262626, x
  MOVQ                    x, inv
  SUBL                    $0x01010101, x
  NOTL                    inv
  ANDL                    inv, x
  ANDL                    $0x80808080, x 
  findAmp                 (x,4) 
scan2:                                                          
  CMPQ                    srcLen, $2
  JL                      scan1
  MOVWQZX                 (src), x
  XORW                    $0x2626, x
  MOVQ                    x, inv
  SUBW                    $0x0101, x
  NOTW                    inv
  ANDW                    inv, x
  ANDW                    $0x8080, x
  findAmp                 (x,2) 
scan1:                                                          
  CMPQ                    srcLen, $1
  JL                      noEsc
  SUBQ                    $1, srcLen
  CMPB                    (src), $'&'
  JNE                     noEsc
  MOVQ                    s_base+0(FP), srcLen
  SUBQ                    srcLen, src
  ADDQ                    src, result
  MOVQ                    result, ret+16(FP)
  RET

// convert block-local index to global index, and then return 
return256:
  KMOVQ                   e0, idx
  TZCNTQ                  idx, idx
  CMPQ                    idx, $64
  JNE                     return
  KMOVQ                   e1, idx
  TZCNTQ                  idx, idx
  ADDQ                    $64, result
  CMPQ                    idx, $64
  JNE                     return
  ADDQ                    $64, result
return128:
  KMOVQ                   e2, idx
  TZCNTQ                  idx, idx
  CMPQ                    idx, $64
  JNE                     return
  KMOVQ                   e3, idx
  TZCNTQ                  idx, idx
  ADDQ                    $64, result
return:
  ADDQ                    idx, result
  MOVQ                    s_base+0(FP), srcLen
  SUBQ                    srcLen, src
  ADDQ                    src, result
  MOVQ                    result, ret+16(FP)
  RET
noEsc:
  MOVQ                    $-1, result
  MOVQ                    result, ret+16(FP)
  RET
