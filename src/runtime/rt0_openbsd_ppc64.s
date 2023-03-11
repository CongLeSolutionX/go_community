// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

DATA	runtime·toc(SB)/8, $0
GLOBL	runtime·toc(SB), NOPTR, $8

TEXT _rt0_ppc64_openbsd(SB),NOSPLIT,$0
	// Load the .TOC. pointer into R2.
	MOVD	$TOC(SB), R2
	BR	main(SB)

TEXT main(SB),NOSPLIT,$-8
	// Store TOC pointer - this will have been loaded into R2 either
	// by _rt0_ppc64_openbsd or by __start, when externally linked.
	MOVD	$runtime·toc(SB), R12
	MOVD	R2, (R12)

	// Make sure R0 is zero before _main
	XOR	R0, R0

	MOVD	$runtime·rt0_go(SB), R12
	MOVD	R12, CTR
	BR	(CTR)
