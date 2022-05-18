// Copyright 2022 The Go Authors. All rights reserved.
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
)
func main() {
  if os.Args[1] == "-V=full" {
    fmt.Printf("link version something buildID=blah\n")
  } else {
    fmt.Fprintf(os.Stderr, "Now is the winter of our discontent\n")
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
		t.Fatalf("can't built sayquote.go: %v\n%s", err2, string(out))
	}

	// The link in this case is obviously going to fail (since the
	// toolexec wrapper will run sayquote.exe instead of "link"), but
	// we should still see our quote.
	targexe := filepath.Join(dir, "targexe.exe")
	args := []string{"build", "-x", "-toolexec", te + " " + quoter,
		"-o", targexe, quotesrc}
	//t.Logf("%s %+v\n", testenv.GoToolPath(t), args)
	out, err = exec.Command(testenv.GoToolPath(t), args...).CombinedOutput()
	if err == nil {
		t.Fatalf("expected build failure: out:\n%s\n", out)
	}
	if !strings.Contains(string(out), "Now is the winter of our discontent") {
		t.Fatalf("toolexec interposition failed: out:\n%s\n", out)
	}
}
