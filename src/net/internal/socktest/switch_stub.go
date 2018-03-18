// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !darwin,!dragonfly,!freebsd,!linux,!nacl,!netbsd,!openbsd,!solaris,!windows

package socktest

// A State represents the state of a socket.
type State struct {
	Sysfd     uintptr // socket descriptor
	Cookie    Cookie
	Err       error // error status of socket system call
	SocketErr error // error status of socket by SO_ERROR
}

func (b *binding) newState() *State { return nil }

func familyString(family int) string { return "<nil>" }

func typeString(sotype int) string { return "<nil>" }

func protocolString(proto int) string { return "<nil>" }
