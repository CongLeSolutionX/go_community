// Copyright 2020 The Go Authors. All rights reserved.
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

func checkPipes(fds [][2]int) bool {
	for _, pipe := range fds {
		// Check whether the each pipe has been closed.
		_, _, errno1 := syscall.Syscall(unix.FcntlSyscall, uintptr(pipe[0]), syscall.F_GETFD, 0)
		_, _, errno2 := syscall.Syscall(unix.FcntlSyscall, uintptr(pipe[1]), syscall.F_GETFD, 0)
		if errno1 == 0 || errno2 == 0 {
			return false
		}
	}
	return true
}

func TestSplicePipePool(t *testing.T) {
	const N = 64
loop:
	for try := 0; try < 3; try++ {
		if try == 1 && testing.Short() {
			break
		}
		var fds [][2]int
		for i := 0; i < N; i++ {
			p, _, err := poll.GetPipe()
			if err != nil {
				t.Skip("failed to create pipe, skip this test")
			}
			prfd, pwfd := poll.GetPipePair(p)
			poll.PutPipe(p)
			p = nil
			fds = append(fds, [2]int{prfd, pwfd})
		}

		// Trigger garbage collection to free the pipe in sync.Pool and test whether or not
		// the pipe buffer has been closed as we expected.
		for i := 0; i < 5; i++ {
			runtime.GC()
			time.Sleep(time.Duration(i*100+10) * time.Millisecond)
			if checkPipes(fds) {
				continue loop
			}
		}
		t.Fatal("at least one pipe is still open")
	}
}
