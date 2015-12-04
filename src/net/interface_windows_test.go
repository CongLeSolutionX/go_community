// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/syscall/windows"
	"testing"
)

func TestWindowsInterfaces(t *testing.T) {
	aas, err := adapterAddresses()
	if err != nil {
		t.Fatal(err)
	}
	ift, err := Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	for i, ifi := range ift {
		aa := aas[i]
		if len(ifi.HardwareAddr) != int(aa.PhysicalAddressLength) {
			t.Errorf("got %d; want %d", len(ifi.HardwareAddr), aa.PhysicalAddressLength)
		}
		if ifi.MTU > 0x7fffffff {
			t.Errorf("%s: got %d; want less than or equal to 1<<31 - 1", ifi.Name, ifi.MTU)
		}
		if ifi.Flags&FlagUp != 0 && aa.OperStatus != windows.IfOperStatusUp {
			t.Errorf("%s: got %v; should not include FlagUp", ifi.Name, ifi.Flags)
		}
		if ifi.Flags&FlagLoopback != 0 && aa.IfType != windows.IF_TYPE_SOFTWARE_LOOPBACK {
			t.Errorf("%s: got %v; should not include FlagLoopback", ifi.Name, ifi.Flags)
		}
	}
}
