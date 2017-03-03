// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_amd64_openbsd(SB),NOSPLIT,$-8
	JMP	_rt0_amd64(SB)

TEXT _rt0_amd64_openbsd_lib(SB),NOSPLIT,$0
	// OpenBSD does not pass argc/argv to DT_INIT_ARRAY functions.
	MOVQ	$0, DI
	MOVQ	DI, SI
	JMP	_rt0_amd64_lib(SB)
