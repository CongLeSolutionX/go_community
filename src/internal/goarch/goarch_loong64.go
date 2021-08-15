// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build loong64
// +build loong64

package goarch

const (
	_ArchFamily          = LOONG64
	_BigEndian           = false
	_DefaultPhysPageSize = 16384
	_PCQuantum           = 4
	_Int64Align          = 8
	_MinFrameSize        = 8
	_StackAlign          = PtrSize
)

type Uintreg uint64
