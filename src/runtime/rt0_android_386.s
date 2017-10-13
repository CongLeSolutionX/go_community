// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

TEXT _rt0_386_android(SB),NOSPLIT,$-8
	JMP	_rt0_386(SB)

TEXT _rt0_386_android_lib(SB),NOSPLIT,$0
	PUSHL	$_rt0_386_android_argv(SB)  // argv
	PUSHL	$1  // argc
	JMP	_rt0_386_lib(SB)

TEXT main(SB),NOSPLIT,$-8
	// Remove the return address from the stack.
	// rt0_go doesn't expect it to be there.
	ADDL	$4, SP
	JMP	runtime·rt0_go(SB)

// TODO: wire up necessary VDSO (see os_linux_386.go)
