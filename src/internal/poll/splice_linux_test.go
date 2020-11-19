// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll_test

import (
	"internal/poll"
	"internal/syscall/unix"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestSplicePipePool(t *testing.T) {
	p, _, err := poll.GetPipe()
	if err != nil {
		t.Skip("failed to create pipe skip")
	}
	prfd, pwfd := poll.GetPipePair(p)
	poll.PutPipe(p)
	p = nil

	// Trigger a garbage collection to free the pipe in sync.Pool and test whether or not
	// the pipe buffer has been closed as we expected.
	runtime.GC()
	time.Sleep((100 + 10) * time.Millisecond)
	runtime.GC()
	time.Sleep((2*100 + 10) * time.Millisecond)

	// Check whether the pipe has been closed.
	_, _, errno1 := syscall.Syscall(unix.FcntlSyscall, uintptr(prfd), syscall.F_GETFD, 0)
	_, _, errno2 := syscall.Syscall(unix.FcntlSyscall, uintptr(pwfd), syscall.F_GETFD, 0)
	if errno1 == 0 || errno2 == 0 {
		t.Fatalf("pipe is still open, prfd errno: %d, pwfd errno: %d\n", errno1, errno2)
	}
}
