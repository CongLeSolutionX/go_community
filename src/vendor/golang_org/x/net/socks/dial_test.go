// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socks

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	targetHostname = "fqdn.doesnotexist"
	targetHostIP   = "2001:db8::1"
	targetPort     = "5963"
)

func TestDial(t *testing.T) {
	var wg sync.WaitGroup

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		ln, err = net.Listen("tcp", "[::1]:0")
	}
	if err != nil {
		t.Fatal(err)
	}
	srvErr := make(chan error, 10)
	defer func() {
		ln.Close()
		wg.Wait()
		close(srvErr)
		for {
			err, ok := <-srvErr
			if !ok {
				return
			}
			if err != nil {
				//t.Log(err)
			}
		}
	}()

	t.Run("Connect", func(t *testing.T) {
		wg.Add(1)
		go proxyServer(ln, srvErr, &wg)
		d, err := NewDialer(ln.Addr().Network(), ln.Addr().String(), CmdConnect)
		if err != nil {
			t.Error(err)
			return
		}
		d.AuthMethods = []AuthMethod{
			AuthMethodNotRequired,
			AuthMethodUsernamePassword,
		}
		d.Authenticate = (&UsernamePassword{"username", "password"}).Authenticate
		dialErr := make(chan error)
		go func() {
			c, err := d.Dial(ln.Addr().Network(), net.JoinHostPort(targetHostIP, targetPort))
			if err == nil {
				t.Log(c.(*Conn).BoundAddr())
				c.Close()
			}
			dialErr <- err
		}()
		if err := <-dialErr; err != nil {
			t.Error(err)
			return
		}
	})
	t.Run("Cancel", func(t *testing.T) {
		wg.Add(1)
		go blackholeProxyServer(ln, srvErr, &wg)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		d, err := NewDialer(ln.Addr().Network(), ln.Addr().String(), CmdConnect)
		if err != nil {
			t.Error(err)
			return
		}
		dialErr := make(chan error)
		go func() {
			c, err := d.DialContext(ctx, ln.Addr().Network(), net.JoinHostPort(targetHostname, targetPort))
			if err == nil {
				c.Close()
			}
			dialErr <- err
		}()
		cancel()
		err = <-dialErr
		if perr, nerr := parseDialError(err); perr != context.Canceled && nerr == nil {
			t.Errorf("got %v; want context.Canceled or equivalent", err)
			return
		}
	})
	t.Run("Deadline", func(t *testing.T) {
		wg.Add(1)
		go blackholeProxyServer(ln, srvErr, &wg)
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
		defer cancel()
		d, err := NewDialer(ln.Addr().Network(), ln.Addr().String(), CmdBind)
		if err != nil {
			t.Error(err)
			return
		}
		dialErr := make(chan error)
		go func() {
			c, err := d.DialContext(ctx, ln.Addr().Network(), net.JoinHostPort(targetHostname, targetPort))
			if err == nil {
				c.Close()
			}
			dialErr <- err
		}()
		err = <-dialErr
		if perr, nerr := parseDialError(err); perr != context.DeadlineExceeded && nerr == nil {
			t.Errorf("got %v; want context.DeadlineExceeded or equivalent", err)
			return
		}
	})
}

func proxyServer(ln net.Listener, ch chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	c, err := ln.Accept()
	if err != nil {
		ch <- err
		return
	}
	defer c.Close()
	b := make([]byte, 257)
	if _, err := io.ReadFull(c, b[:2]); err != nil {
		ch <- err
		return
	}
	io.ReadFull(c, b[:b[1]])
	b[0] = protocolVersion5
	b[1] = byte(AuthMethodNotRequired)
	if _, err := c.Write(b[:2]); err != nil {
		ch <- err
		return
	}
	if _, err := io.ReadFull(c, b[:6+net.IPv6len]); err != nil {
		ch <- err
		return
	}
	if b[0] != protocolVersion5 || Command(b[1]) != CmdConnect || b[2] != 0x00 || b[3] != addrTypeIPv6 {
		ch <- fmt.Errorf("got unexpected message: %#02x %#02x %#02x %#02x", b[0], b[1], b[2], b[3])
		return
	}
	b = b[:0]
	b = append(b, protocolVersion5, statusSucceeded, 0x00, addrTypeIPv6)
	b = append(b, net.ParseIP(targetHostIP)...)
	port, err := strconv.Atoi(targetPort)
	if err != nil {
		ch <- err
		return
	}
	b = append(b, byte(port>>8), byte(port))
	if _, err := c.Write(b); err != nil {
		ch <- err
		return
	}
}

func blackholeProxyServer(ln net.Listener, ch chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	c, err := ln.Accept()
	if err != nil {
		ch <- err
		return
	}
	defer c.Close()
	b := make([]byte, 257)
	if _, err := io.ReadFull(c, b[:2]); err != nil {
		ch <- err
		return
	}
	io.ReadFull(c, b[:b[1]])
	b[0] = protocolVersion5
	b[1] = byte(AuthMethodNotRequired)
	if _, err := c.Write(b[:2]); err != nil {
		ch <- err
		return
	}
	if _, err := io.ReadFull(c, b[:6+1+len(targetHostname)]); err != nil {
		ch <- err
		return
	}
	for {
		if _, err := c.Read(b); err != nil {
			ch <- err
			return
		}
	}
}

func parseDialError(err error) (perr, nerr error) {
	if e, ok := err.(*net.OpError); ok {
		err = e.Err
		nerr = e
	}
	if e, ok := err.(*os.SyscallError); ok {
		err = e.Err
	}
	perr = err
	return
}
