package proxy

import "net"

// SOCKS5Connect takes an existing connection to a socks5 proxy server at addr,
// and commands the server to extend that connection to target,
// which must be a canonical address with a host and port.
func SOCKS5Connect(conn net.Conn, addr string, auth *Auth, target string) error {
	s, err := SOCKS5("", addr, auth, nil)
	if err != nil {
		return err
	}
	return s.(*socks5).connect(conn, target)
}
