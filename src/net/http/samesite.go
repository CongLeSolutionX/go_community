package http

import (
	"regexp"
	"strconv"
)

// ClientIsSameSiteIncompatible checks the UserAgent for potential SameSite
// incompatible browsers
// Based on psuedo code: https://www.chromium.org/updates/same-site/incompatible-clients
func ClientIsSameSiteIncompatible(useragent string) bool {
	return hasWebKitSameSiteBug(useragent) ||
		dropsUnrecognizedSameSiteCookies(useragent)
}

func hasWebKitSameSiteBug(useragent string) bool {
	return isIosVersion(12, useragent) ||
		(isMacosxVersion(10, 14, useragent) &&
			(isSafari(useragent) || isMacEmbeddedBrowser(useragent)))
}

func dropsUnrecognizedSameSiteCookies(useragent string) bool {
	if isUcBrowser(useragent) {
		return !isUcBrowserVersionAtLeast(12, 13, 2, useragent)
	}
	return isChromiumBased(useragent) &&
		isChromiumVersionAtLeast(51, useragent) &&
		!isChromiumVersionAtLeast(67, useragent)
}

// Regex parsing of User-Agent string. (See note above!)

func isIosVersion(major int, useragent string) bool {
	valid := regexp.MustCompile(`\(iP.+; CPU .*OS (\d+)[_\d]*.*\) AppleWebKit\/`)

	// Extract digits from first capturing group.
	matches := valid.FindStringSubmatch(useragent)
	if matches == nil {
		return false
	}

	return matches[1] == strconv.Itoa(major)
}

func isMacosxVersion(major int, minor int, useragent string) bool {
	valid := regexp.MustCompile(`\(Macintosh;.*Mac OS X (\d+)_(\d+)[_\d]*.*\) AppleWebKit\/`)
	// Extract digits from first and second capturing groups.
	matches := valid.FindStringSubmatch(useragent)
	if matches == nil {
		return false
	}

	return (matches[1] == strconv.Itoa(major)) &&
		(matches[2] == strconv.Itoa(minor))
}

func isSafari(useragent string) bool {
	valid := regexp.MustCompile(`Version\/.* Safari\/`)
	return valid.MatchString(useragent) &&
		!isChromiumBased(useragent)
}

func isMacEmbeddedBrowser(useragent string) bool {
	valid := regexp.MustCompile(`^Mozilla\/[\.\d]+ \(Macintosh;.*Mac OS X [_\d]+\) AppleWebKit\/[\.\d]+ \(KHTML, like Gecko\)$`)
	return valid.MatchString(useragent)
}

func isChromiumBased(useragent string) bool {
	valid := regexp.MustCompile(`Chrom(e|ium)`)
	return valid.MatchString(useragent)
}

func isChromiumVersionAtLeast(major int, useragent string) bool {
	valid := regexp.MustCompile(`Chrom[^ \/]+\/(\d+)[\.\d]* `)
	// Extract digits from first capturing group.
	matches := valid.FindStringSubmatch(useragent)
	if matches == nil {
		return false
	}
	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}
	return version >= major
}

func isUcBrowser(useragent string) bool {
	valid := regexp.MustCompile(`UCBrowser\/`)
	return valid.MatchString(useragent)
}

func isUcBrowserVersionAtLeast(major int, minor int, build int, useragent string) bool {
	valid := regexp.MustCompile(`UCBrowser\/(\d+)\.(\d+)\.(\d+)[\.\d]* `)
	// Extract digits from three capturing groups.
	matches := valid.FindStringSubmatch(useragent)
	if matches == nil {
		return false
	}

	majorVersion, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}

	minorVersion, err := strconv.Atoi(matches[2])
	if err != nil {
		return false
	}

	buildVersion, err := strconv.Atoi(matches[3])
	if err != nil {
		return false
	}

	if majorVersion != major {
		return majorVersion > major
	}
	if minorVersion != minor {
		return minorVersion > minor
	}
	return buildVersion >= build
}
