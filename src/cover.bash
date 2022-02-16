#!/usr/bin/env bash
# Copyright 2013 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

if [ ! -f make.bash ]; then
	echo 'cover.bash must be run from $GOROOT/src' 1>&2
	exit 1
fi
. ./make.bash --no-banner
go install -cover cmd std
# TODOs:
# - fix hard-coded dist test rebuild of tools
# - set GOCOVERDIR prior to run below
go tool dist test -cover 
