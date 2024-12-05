package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glasskube/cloud/internal/server"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/glasskube/cloud/internal/frontend"
	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()

	s, err := server.New(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to init server: %w", err))
	}
	defer func() { _ = s.Shutdown() }()

	router := chi.NewRouter()
	router.Use(
		// Handles panics
		middleware.Recoverer,
		// Reject bodies larger than 1MiB
		middleware.RequestSize(1048576),
	)
	router.Mount("/api", server.ApiRouter(s))
	router.With(
		middleware.Compress(5, "text/html", "text/css", "text/javascript"),
	).Handle("/*", StaticFileHandler())

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		_ = s.Shutdown()
		os.Exit(0)
	}()

	addr := ":8080"
	s.GetLogger().Sugar().Infof("listen on %v", addr)
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
