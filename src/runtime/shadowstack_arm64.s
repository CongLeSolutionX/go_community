#include "textflag.h"

TEXT ·shadowTrampolineASM<ABIInternal>(SB),NOSPLIT|NOFRAME,$0-0
	CALL runtime·shadowTrampolineGo<ABIInternal>(SB)
	MOVD R0, R30
	RET
