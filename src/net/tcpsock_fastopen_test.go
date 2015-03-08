// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package net

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
)

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
		b.Skip("TCP fast open is not supported")
	}

	b.StopTimer()
	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCPFirstDataDelivery(b, ln, &d)
}

func benchmarkTCPFirstDataDelivery(b *testing.B, ln Listener, d *Dialer) {
	const msgLen = 512
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ln Listener) {
		defer wg.Done()
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(c Conn) {
				var buf [msgLen]byte
				var nr int
				for nr < msgLen {
					n, err := c.Read(buf[:])
					if err != nil {
						b.Error(err)
						break
					}
					nr += n
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
		var buf [msgLen]byte
		nw, err := c.Write(buf[:])
		if err != nil {
			b.Error(err)
		}
		if nw != msgLen {
			b.Errorf("got %d; want %d", nw, msgLen)
		}
		c.Close()
	}
	ln.Close()
	wg.Wait() // wait for tester goroutines to stop
}

func TestTCPFastOpenSelfConnect(t *testing.T) {
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		t.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	testTCPSelfConnect(t, ln, &d)
}

func TestTCPFastOpen(t *testing.T) {
	ifi := loopbackInterface()
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen || ifi == nil {
		t.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan error, 1)
	handler := func(ls *localServer, ln Listener) {
		persistentTransponder(ln, 1, ifi.MTU, nil, ch) // msgLen must be greater than TCP MSS for reassembly test
	}
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		ln.Close()
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

		b := make([]byte, ifi.MTU) // msgLen must be greater than TCP MSS for reassembly test
		tc := &testConn{Conn: c, ch: ch}
		if !tc.write(b) || !tc.read(b[:tc.nw]) {
			continue
		}
		for _, a := range []Addr{c.LocalAddr(), c.RemoteAddr()} {
			if aa, ok := a.(*TCPAddr); !ok || aa.Port == 0 {
				ch <- fmt.Errorf("got %v; expected a proper address with non-zero port number", a)
			}
		}
		if tc.nr != tc.nw {
			ch <- fmt.Errorf("got %d bytes read; want %d", tc.nr, tc.nw)
		}
	}

	ls.teardown()
	for err := range ch {
		t.Error(err)
	}
}

func newLocalFastOpenListener(network string) (Listener, error) {
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
