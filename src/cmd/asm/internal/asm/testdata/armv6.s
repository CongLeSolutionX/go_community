// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "../../../../../runtime/textflag.h"

TEXT	foo(SB), DUPOK|NOSPLIT, $0

	ADDF	F0, F1, F2    // 002a31ee
	ADDD.EQ	F3, F4, F5    // 035b340e
	ADDF.NE	F0, F2        // 002a321e
	ADDD	F3, F5        // 035b35ee

	END
