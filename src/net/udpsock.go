// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import "syscall"

// UDPAddr represents the address of a UDP end point.
type UDPAddr struct {
	IP   IP
	Port int
	Zone string // IPv6 scoped addressing zone
}

// Network returns the address's network name, "udp".
func (a *UDPAddr) Network() string { return "udp" }

func (a *UDPAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	ip := ipEmptyString(a.IP)
	if a.Zone != "" {
		return JoinHostPort(ip+"%"+a.Zone, itoa(a.Port))
	}
	return JoinHostPort(ip, itoa(a.Port))
}

func (a *UDPAddr) isWildcard() bool {
	if a == nil || a.IP == nil {
		return true
	}
	return a.IP.IsUnspecified()
}

// ResolveUDPAddr parses addr as a UDP address of the form "host:port"
// or "[ipv6-host%zone]:port" and resolves a pair of domain name and
// port name on the network net, which must be "udp", "udp4" or
// "udp6".  A literal address or host name for IPv6 must be enclosed
// in square brackets, as in "[::1]:80", "[ipv6-host]:http" or
// "[ipv6-host%zone]:80".
func ResolveUDPAddr(net, addr string) (*UDPAddr, error) {
	switch net {
	case "udp", "udp4", "udp6":
	case "": // a hint wildcard for Go 1.0 undocumented behavior
		net = "udp"
	default:
		return nil, UnknownNetworkError(net)
	}
	addrs, err := internetAddrList(net, addr, noDeadline)
	if err != nil {
		return nil, err
	}
	return addrs.first(isIPv4).(*UDPAddr), nil
}

// UDPConn is the implementation of the Conn and PacketConn interfaces
// for UDP network connections.
type UDPConn struct {
	conn
}

// ReadFromUDP reads a UDP packet from c, copying the payload into b.
// It returns the number of bytes copied into b and the return address
// that was on the packet.
//
// ReadFromUDP can be made to time out and return an error with
// Timeout() == true after a fixed time limit; see SetDeadline and
// SetReadDeadline.
func (c *UDPConn) ReadFromUDP(b []byte) (int, *UDPAddr, error) {
	if !c.ok() {
		return 0, nil, syscall.EINVAL
	}
	return c.readFromUDP(b)
}

// ReadFrom implements the PacketConn ReadFrom method.
func (c *UDPConn) ReadFrom(b []byte) (int, Addr, error) {
	if !c.ok() {
		return 0, nil, syscall.EINVAL
	}
	n, addr, err := c.readFromUDP(b)
	if addr == nil {
		return n, nil, err
	}
	return n, addr, err
}

// ReadMsgUDP reads a packet from c, copying the payload into b and
// the associated out-of-band data into oob.  It returns the number
// of bytes copied into b, the number of bytes copied into oob, the
// flags that were set on the packet and the source address of the
// packet.
func (c *UDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *UDPAddr, err error) {
	if !c.ok() {
		return 0, 0, 0, nil, syscall.EINVAL
	}
	return c.readMsgUDP(b, oob)
}

// WriteToUDP writes a UDP packet to addr via c, copying the payload
// from b.
//
// WriteToUDP can be made to time out and return an error with
// Timeout() == true after a fixed time limit; see SetDeadline and
// SetWriteDeadline.  On packet-oriented connections, write timeouts
// are rare.
func (c *UDPConn) WriteToUDP(b []byte, addr *UDPAddr) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	return c.writeToUDP(b, addr)
}

// WriteTo implements the PacketConn WriteTo method.
func (c *UDPConn) WriteTo(b []byte, addr Addr) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	a, ok := addr.(*UDPAddr)
	if !ok {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: syscall.EINVAL}
	}
	return c.writeToUDP(b, a)
}

// WriteMsgUDP writes a packet to addr via c if c isn't connected, or
// to c's remote destination address if c is connected (in which case
// addr must be nil).  The payload is copied from b and the associated
// out-of-band data is copied from oob.  It returns the number of
// payload and out-of-band bytes written.
func (c *UDPConn) WriteMsgUDP(b, oob []byte, addr *UDPAddr) (n, oobn int, err error) {
	if !c.ok() {
		return 0, 0, syscall.EINVAL
	}
	return c.writeMsgUDP(b, oob, addr)
}

// DialUDP connects to the remote address raddr on the network net,
// which must be "udp", "udp4", or "udp6".  If laddr is not nil, it is
// used as the local address for the connection.
func DialUDP(net string, laddr, raddr *UDPAddr) (*UDPConn, error) {
	switch net {
	case "udp", "udp4", "udp6":
	default:
		return nil, &OpError{Op: "dial", Net: net, Source: laddr, Addr: raddr, Err: UnknownNetworkError(net)}
	}
	if raddr == nil {
		return nil, &OpError{Op: "dial", Net: net, Source: laddr, Addr: raddr, Err: errMissingAddress}
	}
	return dialUDP(net, laddr, raddr, noDeadline)
}

// ListenUDP listens for incoming UDP packets addressed to the local
// address laddr.  Net must be "udp", "udp4", or "udp6".  If laddr has
// a port of 0, ListenUDP will choose an available port.
// The LocalAddr method of the returned UDPConn can be used to
// discover the port.  The returned connection's ReadFrom and WriteTo
// methods can be used to receive and send UDP packets with per-packet
// addressing.
func ListenUDP(net string, laddr *UDPAddr) (*UDPConn, error) {
	switch net {
	case "udp", "udp4", "udp6":
	default:
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: laddr, Err: UnknownNetworkError(net)}
	}
	if laddr == nil {
		laddr = &UDPAddr{}
	}
	return listenUDP(net, laddr)
}

// ListenMulticastUDP listens for incoming multicast UDP packets
// addressed to the group address gaddr on ifi, which specifies the
// interface to join.  ListenMulticastUDP uses default multicast
// interface if ifi is nil.
func ListenMulticastUDP(net string, ifi *Interface, gaddr *UDPAddr) (*UDPConn, error) {
	switch net {
	case "udp", "udp4", "udp6":
	default:
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: gaddr, Err: UnknownNetworkError(net)}
	}
	if gaddr == nil || gaddr.IP == nil {
		return nil, &OpError{Op: "listen", Net: net, Source: nil, Addr: gaddr, Err: errMissingAddress}
	}
	return listenMulticastUDP(net, ifi, gaddr)
}
