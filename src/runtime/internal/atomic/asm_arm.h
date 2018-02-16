// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#define _armv7_test \
	MOVB	runtime·goarm(SB), R11; \
	CMP	$7, R11; \
	BLT	2(PC)

#define DMB_ISH \
	_armv7_test; \
	DMB MB_ISH

#define DMB_ISHST \
	_armv7_test; \
	DMB MB_ISHST

#define DMB_ST \
	_armv7_test; \
	DMB MB_ST
