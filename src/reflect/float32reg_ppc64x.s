// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ppc64 || ppc64le
// +build ppc64 ppc64le

#include "textflag.h"

// Convert float32->uint64
TEXT ·archFloat32ToReg(SB),NOSPLIT,$0-16
	FMOVS	val+0(FP), F1
	FMOVD	F1, ret+8(FP)
	RET

// Convert uint64->float32
TEXT ·archFloat32FromReg(SB),NOSPLIT,$0-12
	FMOVD	reg+0(FP), F1
	FMOVS	F1, ret+8(FP)
	RET

