// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package net

import (
	"net/internal/socktest"
	"strings"
	"syscall"
)

func disableSocketConnect(tag, network string) {
	ss := strings.Split(network, ":")
	callpathSW.Register(tag, func(s uintptr, cookie socktest.Cookie) {
		callpathSW.AddFilter(s, socktest.FilterConnect, func(st *socktest.State) (socktest.AfterFilter, error) {
			switch ss[0] {
			case "tcp4":
				if st.Cookie.Family() == syscall.AF_INET && st.Cookie.Type() == syscall.SOCK_STREAM {
					return nil, syscall.EHOSTUNREACH
				}
			case "udp4":
				if st.Cookie.Family() == syscall.AF_INET && st.Cookie.Type() == syscall.SOCK_DGRAM {
					return nil, syscall.EHOSTUNREACH
				}
			case "ip4":
				if st.Cookie.Family() == syscall.AF_INET && st.Cookie.Type() == syscall.SOCK_RAW {
					return nil, syscall.EHOSTUNREACH
				}
			case "tcp6":
				if st.Cookie.Family() == syscall.AF_INET6 && st.Cookie.Type() == syscall.SOCK_STREAM {
					return nil, syscall.EHOSTUNREACH
				}
			case "udp6":
				if st.Cookie.Family() == syscall.AF_INET6 && st.Cookie.Type() == syscall.SOCK_DGRAM {
					return nil, syscall.EHOSTUNREACH
				}
			case "ip6":
				if st.Cookie.Family() == syscall.AF_INET6 && st.Cookie.Type() == syscall.SOCK_RAW {
					return nil, syscall.EHOSTUNREACH
				}
			}
			return nil, nil
		})
	})
}
