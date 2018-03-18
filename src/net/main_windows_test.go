// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "internal/poll"

func installTestHooks() {
	socketFunc = callpathSW.Socket
	wsaSocketFunc = callpathSW.WSASocket
	connectFunc = callpathSW.Connect
	listenFunc = callpathSW.Listen

	poll.CloseFunc = callpathSW.Closesocket
	poll.ConnectExFunc = callpathSW.ConnectEx
	poll.AcceptFunc = callpathSW.AcceptEx
}

// forceCloseSockets must be called only from TestMain.
func forceCloseSockets() {
	for _, st := range callpathSW.Sockets() {
		poll.CloseFunc(st.Sysfd)
	}
}
