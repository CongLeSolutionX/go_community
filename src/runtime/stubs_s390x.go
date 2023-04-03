// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

// Called from assembly only; declared for go vet.
func load_g()
func save_g()

// getcallerfp returns the value of the frame pointer register or 0 if not implemented.
func getcallerfp() uintptr { return 0 }
