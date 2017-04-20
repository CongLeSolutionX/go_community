// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

// func Sincos(d float64) (sin, cos float64)
TEXT ·Sincos(SB),NOSPLIT,$0
	JMP	·sincos(SB)
