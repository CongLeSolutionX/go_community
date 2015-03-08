// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package net

import (
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
)

func TestTCPFastOpen(t *testing.T) {
	ifi := loopbackInterface()
	if !supportsPassiveTCPFastOpen || ifi == nil {
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	ln, err := newLocalTCPFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	ch := make(chan error, 1)
	handler := func(ls *localServer, ln Listener) { persistentTransponder(ln, ifi.MTU, ch) }
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := ls.buildup(handler); err != nil {
		ls.teardown()
		t.Fatal(err)
	}

	var d Dialer
	for _, toggle := range []bool{true, false, true} {
		d.FastOpen = toggle
		c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			if perr := parseDialError(err); perr != nil {
				t.Error(perr)
			}
			ls.teardown()
			t.Fatal(err)
		}
		defer c.Close()

		b := make([]byte, ifi.MTU) // must be greater than TCP MSS
		nw, err := c.Write(b)
		if err != nil {
			if perr := parseWriteError(err); perr != nil {
				t.Error(perr)
			}
			ls.teardown()
			t.Fatal(err)
		}
		if nw != len(b) {
			t.Errorf("fastopen=%t: got %d; want %d", toggle, nw, len(b))
		}

		for _, a := range []Addr{c.LocalAddr(), c.RemoteAddr()} {
			if aa, ok := a.(*TCPAddr); !ok || aa.Port == 0 {
				t.Errorf("got %v; expected a proper address with non-zero port number", a)
			}
		}

		nr, err := c.Read(b)
		if err != nil {
			if perr := parseReadError(err); perr != nil {
				t.Error(perr)
			}
			ls.teardown()
			t.Fatal(err)
		}
		if nr != nw {
			t.Errorf("fastopen=%t: got %d; want %d", toggle, nr, nw)
		}
	}

	ls.teardown()
	for err := range ch {
		t.Error(err)
	}
}

func TestTCPFastOpenConcurrentWrite(t *testing.T) {
	if !supportsPassiveTCPFastOpen {
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	ln, err := newLocalTCPFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	handler := func(ls *localServer, ln Listener) {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		var b [64]byte
		for {
			if _, err := c.Read(b[:]); err != nil {
				break
			}
		}
	}
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.teardown()
	if err := ls.buildup(handler); err != nil {
		t.Fatal(err)
	}

	d := Dialer{FastOpen: true}
	c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		if perr := parseDialError(err); perr != nil {
			t.Error(perr)
		}
		t.Fatal(err)
	}
	defer c.Close()

	var b [64]byte
	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			if _, err := c.Write(b[:]); err != nil {
				if perr := parseWriteError(err); perr != nil {
					t.Error(perr)
				}
				t.Error(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func newLocalTCPFastOpenListener(network string) (Listener, error) {
	ln, err := newLocalListener(network)
	if err != nil {
		return nil, err
	}
	f, err := ln.(*TCPListener).File()
	ln.Close()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var n, v int
	switch runtime.GOOS {
	case "darwin":
		n, v = 0x105, 1
	case "linux":
		n, v = 0x17, listenerBacklog
	}
	if err := syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_TCP, n, v); err != nil {
		return nil, os.NewSyscallError("setsockopt", err)
	}
	return FileListener(f)
}

func BenchmarkTCPFirstDataDeliveryBy3WayHandShake(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	b.StopTimer()
	ln, err := newLocalListener("tcp")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCPFirstDataDelivery(b, ln, &d)
}

func BenchmarkTCPFirstDataDeliveryByFastOpen(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skipf("not supported on %s", runtime.GOOS)
	}

	b.StopTimer()
	ln, err := newLocalTCPFastOpenListener("tcp")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCPFirstDataDelivery(b, ln, &d)
}

func benchmarkTCPFirstDataDelivery(b *testing.B, ln Listener, d *Dialer) {
	ch := make(chan error)
	var wg sync.WaitGroup
	go func(ln Listener) {
		defer close(ch)
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(c Conn) {
				var data [512]byte
				if _, err := c.Read(data[:]); err != nil {
					b.Error(err)
				}
				c.Close()
				wg.Done()
			}(c)
		}
	}(ln)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			b.Error(err)
			continue
		}
		var data [512]byte
		if _, err := c.Write(data[:]); err != nil {
			b.Error(err)
		}
		c.Close()
	}
	ln.Close()
	wg.Wait() // wait for reader goroutines to stop
	<-ch      // wait for acceptor goroutine to stop
}
