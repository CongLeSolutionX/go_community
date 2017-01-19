package cgotest

/*
#define PI 3.14
*/
import "C"

import "testing"

func test18720(t *testing.T) {
	if C.PI != 3.14 {
		t.Fatalf("expected 3.14, but got %f", C.PI)
	}
}
