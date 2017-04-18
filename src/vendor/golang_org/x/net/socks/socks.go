// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package socks provides a SOCKS version 5 client implementation.
//
// SOCKS protocol version 5 is defined in RFC 1928.
// Username/Password authentication for SOCKS version 5 is defined in
// RFC 1929.
package socks

// This package is supposed to be used by the net/http package of
// standard library and proxy package of golang.org/x/net repository.

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
)

// A Command represents a SOCKS command.
type Command int

func (cmd Command) String() string {
	switch cmd {
	case CmdConnect:
		return "socks connect"
	case CmdBind:
		return "socks bind"
	default:
		return "socks " + strconv.Itoa(int(cmd))
	}
}

// An AuthMethod represents a SOCKS authentication method.
type AuthMethod int

const (
	protocolVersion5 = 0x05

	CmdConnect Command = 0x01 // establishes an active-open forward proxy connection
	CmdBind    Command = 0x02 // establishes a passive-open forward proxy connection

	addrTypeIPv4 = 0x01
	addrTypeFQDN = 0x03
	addrTypeIPv6 = 0x04

	AuthMethodNotRequired         AuthMethod = 0x00 // no authentication required
	AuthMethodUsernamePassword    AuthMethod = 0x02 // use username/password
	AuthMethodNoAcceptableMethods AuthMethod = 0xff // no acceptable authetication methods

	statusSucceeded = 0x00
)

var cmdErrors = [...]string{
	"succeeded",
	"general SOCKS server failure",
	"connection not allowed by ruleset",
	"network unreachable",
	"host unreachable",
	"connection refused",
	"TTL expired",
	"command not supported",
	"address type not supported",
}

// An Addr represents a SOCKS-specific address.
// Either Name or IP is used exclusively.
type Addr struct {
	Name string // fully-qualified domain name
	IP   net.IP
	Port int
}

func (a *Addr) Network() string { return "socks" }

func (a *Addr) String() string {
	if a == nil {
		return "<nil>"
	}
	port := strconv.Itoa(a.Port)
	if a.IP == nil {
		return net.JoinHostPort(a.Name, port)
	}
	return net.JoinHostPort(a.IP.String(), port)
}

// A Conn represents a forward proxy connection.
type Conn struct {
	net.Conn

	boundAddr net.Addr
}

// BoundAddr returns the server bound address defined in RFC 1928.
func (c *Conn) BoundAddr() net.Addr {
	if c == nil {
		return nil
	}
	return c.boundAddr
}

// A Dialer holds SOCKS-specific options.
type Dialer struct {
	cmd          Command // either CmdConnect or CmdBind
	proxyNetwork string  // network between a proxy server and a client
	proxyAddress string  // proxy server address

	// ProxyDial specifies the optional dial function for
	// establishing the transport connection.
	ProxyDial func(context.Context, string, string) (net.Conn, error)

	// AuthMethods specifies the list of request authention
	// methods.
	// If empty, SOCKS client requests only AuthMethodNotRequired.
	AuthMethods []AuthMethod

	// Authenticate specifies the optional authentication
	// function. It must be non-nil when AuthMethods is not empty.
	// It's the authentication function's responsibility to handle
	// the given context except IO on the given io.ReadWriter.
	Authenticate func(context.Context, io.ReadWriter, AuthMethod) error
}

// DialContext connects to the provided address on the provided
// network.
//
// When d is for CmdConnect, address must specify the pair of address
// and port number of the final destination.
// When d is for CmdBind, address must specify the pair of address and
// port number on the proxy server.
//
// The returned error value may be a net.OpError. When the Op field of
// net.OpError contains "socks", the Source field contains a proxy
// server address and the Addr field contains a command target
// address.
//
// See func Dial of the net package of standard library for a
// description of the network and address parameters.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		proxy, dst, _ := d.pathAddrs(network, address)
		return nil, &net.OpError{Op: d.cmd.String(), Net: network, Source: proxy, Addr: dst, Err: errors.New("network not implemented")}
	}
	switch d.cmd {
	case CmdConnect, CmdBind:
	default:
		proxy, dst, _ := d.pathAddrs(network, address)
		return nil, &net.OpError{Op: d.cmd.String(), Net: network, Source: proxy, Addr: dst, Err: errors.New("command not implemented")}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	var c net.Conn
	if d.ProxyDial != nil {
		c, err = d.ProxyDial(ctx, d.proxyNetwork, d.proxyAddress)
	} else {
		var dd net.Dialer
		c, err = dd.DialContext(ctx, d.proxyNetwork, d.proxyAddress)
	}
	if err != nil {
		proxy, dst, _ := d.pathAddrs(network, address)
		return nil, &net.OpError{Op: d.cmd.String(), Net: network, Source: proxy, Addr: dst, Err: err}
	}
	a, err := d.connect(ctx, c, address)
	if err != nil {
		c.Close()
		proxy, dst, _ := d.pathAddrs(network, address)
		return nil, &net.OpError{Op: d.cmd.String(), Net: network, Source: proxy, Addr: dst, Err: err}
	}
	return &Conn{Conn: c, boundAddr: a}, nil
}

// Dial connects to the provided address on the provided network.
//
// Deprecated: Use DialContext instead.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *Dialer) pathAddrs(network, address string) (proxy, dst net.Addr, err error) {
	proxy, err = net.ResolveTCPAddr(d.proxyNetwork, d.proxyAddress)
	if err != nil {
		return nil, nil, err
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, nil, err
	}
	portnum, err := strconv.Atoi(port)
	if err != nil {
		return nil, nil, err
	}
	a := &Addr{Port: portnum}
	// The final destination address can be unresolvable on the
	// session initiator.
	a.IP = net.ParseIP(host)
	if a.IP == nil {
		a.Name = host
	}
	return proxy, a, nil
}

// NewDialer returns a new Dialer.
//
// The provided network and address must specify a proxy server.
func NewDialer(network, address string, cmd Command) (*Dialer, error) {
	return &Dialer{proxyNetwork: network, proxyAddress: address, cmd: cmd}, nil
}

const (
	authUsernamePasswordVersion = 0x01
	authStatusSucceeded         = 0x00
)

// A UsernamePassword holds information for username/password
// authentication method.
type UsernamePassword struct {
	Username string
	Password string
}

// Authenticate authenticates a pair of username and password with the
// proxy server.
func (up *UsernamePassword) Authenticate(ctx context.Context, rw io.ReadWriter, auth AuthMethod) error {
	switch auth {
	case AuthMethodNotRequired:
		return nil
	case AuthMethodUsernamePassword:
		if len(up.Username) == 0 || len(up.Username) > 255 || len(up.Password) == 0 || len(up.Password) > 255 {
			return errors.New("invalid username/password")
		}
		b := []byte{authUsernamePasswordVersion}
		b = append(b, byte(len(up.Username)))
		b = append(b, up.Username...)
		b = append(b, byte(len(up.Password)))
		b = append(b, up.Password...)
		if _, err := rw.Write(b); err != nil {
			return err
		}
		if _, err := io.ReadFull(rw, b[:2]); err != nil {
			return err
		}
		if b[0] != authUsernamePasswordVersion {
			return errors.New("invalid username/password version")
		}
		if b[1] != authStatusSucceeded {
			return errors.New("username/password authentication failed")
		}
		return nil
	default:
		return errors.New("unsupported authentication method " + strconv.Itoa(int(auth)))
	}
}
