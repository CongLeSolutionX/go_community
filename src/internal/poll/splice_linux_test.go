// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll_test

import (
	"internal/poll"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// checkPipes returns true if all pipes are closed properly, false otherwise.
func checkPipes(fds [][2]int) bool {
	for _, fd := range fds {
		// Check whether the each pipe has been closed.
		err1 := syscall.FcntlFlock(uintptr(fd[0]), syscall.F_GETFD, nil)
		err2 := syscall.FcntlFlock(uintptr(fd[1]), syscall.F_GETFD, nil)
		if err1 == nil || err2 == nil {
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

		var (
			p   *poll.SplicePipe
			ps  []*poll.SplicePipe
			fds [][2]int
			err error
		)
		for i := 0; i < N; i++ {
			p, _, err = poll.GetPipe()
			if err != nil {
				t.Skip("failed to create pipe, skip this test")
			}
			prfd, pwfd := poll.GetPipeFds(p)
			fds = append(fds, [2]int{prfd, pwfd})
			ps = append(ps, p)
		}
		for _, p = range ps {
			poll.PutPipe(p)
		}
		ps = nil

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

func BenchmarkSplicePipe(b *testing.B) {
	b.Run("SplicePipeWithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p, _, _ := poll.GetPipe()
			poll.PutPipe(p)
		}
	})
	b.Run("SplicePipeWithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := poll.NewPipe()
			poll.DestroyPipe(p)
		}
	})
}

func BenchmarkSplicePipePoolParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p, _, _ := poll.GetPipe()
			poll.PutPipe(p)
		}
	})
}

func BenchmarkSplicePipeNativeParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p := poll.NewPipe()
			poll.DestroyPipe(p)
		}
	})
}
