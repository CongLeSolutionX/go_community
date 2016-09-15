// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package net

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

var dnsReadConfigTests = []struct {
	name string
	want *dnsConfig
}{
	{
		name: "testdata/resolv.conf",
		want: &dnsConfig{
			servers:    []string{"8.8.8.8:53", "[2001:4860:4860::8888]:53", "[fe80::1%lo0]:53"},
			search:     []string{"localdomain."},
			ndots:      5,
			timeout:    10 * time.Second,
			attempts:   3,
			rotate:     true,
			unknownOpt: true, // the "options attempts 3" line
		},
	},
	{
		name: "testdata/domain-resolv.conf",
		want: &dnsConfig{
			servers:  []string{"8.8.8.8:53"},
			search:   []string{"localdomain."},
			ndots:    1,
			timeout:  5 * time.Second,
			attempts: 2,
		},
	},
	{
		name: "testdata/search-resolv.conf",
		want: &dnsConfig{
			servers:  []string{"8.8.8.8:53"},
			search:   []string{"test.", "invalid."},
			ndots:    1,
			timeout:  5 * time.Second,
			attempts: 2,
		},
	},
	{
		name: "testdata/empty-resolv.conf",
		want: &dnsConfig{
			servers:  defaultNS,
			ndots:    1,
			timeout:  5 * time.Second,
			attempts: 2,
			search:   []string{"domain.local."},
		},
	},
	{
		name: "testdata/invalid-ndots-resolv.conf",
		want: &dnsConfig{
			servers:  defaultNS,
			ndots:    0,
			timeout:  5 * time.Second,
			attempts: 2,
			search:   []string{"domain.local."},
		},
	},
	{
		name: "testdata/large-ndots-resolv.conf",
		want: &dnsConfig{
			servers:  defaultNS,
			ndots:    15,
			timeout:  5 * time.Second,
			attempts: 2,
			search:   []string{"domain.local."},
		},
	},
	{
		name: "testdata/negative-ndots-resolv.conf",
		want: &dnsConfig{
			servers:  defaultNS,
			ndots:    0,
			timeout:  5 * time.Second,
			attempts: 2,
			search:   []string{"domain.local."},
		},
	},
	{
		name: "testdata/openbsd-resolv.conf",
		want: &dnsConfig{
			ndots:    1,
			timeout:  5 * time.Second,
			attempts: 2,
			lookup:   []string{"file", "bind"},
			servers:  []string{"169.254.169.254:53", "10.240.0.1:53"},
			search:   []string{"c.symbolic-datum-552.internal."},
		},
	},
}

func TestDNSReadConfig(t *testing.T) {
	origGetHostname := getHostname
	defer func() { getHostname = origGetHostname }()
	getHostname = func() (string, error) { return "host.domain.local", nil }

	for _, tt := range dnsReadConfigTests {
		conf := dnsReadConfig(tt.name)
		if conf.err != nil {
			t.Fatal(conf.err)
		}
		conf.mtime = time.Time{}
		if !reflect.DeepEqual(conf, tt.want) {
			t.Errorf("%s:\ngot: %+v\nwant: %+v", tt.name, conf, tt.want)
		}
	}
}

func TestDNSReadMissingFile(t *testing.T) {
	origGetHostname := getHostname
	defer func() { getHostname = origGetHostname }()
	getHostname = func() (string, error) { return "host.domain.local", nil }

	conf := dnsReadConfig("a-nonexistent-file")
	if !os.IsNotExist(conf.err) {
		t.Errorf("missing resolv.conf:\ngot: %v\nwant: %v", conf.err, os.ErrNotExist)
	}
	conf.err = nil
	want := &dnsConfig{
		servers:  defaultNS,
		ndots:    1,
		timeout:  5 * time.Second,
		attempts: 2,
		search:   []string{"domain.local."},
	}
	if !reflect.DeepEqual(conf, want) {
		t.Errorf("missing resolv.conf:\ngot: %+v\nwant: %+v", conf, want)
	}
}

var dnsDefaultSearchTests = []struct {
	name string
	err  error
	want []string
}{
	{
		name: "host.long.domain.local",
		want: []string{"long.domain.local."},
	},
	{
		name: "host.local",
		want: []string{"local."},
	},
	{
		name: "host",
		want: nil,
	},
	{
		name: "host.domain.local",
		err:  errors.New("errored"),
		want: nil,
	},
	{
		// ensures we don't return []string{""}
		// which causes duplicate lookups
		name: "foo.",
		want: nil,
	},
}

func TestDNSDefaultSearch(t *testing.T) {
	origGetHostname := getHostname
	defer func() { getHostname = origGetHostname }()

	for _, tt := range dnsDefaultSearchTests {
		getHostname = func() (string, error) { return tt.name, tt.err }
		got := dnsDefaultSearch()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("dnsDefaultSearch with hostname %q and error %+v = %q, wanted %q", tt.name, tt.err, got, tt.want)
		}
	}
}

var dnsQueryServersTests = []struct {
	servers []string
	rotate  bool

	want [][]string
}{
	{
		servers: []string{"1.2.3.4"},
		want:    [][]string{{"1.2.3.4"}, {"1.2.3.4"}, {"1.2.3.4"}, {"1.2.3.4"}},
	},
	{
		servers: []string{"1.2.3.4"},
		rotate:  true,
		want:    [][]string{{"1.2.3.4"}, {"1.2.3.4"}, {"1.2.3.4"}, {"1.2.3.4"}},
	},
	{
		servers: []string{"1.2.3.4", "2.3.4.5", "3.4.5.6"},
		want:    [][]string{{"1.2.3.4", "2.3.4.5", "3.4.5.6"}, {"1.2.3.4", "2.3.4.5", "3.4.5.6"}, {"1.2.3.4", "2.3.4.5", "3.4.5.6"}, {"1.2.3.4", "2.3.4.5", "3.4.5.6"}},
	},
	{
		servers: []string{"1.2.3.4", "2.3.4.5", "3.4.5.6"},
		rotate:  true,
		want:    [][]string{{"1.2.3.4", "2.3.4.5", "3.4.5.6"}, {"2.3.4.5", "3.4.5.6", "1.2.3.4"}, {"3.4.5.6", "1.2.3.4", "2.3.4.5"}, {"1.2.3.4", "2.3.4.5", "3.4.5.6"}},
	},
}

func TestDNSQueryServers(t *testing.T) {
	for _, tt := range dnsQueryServersTests {
		conf := &dnsConfig{servers: tt.servers, rotate: tt.rotate}

		var got [][]string
		for i := 0; i < 4; i++ {
			got = append(got, conf.queryServers())
		}

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("queryServers with servers %v and rotate=%t\ngot:\n%v\nwant:\n%v", tt.servers, tt.rotate, got, tt.want)
		}
	}
}
