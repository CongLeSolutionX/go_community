// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"context"
	"syscall"
	"testing"
	"time"
)

// Issue 16523
func TestDialContextCancelRace(t *testing.T) {
	oldGetsockoptIntFunc := getsockoptIntFunc
	oldTestHookCanceledDial := testHookCanceledDial
	defer func() {
		getsockoptIntFunc = oldGetsockoptIntFunc
		testHookCanceledDial = oldTestHookCanceledDial
	}()

	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	sawCancel := make(chan bool, 1)
	testHookCanceledDial = func() {
		sawCancel <- true
	}

	ctx, cancelCtx := context.WithCancel(context.Background())
	getsockoptIntFunc = func(fd, level, opt int) (val int, err error) {
		val, err = oldGetsockoptIntFunc(fd, level, opt)
		if level == syscall.SOL_SOCKET && opt == syscall.SO_ERROR && err == nil && val == 0 {
			// Cancel the context at just the moment which
			// caused the race in issue 16523.
			cancelCtx()

			// And wait for the "interrupter" goroutine to
			// cancel the dial by messing with its write
			// timeout before returning.
			select {
			case <-sawCancel:
			case <-time.After(5 * time.Second):
				t.Errorf("didn't see cancel after 5 seconds")
			}
		}
		return
	}

	go func() {
		c, err := ln.Accept()
		if err == nil {
			c.Close()
		}
	}()

	var d Dialer
	c, err := d.DialContext(ctx, "tcp", ln.Addr().String())
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
	if oe.Err != ctx.Err() {
		t.Errorf("DialContext = (%v, %v); want OpError with error %v", c, err, ctx.Err())
	}
}
