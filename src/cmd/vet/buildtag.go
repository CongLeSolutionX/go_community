// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"
)

var (
	nl         = []byte("\n")
	slashSlash = []byte("//")
	plusBuild  = []byte("+build")
)

// checkBuildTag checks that build tags are in the correct location and well-formed.
func checkBuildTag(f *File) {
	if !vet("buildtags") {
		return
	}
	// badf is like File.Badf, but it uses a line number instead of
	// token.Pos.
	badf := func(line int, format string, args ...interface{}) {
		format = "%s:%d:" + format + "\n"
		args = append([]interface{}{f.name, line}, args...)
		fmt.Fprintf(os.Stderr, format, args...)
		setExit(1)
	}

	// we must look at the raw lines, as build tags may appear in non-Go
	// files such as assembly files.
	lines := bytes.SplitAfter(f.content, nl)

	// lineWithComment records all source lines that contain a //-style
	// comment in a Go source file. If the current source file is not Go,
	// the map is nil.
	var lineWithComment map[int]bool
	if f.file != nil {
		lineWithComment = make(map[int]bool)
		for _, group := range f.file.Comments {
			for _, comment := range group.List {
				line := f.fset.Position(comment.Pos()).Line
				lineWithComment[line] = true
			}
		}
	}

	// Determine cutpoint where +build comments are no longer valid.
	// They are valid in leading // comments in the file followed by
	// a blank line.
	var cutoff int
	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			cutoff = i
			continue
		}
		if bytes.HasPrefix(line, slashSlash) {
			continue
		}
		break
	}

	for i, line := range lines {
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, slashSlash) {
			continue
		}
		if lineWithComment != nil && !lineWithComment[i+1] {
			// This is a line in a Go source file that looks like a
			// comment, but actually isn't - such as part of a raw
			// string.
			continue
		}

		text := bytes.TrimSpace(line[2:])
		if bytes.HasPrefix(text, plusBuild) {
			fields := bytes.Fields(text)
			if !bytes.Equal(fields[0], plusBuild) {
				// Comment is something like +buildasdf not +build.
				badf(i+1, "possible malformed +build comment")
				continue
			}
			if i >= cutoff {
				badf(i+1, "+build comment must appear before package clause and be followed by a blank line")
				continue
			}
			// Check arguments.
		Args:
			for _, arg := range fields[1:] {
				for _, elem := range strings.Split(string(arg), ",") {
					if strings.HasPrefix(elem, "!!") {
						badf(i+1, "invalid double negative in build constraint: %s", arg)
						break Args
					}
					elem = strings.TrimPrefix(elem, "!")
					for _, c := range elem {
						if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' && c != '.' {
							badf(i+1, "invalid non-alphanumeric build constraint: %s", arg)
							break Args
						}
					}
				}
			}
			continue
		}
		// Comment with +build but not at beginning.
		if bytes.Contains(line, plusBuild) && i < cutoff {
			badf(i+1, "possible malformed +build comment")
			continue
		}
	}
}
