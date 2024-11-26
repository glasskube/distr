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
	deploymentTargetOutputExpr = `
		id, created_at, name, type,
		CASE WHEN geolocation_lat IS NOT NULL AND geolocation_lon IS NOT NULL
		  THEN (geolocation_lat, geolocation_lon) END AS geolocation
	`
)

func GetDeploymentTargets(ctx context.Context) ([]types.DeploymentTarget, error) {
	db := internalctx.GetDbOrPanic(ctx)
	if rows, err := db.Query(ctx, "SELECT "+deploymentTargetOutputExpr+" FROM DeploymentTarget"); err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	} else if result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentTarget]); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTargets: %w", err)
	} else {
		return result, nil
	}
}

func GetDeploymentTarget(ctx context.Context, id string) (*types.DeploymentTarget, error) {
	db := internalctx.GetDbOrPanic(ctx)
	rows, err := db.Query(ctx,
		"SELECT "+deploymentTargetOutputExpr+" FROM DeploymentTarget WHERE id = @id",
		pgx.NamedArgs{"id": id})
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentTarget])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.NotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTarget: %w", err)
	} else {
		return &result, nil
	}
}
