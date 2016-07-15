// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httptest

import (
	"errors"
	"net"
	"sync"
)

// A pipeListener is a net.Listener implementation that serves
// in-memory pipes.
type pipeListener struct {
	addr string

	m    sync.Mutex
	done chan struct{}
	c    chan net.Conn
}

func newPipeListener(addr string) *pipeListener {
	return &pipeListener{
		addr: addr,

		done: make(chan struct{}),
		c:    make(chan net.Conn),
	}
}

func (lis *pipeListener) Accept() (net.Conn, error) {
	select {
	case <-lis.done:
		return nil, errors.New("closed")
	case c := <-lis.c:
		return c, nil
	}
}

func (lis *pipeListener) Close() error {
	lis.m.Lock()
	defer lis.m.Unlock()

	select {
	case <-lis.done:
		return errors.New("closed")
	default:
	}

	close(lis.done)
	return nil
}

func (lis *pipeListener) Addr() net.Addr {
	return pipeAddr(lis.addr)
}

// Dial "dials" the listener, creating a pipe. It returns the client's
// end of the connection and causes a waiting Accept call to return
// the server's end.
func (lis *pipeListener) Dial(network, addr string) (net.Conn, error) {
	s, c := net.Pipe()
	select {
	case lis.c <- s:
	case <-lis.done:
		return nil, errors.New("closed")
	}
	return c, nil
}

type pipeAddr string

func (pipeAddr) Network() string {
	return "pipe"
}

func (addr pipeAddr) String() string {
	return string(addr)
}
