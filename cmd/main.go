package main

import (
	"fmt"
	"net/http"

	"github.com/glasskube/cloud/internal/frontend"
	"github.com/go-chi/chi/v5"
)

func main() {
	router := chi.NewRouter()
	router.Mount("/api", ApiRouter())
	router.Handle("/*", http.FileServerFS(frontend.BrowserFS()))

	addr := ":8080"
	fmt.Printf("listen on %v", addr)
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
