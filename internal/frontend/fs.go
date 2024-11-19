package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist/cloud-ui/browser/*
var embeddedFsys embed.FS

func BrowserFS() fs.FS {
	if fs, err := fs.Sub(embeddedFsys, "dist/cloud-ui/browser"); err != nil {
		panic(err)
	} else {
		return fs
	}
}
