package version

var (
	// VERSION should be updated by hand at each release
	VERSION = "0.4.0"

	// GITCOMMIT will be overwritten automatically by the build system
	GITCOMMIT = "HEAD"

	// BUILDTIME will be overwritten automatically by the build system
	BUILDTIME = ""

	// SHOWWARNING might be overwritten by the build system to not show the warning
	SHOWWARNING = "true"
)

// ShowWarning returns wether the warning should be printed or not
func ShowWarning() bool {
	return SHOWWARNING != "false"
}
