#!/bin/sh

set -ex

PKG=cmd/compile/internal/ssa

go install cmd/compile
go build $PKG
GODEBUG=gcratetrace=1 go build -gcflags=-env=__UNUSED=$RANDOM $PKG
