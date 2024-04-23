// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build plan9

package os

import (
	"internal/filepathlite"
)

func checkPathEscapes(_ *Root, name string) error {
	if !filepathlite.IsLocal(name) {
		return errPathEscapes
	}
	return nil
}
