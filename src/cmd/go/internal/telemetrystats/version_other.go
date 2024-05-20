//go:build !cmd_go_bootstrap && !unix && !windows

package telemetrystats

func incrementVersionCounters() {
	telemetry.Inc("go/platform:version-not-supported")
}
