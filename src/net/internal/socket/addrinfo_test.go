// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo,!netgo
// +build !nacl,!plan9,!solaris,!windows

package socket

import (
	"strings"
	"testing"
)

var getaddrinfoPortTests = []struct {
	network, service string
	port             int
}{
	{"tcp", "http", 80},
	{"tcp4", "http", 80},
	{"tcp6", "http", 80},
}

func TestGetaddrinfoPort(t *testing.T) {
	for i, tt := range getaddrinfoPortTests {
		port, err := GetaddrinfoPort(tt.network, tt.service)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if port != tt.port {
			t.Errorf("#%d: got %d; want %d", i, port, tt.port)
			continue
		}
	}
}

var getaddrinfoAddrTests = []struct {
	name string
}{
	{"golang.org"},
	{"golang.org."},
}

func TestGetaddrinfoAddr(t *testing.T) {
	if testing.Short() {
		t.Skip("avoid external network")
	}

	for i, tt := range getaddrinfoAddrTests {
		sas, err := GetaddrinfoAddr(tt.name)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if len(sas) == 0 {
			t.Errorf("#%d: num address records = 0; want >0", i)
			continue
		}
	}
}

var getaddrinfoCNAMETests = []struct {
	name, suffix string
}{
	{"www.iana.com", ".icann.org"},
	{"www.iana.com.", ".icann.org"},
}

func TestGetaddrinfoCNAME(t *testing.T) {
	if testing.Short() {
		t.Skip("avoid external network")
	}

	for i, tt := range getaddrinfoCNAMETests {
		cname, err := GetaddrinfoCNAME(tt.name)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if !strings.HasSuffix(cname, tt.suffix) && !strings.HasSuffix(cname, tt.suffix+".") {
			t.Errorf("#%d: got %s; want a record containing %s", i, cname, tt.suffix)
			continue
		}
	}
}
