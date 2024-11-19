package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/glasskube/cloud/internal/frontend"
	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()
	router.Mount("/api", ApiRouter())
	router.Handle("/*", StaticFileHandler())

	addr := ":8080"
	fmt.Printf("listen on %v\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		panic(err)
	}
	fmt.Println("ok bye")
}

func ApiRouter() chi.Router {
	router := chi.NewRouter()
	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusAccepted) })
	// TODO: add api routes here
	return router
}

func StaticFileHandler() http.HandlerFunc {
	fsys := frontend.BrowserFS()
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
