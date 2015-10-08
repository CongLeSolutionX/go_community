// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(mwhudson): document

#ifdef GOARCH_ppc64
#define FIXED_FRAME 8
#endif

#ifdef GOARCH_ppc64le
#define FIXED_FRAME 8
#endif
