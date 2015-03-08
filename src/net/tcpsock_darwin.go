// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

func probeTCPStack() (supportsPassiveTCPFastOpen, supportsActiveTCPFastOpen bool) {
	n, err := syscall.SysctlUint32("net.inet.tcp.fastopen")
	if err != nil {
		return
	}
	if n&0x1 != 0 {
		supportsPassiveTCPFastOpen = true
	}
	if n&0x2 != 0 {
		// TODO(mikioh): We will enable active open feature
		// when syscall.Syscall9(SYS_CONNECTX, ...) is plugged
		// in net package.
		//supportsActiveTCPFastOpen = true
	}
	return
}
