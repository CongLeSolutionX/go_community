// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package embedlite is a lightweight replacement for embed,
// not importing any package at all. Just like embed, importing the package
// allows using go:embed directives. Unlike embed, only embedding into strings
// or byte slices is allowed.
//
// This restriction is in place since embedlite does not import io/fs.
// This makes it possible to embed single files into Go packages which cannot
// import io/fs due to import cycles, such as time/tzdata.
package embedlite
