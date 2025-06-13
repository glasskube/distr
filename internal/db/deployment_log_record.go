package db

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/glasskube/distr/api"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	deploymentLogRecordOutputExpr = `
	lr.id, lr.created_at, lr.deployment_id, lr.deployment_revision_id, lr.resource, lr.timestamp, lr.severity, lr.body
	`
)

func SaveDeploymentLogRecords(ctx context.Context, records []api.DeploymentLogRecord) error {
	db := internalctx.GetDb(ctx)
	_, err := db.CopyFrom(
		ctx,
		pgx.Identifier{"deploymentlogrecord"},
		[]string{"deployment_id", "deployment_revision_id", "resource", "timestamp", "severity", "body"},
		pgx.CopyFromSlice(len(records), func(i int) ([]any, error) {
			r := records[i]
			return []any{r.DeploymentID, r.DeploymentRevisionID, r.Resource, r.Timestamp, r.Severity, r.Body}, nil
		}),
	)
	return err
}

func ValidateDeploymentLogRecords(
	ctx context.Context,
	deploymentTargetID uuid.UUID,
	records []api.DeploymentLogRecord,
) error {
	if len(records) == 0 {
		return nil
	}

	db := internalctx.GetDb(ctx)

	tuples := map[struct{ deploymentID, revisionID uuid.UUID }]struct{}{}
	for _, record := range records {
		tuples[struct{ deploymentID, revisionID uuid.UUID }{
			deploymentID: record.DeploymentID,
			revisionID:   record.DeploymentRevisionID,
		}] = struct{}{}
	}

	for tuple := range tuples {
		rows, err := db.Query(
			ctx,
			`SELECT 1
			FROM Deployment d
			JOIN DeploymentRevision dr ON d.id = dr.deployment_id
			WHERE d.deployment_target_id = @deploymentTargetId
				AND d.id = @deploymentId
				AND dr.id = @deploymentRevisionId`,
			pgx.NamedArgs{
				"deploymentTargetId":   deploymentTargetID,
				"deploymentId":         tuple.deploymentID,
				"deploymentRevisionId": tuple.revisionID,
			},
		)
		if err != nil {
			return fmt.Errorf("could not query DeploymentTarget: %w", err)
		}
		if _, err := pgx.CollectRows(rows, pgx.RowTo[int64]); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: deployment %s and revision %s does not exist in deployment target %s",
					apierrors.ErrNotFound, tuple.deploymentID, tuple.revisionID, deploymentTargetID)
			}
			return fmt.Errorf("could not collect DeploymentTarget: %w", err)
		}
	}

	return nil
}

func GetDeploymentLogRecordResources(ctx context.Context,
	deploymentID uuid.UUID,
) ([]string, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT DISTINCT resource FROM DeploymentLogRecord WHERE deployment_id = @deploymentId`,
		pgx.NamedArgs{"deploymentId": deploymentID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query DeploymentLogRecord: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("could not collect DeploymentLogRecord: %w", err)
	}
	slices.Sort(result)
	return result, nil
}

func GetDeploymentLogRecords(
	ctx context.Context,
	deploymentID uuid.UUID,
	resource string,
	limit int,
	before time.Time,
	after time.Time,
) ([]types.DeploymentLogRecord, error) {
	if before.IsZero() {
		before = time.Now()
	}
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT `+deploymentLogRecordOutputExpr+`
		FROM DeploymentLogRecord lr
		WHERE lr.deployment_id = @deploymentId
			AND lr.resource = @resource
			AND lr.timestamp BETWEEN @after AND @before
		ORDER BY lr.timestamp DESC
		LIMIT @limit`,
		pgx.NamedArgs{
			"deploymentId": deploymentID,
			"resource":     resource,
			"limit":        limit,
			"before":       before,
			"after":        after,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query DeploymentLogRecord: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.DeploymentLogRecord])
	if err != nil {
		return nil, fmt.Errorf("could not collect DeploymentLogRecord: %w", err)
	}
	return result, nil
}

// CleanupDeploymentLogRecords deletes logrecords for all deployments but keeps the
// last [env.LogRecordEntriesMaxCount] records for each (deployment_id, resource) group.
//
// If [env.LogRecordEntriesMaxCount] is nil, no cleanup is performed.
func CleanupDeploymentLogRecords(ctx context.Context) (int64, error) {
	limit := env.LogRecordEntriesMaxCount()
	if limit == nil {
		return 0, nil
	}

	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM DeploymentLogRecord
		WHERE id NOT IN (
			SELECT keep.id FROM (
				SELECT DISTINCT lr.deployment_id, lr.resource FROM DeploymentLogRecord lr
			) rn
			JOIN LATERAL (
				SELECT *
				FROM DeploymentLogRecord lr
				WHERE lr.deployment_id = rn.deployment_id AND lr.resource = rn.resource
				ORDER BY lr.timestamp DESC
				LIMIT @limit
			) keep ON true
		)`,
		pgx.NamedArgs{"limit": limit},
	)
	if err != nil {
		return 0, fmt.Errorf("error cleaning up DeploymentLogRecords: %w", err)
	} else {
		return cmd.RowsAffected(), nil
	}
}
