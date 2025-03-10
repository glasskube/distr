package db

import (
	"context"
	"fmt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	artifactOutputExpr         = `a.id, a.created_at, a.organization_id, a.name `
	artifactWithTagsOutputExpr = artifactOutputExpr + `,
		coalesce((
			SELECT array_agg(row(at.id, at.created_at, at.hash, at.labels, at.artifact_id) ORDER BY av.created_at ASC)
			FROM ArtifactTag at
			WHERE at.artifact_id = a.id
		), array[]::record[]) AS tags `
)

func GetArtifactsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.Artifact, error) {
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
		SELECT `+artifactWithTagsOutputExpr+`
		FROM Artifact a WHERE a.organization_id = :orgId
		ORDER BY a.name`, pgx.NamedArgs{
		"orgId": orgID,
	}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Artifact]); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		return artifacts, nil
	}
}

func GetArtifactsByLicenseOwnerID(ctx context.Context, ownerID uuid.UUID) ([]types.Artifact, error) {
	// TODO impl
	return nil, nil
}
