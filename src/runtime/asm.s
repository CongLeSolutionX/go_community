// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// Funcdata for functions with no local variables in frame.
// Define two zero-length bitmaps, because the same Index is used
// for the local variables as for the argument frame, and assembly
// frames have two argument bitmaps, one without results and one with results.
DATA runtime·no_pointers_stackmap+0x00(SB)/4, $2
DATA runtime·no_pointers_stackmap+0x04(SB)/4, $0
GLOBL runtime·no_pointers_stackmap(SB),RODATA, $8

TEXT runtime∕internal∕base·Nop(SB),NOSPLIT,$0-0
	RET

GLOBL runtime∕internal∕base·Mheap_(SB), NOPTR, $0
GLOBL runtime∕internal∕base·Memstats(SB), NOPTR, $0
