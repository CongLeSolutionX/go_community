// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for vendoring semantics.

package main_test

import (
	"internal/testenv"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVendorGOPATH(t *testing.T) {
	tg := testgo(t)
	defer tg.cleanup()
	changeVolume := func(s string, f func(s string) string) string {
		vol := filepath.VolumeName(s)
		return f(vol) + s[len(vol):]
	}
	gopath := changeVolume(filepath.Join(tg.pwd(), "testdata"), strings.ToLower)
	tg.setenv("GOPATH", gopath)
	cd := changeVolume(filepath.Join(tg.pwd(), "testdata/src/vend/hello"), strings.ToUpper)
	tg.cd(cd)
	tg.run("run", "hello.go")
	tg.grepStdout("hello, world", "missing hello world output")
}

func TestVendorGet(t *testing.T) {
	tooSlow(t)
	tg := testgo(t)
	defer tg.cleanup()
	tg.tempFile("src/v/m.go", `
		package main
		import ("fmt"; "vendor.org/p")
		func main() {
			fmt.Println(p.C)
		}`)
	tg.tempFile("src/v/m_test.go", `
		package main
		import ("fmt"; "testing"; "vendor.org/p")
		func TestNothing(t *testing.T) {
			fmt.Println(p.C)
		}`)
	tg.tempFile("src/v/vendor/vendor.org/p/p.go", `
		package p
		const C = 1`)
	tg.setenv("GOPATH", tg.path("."))
	tg.cd(tg.path("src/v"))
	tg.run("run", "m.go")
	tg.run("test")
	tg.run("list", "-f", "{{.Imports}}")
	tg.grepStdout("v/vendor/vendor.org/p", "import not in vendor directory")
	tg.run("list", "-f", "{{.TestImports}}")
	tg.grepStdout("v/vendor/vendor.org/p", "test import not in vendor directory")
	tg.run("get", "-d")
	tg.run("get", "-t", "-d")
}

func TestLegacyModGet(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)
	testenv.MustHaveExecPath(t, "git")

	tg := testgo(t)
	defer tg.cleanup()
	tg.makeTempdir()
	tg.setenv("GOPATH", tg.path("d1"))
	tg.run("get", "vcs-test.golang.org/git/modlegacy1-old.git/p1")
	tg.run("list", "-f", "{{.Deps}}", "vcs-test.golang.org/git/modlegacy1-old.git/p1")
	tg.grepStdout("new.git/p2", "old/p1 should depend on new/p2")
	tg.grepStdoutNot("new.git/v2/p2", "old/p1 should NOT depend on new/v2/p2")
	tg.run("build", "vcs-test.golang.org/git/modlegacy1-old.git/p1", "vcs-test.golang.org/git/modlegacy1-new.git/p1")

	tg.setenv("GOPATH", tg.path("d2"))

	tg.must(os.RemoveAll(tg.path("d2")))
	tg.run("get", "github.com/rsc/vgotest5")
	tg.run("get", "github.com/rsc/vgotest4")
	tg.run("get", "github.com/myitcv/vgo_example_compat")

	if testing.Short() {
		return
	}

	tg.must(os.RemoveAll(tg.path("d2")))
	tg.run("get", "github.com/rsc/vgotest4")
	tg.run("get", "github.com/rsc/vgotest5")
	tg.run("get", "github.com/myitcv/vgo_example_compat")

	tg.must(os.RemoveAll(tg.path("d2")))
	tg.run("get", "github.com/rsc/vgotest4", "github.com/rsc/vgotest5")
	tg.run("get", "github.com/myitcv/vgo_example_compat")

	tg.must(os.RemoveAll(tg.path("d2")))
	tg.run("get", "github.com/rsc/vgotest5", "github.com/rsc/vgotest4")
	tg.run("get", "github.com/myitcv/vgo_example_compat")

	tg.must(os.RemoveAll(tg.path("d2")))
	tg.run("get", "github.com/myitcv/vgo_example_compat")
	tg.run("get", "github.com/rsc/vgotest4", "github.com/rsc/vgotest5")

	pkgs := []string{"github.com/myitcv/vgo_example_compat", "github.com/rsc/vgotest4", "github.com/rsc/vgotest5"}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				if i == j || i == k || k == j {
					continue
				}
				tg.must(os.RemoveAll(tg.path("d2")))
				tg.run("get", pkgs[i], pkgs[j], pkgs[k])
			}
		}
	}
}
