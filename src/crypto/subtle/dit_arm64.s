//go:build !purego

#include "textflag.h"

TEXT ·enableDIT(SB),$0-1
    MRS DIT, R0
    UBFX $24, R0, $1, R1
    MOVB R1, ret+0(FP)
    MSR $1, DIT
    RET

TEXT ·ditEnabled(SB),$0-1
    MRS DIT, R0
    UBFX $24, R0, $1, R1
    MOVB R1, ret+0(FP)
    RET

TEXT ·disableDIT(SB),$0
    MSR $0, DIT
    RET
