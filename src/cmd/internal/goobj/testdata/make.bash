#!/usr/bin/env bash

set -e

goos=$(go env GOOS)
goarch=$(go env GOARCH)

function cleanup() {
	rm -f go1.o go2.o
}
trap cleanup EXIT

go tool compile -o go1.o go1.go
go tool compile -o go2.o go2.go

cp go1.o $goos-$goarch-goobj
go tool pack c $goos-$goarch-archive go1.o go2.o
