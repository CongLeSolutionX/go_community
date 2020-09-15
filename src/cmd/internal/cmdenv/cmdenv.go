// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cmdenv parses environment variables that identify commands for
// external build tools.
package cmdenv

import (
	"os/exec"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Split splits a command found in an environment variable
// (such as CC or CXX) into an executable path and list of arguments..
// The executable path may itself contain spaces.
// We assume UTF-8 encoding.
func Split(cmd string) (exe string, args []string) {
	// According to
	// https://www.gnu.org/software/make/manual/html_node/Implicit-Variables.html:
	// 	The “name of a program” may also contain some command arguments, but it
	// 	must start with an actual executable program name.
	//
	// No documentation that I (bcmills) could find described any formal mechanism
	// for quoting or escaping paths in other tools that make use of the CC
	// environment variable. Since the typical case is either a single path to a
	// binary (without flags, but possibly including spaces in the path) or a
	// command in $PATH with additional arguments (without spaces in the command
	// name), we arbitrarily split at the first space for which exec.LookPath can
	// find the prefix, and treat the remaining space-separated fields as
	// arguments.

	cmd = strings.TrimSpace(cmd)

	var (
		r        rune
		wasSpace = false
		size     = 0
	)
	for i := 0; i < len(cmd); i += size {
		r, size = utf8.DecodeRuneInString(cmd[i:])
		if !unicode.IsSpace(r) {
			wasSpace = false
			continue
		}
		if wasSpace {
			continue // Already tried a split across an adjacent space.
		}
		wasSpace = true

		exe, args := cmd[:i], strings.TrimSpace(cmd[i+size:])
		if _, err := exec.LookPath(exe); err == nil {
			return exe, strings.Fields(args)
		}
	}

	// No space-separated prefix of cmd was a valid executable.
	//
	// Either cmd does not contain spaces (the common case), or the spaces are all
	// part of the path to the executable, or the executable doesn't exist at all.
	return cmd, nil
}
