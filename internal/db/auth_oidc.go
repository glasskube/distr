package db

import (
	"context"
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
