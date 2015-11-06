// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build openbsd

package cgo

// We override pthread_create to support PT_TLS.
//go:cgo_export_dynamic pthread_create pthread_create
