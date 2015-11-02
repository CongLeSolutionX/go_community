// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT runtime·casp1(SB),NOSPLIT,$0
	B	runtime·cas(SB)

// TODO(minux): this is only valid for ARMv6+
// bool armcas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	}else
//		return 0;
TEXT runtime·cas(SB),NOSPLIT,$0
	B runtime·armcas(SB)
