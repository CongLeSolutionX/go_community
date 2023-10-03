// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//go:build wasm32

#include "textflag.h"

// See memclrNoHeapPointers Go doc for important implementation constraints.

// func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr)
TEXT runtime·memclrNoHeapPointers(SB), NOSPLIT, $0-16
	MOVW ptr+0(FP), R0
	MOVW n+4(FP), R1

	Get R0
	I32Const $0
	Get R1
	MemoryFill
	RET
