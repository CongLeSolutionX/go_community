// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// go mod editwork

package modcmd

import (
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/lockedfile"
	"cmd/go/internal/modfetch"
	"cmd/go/internal/modload"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"golang.org/x/mod/modfile"
)

var cmdEditwork = &base.Command{
	UsageLine: "go mod editwork [editing flags] [go.work]",
	Short:     "edit go.work from tools or scripts",
	Long:      `TODO long form documentatino`,
}

var (
	editworkFmt   = cmdEditwork.Flag.Bool("fmt", false, "")
	editworkGo    = cmdEditwork.Flag.String("go", "", "")
	editworkJSON  = cmdEditwork.Flag.Bool("json", false, "")
	editworkPrint = cmdEditwork.Flag.Bool("print", false, "")
	workedits     []func(file *modfile.WorkFile) // edits specified in flags
)

func init() {
	cmdEditwork.Run = runEditwork // break init cycle

	cmdEditwork.Flag.Var(flagFunc(flagEditworkDirectory), "directory", "")
	cmdEditwork.Flag.Var(flagFunc(flagEditworkDropDirectory), "dropdirectory", "")
	cmdEditwork.Flag.Var(flagFunc(flagEditworkReplace), "replace", "")
	cmdEditwork.Flag.Var(flagFunc(flagEditworkDropReplace), "dropreplace", "")

	base.AddModCommonFlags(&cmdEditwork.Flag)
	base.AddBuildFlagsNX(&cmdEditwork.Flag)
}

func runEditwork(ctx context.Context, cmd *base.Command, args []string) {
	modload.WorkspacesEnabled = true

	anyFlags :=
		*editworkGo != "" ||
			*editworkJSON ||
			*editworkPrint ||
			*editworkFmt ||
			len(workedits) > 0

	if !anyFlags {
		base.Fatalf("go mod edit: no flags specified (see 'go help mod edit').")
	}

	if *editworkJSON && *editworkPrint {
		base.Fatalf("go mod edit: cannot use both -json and -print")
	}

	if len(args) > 1 {
		base.Fatalf("go mod edit: too many arguments")
	}
	var gowork string
	if len(args) == 1 {
		gowork = args[0]
	} else {
		gowork = modload.WorkFilePath()
	}

	if *editworkGo != "" {
		if !modfile.GoVersionRE.MatchString(*editworkGo) {
			base.Fatalf(`go mod: invalid -go option; expecting something like "-go %s"`, modload.LatestGoVersion())
		}
	}

	data, err := lockedfile.Read(gowork)
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	workFile, err := modfile.ParseWork(gowork, data, nil)
	if err != nil {
		base.Fatalf("go: errors parsing %s:\n%s", base.ShortPath(gowork), err)
	}

	if *editGo != "" {
		if err := workFile.AddGoStmt(*editGo); err != nil {
			base.Fatalf("go: internal error: %v", err)
		}
	}

	if len(workedits) > 0 {
		for _, edit := range workedits {
			edit(workFile)
		}
	}
	workFile.SortBlocks()
	workFile.Cleanup() // clean file after edits

	if *editJSON {
		editworkPrintJSON(workFile)
		return
	}

	out, err := workFile.Format()
	if err != nil {
		base.Fatalf("go: %v", err)
	}

	if *editPrint {
		os.Stdout.Write(out)
		return
	}

	// Make a best-effort attempt to acquire the side lock, only to exclude
	// previous versions of the 'go' command from making simultaneous edits.
	if unlock, err := modfetch.SideLock(); err == nil {
		defer unlock()
	}

	err = lockedfile.Transform(gowork, func(lockedData []byte) ([]byte, error) {
		if !bytes.Equal(lockedData, data) {
			return nil, errors.New("go.work changed during editing; not overwriting")
		}
		return out, nil
	})
	if err != nil {
		base.Fatalf("go: %v", err)
	}
}

// flagEditworkDirectory implements the -directory flag.
func flagEditworkDirectory(arg string) {
	workedits = append(workedits, func(f *modfile.WorkFile) {
		if err := f.AddDirectory(arg, ""); err != nil {
			base.Fatalf("go mod: -retract=%s: %v", arg, err)
		}
	})
}

// flagEditworkDropDirectory implements the -dropdirectory flag.
func flagEditworkDropDirectory(arg string) {
	workedits = append(workedits, func(f *modfile.WorkFile) {
		if err := f.DropDirectory(arg); err != nil {
			base.Fatalf("go mod: -dropretract=%s: %v", arg, err)
		}
	})
}

// flagReplace implements the -replace flag.
func flagEditworkReplace(arg string) {
	var i int
	if i = strings.Index(arg, "="); i < 0 {
		base.Fatalf("go mod: -replace=%s: need old[@v]=new[@w] (missing =)", arg)
	}
	old, new := strings.TrimSpace(arg[:i]), strings.TrimSpace(arg[i+1:])
	if strings.HasPrefix(new, ">") {
		base.Fatalf("go mod: -replace=%s: separator between old and new is =, not =>", arg)
	}
	oldPath, oldVersion, err := parsePathVersionOptional("old", old, false)
	if err != nil {
		base.Fatalf("go mod: -replace=%s: %v", arg, err)
	}
	newPath, newVersion, err := parsePathVersionOptional("new", new, true)
	if err != nil {
		base.Fatalf("go mod: -replace=%s: %v", arg, err)
	}
	if newPath == new && !modfile.IsDirectoryPath(new) {
		base.Fatalf("go mod: -replace=%s: unversioned new path must be local directory", arg)
	}

	workedits = append(workedits, func(f *modfile.WorkFile) {
		if err := f.AddReplace(oldPath, oldVersion, newPath, newVersion); err != nil {
			base.Fatalf("go mod: -replace=%s: %v", arg, err)
		}
	})
}

// flagDropReplace implements the -dropreplace flag.
func flagEditworkDropReplace(arg string) {
	path, version, err := parsePathVersionOptional("old", arg, true)
	if err != nil {
		base.Fatalf("go mod: -dropreplace=%s: %v", arg, err)
	}
	workedits = append(workedits, func(f *modfile.WorkFile) {
		if err := f.DropReplace(path, version); err != nil {
			base.Fatalf("go mod: -dropreplace=%s: %v", arg, err)
		}
	})
}

// editPrintJSON prints the -json output.
func editworkPrintJSON(workFile *modfile.WorkFile) {
	var f workfileJSON
	if workFile.Go != nil {
		f.Go = workFile.Go.Version
	}
	for _, d := range workFile.Directory {
		f.Directory = append(f.Directory, directoryJSON{DiskPath: d.DiskPath, ModPath: d.ModulePath})
	}

	for _, r := range workFile.Replace {
		f.Replace = append(f.Replace, replaceJSON{r.Old, r.New})
	}
	data, err := json.MarshalIndent(&f, "", "\t")
	if err != nil {
		base.Fatalf("go: internal error: %v", err)
	}
	data = append(data, '\n')
	os.Stdout.Write(data)
}

// workfileJSON is the -json output data structure.
type workfileJSON struct {
	Go        string `json:",omitempty"`
	Directory []directoryJSON
	Replace   []replaceJSON
}

type directoryJSON struct {
	DiskPath string
	ModPath  string `json:",omitempty"`
}
