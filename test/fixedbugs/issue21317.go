// +build !js,gc
// run

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// As of "Mon 6 Nov 2017", run.go doesn't yet have proper
// column matching so instead match the output manually
// by exec-ing

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	f, err := ioutil.TempFile("", "issue21317.go")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(f, `
package main

import "fmt"

func main() {
        n, err := fmt.Println(1)
}
`)
	f.Close()
	defer os.RemoveAll(f.Name())

	importcfg, err := os.CreateTemp("", "importcfg")
	output, err := exec.Command("go", "list", "-export", "-f", "{{if .Export}}packagefile {{.ImportPath}}={{.Export}}{{end}}", "std").Output()
	if err != nil {
		log.Fatal(err)
	}
	importcfg.Write(output)

	// compile and test output
	cmd := exec.Command("go", "tool", "compile", "-p=main", "-importcfg="+importcfg.Name(), f.Name())
	out, err := cmd.CombinedOutput()
	if err == nil {
		log.Fatalf("expected cmd/compile to fail")
	}
	wantErrs := []string{
		"7:9: n declared and not used",
		"7:12: err declared and not used",
	}
	outStr := string(out)
	for _, want := range wantErrs {
		if !strings.Contains(outStr, want) {
			log.Fatalf("failed to match %q\noutput: %q", want, outStr)
		}
	}
}
