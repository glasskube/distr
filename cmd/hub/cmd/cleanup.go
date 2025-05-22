package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/glasskube/distr/internal/cleanup"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	deploymentTargetStatus   = "DeploymentTargetStatus"
	deploymentTargetMetrics  = "DeploymentTargetMetrics"
	deploymentRevisionStatus = "DeploymentRevisionStatus"
)

type CleanupOptions struct{ Type string }

var CleanupCommand = &cobra.Command{
	Use: "cleanup <type>",
	Long: fmt.Sprintf("type must be one of: %v, %v, %v",
		deploymentTargetStatus, deploymentRevisionStatus, deploymentTargetMetrics),
	Short:     "delete old data",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []cobra.Completion{deploymentTargetStatus, deploymentRevisionStatus},
	PreRun:    func(cmd *cobra.Command, args []string) { env.Initialize() },
	Run: func(cmd *cobra.Command, args []string) {
		runCleanup(cmd.Context(), CleanupOptions{Type: args[0]})
	},
}

func init() {
	RootCommand.AddCommand(CleanupCommand)
}

func runCleanup(ctx context.Context, opts CleanupOptions) {
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()
	log := registry.GetLogger()

	var cleanupFunc func(context.Context) error
	switch opts.Type {
	case deploymentTargetStatus:
		cleanupFunc = cleanup.RunDeploymentTargetStatusCleanup
	case deploymentRevisionStatus:
		cleanupFunc = cleanup.RunDeploymentRevisionStatusCleanup
	case deploymentTargetMetrics:
		cleanupFunc = cleanup.RunDeploymentTargetMetricsCleanup
	default:
		log.Sugar().Errorf("invalid cleanup type: %v", opts.Type)
		os.Exit(1)
	}

	ctx = internalctx.WithDb(ctx, registry.GetDbPool())
	ctx = internalctx.WithLogger(ctx, log)

	if err := cleanupFunc(ctx); err != nil {
		log.Error("cleanup failed", zap.Error(err))
		os.Exit(1)
	}
}
