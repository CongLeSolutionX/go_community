// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm64

/*

Go Assembly for ARM64 Reference Manual

1. Alphabetical list of basic instructions
    // TODO

2. Alphabetical list of float-point instructions
    // TODO

3. Alphabetical list of SIMD instructions

    VADD: Add (vector).
      VADD	<Vm>.T, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        8B, 16B, H4, H8, S2, S4, D2

    VADDP: Add Pairwise (vector)
      VADDP	<Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D2

    VADDV: Add across Vector.
      VADDV	<Vn>.<T>, Vd
        <T> Is an arrangement specifier and can have the following values:
        8B, 16B, H4, H8, S4

    VAND: Bitwise AND (vector)
      VAND	<Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16

    VCMEQ: Compare bitwise Equal (vector)
      VCMEQ	<Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D2

    VDUP: Duplicate vector element to vector or scalar.
      VDUP	<Vn>.<Ts>[index], <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        8B, 16B, H4, H8, S2, S4, D2
        <Ts> Is an element size specifier and can have the following values:
        B, H, S, D

    VEOR: Bitwise exclusive OR (vector, register)
      VEOR	<Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16

    VEXT:  Extracts vector elements from src SIMD registers to dst SIMD register
      VEXT	$index, <Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangment specifier and can be B8, B16
        $index is the lowest numberred byte element to be exracted.

    VLD1: Load multiple single-element structures
      VLD1	(Rn), [<Vt>.<T>, <Vt2>.<T> ...]     // no offset
      VLD1.P	imm(Rn), [<Vt>.<T>, <Vt2>.<T> ...]  // immediate offset variant
      VLD1.P	(Rn)(Rm), [<Vt>.<T>, <Vt2>.<T> ...] // register offset variant
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D1, D2

    VLD1: Load one single-element structures
      VLD1	(Rn), <Vt>.<T>[index]     // no offset
      VLD1.P	imm(Rn), <Vt>.<T>[index]  // immediate offset variant
      VLD1.P	(Rn)(Rm), <Vt>.<T>[index] // register offset variant
        <T> Is an arrangement specifier and can have the following values:
        B, H, S D

    VMOV: move
      VMOV	<Vn>.<T>[index], Rd // Move vector element to general-purpose register.
        <T> Is a source width specifier and can have the following values:
        B, H, S (Wd)
        D (Xd)

      VMOV	Rn, <Vd>.<T> // Duplicate general-purpose register to vector.
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4 (Wn)
        D2 (Xn)

      VMOV	<Vn>.<T>, <Vd>.<T> // Move vector.
        <T> Is an arrangement specifier and can have the following values:
        B8, B16

      VMOV	Rn, <Vd>.<T>[index] // Move general-purpose register to a vector element.
        <T> Is a source width specifier and can have the following values:
        B, H, S (Wd)
        D (Xd)

      VMOV	<Vn>.<T>[index], Vn  // Move vector element to scalar.
        <T> Is an element size specifier and can have the following values:
        B, H, S, D

      VMOV	<Vn>.<T>[index2], <Vd>.<T>[index1]  // Move vector element to another.
        <T> Is an element size specifier and can have the following values:
        B, H, S, D

    VMOVI: Move Immediate (vector).
      VMOVI	$imm8, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        8B, 16B

    VMOVS: Load SIMD&FP Register (immediate offset). ARMv8: LDR (immediate, SIMD&FP)
      Store SIMD&FP register (immediate offset). ARMv8: STR (immediate, SIMD&FP)
      VMOVS	(Rn), Vn
      VMOVS.W	imm(Rn), Vn
      VMOVS.P	imm(Rn), Vn
      VMOVS	Vn, (Rn)
      VMOVS.W	Vn, imm(Rn)
      VMOVS.P	Vn, imm(Rn)

    VORR: Bitwise inclusive OR (vector, register)
      VORR	<Vm>.<T>, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16

    VRBIT: Reverse bit order (vector)
      VRBIT	<Vn>.<T>, <Vd>.<T>
        <T> Is an arrangment specifier and can be B8, B16

    VREV32: Reverse elements in 32-bit words (vector).
      REV32 <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8

    VSHL: Shift Left(immediate)
      VSHL 	$shift, <Vn>.<T>, <Vd>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D1, D2
        $shift Is the left shift amount

    VST1: Store multiple single-element structures
      VST1	[<Vt>.<T>, <Vt2>.<T> ...], (Rn)         // no offset
      VST1.P	[<Vt>.<T>, <Vt2>.<T> ...], imm(Rn)      // immediate offset variant
      VST1.P	[<Vt>.<T>, <Vt2>.<T> ...], (Rn)(Rm)     // register offset variant
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D1, D2

    VST1: Store one single-element structures
      VST1	<Vt>.<T>.<Index>, (Rn)         // no offset
      VST1.P	<Vt>.<T>.<Index>, imm(Rn)      // immediate offset variant
      VST1.P	<Vt>.<T>.<Index>, (Rn)(Rm)     // register offset variant
        <T> Is an arrangement specifier and can have the following values:
        B, H, S, D

    VUSHR: Unsigned shift right(immediate)
      VUSHR	$shift, <Vn>.<T>, <Vm>.<T>
        <T> Is an arrangement specifier and can have the following values:
        B8, B16, H4, H8, S2, S4, D1, D2
        $shift is the right shift amount


4. Alphabetical list of cryptographic extension instructions

    PMULL{2}:  Polynomial multiply long.
      PMULL{2}	<Vn>.<Tb>, <Vm>.<Tb>, <Vd>.<Ta> // PMULL multiplies corresponding elements in the lower half of the vectors of two source SIMD registers and PMULL{2} operates in the upper half
        <Ta> Is an arrangement specifier, it can be H8, 1Q
	<Tb> Is an arrangement specifier, it can be B8, B16, D1, D2

    SHA1C, SHA1M, SHA1P: SHA1 hash update.
      SHA1C	<Vm>.S4, Vn, Vd
      SHA1M	<Vm>.S4, Vn, Vd
      SHA1P	<Vm>.S4, Vn, Vd

    SHA1H: SHA1 fixed rotate.
      SHA1H	Vn, Vd

    SHA1SU0:   SHA1 schedule update 0.
    SHA256SU1: SHA256 schedule update 1.
      SHA1SU0	<Vm>.S4, <Vn>.S4, <Vd>.S4
      SHA256SU1	<Vm>.S4, <Vn>.S4, <Vd>.S4

    SHA1SU1:   SHA1 schedule update 1.
    SHA256SU0: SHA256 schedule update 0.
      SHA1SU1	<Vn>.S4, <Vd>.S4
      SHA256SU0	<Vn>.S4, <Vd>.S4

    SHA256H, SHA256H2: SHA256 hash update.
      SHA256H	<Vm>.S4, Vn, Vd
      SHA256H2	<Vm>.S4, Vn, Vd

*/
