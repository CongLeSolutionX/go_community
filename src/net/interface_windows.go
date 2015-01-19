// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/syscall/windows"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// If the ifindex is zero, interfaceTable returns mappings of all
// network interfaces.  Otherwise it returns a mapping of a specific
// interface.
func interfaceTable(ifindex int) ([]Interface, error) {
	var (
		err error
		ift []Interface
	)

	var size uint32
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, nil, &size)
	if err != nil && err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	c := size / uint32(unsafe.Sizeof(windows.IpAdapterAddresses{}))
	addrs := make([]windows.IpAdapterAddresses, c)
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, &addrs[0], &size)
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	var ptable *windows.IpInterfaceNameInfo
	err = windows.NhpAllocateAndGetInterfaceInfoFromStack(&ptable, &size, true, windows.GetProcessHeap(), 0)
	if err != nil {
		return nil, os.NewSyscallError("NhpAllocateAndGetInterfaceInfoFromStack", err)
	}
	paddr := &addrs[0]
	for paddr != nil {
		index := paddr.IfIndex
		if paddr.Ipv6IfIndex != 0 {
			index = paddr.Ipv6IfIndex
		}
		if ifindex == 0 || ifindex == int(index) {
			var flags Flags
			if paddr.Flags&windows.IfOperStatusUp != 0 {
				flags |= windows.IFF_UP
			}
			if paddr.IfType&windows.IF_TYPE_SOFTWARE_LOOPBACK != 0 {
				flags |= windows.IFF_LOOPBACK
			}
			tables := (*[100]windows.IpInterfaceNameInfo)(unsafe.Pointer(ptable))
			for n := 0; n < int(size); n++ {
				if index == tables[n].Index {
					if tables[n].AccessType&windows.IF_ACCESS_BROADCAST != 0 {
						flags |= windows.IFF_BROADCAST
					}
					if tables[n].AccessType&windows.IF_ACCESS_POINT_TO_POINT != 0 {
						flags |= windows.IFF_POINTOPOINT
					}
					if tables[n].AccessType&windows.IF_ACCESS_POINT_TO_MULTI_POINT != 0 {
						flags |= windows.IFF_MULTICAST
					}
				}
			}
			ifi := Interface{
				Index:        int(index),
				MTU:          int(paddr.Mtu),
				Name:         utf16PtrToString(paddr.FriendlyName),
				HardwareAddr: HardwareAddr(paddr.PhysicalAddress[:]),
				Flags:        flags,
			}
			ift = append(ift, ifi)
		}
		paddr = (*windows.IpAdapterAddresses)(unsafe.Pointer(paddr.Next))
	}

	return ift, nil
}

// If the ifi is nil, interfaceAddrTable returns addresses for all
// network interfaces.  Otherwise it returns addresses for a specific
// interface.
func interfaceAddrTable(ifi *Interface) ([]Addr, error) {
	var (
		err  error
		ifat []Addr
	)

	var size uint32
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, nil, &size)
	if err != nil && err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	c := size / uint32(unsafe.Sizeof(windows.IpAdapterAddresses{}))
	addrs := make([]windows.IpAdapterAddresses, c)
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, &addrs[0], &size)
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	paddr := &addrs[0]
	for paddr != nil {
		index := paddr.IfIndex
		if paddr.Ipv6IfIndex != 0 {
			index = paddr.Ipv6IfIndex
		}
		if ifi == nil || ifi.Index == int(index) {
			puni := paddr.FirstUnicastAddress
			for puni != nil {
				if puni.Flags&windows.IP_ADAPTER_ADDRESS_DNS_ELIGIBLE != 0 &&
					puni.Flags&windows.IP_ADAPTER_ADDRESS_TRANSIENT == 0 {
					if sa, err := puni.Address.Sockaddr.Sockaddr(); err == nil {
						switch sav := sa.(type) {
						case *syscall.SockaddrInet4:
							ifa := &IPNet{IP: make(IP, IPv4len), Mask: CIDRMask(int(puni.OnLinkPrefixLength), 8*IPv4len)}
							copy(ifa.IP, sav.Addr[:])
							ifat = append(ifat, ifa)
						case *syscall.SockaddrInet6:
							ifa := &IPNet{IP: make(IP, IPv6len), Mask: CIDRMask(int(puni.OnLinkPrefixLength), 8*IPv6len)}
							copy(ifa.IP, sav.Addr[:])
							ifat = append(ifat, ifa)
						}
					}
				}
				puni = (*windows.IpAdapterUnicastAddress)(unsafe.Pointer(puni.Next))
			}
		}
		paddr = (*windows.IpAdapterAddresses)(unsafe.Pointer(paddr.Next))
	}

	return ifat, nil
}

// interfaceMulticastAddrTable returns addresses for a specific
// interface.
func interfaceMulticastAddrTable(ifi *Interface) ([]Addr, error) {
	var (
		err  error
		ifat []Addr
	)

	var size uint32
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, nil, &size)
	if err != nil && err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	c := size / uint32(unsafe.Sizeof(windows.IpAdapterAddresses{}))
	addrs := make([]windows.IpAdapterAddresses, c)
	err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, &addrs[0], &size)
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersAddresses", err)
	}

	paddr := &addrs[0]
	for paddr != nil {
		index := paddr.IfIndex
		if paddr.Ipv6IfIndex != 0 {
			index = paddr.Ipv6IfIndex
		}
		if ifi == nil || ifi.Index == int(index) {
			pmul := paddr.FirstMulticastAddress
			for pmul != nil {
				if pmul.Flags&windows.IP_ADAPTER_ADDRESS_DNS_ELIGIBLE != 0 &&
					pmul.Flags&windows.IP_ADAPTER_ADDRESS_TRANSIENT == 0 {
					if sa, err := pmul.Address.Sockaddr.Sockaddr(); err == nil {
						switch sav := sa.(type) {
						case *syscall.SockaddrInet4:
							ifa := &IPAddr{IP: make(IP, IPv4len)}
							copy(ifa.IP, sav.Addr[:])
							ifat = append(ifat, ifa.toAddr())
						case *syscall.SockaddrInet6:
							ifa := &IPAddr{IP: make(IP, IPv6len)}
							copy(ifa.IP, sav.Addr[:])
							ifat = append(ifat, ifa.toAddr())
						}
					}
				}
				pmul = (*windows.IpAdapterMulticastAddress)(unsafe.Pointer(pmul.Next))
			}
		}
		paddr = (*windows.IpAdapterAddresses)(unsafe.Pointer(paddr.Next))
	}

	return ifat, nil
}

func utf16PtrToString(p *uint16) string {
	a := (*[10000]uint16)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(utf16.Decode(a[:i]))
}
