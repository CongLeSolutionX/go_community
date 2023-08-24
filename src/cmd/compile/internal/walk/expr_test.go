// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package walk

import (
	"bufio"
	"bytes"
	"internal/testenv"
	"path/filepath"
	"regexp"
	"testing"
)

func TestUsemethod(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	tmpdir := t.TempDir()

	t.Run("StructOf", func(t *testing.T) {
		src := filepath.Join("testdata", "usemethod.go")
		exe := filepath.Join(tmpdir, "usemethod.go.exe")
		cmd := testenv.Command(t, testenv.GoToolPath(t), "build", "-ldflags=-dumpdep", "-o", exe, src)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("%v: %v:\n%s", cmd.Args, err, out)
		}

		wrong := regexp.MustCompile("reflect.StructOf.*ReflectMethod")

		s := bufio.NewScanner(bytes.NewBuffer(out))
		for s.Scan() {
			if wrong.MatchString(s.Text()) {
				t.Fatalf("reflect.StructOf must not be a ReflectMethod")
			}
		}
	})
}
