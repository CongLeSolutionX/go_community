// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_core "runtime/internal/core"
)

//go:noescape
func Jmpdefer(fv *_core.Funcval, argp uintptr)
func Goexit()

func morestack()
func rt0_go()

func systemstack_switch()
