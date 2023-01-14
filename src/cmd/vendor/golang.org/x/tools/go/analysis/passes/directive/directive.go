// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package directive defines an Analyzer that checks known Go toolchain directives.
package directive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/internal/analysisutil"
)

const Doc = "check Go toolchain directives such as //go:debug"

var Analyzer = &analysis.Analyzer{
	Name: "directive",
	Doc:  Doc,
	Run:  runDirective,
}

func runDirective(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		checkGoFile(pass, f)
	}
	for _, name := range pass.OtherFiles {
		if err := checkOtherFile(pass, name); err != nil {
			return nil, err
		}
	}
	for _, name := range pass.IgnoredFiles {
		if strings.HasSuffix(name, ".go") {
			f, err := parser.ParseFile(pass.Fset, name, nil, parser.ParseComments)
			if err != nil {
				// Not valid Go source code - not our job to diagnose, so ignore.
				return nil, nil
			}
			checkGoFile(pass, f)
		} else {
			if err := checkOtherFile(pass, name); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func checkGoFile(pass *analysis.Pass, f *ast.File) {
	var check checker
	check.init(pass)

	for _, group := range f.Comments {
		// A +build comment is ignored after or adjoining the package declaration.
		if group.End()+1 >= f.Package {
			check.headerOK = false
		}
		// A //go:build comment is ignored after the package declaration
		// (but adjoining it is OK, in contrast to +build comments).
		if group.Pos() >= f.Package {
			check.headerOK = false
		}

		// Check each line of a //-comment.
		for _, c := range group.List {
			check.comment(c.Slash, c.Text, true)
		}
	}
}

func checkOtherFile(pass *analysis.Pass, filename string) error {
	var check checker
	check.init(pass)

	// We cannot use the Go parser, since is not a Go source file.
	// Read the raw bytes instead.
	content, tf, err := analysisutil.ReadFile(pass.Fset, filename)
	if err != nil {
		return err
	}

	check.headerOK = false
	check.nonGoFile(token.Pos(tf.Base()), string(content))
	return nil
}

type checker struct {
	pass     *analysis.Pass
	headerOK bool // header directives are OK
	inStar   bool // currently in a /* */ comment
}

func (check *checker) init(pass *analysis.Pass) {
	check.pass = pass
	check.headerOK = true
}

func (check *checker) nonGoFile(pos token.Pos, fullText string) {
	// Process each line.
	text := fullText
	inStar := false
	for text != "" {
		i := strings.Index(text, "\n")
		if i < 0 {
			i = len(text)
		} else {
			i++
		}
		offset := len(fullText) - len(text)
		line := text[:i]
		text = text[i:]

		if !inStar && strings.HasPrefix(line, "//") {
			check.comment(pos+token.Pos(offset), line, false)
			continue
		}

		// Skip over, cut out any /* */ comments.
		for {
			line = strings.TrimSpace(line)
			if inStar {
				i := strings.Index(line, "*/")
				if i < 0 {
					line = ""
					break
				}
				line = line[i+len("*/"):]
				inStar = false
				continue
			}
			if strings.HasPrefix(line, "/*") {
				inStar = true
				line = line[len("/*"):]
				continue
			}
			break
		}
		if line != "" {
			// Found non-comment non-blank line.
			// Ends space for valid //go:build comments,
			// but also ends the fraction of the file we can
			// reliably parse. From this point on we might
			// incorrectly flag "comments" inside multiline
			// string constants or anything else (this might
			// not even be a Go program). So stop.
			break
		}
	}
}

func (check *checker) comment(pos token.Pos, line string, isGo bool) {
	if !strings.HasPrefix(line, "//go:") {
		return
	}
	// testing hack: stop at // ERROR
	if i := strings.Index(line, " // ERROR "); i >= 0 {
		line = line[:i]
	}

	verb := line
	if i := strings.IndexFunc(verb, unicode.IsSpace); i >= 0 {
		verb = verb[:i]
		if line[i] != ' ' && line[i] != '\t' && line[i] != '\n' {
			r, _ := utf8.DecodeRuneInString(line[i:])
			check.pass.Reportf(pos, "invalid space %#q in %s directive", r, verb)
		}
	}

	switch verb {
	default:
		// TODO: Use the go language version for the file.
		// If that version is not newer than us, then we can
		// report unknown directives.

	case "//go:build":
		// Ignore. The buildtag analyzer reports misplaced comments.

	case "//go:debug":
		if !check.headerOK || !isGo {
			check.pass.Reportf(pos, "misplaced //go:debug directive")
		}
	}
}
