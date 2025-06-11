package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/glasskube/distr/internal/buildconfig"
	"github.com/glasskube/distr/internal/cleanup"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/util"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const (
	deploymentTargetStatus   = "DeploymentTargetStatus"
	deploymentTargetMetrics  = "DeploymentTargetMetrics"
	deploymentRevisionStatus = "DeploymentRevisionStatus"
	deploymentLogRecord      = "DeploymentLogRecord"
)

type CleanupOptions struct{ Type string }

var CleanupCommand = &cobra.Command{
	Use: "cleanup <type>",
	Long: fmt.Sprintf(
		"type must be one of: %v, %v, %v, %v",
		deploymentTargetStatus,
		deploymentRevisionStatus,
		deploymentTargetMetrics,
		deploymentLogRecord,
	),
	Short: "delete old data",
	Args:  cobra.ExactArgs(1),
	ValidArgs: []cobra.Completion{
		deploymentTargetStatus,
		deploymentRevisionStatus,
		deploymentTargetMetrics,
		deploymentLogRecord,
	},
	PreRun: func(cmd *cobra.Command, args []string) { env.Initialize() },
	Run: func(cmd *cobra.Command, args []string) {
		runCleanup(cmd.Context(), CleanupOptions{Type: args[0]})
	},
}

func init() {
	RootCommand.AddCommand(CleanupCommand)
}

func runCleanup(ctx context.Context, opts CleanupOptions) {
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown(ctx)) }()
	log := registry.GetLogger()

	var cleanupFunc func(context.Context) error
	switch opts.Type {
	case deploymentTargetStatus:
		cleanupFunc = cleanup.RunDeploymentTargetStatusCleanup
	case deploymentRevisionStatus:
		cleanupFunc = cleanup.RunDeploymentRevisionStatusCleanup
	case deploymentTargetMetrics:
		cleanupFunc = cleanup.RunDeploymentTargetMetricsCleanup
	case deploymentLogRecord:
		cleanupFunc = cleanup.RunDeploymentLogRecordCleanup
	default:
		log.Sugar().Errorf("invalid cleanup type: %v", opts.Type)
		os.Exit(1)
	}

	ctx = internalctx.WithDb(ctx, registry.GetDbPool())
	ctx = internalctx.WithLogger(ctx, log)

	ctx, span := registry.GetTracers().Always().
		Tracer("github.com/glasskube/distr/cmd/hub/cmd", trace.WithInstrumentationVersion(buildconfig.Version())).
		Start(ctx, fmt.Sprintf("cleanup_%v", opts.Type), trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	if err := cleanupFunc(ctx); err != nil {
		log.Error("cleanup failed", zap.Error(err))
		span.SetStatus(codes.Error, "cleanupFunc error")
		span.RecordError(err)
		os.Exit(1)
	}
	span.SetStatus(codes.Ok, "cleanupFunc finished")
}
