package cleanup

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"go.uber.org/zap"
)

func RunDeploymentLogRecordCleanup(ctx context.Context) error {
	log := internalctx.GetLogger(ctx)
	count, err := db.CleanupDeploymentLogRecords(ctx)
	log.Info("DeploymentLogRecord cleanup finished", zap.Int64("rowsDeleted", count), zap.Error(err))
	return err
}
