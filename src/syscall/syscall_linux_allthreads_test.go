// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux
// +build !ppc64

package syscall_test

import (
	"runtime"
	"syscall"
	"testing"
)

// reference uapi/linux/prctl.h
const (
	PR_GET_KEEPCAPS uintptr = 7
	PR_SET_KEEPCAPS         = 8
)

// TestAllThreadsSyscall tests that the go runtime can perform
// syscalls that execute on all OSThreads - with which to support
// POSIX semantics for security state changes.
func TestAllThreadsSyscall(t *testing.T) {
	if _, _, err := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, PR_SET_KEEPCAPS, 0, 0); err == syscall.ENOTSUP {
		t.Skip("AllThreadsSyscall disabled with cgo")
	}

	fns := []struct {
		label string
		fn    func(uintptr) error
	}{
		{
			label: "prctl<3-args>",
			fn: func(v uintptr) error {
				_, _, e := syscall.AllThreadsSyscall(syscall.SYS_PRCTL, PR_SET_KEEPCAPS, v, 0)
				if e != 0 {
					return e
				}
				return nil
			},
		},
		{
			label: "prctl<6-args>",
			fn: func(v uintptr) error {
				_, _, e := syscall.AllThreadsSyscall6(syscall.SYS_PRCTL, PR_SET_KEEPCAPS, v, 0, 0, 0, 0)
				if e != 0 {
					return e
				}
				return nil
			},
		},
	}

	waiter := func(q <-chan uintptr, r chan<- uintptr, once bool) {
		for x := range q {
			runtime.LockOSThread()
			v, _, e := syscall.Syscall(syscall.SYS_PRCTL, PR_GET_KEEPCAPS, 0, 0)
			if e != 0 {
				t.Errorf("tid=%d prctl(PR_GET_KEEPCAPS) failed: %v", syscall.Gettid(), e)
			} else if x != v {
				t.Errorf("tid=%d prctl(PR_GET_KEEPCAPS) mismatch: got=%d want=%d", syscall.Gettid(), v, x)
			}
			r <- v
			if once {
				break
			}
			runtime.UnlockOSThread()
		}
	}

	// launches per fns member.
	const launches = 11
	question := make(chan uintptr)
	response := make(chan uintptr)
	defer close(question)

	routines := 0
	for i, v := range fns {
		for j := 0; j < launches; j++ {
			// Add another goroutine - the closest thing
			// we can do to encourage more OS thread
			// creation - while the test is running.  The
			// actual thread creation may or may not be
			// needed, based on the number of available
			// unlocked OS threads at the time waiter
			// calls runtime.LockOSThread(), but the goal
			// of doing this every time through the loop
			// is to race thread creation with v.fn(want)
			// being executed. Via the once boolean we
			// also encourage one in 5 waiters to return
			// locked after participating in only one
			// question response sequence. This allows the
			// test to race thread destruction too.
			once := routines%5 == 4
			go waiter(question, response, once)

			// Keep a count of how many goroutines are
			// going to participate in the
			// question/response test. This will count up
			// towards 2*launches minus the count of
			// routines that have been invoked with
			// once=true.
			routines++

			// Decide what value we want to set the
			// process-shared KEEPCAPS. Note, there is
			// an explicit repeat of 0 when we change the
			// variant of the syscall being used.
			want := uintptr(j & 1)

			// Invoke the AllThreadsSyscall* variant.
			if err := v.fn(want); err != nil {
				t.Errorf("[%d,%d] %s(PR_SET_KEEPCAPS, %d, ...): %v", i, j, v.label, j&1, err)
			}

			// At this point, we want all launched Go
			// routines to confirm that they see the
			// wanted value for KEEPCAPS.
			for k := 0; k < routines; k++ {
				question <- want
			}

			// At this point, we should have a large
			// number of locked OS threads all wanting to
			// reply.
			for k := 0; k < routines; k++ {
				if got := <-response; got != want {
					t.Errorf("[%d,%d,%d] waiter result got=%d, want=%d", i, j, k, got, want)
				}
			}

			// Provide an explicit opportunity for this Go
			// routine to change Ms.
			runtime.Gosched()

			if once {
				// One waiter routine will have exited.
				routines--
			}

			// Whatever M we are now running on, confirm
			// we see the wanted value too.
			if v, _, e := syscall.Syscall(syscall.SYS_PRCTL, PR_GET_KEEPCAPS, 0, 0); e != 0 {
				t.Errorf("[%d,%d] prctl(PR_GET_KEEPCAPS) failed: %v", i, j, e)
			} else if v != want {
				t.Errorf("[%d,%d] prctl(PR_GET_KEEPCAPS) gave wrong value: got=%v, want=1", i, j, v)
			}
		}
	}
}
