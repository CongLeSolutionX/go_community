// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"reflect"
	"sort"
	"testing"
)

var lookupStaticHostTests = []struct {
	host  string
	addrs []string
}{
	// internet address and host name
	{
		"localhost",
		[]string{"127.0.0.1", "fe80::1"},
	},
	{
		"broadcasthost",
		[]string{"255.255.255.255"},
	},
	{
		"odin",
		[]string{"127.0.0.2", "127.0.0.3", "::2"},
	},
	{
		"thor",
		[]string{"127.1.1.1"},
	},

	// internet address with zone identifier and host name
	{
		"ipv6-localhost",
		[]string{"fe80::1%lo0"},
	},

	// internet address, host name and aliases
	{
		"ullr",
		[]string{"127.1.1.2"},
	},
	{
		"ullr.localdomain",
		[]string{"127.1.1.2"},
	},

	// bogus entries
	{
		"loki",
		nil,
	},
}

func TestLookupStaticHost(t *testing.T) {
	p := testHookHostsPath
	testHookHostsPath = "testdata/hosts"
	defer func() { testHookHostsPath = p }()

	for i, tt := range lookupStaticHostTests {
		addrs := lookupStaticHost(tt.host)
		if !reflect.DeepEqual(addrs, tt.addrs) {
			t.Errorf("#%d: got %v; want %v", i, addrs, tt.addrs)
		}
	}

	testHookHostsPath = "testdata/hosts_singleline" // see golang.org/issue/6646
	ttaddrs := []string{"127.0.0.2"}

	addrs := lookupStaticHost("odin")
	if !reflect.DeepEqual(addrs, ttaddrs) {
		t.Errorf("got %v; want %v", addrs, ttaddrs)
	}
}

var lookupStaticAddrTests = []struct {
	addr  string
	hosts []string
}{
	{
		"fe80::1",
		[]string{"localhost"},
	},
	{
		"fe80:0:0:0:0:0:0:1",
		[]string{"localhost"},
	},
}

func TestLookupStaticAddr(t *testing.T) {
	p := testHookHostsPath
	testHookHostsPath = "testdata/hosts"
	defer func() { testHookHostsPath = p }()

	for i, tt := range lookupStaticAddrTests {
		hosts := lookupStaticAddr(tt.addr)
		if !reflect.DeepEqual(hosts, tt.hosts) {
			t.Errorf("#%d: got %v; want %v", i, hosts, tt.hosts)
		}
	}
}

func TestLookupHost(t *testing.T) {
	// Can't depend on this to return anything in particular,
	// but if it does return something, make sure it doesn't
	// duplicate addresses (a common bug due to the way
	// getaddrinfo works).
	addrs, _ := LookupHost("localhost")
	sort.Strings(addrs)
	for i := 0; i+1 < len(addrs); i++ {
		if addrs[i] == addrs[i+1] {
			t.Fatalf("LookupHost(\"localhost\") = %v, has duplicate addresses", addrs)
		}
	}
}
