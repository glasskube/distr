package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glasskube/cloud/api"

	"github.com/glasskube/cloud/internal/env"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentOutputExpr = `
		d.id, d.created_at, d.deployment_target_id, d.release_name
	`
)

func GetDeploymentsForDeploymentTarget(ctx context.Context, deploymentTargetId string) ([]types.Deployment, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT"+deploymentOutputExpr+
			"FROM Deployment d "+
			"WHERE d.deployment_target_id = @deploymentTargetId "+
			"ORDER BY d.created_at DESC",
		pgx.NamedArgs{"deploymentTargetId": deploymentTargetId})
	if err != nil {
		return nil, fmt.Errorf("failed to query Deployments: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Deployment]); err != nil {
		return nil, fmt.Errorf("failed to get Deployments: %w", err)
	} else {
		return result, nil
	}
}

func GetDeployment(ctx context.Context, id string) (*types.Deployment, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT"+deploymentOutputExpr+
			"FROM Deployment d "+
			"WHERE d.id = @id",
		pgx.NamedArgs{"id": id})
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

func GetLatestDeploymentForDeploymentTarget(ctx context.Context, deploymentTargetId string) (
	*types.DeploymentWithLatestRevision, error) {
	// TODO all these methods also need the orgId criteria
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`SELECT`+deploymentOutputExpr+`,
				dr.application_version_id as application_version_id, dr.values_yaml as values_yaml,
				dr.id as deployment_revision_id,
				a.id AS application_id, a.name AS application_name, av.name AS application_version_name
			FROM Deployment d
				JOIN DeploymentRevision dr ON d.id = dr.deployment_id
				JOIN ApplicationVersion av ON dr.application_version_id = av.id
				JOIN Application a ON av.application_id = a.id
			WHERE d.deployment_target_id = @deploymentTargetId
			ORDER BY d.created_at DESC, dr.created_at DESC LIMIT 1`,
		pgx.NamedArgs{"deploymentTargetId": deploymentTargetId})
	if err != nil {
		return nil, fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentWithLatestRevision])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get Deployment: %w", err)
	} else {
		return &result, nil
	}
}

func CreateDeployment(ctx context.Context, d *api.DeploymentRequest) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO Deployment AS d
			(deployment_target_id, release_name)
			VALUES (@deploymentTargetId, @releaseName)
			RETURNING`+deploymentOutputExpr,
		pgx.NamedArgs{
			"deploymentTargetId": d.DeploymentTargetId,
			"releaseName":        d.ReleaseName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Deployment])
	if err != nil {
		return fmt.Errorf("could not save Deployment: %w", err)
	} else {
		d.ID = result.ID
		return nil
	}
}

func CreateDeploymentRevision(ctx context.Context, d *api.DeploymentRequest) (*types.DeploymentRevision, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO DeploymentRevision AS d
			(deployment_id, application_version_id, values_yaml)
			VALUES (@deploymentId, @applicationVersionId, @valuesYaml)
			RETURNING d.id, d.created_at, d.deployment_id, d.application_version_id`,
		pgx.NamedArgs{
			"deploymentId":         d.ID,
			"applicationVersionId": d.ApplicationVersionId,
			"valuesYaml":           d.ValuesYaml,
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
	revisionID string,
	statusType types.DeploymentStatusType,
	message string,
) error {
	db := internalctx.GetDb(ctx)
	var id string
	rows := db.QueryRow(ctx, `
		INSERT INTO DeploymentRevisionStatus (deployment_revision_id, message, type)
		VALUES (@deploymentRevisionId, @message, @type)
		RETURNING id`,
		pgx.NamedArgs{"deploymentRevisionId": revisionID, "message": message, "type": statusType})
	if err := rows.Scan(&id); err != nil {
		return err
	} else {
		return nil
	}
}

func CreateDeploymentRevisionStatusWithCreatedAt(
	ctx context.Context,
	revisionID string,
	message string,
	createdAt time.Time,
) error {
	db := internalctx.GetDb(ctx)
	var id string
	rows := db.QueryRow(ctx,
		"INSERT INTO DeploymentRevisionStatus (deployment_revision_id, message, type, created_at) "+
			"VALUES (@deploymentRevisionId, @message, @type, @createdAt) RETURNING id",
		pgx.NamedArgs{
			"deploymentRevisionId": revisionID,
			"message":              message,
			"type":                 types.DeploymentStatusTypeOK,
			"createdAt":            createdAt,
		})
	if err := rows.Scan(&id); err != nil {
		return err
	} else {
		return nil
	}
}

func CleanupDeploymentRevisionStatus(ctx context.Context, revisionID string) (int64, error) {
	if env.StatusEntriesMaxAge() == nil {
		return 0, nil
	}
	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(ctx, `
		DELETE FROM DeploymentRevisionStatus
		       WHERE deployment_revision_id = @deploymentRevisionId AND
		             current_timestamp - created_at > @statusEntriesMaxAge`,
		pgx.NamedArgs{"deploymentRevisionId": revisionID, "statusEntriesMaxAge": env.StatusEntriesMaxAge()}); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}

func GetDeploymentStatus(
	ctx context.Context,
	deploymentId string,
	maxRows int,
) ([]types.DeploymentRevisionStatus, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT drs.id, drs.created_at, drs.deployment_revision_id, drs.type, drs.message
		FROM DeploymentRevisionStatus drs
			INNER JOIN DeploymentRevision dr ON dr.id = drs.deployment_revision_id
		WHERE dr.deployment_id = @deploymentId
		ORDER BY created_at DESC
		LIMIT @maxRows`,
		pgx.NamedArgs{"deploymentId": deploymentId, "maxRows": maxRows})
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentRevisionStatus: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentRevisionStatus]); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentRevisionStatus: %w", err)
	} else {
		return result, nil
	}
}
