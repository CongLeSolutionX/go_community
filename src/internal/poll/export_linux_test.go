// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Export guts for testing on linux.
// Since testing imports os and os imports internal/poll,
// the internal/poll tests can not be in package poll.

package poll

var (
	GetPipe    = getPipe
	PutPipe    = putPipe
	GetPipeFds = func(p *SplicePipe) (int, int) {
		return p.rfd, p.wfd
	}
	NewPipe     = newPipe
	DestroyPipe = destroyPipe
)

type SplicePipe = splicePipe
