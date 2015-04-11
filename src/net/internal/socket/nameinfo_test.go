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

var getnameinfoPTRTests = []struct {
	ip     []byte
	suffix string
}{
	{[]byte{8, 8, 8, 8}, ".google.com"},
	{[]byte{8, 8, 4, 4}, ".google.com"},
}

func TestGetnameinfoPTR(t *testing.T) {
	if testing.Short() {
		t.Skip("avoid external network")
	}

	for i, tt := range getnameinfoPTRTests {
		ptr, err := GetnameinfoPTR(tt.ip)
		if err != nil {
			t.Errorf("#%d: %v", i, err)
			continue
		}
		if !strings.HasSuffix(ptr, tt.suffix) && !strings.HasSuffix(ptr, tt.suffix+".") {
			t.Errorf("#%d: got %s; want a record containing %s", i, ptr, tt.suffix)
			continue
		}
	}
}
