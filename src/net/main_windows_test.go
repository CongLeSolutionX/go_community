// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"syscall"
	"time"
)

var (
	testHookDialChannel = func() { time.Sleep(time.Millisecond) } // see golang.org/issue/5349
	testErrnoInProgress = syscall.ERROR_IO_PENDING

	// Placeholders for saving original socket system calls.
	origSocket      = socketFunc
	origClosesocket = closeFunc
	origConnect     = connectFunc
	origConnectEx   = connectExFunc
)

func installTestHooks() {
	socketFunc = sw.Socket
	closeFunc = sw.Closesocket
	connectFunc = sw.Connect
	connectExFunc = sw.ConnectEx
}

func uninstallTestHooks() {
	socketFunc = origSocket
	closeFunc = origClosesocket
	connectFunc = origConnect
	connectExFunc = origConnectEx
}

func forceCloseSockets() {
	for s := range sw.Sockets() {
		closeFunc(s)
	}
}
