#include "textflag.h"

TEXT ·f(SB),0,$0-16
	MOVW	x+0(FP), R1
	MOVW	x+4(FP), R2
	DIVU	R1, R2
	DIV	R1, R2
	MODU	R1, R2
	MOD	R1, R2
        RET
