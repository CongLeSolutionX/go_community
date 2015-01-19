// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	"unsafe"
)

//go:noescape
func mach_msg_trap(h unsafe.Pointer, op int32, send_size, rcv_size, rcv_name, timeout, notify uint32) int32

func mach_reply_port() uint32
func Mach_task_self() uint32

//go:noescape
func Sigaction(mode uint32, new, old *Sigactiont)

func sigtramp()

func Raise(int32)
