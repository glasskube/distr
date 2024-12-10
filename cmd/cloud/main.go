package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glasskube/cloud/internal/svc"
	"github.com/glasskube/cloud/internal/util"
)

func main() {
	ctx := context.Background()
	registry, err := svc.NewDefault(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize application: %w", err))
	}
	defer func() { util.Must(registry.Shutdown()) }()

	server := registry.GetServer()
	go onSigterm(func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		server.Shutdown(ctx)
		cancel()
	})

	util.Must(server.Start(":8080"))
}

func onSigterm(callback func()) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
	<-sigint
	callback()
}
