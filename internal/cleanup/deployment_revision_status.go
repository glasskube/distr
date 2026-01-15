package cleanup

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"go.uber.org/zap"
)

func RunDeploymentRevisionStatusCleanup(ctx context.Context) error {
	log := internalctx.GetLogger(ctx)
	if count, err := db.CleanupDeploymentRevisionStatus(ctx); err != nil {
		return err
	} else {
		log.Info("DeploymentRevisionStatus cleanup finished", zap.Int64("rowsDeleted", count))
		return nil
	}
}
