// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"
	"net/internal/socktest"
	"os"
	"runtime"
	"syscall"
	"testing"
)

func (e *OpError) isValid() error {
	if e.Op == "" {
		return fmt.Errorf("OpError.Op is empty: %v", e)
	}
	if e.Net == "" {
		return fmt.Errorf("OpError.Net is empty: %v", e)
	}
	switch addr := e.Addr.(type) {
	case *TCPAddr, *UDPAddr, *IPAddr, *IPNet, *UnixAddr, *pipeAddr:
		if addr == nil {
			return fmt.Errorf("OpError.Addr is empty: %v", e)
		}
	}
	return nil
}

var dialErrorParsers = []func(error) (error, bool){
	func(err error) (error, bool) {
		switch err := err.(type) {
		case *OpError:
			if err := err.isValid(); err != nil {
				return err, false
			}
			return err.Err, true
		}
		return fmt.Errorf("unexpected type on 1st nested level: %T", err), false
	},
	func(err error) (error, bool) {
		switch err := err.(type) {
		case syscall.Errno:
			return nil, false
		case *AddrError, *DNSError, InvalidAddrError, *ParseError, UnknownNetworkError, *timeoutError:
			return nil, false
		case *os.SyscallError:
			return err.Err, true
		case *DNSConfigError:
			return err.Err, true
		}
		switch err {
		case errClosing, errMissingAddress:
			return nil, false
		}
		return fmt.Errorf("unexpected type on 2nd nested level: %T", err), false
	},
	func(err error) (error, bool) {
		switch err.(type) {
		case syscall.Errno:
			return nil, false
		}
		return fmt.Errorf("unexpected type on 3rd nested level: %T", err), false
	},
}

func parseDialError(err error) error {
	var more bool
	for _, parser := range dialErrorParsers {
		if err, more = parser(err); !more {
			break
		}
	}
	return err
}

var dialErrorTests = []struct {
	network, address string
}{
	{"foo", ""},
	{"bar", "baz"},
	{"datakit", "mh/astro/r70"},
	{"tcp", ""},
	{"tcp", "127.0.0.1:☺"},
	{"tcp", "no-such-name:80"},
	{"tcp", "mh/astro/r70:http"},

	{"tcp", "127.0.0.1:0"},
	{"udp", "127.0.0.1:0"},
	{"ip:icmp", "127.0.0.1"},

	{"unix", "/path/to/somewhere"},
	{"unixgram", "/path/to/somewhere"},
	{"unixpacket", "/path/to/somewhere"},
}

func TestDialError(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	}

	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = func(fn func(string) ([]IPAddr, error), host string) ([]IPAddr, error) {
		return nil, &DNSError{Err: "dial error test", Name: "name", Server: "server", IsTimeout: true}
	}
	sw.Set(socktest.FilterConnect, func(so *socktest.Status) (socktest.AfterFilter, error) {
		return nil, syscall.EOPNOTSUPP
	})
	defer sw.Set(socktest.FilterConnect, nil)

	d := Dialer{Timeout: someTimeout}
	for i, tt := range dialErrorTests {
		c, err := d.Dial(tt.network, tt.address)
		if err == nil {
			t.Errorf("#%d: should fail; %s:%s->%s", i, tt.network, c.LocalAddr(), c.RemoteAddr())
			c.Close()
			continue
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
	}
}

var listenErrorTests = []struct {
	network, address string
}{
	{"foo", ""},
	{"bar", "baz"},
	{"datakit", "mh/astro/r70"},
	{"tcp", "127.0.0.1:☺"},
	{"tcp", "no-such-name:80"},
	{"tcp", "mh/astro/r70:http"},
}

func TestListenError(t *testing.T) {
	switch runtime.GOOS {
	case "plan9":
		t.Skipf("%s does not have full support of socktest", runtime.GOOS)
	}

	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = func(fn func(string) ([]IPAddr, error), host string) ([]IPAddr, error) {
		return nil, &DNSError{Err: "listen error test", Name: "name", Server: "server", IsTimeout: true}
	}

	for i, tt := range listenErrorTests {
		ln, err := Listen(tt.network, tt.address)
		if err == nil {
			t.Errorf("#%d: should fail; %s:%s->", i, tt.network, ln.Addr())
			ln.Close()
			continue
		}
		if err = parseDialError(err); err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
	}
}
