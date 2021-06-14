// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// go mod initwork

package modcmd

import (
	"cmd/go/internal/base"
	"cmd/go/internal/modload"
	"context"
	"path/filepath"
)

var cmdInitwork = &base.Command{
	UsageLine: "go mod initwork [moddirs]",
	Short:     "initialize workspace file",
	Long:      `todo`,
	Run:       runInitwork,
}

func init() {
	base.AddModCommonFlags(&cmdInitwork.Flag)
}

func runInitwork(ctx context.Context, cmd *base.Command, args []string) {
	modload.ForceUseModules = true
	modload.WorkspacesEnabled = true

	// TODO(matloob): support using the -workfile path
	// To do that properly, we'll have to make the module directories
	// make dirs relative to workFile path before adding the paths to
	// the directory entries

	workFile := filepath.Join(base.Cwd(), "go.work")

	modload.CreateWorkFile(ctx, workFile, args)
}
