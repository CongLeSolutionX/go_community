// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sanitizers_test

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestASAN(t *testing.T) {
	goos, err := goEnv("GOOS")
	if err != nil {
		t.Fatal(err)
	}
	goarch, err := goEnv("GOARCH")
	if err != nil {
		t.Fatal(err)
	}
	// The asan tests require support for the -asan option.
	if !aSanSupported(goos, goarch) {
		t.Skipf("skipping on %s/%s; -asan option is not supported.", goos, goarch)
	}

	// The current implementation is only compatible with the ASan library from version
	// v7 to v9 (See the description in src/runtime/asan/asan.go). Therefore, using the
	// -asan option must use a compatible version of ASan library, which requires that
	// the gcc version is not less than 7 and the clang version is not less than 4,
	// otherwise a segmentation fault will occur.
	cc, err := goEnv("CC")
	if err != nil {
		t.Fatal(err)
	}
	if cc != "gcc" && cc != "clang" {
		t.Skipf("skipping: The expected C compiler is gcc or clang, not %s", cc)
	}
	out, err := exec.Command(cc, "--version").CombinedOutput()
	if err != nil {
		t.Skipf("skipping: error executing C compiler %s: %v", cc, err)
	}
	re := regexp.MustCompile(`([0-9]+)\.([0-9]+)\.0`)
	matches := re.FindSubmatch(out)
	if len(matches) < 3 {
		t.Skipf("skipping: can't determine C compiler %s version from\n%s\n", cc, out)
	}
	major, err1 := strconv.Atoi(string(matches[1]))
	minor, err2 := strconv.Atoi(string(matches[2]))
	if err1 != nil || err2 != nil {
		t.Skipf("skipping: can't determine C compiler %s version: %v, %v", cc, err1, err2)
	}
	if cc == "gcc" {
		if major < 7 {
			t.Skipf("skipping: too old version of gcc %d.%d uses v6 or lower version of ASan library", major, minor)
		}
	} else {
		if major < 4 {
			t.Skipf("skipping: too old version of clang %d.%d uses v6 or lower version of ASan library", major, minor)
		}
	}
	t.Parallel()
	requireOvercommit(t)
	config := configure("address")
	config.skipIfCSanitizerBroken(t)

	mustRun(t, config.goCmd("build", "std"))

	cases := []struct {
		src               string
		memoryAccessError string
	}{
		{src: "asan1_fail.go", memoryAccessError: "heap-use-after-free"},
		{src: "asan2_fail.go", memoryAccessError: "heap-buffer-overflow"},
		{src: "asan3_fail.go", memoryAccessError: "use-after-poison"},
		{src: "asan4_fail.go", memoryAccessError: "use-after-poison"},
		{src: "asan_useAfterReturn.go"},
	}
	for _, tc := range cases {
		tc := tc
		name := strings.TrimSuffix(tc.src, ".go")
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := newTempDir(t)
			defer dir.RemoveAll(t)

			outPath := dir.Join(name)
			mustRun(t, config.goCmd("build", "-o", outPath, srcPath(tc.src)))

			cmd := hangProneCmd(outPath)
			if tc.memoryAccessError != "" {
				out, err := cmd.CombinedOutput()
				if err != nil && strings.Contains(string(out), tc.memoryAccessError) {
					return
				}
				t.Fatalf("%#q exited without expected memory access error\n%s; got failure\n%s", strings.Join(cmd.Args, " "), tc.memoryAccessError, out)
			}
			mustRun(t, cmd)
		})
	}
}
