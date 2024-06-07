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

# Don't collide with any gomote groups the user is using
export GOMOTE_GROUP=

pushGo() {
    gomote push $VM
}

buildGo() {
    gomote run $VM go/src/make.bash
}

setupSweet() {
    # On the perf gomotes, x/benchmarks is already checked out
    #gomote run -system $VM git clone https://go.googlesource.com/benchmarks

    gomote run -dir benchmarks/sweet $VM ./go/bin/go build ./cmd/sweet
    gomote run $VM ./benchmarks/sweet/sweet get
}

runSweet() {
    VMWD=$(gomote run -system $VM pwd)

    config=$(mktemp)
    cat >>$config <<EOF
[[config]]
  name = "gcrate"
  goroot = "$VMWD/go"
  envexec = ["GODEBUG=gcratetrace=1"]

[[config]]
  name = "trace"
  goroot = "$VMWD/go"
  diagnostics = ["trace"]
EOF
    gomote put $VM $config benchmarks/sweet/config.toml

    gomote run $VM ./benchmarks/sweet/sweet run -shell -count 1 config.toml

    # List files for reference
    gomote run -system $VM ls -lR benchmarks/sweet/results
}

# Run gclab on Sweet traces and collect results
runGCLab() {
    gomote run -system $VM sh -c './go/bin/go build cmd/gclab && cd benchmarks/sweet/results && FILES="$(find . -wholename "*/trace.debug/*.trace")" && for f in $FILES; do echo $f; ../../../gclab $f > $f.gclab 2>&1 || true; done'
}

# Fetch sweet results (excluding traces). If gclab has run, this will
# include results from it.
fetchSweetResults() {
    # Times out
    #gomote gettar -dir ./benchmarks/sweet/results $VM

    # Fetch everything except raw traces.
    gomote run -system $VM tar cz --exclude='trace.debug' benchmarks/sweet/results > $VM-$(date +%Y%m%d_%H%M%S).tar.gz
}

# XXX Rerun all of this with limited GOMAXPROCS

pushGo
buildGo
setupSweet
runSweet
runGCLab
fetchSweetResults
