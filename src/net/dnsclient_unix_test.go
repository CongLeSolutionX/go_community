// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

var dnsTransportFallbackTests = []struct {
	server  string
	name    string
	qtype   uint16
	timeout int
	rcode   int
}{
	// Querying "com." with qtype=255 usually makes an answer
	// which requires more than 512 bytes.
	{"8.8.8.8:53", "com.", dnsTypeALL, 2, dnsRcodeSuccess},
	{"8.8.4.4:53", "com.", dnsTypeALL, 4, dnsRcodeSuccess},
}

func TestDNSTransportFallback(t *testing.T) {
	if testing.Short() || !*testExternal {
		t.Skip("avoid external network")
	}

	for _, tt := range dnsTransportFallbackTests {
		timeout := time.Duration(tt.timeout) * time.Second
		msg, err := exchange(tt.server, tt.name, tt.qtype, timeout)
		if err != nil {
			t.Error(err)
			continue
		}
		switch msg.rcode {
		case tt.rcode, dnsRcodeServerFailure:
		default:
			t.Errorf("got %v from %v; want %v", msg.rcode, tt.server, tt.rcode)
			continue
		}
	}
}

// See RFC 6761 for further information about the reserved, pseudo
// domain names.
var specialDomainNameTests = []struct {
	name  string
	qtype uint16
	rcode int
}{
	// Name resolution APIs and libraries should not recognize the
	// followings as special.
	{"1.0.168.192.in-addr.arpa.", dnsTypePTR, dnsRcodeNameError},
	{"test.", dnsTypeALL, dnsRcodeNameError},
	{"example.com.", dnsTypeALL, dnsRcodeSuccess},

	// Name resolution APIs and libraries should recognize the
	// followings as special and should not send any queries.
	// Though, we test those names here for verifying nagative
	// answers at DNS query-response interaction level.
	{"localhost.", dnsTypeALL, dnsRcodeNameError},
	{"invalid.", dnsTypeALL, dnsRcodeNameError},
}

func TestSpecialDomainName(t *testing.T) {
	if testing.Short() || !*testExternal {
		t.Skip("avoid external network")
	}

	server := "8.8.8.8:53"
	for _, tt := range specialDomainNameTests {
		msg, err := exchange(server, tt.name, tt.qtype, 0)
		if err != nil {
			t.Error(err)
			continue
		}
		switch msg.rcode {
		case tt.rcode, dnsRcodeServerFailure:
		default:
			t.Errorf("got %v from %v; want %v", msg.rcode, server, tt.rcode)
			continue
		}
	}
}

type testResolvConf struct {
	dir  string
	path string
	*resolverConfig
}

func newTestResolvConf() (*testResolvConf, error) {
	dir, err := ioutil.TempDir("", "go-resolvconftest")
	if err != nil {
		return nil, err
	}
	return &testResolvConf{
		dir:            dir,
		path:           path.Join(dir, "resolv.conf"),
		resolverConfig: &resolvConf,
	}, nil
}

func (conf *testResolvConf) write(lines []string) error {
	conf.update("/etc/resolv.conf")
	// Make sure the file mtime will be different once we're done
	// here, even on systems with coarse (1s) mtime resolution.
	time.Sleep(time.Second)

	f, err := os.OpenFile(conf.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(strings.Join(lines, "\n")); err != nil {
		f.Close()
		return err
	}
	f.Close()
	conf.acquireSema()
	conf.lastChecked = time.Time{}
	conf.releaseSema()
	return nil
}

func (conf *testResolvConf) servers() []string {
	conf.mu.RLock()
	servers := conf.dnsConfig.servers
	conf.mu.RUnlock()
	return servers
}

func (conf *testResolvConf) teardown() error {
	conf.acquireSema()
	conf.lastChecked = time.Time{}
	conf.releaseSema()
	conf.update("/etc/resolv.conf")
	return os.RemoveAll(conf.dir)
}

var updateResolvConfTests = []struct {
	name  string
	lines []string
	out   []string
}{
	{
		"golang.org",
		[]string{"nameserver 8.8.8.8"},
		[]string{"8.8.8.8"},
	},
	{
		"",
		nil, // an empty resolv.conf should use defaultNS as name servers
		defaultNS,
	},
	{
		"golang.org",
		[]string{"nameserver 8.8.4.4"},
		[]string{"8.8.4.4"},
	},
}

func TestUpdateResolvConf(t *testing.T) {
	if testing.Short() || !*testExternal {
		t.Skip("avoid external network")
	}

	conf, err := newTestResolvConf()
	if err != nil {
		t.Fatal(err)
	}
	defer conf.teardown()

	for _, tt := range updateResolvConfTests {
		if err := conf.write(tt.lines); err != nil {
			t.Error(err)
			continue
		}
		conf.update(conf.path)
		if tt.name != "" {
			ips, err := goLookupIP(tt.name)
			if err != nil {
				t.Error(err)
			}
			if len(ips) == 0 {
				t.Errorf("no records for %s", tt.name)
			}
		}
		servers := conf.servers()
		if !reflect.DeepEqual(servers, tt.out) {
			t.Errorf("got %v; want %v", servers, tt.out)
			continue
		}
	}
}

func BenchmarkGoLookupIP(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	for i := 0; i < b.N; i++ {
		goLookupIP("www.example.com")
	}
}

func BenchmarkGoLookupIPNoSuchHost(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	for i := 0; i < b.N; i++ {
		goLookupIP("some.nonexistent")
	}
}

func BenchmarkGoLookupIPWithBrokenNameServer(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	// This looks ugly but it's safe as long as benchmarks are run
	// sequentially in package testing.
	resolvConf.update("/etc/resolv.conf")
	resolvConf.acquireSema()
	orig := *resolvConf.dnsConfig
	defer func() {
		resolvConf.dnsConfig = &orig
		resolvConf.releaseSema()
	}()

	resolvConf.dnsConfig.servers = append([]string{"203.0.113.254"}, resolvConf.dnsConfig.servers...) // use TEST-NET-3 block, see RFC 5737
	for i := 0; i < b.N; i++ {
		goLookupIP("www.example.com")
	}
}
