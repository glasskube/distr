package main

import (
	"fmt"
	"net/http"

	"github.com/glasskube/cloud/internal/frontend"
)

func main() {
	fileServer := http.FileServerFS(frontend.BrowserFS())
	addr := ":8080"
	fmt.Printf("listen on %v", addr)
	if err := http.ListenAndServe(addr, fileServer); err != nil {
		panic(err)
	}
}
