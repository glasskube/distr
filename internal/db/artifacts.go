package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	artifactOutputExpr         = `a.id, a.created_at, a.organization_id, a.name `
	artifactWithTagsOutputExpr = artifactOutputExpr + `,
		coalesce((
			SELECT array_agg(row(av.id, av.created_at, av.hash, av.labels, av.artifact_id) ORDER BY av.created_at ASC)
			FROM ArtifactVersion av
			WHERE av.artifact_id = a.id
		), array[]::record[]) AS versions `
)

func GetArtifactsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.Artifact, error) {
	// TODO impl
	/*db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
		SELECT `+artifactWithTagsOutputExpr+`
		FROM Artifact a WHERE a.organization_id = @orgId
		ORDER BY a.name`, pgx.NamedArgs{
		"orgId": orgID,
	}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Artifact]); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		return artifacts, nil
	}*/
	return nil, nil
}

func GetArtifactsByLicenseOwnerID(ctx context.Context, ownerID uuid.UUID) ([]types.Artifact, error) {
	// TODO impl
	return nil, nil
}

func GetOrCreateArtifact(ctx context.Context, org *types.Organization, artifactName string) (*types.Artifact, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT `+artifactOutputExpr+`
			FROM Artifact a
			WHERE a.name = @name AND a.organization_id = @orgId`,
		pgx.NamedArgs{
			"name":  artifactName,
			"orgId": org.ID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query artifact: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Artifact]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			artifact := &types.Artifact{Name: artifactName, OrganizationID: org.ID}
			err = CreateArtifact(ctx, artifact)
			return artifact, err
		}
		return nil, fmt.Errorf("could not collect artifact: %w", err)
	} else {
		return &result, nil
	}
}

func CreateArtifact(ctx context.Context, artifact *types.Artifact) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO Artifact AS a (name, organization_id) VALUES (@name, @organizationId) RETURNING `+artifactOutputExpr,
		pgx.NamedArgs{
			"name":           artifact.Name,
			"organizationId": artifact.OrganizationID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert Artifact: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Artifact]); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else {
		*artifact = result
		return nil
	}
}

func CreateArtifactVersion(ctx context.Context, av *types.ArtifactVersion) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO ArtifactVersion AS av (
            name, created_by_useraccount_id, artifact_id
        ) VALUES (
        	@name, @createdById, @artifactId
        ) RETURNING *`,
		pgx.NamedArgs{
			"name":        av.Name,
			"createdById": av.CreatedByUserAccountID,
			"artifactId":  av.ArtifactID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert Artifact: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactVersion]); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else {
		*av = result
		return nil
	}
}
