// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Simple context type.

package http

import (
	"errors"
	"io"
	"sync"
	"time"
)

// A context is a very light version of golang.org/x/net/context.
// It supports cancelation.
// A nil *context is a valid context. (it just can't be canceled)
type context struct {
	Cancel     chan struct{} // closed on cancel
	cancelOnce sync.Once
}

func newContext() *context {
	return &context{
		Cancel: make(chan struct{}),
	}
}

var errCanceled = errors.New("net/http: operation canceled")

func (ctx *context) cancel()       { ctx.cancelOnce.Do(ctx.closeChannel) }
func (ctx *context) closeChannel() { close(ctx.Cancel) }

func (ctx *context) valid() error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Cancel:
		return errCanceled
	default:
	}
	return nil
}

func newContextReader(ctx *context, r io.Reader) io.Reader {
	return &contextReader{
		ctx:    ctx,
		r:      r,
		readch: make(chan intError, 1),
	}
}

type intError struct {
	int
	error
}

// A contextReader is a cancelable Reader wrapper.
type contextReader struct {
	ctx    *context
	r      io.Reader
	readch chan intError
}

func (r *contextReader) Read(p []byte) (n int, err error) {
	if r.ctx == nil {
		return r.r.Read(p)
	}
	// Check whether it's canceled before even starting the read
	// goroutine.
	if err := r.ctx.valid(); err != nil {
		return 0, err
	}
	go func() {
		n, err := r.r.Read(p)
		r.readch <- intError{n, err}
	}()
	select {
	case v := <-r.readch:
		return v.int, v.error
	case <-r.ctx.Cancel:
		// At this point we still have a goroutine running in
		// the Read call.  To avoid races like in issue 12796
		// where people reuse readers between subsequent
		// requests, we want to wait for that reader to
		// finish.  There's no great number
		// here. Unfortunately contexts & cancelation aren't
		// part of the io.Reader interface.
		t := time.NewTimer(250 * time.Millisecond)
		select {
		case <-t.C:
			// We tried.
			// Hope the user isn't re-using a reader with mutable state.
			// See Issue 12796.
		case <-r.readch:
			t.Stop()
		}
		return 0, errCanceled
	}

}
