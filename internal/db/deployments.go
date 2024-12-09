package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentOutputExpr = `
		id, created_at, deployment_target_id, application_version_id
	`
)

func GetDeploymentsForDeploymentTarget(ctx context.Context, deploymentTargetId string) ([]types.Deployment, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+deploymentOutputExpr+" FROM Deployment WHERE deployment_target_id = @deploymentTargetId"+
			" ORDER BY created_at DESC",
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
		"SELECT "+deploymentOutputExpr+" FROM Deployment WHERE id = @id",
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
		`select d.id, d.created_at, d.deployment_target_id, d.application_version_id,
       			a.id as application_id, a.name as application_name, av.name as application_version_name
			from deployment d join applicationversion av on(d.application_version_id=av.id)
				join application a on (av.application_id=a.id)
			where deployment_target_id = @deploymentTargetId
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

func GetLatestDeploymentComposeFile(ctx context.Context, deploymentTargetId string, orgId string) ([]byte, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
		SELECT av.compose_file_data
		FROM Deployment d
		INNER JOIN ApplicationVersion av ON d.application_version_id = av.id
		INNER JOIN DeploymentTarget dt ON d.deployment_target_id = dt.id
		WHERE d.deployment_target_id = @deploymentTargetId AND dt.organization_id = @orgId
		ORDER BY d.created_at DESC LIMIT 1`, pgx.NamedArgs{
		"deploymentTargetId": deploymentTargetId,
		"orgId":              orgId,
	}); err != nil {
		return nil, fmt.Errorf("could not get latest deployment: %w", err)
	} else if data, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[[]byte]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get latest deployment: %w", err)
	} else {
		return data, nil
	}
}

func CreateDeployment(ctx context.Context, d *types.Deployment) error {
	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{"deployment_target_id": d.DeploymentTargetId, "application_version_id": d.ApplicationVersionId}
	rows, err := db.Query(ctx,
		`INSERT INTO Deployment (deployment_target_id, application_version_id)
			VALUES (@deployment_target_id, @application_version_id) RETURNING `+deploymentOutputExpr,
		args)
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

func CreateDeploymentStatus(ctx context.Context, deployment *types.Deployment, message string) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO DeploymentStatus (deployment_id, message) VALUES (@deploymentId, @message)",
		pgx.NamedArgs{"deploymentId": deployment.ID, "message": message})
	if err != nil {
		return err
	} else {
		rows.Close()
		return nil
	}
}
