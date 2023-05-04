package cgotest

import (
	"testing"

	"cmd/cgo/testdata/test/issue9026"
)

func test9026(t *testing.T) { issue9026.Test(t) }
