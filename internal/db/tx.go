package db

import (
	"context"
	"errors"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/jackc/pgx/v5"
	"go.uber.org/multierr"
)

func RunTx(ctx context.Context, f func(ctx context.Context) error) (finalErr error) {
	db := internalctx.GetDb(ctx)
	if tx, err := db.Begin(ctx); err != nil {
		return err
	} else {
		defer func() {
			// Rollback is safe to call after commit but we have to silence ErrTxClosed
			if err := tx.Rollback(ctx); !errors.Is(err, pgx.ErrTxClosed) {
				multierr.AppendInto(&finalErr, err)
			}
		}()
		if err := f(internalctx.WithDb(ctx, tx)); err != nil {
			return err
		} else {
			return tx.Commit(ctx)
		}
	}
}
