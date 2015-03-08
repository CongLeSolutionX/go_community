// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package net

import (
	"context"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
)

func BenchmarkTCPFirstDataDelivery(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	ln, err := newLocalListener("tcp")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCPFirstDataDelivery(b, ln, &d)
}

func BenchmarkTCPFastOpenFirstDataDelivery(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

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
				nr, err := io.ReadFull(c, buf[:])
				if err != nil {
					b.Error(err)
				}
				if nr != msgLen {
					b.Errorf("got %d; want %d", nr, msgLen)
				}
				c.Close()
				wg.Done()
			}(c)
		}
	}(ln)

	b.ResetTimer()

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

func BenchmarkTCP4FastOpenOneShot(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCP(b, false, false, ln, &d)
}

func BenchmarkTCP4FastOpenOneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true, Timeout: time.Hour} // not intended to fire
	benchmarkTCP(b, false, true, ln, &d)
}

func BenchmarkTCP4FastOpenPersistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCP(b, true, false, ln, &d)
}

func BenchmarkTCP4FastOpenPersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true, Timeout: time.Hour} // not intended to fire
	benchmarkTCP(b, true, true, ln, &d)
}

func BenchmarkTCP6FastOpenOneShot(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCP(b, false, false, ln, &d)
}

func BenchmarkTCP6FastOpenOneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true, Timeout: time.Hour} // not intended to fire
	benchmarkTCP(b, false, true, ln, &d)
}

func BenchmarkTCP6FastOpenPersistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCP(b, true, false, ln, &d)
}

func BenchmarkTCP6FastOpenPersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true, Timeout: time.Hour} // not intended to fire
	benchmarkTCP(b, true, true, ln, &d)
}

func BenchmarkTCP4FastOpenConcurrentReadWrite(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCPConcurrentReadWrite(b, ln, &d)
}

func BenchmarkTCP6FastOpenConcurrentReadWrite(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		b.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	d := Dialer{FastOpen: true}
	benchmarkTCPConcurrentReadWrite(b, ln, &d)
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
	defer ln.Close()
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.teardown()
	tpch := make(chan error, 1)
	handler := func(ls *localServer, ln Listener) {
		persistentTransponder(ln, 1, ifi.MTU, nil, tpch) // msgLen must be greater than TCP MSS for reassembly test
	}
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
	trch := make(chan error, 1)
	b := make([]byte, ifi.MTU) // msgLen must be greater than TCP MSS for reassembly test
	go transceiver(c, b, trch)

	go func() { // for data race detection
		for i := 0; i < 10; i++ {
			c.LocalAddr()
			c.RemoteAddr()
			c.SetDeadline(time.Now().Add(someTimeout))
			c.SetWriteDeadline(time.Now().Add(someTimeout))
			c.SetReadDeadline(time.Now().Add(someTimeout))
		}
	}()

	for err := range trch {
		t.Error(err)
	}
	for _, addr := range []Addr{c.LocalAddr(), c.RemoteAddr()} {
		if a, ok := addr.(*TCPAddr); !ok || a.Port == 0 {
			t.Errorf("got %v; expected a proper address with non-zero port number", addr)
		}
	}
	ln.Close()
	for err := range tpch {
		t.Error(err)
	}
}

func TestTCPFastOpenDialerTimeout(t *testing.T) {
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		t.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	d := Dialer{FastOpen: true}
	for _, tmo := range []time.Duration{-3 * time.Second, 0} {
		d.Timeout = tmo
		c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			if perr := parseDialError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		defer c.Close()

		// The first Write call must follow the Dialer's
		// timeout and deadline values, and must ignore write
		// deadline values.
		c.SetDeadline(aLongTimeAgo)
		c.SetWriteDeadline(aLongTimeAgo)
		_, err = c.Write([]byte("TCP FATOPEN FIRST WRITE DIALER TIMEOUT TEST"))
		if err == nil {
			if d.Timeout != 0 {
				t.Fatal("should fail")
			}
			continue
		}
		if perr := parseWriteError(err); perr != nil {
			t.Error(perr)
		}
		if nerr, ok := err.(Error); !ok || !nerr.Timeout() {
			t.Fatal(err)
		}
	}
}

func TestTCPFastOpenDialContextCancel(t *testing.T) {
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		t.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	ctx, cancel := context.WithCancel(context.Background())
	d := Dialer{FastOpen: true}
	c, err := d.DialContext(ctx, ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		if perr := parseDialError(err); perr != nil {
			t.Error(perr)
		}
		t.Fatal(err)
	}
	defer c.Close()

	cancel()
	_, err = c.Write([]byte("TCP FATOPEN FIRST WRITE DIALER CANCEL TEST"))
	if err == nil {
		t.Fatal("should fail")
	}
	if perr := parseWriteError(err); perr != nil {
		t.Error(perr)
	}
	if operr, ok := err.(*OpError); !ok || operr.Err != errCanceled {
		t.Fatal(err)
	}
}

func TestTCPFastOpenStress(t *testing.T) {
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		t.Skip("TCP fast open is not supported")
	}

	ln, err := newLocalFastOpenListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	nmsgs := int(1e4)
	if testing.Short() {
		nmsgs = 1e2
	}
	testTCPStress(t, ln, 2, nmsgs, 512)
}

func TestTCPFastOpenSelfConnect(t *testing.T) {
	if !supportsPassiveTCPFastOpen || !supportsActiveTCPFastOpen {
		t.Skip("TCP fast open is not supported")
	}

	d := Dialer{FastOpen: true}
	testTCPSelfConnect(t, &d, 100)
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
