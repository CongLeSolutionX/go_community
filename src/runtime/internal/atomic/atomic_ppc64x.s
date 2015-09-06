// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

// uint32 runtime∕internal∕atomic·Atomicload(uint32 volatile* addr)
TEXT ·Atomicload(SB),NOSPLIT,$-8-12
	MOVD	addr+0(FP), R3
	SYNC
	MOVWZ	0(R3), R3
	CMPW	R3, R3, CR7
	BC	4, 30, 1(PC) // bne- cr7,0x4
	ISYNC
	MOVW	R3, ret+8(FP)
	RET

// uint64 runtime∕internal∕atomic·Atomicload64(uint64 volatile* addr)
TEXT ·Atomicload64(SB),NOSPLIT,$-8-16
	MOVD	addr+0(FP), R3
	SYNC
	MOVD	0(R3), R3
	CMP	R3, R3, CR7
	BC	4, 30, 1(PC) // bne- cr7,0x4
	ISYNC
	MOVD	R3, ret+8(FP)
	RET

// void *runtime∕internal∕atomic·Atomicloadp(void *volatile *addr)
TEXT ·Atomicloadp(SB),NOSPLIT,$-8-16
	MOVD	addr+0(FP), R3
	SYNC
	MOVD	0(R3), R3
	CMP	R3, R3, CR7
	BC	4, 30, 1(PC) // bne- cr7,0x4
	ISYNC
	MOVD	R3, ret+8(FP)
	RET
