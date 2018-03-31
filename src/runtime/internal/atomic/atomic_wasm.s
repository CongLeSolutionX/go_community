// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT runtime∕internal∕atomic·StorepNoWB(SB), NOSPLIT, $0-16
	I64Load ptr+0(FP)
	I32WrapI64
	I64Load val+8(FP)
	I64Store $0
	RET
