package db

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/glasskube/cloud/internal/apierrors"
	"github.com/glasskube/cloud/internal/auth"
	internalctx "github.com/glasskube/cloud/internal/context"
	"github.com/glasskube/cloud/internal/types"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentTargetOutputExprBase = `
		dt.id, dt.created_at, dt.name, dt.type, dt.access_key_salt, dt.access_key_hash, dt.organization_id,
		CASE WHEN dt.geolocation_lat IS NOT NULL AND dt.geolocation_lon IS NOT NULL
		  	THEN (dt.geolocation_lat, dt.geolocation_lon) END AS geolocation
	`
	deploymentTargetOutputExpr = deploymentTargetOutputExprBase +
		", (" + userAccountOutputExpr + ") as created_by"
	deploymentTargetWithStatusOutputExpr = deploymentTargetOutputExpr + `,
		CASE WHEN status.id IS NOT NULL
			THEN (status.id, status.created_at, status.message) END AS current_status
	`
	deploymentTargetJoinExpr = `
		LEFT JOIN DeploymentTargetStatus status ON dt.id = status.deployment_target_id
		LEFT JOIN UserAccount u ON dt.created_by_user_account_id = u.id
	`
	deploymentTargetWhereExprBase = `
		WHERE (
			status.id IS NULL OR status.created_at = (
				SELECT max(s.created_at) FROM DeploymentTargetStatus s WHERE s.deployment_target_id = status.deployment_target_id
			)
		)
	`
	deploymentTargetFromExpr = `
		FROM DeploymentTarget dt
	` + deploymentTargetJoinExpr + deploymentTargetWhereExprBase
)

func GetDeploymentTargets(ctx context.Context) ([]types.DeploymentTargetWithCreatedBy, error) {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return nil, err
	}
	userId, err := auth.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	userRole, err := auth.CurrentUserRole(ctx)
	if err != nil {
		return nil, err
	}

	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx,
		"SELECT "+deploymentTargetWithStatusOutputExpr+" "+deploymentTargetFromExpr+" "+
			"AND dt.organization_id = @orgId "+
			"AND (dt.created_by_user_account_id = @userId OR @userRole = 'vendor')",
		pgx.NamedArgs{"orgId": orgId, "userId": userId, "userRole": userRole},
	); err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	} else if result, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[types.DeploymentTargetWithCreatedBy],
	); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTargets: %w", err)
	} else {
		return result, nil
	}
}

func GetDeploymentTarget(ctx context.Context, id string, orgId *string) (*types.DeploymentTargetWithCreatedBy, error) {
	db := internalctx.GetDb(ctx)
	var args pgx.NamedArgs
	query := "SELECT " + deploymentTargetWithStatusOutputExpr + " " + deploymentTargetFromExpr + " AND dt.id = @id"
	if orgId != nil {
		args = pgx.NamedArgs{"id": id, "orgId": *orgId}
		query = query + " AND dt.organization_id = @orgId"
	} else {
		args = pgx.NamedArgs{"id": id}
	}
	rows, err := db.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentTargetWithCreatedBy])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTarget: %w", err)
	} else {
		return &result, nil
	}
}

func CreateDeploymentTarget(ctx context.Context, dt *types.DeploymentTargetWithCreatedBy) error {
	if dt.OrganizationID == "" {
		if orgId, err := auth.CurrentOrgId(ctx); err != nil {
			return err
		} else {
			dt.OrganizationID = orgId
		}
	}
	if dt.CreatedBy == nil {
		if userId, err := auth.CurrentUserId(ctx); err != nil {
			return err
		} else {
			dt.CreatedBy = &types.UserAccount{ID: userId}
		}
	}

	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{
		"name":   dt.Name,
		"type":   dt.Type,
		"orgId":  dt.OrganizationID,
		"userId": dt.CreatedBy.ID,
		"lat":    nil,
		"lon":    nil,
	}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(
		ctx,
		`WITH inserted AS (
			INSERT INTO DeploymentTarget
			(name, type, organization_id, created_by_user_account_id, geolocation_lat, geolocation_lon)
			VALUES (@name, @type, @orgId, @userId, @lat, @lon) RETURNING *
		)
		SELECT `+deploymentTargetOutputExpr+` FROM inserted dt`+deploymentTargetJoinExpr+deploymentTargetWhereExprBase,
		args,
	)
	if err != nil {
		return fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.DeploymentTargetWithCreatedBy])
	if err != nil {
		return fmt.Errorf("could not save DeploymentTarget: %w", err)
	} else {
		*dt = result
		return nil
	}
}

func UpdateDeploymentTarget(ctx context.Context, dt *types.DeploymentTargetWithCreatedBy) error {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return err
	}

	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{"id": dt.ID, "name": dt.Name, "orgId": orgId, "lat": nil, "lon": nil}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(ctx,
		`WITH updated AS (
			UPDATE DeploymentTarget AS dt SET name = @name, geolocation_lat = @lat, geolocation_lon = @lon
			WHERE id = @id AND organization_id = @orgId RETURNING *
		)
		SELECT `+deploymentTargetOutputExpr+` FROM updated dt`+deploymentTargetJoinExpr+deploymentTargetWhereExprBase,
		args)
	if err != nil {
		return fmt.Errorf("could not update DeploymentTarget: %w", err)
	} else if updated, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.DeploymentTargetWithCreatedBy]); err != nil {
		return fmt.Errorf("could not get updated DeploymentTarget: %w", err)
	} else {
		*dt = updated
		return nil
	}
}

func UpdateDeploymentTargetAccess(ctx context.Context, dt *types.DeploymentTarget) error {
	orgId, err := auth.CurrentOrgId(ctx)
	if err != nil {
		return err
	}
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE DeploymentTarget AS dt SET access_key_salt = @accessKeySalt, access_key_hash = @accessKeyHash "+
			"WHERE id = @id AND organization_id = @orgId RETURNING "+
			deploymentTargetOutputExprBase,
		pgx.NamedArgs{"accessKeySalt": dt.AccessKeySalt, "accessKeyHash": dt.AccessKeyHash, "id": dt.ID, "orgId": orgId})
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

func CreateDeploymentTargetStatus(ctx context.Context, dt *types.DeploymentTarget, message string) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO DeploymentTargetStatus (deployment_target_id, message) VALUES (@deploymentTargetId, @message)",
		pgx.NamedArgs{"deploymentTargetId": dt.ID, "message": message})
	if err != nil {
		return err
	} else {
		rows.Close()
		return nil
	}
}
