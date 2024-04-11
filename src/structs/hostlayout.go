// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package structs defines types used as directives
// when they appear as field types within a structure.
package structs

// HostLayout, as a field type, signals that the size, alignment,
// and order of fields conform to requirements of the host
// platform and may not match the Go compiler’s defaults.
type HostLayout struct{}
