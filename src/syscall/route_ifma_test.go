// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd

package syscall_test

import (
	"fmt"
	"syscall"
)

func parseRoutingSockaddr(m syscall.RoutingMessage) ([]syscall.Sockaddr, error) {
	switch m := m.(type) {
	case *syscall.RouteMessage:
		sas, err := syscall.ParseRoutingSockaddr(m)
		if err != nil {
			return nil, fmt.Errorf("%T: %v", m, err)
		}
		n := addrFlags(m.Header.Addrs).count()
		if len(sas) != n {
			return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
		}
		return sas, nil
	case *syscall.InterfaceMessage:
		sas, err := syscall.ParseRoutingSockaddr(m)
		if err != nil {
			return nil, fmt.Errorf("%T: %v", m, err)
		}
		n := addrFlags(m.Header.Addrs).count()
		if len(sas) != n {
			return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
		}
		return sas, nil
	case *syscall.InterfaceAddrMessage:
		sas, err := syscall.ParseRoutingSockaddr(m)
		if err != nil {
			return nil, fmt.Errorf("%T: %v", m, err)
		}
		n := addrFlags(m.Header.Addrs).count()
		if len(sas) != n {
			return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
		}
		return sas, nil
	case *syscall.InterfaceMulticastAddrMessage:
		sas, err := syscall.ParseRoutingSockaddr(m)
		if err != nil {
			return nil, fmt.Errorf("%T: %v", m, err)
		}
		n := addrFlags(m.Header.Addrs).count()
		if len(sas) != n {
			return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
		}
		return sas, nil
	default:
		panic(fmt.Sprintf("unknown routing message type: %T", m))
	}
}
