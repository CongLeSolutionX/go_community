// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !arm64 && !amd64

// Dummy assembly file that allows us to linkname into the runtime.

TEXT ·loadRegisters(SB),0,$0-8
	RET
TEXT ·spillRegisters(SB),0,$0-16
	RET
