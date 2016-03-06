// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !darwin,!linux

package net

import "syscall"

func probeTCPStack() (supportsPassiveTCPFastOpen, supportsActiveTCPFastOpen bool) {
	return
}
