TEXT runtime∕internal∕atomic·Cas(SB),NOSPLIT,$0
	B	runtime∕internal∕atomic·Armcas(SB)

TEXT runtime∕internal∕atomic·Casp1(SB),NOSPLIT,$0
	B	runtime∕internal∕atomic·Cas(SB)

