// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"cmd/go/internal/base"
	"os"
	"strings"
	"testing"
)

func TestChdir(t *testing.T) {
	script, err := os.ReadFile("testdata/script/chdir.txt")
	if err != nil {
		t.Fatal(err)
	}

	var walk func(string, *base.Command)
	walk = func(name string, cmd *base.Command) {
		if len(cmd.Commands) > 0 {
			for _, sub := range cmd.Commands {
				walk(name+" "+sub.Name(), sub)
			}
			return
		}
		if !cmd.Runnable() {
			return
		}
		if cmd.CustomFlags {
			if !strings.Contains(string(script), "# "+name+"\n") {
				t.Errorf("%s has custom flags, not tested in testdata/script/chdir.txt", name)
			}
			return
		}
		f := cmd.Flag.Lookup("C")
		if f == nil {
			t.Errorf("%s has no -C flag", name)
		} else if f.Usage != "AddChdirFlag" {
			t.Errorf("%s has -C flag but not from AddChdirFlag", name)
		}
	}
	walk("go", base.Go)
}
