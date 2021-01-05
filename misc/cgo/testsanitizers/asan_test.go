// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sanitizers_test

import (
	"strings"
	"testing"
)

func TestASAN(t *testing.T) {
	t.Parallel()
	requireOvercommit(t)
	config := configure("address")
	config.skipIfCSanitizerBroken(t)

	mustRun(t, config.goCmd("build", "std"))

	cases := []struct {
		src       string
		noWantErr bool
	}{
		{src: "asan1_fail.go"},
		{src: "asan2_fail.go"},
		{src: "asan3_fail.go"},
		{src: "asan4_fail.go"},
		{src: "asan_useAfterReturn.go", noWantErr: true},
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
			if !tc.noWantErr {
				out, err := cmd.CombinedOutput()
				if err != nil {
					return
				}
				t.Fatalf("%#q exited without error; want ASAN failure\n%s", strings.Join(cmd.Args, " "), out)
			}
			mustRun(t, cmd)
		})
	}
}
