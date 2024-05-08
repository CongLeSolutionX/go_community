// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package telemetrycmd implements the "go telemetry" command.
package telemetrycmd

import (
	"cmd/go/internal/base"
	"cmd/internal/telemetry"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var CmdTelemetry = &base.Command{
	UsageLine: "go telemetry",
	Short:     "manage telemetry data and settings",
	Long: `Telemetry is used to manage Go telemetry data and settings.

See https://go.dev/doc/telemetry for more information on telemetry.
`,

	Commands: []*base.Command{
		cmdOn,
		cmdLocal,
		cmdOff,
		cmdConfig,
		cmdClean,
	},
}

var cmdOn = &base.Command{
	UsageLine: "go telemetry on",
	Short:     "enable telemetry collection and uploading",
	Long: `On enables telemetry collection and uploading.

When telemetry is enabled, telemetry data is written to the local file system
and periodically sent to https://telemetry.go.dev/. Uploaded data is used to
help improve the Go toolchain and related tools, and it will be published as
part of a public dataset.

For more details, see https://telemetry.go.dev/privacy.
This data is collected in accordance with the Google Privacy Policy
(https://policies.google.com/privacy).

To disable telemetry uploading, but keep local data collection, run
“go telemetry local”.
To disable both collection and uploading, run “go telemetry off“.
`,
	Run: runOn,
}

func runOn(ctx context.Context, cmd *base.Command, args []string) {
	if old := telemetry.Mode(); old == "on" {
		return
	}
	if err := telemetry.SetMode("on"); err != nil {
		base.Fatalf("go: failed to enable telemetry: %v", err)
	}
	// We could perhaps only show the telemetry on message when the mode goes
	// from off->on (i.e. check the previous state before calling setMode),
	// but that seems like an unnecessary optimization.
	fmt.Fprintln(os.Stderr, telemetryOnMessage())
}

func telemetryOnMessage() string {
	return `Telemetry uploading is now enabled and data will be periodically sent to
https://telemetry.go.dev/. Uploaded data is used to help improve the Go
toolchain and related tools, and it will be published as part of a public
dataset.

For more details, see https://telemetry.go.dev/privacy.
This data is collected in accordance with the Google Privacy Policy
(https://policies.google.com/privacy).

To disable telemetry uploading, but keep local data collection, run
“gotelemetry local”.
To disable both collection and uploading, run “gotelemetry off“.`
}

var cmdLocal = &base.Command{
	UsageLine: "go telemetry local",
	Short:     "enable telemetry collection but disable uploading",
	Long: `Local enables telemetry collection but not uploading.

When telemetry is in local mode, counter data is written to the local file
system, but will not be uploaded to remote servers.

To enable telemetry uploading, run “go telemetry on”.
To disable both collection and uploading, run “go telemetry off”
`,
	Run: runLocal,
}

func runLocal(ctx context.Context, cmd *base.Command, args []string) {
	if old := telemetry.Mode(); old == "local" {
		return
	}
	if err := telemetry.SetMode("local"); err != nil {
		base.Fatalf("go: failed to set the telemetry mode to local: %v", err)
	}
}

var cmdOff = &base.Command{
	UsageLine: "go telemetry off",
	Short:     "disable telemetry collection and uploading",
	Long: `Off disables telemetry collection and uploading.

When telemetry is disabled, local counter data is neither collected nor uploaded.

To enable local collection (but not uploading) of telemetry data, run
“go telemetry local“.
To enable both collection and uploading, run “gotelemetry on”.
`,
	Run: runOff,
}

func runOff(ctx context.Context, cmd *base.Command, args []string) {
	if old := telemetry.Mode(); old == "off" {
		return
	}
	if err := telemetry.SetMode("off"); err != nil {
		base.Fatalf("Failed to disable telemetry: %v", err)
	}
}

var cmdConfig = &base.Command{
	UsageLine: "go telemetry config",
	Short:     "print the current telemetry configuration",
	Long: `Config prints the current telemetry configuration.

The configuration consists of the mode, the time the mode was effective, and
the directory telemetry data is written to. If the -json flag is provided
it will be written in JSON format.
`,
	Run: runConfig,
}

var configJson = cmdConfig.Flag.Bool("json", false, "output config as JSON")

type config struct {
	Mode  string
	Start time.Time
	Dir   string
}

func runConfig(ctx context.Context, cmd *base.Command, args []string) {
	config := config{
		Mode:  telemetry.Mode(),
		Start: telemetry.ModeEffective(),
		Dir:   telemetry.Dir(),
	}
	if *configJson {
		b, err := json.MarshalIndent(config, "", "\t")
		if err != nil {
			base.Fatalf("%s", err)
		}
		fmt.Println(string(b))
		return
	}
	fmt.Println("mode: %s", config.Mode)
	fmt.Println("start: %v", config.Start)
	fmt.Println("dir: %s", config.Dir)
}

var cmdClean = &base.Command{
	UsageLine: "go telemetry clean",
	Short:     "remove all local telemetry data",
	Long: `Clean removes locally collected counters and reports.

Removing counter files that are currently in use may fail on some operating
systems.

go telemetry clean does not affect the current telemetry mode.
`,
	Run: runClean,
}

func runClean(ctx context.Context, cmd *base.Command, args []string) {
	// For now, be careful to only remove counter files and reports.
	// It would probably be OK to just remove everything, but it may
	// be useful to preserve the weekends file.
	telemetryDir := telemetry.Dir()
	for dir, suffixes := range map[string][]string{
		filepath.Join(telemetryDir, "local"):  {".v1.count", ".json"},
		filepath.Join(telemetryDir, "upload"): {".json"},
	} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				base.Errorf("go: failed to read telemetry dir: %v", err)
			}
			continue
		}
		for _, entry := range entries {
			// TODO: use slices.ContainsFunc once it is available in all supported Go
			// versions.
			remove := false
			for _, suffix := range suffixes {
				if strings.HasSuffix(entry.Name(), suffix) {
					remove = true
					break
				}
			}
			if remove {
				path := filepath.Join(dir, entry.Name())
				if err := os.Remove(path); err != nil {
					base.Errorf("go: failed to remove %s: %v", path, err)
				}
			}
		}
	}
}
