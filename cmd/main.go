package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/glasskube/cloud/internal/server"
)

func main() {
	ctx := context.Background()

	s, err := server.New(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to init server: %w", err))
	}
	defer func() { _ = s.Shutdown() }()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
		<-sigint
		_ = s.Shutdown()
		os.Exit(0)
	}()

	addr := ":8080"
	s.GetLogger().Sugar().Infof("listen on %v", addr)
	if err := http.ListenAndServe(addr, server.NewRouter(s)); err != nil {
		panic(err)
	}
}
