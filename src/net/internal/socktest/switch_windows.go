// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socktest

import "syscall"

// A State represents the state of a socket.
type State struct {
	Sysfd     syscall.Handle // socket descriptor
	Cookie    Cookie
	Err       error // error status of socket system call
	SocketErr error // error status of socket by SO_ERROR
}

func (b *binding) newState() *State {
	return &State{
		Sysfd:  syscall.Handle(b.sysfd),
		Cookie: b.cookie,
	}
}
