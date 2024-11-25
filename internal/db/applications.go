package db

import (
	"context"
	"fmt"
	"time"

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
			select a.id,
			       a.created_at,
			       a.name,
			       a.type,
			       av.id,
			       av.created_at,
			       av.name,
			       av.compose_file_data
			from Application a
			    left join ApplicationVersion av on a.id = av.application_id`); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if joinedStructs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[applicationWithOptionalVersionRow]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return collectApplicationsWithVersions(joinedStructs), nil
	}
}

func GetApplication(ctx context.Context, id string) (*types.Application, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, `
			select a.id,
			       a.created_at,
			       a.name,
			       a.type,
			       av.id,
			       av.created_at,
			       av.name,
			       av.compose_file_data
			from Application a
			    left join ApplicationVersion av on a.id = av.application_id
			where a.id = @id
		`, pgx.NamedArgs{"id": id}); err != nil {
		return nil, fmt.Errorf("failed to query application: %w", err)
	} else if joinedStructs, err := pgx.CollectRows(rows, pgx.RowToStructByPos[applicationWithOptionalVersionRow]); err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	} else if applications := collectApplicationsWithVersions(joinedStructs); len(applications) != 1 {
		if len(applications) == 0 {
			return nil, nil
		} else {
			return nil, pgx.ErrTooManyRows
		}
	} else {
		return &applications[0], nil
	}
}

func collectApplicationsWithVersions(joinedStructs []applicationWithOptionalVersionRow) []types.Application {
	applicationsMap := make(map[string]*types.Application)
	for _, joinedStruct := range joinedStructs {
		if _, ok := applicationsMap[joinedStruct.ApplicationId]; !ok {
			applicationsMap[joinedStruct.ApplicationId] = &types.Application{
				ID:        joinedStruct.ApplicationId,
				CreatedAt: joinedStruct.ApplicationCreatedAt,
				Name:      joinedStruct.ApplicationName,
				Type:      joinedStruct.ApplicationType,
				Versions:  make([]types.ApplicationVersion, 0),
			}
		}

		existing, _ := applicationsMap[joinedStruct.ApplicationId]

		if joinedStruct.ApplicationVersionId != nil {
			version := types.ApplicationVersion{
				ID:              *joinedStruct.ApplicationVersionId,
				CreatedAt:       *joinedStruct.ApplicationVersionCreatedAt,
				Name:            *joinedStruct.ApplicationVersionName,
				ComposeFileData: joinedStruct.ApplicationVersionComposeFileData,
			}
			existing.Versions = append(existing.Versions, version)
		}
	}
	applications := make([]types.Application, 0, len(applicationsMap))
	for _, application := range applicationsMap {
		applications = append(applications, *application)
	}
	return applications
}

type applicationWithOptionalVersionRow struct {
	ApplicationId                     string
	ApplicationCreatedAt              time.Time
	ApplicationName                   string
	ApplicationType                   types.DeploymentType
	ApplicationVersionId              *string
	ApplicationVersionCreatedAt       *time.Time
	ApplicationVersionName            *string
	ApplicationVersionComposeFileData *[]byte
}

func CreateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDbOrPanic(ctx)
	row := db.QueryRow(ctx,
		"INSERT INTO ApplicationVersion (name, application_id) VALUES (@name, @applicationId) RETURNING id",
		pgx.NamedArgs{"name": applicationVersion.Name, "applicationId": applicationVersion.ApplicationId})
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
