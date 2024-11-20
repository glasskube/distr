package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func GetApplications(ctx context.Context) ([]types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "select id, name from application"); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func GetApplication(ctx context.Context, id string) (*types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "select id, name from application where id = @id", pgx.NamedArgs{"id": id}); err != nil {
		return nil, fmt.Errorf("failed to query application: %w", err)
	} else if application, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Application]); err != nil {
		if errors.As(err, &sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	} else {
		return &application, nil
	}
}
