package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	applicationOutputExpr             = `a.id, a.created_at, a.organization_id, a.name, a.type`
	applicationWithVersionsOutputExpr = applicationOutputExpr + `,
		coalesce((
			SELECT array_agg(row(av.id, av.created_at, av.archived_at, av.name, av.application_id,
				av.chart_type, av.chart_name, av.chart_url, av.chart_version) ORDER BY av.created_at ASC)
			FROM applicationversion av
			WHERE av.application_id = a.id
		), array[]::record[]) as versions `
)

func CreateApplication(ctx context.Context, application *types.Application, orgID uuid.UUID) error {
	application.OrganizationID = orgID
	db := internalctx.GetDb(ctx)
	row := db.QueryRow(ctx,
		"INSERT INTO Application (name, type, organization_id) VALUES (@name, @type, @orgId) RETURNING id, created_at",
		pgx.NamedArgs{"name": application.Name, "type": application.Type, "orgId": application.OrganizationID})
	if err := row.Scan(&application.ID, &application.CreatedAt); err != nil {
		return fmt.Errorf("could not save application: %w", err)
	}
	return nil
}

func UpdateApplication(ctx context.Context, application *types.Application, orgID uuid.UUID) error {
	application.OrganizationID = orgID
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE Application SET name = @name WHERE id = @id AND organization_id = @orgId RETURNING *",
		pgx.NamedArgs{"id": application.ID, "name": application.Name, "orgId": application.OrganizationID})
	if err != nil {
		return fmt.Errorf("could not update application: %w", err)
	} else if updated, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.Application]); err != nil {
		return fmt.Errorf("could not get updated application: %w", err)
	} else {
		*application = updated
		return nil
	}
}

func DeleteApplicationWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, `DELETE FROM Application WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
	} else if cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("could not delete Application: %w", err)
	}

	return nil
}

func GetApplicationsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.Application, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT `+applicationWithVersionsOutputExpr+`
			FROM Application a
			WHERE a.organization_id = @orgId
			ORDER BY a.name
			`, pgx.NamedArgs{"orgId": orgID}); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err :=
		pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func GetApplicationsWithLicenseOwnerID(ctx context.Context, id uuid.UUID) ([]types.Application, error) {
	db := internalctx.GetDb(ctx)
	// TODO: Only include versions from at least one license
	if rows, err := db.Query(ctx, `
			SELECT DISTINCT `+applicationWithVersionsOutputExpr+`
			FROM ApplicationLicense al
				LEFT JOIN Application a ON al.application_id = a.id
			WHERE al.owner_useraccount_id = @id
			ORDER BY a.name
			`, pgx.NamedArgs{"id": id}); err != nil {
		return nil, fmt.Errorf("failed to query applications: %w", err)
	} else if applications, err :=
		pgx.CollectRows(rows, pgx.RowToStructByName[types.Application]); err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	} else {
		return applications, nil
	}
}

func GetApplication(ctx context.Context, id, orgID uuid.UUID) (*types.Application, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT `+applicationWithVersionsOutputExpr+`
			FROM Application a
			WHERE a.id = @id AND a.organization_id = @orgId
		`, pgx.NamedArgs{"id": id, "orgId": orgID}); err != nil {
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

func GetApplicationForApplicationVersionID(ctx context.Context, id, orgID uuid.UUID) (*types.Application, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT `+applicationWithVersionsOutputExpr+`
			FROM ApplicationVersion v
				LEFT JOIN Application a ON a.id = v.application_id
			WHERE v.id = @id AND a.organization_id = @orgId
		`, pgx.NamedArgs{"id": id, "orgId": orgID}); err != nil {
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
		"applicationId": applicationVersion.ApplicationID,
		"chartType":     applicationVersion.ChartType,
		"chartName":     applicationVersion.ChartName,
		"chartUrl":      applicationVersion.ChartUrl,
		"chartVersion":  applicationVersion.ChartVersion,
	}
	if applicationVersion.ComposeFileData != nil {
		args["composeFileData"] = applicationVersion.ComposeFileData
	}
	if applicationVersion.ValuesFileData != nil {
		args["valuesFileData"] = applicationVersion.ValuesFileData
	}
	if applicationVersion.TemplateFileData != nil {
		args["templateFileData"] = applicationVersion.TemplateFileData
	}

	row, err := db.Query(ctx,
		`INSERT INTO ApplicationVersion AS av (name, application_id, chart_type, chart_name, chart_url, chart_version,
				compose_file_data, values_file_data, template_file_data)
			VALUES (@name, @applicationId, @chartType, @chartName, @chartUrl, @chartVersion, @composeFileData::bytea,
				@valuesFileData::bytea, @templateFileData::bytea)
			RETURNING av.id, av.created_at, av.name, av.chart_type, av.chart_name, av.chart_url, av.chart_version,
				av.values_file_data, av.template_file_data, av.compose_file_data, av.application_id`, args)
	if err != nil {
		return fmt.Errorf("can not create ApplicationVersion: %w", err)
	} else if result, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[types.ApplicationVersion]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		} else if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.UniqueViolation {
			err = apierrors.ErrAlreadyExists
		}
		return fmt.Errorf("could not scan ApplicationVersion: %w", err)
	} else {
		*applicationVersion = result
		return nil
	}
}

func UpdateApplicationVersion(ctx context.Context, applicationVersion *types.ApplicationVersion) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`UPDATE ApplicationVersion AS av SET name = @name, archived_at = @archivedAt WHERE id = @id
		RETURNING av.id, av.created_at, av.name, av.chart_type, av.chart_name, av.chart_url, av.chart_version,
				av.values_file_data, av.template_file_data, av.compose_file_data, av.application_id`,
		pgx.NamedArgs{
			"id":         applicationVersion.ID,
			"name":       applicationVersion.Name,
			"archivedAt": applicationVersion.ArchivedAt,
		})
	if err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.UniqueViolation {
			err = apierrors.ErrAlreadyExists
		}
		return fmt.Errorf("can not update ApplicationVersion: %w", err)
	} else if updated, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ApplicationVersion]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return fmt.Errorf("could not scan ApplicationVersion: %w", err)
	} else {
		*applicationVersion = updated
		return nil
	}
}

func GetApplicationVersion(ctx context.Context, applicationVersionID uuid.UUID) (*types.ApplicationVersion, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT av.id, av.created_at, av.archived_at, av.name, av.chart_type, av.chart_name, av.chart_url, av.chart_version,
			av.values_file_data, av.template_file_data, av.compose_file_data, av.application_id
		FROM ApplicationVersion av
		WHERE id = @id`,
		pgx.NamedArgs{"id": applicationVersionID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get ApplicationVersion: %w", err)
	} else if data, err := pgx.CollectExactlyOneRow(rows,
		pgx.RowToStructByName[types.ApplicationVersion]); err != nil {
		if err == pgx.ErrNoRows {
			return nil, apierrors.ErrNotFound
		}
		return nil, err
	} else {
		return &data, nil
	}
}
