package webui

import (
	"io/fs"
	"embed"
)

var assets embed.FS

// FS contains the web UI assets.
var FS, _ = fs.Sub(assets, "static")
