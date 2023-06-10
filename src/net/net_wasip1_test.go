// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build wasip1

package net

import (
	"errors"
	"internal/fakenet"
	"path/filepath"
	"syscall"
	"testing"
)

// WASI preview 1 has limited socket support, so we cannot run the net package
// test suite. However, we do want to validate that programs compiled to wasip1
// will behave in predictable ways, won't panic nor see unexpected errors. The
// tests in this files are intended to validate this.

func withoutFakenet(f func()) {
	enabled := fakenet.IsEnabled()
	defer fakenet.SetEnabled(enabled)
	// Importing the internal/testenv package neables the fake network facility
	// on wasip1, but the tests in this file are testing the behavior of the
	// public interfaces that programs can depend on, where the fake network is
	// disabled.
	fakenet.SetEnabled(false)
	f()
}

func TestWasip1ListenNotSupported(t *testing.T) {
	withoutFakenet(func() {
		tmp := t.TempDir()
		for _, test := range []struct{ network, address string }{
			{"unix", filepath.Join(tmp, "unix.sock")},
			{"unixgram", filepath.Join(tmp, "unixgram.sock")},
			{"tcp", ":0"},
			{"tcp4", ":0"},
			{"tcp6", ":0"},
		} {
			t.Run(test.network, func(t *testing.T) {
				l, err := Listen(test.network, test.address)
				if !errors.Is(err, syscall.EPROTONOSUPPORT) {
					t.Errorf("%s protocol should not be supported on wasip1 (err=%v)", test.network, err)
				}
				if l != nil {
					t.Errorf("listener must be nil but got %T", l)
				}
			})
		}
	})
}

func TestWasip1ListenPacketNotSupported(t *testing.T) {
	withoutFakenet(func() {
		tmp := t.TempDir()
		for _, test := range []struct{ network, address string }{
			{"unixgram", filepath.Join(tmp, "unixgram.sock")},
			{"udp", ":0"},
			{"udp4", ":0"},
			{"udp6", ":0"},
		} {
			t.Run(test.network, func(t *testing.T) {
				l, err := ListenPacket(test.network, test.address)
				if !errors.Is(err, syscall.EPROTONOSUPPORT) {
					t.Errorf("%s protocol should not be supported on wasip1 (err=%v)", test.network, err)
				}
				if l != nil {
					t.Errorf("listener must be nil but got %T", l)
				}
			})
		}
	})
}

func TestWasip1DialNotSupported(t *testing.T) {
	withoutFakenet(func() {
		tmp := t.TempDir()
		for _, test := range []struct{ network, address string }{
			{"unix", filepath.Join(tmp, "unix.sock")},
			{"unixgram", filepath.Join(tmp, "unixgram.sock")},
			{"udp", "127.0.0.1:1234"},
			{"udp4", "127.0.0.1:1234"},
			{"udp6", "[::1]:1234"},
			{"tcp", "127.0.0.1:1234"},
			{"tcp4", "127.0.0.1:1234"},
			{"tcp6", "[::1]:1234"},
		} {
			t.Run(test.network, func(t *testing.T) {
				c, err := Dial(test.network, test.address)
				if !errors.Is(err, syscall.EPROTONOSUPPORT) {
					t.Errorf("%s protocol should not be supported on wasip1 (err=%v)", test.network, err)
				}
				if c != nil {
					t.Errorf("connection must be nil but got %T", c)
				}
			})
		}
	})
}
