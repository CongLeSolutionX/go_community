// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package load

import (
	"errors"
	"fmt"
	"strings"
)

var ErrNotGoDebug = errors.New("not //go:debug line")

func ParseGoDebug(text string) (key, value string, err error) {
	if !strings.HasPrefix(text, "//go:debug") {
		return "", "", ErrNotGoDebug
	}
	i := strings.IndexAny(text, " \t")
	if i < 0 {
		if strings.TrimSpace(text) == "//go:debug" {
			return "", "", fmt.Errorf("missing key=value")
		}
		return "", "", ErrNotGoDebug
	}
	k, v, ok := strings.Cut(strings.TrimSpace(text[i:]), "=")
	if !ok {
		return "", "", fmt.Errorf("missing key=value")
	}
	if strings.Contains(k, " \t") {
		return "", "", fmt.Errorf("key contains space")
	}
	if strings.Contains(k, " \t") {
		return "", "", fmt.Errorf("value contains space")
	}
	if strings.Contains(k, ",") {
		return "", "", fmt.Errorf("key contains comma")
	}
	if strings.Contains(k, ",") {
		return "", "", fmt.Errorf("value contains comma")
	}
	return k, v, nil
}
