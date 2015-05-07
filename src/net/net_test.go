// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestCloseRead(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	for _, network := range []string{"tcp", "unix", "unixpacket"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(ln.Addr().String())
		}
		defer ln.Close()

		c, err := Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(c.LocalAddr().String())
		}
		defer c.Close()

		switch c := c.(type) {
		case *TCPConn:
			err = c.CloseRead()
		case *UnixConn:
			err = c.CloseRead()
		}
		if err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		var b [1]byte
		n, err := c.Read(b[:])
		if n != 0 || err == nil {
			t.Fatalf("got (%d, %v); want (0, error)", n, err)
		}
	}
}

func TestCloseWrite(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "plan9":
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	handler := func(ls *localServer, ln Listener) {
		c, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer c.Close()

		var b [1]byte
		n, err := c.Read(b[:])
		if n != 0 || err != io.EOF {
			t.Errorf("got (%d, %v); want (0, io.EOF)", n, err)
			return
		}
		switch c := c.(type) {
		case *TCPConn:
			err = c.CloseWrite()
		case *UnixConn:
			err = c.CloseWrite()
		}
		if err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Error(err)
			return
		}
		n, err = c.Write(b[:])
		if err == nil {
			t.Errorf("got (%d, %v); want (any, error)", n, err)
			return
		}
	}

	for _, network := range []string{"tcp", "unix", "unixpacket"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ls, err := newLocalServer(network)
		if err != nil {
			t.Fatal(err)
		}
		defer ls.teardown()
		if err := ls.buildup(handler); err != nil {
			t.Fatal(err)
		}

		c, err := Dial(ls.Listener.Addr().Network(), ls.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(c.LocalAddr().String())
		}
		defer c.Close()

		switch c := c.(type) {
		case *TCPConn:
			err = c.CloseWrite()
		case *UnixConn:
			err = c.CloseWrite()
		}
		if err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		var b [1]byte
		n, err := c.Read(b[:])
		if n != 0 || err != io.EOF {
			t.Fatalf("got (%d, %v); want (0, io.EOF)", n, err)
		}
		n, err = c.Write(b[:])
		if err == nil {
			t.Fatalf("got (%d, %v); want (any, error)", n, err)
		}
	}
}

func TestConnClose(t *testing.T) {
	for _, network := range []string{"tcp", "unix", "unixpacket"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(ln.Addr().String())
		}
		defer ln.Close()

		c, err := Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(c.LocalAddr().String())
		}
		defer c.Close()

		if err := c.Close(); err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		var b [1]byte
		n, err := c.Read(b[:])
		if n != 0 || err == nil {
			t.Fatalf("got (%d, %v); want (0, error)", n, err)
		}
	}
}

func TestListenerClose(t *testing.T) {
	for _, network := range []string{"tcp", "unix", "unixpacket"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		ln, err := newLocalListener(network)
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unix", "unixpacket":
			defer os.Remove(ln.Addr().String())
		}
		defer ln.Close()

		if err := ln.Close(); err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		c, err := ln.Accept()
		if err == nil {
			c.Close()
			t.Fatal("should fail")
		}
	}
}

func TestPacketConnClose(t *testing.T) {
	for _, network := range []string{"udp", "unixgram"} {
		if !testableNetwork(network) {
			t.Logf("skipping %s test", network)
			continue
		}

		c, err := newLocalPacketListener(network)
		if err != nil {
			t.Fatal(err)
		}
		switch network {
		case "unixgram":
			defer os.Remove(c.LocalAddr().String())
		}
		defer c.Close()

		if err := c.Close(); err != nil {
			if perr := parseCloseError(err); perr != nil {
				t.Error(perr)
			}
			t.Fatal(err)
		}
		var b [1]byte
		n, _, err := c.ReadFrom(b[:])
		if n != 0 || err == nil {
			t.Fatalf("got (%d, %v); want (0, error)", n, err)
		}
	}
}

// See golang.org/issue/6987.
func TestAcceptIgnoreBrokenConnRequest(t *testing.T) {
	switch runtime.GOOS {
	case "android", "nacl":
		t.Skip("not supported on %s", runtime.GOOS)
	case "darwin":
		switch runtime.GOARCH {
		case "arm", "arm64":
			t.Skip("not supported on %s/%s", runtime.GOOS, runtime.GOOS)
		}
	}

	const ipcKey = "GO_NETTEST_DIAL_ADDR"

	if addr := os.Getenv(ipcKey); addr != "" {
		c, err := Dial("tcp", addr)
		if err != nil {
			os.Exit(1)
		}
		defer c.Close()
		os.Stdout.Close()
		// In child process, the process will be killed here.
		time.Sleep(time.Minute)
		os.Exit(1)
	}

	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	// Start a child process that connects to the listener, then
	// kill it because it's the easy way to make SYN_RCVD state
	// sockets in pending queue become broken connection requests.
	cmd := exec.Command(os.Args[0], "-test.run=TestAcceptIgnoreBrokenConnRequest", "-test.short=true")
	cmd.Env = append(os.Environ(), ipcKey+"="+ln.Addr().String())
	pr, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var b [1]byte
	pr.Read(b[:])
	defer cmd.Wait()
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	// Give a chance to send SYN_RESET to the kernel.
	time.Sleep(200 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			if perr := parseDialError(err); perr != nil {
				t.Error(perr)
			}
			t.Error(err)
			return
		}
		c.Close()
	}()
	if err := ln.(*TCPListener).SetDeadline(time.Now().Add(200 * time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			if perr := parseAcceptError(err); perr != nil {
				t.Error(perr)
			}
			if nerr, ok := err.(Error); !ok || !nerr.Timeout() {
				t.Fatal(err)
			}
			break
		}
		c.Close()
	}
	wg.Wait()
}
