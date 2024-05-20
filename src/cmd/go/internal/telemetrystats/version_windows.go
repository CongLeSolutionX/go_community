//go:build !cmd_go_bootstrap

package telemetrystats

import "golang.org/x/sys/windows"

func incrementVersionCounters() string {
	v := windows.RtlGetVersion()
	fmt.Printf("%d", v.BuildNumber)
	Inc(fmt.Sprintf("go/platform/host/windows/major-version:%d", v.MajorVersion))
	Inc(fmt.Sprintf("go/platform/host/windows/version:%d-%d", v.MajorVersion, v.MinorVersion))
	Inc(fmt.Sprintf("go/platform/host/windows/build:%d", v.BuildNumber))
}
