// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gofmt

import (
	"flag"
	"os"
	"testing"
)

func TestArgsWithAllFlagOptions(t *testing.T) {
	f := &Flag{
		List:        true,
		Write:       true,
		RewriteRule: "rule",
		SimplifyAST: true,
		DoDiff:      true,
		AllErrors:   true,
		Cpuprofile:  "cpuprofile",
	}

	expected := []string{"-e", "-cpuprofile", "cpuprofile", "-d", "-l", "-r", "rule", "-s", "-w"}
	got := f.Args()
	if len(got) != len(expected) {
		t.Fatal("The lengths of got and expected are different")
	}

	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("Got and expected are different [index=%d, got=%v, expected=%v]", i, got[i], expected[i])
		}
	}
}

func TestArgsWithNoFlagOptions(t *testing.T) {
	f := &Flag{}

	if len(f.Args()) != 0 {
		t.Fatal("The expected is not empty")
	}
}

func TestInitGofmtFlagWithAllOptions(t *testing.T) {
	var gofmtFlag Flag

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	gofmtFlag.InitGofmtFlag(flagSet)

	args := []string{"-e", "-cpuprofile", "cpuprofile", "-d", "-l", "-r", "rule", "-s", "-w"}
	flagSet.Parse(args)

	if gofmtFlag.List != true {
		t.Errorf("-l is not true")
	}
	if gofmtFlag.Write != true {
		t.Errorf("-w is not true")
	}
	if gofmtFlag.RewriteRule != "rule" {
		t.Errorf("-r differs from expectation [expected='rule', actual='%s']", gofmtFlag.RewriteRule)
	}
	if gofmtFlag.SimplifyAST != true {
		t.Errorf("-s is not true")
	}
	if gofmtFlag.DoDiff != true {
		t.Errorf("-d is not true")
	}
	if gofmtFlag.AllErrors != true {
		t.Errorf("-e is not true")
	}
	if gofmtFlag.Cpuprofile != "cpuprofile" {
		t.Errorf("-cpuprofile differs from expectation [expected='cpuprofile', actual='%s']", gofmtFlag.Cpuprofile)
	}
}

func TestInitGofmtFlagWithNoOptions(t *testing.T) {
	var gofmtFlag Flag

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	gofmtFlag.InitGofmtFlag(flagSet)

	flagSet.Parse([]string{})

	if gofmtFlag.List != false {
		t.Errorf("-l is not false")
	}
	if gofmtFlag.Write != false {
		t.Errorf("-w is not false")
	}
	if gofmtFlag.RewriteRule != "" {
		t.Errorf("-r is not empty")
	}
	if gofmtFlag.SimplifyAST != false {
		t.Errorf("-s is not false")
	}
	if gofmtFlag.DoDiff != false {
		t.Errorf("-d is not false")
	}
	if gofmtFlag.AllErrors != false {
		t.Errorf("-e is not false")
	}
	if gofmtFlag.Cpuprofile != "" {
		t.Errorf("-c is not empty")
	}
}
