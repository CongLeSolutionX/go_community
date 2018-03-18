// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package net

import "internal/poll"

var extraTestHookInstallers []func()

func installTestHooks() {
	socketFunc = callpathSW.Socket
	connectFunc = callpathSW.Connect
	listenFunc = callpathSW.Listen
	getsockoptIntFunc = callpathSW.GetsockoptInt

	poll.CloseFunc = callpathSW.Close
	poll.AcceptFunc = callpathSW.Accept

	for _, fn := range extraTestHookInstallers {
		fn()
	}
}

// forceCloseSockets must be called only from TestMain.
func forceCloseSockets() {
	for _, st := range callpathSW.Sockets() {
		poll.CloseFunc(st.Sysfd)
	}
}
