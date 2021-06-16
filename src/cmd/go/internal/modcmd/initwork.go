// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// go mod initwork

package modcmd

import (
	"cmd/go/internal/base"
	"cmd/go/internal/modload"
	"cmd/go/internal/search"
	"context"
	"io/fs"
	"log"
	"path/filepath"
)

var cmdInitwork = &base.Command{
	UsageLine: "go mod initwork [moddirs]",
	Short:     "initialize workspace file",
	Long: `Initwork initializes and writes a new go.work file in the current directory, in
effect creating a new workspace at the current directory.

Init optionally accepts paths to the workspace modules as arguments. If the
argument is omitted, an empty workspace with no modules will be created.

See the workspaces design proposal at
https://go.googlesource.com/proposal/+/master/design/45713-workspace.md for
more information.
`,
	Run: runInitwork,
}

func init() {
	base.AddModCommonFlags(&cmdInitwork.Flag)
	base.AddWorkfileFlag(&cmdInitwork.Flag)
}

func runInitwork(ctx context.Context, cmd *base.Command, args []string) {
	modload.InitWorkfile()

	modload.ForceUseModules = true

	// TODO(matloob): support using the -workfile path
	// To do that properly, we'll have to make the module directories
	// make dirs relative to workFile path before adding the paths to
	// the directory entries

	workFile := filepath.Join(base.Cwd(), "go.work")

	// TODO(matloob) standardize paths

	var expandedArgs []string
	hasDotDotDot := false
	for _, arg := range args {
		carg := filepath.ToSlash(filepath.Clean(arg))
		if !search.NewMatch(carg).IsLiteral() {

		}
		if filepath.Base(carg) != "..." {
			if !search.NewMatch(carg).IsLiteral() {
				base.Fatalf("go: initwork does not accept the pattern %q", arg)
			}
			expandedArgs = append(expandedArgs, carg)
			// TODO(matloob): check for go.mod?
			continue
		}
		hasDotDotDot = true
		noDotDotDot := filepath.Dir(carg)
		if !search.NewMatch(noDotDotDot).IsLiteral() {
			base.Fatalf("go: initwork does not accept patterns other than \"...\"; found %q", arg)
		}
		filepath.Walk(noDotDotDot, func(path string, info fs.FileInfo, err error) error {
			if !info.Mode().IsDir() && filepath.Base(path) == "go.mod" {
				expandedArgs = append(expandedArgs, filepath.ToSlash(filepath.Clean(filepath.Dir(path))))
			}

			return nil
		})
	}

	if hasDotDotDot {
		log.Print("WARNING: '...' PATTERNS ARE *NOT* EXPANDED IN THE WORKSPACES PROPOSAL")
	}

	modload.CreateWorkFile(ctx, workFile, expandedArgs)
}
