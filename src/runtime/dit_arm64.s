//go:build !purego

#include "textflag.h"

TEXT ·enableDIT(SB),$0
    MSR $1, DIT
    RET
