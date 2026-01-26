package cleanup

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"go.uber.org/zap"
)

func RunDeploymentTargetLogRecordCleanup(ctx context.Context) error {
	log := internalctx.GetLogger(ctx)
	count, err := db.CleanupDeploymentTargetLogRecords(ctx)
	log.Info("DeploymentTargetLogRecord cleanup finished", zap.Int64("rowsDeleted", count), zap.Error(err))
	return err
}
