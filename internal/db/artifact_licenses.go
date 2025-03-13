package db

import (
	"context"
	"fmt"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func GetArtifactLicenses(ctx context.Context, orgID uuid.UUID) ([]types.ArtifactLicense, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT al.id, al.created_at, al.name, al.expires_at
		FROM ArtifactLicense al
		WHERE al.organization_id = @orgId
		ORDER BY al.name`,
		pgx.NamedArgs{"orgId": orgID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactLicense: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.ArtifactLicense])
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactLicense: %w", err)
	}
	return result, nil
}
