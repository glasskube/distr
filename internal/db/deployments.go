package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glasskube/cloud/internal/env"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentOutputExpr = `
		d.id, d.created_at, d.deployment_target_id, d.application_version_id, d.release_name, d.values_yaml
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
	*types.DeploymentWithData, error) {
	// TODO all these methods also need the orgId criteria
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		`SELECT`+deploymentOutputExpr+`,
				a.id AS application_id, a.name AS application_name, av.name AS application_version_name
			FROM Deployment d
				JOIN ApplicationVersion av ON d.application_version_id = av.id
				JOIN Application a ON av.application_id = a.id
			WHERE d.deployment_target_id = @deploymentTargetId
			ORDER BY d.created_at DESC LIMIT 1`,
		pgx.NamedArgs{"deploymentTargetId": deploymentTargetId})
	if err != nil {
		return nil, fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentWithData])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get Deployment: %w", err)
	} else {
		return &result, nil
	}
}

func GetLatestDeploymentComposeFile(
	ctx context.Context,
	deploymentTargetId string,
	orgId string,
) (string, []byte, error) {
	db := internalctx.GetDb(ctx)
	var deploymentId string
	var file []byte
	rows := db.QueryRow(ctx, `
		SELECT d.id, av.compose_file_data
		FROM Deployment d
		INNER JOIN ApplicationVersion av ON d.application_version_id = av.id
		INNER JOIN DeploymentTarget dt ON d.deployment_target_id = dt.id
		WHERE d.deployment_target_id = @deploymentTargetId AND dt.organization_id = @orgId
		ORDER BY d.created_at DESC LIMIT 1`, pgx.NamedArgs{
		"deploymentTargetId": deploymentTargetId,
		"orgId":              orgId,
	})
	if err := rows.Scan(&deploymentId, &file); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, apierrors.ErrNotFound
		}
		return "", nil, fmt.Errorf("failed to get latest deployment: %w", err)
	} else {
		return deploymentId, file, nil
	}
}

func CreateDeployment(ctx context.Context, d *types.Deployment) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO Deployment AS d
			(deployment_target_id, application_version_id, release_name, values_yaml)
			VALUES (@deploymentTargetId, @applicationVersionId, @releaseName, @valuesYaml)
			RETURNING`+deploymentOutputExpr,
		pgx.NamedArgs{
			"deploymentTargetId":   d.DeploymentTargetId,
			"applicationVersionId": d.ApplicationVersionId,
			"releaseName":          d.ReleaseName,
			"valuesYaml":           d.ValuesYaml,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to query Deployments: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Deployment])
	if err != nil {
		return fmt.Errorf("could not save Deployment: %w", err)
	} else {
		*d = result
		return nil
	}
}

func CreateDeploymentStatus(ctx context.Context, deploymentID string, message string) error {
	statusType := types.DeploymentStatusTypeOK // TODO get type
	db := internalctx.GetDb(ctx)
	var id string
	rows := db.QueryRow(ctx,
		"INSERT INTO DeploymentStatus (deployment_id, message, type) VALUES (@deploymentId, @message, @type) RETURNING id",
		pgx.NamedArgs{"deploymentId": deploymentID, "message": message, "type": statusType})
	if err := rows.Scan(&id); err != nil {
		return err
	} else {
		return nil
	}
}

func CreateDeploymentStatusWithCreatedAt(
	ctx context.Context,
	deploymentID string,
	message string,
	createdAt time.Time,
) error {
	db := internalctx.GetDb(ctx)
	var id string
	rows := db.QueryRow(ctx,
		"INSERT INTO DeploymentStatus (deployment_id, message, created_at) "+
			"VALUES (@deploymentId, @message, @createdAt) RETURNING id",
		pgx.NamedArgs{"deploymentId": deploymentID, "message": message, "createdAt": createdAt})
	if err := rows.Scan(&id); err != nil {
		return err
	} else {
		return nil
	}
}

func CleanupDeploymentStatus(ctx context.Context, deploymentId string) (int64, error) {
	if env.StatusEntriesMaxAge() == nil {
		return 0, nil
	}
	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(ctx, `
		DELETE FROM DeploymentStatus
		       WHERE deployment_id = @deploymentId AND
		             current_timestamp - created_at > @statusEntriesMaxAge`,
		pgx.NamedArgs{"deploymentId": deploymentId, "statusEntriesMaxAge": env.StatusEntriesMaxAge()}); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}

func GetDeploymentStatus(ctx context.Context, deploymentId string, maxRows int) ([]types.DeploymentStatus, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT id, created_at, deployment_id, type, message
		FROM DeploymentStatus
		WHERE deployment_id = @deploymentId
		ORDER BY created_at DESC
		LIMIT @maxRows`,
		pgx.NamedArgs{"deploymentId": deploymentId, "maxRows": maxRows})
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentStatus: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentStatus]); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentStatus: %w", err)
	} else {
		return result, nil
	}
}
