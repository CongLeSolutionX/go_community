// run

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that batch files are maintained as CRLF files (consistent
// behavior on all operating systems). See golang.org/issue/37791.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	batches, _ := filepath.Glob(runtime.GOROOT() + "/src/*.bat")
	for _, bat := range batches {
		body, _ := ioutil.ReadFile(bat)
		if !bytes.Contains(body, []byte("\r\n")) {
			fmt.Printf("Windows batch file %s does not contain CRLF line termination.\n", bat)
			os.Exit(1)
		}
	}
}
