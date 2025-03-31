package dashboard

import (
	"embed"
	"io/fs"
)

//go:embed dist
var assets embed.FS

// WebUI contains the Web UI assets.
var WebUI, _ = fs.Sub(assets, "dist")
