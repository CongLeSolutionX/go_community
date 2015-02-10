// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd netbsd openbsd

package syscall_test

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRouteRIB(t *testing.T) {
	for _, facility := range []int{syscall.NET_RT_DUMP, syscall.NET_RT_IFLIST} {
		for _, param := range []int{syscall.AF_UNSPEC, syscall.AF_INET, syscall.AF_INET6} {
			b, err := syscall.RouteRIB(facility, param)
			if err != nil {
				t.Error(facility, param, err)
				continue
			}
			msgs, err := syscall.ParseRoutingMessage(b)
			if err != nil {
				t.Error(facility, param, err)
				continue
			}
			if _, err := parseRoutingSockaddr(msgs); err != nil {
				t.Error(facility, param, err)
				continue
			}
		}
	}
}

func TestRouteMonitor(t *testing.T) {
	if testing.Short() || os.Getuid() != 0 {
		t.Skip("must be root")
	}

	s, err := syscall.Socket(syscall.AF_ROUTE, syscall.SOCK_RAW, syscall.AF_UNSPEC)
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Close(s)

	tmo := time.After(30 * time.Second)
	go func() {
		b := make([]byte, os.Getpagesize())
		for {
			n, err := syscall.Read(s, b)
			if err != nil {
				return
			}
			msgs, err := syscall.ParseRoutingMessage(b[:n])
			if err != nil {
				t.Error(err)
				return
			}
			sas, err := parseRoutingSockaddr(msgs)
			if err != nil {
				t.Error(err)
				return
			}
			t.Log(sockaddrs(sas))
		}
	}()
	<-tmo
}

func parseRoutingSockaddr(msgs []syscall.RoutingMessage) ([]syscall.Sockaddr, error) {
	var sss []syscall.Sockaddr
	for _, m := range msgs {
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
			sss = append(sss, sas...)
		case *syscall.InterfaceMessage:
			sas, err := syscall.ParseRoutingSockaddr(m)
			if err != nil {
				return nil, fmt.Errorf("%T: %v", m, err)
			}
			n := addrFlags(m.Header.Addrs).count()
			if len(sas) != n {
				return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
			}
			sss = append(sss, sas...)
		case *syscall.InterfaceAddrMessage:
			sas, err := syscall.ParseRoutingSockaddr(m)
			if err != nil {
				return nil, fmt.Errorf("%T: %v", m, err)
			}
			n := addrFlags(m.Header.Addrs).count()
			if len(sas) != n {
				return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
			}
			sss = append(sss, sas...)
		case *syscall.InterfaceMulticastAddrMessage:
			sas, err := syscall.ParseRoutingSockaddr(m)
			if err != nil {
				return nil, fmt.Errorf("%T: %v", m, err)
			}
			n := addrFlags(m.Header.Addrs).count()
			if len(sas) != n {
				return nil, fmt.Errorf("%T: got %v; want %v, %v", m, len(sas), n, addrFlags(m.Header.Addrs))
			}
			sss = append(sss, sas...)
		}
	}
	return sss, nil
}

type addrFamily byte

func (f addrFamily) String() string {
	switch f {
	case syscall.AF_UNSPEC:
		return "unspec"
	case syscall.AF_LINK:
		return "link"
	case syscall.AF_INET:
		return "inet4"
	case syscall.AF_INET6:
		return "inet6"
	default:
		return fmt.Sprintf("unknown %d", f)
	}
}

type addrFlags uint32

var addrFlagNames = [...]string{
	"dst",
	"gateway",
	"netmask",
	"genmask",
	"ifp",
	"ifa",
	"author",
	"brd",
	"mpls1 or tag or src", // dragonfly, netbsd, openbsd
	"mpls2 or srcmask",    // dragonfly, openbsd
	"mpls3 or label",      // dragonfly, openbsd
}

func (f addrFlags) count() int {
	var n int
	for i := 0; i < syscall.RTAX_MAX; i++ {
		if f&(1<<uint(i)) != 0 {
			n++
		}
	}
	return n
}

func (f addrFlags) String() string {
	var s string
	for i, name := range addrFlagNames {
		if f&(1<<uint(i)) != 0 {
			if s != "" {
				s += "|"
			}
			s += name
		}
	}
	if s == "" {
		return "<nil>"
	}
	return s
}

type sockaddrs []syscall.Sockaddr

func (sas sockaddrs) String() string {
	var s string
	for _, sa := range sas {
		switch sa := sa.(type) {
		case *syscall.SockaddrDatalink:
			s += fmt.Sprintf("[%v/%v/%v, t/n/a/s=%v/%v/%v/%v] ", sa.Len, addrFamily(sa.Family), sa.Index, sa.Type, sa.Nlen, sa.Alen, sa.Slen)
		case *syscall.SockaddrInet4:
			s += fmt.Sprintf("%v ", net.IP(sa.Addr[:]))
		case *syscall.SockaddrInet6:
			s += fmt.Sprintf("%v ", net.IP(sa.Addr[:]))
		default:
			panic("unknown sockaddr type")
		}
	}
	if s == "" {
		return "<nil>"
	}
	return s
}
