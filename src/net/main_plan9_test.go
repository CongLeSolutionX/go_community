// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

func installTestHooks() {}

// forceCloseSockets must be called only from TestMain.
func forceCloseSockets() {}

func disableSocketConnect(tag, network string) {}
