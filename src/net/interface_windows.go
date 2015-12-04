// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/syscall/windows"
	"os"
	"syscall"
	"unsafe"
)

// supportsVistaIP reports whether the platform implements new IP
// stack and ABIs supported on Windows Vista and above.
var supportsVistaIP bool

func init() {
	supportsVistaIP = probeWindowsIPStack()
}

func probeWindowsIPStack() (supportsVistaIPv6 bool) {
	v, err := syscall.GetVersion()
	if err != nil {
		return true // Windows 10 and above will deprecate this API
	}
	if byte(v) < 6 { // major version of Windows Vista is 6
		return false
	}
	return true
}

// adapterAddresses returns a list of IP adapter and address
// structures. The structure contains an IP adapter and flattened
// multiple IP addresses including unicast, anycast and multicast
// addresses.
func adapterAddresses() ([]*windows.IpAdapterAddresses, error) {
	const initSize = 15000 // recommended initial size
	b := make([]byte, initSize)
	l := uint32(initSize)
	var err error
	for i := 0; i < 3; i++ {
		err = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0, &b[0], &l)
		if err == nil {
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		b = make([]byte, l)
	}
	if err != nil {
		return nil, os.NewSyscallError("getadaptersaddresses", err)
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

// If the ifindex is zero, interfaceTable returns mappings of all
// network interfaces.  Otherwise it returns a mapping of a specific
// interface.
func interfaceTable(ifindex int) ([]Interface, error) {
	aas, err := adapterAddresses()
	if err != nil {
		return nil, err
	}
	var ift []Interface
	for _, aa := range aas {
		index := aa.IfIndex
		if index == 0 { // ipv6IfIndex is a sustitute for ifIndex on some kernel versions
			index = aa.Ipv6IfIndex
		}
		if ifindex == 0 || ifindex == int(index) {
			ifi := Interface{
				Index: int(index),
				Name:  syscall.UTF16ToString((*(*[10000]uint16)(unsafe.Pointer(aa.FriendlyName)))[:]),
			}
			if aa.OperStatus == windows.IfOperStatusUp {
				ifi.Flags |= FlagUp
			}
			// For now we need to infer link-layer service
			// capabilities from media types.
			// We will be able to use
			// MIB_IF_ROW2.AccessType once we drop support
			// for Windows XP.
			switch aa.IfType {
			case windows.IF_TYPE_ETHERNET_CSMACD, windows.IF_TYPE_ISO88025_TOKENRING, windows.IF_TYPE_IEEE80211, windows.IF_TYPE_IEEE1394:
				ifi.Flags |= FlagBroadcast | FlagMulticast
			case windows.IF_TYPE_PPP, windows.IF_TYPE_TUNNEL:
				ifi.Flags |= FlagPointToPoint | FlagMulticast
			case windows.IF_TYPE_SOFTWARE_LOOPBACK:
				ifi.Flags |= FlagLoopback | FlagMulticast
			case windows.IF_TYPE_ATM:
				ifi.Flags |= FlagBroadcast |
					FlagPointToPoint |
					FlagMulticast // assume  all services available; LANE, point-to-point and point-to-multipoint
			}
			if aa.Mtu == 0xffffffff {
				ifi.MTU = -1
			} else {
				ifi.MTU = int(aa.Mtu)
			}
			if aa.PhysicalAddressLength > 0 {
				ifi.HardwareAddr = make(HardwareAddr, aa.PhysicalAddressLength)
				copy(ifi.HardwareAddr, aa.PhysicalAddress[:])
			}
			ift = append(ift, ifi)
			if ifindex == int(ifi.Index) {
				break
			}
		}
	}
	return ift, nil
}

// If the ifi is nil, interfaceAddrTable returns addresses for all
// network interfaces.  Otherwise it returns addresses for a specific
// interface.
func interfaceAddrTable(ifi *Interface) ([]Addr, error) {
	aas, err := adapterAddresses()
	if err != nil {
		return nil, err
	}
	var ifat []Addr
	for _, aa := range aas {
		index := aa.IfIndex
		if index == 0 { // ipv6IfIndex is a sustitute for ifIndex on some kernel versions
			index = aa.Ipv6IfIndex
		}
		if ifi == nil || ifi.Index == int(index) {
			puni := aa.FirstUnicastAddress
			for ; puni != nil; puni = puni.Next {
				if sa, err := puni.Address.Sockaddr.Sockaddr(); err == nil {
					var ifa Addr
					switch sa := sa.(type) {
					case *syscall.SockaddrInet4:
						if supportsVistaIP {
							ipn := &IPNet{IP: make(IP, IPv4len), Mask: CIDRMask(int(puni.OnLinkPrefixLength), 8*IPv4len)}
							copy(ipn.IP, sa.Addr[:])
							ifa = ipn
						} else {
							ipa := &IPAddr{IP: make(IP, IPv4len)}
							copy(ipa.IP, sa.Addr[:])
							ifa = ipa
						}
						ifat = append(ifat, ifa)
					case *syscall.SockaddrInet6:
						if supportsVistaIP {
							ipn := &IPNet{IP: make(IP, IPv6len), Mask: CIDRMask(int(puni.OnLinkPrefixLength), 8*IPv6len)}
							copy(ipn.IP, sa.Addr[:])
							ifa = ipn
						} else {
							ipa := &IPAddr{IP: make(IP, IPv6len)}
							copy(ipa.IP, sa.Addr[:])
							ifa = ipa
						}
						ifat = append(ifat, ifa)
					}
				}
			}
			pany := aa.FirstAnycastAddress
			for ; pany != nil; pany = pany.Next {
				if sa, err := pany.Address.Sockaddr.Sockaddr(); err == nil {
					switch sa := sa.(type) {
					case *syscall.SockaddrInet4:
						ifa := &IPAddr{IP: make(IP, IPv4len)}
						copy(ifa.IP, sa.Addr[:])
						ifat = append(ifat, ifa)
					case *syscall.SockaddrInet6:
						ifa := &IPAddr{IP: make(IP, IPv6len)}
						copy(ifa.IP, sa.Addr[:])
						ifat = append(ifat, ifa)
					}
				}
			}
		}
	}
	return ifat, nil
}

// interfaceMulticastAddrTable returns addresses for a specific
// interface.
func interfaceMulticastAddrTable(ifi *Interface) ([]Addr, error) {
	aas, err := adapterAddresses()
	if err != nil {
		return nil, err
	}
	var ifat []Addr
	for _, aa := range aas {
		index := aa.IfIndex
		if index == 0 { // ipv6IfIndex is a sustitute for ifIndex on some kernel versions
			index = aa.Ipv6IfIndex
		}
		if ifi == nil || ifi.Index == int(index) {
			pmul := aa.FirstMulticastAddress
			for ; pmul != nil; pmul = pmul.Next {
				if sa, err := pmul.Address.Sockaddr.Sockaddr(); err == nil {
					switch sa := sa.(type) {
					case *syscall.SockaddrInet4:
						ifa := &IPAddr{IP: make(IP, IPv4len)}
						copy(ifa.IP, sa.Addr[:])
						ifat = append(ifat, ifa)
					case *syscall.SockaddrInet6:
						ifa := &IPAddr{IP: make(IP, IPv6len)}
						copy(ifa.IP, sa.Addr[:])
						ifat = append(ifat, ifa)
					}
				}
			}
		}
	}
	return ifat, nil
}
