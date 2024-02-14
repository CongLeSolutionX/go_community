// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// If !iscgo, this is a no-op.
//
// NOTE: mcall() assumes this clobbers only R30 (REGTMP).
TEXT ·save_g(SB),NOSPLIT|NOFRAME,$0-0
	MOVB	·iscgo(SB), R30
	BEQ	R30, nocgo

	MOVV	g, ·tls_g(SB)

nocgo:
	RET

TEXT ·load_g(SB),NOSPLIT|NOFRAME,$0-0
	MOVV	·tls_g(SB), g
	RET

GLOBL ·tls_g(SB), TLSBSS, $8
