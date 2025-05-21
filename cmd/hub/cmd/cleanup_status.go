package cmd

import (
	"context"
	"os"

	"github.com/glasskube/distr/internal/cleanup"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var CleanupStatusCommand = &cobra.Command{
	Use:    "status",
	Args:   cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) { env.Initialize() },
	Run: func(cmd *cobra.Command, args []string) {
		runCleanupStatus(cmd.Context())
	},
}

func init() {
	CleanupCommand.AddCommand(CleanupStatusCommand)
}

func runCleanupStatus(ctx context.Context) {
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { util.Must(registry.Shutdown()) }()
	log := registry.GetLogger()
	ctx = internalctx.WithDb(ctx, registry.GetDbPool())
	ctx = internalctx.WithLogger(ctx, log)
	if err := cleanup.RunStatusCleanup(ctx); err != nil {
		log.Error("status cleanup failed", zap.Error(err))
		os.Exit(1)
	}
}
