// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"flag"
	"reflect"
	"strings"
	"testing"
)

func TestPassFlagToTestIncludesAllTestFlags(t *testing.T) {
	flag.VisitAll(func(f *flag.Flag) {
		if !strings.HasPrefix(f.Name, "test.") {
			return
		}
		name := strings.TrimPrefix(f.Name, "test.")
		switch name {
		case "testlogfile", "paniconexit0":
			// These are internal flags.
		default:
			if !passFlagToTest[name] {
				t.Errorf("passFlagToTest missing entry for %q (flag test.%s)", name, name)
				t.Logf("(Run 'go generate cmd/go/internal/test' if it should be added.)")
			}
		}
	})

	for name := range passFlagToTest {
		if flag.Lookup("test."+name) == nil {
			t.Errorf("passFlagToTest contains %q, but flag -test.%s does not exist in test binary", name, name)
		}

		if CmdTest.Flag.Lookup(name) == nil {
			t.Errorf("passFlagToTest contains %q, but flag -%s does not exist in 'go test' subcommand", name, name)
		}
	}
}

func TestVetAnalyzersSetIsCorrect(t *testing.T) {
	// TODO: is there a better source of truth than this?
	want := map[string]bool{
		"composites.whitelist":       true,
		"structtag":                  true,
		"atomic":                     true,
		"buildtag":                   true,
		"asmdecl":                    true,
		"cgocall":                    true,
		"lostcancel":                 true,
		"printf":                     true,
		"tests":                      true,
		"assign":                     true,
		"composites":                 true,
		"nilfunc":                    true,
		"shift":                      true,
		"testinggoroutine":           true,
		"unusedresult":               true,
		"bools":                      true,
		"httpresponse":               true,
		"stdmethods":                 true,
		"unmarshal":                  true,
		"unusedresult.funcs":         true,
		"copylocks":                  true,
		"unusedresult.stringmethods": true,
		"loopclosure":                true,
		"printf.funcs":               true,
		"unsafeptr":                  true,
		"unreachable":                true,
		"errorsas":                   true,
		"framepointer":               true,
		"ifaceassert":                true,
		"sigchanyzer":                true,
		"stringintconv":              true,
	}

	if !reflect.DeepEqual(want, passAnalyzersToVet) {
		t.Errorf("want %v; got %v", want, passAnalyzersToVet)
	}
}
