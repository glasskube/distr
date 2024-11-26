package db

import (
	"context"
	"fmt"

	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

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
	} else if updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return fmt.Errorf("could not get updated application: %w", err)
	} else {
		*application = updated
		return nil
	}
}

func GetApplications(ctx context.Context) ([]types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, `
			SELECT a.id,
			       a.created_at,
			       a.name,
			       a.type,
			       CASE WHEN count(av.id) > 0
			           THEN array_agg(row(av.id, av.created_at, av.name))
						 END as versions
			FROM Application a
			    LEFT JOIN ApplicationVersion av on a.id = av.application_id
			GROUP BY a.id
			`); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err :=
		pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func GetApplication(ctx context.Context, id string) (*types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, `
			SELECT a.id,
			       a.created_at,
			       a.name,
			       a.type,
			       CASE WHEN count(av.id) > 0
			           THEN array_agg(row(av.id, av.created_at, av.name))
						 END as versions
			FROM Application a
			    LEFT JOIN ApplicationVersion av on a.id = av.application_id
			WHERE a.id = @id
			GROUP BY a.id
		`, pgx.NamedArgs{"id": id}); err != nil {
		return nil, fmt.Errorf("failed to query application: %w", err)
	} else if application, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Application]); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	} else {
		return &application, nil
	}
}

func CreateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDbOrPanic(ctx)
	args := pgx.NamedArgs{
		"name":          applicationVersion.Name,
		"applicationId": applicationVersion.ApplicationId,
	}
	if applicationVersion.ComposeFileData != nil {
		args["composeFileData"] = *applicationVersion.ComposeFileData
	}
	row := db.QueryRow(ctx,
		`INSERT INTO ApplicationVersion (name, application_id, compose_file_data)
					VALUES (@name, @applicationId, @composeFileData::bytea) RETURNING id`, args)
	if err := row.Scan(&applicationVersion.ID); err != nil {
		return fmt.Errorf("could not save application: %w", err)
	}
	return nil
}

func UpdateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDbOrPanic(ctx)
	rows, err := db.Query(ctx,
		"UPDATE ApplicationVersion SET name = @name WHERE id = @id RETURNING *",
		pgx.NamedArgs{"id": applicationVersion.ID, "name": applicationVersion.Name})
	if err != nil {
		return fmt.Errorf("could not update applicationversion: %w", err)
	} else if updated, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[types.ApplicationVersion]); err != nil {
		return fmt.Errorf("could not get updated applicationversion: %w", err)
	} else {
		*applicationVersion = updated
		return nil
	}
}

func GetApplicationVersionComposeFile(ctx context.Context, applicationVersionId string) ([]byte, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "SELECT compose_file_data FROM ApplicationVersion WHERE id = @id", pgx.NamedArgs{
		"id": applicationVersionId,
	}); err != nil {
		return nil, fmt.Errorf("could not get applicationversion: %w", err)
	} else if data, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[[]byte]); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	} else {
		return data, nil
	}
}
