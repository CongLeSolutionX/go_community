// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !netgo
// +build darwin dragonfly freebsd linux netbsd openbsd

package net

import (
	intsocket "net/internal/socket"
	"syscall"
)

func cgoLookupHost(name string) ([]string, error, bool) {
	sas, err := intsocket.GetaddrinfoAddr(name)
	if err != nil {
		return nil, err, true
	}
	var lits []string
	for _, sa := range sas {
		switch sa := sa.(type) {
		case *syscall.SockaddrInet4:
			lits = append(lits, IP(sa.Addr[:]).String())
		case *syscall.SockaddrInet6:
			lits = append(lits, IP(sa.Addr[:]).String())
		}
	}
	return lits, nil, true
}

func cgoLookupPort(network, service string) (int, error, bool) {
	acquireThread()
	defer releaseThread()

	port, err := intsocket.GetaddrinfoPort(network, service)
	if err != nil {
		return 0, err, true
	}
	return port, nil, true
}

func cgoLookupIP(name string) ([]IPAddr, error, bool) {
	acquireThread()
	defer releaseThread()

	sas, err := intsocket.GetaddrinfoAddr(name)
	if err != nil {
		return nil, err, true
	}
	var addrs []IPAddr
	for _, sa := range sas {
		switch sa := sa.(type) {
		case *syscall.SockaddrInet4:
			addrs = append(addrs, IPAddr{IP: copyIP(sa.Addr[:])})
		case *syscall.SockaddrInet6:
			addrs = append(addrs, IPAddr{IP: copyIP(sa.Addr[:]), Zone: zoneToString(int(sa.ZoneId))})
		}
	}
	return addrs, nil, true
}

func cgoLookupCNAME(name string) (string, error, bool) {
	acquireThread()
	defer releaseThread()

	cname, err := intsocket.GetaddrinfoCNAME(name)
	if err != nil {
		return "", err, true
	}
	return cname, nil, true
}

func cgoLookupPTRName(addr string) ([]string, error, bool) {
	acquireThread()
	defer releaseThread()

	ip := ParseIP(addr)
	if ip == nil {
		return nil, &AddrError{Err: "non-IP address", Addr: addr}, false
	}
	ptr, err := intsocket.GetnameinfoPTR(ip)
	if err != nil {
		return nil, err, true
	}
	return []string{ptr}, nil, true
}

func copyIP(x []byte) IP {
	if len(x) < 16 {
		return IP(x).To16()
	}
	y := make(IP, len(x))
	copy(y, x)
	return y
}
