// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package poll_test

import (
	. "internal/poll"
	"testing"
)

func TestPoolInitializingFd(t *testing.T) {
	for _, net := range []string{"file", "console", "dir"} {
		fd := FD{}
		fd.Init(net)
		if fd.PoolFDInitialized() {
			t.Fatalf("pool fd initialized for unowned net type")
		}
	}

	for _, net := range []string{"tcp", "udp", "ip", "unix"} {
		fd := FD{}
		fd.Init(net)
		if !fd.PoolFDInitialized() {
			t.Fatalf("pool fd not initialized for netpoll owned type")
		}
	}
}
