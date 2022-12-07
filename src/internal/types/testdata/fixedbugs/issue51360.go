// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	len.Println /* ERR cannot select on len */
	len.Println /* ERR cannot select on len */ ()
	_ = len.Println /* ERR cannot select on len */
	_ = len /* ERR cannot index len */ [0]
	_ = *len /* ERR cannot indirect len */
}
