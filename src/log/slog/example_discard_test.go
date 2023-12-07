// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slog_test

import (
	"context"
	"log/slog"
)

type discardHandler struct {
	slog.JSONHandler
}

func (d *discardHandler) Enabled(context.Context, slog.Level) bool { return false }

func Example_discardLogs() {
	l := slog.New(&discardHandler{})
	l.Error("this message will not be logged anywhere")
	// Output:
}
