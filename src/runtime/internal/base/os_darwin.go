// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

func Bsdthread_create(stk, arg unsafe.Pointer, fn uintptr) int32

//go:noescape
func mach_msg_trap(h unsafe.Pointer, op int32, send_size, rcv_size, rcv_name, timeout, notify uint32) int32

func mach_reply_port() uint32
func Mach_task_self() uint32

//go:noescape
func Sigprocmask(how uint32, new, old *uint32)

//go:noescape
func sigaction(mode uint32, new, old *sigactiont)

//go:noescape
func sigaltstack(new, old *stackt)

func sigtramp()

//go:noescape
func setitimer(mode int32, new, old *itimerval)

func Raise(sig int32)
func raiseproc(int32)
