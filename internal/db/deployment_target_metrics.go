package db

import (
	"context"
	"fmt"

	"github.com/distr-sh/distr/api"
	internalctx "github.com/distr-sh/distr/internal/context"
	"github.com/distr-sh/distr/internal/env"
	"github.com/distr-sh/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DeploymentTargetLatestMetrics struct {
	ID uuid.UUID `db:"id" json:"id"`
	api.AgentDeploymentTargetMetrics
}

func GetLatestDeploymentTargetMetrics(
	ctx context.Context,
	orgID uuid.UUID,
	customerOrganizationID *uuid.UUID,
) ([]DeploymentTargetLatestMetrics, error) {
	db := internalctx.GetDb(ctx)
	isVendorUser := customerOrganizationID == nil

	if rows, err := db.Query(ctx,
		`SELECT dt.id, dtm.cpu_cores_millis, dtm.cpu_usage, dtm.memory_bytes, dtm.memory_usage FROM
			DeploymentTarget dt
			LEFT JOIN UserAccount u
				ON dt.created_by_user_account_id = u.id
			LEFT JOIN Organization_UserAccount j
				ON u.id = j.user_account_id
			LEFT JOIN (
				-- copied from getting deployment target latest status:
				-- find the creation date of the latest status entry for each deployment target
				-- IMPORTANT: The sub-query here might seem inefficient but it is MUCH FASTER than using a GROUP BY clause
				-- because it can utilize an index!!
				SELECT
					dt1.id AS deployment_target_id,
					(SELECT max(created_at) FROM DeploymentTargetMetrics WHERE deployment_target_id = dt1.id) AS max_created_at
				FROM DeploymentTarget dt1
			) metrics_max
				ON dt.id = metrics_max.deployment_target_id
			INNER JOIN DeploymentTargetMetrics dtm
				ON dt.id = dtm.deployment_target_id
					AND dtm.created_at = metrics_max.max_created_at
			WHERE dt.organization_id = @orgId
			AND (@isVendorUser OR dt.customer_organization_id = @customerOrganizationId)
			AND dt.metrics_enabled = true
			ORDER BY u.name, u.email, dt.name`,
		pgx.NamedArgs{"orgId": orgID, "customerOrganizationId": customerOrganizationID, "isVendorUser": isVendorUser},
	); err != nil {
		return nil, fmt.Errorf("failed to query DeploymentTargets: %w", err)
	} else if result, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[DeploymentTargetLatestMetrics],
	); err != nil {
		return nil, fmt.Errorf("failed to get DeploymentTargets: %w", err)
	} else {
		return result, nil
	}
}

func CreateDeploymentTargetMetrics(
	ctx context.Context,
	dt *types.DeploymentTarget,
	metrics *api.AgentDeploymentTargetMetrics,
) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO DeploymentTargetMetrics "+
			"(deployment_target_id, cpu_cores_millis, cpu_usage, memory_bytes, memory_usage) "+
			"VALUES (@deploymentTargetId, @cpuCoresMillis, @cpuUsage, @memoryBytes, @memoryUsage)",
		pgx.NamedArgs{
			"deploymentTargetId": dt.ID,
			"cpuCoresMillis":     metrics.CPUCoresMillis,
			"cpuUsage":           metrics.CPUUsage,
			"memoryBytes":        metrics.MemoryBytes,
			"memoryUsage":        metrics.MemoryUsage,
		})
	if err != nil {
		return err
	} else {
		rows.Close()
		return rows.Err()
	}
}

func CleanupDeploymentTargetMetrics(ctx context.Context) (int64, error) {
	if env.MetricsEntriesMaxAge() == nil {
		return 0, nil
	}
	db := internalctx.GetDb(ctx)
	if cmd, err := db.Exec(
		ctx,
		`DELETE FROM DeploymentTargetMetrics dtm
		USING (
			SELECT
				dt.id AS deployment_target_id,
				(SELECT max(created_at) FROM DeploymentTargetMetrics WHERE deployment_target_id = dt.id)
					AS max_created_at
			FROM DeploymentTarget dt
		) max_created_at
		WHERE dtm.deployment_target_id = max_created_at.deployment_target_id
			AND dtm.created_at < max_created_at.max_created_at
			AND current_timestamp - dtm.created_at > @metricsEntriesMaxAge`,
		pgx.NamedArgs{"metricsEntriesMaxAge": env.MetricsEntriesMaxAge()},
	); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}
