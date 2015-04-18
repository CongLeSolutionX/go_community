// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"syscall"
)

func fileConn(f *os.File) (Conn, error) {
	// TODO: Implement this
	return nil, syscall.EWINDOWS
}

func fileListener(f *os.File) (l Listener, err error) {
	// TODO: Implement this
	return nil, syscall.EWINDOWS
}

func filePacketConn(f *os.File) (c PacketConn, err error) {
	// TODO: Implement this
	return nil, syscall.EWINDOWS
}
