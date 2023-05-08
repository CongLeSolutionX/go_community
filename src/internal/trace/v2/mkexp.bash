#!/usr/bin/env bash
# Copyright 2023 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# This script copies this directory to golang.org/x/exp/trace.
# Just point it at an golang.org/x/exp checkout.

set -e
if [ ! -f mkexp.bash ]; then
	echo 'mkexp.bash must be run from $GOROOT/src/internal/trace/v2' 1>&2
	exit 1
fi

if [ "$#" -ne 1 ]; then
    echo 'mkexp.bash expects one argument: a path to a golang.org/x/exp git checkout'
	exit 1
fi

mkdir -p $1/trace
cp -r ./* $1/trace
rm $1/trace/mkexp.bash
mv $1/trace/tools $1/trace/cmd
mv $1/trace/raw $1/trace/internal/raw
find $1/trace -name '*.go' | xargs -- sed -i 's/internal\/trace\/v2/golang.org\/x\/exp\/trace/'
find $1/trace -name '*.go' | xargs -- sed -i 's/golang.org\/x\/exp\/trace\/raw/golang.org\/x\/exp\/trace\/internal\/raw/'
