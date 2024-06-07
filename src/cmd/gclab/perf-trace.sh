#!/bin/sh

if [[ x"$VM" = x ]]; then
  echo "Run one of:"
  echo
  echo "  export VM=\$(gomote create gotip-linux-amd64_c2s16-perf_vs_release)"
  echo
  echo "  export VM=\$(gomote create gotip-linux-amd64_c3h88-perf_vs_release)"
  exit 1
fi

set -ex

gomote push $VM
gomote run $VM go/src/make.bash

gomote run -system $VM git clone https://go.googlesource.com/benchmarks
gomote run -dir benchmarks $VM git fetch https://go.googlesource.com/benchmarks refs/changes/70/600070/2
gomote run -dir benchmarks $VM git checkout FETCH_HEAD

gomote run -dir benchmarks/sweet $VM ./go/bin/go build ./cmd/sweet
gomote run $VM ./benchmarks/sweet/sweet get

VMWD=$(gomote run -system $VM pwd)

config=$(mktemp)
cat >>$config <<EOF
[[config]]
  name = "trace"
  goroot = "$VMWD/go"
  diagnostics = ["trace"]

#[[config]]
#  name = "gcrate"
#  goroot = "$VMWD/go"
#  envexec = ["GODEBUG=gcratetrace=1"]
EOF
gomote put $VM $config benchmarks/sweet/config.toml

gomote run $VM ./benchmarks/sweet/sweet run -shell -count 1 config.toml

# Times out
#gomote gettar -dir ./benchmarks/sweet/results $VM
gomote run -system $VM tar cz benchmarks/sweet/results > $VM.tar.gz
