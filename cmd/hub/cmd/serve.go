package cmd

import (
	"context"
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
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v83"
)

type ServeOptions struct{ Migrate bool }

var serveOpts = ServeOptions{Migrate: true}

var ServeCommand = &cobra.Command{
	Use:    "serve",
	Short:  "run the Distr Hub server",
	Args:   cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) { env.Initialize() },
	Run: func(cmd *cobra.Command, args []string) {
		runServe(cmd.Context(), serveOpts)
	},
}

func init() {
	ServeCommand.Flags().BoolVar(&serveOpts.Migrate, "migrate", serveOpts.Migrate,
		"run database migrations before starting the server")

	RootCommand.AddCommand(ServeCommand)
}

func runServe(ctx context.Context, opts ServeOptions) {
	util.Must(sentry.Init(sentry.ClientOptions{
		Dsn:              env.SentryDSN(),
		Debug:            env.SentryDebug(),
		Environment:      env.SentryEnvironment(),
		EnableTracing:    env.OtelExporterSentryEnabled(),
		TracesSampleRate: 1.0,
		Release:          buildconfig.Version(),
	}))
	defer sentry.Flush(5 * time.Second)
	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().RecoverWithContext(ctx, err)
			panic(err)
		}
	}()

	if key := env.StripeAPIKey(); key != nil {
		stripe.Key = *key
	}

	registry := util.Require(svc.New(ctx, svc.ExecDbMigration(opts.Migrate)))
	defer func() { util.Must(registry.Shutdown(ctx)) }()

	util.Must(db.CreateAgentVersion(internalctx.WithDb(ctx, registry.GetDbPool())))

	server := registry.GetServer()
	artifactsServer := registry.GetArtifactsServer()

	sigCtx, _ := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	context.AfterFunc(sigCtx, func() {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		server.Shutdown(ctx)
		artifactsServer.Shutdown(ctx)
		cancel()
	})

	go func() { util.Must(server.Start(":8080")) }()
	go func() { util.Must(artifactsServer.Start(":8585")) }()
	registry.GetJobsScheduler().Start()
	server.WaitForShutdown()
	artifactsServer.WaitForShutdown()
}
