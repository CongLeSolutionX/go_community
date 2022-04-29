// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goroot

import (
	"log"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func getStdpkgs() []string {
	allPkgs := make(map[string]bool)
	for _, t := range []string{"linux", "wasm,js", "windows"} { // covering set?
		stdpkgs, err := exec.Command("go", "list", "-tags", t, "std", "cmd", "builtin").Output()
		if err != nil {
			log.Fatal(err)
		}
		for _, v := range strings.Fields(string(stdpkgs)) {
			allPkgs[v] = true
		}
	}
	var l []string
	for v := range allPkgs {
		l = append(l, v)
	}
	sort.Strings(l)
	return l
}

func TestUpToDate(t *testing.T) {
	wantstd := getStdpkgs()
	if !reflect.DeepEqual(std, wantstd) {
		t.Error("List of standard packages not up to date. Please re-run mkstdpkgs.sh")
	}
}
