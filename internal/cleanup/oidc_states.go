package cleanup

import (
	"context"

	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/db"
	"go.uber.org/zap"
)

func RunOIDCStateCleanup(ctx context.Context) error {
	log := internalctx.GetLogger(ctx)
	if count, err := db.CleanupOIDCStates(ctx); err != nil {
		return err
	} else {
		log.Info("OIDCStates cleanup finished", zap.Int64("rowsDeleted", count))
		return nil
	}
}
