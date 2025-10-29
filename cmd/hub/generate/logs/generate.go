package main

import (
	"context"
	"time"

	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/db"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/svc"
	"github.com/glasskube/distr/internal/types"
	"github.com/glasskube/distr/internal/util"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()
	env.Initialize()
	registry := util.Require(svc.NewDefault(ctx))
	defer func() { _ = registry.Shutdown(ctx) }()
	ctx = internalctx.WithDb(ctx, registry.GetDbPool())

	deploymentID := uuid.MustParse("a8bd20ef-f477-4349-80ba-11198632a6fb")
	revisionID := uuid.MustParse("2109c2a9-cb56-4b12-9bb0-ec3a0e682e4d")
	statusCount := 2000000
	statusInterval := 5 * time.Second

	now := time.Now().UTC()
	createdAt := now.Add(time.Duration(statusCount) * -statusInterval)
	var ds []types.DeploymentLogRecord
	for createdAt.Before(now) {
		ds = append(ds, types.DeploymentLogRecord{
			CreatedAt: createdAt,
			Resource:  "example-resource",
			Timestamp: createdAt,
			Severity:  "error",
			Body:      "example log record",
		})
		createdAt = createdAt.Add(statusInterval)
	}
	util.Must(db.BulkCreateDeploymentLogRecordWithCreatedAt(ctx, deploymentID, revisionID, ds))
}
