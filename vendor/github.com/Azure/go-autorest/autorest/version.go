package autorest

import (
	"fmt"
	"strings"
	"sync"
)

const (
	major = 7
	minor = 3
	patch = 1
	tag   = ""
)

var versionLock sync.Once
var version string

// Version returns the semantic version (see http://semver.org).
func Version() string {
	versionLock.Do(func() {
		version = fmt.Sprintf("v%d.%d.%d", major, minor, patch)

		if trimmed := strings.TrimPrefix(tag, "-"); trimmed != "" {
			version = fmt.Sprintf("%s-%s", version, trimmed)
		}
	})
	return version
}
