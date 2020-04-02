// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·x509_CFDataGetLength_trampoline(SB),NOSPLIT,$0-0
	JMP	x509_CFDataGetLength(SB)
TEXT ·x509_CFDataCreate_trampoline(SB),NOSPLIT,$0-0
	JMP	x509_CFDataCreate(SB)
TEXT ·x509_CFRelease_trampoline(SB),NOSPLIT,$0-0
	JMP	x509_CFRelease(SB)
