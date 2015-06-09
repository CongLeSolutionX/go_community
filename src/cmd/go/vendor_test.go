// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for vendoring semantics.

package main_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestVendorImports(t *testing.T) {
	tg := testgo(t)
	defer tg.cleanup()
	tg.setenv("GOPATH", filepath.Join(tg.pwd(), "testdata"))
	tg.setenv("GO15VENDOREXPERIMENT", "1")
	tg.run("list", "-f", "{{.ImportPath}} {{.Imports}}", "vend/...")
	want := `
		vend [vend/vendor/p r]
		vend/hello [fmt vend/vendor/strings]
		vend/subdir [vend/vendor/p r]
		vend/vendor/p []
		vend/vendor/q []
		vend/vendor/strings []
		vend/x [vend/x/vendor/p vend/vendor/q vend/x/vendor/r]
		vend/x/vendor/p []
		vend/x/vendor/r []
	`
	want = strings.Replace(want+"\t", "\n\t\t", "\n", -1)
	want = strings.TrimPrefix(want, "\n")

	have := tg.stdout.String()

	if have != want {
		// Output is sorted, so diff is easy to prepare.
		var diff bytes.Buffer
		have := splitLines(have)
		want := splitLines(want)
		for len(have) > 0 || len(want) > 0 {
			if len(want) == 0 || len(have) > 0 && have[0] < want[0] {
				fmt.Fprintf(&diff, "unexpected: %s\n", have[0])
				have = have[1:]
				continue
			}
			if len(have) == 0 || len(want) > 0 && want[0] < have[0] {
				fmt.Fprintf(&diff, "missing: %s\n", want[0])
				want = want[1:]
				continue
			}
			fmt.Fprintf(&diff, "\t%s\n", want[0])
			want = want[1:]
			have = have[1:]
		}
		t.Errorf("incorrect go list output:\n%s", diff.String())
	}
}

func splitLines(s string) []string {
	x := strings.Split(s, "\n")
	if x[len(x)-1] == "" {
		x = x[:len(x)-1]
	}
	return x
}

func TestVendorHello(t *testing.T) {
	tg := testgo(t)
	defer tg.cleanup()
	tg.setenv("GOPATH", filepath.Join(tg.pwd(), "testdata"))
	tg.setenv("GO15VENDOREXPERIMENT", "1")
	tg.cd(filepath.Join(tg.pwd(), "testdata/src/vend/hello"))
	tg.run("run", "hello.go")
	tg.grepStdout("hello, world", "missing hello world output")
}
