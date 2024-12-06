package handlers

import (
	"io/fs"
	"net/http"
)

func StaticFileHandler(fsys fs.FS) http.HandlerFunc {
	server := http.FileServer(http.FS(fsys))
	return func(w http.ResponseWriter, r *http.Request) {
		// check if the requested file exists and use index.html if it does not.
		if _, err := fs.Stat(fsys, r.URL.Path[1:]); err != nil {
			http.StripPrefix(r.URL.Path, server).ServeHTTP(w, r)
		} else {
			server.ServeHTTP(w, r)
		}
	}
}
