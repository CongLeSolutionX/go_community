//go:build aix && ppc64

#include "textflag.h"

// Sometimes aix will execute the instruction despite the
// target page being set to no-execute. Ensure icache is
// synchronized to preserve the behavior of this test.
//
// func syncIcache(p uintptr)
TEXT main·syncIcache(SB), NOSPLIT|NOFRAME, $0-0
	SYNC
	MOVD (R3), R3
	ICBI (R3)
	ISYNC
	RET
