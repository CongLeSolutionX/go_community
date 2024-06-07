#!/bin/sh

set -ex

go install cmd/compile
go build runtime
TRACE=$(mktemp trace.XXXXXXXXXX)
go build -gcflags=-traceprofile=$TRACE runtime
mv $TRACE ./trace
ls -lh ./trace
