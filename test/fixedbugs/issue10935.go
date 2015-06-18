// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Stat() would fail with File not Found if called against a deduped file.
// This may occur on a local NTFS volume, or over SMB file shares that serve
// deduplicated volumes.

package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: issue10935.exe deduped_file\n")
		return
	}

	dedupedFile := os.Args[1]
	fi, err := os.Stat(dedupedFile)

	if err != nil {
		fmt.Printf("Error in os.Stat: %v\n", err)
		return
	}

	mode := fi.Mode()
	if mode&os.ModeSymlink != 0 {
		fmt.Printf("Error: File '%v' is considered a symlink, and should not be.\n", fi.Name())
		return
	}

	if !mode.IsRegular() {
		fmt.Printf("Error: File '%v' should be considered Regular, but has mode %v.\n", fi.Name(), mode)
		return
	}

	fmt.Printf("Success. File '%v' is regular with mode %v.\n", fi.Name(), mode)
}
