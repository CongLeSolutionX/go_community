// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !purego

package runtime

// enableDIT is a lightweight version of crypto/subtle.enableDIT that we
// redefine since we can't linkname to it from runtime.
func enableDIT()
