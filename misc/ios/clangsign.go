// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// The clangsign binary is a clang wrapper that signs binaries with the
// ldid utility after building. Signing is required when running on
// self-hosted iOS (virtual) devices.
// Use -ldflags="-X main.sdkpath=<path to iPhoneOS.sdk>" when building
// the wrapper.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

var sdkpath = ""

func main() {
	if sdkpath == "" {
		fmt.Fprintf(os.Stderr, "no SDK is set; use -ldflags=\"-X main.sdkpath=<sdk path>\" when building this wrapper.\n")
		os.Exit(1)
	}
	args := os.Args[1:]
	cmd := exec.Command("clang", "-isysroot", "/var/root/sdks/iPhoneOS11.2.sdk", "-mios-version-min=6.0")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			os.Exit(err.ExitCode())
		}
		os.Exit(1)
	}
	for i, a := range args[:len(args)-1] {
		if a != "-o" {
			continue
		}
		// ldid -S will will fail if the file is not an executable, so
		// ignore errors.
		exec.Command("ldid", "-S", args[i+1]).Run()
		break
	}
	os.Exit(0)
}
