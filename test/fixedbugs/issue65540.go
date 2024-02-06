// compile

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

// This is nested enough that if you have quadratic behavior you
// should take millions of years to compute theses types.

type S0 struct{ a, b S1 }
type S1 struct{ a, b S2 }
type S2 struct{ a, b S3 }
type S3 struct{ a, b S4 }
type S4 struct{ a, b S5 }
type S5 struct{ a, b S6 }
type S6 struct{ a, b S7 }
type S7 struct{ a, b S8 }
type S8 struct{ a, b S9 }
type S9 struct{ a, b S10 }
type S10 struct{ a, b S11 }
type S11 struct{ a, b S12 }
type S12 struct{ a, b S13 }
type S13 struct{ a, b S14 }
type S14 struct{ a, b S15 }
type S15 struct{ a, b S16 }
type S16 struct{ a, b S17 }
type S17 struct{ a, b S18 }
type S18 struct{ a, b S19 }
type S19 struct{ a, b S20 }
type S20 struct{ a, b S21 }
type S21 struct{ a, b S22 }
type S22 struct{ a, b S23 }
type S23 struct{ a, b S24 }
type S24 struct{ a, b S25 }
type S25 struct{ a, b S26 }
type S26 struct{ a, b S27 }
type S27 struct{ a, b S28 }
type S28 struct{ a, b S29 }
type S29 struct{ a, b S30 }
type S30 struct{ a, b S31 }
type S31 struct{ a, b S32 }
type S32 struct{ a, b S33 }
type S33 struct{ a, b S34 }
type S34 struct{ a, b S35 }
type S35 struct{ a, b S36 }
type S36 struct{ a, b S37 }
type S37 struct{ a, b S38 }
type S38 struct{ a, b S39 }
type S39 struct{ a, b S40 }
type S40 struct{ a, b S41 }
type S41 struct{ a, b S42 }
type S42 struct{ a, b S43 }
type S43 struct{ a, b S44 }
type S44 struct{ a, b S45 }
type S45 struct{ a, b int }

func Copy(a, b *S0) {
	// prevent a really smart compiler to be lazy about this
	*a = *b
}
