// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os/exec"
	"strings"
	"testing"
)

func TestInterfacesWindows(t *testing.T) {
	b, err := exec.Command("ipconfig").Output()
	if err != nil {
		t.Fatal("failed to run ipconfig", err)
	}
	output := string(b)

	ni, err := Interfaces()
	if err != nil {
		t.Errorf("Interface failed: %v", err)
	}
	for _, i := range ni {
		// TODO(mattn): When interface name has multi-byte characters, this
		// test doesn't work because exec.Cmd#Output may not be utf-8 on windows.

		//if !strings.Contains(output, i.Name) {
		//	t.Errorf("interface name is not contained in output of ipconfig: %v", i.Name)
		//}
		addrs, err := i.Addrs()
		if err != nil {
			t.Errorf("Addrs failed: %v", err)
		}
		for _, addr := range addrs {
			ip := addr.String()
			ipmask := strings.Split(ip, "/")
			if len(ipmask) != 2 {
				t.Errorf("interface address must have netmask: %v", ipmask)
			}
			if !strings.Contains(output, ipmask[0]) {
				t.Errorf("interface address is not contained in output of ipconfig: %v", ip)
			}
			// TODO(mattn): Should check netmask
		}
	}
}

func TestInterfaceAddrsWindows(t *testing.T) {
	b, err := exec.Command("ipconfig").Output()
	if err != nil {
		t.Fatal("failed to run ipconfig", err)
	}
	output := string(b)

	na, err := InterfaceAddrs()
	if err != nil {
		t.Errorf("Interface failed: %v", err)
	}
	for _, a := range na {
		ip := a.String()
		ipmask := strings.Split(ip, "/")
		if len(ipmask) != 2 {
			t.Errorf("interface address must have netmask: %v", ipmask)
		}
		if !strings.Contains(output, ipmask[0]) {
			t.Errorf("interface address is not contained in output of ipconfig: %v", ip)
		}
		// TODO(mattn): Should check netmask
	}
}
