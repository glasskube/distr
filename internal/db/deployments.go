package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	deploymentOutputExpr = `
		d.id, d.created_at, d.deployment_target_id, d.release_name, d.application_license_id, d.docker_type,
		d.logs_enabled
	`
)

func GetDeployment(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
	orgID uuid.UUID,
	userRole types.UserRole,
) (*types.Deployment, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT"+deploymentOutputExpr+
			"FROM Deployment d "+
			"INNER JOIN DeploymentTarget dt ON d.deployment_target_id = dt.id "+
			"WHERE d.id = @id AND dt.organization_id = @orgId "+
			"AND (@userRole = 'vendor' OR dt.created_by_user_account_id = @userId)",
		pgx.NamedArgs{
			"id":       id,
			"userId":   userID,
			"orgId":    orgID,
			"userRole": userRole,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Deployment])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get Deployment: %w", err)
	} else {
		return &result, nil
	}
}

func GetDeploymentsForDeploymentTarget(
	ctx context.Context,
	deploymentTargetID uuid.UUID,
) ([]types.DeploymentWithLatestRevision, error) {
	// TODO all these methods also need the orgId criteria
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT`+deploymentOutputExpr+`,
				dr.application_version_id as application_version_id,
				dr.values_yaml as values_yaml,
				dr.env_file_data as env_file_data,
				dr.id as deployment_revision_id,
				dr.created_at AS deployment_revision_created_at,
				a.id AS application_id,
				a.name AS application_name,
				av.name AS application_version_name,
				CASE WHEN drs.id IS NOT NULL THEN (
					drs.id,
					drs.created_at,
					drs.deployment_revision_id,
					drs.type, drs.message
				) END AS latest_status
			FROM Deployment d
				LEFT JOIN (
					SELECT deployment_id, max(created_at) AS max_created_at
					FROM DeploymentRevision
					GROUP BY deployment_id
				) dr_max ON d.id = dr_max.deployment_id
				JOIN DeploymentRevision dr
					ON d.id = dr.deployment_id
					AND dr.created_at = dr_max.max_created_at
				JOIN ApplicationVersion av ON dr.application_version_id = av.id
				JOIN Application a ON av.application_id = a.id
				-- Join the DeploymentRevision table again because we ALSO need the latest deployment revision for
				-- which exists a status. Otherwise, the deployment is shown as "no status" after an update
				LEFT JOIN (
					SELECT deployment_id, max(created_at) AS max_created_at
					FROM DeploymentRevision dr1
					WHERE exists(SELECT id FROM DeploymentRevisionStatus WHERE deployment_revision_id = dr1.id)
					GROUP BY deployment_id
				) dr_max_status ON d.id = dr_max_status.deployment_id
				LEFT JOIN DeploymentRevision dr_status
					ON d.id = dr_status.deployment_id
					AND dr_status.created_at = dr_max_status.max_created_at
				LEFT JOIN (
					SELECT
						dr1.id AS deployment_revision_id,
						(SELECT max(created_at) FROM DeploymentRevisionStatus WHERE deployment_revision_id = dr1.id) AS max_created_at
					FROM DeploymentRevision dr1
				) status_max ON dr_status.id = status_max.deployment_revision_id
				LEFT JOIN DeploymentRevisionStatus drs
					ON dr_status.id = drs.deployment_revision_id
					AND drs.created_at = status_max.max_created_at
			WHERE d.deployment_target_id = @deploymentTargetId
			ORDER BY d.created_at`,
		pgx.NamedArgs{"deploymentTargetId": deploymentTargetID})
	if err != nil {
		return nil, fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentWithLatestRevision])
	if err != nil {
		return nil, fmt.Errorf("failed to scan Deployments: %w", err)
	} else {
		return result, nil
	}
}

func CreateDeployment(ctx context.Context, request *api.DeploymentRequest) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO Deployment AS d
			(deployment_target_id, release_name, application_license_id, docker_type)
			VALUES (@deploymentTargetId, @releaseName, @applicationLicenseId, @dockerType)
			RETURNING`+deploymentOutputExpr,
		pgx.NamedArgs{
			"deploymentTargetId":   request.DeploymentTargetID,
			"releaseName":          request.ReleaseName,
			"applicationLicenseId": request.ApplicationLicenseID,
			"dockerType":           request.DockerType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Deployment])
	if err != nil {
		return fmt.Errorf("could not save Deployment: %w", err)
	} else {
		request.DeploymentID = &result.ID
		return nil
	}
}

func UpdateDeployment(ctx context.Context, deployment *types.Deployment) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`UPDATE Deployment AS d
		SET logs_enabled = @logsEnabled
		WHERE id = @id
		RETURNING`+deploymentOutputExpr,
		pgx.NamedArgs{
			"id":          deployment.ID,
			"logsEnabled": deployment.LogsEnabled,
		},
	)
	if err != nil {
		return fmt.Errorf("could not update Deployment: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Deployment]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return fmt.Errorf("could not update Deployment: %w", err)
	} else {
		*deployment = result
		return nil
	}
}

func DeleteDeploymentWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	res, err := db.Exec(ctx, "DELETE FROM Deployment WHERE id = @id", pgx.NamedArgs{"id": id})
	if err == nil && res.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("could not delete Deployment: %w", err)
	}
	return nil
}

func CreateDeploymentRevision(ctx context.Context, request *api.DeploymentRequest) (*types.DeploymentRevision, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO DeploymentRevision AS d
			(deployment_id, application_version_id, values_yaml, env_file_data)
			VALUES (@deploymentId, @applicationVersionId, @valuesYaml, @envFileData)
			RETURNING d.id, d.created_at, d.deployment_id, d.application_version_id`,
		pgx.NamedArgs{
			"deploymentId":         request.DeploymentID,
			"applicationVersionId": request.ApplicationVersionID,
			"valuesYaml":           request.ValuesYaml,
			"envFileData":          request.EnvFileData,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentRevision: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentRevision])
	if err != nil {
		return nil, fmt.Errorf("could not save DeploymentRevision: %w", err)
	} else {
		return &result, nil
	}
}

func CreateDeploymentRevisionStatus(
	ctx context.Context,
	revisionID uuid.UUID,
	statusType types.DeploymentStatusType,
	message string,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(ctx, `
		INSERT INTO DeploymentRevisionStatus (deployment_revision_id, message, type)
		VALUES (@deploymentRevisionId, @message, @type)`,
		pgx.NamedArgs{"deploymentRevisionId": revisionID, "message": message, "type": statusType})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else {
		return nil
	}
}

func BulkCreateDeploymentRevisionStatusWithCreatedAt(
	ctx context.Context,
	deploymentRevisionID uuid.UUID,
	statuses []types.DeploymentRevisionStatus,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"deploymentrevisionstatus"},
		[]string{"deployment_revision_id", "type", "message", "created_at"},
		pgx.CopyFromSlice(len(statuses), func(i int) ([]any, error) {
			return []any{
				deploymentRevisionID,
				types.DeploymentStatusTypeOK,
				statuses[i].Message,
				statuses[i].CreatedAt,
			}, nil
		}),
	)
	return err
}

// CleanupDeploymentRevisionStatus deletes all DeploymentRevisionStatus entries older than [env.StatusEntriesMaxAge()],
// always keeping the latest entry across all DeploymentRevisions of every Deployment
func CleanupDeploymentRevisionStatus(ctx context.Context) (int64, error) {
	if env.StatusEntriesMaxAge() == nil {
		return 0, nil
	}

	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(
		ctx,
		`DELETE FROM DeploymentRevisionStatus drs
		USING (
			SELECT
				dr1.id AS deployment_revision_id,
				max(dr2.max_created_at) AS max_created_at
			FROM DeploymentRevision dr1
			JOIN (
				SELECT dr.id, dr.deployment_id, (
					SELECT max(drs.created_at)
					FROM DeploymentRevisionStatus drs
					WHERE drs.deployment_revision_id = dr.id
				) AS max_created_at
				FROM DeploymentRevision dr
			) dr2 ON dr1.deployment_id = dr2.deployment_id
			GROUP BY dr1.id
		) max_created_at
		WHERE drs.deployment_revision_id = max_created_at.deployment_revision_id
			AND drs.created_at < max_created_at.max_created_at
			AND current_timestamp - drs.created_at > @statusEntriesMaxAge`,
		pgx.NamedArgs{"statusEntriesMaxAge": env.StatusEntriesMaxAge()},
	); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}

func GetDeploymentStatus(
	ctx context.Context,
	deploymentID uuid.UUID,
	maxRows int,
	before time.Time,
	after time.Time,
) ([]types.DeploymentRevisionStatus, error) {
	if before.IsZero() {
		before = time.Now()
	}

	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		"SELECT id from DeploymentRevision WHERE deployment_id = @deploymentId",
		pgx.NamedArgs{"deploymentId": deploymentID},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentRevision for status: %w", err)
	}
	deploymentRevisionIDs, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, fmt.Errorf("failed to scan DeploymentRevision for status: %w", err)
	}

	rows, err = db.Query(
		ctx,
		`SELECT id, created_at, deployment_revision_id, type, message
		FROM DeploymentRevisionStatus
		WHERE deployment_revision_id = ANY (@deploymentRevisionIds)
			AND created_at BETWEEN @after AND @before
		ORDER BY created_at DESC
		LIMIT @maxRows`,
		pgx.NamedArgs{
			"deploymentRevisionIds": deploymentRevisionIDs,
			"maxRows":               maxRows,
			"before":                before,
			"after":                 after,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentRevisionStatus: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentRevisionStatus]); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentRevisionStatus: %w", err)
	} else {
		return result, nil
	}
}
