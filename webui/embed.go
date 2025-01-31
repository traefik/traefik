package webui

import (
	"embed"
	"io/fs"
)

// Files starting with . and _ are excluded by default
//
//go:embed static
var assets embed.FS

// FS contains the web UI assets.
var FS, _ = fs.Sub(assets, "static")
