#!/usr/bin/env bash
# Copyright 2022 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

if [ ! -f make.bash ]; then
	echo 'cover.bash must be run from $GOROOT/src' 1>&2
	exit 1
fi

bash make.bash
if [ $? != 0 ]; then
  echo "make.bash failed, not proceeding"
  exit 1
fi

set -e

go test -cover -coverpkg all all
