package loadbalancer

import (
	"regexp"
	"strconv"
)

var (
	iosVersionRegex       = regexp.MustCompile(`\(iP.+; CPU .*OS (\d+)[_\d]*.*\) AppleWebKit/`)
	macosVersionRegex     = regexp.MustCompile(`\(Macintosh;.*Mac OS X (\d+)_(\d+)[_\d]*.*\) AppleWebKit/`)
	safariRegex           = regexp.MustCompile(`Version/.* Safari/`)
	macEmbeddedRegex      = regexp.MustCompile(`^Mozilla/[\d.]+ \(Macintosh;.*Mac OS X [\d_]+\) AppleWebKit/[\d.]+ \(KHTML, like Gecko\)$`)
	chromiumRegex         = regexp.MustCompile(`Chrom(e|ium)`)
	chromiumVersionRegex  = regexp.MustCompile(`Chrom[^ /]+/(\d+)`)
	ucBrowserRegex        = regexp.MustCompile(`UCBrowser/`)
	ucBrowserVersionRegex = regexp.MustCompile(`UCBrowser/(\d+)\.(\d+)\.(\d+)`)
)

func shouldSendSameSiteNone(userAgent string) bool {
	return !isSameSiteNoneIncompatible(userAgent)
}

func isSameSiteNoneIncompatible(userAgent string) bool {
	return hasWebKitSameSiteBug(userAgent) || dropsUnrecognizedSameSiteCookies(userAgent)
}

func hasWebKitSameSiteBug(userAgent string) bool {
	if matches := iosVersionRegex.FindStringSubmatch(userAgent); len(matches) == 2 {
		major, err := strconv.Atoi(matches[1])
		return err == nil && major == 12
	}

	matches := macosVersionRegex.FindStringSubmatch(userAgent)
	if len(matches) != 3 {
		return false
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil || major != 10 {
		return false
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil || minor != 14 {
		return false
	}

	return isSafari(userAgent) || macEmbeddedRegex.MatchString(userAgent)
}

func dropsUnrecognizedSameSiteCookies(userAgent string) bool {
	if ucBrowserRegex.MatchString(userAgent) {
		return !isUcBrowserVersionAtLeast(userAgent, 12, 13, 2)
	}

	return chromiumRegex.MatchString(userAgent) && isChromiumVersionAtLeast(userAgent, 51) && !isChromiumVersionAtLeast(userAgent, 67)
}

func isSafari(userAgent string) bool {
	return safariRegex.MatchString(userAgent) && !chromiumRegex.MatchString(userAgent)
}

func isChromiumVersionAtLeast(userAgent string, major int) bool {
	matches := chromiumVersionRegex.FindStringSubmatch(userAgent)
	if len(matches) != 2 {
		return false
	}

	version, err := strconv.Atoi(matches[1])
	return err == nil && version >= major
}

func isUcBrowserVersionAtLeast(userAgent string, major, minor, build int) bool {
	matches := ucBrowserVersionRegex.FindStringSubmatch(userAgent)
	if len(matches) != 4 {
		return false
	}

	majorVersion, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}
	if majorVersion != major {
		return majorVersion > major
	}

	minorVersion, err := strconv.Atoi(matches[2])
	if err != nil {
		return false
	}
	if minorVersion != minor {
		return minorVersion > minor
	}

	buildVersion, err := strconv.Atoi(matches[3])
	return err == nil && buildVersion >= build
}
