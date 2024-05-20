//go:build !unix && !windows

package telemetrystats

func incrementVersionCounters() {
	telemetry.Inc("go/platform:version-not-supported")
}
