package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

func CreateApplication(ctx context.Context, application *types.Application) error {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return err
	}

	db := internalctx.GetDb(ctx)
	row := db.QueryRow(ctx,
		"INSERT INTO Application (name, type, organization_id) VALUES (@name, @type, @orgId) RETURNING id, created_at",
		pgx.NamedArgs{"name": application.Name, "type": application.Type, "orgId": orgId})
	if err := row.Scan(&application.ID, &application.CreatedAt); err != nil {
		return fmt.Errorf("could not save application: %w", err)
	}
	return nil
}

func UpdateApplication(ctx context.Context, application *types.Application) error {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return err
	}

	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE Application SET name = @name WHERE id = @id AND organization_id = @orgId RETURNING *",
		pgx.NamedArgs{"id": application.ID, "name": application.Name, "orgId": orgId})
	if err != nil {
		return fmt.Errorf("could not update application: %w", err)
	} else if updated, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.Application]); err != nil {
		return fmt.Errorf("could not get updated application: %w", err)
	} else {
		*application = updated
		return nil
	}
}

func GetApplications(ctx context.Context) ([]types.Application, error) {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return nil, err
	}

	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT
			    a.id,
			    a.created_at,
			    a.name,
			    a.type,
			    coalesce((
			    	SELECT array_agg(row(av.id, av.created_at, av.name) ORDER BY av.created_at DESC)
			    	FROM applicationversion av
			    	WHERE av.application_id = a.id
			    ), array[]::record[]) as versions
			FROM Application a
			WHERE a.organization_id = @orgId
			`, pgx.NamedArgs{"orgId": orgId}); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err :=
		pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func GetApplication(ctx context.Context, id string) (*types.Application, error) {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return nil, err
	}

	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT
			    a.id,
			    a.created_at,
			    a.name,
			    a.type,
			    coalesce((
			    	SELECT array_agg(row(av.id, av.created_at, av.name) ORDER BY av.created_at DESC)
			    	FROM applicationversion av
			    	WHERE av.application_id = a.id
			    ), array[]::record[]) as versions
			FROM Application a
			WHERE a.id = @id AND a.organization_id = @orgId
		`, pgx.NamedArgs{"id": id, "orgId": orgId}); err != nil {
		return nil, fmt.Errorf("failed to query application: %w", err)
	} else if application, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Application]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	} else {
		return &application, nil
	}
}

func CreateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{
		"name":          applicationVersion.Name,
		"applicationId": applicationVersion.ApplicationId,
	}
	if applicationVersion.ComposeFileData != nil {
		args["composeFileData"] = *applicationVersion.ComposeFileData
	}
	row := db.QueryRow(ctx,
		`INSERT INTO ApplicationVersion (name, application_id, compose_file_data)
					VALUES (@name, @applicationId, @composeFileData::bytea) RETURNING id, created_at`, args)
	if err := row.Scan(&applicationVersion.ID, &applicationVersion.CreatedAt); err != nil {
		return fmt.Errorf("could not save application: %w", err)
	}
	return nil
}

func UpdateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDb(ctx)
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
	db := internalctx.GetDb(ctx)
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
