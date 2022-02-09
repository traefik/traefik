package webui

import (
	"embed"
	"io/fs"
)

//go:embed static
var assets embed.FS

// FS contains the web UI assets.
var FS, _ = fs.Sub(assets, "static")
