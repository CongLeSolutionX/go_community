set -eu

if [[ ! -e "code.json" ]]; then
	cp ../../encoding/json/testdata/code.json.gz .
	gunzip code.json.gz
fi
gotip build -o ./baseline
../../../bin/go build -o ./candidate
export GODEBUG="memprofilerate=$1"
export BASEPREFIX="rate-$1-${2:-}"

echo "BASELINE =========================="
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"
env GODEBUG="${GODEBUG},runtimecontentionstacks" PREFIX="baseline-${BASEPREFIX}"  ./baseline | grep "gcs"

echo "ENABLED =========================="
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs" 
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs" 
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs" 
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs" 
env GODEBUG="${GODEBUG},profileruntimelocks,trackroots=1" PREFIX="candidate-${BASEPREFIX}" ./candidate | grep "gcs" 

echo "DISABLED =========================="
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
env GODEBUG="${GODEBUG},profileruntimelocks" PREFIX="candidate-tracking-disabled-${BASEPREFIX}" ./candidate | grep "gcs"
