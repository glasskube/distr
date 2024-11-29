package db

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/glasskube/cloud/internal/apierrors"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentTargetOutputExpr = `
		dt.id, dt.created_at, dt.name, dt.type,
		CASE WHEN dt.geolocation_lat IS NOT NULL AND dt.geolocation_lon IS NOT NULL
		  	THEN (dt.geolocation_lat, dt.geolocation_lon) END AS geolocation
	`
	deploymentTargetWithStatusOutputExpr = deploymentTargetOutputExpr + `,
		CASE WHEN status.id IS NOT NULL
			THEN (status.id, status.created_at, status.message) END AS current_status
	`
	deploymentTargetFromExpr = `
		FROM DeploymentTarget dt LEFT JOIN DeploymentTargetStatus status ON dt.id = status.deployment_target_id
		WHERE (
			status.id IS NULL OR status.created_at = (
				SELECT max(s.created_at) FROM DeploymentTargetStatus s WHERE s.deployment_target_id = status.deployment_target_id
			)
		)
`
)

func GetDeploymentTargets(ctx context.Context) ([]types.DeploymentTarget, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx,
		"SELECT "+deploymentTargetWithStatusOutputExpr+" "+deploymentTargetFromExpr); err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentTarget]); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTargets: %w", err)
	} else {
		return result, nil
	}
}

func GetDeploymentTarget(ctx context.Context, id string) (*types.DeploymentTarget, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+deploymentTargetWithStatusOutputExpr+" "+deploymentTargetFromExpr+" AND dt.id = @id",
		pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentTarget])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTarget: %w", err)
	} else {
		return &result, nil
	}
}

func CreateDeploymentTarget(ctx context.Context, dt *types.DeploymentTarget) error {
	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{"name": dt.Name, "type": dt.Type, "lat": nil, "lon": nil}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(ctx,
		"INSERT INTO DeploymentTarget AS dt (name, type, geolocation_lat, geolocation_lon) "+
			"VALUES (@name, @type, @lat, @lon) RETURNING "+
			deploymentTargetOutputExpr,
		args)
	if err != nil {
		return fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.DeploymentTarget])
	if err != nil {
		return fmt.Errorf("could not save DeploymentTarget: %w", err)
	} else {
		*dt = result
		return nil
	}
}

func UpdateDeploymentTarget(ctx context.Context, dt *types.DeploymentTarget) error {
	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{"id": dt.ID, "name": dt.Name, "lat": nil, "lon": nil}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(ctx,
		"UPDATE DeploymentTarget AS dt SET name = @name, geolocation_lat = @lat, geolocation_lon = @lon "+
			" WHERE id = @id RETURNING "+
			deploymentTargetOutputExpr,
		args)
	if err != nil {
		return fmt.Errorf("could not update DeploymentTarget: %w", err)
	} else if updated, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.DeploymentTarget]); err != nil {
		return fmt.Errorf("could not get updated DeploymentTarget: %w", err)
	} else {
		*dt = updated
		return nil
	}
}
