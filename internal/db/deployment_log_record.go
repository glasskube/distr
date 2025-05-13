package db

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/glasskube/distr/api"
	internalctx "github.com/glasskube/distr/internal/context"
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
			AND lr.timestamp < @before
		ORDER BY lr.timestamp DESC
		LIMIT @limit`,
		pgx.NamedArgs{
			"deploymentId": deploymentID,
			"resource":     resource,
			"limit":        limit,
			"before":       before,
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
