//go:build !cmd_go_bootstrap && windows

package telemetrystats

import (
	"fmt"

	"cmd/internal/telemetry"

	"golang.org/x/sys/windows"
)

func incrementVersionCounters() {
	v := windows.RtlGetVersion()
	telemetry.Inc(fmt.Sprintf("go/platform/host/windows/major-version:%d", v.MajorVersion))
	telemetry.Inc(fmt.Sprintf("go/platform/host/windows/version:%d-%d", v.MajorVersion, v.MinorVersion))
	telemetry.Inc(fmt.Sprintf("go/platform/host/windows/build:%d", v.BuildNumber))
}
