// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build mips64 mips64le

#include "textflag.h"

#define SYNC	WORD $0xf

// uint32 runtime·atomicload(uint32 volatile* addr)
TEXT ·atomicload(SB),NOSPLIT,$-8-12
	MOVV	addr+0(FP), R1
	SYNC
	MOVWU	0(R1), R1
	SYNC
	MOVW	R1, ret+8(FP)
	RET

// uint64 runtime·atomicload64(uint64 volatile* addr)
TEXT ·atomicload64(SB),NOSPLIT,$-8-16
	MOVV	addr+0(FP), R1
	SYNC
	MOVV	0(R1), R1
	SYNC
	MOVV	R1, ret+8(FP)
	RET

// void *runtime·atomicloadp(void *volatile *addr)
TEXT ·atomicloadp(SB),NOSPLIT,$-8-16
	MOVV	addr+0(FP), R1
	SYNC
	MOVV	0(R1), R1
	SYNC
	MOVV	R1, ret+8(FP)
	RET

TEXT ·publicationBarrier(SB),NOSPLIT,$-8-0
	SYNC
	RET
