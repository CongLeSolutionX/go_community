package proflabel_test

import (
	"fmt"
	"reflect"
	"runtime/internal/proflabel"
	"testing"
)

func TestProflabelRoundtrip(t *testing.T) {
	labels := &proflabel.Labels{List: []proflabel.Label{{Key: "key", Value: "value"}}}
	proflabel.Set(labels)
	gotLabels := proflabel.Get()
	if !reflect.DeepEqual(labels, gotLabels) {
		t.Errorf("label set on goroutine: got %v, want %v", gotLabels, labels)
	}
}

func TestProflabelChildGoroutine(t *testing.T) {
	labels := &proflabel.Labels{List: []proflabel.Label{{Key: "key", Value: "value"}}}
	proflabel.Set(labels)
	ch := make(chan *proflabel.Labels)
	go func() {
		ch <- proflabel.Get()
		ch <- nil // synchronize
		ch <- proflabel.Get()
	}()

	// check parent labels as baseline
	gotParentLabels := proflabel.Get()
	if !reflect.DeepEqual(labels, gotParentLabels) {
		fmt.Println("labels", labels)
		fmt.Println("got parent labels", gotParentLabels)
		t.Errorf("label set on parent goroutine: got %v, want %v", gotParentLabels, labels)
	}

	// child labels should be the same
	gotChildLabels := <-ch
	if !reflect.DeepEqual(labels, gotChildLabels) {
		t.Errorf("label set on child goroutine: got %v, want %v", gotChildLabels, labels)
	}

	proflabel.Set(nil) // clear parent labels
	<-ch               // synchronize with child goroutine

	// parent labels should be nil after the Set call
	gotParentLabels = proflabel.Get()
	if !reflect.DeepEqual((*proflabel.Labels)(nil), gotParentLabels) {
		t.Errorf("label set on parent goroutine: got %v, want %v", gotParentLabels, nil)
	}
}
