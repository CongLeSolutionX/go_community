//go:build !cmd_go_bootstrap && !unix && !windows

package telemetrystats

import "cmd/internal/telemetry"

func incrementVersionCounters() {
	telemetry.Inc("go/platform:version-not-supported")
}
