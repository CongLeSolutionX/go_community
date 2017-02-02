// run
// +build !nacl

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// nacl is excluded because this runs a compiler.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func updateEnv(env *[]string, key, val string) {
	if val != "" {
		var found bool
		key = key + "="
		for i, kv := range *env {
			if strings.HasPrefix(kv, key) {
				(*env)[i] = key + val
				found = true
				break
			}
		}
		if !found {
			*env = append(*env, key+val)
		}
	}
}

func main() {
	// Set this to see how other platforms score.
	testarch := os.Getenv("TESTARCH")
	debug := os.Getenv("TESTDEBUG") != ""

	cmd := exec.Command("go", "build", "-gcflags", "-S", "fixedbugs/issue18902b.go")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Env = os.Environ()

	updateEnv(&cmd.Env, "GOARCH", testarch)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s\n%s", err, buf.Bytes())
		return
	}
	begin := "\"\".(*gcSortBuf).flush"
	s := buf.String()
	i := strings.Index(s, begin)
	if i < 0 {
		fmt.Printf("Failed to find expected symbol %s in output\n%s\n", begin, s)
		return
	}
	s = s[i:]
	r := strings.NewReader(s)
	scanner := bufio.NewScanner(r)
	first := true
	beforeLineNumber := "issue18902b.go:"
	lbln := len(beforeLineNumber)

	var scannedCount, changes, sumdiffs float64

	prevVal := 0
	for scanner.Scan() {
		line := scanner.Text()
		if first {
			first = false
			continue
		}
		i = strings.Index(line, beforeLineNumber)
		if i < 0 {
			// We're done
			if scannedCount < 200 { // When test written, 251 observed.
				fmt.Printf("Scanned only %d lines, was expecting more than 200", scannedCount)
				return
			}
			// Note: when test was written, before changes=92, after=62 (now 50 w/ rematerialization NoXPos)
			// and before sumdiffs=784, after=446 (now 180 w/ rematerialization NoXPos)
			// Split the difference to declare that there's a problem.
			// Also normalize against instruction count in case we unroll loops, etc.
			if changes/scannedCount >= (50+92)/(2*scannedCount) || sumdiffs/scannedCount >= (180+784)/(2*scannedCount) {
				fmt.Printf("Line numbers change too much, # of changes=%.f, sumdiffs=%.f, # of instructions=%.f\n", changes, sumdiffs, scannedCount)
			}
			return
		}
		scannedCount++
		i += lbln
		lineVal, err := strconv.Atoi(line[i : i+3])
		if err != nil {
			fmt.Printf("Expected 3-digit line number after %s in %s\n", beforeLineNumber, line)
		}
		if prevVal == 0 {
			prevVal = lineVal
		}
		diff := lineVal - prevVal
		if diff < 0 {
			diff = -diff
		}
		if diff != 0 {
			changes++
			sumdiffs += float64(diff)
		}
		// If things change too much, uncomment the line below to figure out what's up.
		// The "before" behavior can be recreated DebugFriendlySetPosFrom (currently in gc/ssa.go)
		// by inserting unconditional
		//   	s.SetPos(v.Pos)
		// at the top of the function.

		if debug {
			fmt.Printf("%d %.f %.f %s\n", lineVal, changes, sumdiffs, line)
		}
		prevVal = lineVal
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Reading standard input:", err)
		return
	}
}
