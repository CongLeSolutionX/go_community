// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// This runs on a foreign stack, without an m or a g. No stack split.
//go:nosplit
//go:norace
//go:nowritebarrierrec
func badsignal(sig uintptr) {
	cgocallback(unsafe.Pointer(funcPC(badsignalgo)), noescape(unsafe.Pointer(&sig)), unsafe.Sizeof(sig))
}

func badsignalgo(sig uintptr) {
	if !sigsend(uint32(sig)) {
		// A foreign thread received the signal sig, and the
		// Go code does not want to handle it.
		raisebadsignal(int32(sig))
	}
}
