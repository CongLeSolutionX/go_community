// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that environment variables are accessible through
// package os.

package main

import "os"

func main() {
	ga := os.Getenv("PATH")
	if ga == "" {
		print("PATH is empty\n")
		os.Exit(1)
	}
	xxx := os.Getenv("DOES_NOT_EXIST")
	if xxx != "" {
		print("$DOES_NOT_EXIST=", xxx, "\n")
		os.Exit(1)
	}
}
