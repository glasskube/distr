package db

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentTargetOutputExprBase = `
		dt.id,
		dt.created_at,
		dt.name,
		dt.type,
		dt.access_key_salt,
		dt.access_key_hash,
		dt.namespace,
		dt.scope,
		dt.organization_id,
		dt.created_by_user_account_id,
		dt.agent_version_id,
		dt.reported_agent_version_id,
		CASE WHEN dt.geolocation_lat IS NOT NULL AND dt.geolocation_lon IS NOT NULL
		  	THEN (dt.geolocation_lat, dt.geolocation_lon) END
			AS geolocation
	`
	deploymentTargetOutputExpr = deploymentTargetOutputExprBase +
		", (" + userAccountWithRoleOutputExpr + ") as created_by"
	deploymentTargetWithStatusOutputExpr = deploymentTargetOutputExpr + `,
		CASE WHEN status.id IS NOT NULL
			THEN (status.id, status.created_at, status.message) END
			AS current_status,
		CASE WHEN agv.id IS NOT NULL
			THEN (agv.id, agv.created_at, agv.name, agv.manifest_file_revision, agv.compose_file_revision) END
			AS agent_version
	`
	deploymentTargetJoinExpr = `
		LEFT JOIN (
			-- find the creation date of the latest status entry for each deployment target
			-- IMPORTANT: The sub-query here might seem inefficient but it is MUCH FASTER than using a GROUP BY clause
			-- because it can utilize an index!!
			SELECT
				dt1.id AS deployment_target_id,
				(SELECT max(created_at) FROM DeploymentTargetStatus WHERE deployment_target_id = dt1.id) AS max_created_at
			FROM DeploymentTarget dt1
		) status_max
		 	ON dt.id = status_max.deployment_target_id
		LEFT JOIN DeploymentTargetStatus status
			ON dt.id = status.deployment_target_id
			AND status.created_at = status_max.max_created_at
		LEFT JOIN AgentVersion agv
			ON dt.agent_version_id = agv.id
		LEFT JOIN UserAccount u
			ON dt.created_by_user_account_id = u.id
		LEFT JOIN Organization_UserAccount j
			ON u.id = j.user_account_id
	`
	deploymentTargetFromExpr = `
		DeploymentTarget dt
	` + deploymentTargetJoinExpr
)

func GetDeploymentTargets(
	ctx context.Context,
	orgID, userID uuid.UUID,
	userRole types.UserRole,
) ([]types.DeploymentTargetWithCreatedBy, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx,
		"SELECT"+deploymentTargetWithStatusOutputExpr+"FROM"+deploymentTargetFromExpr+
			"WHERE dt.organization_id = @orgId "+
			"AND (dt.created_by_user_account_id = @userId OR @userRole = 'vendor') "+
			"ORDER BY u.name, u.email, dt.name",
		pgx.NamedArgs{"orgId": orgID, "userId": userID, "userRole": userRole},
	); err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	} else if result, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[types.DeploymentTargetWithCreatedBy],
	); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTargets: %w", err)
	} else {
		for i := range result {
			if err := addDeploymentToTarget(ctx, &result[i]); err != nil {
				return nil, err
			}
		}
		return result, nil
	}
}

func GetDeploymentTarget(
	ctx context.Context,
	id uuid.UUID,
	orgID *uuid.UUID,
) (*types.DeploymentTargetWithCreatedBy, error) {
	db := internalctx.GetDb(ctx)
	var args pgx.NamedArgs
	query := "SELECT" + deploymentTargetWithStatusOutputExpr + "FROM" + deploymentTargetFromExpr + "WHERE dt.id = @id"
	if orgID != nil {
		args = pgx.NamedArgs{"id": id, "orgId": *orgID}
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
		return &result, addDeploymentToTarget(ctx, &result)
	}
}

func GetDeploymentTargetForDeploymentID(
	ctx context.Context,
	id uuid.UUID,
) (*types.DeploymentTargetWithCreatedBy, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		fmt.Sprintf("SELECT %v FROM %v JOIN Deployment d ON dt.id = d.deployment_target_id WHERE d.id = @id",
			deploymentTargetWithStatusOutputExpr, deploymentTargetFromExpr),
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.DeploymentTargetWithCreatedBy])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apierrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTarget: %w", err)
	} else {
		return &result, addDeploymentToTarget(ctx, &result)
	}
}

func CreateDeploymentTarget(
	ctx context.Context,
	dt *types.DeploymentTargetWithCreatedBy,
	orgID, createdByID uuid.UUID,
) error {
	dt.OrganizationID = orgID
	if dt.CreatedBy == nil {
		dt.CreatedBy = &types.UserAccountWithUserRole{ID: createdByID}
	}

	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{
		"name":           dt.Name,
		"type":           dt.Type,
		"orgId":          dt.OrganizationID,
		"userId":         dt.CreatedBy.ID,
		"namespace":      dt.Namespace,
		"scope":          dt.Scope,
		"lat":            nil,
		"lon":            nil,
		"agentVersionId": dt.AgentVersionID,
	}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(
		ctx,
		`WITH inserted AS (
			INSERT INTO DeploymentTarget
			(name, type, organization_id, created_by_user_account_id, namespace, scope, geolocation_lat,
				geolocation_lon, agent_version_id)
			VALUES (@name, @type, @orgId, @userId, @namespace, @scope, @lat, @lon, @agentVersionId)
			RETURNING *
		)
		SELECT `+deploymentTargetOutputExpr+` FROM inserted dt`+deploymentTargetJoinExpr,
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
		return addDeploymentToTarget(ctx, dt)
	}
}

func UpdateDeploymentTarget(ctx context.Context, dt *types.DeploymentTargetWithCreatedBy, orgID uuid.UUID) error {
	agentUpdateStr := ""
	db := internalctx.GetDb(ctx)
	args := pgx.NamedArgs{
		"id":    dt.ID,
		"name":  dt.Name,
		"orgId": orgID,
		"lat":   nil,
		"lon":   nil,
	}
	if dt.AgentVersionID != nil {
		args["agentVersionId"] = dt.AgentVersionID
		agentUpdateStr = ", agent_version_id = @agentVersionId "
	}
	if dt.Geolocation != nil {
		maps.Copy(args, pgx.NamedArgs{"lat": dt.Geolocation.Lat, "lon": dt.Geolocation.Lon})
	}
	rows, err := db.Query(ctx,
		`WITH updated AS (
			UPDATE DeploymentTarget AS dt SET
				name = @name,
				geolocation_lat = @lat,
				geolocation_lon = @lon `+agentUpdateStr+`
			WHERE id = @id AND organization_id = @orgId RETURNING *
		)
		SELECT `+deploymentTargetWithStatusOutputExpr+` FROM updated dt`+deploymentTargetJoinExpr,
		args)
	if err != nil {
		return fmt.Errorf("could not update DeploymentTarget: %w", err)
	} else if updated, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByNameLax[types.DeploymentTargetWithCreatedBy]); err != nil {
		return fmt.Errorf("could not get updated DeploymentTarget: %w", err)
	} else {
		*dt = updated
		return addDeploymentToTarget(ctx, dt)
	}
}

func DeleteDeploymentTargetWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(ctx, `DELETE FROM DeploymentTarget WHERE id = @id`, pgx.NamedArgs{"id": id}); err != nil {
		return err
	} else if cmd.RowsAffected() == 0 {
		return apierrors.ErrNotFound
	} else {
		return nil
	}
}

func UpdateDeploymentTargetAccess(ctx context.Context, dt *types.DeploymentTarget, orgID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"UPDATE DeploymentTarget AS dt SET access_key_salt = @accessKeySalt, access_key_hash = @accessKeyHash "+
			"WHERE id = @id AND organization_id = @orgId RETURNING "+
			deploymentTargetOutputExprBase,
		pgx.NamedArgs{"accessKeySalt": dt.AccessKeySalt, "accessKeyHash": dt.AccessKeyHash, "id": dt.ID, "orgId": orgID})
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

func UpdateDeploymentTargetReportedAgentVersionID(
	ctx context.Context,
	dt *types.DeploymentTargetWithCreatedBy,
	agentVersionID uuid.UUID,
) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`WITH updated AS (
			UPDATE DeploymentTarget AS dt
			SET reported_agent_version_id = @agentVersionId
			WHERE id = @id
			RETURNING *
		)
		SELECT`+deploymentTargetWithStatusOutputExpr+`FROM updated dt`+deploymentTargetJoinExpr,
		pgx.NamedArgs{"id": dt.ID, "agentVersionId": agentVersionID},
	)
	if err != nil {
		return err
	} else if updated, err := pgx.CollectExactlyOneRow(rows,
		pgx.RowToAddrOfStructByName[types.DeploymentTargetWithCreatedBy]); err != nil {
		return err
	} else {
		*dt = *updated
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

func CleanupDeploymentTargetStatus(ctx context.Context, dt *types.DeploymentTarget) (int64, error) {
	if env.StatusEntriesMaxAge() == nil {
		return 0, nil
	}
	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(ctx, `
		DELETE FROM DeploymentTargetStatus
		       WHERE deployment_target_id = @deploymentTargetId AND
		             current_timestamp - created_at > @statusEntriesMaxAge`,
		pgx.NamedArgs{"deploymentTargetId": dt.ID, "statusEntriesMaxAge": env.StatusEntriesMaxAge()}); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}

func addDeploymentToTarget(ctx context.Context, dt *types.DeploymentTargetWithCreatedBy) error {
	if latest, err := GetLatestDeploymentForDeploymentTarget(ctx, dt.ID); errors.Is(err, apierrors.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	} else {
		dt.Deployment = latest
		return nil
	}
}
