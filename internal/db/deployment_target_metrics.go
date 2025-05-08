package db

import (
	"context"

	"github.com/glasskube/distr/api"
	context2 "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DeploymentTargetWithLatestMetrics struct {
	types.DeploymentTarget
	api.AgentDeploymentTargetMetrics
}

func GetLatestDeploymentTargetMetrics(ctx context.Context, orgID, userID uuid.UUID, userRole types.UserRole) (
	[]DeploymentTargetWithLatestMetrics, error) {
	return make([]DeploymentTargetWithLatestMetrics, 0), nil
}

func CreateDeploymentTargetMetrics(
	ctx context.Context,
	dt *types.DeploymentTarget,
	metrics *api.AgentDeploymentTargetMetrics,
) error {
	db := context2.GetDb(ctx)
	rows, err := db.Query(ctx,
		"INSERT INTO DeploymentTargetMetrics (deployment_target_id, cpu_cores_m, cpu_usage, memory_bytes, memory_usage) "+
			"VALUES (@deploymentTargetId, @cpuCoresM, @cpuUsage, @memoryBytes, @memoryUsage)",
		pgx.NamedArgs{
			"deploymentTargetId": dt.ID,
			"cpuCoresM":          metrics.CPUCoresM,
			"cpuUsage":           metrics.CPUUsage,
			"memoryBytes":        metrics.MemoryBytes,
			"memoryUsage":        metrics.MemoryUsage,
		})
	if err != nil {
		return err
	} else {
		// TODO check error handling again
		rows.Close()
		return nil
	}
}

func CleanupDeploymentTargetMetrics(ctx context.Context, dt *types.DeploymentTarget) (int64, error) {
	if env.MetricsEntriesMaxAge() == nil {
		return 0, nil
	}
	db := context2.GetDb(ctx)
	if cmd, err := db.Exec(ctx, `
		DELETE FROM DeploymentTargetMetrics
		       WHERE deployment_target_id = @deploymentTargetId AND
		             current_timestamp - created_at > @metricsEntriesMaxAge`,
		pgx.NamedArgs{"deploymentTargetId": dt.ID, "metricsEntriesMaxAge": env.MetricsEntriesMaxAge()}); err != nil {
		return 0, err
	} else {
		return cmd.RowsAffected(), nil
	}
}
