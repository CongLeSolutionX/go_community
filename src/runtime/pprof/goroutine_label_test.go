// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pprof_test

import (
	"context"
	"reflect"
	"runtime/internal/proflabel"
	"runtime/pprof"
	"testing"
)

func TestSetGoroutineLabels(t *testing.T) {
	sync := make(chan struct{})

	wantLabels := map[string]string{}
	if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Errorf("Expected parent goroutine's profile labels to be empty before test, got %v", gotLabels)
	}
	go func() {
		if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
			t.Errorf("Expected child goroutine's profile labels to be empty before test, got %v", gotLabels)
		}
		sync <- struct{}{}
	}()
	<-sync

	wantLabels = map[string]string{"key": "value"}
	ctx := pprof.WithLabels(context.Background(), pprof.Labels("key", "value"))
	pprof.SetGoroutineLabels(ctx)
	if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Errorf("parent goroutine's profile labels: got %v, want %v", gotLabels, wantLabels)
	}
	go func() {
		if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
			t.Errorf("child goroutine's profile labels: got %v, want %v", gotLabels, wantLabels)
		}
		sync <- struct{}{}
	}()
	<-sync

	wantLabels = map[string]string{}
	ctx = context.Background()
	pprof.SetGoroutineLabels(ctx)
	if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Errorf("Expected parent goroutine's profile labels to be empty, got %v", gotLabels)
	}
	go func() {
		if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
			t.Errorf("Expected child goroutine's profile labels to be empty, got %v", gotLabels)
		}
		sync <- struct{}{}
	}()
	<-sync
}

func TestDo(t *testing.T) {
	wantLabels := map[string]string{}
	if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Errorf("Expected parent goroutine's profile labels to be empty before Do, got %v", gotLabels)
	}

	pprof.Do(context.Background(), pprof.Labels("key1", "value1", "key2", "value2"), func(ctx context.Context) {
		wantLabels := map[string]string{"key1": "value1", "key2": "value2"}
		if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
			t.Errorf("parent goroutine's profile labels: got %v, want %v", gotLabels, wantLabels)
		}

		sync := make(chan struct{})
		go func() {
			wantLabels := map[string]string{"key1": "value1", "key2": "value2"}
			if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
				t.Errorf("child goroutine's profile labels: got %v, want %v", gotLabels, wantLabels)
			}
			sync <- struct{}{}
		}()
		<-sync

	})

	wantLabels = map[string]string{}
	if gotLabels := toMap(proflabel.Get()); !reflect.DeepEqual(gotLabels, wantLabels) {
		t.Errorf("Expected parent goroutine's profile labels to be empty after Do, got %v", gotLabels)
	}
}

func toMap(labels *proflabel.Labels) map[string]string {
	m := make(map[string]string)
	for labels != nil {
		for _, label := range labels.List {
			if _, ok := m[label.Key]; !ok {
				m[label.Key] = label.Value
			}
		}
		labels = labels.Next
	}
	return m
}
