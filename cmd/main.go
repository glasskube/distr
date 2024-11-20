package main

import (
	"fmt"
	"github.com/glasskube/cloud/internal/server"
	"github.com/go-chi/chi/v5/middleware"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glasskube/cloud/internal/frontend"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := server.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init server: %v\n", err)
		panic(err)
	}
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Mount("/api", server.ApiRouter())
	router.Handle("/*", StaticFileHandler())

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		_ = server.Shutdown()
		fmt.Println("ok bye")
		os.Exit(1)
	}()

	addr := ":8080"
	fmt.Printf("listen on %v\n", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		panic(err)
	}
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
