package webapp

import (
	"embed"
	"io/fs"
)

// dist is populated by scripts/build.ps1 after the frontend production build.
// The checked-in placeholder keeps ordinary Go tooling usable in a clean clone.
//
//go:embed all:dist
var content embed.FS

func FS() (fs.FS, error) {
	return fs.Sub(content, "dist")
}
