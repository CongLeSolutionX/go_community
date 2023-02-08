// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"cmd/go/internal/cfg"
	"cmd/go/internal/test/internal/genflags"
	"internal/testenv"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cfg.SetGOROOT(testenv.GOROOT(nil), false)
	os.Exit(m.Run())
}

func TestPassFlagToTestIncludesAllTestFlags(t *testing.T) {
	unwanted := map[string]bool{}
	for name, ok := range passFlagToTest {
		unwanted[name] = ok
	}

	for _, name := range genflags.PassFlagToTest() {
		if got, ok := passFlagToTest[name]; !got {
			t.Errorf("passFlagToTest[%q] = %v, %v; want true", name, got, ok)
		}
		delete(unwanted, name)
	}
	if len(unwanted) > 0 {
		t.Errorf("unexpected entries in passFlagToTest: %v", unwanted)
	}

	for name, ok := range passFlagToTest {
		if ok && CmdTest.Flag.Lookup(name) == nil {
			t.Errorf("passFlagToTest contains %q, but flag -%s does not exist in 'go test' subcommand", name, name)
		}
	}
}

func TestVetAnalyzersSetIsCorrect(t *testing.T) {
	unwanted := map[string]bool{}
	for name, ok := range passAnalyzersToVet {
		unwanted[name] = ok
	}

	wantNames, err := genflags.VetAnalyzers()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range wantNames {
		if got, ok := passAnalyzersToVet[name]; !got {
			t.Errorf("passAnalyzersToVet[%q] = %v, %v; want true", name, got, ok)
		}
		delete(unwanted, name)
	}
	if len(unwanted) > 0 {
		t.Errorf("unexpected entries in passAnalyzersToVet: %v", unwanted)
	}
}
