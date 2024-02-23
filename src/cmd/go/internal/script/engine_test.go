// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package script

import (
	"context"
	"testing"
)

func FuzzQuoteArgs(f *testing.F) {
	state, err := NewState(context.Background(), f.TempDir(), nil /* env */)
	if err != nil {
		f.Fatalf("failed to create state: %v", err)
	}

	f.Add("foo")
	f.Add(`"foo"`)
	f.Add(`'foo'`)
	f.Fuzz(func(t *testing.T, s string) {
		quoted := "cmd " + quoteArgs([]string{s})
		cmd, err := parse("file.txt", 42, quoted)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", quoted, err)
		}
		args := expandArgs(state, cmd.rawArgs, nil /* regexpArgs */)

		if want, got := 1, len(args); want != got {
			t.Fatalf("expected %d args, got %d: %s", want, got, args)
		}

		if want, got := s, args[0]; want != got {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}
