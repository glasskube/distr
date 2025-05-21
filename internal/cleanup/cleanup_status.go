package cleanup

import (
	"context"

	internalctx "github.com/glasskube/distr/internal/context"
)

func RunStatusCleanup(ctx context.Context) error {
	log := internalctx.GetLogger(ctx)
	log.Info("running status cleanup")
	return nil
}
