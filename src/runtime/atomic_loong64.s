// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// DBAR sync load/store operation
#define DBAR	WORD $0x38720000

TEXT ·publicationBarrier(SB),NOSPLIT|NOFRAME,$0-0
	DBAR
	RET
