package main

import (
	"context"
	"math/rand"
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

	deploymentID := uuid.MustParse("98be36e4-aa8a-4596-a5e8-8da0e0974105")
	revisionID := uuid.MustParse("addb2eac-c1e5-4580-a36e-42c011327dd5")
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
			Body:      randomString(1000),
		})
		createdAt = createdAt.Add(statusInterval)
	}
	util.Must(db.BulkCreateDeploymentLogRecordWithCreatedAt(ctx, deploymentID, revisionID, ds))
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
