//go:build !cmd_go_bootstrap

package telemetrystats

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"cmd/internal/telemetry"

	"golang.org/x/sys/unix"
)

func incrementVersionCounters() {
	convert := func(nullterm []byte) string {
		end := bytes.IndexByte(nullterm, 0)
		if end < 0 {
			end = len(nullterm)
		}
		return string(nullterm[:end])
	}

	var v unix.Utsname
	err := unix.Uname(&v)
	if err != nil {
		telemetry.Inc("go/platform/host/%d/version:unknown")
		return
	}
	major, minor, ok := majorMinor(convert(v.Sysname[:]))
	if !ok {
		telemetry.Inc("go/platform/host/%d/version:unknown")
		return
	}
	telemetry.Inc(fmt.Sprintf("go/platform/host/%d/major-version:%d-%d", runtime.GOOS, major))
	telemetry.Inc(fmt.Sprintf("go/platform/host/%d/version:%d-%d", runtime.GOOS, major, minor))

}

func majorMinor(v string) (string, string, bool) {
	firstDot := strings.Index(v, ".")
	if firstDot < 0 {
		return "", "", false
	}
	major := v[:firstDot]
	v = v[firstDot+1:]
	secondDot := strings.Index(v, ".")
	minor := v[:secondDot]
	return major, minor, true
}
