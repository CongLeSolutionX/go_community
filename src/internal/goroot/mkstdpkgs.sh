#!/usr/bin/env bash
# Copyright 2022 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

echo 'package goroot

import (
	"strings"
)

var std = strings.Fields(`' > stdpkgs.go
go list std cmd >> stdpkgs.go
echo '`)
' >> stdpkgs.go
