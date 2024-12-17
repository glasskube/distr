package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/cloud/internal/env"
	"github.com/glasskube/cloud/internal/migrations"
	"github.com/glasskube/cloud/internal/svc"
	"github.com/glasskube/cloud/internal/util"
	"github.com/spf13/pflag"
)

var cliOptions = struct{ Migrate bool }{
	Migrate: true,
}

func init() {
	pflag.BoolVar(&cliOptions.Migrate, "migrate", cliOptions.Migrate, "run database migrations before starting the server")
	pflag.Parse()
}

func main() {
	ctx := context.Background()

	util.Must(sentry.Init(sentry.ClientOptions{Dsn: env.SentryDSN(), Debug: env.SentryDebug()}))
	defer sentry.Flush(5 * time.Second)
	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().RecoverWithContext(ctx, err)
			panic(err)
		}
	}()

	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()

	if cliOptions.Migrate {
		util.Must(migrations.Up(registry.GetLogger()))
	}

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
