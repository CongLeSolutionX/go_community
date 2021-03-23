// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build amd64

#include "textflag.h"

TEXT	·funcPCTestFn(SB),NOSPLIT,$0-0
	RET

GLOBL	·funcPCTestFnAddr(SB), NOPTR, $8
DATA	·funcPCTestFnAddr(SB)/8, $·funcPCTestFn(SB)
