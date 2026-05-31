//go:build embedweb

package server

import (
	"embed"
	"io/fs"
)

//go:embed web/dist
var embeddedWeb embed.FS

func embeddedWebFS() (fs.FS, bool) {
	webFiles, err := fs.Sub(embeddedWeb, "web/dist")
	return webFiles, err == nil
}
