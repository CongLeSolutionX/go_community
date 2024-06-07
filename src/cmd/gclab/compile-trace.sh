#!/bin/sh

set -ex

PKG=cmd/compile/internal/ssa

go install cmd/compile
go build $PKG
TRACE=$(mktemp trace.XXXXXXXXXX)
go build -gcflags=-traceprofile=$TRACE $PKG
mv $TRACE ./trace
ls -lh ./trace
