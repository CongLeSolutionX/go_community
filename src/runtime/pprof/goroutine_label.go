// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof

import (
	"context"
	"runtime/internal/proflabel"
)

// SetGoroutineLabels sets the context's label set onto the current goroutine.
// This is a lower-level API than Do, which should be used instead when possible.
func SetGoroutineLabels(ctx context.Context) {
	ctxLabels, _ := ctx.Value(labelContextKey{}).(labelMap)
	labels := &proflabel.Labels{}
	// TODO(matloob): We won't need this copy code if we just use a proflabel.Labels
	// in the context.
	for key, value := range ctxLabels {
		labels.List = append(labels.List, proflabel.Label{Key: key, Value: value})
	}
	proflabel.Set(labels)
}

// Do calls f with a copy of the parent context with the
// given labels added to the parent's label map.
// Each key/value pair in labels is inserted into the label map in the
// order provided, overriding any previous value for the same key.
// The augmented label map will be set for the duration of the call to f
// and restored once f returns.
func Do(ctx context.Context, labels LabelSet, f func(context.Context)) {
	defer SetGoroutineLabels(ctx)
	ctx = WithLabels(ctx, labels)
	SetGoroutineLabels(ctx)
	f(ctx)
}
