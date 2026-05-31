//go:build !embedweb

package server

import "io/fs"

func embeddedWebFS() (fs.FS, bool) {
	return nil, false
}
