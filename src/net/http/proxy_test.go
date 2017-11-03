// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/url"
	"os"
	"testing"
)

// TODO(mattn):
//	test ProxyAuth

var UseProxyTests = []struct {
	host  string
	match bool
}{
	// Never proxy localhost:
	{"localhost", false},
	{"127.0.0.1", false},
	{"127.0.0.2", false},
	{"[::1]", false},
	{"[::2]", true}, // not a loopback address

	{"192.168.1.1", false},                // matches exact IPv4
	{"192.168.1.2", true},                 // ports do not match
	{"192.168.1.3", false},                // matches exact IPv4:port
	{"192.168.1.4", true},                 // no match
	{"10.0.0.2", false},                   // matches IPv4/CIDR
	{"[2001:db8::52:0:1]", false},         // matches exact IPv6
	{"[2001:db8::52:0:2]", true},          // no match
	{"[2001:db8::52:0:3]", false},         // matches exact [IPv6]:port
	{"[2002:db8:a::123]", false},          // matches IPv6/CIDR
	{"[fe80::424b:c8be:1643:a1b6]", true}, // no match

	{"barbaz.net", false},         // match as .barbaz.net
	{"foobar.com", false},         // have a port but match
	{"foofoobar.com", true},       // not match as a part of foobar.com
	{"baz.com", true},             // not match as a part of barbaz.com
	{"localhost.net", true},       // not match as suffix of address
	{"local.localhost", true},     // not match as prefix as address
	{"barbarbaz.net", true},       // not match because NO_PROXY have a '.'
	{"www.foobar.com", false},     // match because NO_PROXY includes "foobar.com"
	{"wildcard.io", false},        // match as *.wildcard.io
	{"nested.wildcard.io", false}, // match as *.wildcard.io
	{"awildcard.io", true},        // not a match because of '*'
}

func TestUseProxy(t *testing.T) {
	ResetProxyEnv()
	os.Setenv("NO_PROXY", "foobar.com, .barbaz.net, *.wildcard.io, 192.168.1.1, 192.168.1.2:81, 192.168.1.3:80, 10.0.0.0/30, 2001:db8::52:0:1, [2001:db8::52:0:2]:443, [2001:db8::52:0:3]:80, 2002:db8:a::45/64")
	for _, test := range UseProxyTests {
		if useProxy(test.host+":80") != test.match {
			t.Errorf("useProxy(%v) = %v, want %v", test.host, !test.match, test.match)
		}
	}
}

func TestAllNoProxy(t *testing.T) {
	ResetProxyEnv()
	os.Setenv("NO_PROXY", "*")
	for _, test := range UseProxyTests {
		if useProxy(test.host+":80") != false {
			t.Errorf("useProxy(%v) = true, want false", test.host)
		}
	}
}

var cacheKeysTests = []struct {
	proxy  string
	scheme string
	addr   string
	key    string
}{
	{"", "http", "foo.com", "|http|foo.com"},
	{"", "https", "foo.com", "|https|foo.com"},
	{"http://foo.com", "http", "foo.com", "http://foo.com|http|"},
	{"http://foo.com", "https", "foo.com", "http://foo.com|https|foo.com"},
}

func TestCacheKeys(t *testing.T) {
	for _, tt := range cacheKeysTests {
		var proxy *url.URL
		if tt.proxy != "" {
			u, err := url.Parse(tt.proxy)
			if err != nil {
				t.Fatal(err)
			}
			proxy = u
		}
		cm := connectMethod{proxy, tt.scheme, tt.addr}
		if got := cm.key().String(); got != tt.key {
			t.Fatalf("{%q, %q, %q} cache key = %q; want %q", tt.proxy, tt.scheme, tt.addr, got, tt.key)
		}
	}
}

func ResetProxyEnv() {
	for _, v := range []string{"HTTP_PROXY", "http_proxy", "NO_PROXY", "no_proxy"} {
		os.Unsetenv(v)
	}
	ResetCachedEnvironment()
}

func TestInvalidNoProxy(t *testing.T) {
	ResetProxyEnv()
	os.Setenv("NO_PROXY", ":1")
	useProxy("example.com:80") // should not panic
}
