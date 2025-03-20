package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/glasskube/distr/internal/buildconfig"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/util"
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

	util.Must(sentry.Init(sentry.ClientOptions{
		Dsn:     env.SentryDSN(),
		Debug:   env.SentryDebug(),
		Release: buildconfig.Version(),
	}))
	defer sentry.Flush(5 * time.Second)
	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().RecoverWithContext(ctx, err)
			panic(err)
		}
	}()

	registry := util.Require(svc.New(ctx, svc.ExecDbMigration(cliOptions.Migrate)))
	defer func() { util.Must(registry.Shutdown()) }()

	util.Must(db.CreateAgentVersion(internalctx.WithDb(ctx, registry.GetDbPool())))

	server := registry.GetServer()
	artifactsServer := registry.GetArtifactsServer()
	go onSigterm(func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		server.Shutdown(ctx)
		artifactsServer.Shutdown(ctx)
		cancel()
	})

	go func() { util.Must(server.Start(":8080")) }()
	go func() { util.Must(artifactsServer.Start(":8585")) }()
	server.WaitForShutdown()
	artifactsServer.WaitForShutdown()
}

func onSigterm(callback func()) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT)
	<-sigint
	callback()
}
