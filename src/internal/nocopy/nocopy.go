// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nocopy

// NoCopy must be embedded into structs which must not be copied
// after first use.
//
// See https://github.com/golang/go/issues/8005#issuecomment-190753527
// for details.
type NoCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*NoCopy) Lock() {}
