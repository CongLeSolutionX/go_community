// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cnet

import (
	"cmd/gclab/heap"
	"fmt"
	"testing"
)

func TestDensityToDot(t *testing.T) {
	layers := DefaultDensityNetworkConfig.makeLayers(16, 256*heap.MiB)
	dot := layersToDot(layers)
	fmt.Print(dot)
}
