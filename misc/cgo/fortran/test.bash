#!/usr/bin/env bash
# Copyright 2016 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# This directory is intended to test the use of Fortran with cgo.

set -e

FC=$1

goos=$(go env GOOS)

libext="so"
if [ "$goos" == "darwin" ]; then
	libext="dylib"
fi

case "$FC" in
*gfortran*)
  export LIBRARY_PATH=$(dirname $("$FC" -print-file-name=libgfortran.$libext))
  ;;
esac

if [[ "$goos" != "darwin" && "$goos" != "linux" ]]; then
	export CGO_LDFLAGS="$CGO_LDFLAGS -rpath $LIBRARY_PATH -L $LIBRARY_PATH"
fi

if ! $FC helloworld/helloworld.f90 -o main.exe >& /dev/null; then
  echo "skipping Fortran test: could not build helloworld.f90 with $FC"
  exit 0
fi
rm -f main.exe

status=0

if ! go test; then
  echo "FAIL: go test"
  status=1
fi

exit $status
