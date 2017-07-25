// +build consul

package version

// NOTE we rely on other "version_*.go" files to be lexically after
// "version_base.go" in order for this to get properly overridden. Be careful
// adding new versions and pick a name that will follow "version_base.go".
func init() {
	// The main version number that is being run at the moment.
	Version = "0.8.1"

	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = "dev"
}
