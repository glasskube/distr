package db

import (
	"context"
	"errors"
	"fmt"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/multierr"
)

func RunTx(ctx context.Context, txOptions pgx.TxOptions, f func(ctx context.Context) error) (finalErr error) {
	db := internalctx.GetDb(ctx)
	if pool, ok := db.(*pgxpool.Pool); !ok {
		return fmt.Errorf("expected GetDb to return *pgxpool.Pool but got %T instead", db)
	} else if tx, err := pool.BeginTx(ctx, txOptions); err != nil {
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
