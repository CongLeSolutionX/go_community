// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package toolchain

import (
	"context"

	"cmd/go/internal/base"
	"cmd/go/internal/gover"

	"golang.org/x/mod/modfile"
)

// A Reqs collects information about the toolchain version
// required for a given operation.
// After initialization, the requirements are accumulated
// by calling AddGoMod to add a parsed go.mod file
// or calling AddError to add an error returned from
// any function that might have identified a requirement
// and returned a *gover.TooNewError, potentially wrapped.
// Once all the possible requirements have been added,
// calling Switch identifies whether a toolchain switch is
// necessary and does that switch if possible.
// If the switch cannot be performed, then Switch prints
// all the added errors using base.Error and then returns.
type Reqs struct {
	// UseToolchainLines controls whether the
	// Reqs should treat toolchain lines
	// as a minimum requirement for the eventual
	// toolchain switch. Typically it is set during
	// "go work" commands that are processing
	// the direct workspace module roots.
	UseToolchainLines bool

	GoVersion string               // max go requirement observed
	Toolchain string               // max toolchain requirement observed
	Errors    []error              // errors collected so far
	TooNew    []*gover.TooNewError // ToonewErrors (perhaps wrapped) inside errors list
}

// Error reports the given error,
// which is added to r.Errors.
// It is not reported until r.Switch.
func (r *Reqs) AddError(err error) {
	r.Errors = append(r.Errors, err)
	r.addTooNew(err)
}

// addTooNew adds any TooNew errors that can be found in err.
func (r *Reqs) addTooNew(err error) {
	switch err := err.(type) {
	case interface{ Unwrap() []error }:
		for _, e := range err.Unwrap() {
			r.addTooNew(e)
		}

	case interface{ Unwrap() error }:
		r.addTooNew(err.Unwrap())

	case *gover.TooNewError:
		r.GoVersion = gover.Max(r.GoVersion, err.GoVersion)
		if r.UseToolchainLines {
			r.Toolchain = gover.ToolchainMax(r.Toolchain, err.Toolchain)
		}
		r.TooNew = append(r.TooNew, err)
	}
}

// AddGoMod adds the requirements in mf to the running requirements.
func (r *Reqs) AddGoMod(mf *modfile.File) {
	if mf.Go != nil {
		r.GoVersion = gover.Max(r.GoVersion, mf.Go.Version)
	} else {
		r.GoVersion = gover.Max(r.GoVersion, gover.DefaultGoModVersion)
	}

	if r.UseToolchainLines && mf.Toolchain != nil {
		r.Toolchain = gover.ToolchainMax(r.Toolchain, mf.Toolchain.Name)
	}
}

// Switch switches to a newer Go toolchain if any errors or go.mod files
// requiring one have been observed. If it fails to switch, it emits (using base.Error)
// all the errors that have been queued using r.Error.
//
// Switch should almost always be followed by a call to base.ExitIfErrors,
// since it may have emitted many errors that have been delayed until now.
func (r *Reqs) Switch(ctx context.Context) {
	// Switch to newer Go toolchain if necessary and possible.
	v := gover.Max(gover.FromToolchain(r.Toolchain), r.GoVersion)
	if gover.Compare(v, gover.Local()) > 0 {
		TryVersion(ctx, v)
	}

	// Emit any errors now, since the version switch didn't happen.
	for _, err := range r.Errors {
		base.Error(err)
	}
	base.ExitIfErrors()
}

// UpdateGoWork updates a go.work file with the information
// about the running go version and toolchain.
func (r *Reqs) UpdateGoWork(wf *modfile.WorkFile) {
	if !r.UseToolchainLines {
		base.Fatalf("go: internal error: UpdateGoWork without UseToolchainLines")
	}

	goVers := r.GoVersion
	toolchain := r.Toolchain

	if goVers == "" && toolchain == "" {
		return
	}

	// Compute max of input goVers, toolchain and what's in the file.
	oldGo := gover.DefaultGoModVersion
	if wf.Go != nil {
		oldGo = wf.Go.Version
	}
	oldToolchain := "go" + oldGo
	if wf.Toolchain != nil {
		oldToolchain = wf.Toolchain.Name
	}
	if gover.Compare(goVers, oldGo) < 0 {
		goVers = oldGo
	}
	if gover.Compare(gover.FromToolchain(toolchain), gover.FromToolchain(oldToolchain)) < 0 {
		toolchain = oldToolchain
	}
	if gover.Compare(gover.FromToolchain(toolchain), goVers) < 0 {
		toolchain = "go" + goVers
	}
	if goVers == oldGo && toolchain == oldToolchain {
		return
	}

	// We are writing a new go or toolchain line. For reproducibility,
	// if the toolchain running right now is newer than the new toolchain line,
	// update the toolchain line to record the newer toolchain.
	// The user never sets the toolchain explicitly in a 'go work' command,
	// so this is only happening as a result of a go or toolchain line found
	// in a module.
	toolchain = gover.ToolchainMax(toolchain, "go"+gover.Local())

	// Update the file. Drop toolchain if it matches goVers exactly.
	wf.AddGoStmt(goVers)
	if toolchain == "" || toolchain == "go"+goVers {
		wf.DropToolchainStmt()
	} else {
		wf.AddToolchainStmt(toolchain)
	}
}
