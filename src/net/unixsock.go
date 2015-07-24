// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
	"syscall"
	"time"
)

// UnixAddr represents the address of a Unix domain socket end point.
type UnixAddr struct {
	Name string
	Net  string
}

// Network returns the address's network name, "unix", "unixgram" or
// "unixpacket".
func (a *UnixAddr) Network() string {
	return a.Net
}

func (a *UnixAddr) String() string {
	if a == nil {
		return "<nil>"
	}
	return a.Name
}

func (a *UnixAddr) isWildcard() bool {
	return a == nil || a.Name == ""
}

func (a *UnixAddr) opAddr() Addr {
	if a == nil {
		return nil
	}
	return a
}

// ResolveUnixAddr parses addr as a Unix domain socket address.
// The string network gives the network name, "unix", "unixgram" or
// "unixpacket".
func ResolveUnixAddr(network, addr string) (*UnixAddr, error) {
	switch network {
	case "unix", "unixgram", "unixpacket":
		return &UnixAddr{Name: addr, Net: network}, nil
	default:
		return nil, UnknownNetworkError(network)
	}
}

// UnixConn is an implementation of the Conn interface for connections
// to Unix domain sockets.
type UnixConn struct {
	conn
}

// CloseRead shuts down the reading side of the Unix domain connection.
// Most callers should just use Close.
func (c *UnixConn) CloseRead() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	if err := c.fd.closeRead(); err != nil {
		return &OpError{Op: "close", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return nil
}

// CloseWrite shuts down the writing side of the Unix domain connection.
// Most callers should just use Close.
func (c *UnixConn) CloseWrite() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	if err := c.fd.closeWrite(); err != nil {
		return &OpError{Op: "close", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
	}
	return nil
}

// ReadFromUnix reads a packet from c, copying the payload into b.
// It returns the number of bytes copied into b and the source address
// of the packet.
//
// ReadFromUnix can be made to time out and return an error with
// Timeout() == true after a fixed time limit; see SetDeadline and
// SetReadDeadline.
func (c *UnixConn) ReadFromUnix(b []byte) (int, *UnixAddr, error) {
	if !c.ok() {
		return 0, nil, syscall.EINVAL
	}
	return c.readFromUnix(b)
}

// ReadFrom implements the PacketConn ReadFrom method.
func (c *UnixConn) ReadFrom(b []byte) (int, Addr, error) {
	if !c.ok() {
		return 0, nil, syscall.EINVAL
	}
	n, addr, err := c.readFromUnix(b)
	if addr == nil {
		return n, nil, err
	}
	return n, addr, err
}

// ReadMsgUnix reads a packet from c, copying the payload into b and
// the associated out-of-band data into oob.
// It returns the number of bytes copied into b, the number of bytes
// copied into oob, the flags that were set on the packet, and the
// source address of the packet.
func (c *UnixConn) ReadMsgUnix(b, oob []byte) (n, oobn, flags int, addr *UnixAddr, err error) {
	if !c.ok() {
		return 0, 0, 0, nil, syscall.EINVAL
	}
	return c.readMsgUnix(b, oob)
}

// WriteToUnix writes a packet to addr via c, copying the payload from b.
//
// WriteToUnix can be made to time out and return an error with
// Timeout() == true after a fixed time limit; see SetDeadline and
// SetWriteDeadline.  On packet-oriented connections, write timeouts
// are rare.
func (c *UnixConn) WriteToUnix(b []byte, addr *UnixAddr) (int, error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	return c.writeToUnix(b, addr)
}

// WriteTo implements the PacketConn WriteTo method.
func (c *UnixConn) WriteTo(b []byte, addr Addr) (n int, err error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	a, ok := addr.(*UnixAddr)
	if !ok {
		return 0, &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: addr, Err: syscall.EINVAL}
	}
	return c.writeToUnix(b, a)
}

// WriteMsgUnix writes a packet to addr via c, copying the payload
// from b and the associated out-of-band data from oob.
// It returns the number of payload and out-of-band bytes written.
func (c *UnixConn) WriteMsgUnix(b, oob []byte, addr *UnixAddr) (n, oobn int, err error) {
	if !c.ok() {
		return 0, 0, syscall.EINVAL
	}
	return c.writeMsgUnix(b, oob, addr)
}

// DialUnix connects to the remote address raddr on network, which
// must be "unix", "unixgram" or "unixpacket".
// If laddr is not nil, it is used as the local address for the
// connection.
func DialUnix(network string, laddr, raddr *UnixAddr) (*UnixConn, error) {
	switch network {
	case "unix", "unixgram", "unixpacket":
	default:
		return nil, &OpError{Op: "dial", Net: network, Source: laddr.opAddr(), Addr: raddr.opAddr(), Err: UnknownNetworkError(network)}
	}
	return dialUnix(network, laddr, raddr, noDeadline)
}

// UnixListener is a Unix domain socket listener.
// Clients should typically use variables of type Listener instead of
// assuming Unix domain sockets.
type UnixListener struct {
	fd   *netFD
	path string
}

// AcceptUnix accepts the next incoming call and returns the new
// connection.
func (l *UnixListener) AcceptUnix() (*UnixConn, error) {
	if !l.ok() {
		return nil, syscall.EINVAL
	}
	return l.acceptUnix()
}

// Accept implements the Accept method in the Listener interface; it
// waits for the next call and returns a generic Conn.
func (l *UnixListener) Accept() (Conn, error) {
	if !l.ok() {
		return nil, syscall.EINVAL
	}
	c, err := l.acceptUnix()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Close stops listening on the Unix address.
// Already accepted connections are not closed.
func (l *UnixListener) Close() error {
	if !l.ok() {
		return syscall.EINVAL
	}
	return l.close()
}

// Addr returns the listener's network address.
// The Addr returned is shared by all invocations of Addr, so
// do not modify it.
func (l *UnixListener) Addr() Addr { return l.fd.laddr }

// SetDeadline sets the deadline associated with the listener.
// A zero time value disables the deadline.
func (l *UnixListener) SetDeadline(t time.Time) error {
	if !l.ok() {
		return syscall.EINVAL
	}
	if err := l.fd.setDeadline(t); err != nil {
		return &OpError{Op: "set", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
	}
	return nil
}

// File returns a copy of the underlying os.File, set to blocking
// mode.  It is the caller's responsibility to close f when finished.
// Closing l does not affect f, and closing f does not affect l.
//
// The returned os.File's file descriptor is different from the
// connection's.  Attempting to change properties of the original
// using this duplicate may or may not have the desired effect.
func (l *UnixListener) File() (f *os.File, err error) {
	if !l.ok() {
		return nil, syscall.EINVAL
	}
	return l.file()
}

// ListenUnix announces on the Unix domain socket laddr and returns a
// Unix listener.
// The string network must be "unix" or "unixpacket".
func ListenUnix(network string, laddr *UnixAddr) (*UnixListener, error) {
	switch network {
	case "unix", "unixpacket":
	default:
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: UnknownNetworkError(network)}
	}
	if laddr == nil {
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: errMissingAddress}
	}
	return listenUnix(network, laddr)
}

// ListenUnixgram listens for incoming Unix datagram packets addressed
// to the local address laddr.
// The string network must be "unixgram".
// The returned connection's ReadFrom and WriteTo methods can be used
// to receive and send packets with per-packet addressing.
func ListenUnixgram(network string, laddr *UnixAddr) (*UnixConn, error) {
	switch network {
	case "unixgram":
	default:
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: UnknownNetworkError(network)}
	}
	if laddr == nil {
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: laddr.opAddr(), Err: errMissingAddress}
	}
	return listenUnixgram(network, laddr)
}
