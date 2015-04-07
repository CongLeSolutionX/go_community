#!/bin/bash

# Copyright 2015 The Go Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Run this script to update the packages ./exact and ./types
# in the $GOROOT/src/go directory. They are vendored from the
# original sources in x/tools. Imports are renamed as needed.
#
# Delete this script once go/exact and go/types don't exist anymore in x/tools.

### Safety first.
if [ ! -d $GOPATH ]; then
	echo "GOPATH must be set"
	exit 1
fi
if [ ! -d $GOROOT ]; then
	echo "GOROOT must be set"
	exit 1
fi

GODIR=$GOROOT/src/go

function vendor() (
	SRCDIR=$GOPATH/src/golang.org/x/tools/$1
	DSTDIR=$GODIR/$2

	echo "vendoring $SRCDIR => $DSTDIR"

	# create directory
	rm -rf $DSTDIR
	mkdir -p $DSTDIR
	cd $DSTDIR

	# copy go sources and update import paths
	for f in $SRCDIR/*.go; do
		# copy $f and update imports	
		sed -e 's|"golang.org/x/tools/go/exact"|"go/exact"|' \
		    -e 's|"golang.org/x/tools/go/types"|"go/types"|' \
		    -e 's|"golang.org/x/tools/go/gcimporter"|"go/types/internal/gcimporter"|' \
		    $f | gofmt > tmp.go
		mv -f tmp.go `basename $f`
	done

	# copy testdata, if any
	if [ -e $SRCDIR/testdata ]; then
		cp -R $SRCDIR/testdata/ $DSTDIR/testdata/
	fi
)

function install() (
	PKG=$GODIR/$1

	echo "installing $PKG"
	cd $PKG
	go install
)

function test() (
	PKG=$GODIR/$1

	echo "testing $PKG"
	cd $PKG
	go test
	if [ $? -ne 0 ]; then
		echo "TESTING $PKG FAILED"
		exit 1
	fi
)

### go/exact
vendor go/exact exact
test exact
install exact

### go/types
vendor go/types types
# cannot test w/o gcimporter
install types

### go/gcimporter
vendor go/gcimporter types/internal/gcimporter
test types/internal/gcimporter
install types/internal/gcimporter

### test go/types (requires gcimporter)
test types

# All done.
echo "DONE"
exit 0
