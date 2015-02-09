#include "textflag.h"

// actually a function descriptor for _main<>(SB)
TEXT _rt0_ppc64_linux(SB),NOSPLIT,$0
	DWORD $_main<>(SB)
	DWORD $0
	DWORD $0

TEXT _main<>(SB),NOSPLIT,$-8
	MOVD 0(R1), R3 // Argc
	ADD $8, R1, R4 // Argv
	BR main(SB)

TEXT main(SB),NOSPLIT,$-8
	MOVD	$runtime∕internal∕schedinit·rt0_go(SB), R31
	MOVD	R31, CTR
	BR	(CTR)
