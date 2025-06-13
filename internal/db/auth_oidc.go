package db

import (
	"context"
	"fmt"
	"time"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func CreateOIDCState(ctx context.Context) (uuid.UUID, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, "INSERT INTO OIDCState DEFAULT VALUES RETURNING id")
	if err != nil {
		return uuid.Nil, err
	}
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
}

func DeleteOIDCState(ctx context.Context, id uuid.UUID) (time.Time, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, "DELETE FROM OIDCState WHERE id = @id RETURNING created_at", pgx.NamedArgs{"id": id})
	if err != nil {
		var none time.Time
		return none, err
	}
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[time.Time])
}

func CleanupOIDCStates(ctx context.Context) (int64, error) {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM OIDCState WHERE current_timestamp - created_at > @maxAge`,
		pgx.NamedArgs{"maxAge": 1 * time.Minute},
	)
	if err != nil {
		return 0, fmt.Errorf("error cleaning up OIDCState: %w", err)
	} else {
		return cmd.RowsAffected(), nil
	}
}
