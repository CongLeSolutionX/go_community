// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT ·startTimer(SB),NOSPLIT,$0
	B time·startTimer(SB)

TEXT ·stopTimer(SB),NOSPLIT,$0
	B time·stopTimer(SB)

TEXT ·resetTimer(SB),NOSPLIT,$0
	B time·resetTimer(SB)
