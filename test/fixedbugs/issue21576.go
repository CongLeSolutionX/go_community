// +build !nacl,!js
// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Ensure that even deadlock detection can still
// run even with an import of "_ os/signal".

package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const prog = `
package main

import _ "os/signal"

func main() {
  c := make(chan int)
  c <- 1
}
`

func main() {
	dir, err := ioutil.TempDir("", "21576")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "main.go")
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Creating main.go %v", err)
	}

	_, err = f.Write([]byte(prog))
	f.Close()
	if err != nil {
		log.Fatalf("Write error %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	stderr := new(strings.Builder)
	cmd := exec.CommandContext(ctx, "go", "run", file)
	cmd.Stderr = stderr
	cmd.Stdout = stderr
	if err := cmd.Run(); err == nil {
		log.Fatalf("Passed, expected an error")
	}
	defer cancel()

	output := stderr.String()
	wantMatches := []string{
		"fatal error: all goroutines are asleep - deadlock!",
		"\n",
		"goroutine 1 \\[chan send\\]:",
		"\n",
		"main.main\\(\\)",
		"\n",
		".+:8 .+",
		"\n",
		"exit status .+",
		"\n",
		"\n",
	}
	strOutput := string(output)
	for _, want := range wantMatches {
		reg := regexp.MustCompile(want)
		match := reg.FindString(strOutput)
		if match == "" {
			log.Printf("Unknown match pattern %q", want)
			continue
		}

		index := strings.Index(strOutput, match)
		strOutput = strOutput[:index] + strOutput[index+len(match):]
	}
	if len(strOutput) != 0 {
		log.Fatalf("Unmatched residual error message:\n%q", strOutput)
	}
}
