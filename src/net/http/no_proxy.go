// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP client implementation. See RFC 2616.
//
// This is the low-level Transport implementation of RoundTripper.
// The high-level interface is in client.go.

package http

import (
	"net"
	"os"
	"strings"
	"sync"
)

var (
	httpProxyEnv = &envOnce{
		names: []string{"HTTP_PROXY", "http_proxy"},
	}
	httpsProxyEnv = &envOnce{
		names: []string{"HTTPS_PROXY", "https_proxy"},
	}
	noProxyEnv = &envNoProxyOnce{
		envOnce: envOnce{names: []string{"NO_PROXY", "no_proxy"}},
	}
)

type ignoreProxy interface {
	ignore(host, port string, ip net.IP) bool
}

type cidrIgnore struct {
	cidr *net.IPNet
}

func (i *cidrIgnore) ignore(host, port string, ip net.IP) bool {
	return i.cidr.Contains(ip)
}

type ipIgnore struct {
	ip   net.IP
	port string
}

func (i *ipIgnore) ignore(host, port string, ip net.IP) bool {
	if i.ip.Equal(ip) {
		return i.port == "" || i.port == port
	}
	return false
}

type domainIgnore struct {
	host string
	port string
}

func (i *domainIgnore) ignore(host, port string, ip net.IP) bool {
	if strings.HasSuffix(host, i.host) || host == i.host[1:] {
		return i.port == "" || i.port == port
	}
	return false
}

// envOnce looks up an environment variable (optionally by multiple
// names) once. It mitigates expensive lookups on some platforms
// (e.g. Windows).
type envOnce struct {
	names []string
	once  sync.Once
	val   string
}

func (e *envOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envOnce) init() {
	for _, n := range e.names {
		e.val = os.Getenv(n)
		if e.val != "" {
			return
		}
	}
}

// reset is used by tests
func (e *envOnce) reset() {
	e.once = sync.Once{}
	e.val = ""
}

type envNoProxyOnce struct {
	envOnce
	ipIgnores     []ignoreProxy
	domainIgnores []ignoreProxy
}

func (e *envNoProxyOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envNoProxyOnce) init() {
	e.envOnce.init()

	for _, p := range strings.Split(e.val, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}

		if p == "*" {
			e.val = "*"
			e.ipIgnores = nil
			e.domainIgnores = nil
			return
		}

		// IPv4/CIDR, IPv6/CIDR
		if _, pnet, err := net.ParseCIDR(p); err == nil {
			e.ipIgnores = append(e.ipIgnores, &cidrIgnore{cidr: pnet})
			continue
		}

		// IPv4:port, [IPv6]:port
		phost, pport, err := net.SplitHostPort(p)
		if err == nil {
			if len(phost) == 0 {
				// There is no host part, likely the entry is malformed; ignore.
				continue
			}
			if phost[0] == '[' && phost[len(phost)-1] == ']' {
				phost = phost[1 : len(phost)-1]
			}
		} else {
			phost = p
		}
		// IPv4, IPv6
		if pip := net.ParseIP(phost); pip != nil {
			e.ipIgnores = append(e.ipIgnores, &ipIgnore{ip: pip, port: pport})
			continue
		}

		phost = p
		pport = ""
		if i := strings.IndexByte(p, ':'); i != -1 {
			phost = p[:i]
			pport = p[i+1:]
		}
		if len(phost) == 0 {
			// There is no host part, likely the entry is malformed; ignore.
			continue
		}

		// domain.com or domain.com:80
		// foo.com matches bar.foo.com
		// .domain.com or .domain.com:port
		// *.domain.com or *.domain.com:port
		if strings.HasPrefix(phost, "*.") {
			phost = phost[1:]
		}
		if phost[0] != '.' {
			phost = "." + phost
		}
		e.domainIgnores = append(e.domainIgnores, &domainIgnore{host: phost, port: pport})
	}
}

func (e *envNoProxyOnce) reset() {
	e.envOnce.reset()
	e.ipIgnores = nil
	e.domainIgnores = nil
}

// useProxy reports whether requests to addr should use a proxy,
// according to the NO_PROXY or no_proxy environment variable.
// addr is always a canonicalAddr with a host and port.
func useProxy(addr string) bool {
	if len(addr) == 0 {
		return true
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	if noProxyEnv.Get() == "*" {
		return false
	}

	addr = strings.ToLower(strings.TrimSpace(host))
	if ip != nil {
		for _, i := range noProxyEnv.ipIgnores {
			if i.ignore(addr, port, ip) {
				return false
			}
		}
	}
	for _, i := range noProxyEnv.domainIgnores {
		if i.ignore(addr, port, ip) {
			return false
		}
	}
	return true
}
