package db

import (
	"context"
	"fmt"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func GetApplications(ctx context.Context) ([]types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "select * from Application"); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func CreateApplication(ctx context.Context, appliation *types.Application) error {
	db := internalctx.GetDbOrPanic(ctx)
	row := db.QueryRow(ctx,
		"INSERT INTO Application (name, type) VALUES (@name, @type) RETURNING id",
		pgx.NamedArgs{"name": appliation.Name, "type": appliation.Type})
	if err := row.Scan(&appliation.ID); err != nil {
		return fmt.Errorf("could not save application: %w", err)
	}
	return nil
}

func UpdateApplication(ctx context.Context, application *types.Application) error {
	db := internalctx.GetDbOrPanic(ctx)
	rows, err := db.Query(ctx,
		"UPDATE Application SET name = @name WHERE id = @id RETURNING *",
		pgx.NamedArgs{"id": application.ID, "name": application.Name})
	if err != nil {
		return fmt.Errorf("could not update application: %w", err)
	} else if updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[types.Application]); err != nil {
		return fmt.Errorf("could not get updated application: %w", err)
	} else {
		*application = updated
		return nil
	}
}

func GetApplication(ctx context.Context, id string) (*types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "select * from Application where id = @id", pgx.NamedArgs{"id": id}); err != nil {
		return nil, fmt.Errorf("failed to query application: %w", err)
	} else if application, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.Application]); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	} else {
		return &application, nil
	}
}
