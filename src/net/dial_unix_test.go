// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"context"
	"net/internal/socktest"
	"syscall"
	"testing"
	"time"
)

// Issue 16523
func TestDialContextCancelRace(t *testing.T) {
	oldTestHookCanceledDial := testHookCanceledDial
	defer func() { testHookCanceledDial = oldTestHookCanceledDial }()

	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	sawCancel := make(chan bool, 1)
	testHookCanceledDial = func() { sawCancel <- true }
	ctx, cancelCtx := context.WithCancel(context.Background())
	afterConnect := func(st *socktest.State) error {
		t.Log("connect", st)
		if st.Err == nil {
			// On some operating systems, localhost
			// connects _sometimes_ succeed immediately.
			// Prevent that, so we exercise the code path
			// we're interested in testing. This seems
			// harmless. It makes FreeBSD 10.10 work when
			// run with many iterations. It failed about
			// half the time previously.
			return syscall.EINPROGRESS
		}
		return nil
	}
	afterGetsockopt := func(st *socktest.State) error {
		t.Log("getsockopt", st)
		if st.SocketErr == syscall.Errno(0) {
			t.Log("canceling context")

			// Cancel the context at just the moment which
			// caused the race in issue 16523.
			cancelCtx()

			// And wait for the "interrupter" goroutine to
			// cancel the dial by messing with its write
			// timeout before returning.
			select {
			case <-sawCancel:
				t.Log("saw cancel")
			case <-time.After(5 * time.Second):
				t.Error("didn't see cancel after 5 seconds")
			}
		}
		return nil
	}

	callpathSW.Register("TestDialContextCancelRace", func(s uintptr, cookie socktest.Cookie) {
		callpathSW.AddFilter(s, socktest.FilterConnect, func(st *socktest.State) (socktest.AfterFilter, error) {
			return afterConnect, nil
		})
		callpathSW.AddFilter(s, socktest.FilterGetsockoptInt, func(st *socktest.State) (socktest.AfterFilter, error) {
			return afterGetsockopt, nil
		})
	})
	defer callpathSW.Deregister("TestDialContextCancelRace")

	var d Dialer
	c, err := d.DialContext(ctx, ln.Addr().Network(), ln.Addr().String())
	if err == nil {
		c.Close()
		t.Fatal("unexpected successful dial; want context canceled error")
	}

	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("expected context to be canceled")
	}

	oe, ok := err.(*OpError)
	if !ok || oe.Op != "dial" {
		t.Fatalf("Dial error = %#v; want dial *OpError", err)
	}

	if oe.Err != errCanceled {
		t.Errorf("DialContext = (%v, %v); want OpError with error %v", c, err, errCanceled)
	}
}
