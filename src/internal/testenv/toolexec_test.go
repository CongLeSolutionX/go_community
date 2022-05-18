// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testenv_test

import (
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const shakespeareKingRichard = `
package main
import (
  "fmt"
  "os"
  "os/exec"
)
func main() {
  fmt.Fprintf(os.Stderr, "Now is the winter of our discontent\n")
  cmd := exec.Command(os.Args[2], os.Args[3:]...)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  if err := cmd.Run(); err != nil {
    os.Exit(1)
  }
}
`

func TestBuildToolExec(t *testing.T) {
	t.Parallel()
	te, err := testenv.BuildToolExec(t, "link")

	// If no go build, then make sure BuildToolExec returns ""
	if !testenv.HasGoBuild() {
		if err == nil || te != "" {
			t.Fatalf("expected error since testenv.HasGoBuild is false")
		} else {
			return
		}
	}

	// Otherwise, try it out. Builder a helper first.
	// Build tool, return a path to binary.
	dir := t.TempDir()
	quotesrc := filepath.Join(dir, "sayquote.go")
	if err := os.WriteFile(quotesrc, []byte(shakespeareKingRichard), 0666); err != nil {
		t.Fatalf("os.WriteFile(%s) failed: %v", quotesrc, err)
	}
	quoter := filepath.Join(dir, "sayquote.exe")
	out, err2 := exec.Command(testenv.GoToolPath(t), "build", "-o", quoter, quotesrc).CombinedOutput()
	if err2 != nil {
		t.Logf("%s\n", out)
		t.Fatalf("can't built sayquote.go: %v", err2)
	}

	// The link in this case is obviously going to fail, but we should
	// still see our quote.
	targexe := filepath.Join(dir, "targexe.exe")
	args := []string{"build", "-x", "-toolexec=" + te + " " + quoter,
		"-o", targexe, quotesrc}
	out, err = exec.Command(testenv.GoToolPath(t), args...).CombinedOutput()
	if err == nil {
		t.Fatalf("expected build failure: out:\n%s\n", out)
	}
	if !strings.Contains(string(out), "Now is the winter of our discontent") {
		t.Fatalf("toolexec interposition failed: out:\n%s\n", out)
	}
}
