// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package trace

import (
	"cmd/internal/traceviewer"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"time"
)

var traceStarted int32

func getTraceContext(ctx context.Context) (traceContext, bool) {
	if atomic.LoadInt32(&traceStarted) == 0 {
		return traceContext{}, false
	}
	v := ctx.Value(traceKey{})
	if v == nil {
		return traceContext{}, false
	}
	return v.(traceContext), true
}

// StartSpan starts a trace event with the given name. The Span ends when the returned
// function is called.
func StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	tc, ok := getTraceContext(ctx)
	if !ok {
		return ctx, nil
	}
	childSpan := &Span{t: tc.t, name: name, start: time.Now()}
	tc.t.writeEvent(&traceviewer.Event{
		Name:  childSpan.name,
		Time:  float64(childSpan.start.UnixNano()) / float64(time.Microsecond),
		Phase: "B",
	})
	ctx = context.WithValue(ctx, traceKey{}, traceContext{tc.t})
	return ctx, childSpan
}

type Span struct {
	t *tracer

	name  string
	start time.Time
	end   time.Time
}

func (s *Span) Done() {
	if s == nil {
		return
	}
	s.end = time.Now()
	s.t.writeEvent(&traceviewer.Event{
		Name:  s.name,
		Time:  float64(s.end.UnixNano()) / float64(time.Microsecond),
		Phase: "E",
	})
}

type tracer struct {
	evch chan *traceviewer.Event
}

func (t *tracer) writeEvent(ev *traceviewer.Event) {
	t.evch <- ev
}

// traceKey is the context key for tracing information. It is unexported to prevent collisions with context keys defined in
// other packages.
type traceKey struct{}

type traceContext struct {
	t *tracer
}

// Start starts a trace which writes to the given file.
func Start(ctx context.Context, file string) (context.Context, func() error, error) {
	atomic.StoreInt32(&traceStarted, 1)
	if file == "" {
		return nil, nil, errors.New("no trace file supplied")
	}
	f, err := os.Create(file)
	if err != nil {
		return nil, nil, err
	}
	f.WriteString("[")
	evch := make(chan *traceviewer.Event)
	errch := make(chan error)
	tf := &traceFile{f, evch, errch}
	go tf.writerGoroutine()
	t := &tracer{evch: evch}
	ctx = context.WithValue(ctx, traceKey{}, traceContext{t: t})
	return ctx, tf.flush, nil
}

type traceFile struct {
	f     *os.File
	evch  chan *traceviewer.Event
	errch chan error
}

func (tf *traceFile) writerGoroutine() {
	first := true
	var errs []error
	enc := json.NewEncoder(tf.f)
	enc.SetEscapeHTML(false)
	for ev := range tf.evch {
		if !first {
			tf.f.WriteString(",")
		}
		first = false
		if err := enc.Encode(ev); err != nil {
			errs = append(errs, err)
		}
	}
	for _, err := range errs {
		tf.errch <- err
	}
	close(tf.errch)
}

func (tf *traceFile) flush() error {
	var rerr error
	close(tf.evch)
	for err := range tf.errch {
		if rerr != nil {
			rerr = err // return first error
		}
	}
	tf.f.WriteString("]")
	err := tf.f.Close()
	if rerr != nil {
		rerr = err
	}
	return rerr
}
