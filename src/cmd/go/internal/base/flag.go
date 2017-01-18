// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import "cmd/go/internal/str"

type StringsFlag []string

func (v *StringsFlag) Set(s string) error {
	var err error
	*v, err = str.SplitQuotedFields(s)
	if *v == nil {
		*v = []string{}
	}
	return err
}

func (v *StringsFlag) String() string {
	return "<StringsFlag>"
}
