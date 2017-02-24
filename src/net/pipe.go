// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"io"
	"sync"
	"time"
)

// pipeDeadline is an abstraction for handling timeouts.
// The zero value is ready for use.
type pipeDeadline struct {
	mu     sync.Mutex // Guards timer and cancel
	timer  *time.Timer
	cancel chan struct{}
}

// set sets the point in time when the deadline will time out.
// A timeout event is signaled by closing the channel returned by waiter.
// Once a timeout has occurred, the deadline can be refreshed by specifying a
// t value in the future.
//
// A zero value for t prevents timeout.
func (d *pipeDeadline) set(t time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil && !d.timer.Stop() {
		<-d.cancel // Wait for the timer callback to finish and close cancel
	}
	d.timer = nil

	// Check if cancel is closed already.
	closed := false
	select {
	case <-d.cancel:
		closed = true
	default:
	}

	// Time is zero, then there is no deadline.
	if t.IsZero() {
		if closed || d.cancel == nil {
			d.cancel = make(chan struct{})
		}
		return
	}

	// Time in the future, setup a timer to cancel in the future.
	if dur := t.Sub(time.Now()); dur > 0 {
		if closed || d.cancel == nil {
			d.cancel = make(chan struct{})
		}
		d.timer = time.AfterFunc(dur, func() {
			close(d.cancel)
		})
		return
	}

	// Time in the past, so close immediately.
	if !closed {
		if d.cancel == nil {
			d.cancel = make(chan struct{})
		}
		close(d.cancel)
	}
}

// waiter returns a channel that is closed when the deadline is exceeded.
func (d *pipeDeadline) waiter() chan struct{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel == nil {
		d.cancel = make(chan struct{})
	}
	return d.cancel
}

func isClosedChan(c <-chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

type pipeError struct {
	error   string
	timeout bool
}

func (pe pipeError) Error() string   { return pe.error }
func (pe pipeError) Timeout() bool   { return pe.timeout }
func (pe pipeError) Temporary() bool { return pe.timeout }

var (
	errDeadline = pipeError{"deadline exceeded", true}
	errClosed   = pipeError{"closed connection", false}
)

type pipe struct {
	pr *io.PipeReader
	pw *io.PipeWriter

	readDeadline  pipeDeadline
	writeDeadline pipeDeadline

	mu         sync.Mutex   // Guards closedCh
	readReqCh  chan dataReq // Never closed
	writeReqCh chan dataReq // Never closed
	closedCh   chan struct{}
}

// Pipe creates a synchronous, in-memory, full duplex
// network connection; both ends implement the Conn interface.
// Reads on one end are matched with writes on the other,
// copying data directly between the two; there is no internal
// buffering.
func Pipe() (Conn, Conn) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	return newPipe(r1, w2), newPipe(r2, w1)
}

func newPipe(r *io.PipeReader, w *io.PipeWriter) *pipe {
	p := &pipe{
		pr:         r,
		pw:         w,
		readReqCh:  make(chan dataReq),
		writeReqCh: make(chan dataReq),
		closedCh:   make(chan struct{}),
	}
	go p.readLoop()
	go p.writeLoop()
	return p
}

type (
	dataReq struct {
		b  []byte
		ch chan<- dataResp
	}
	dataResp struct {
		n   int
		err error
	}
)

const chunkSize = 4096

func (p *pipe) Read(b []byte) (int, error) {
	n, err := p.read(b)
	if err != nil {
		err = &OpError{Op: "read", Net: "pipe", Err: err}
	}
	return n, err
}

func (p *pipe) read(b []byte) (int, error) {
	switch {
	case isClosedChan(p.closedCh):
		return 0, errClosed
	case isClosedChan(p.readDeadline.waiter()):
		return 0, errDeadline
	}

	ch := make(chan dataResp)
	select {
	case p.readReqCh <- dataReq{b, ch}:
		res := <-ch
		return res.n, res.err
	case <-p.readDeadline.waiter():
		return 0, errDeadline
	case <-p.closedCh:
		return 0, errClosed
	}
}

func (p *pipe) readLoop() {
	chunk := make([]byte, chunkSize)

	var b []byte
	var err error
	for {
		for len(b) == 0 && err == nil {
			var n int
			n, err = p.pr.Read(chunk)
			if err == io.ErrClosedPipe {
				err = errClosed
			}
			b = chunk[:n]
		}

		select {
		case req := <-p.readReqCh:
			n := copy(req.b, b)
			b = b[n:]
			req.ch <- dataResp{n, err} // Non-buffered, but guaranteed listener in Read
		case <-p.closedCh:
			return
		}
	}
}

func (p *pipe) Write(b []byte) (int, error) {
	n, err := p.write(b)
	if err != nil {
		err = &OpError{Op: "write", Net: "pipe", Err: err}
	}
	return n, err
}

func (p *pipe) write(b []byte) (int, error) {
	switch {
	case isClosedChan(p.closedCh):
		return 0, errClosed
	case isClosedChan(p.writeDeadline.waiter()):
		return 0, errDeadline
	}

	// TODO(dsnet): Optimize away this allocation in each write.
	// Since net.Pipe is mainly used in tests, this is probably not a big deal.
	chunk := make([]byte, chunkSize)

	ch := make(chan dataResp, 1)
	var n int
	for {
		// Copy the input slice and pass that to the writeLoop to ensure that
		// in the event of a deadline, it is still safe for writeLoop to
		// access the chunk even if Write has returned.
		nn := copy(chunk, b)
		b = b[nn:]

		// Send request.
		select {
		case p.writeReqCh <- dataReq{chunk[:nn], ch}:
		case <-p.writeDeadline.waiter():
			return n, errDeadline
		case <-p.closedCh:
			return n, errClosed
		}

		// Receive response.
		select {
		case res := <-ch:
			n += res.n
			if len(b) == 0 || res.err != nil {
				return n, res.err
			}
		case <-p.writeDeadline.waiter():
			return n, errDeadline
		case <-p.closedCh:
			return n, errClosed
		}
	}
}

func (p *pipe) writeLoop() {
	var err error
	for {
		select {
		case req := <-p.writeReqCh:
			var n int
			if err == nil {
				n, err = p.pw.Write(req.b)
				if err == io.ErrClosedPipe {
					err = errClosed
				}
			}
			req.ch <- dataResp{n, err} // Buffered channel from write; won't block
		case <-p.closedCh:
			return
		}
	}
}

func (p *pipe) Close() error {
	err := p.pr.Close()
	err1 := p.pw.Close()
	if err == nil {
		err = err1
	}

	p.mu.Lock()
	if !isClosedChan(p.closedCh) {
		close(p.closedCh)
	}
	p.mu.Unlock()
	return err
}

func (p *pipe) SetDeadline(t time.Time) error {
	if isClosedChan(p.closedCh) {
		return &OpError{Op: "set", Net: "pipe", Err: errClosed}
	}
	p.readDeadline.set(t)
	p.writeDeadline.set(t)
	return nil
}

func (p *pipe) SetReadDeadline(t time.Time) error {
	if isClosedChan(p.closedCh) {
		return &OpError{Op: "set", Net: "pipe", Err: errClosed}
	}
	p.readDeadline.set(t)
	return nil
}

func (p *pipe) SetWriteDeadline(t time.Time) error {
	if isClosedChan(p.closedCh) {
		return &OpError{Op: "set", Net: "pipe", Err: errClosed}
	}
	p.writeDeadline.set(t)
	return nil
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }
func (pipeAddr) String() string  { return "pipe" }

func (p *pipe) LocalAddr() Addr {
	return pipeAddr{}
}

func (p *pipe) RemoteAddr() Addr {
	return pipeAddr{}
}
