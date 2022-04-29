// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goroot

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestUpToDate(t *testing.T) {
	stdlist, err := exec.Command("go", "list", "std", "cmd").Output()
	if err != nil {
		t.Fatal(err)
	}
	wantstd := strings.Fields(string(stdlist))
	if !reflect.DeepEqual(std, wantstd) {
		t.Error("List of standard packages not up to date. Please re-run mkstdpkgs.sh")
	}
}
