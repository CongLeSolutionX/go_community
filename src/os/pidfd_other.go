// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !linux && !windows && !plan9

package os

import "syscall"

func pidfdOpen(pid int) (uintptr, error) {
	return unsetHandle, nil
}

func pidfdSendSignal(_ uintptr, _ syscall.Signal) (_ error, done bool) {
	// Not implemented.
	return nil, false
}

func canUsePidfdSendSignal() bool {
	return false
}
