// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#define _dmb_armv6_preamble \
	MOVB	runtime·goarm(SB), R11; \
	CMP	$7, R11; \
	BEQ	5(PC); \
	CMP	$6, R11; \
	BLT	4(PC); \
	/* ARMv6 dmb using c7, Cache Operations Register */ \
	MCR	15, 0, R0, C7, C10, 5; \
	B	2(PC)

#define DMB_ISHST \
	_dmb_armv6_preamble; \
	/* ARMv7 dmb ishst */ \
	WORD	$0xf57ff05a

#define DMB_ISH \
	_dmb_armv6_preamble; \
	/* ARMv7 dmb ish */ \
	WORD	$0xf57ff05b

#define DMB_ST \
	_dmb_armv6_preamble; \
	/* ARMv7 dmb st */ \
	WORD	$0xf57ff05e
