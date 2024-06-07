#!/bin/sh

set -ex

go install cmd/compile
go build runtime
go build -gcflags=-traceprofile=/tmp/trace runtime
ls -lh /tmp/trace
