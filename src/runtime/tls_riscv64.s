// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "go_tls.h"
#include "funcdata.h"
#include "textflag.h"

// If !iscgo, this is a no-op.
//
// NOTE: mcall() assumes this clobbers only X31 (REG_TMP).
TEXT ·save_g(SB),NOSPLIT|NOFRAME,$0-0
#ifndef GOOS_openbsd
	MOVB	·iscgo(SB), X31
	BEQZ	X31, nocgo
#endif
	MOV	g, ·tls_g(SB)
nocgo:
	RET

TEXT ·load_g(SB),NOSPLIT|NOFRAME,$0-0
	MOV	·tls_g(SB), g
	RET

GLOBL ·tls_g(SB), TLSBSS, $8
