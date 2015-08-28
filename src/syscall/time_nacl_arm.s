// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT runtime·startTimer(SB),NOSPLIT,$0
	B runtime·startTimer(SB)

TEXT runtime·stopTimer(SB),NOSPLIT,$0
	B runtime·stopTimer(SB)
