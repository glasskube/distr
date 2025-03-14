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
	artifactLicenseOutputExpr = " al.id, al.created_at, al.name, al.expires_at, al.owner_useraccount_id, al.organization_id "
	artifactLicenseOwnerExpr  = " CASE WHEN al.owner_useraccount_id IS NOT NULL THEN (" + userAccountOutputExpr + ") END as owner "

	artifactLicenseCompleteOutExpr = artifactLicenseOutputExpr + ", " + artifactLicenseOwnerExpr
)

func GetArtifactLicenses(ctx context.Context, orgID uuid.UUID) ([]types.ArtifactLicense, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT `+artifactLicenseCompleteOutExpr+`,
			coalesce((
				SELECT array_agg(
					row(
						(`+artifactOutputExpr+`),

					)
				)
				FROM ArtifactLicense_Artifact ala
				INNER JOIN Artifact a ON a.id = ala.artifact_id
				WHERE ala.artifact_license_id = al.id
			), ARRAY[]::RECORD[]) as artifacts
		FROM ArtifactLicense al
		LEFT JOIN UserAccount u ON al.owner_useraccount_id = u.id
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

func CreateArtifactLicense(ctx context.Context, license *types.ArtifactLicenseBase) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO ArtifactLicense AS al (
			name, expires_at, organization_id, owner_useraccount_id
		) VALUES (
			@name, @expiresAt, @organizationId, @ownerUserAccountId
		) RETURNING `+artifactLicenseOutputExpr,
		pgx.NamedArgs{
			"name":               license.Name,
			"expiresAt":          license.ExpiresAt,
			"organizationId":     license.OrganizationID,
			"ownerUserAccountId": license.OwnerUserAccountID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert ArtifactLicense: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactLicenseBase]); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else {
		*license = result
		return nil
	}
}

func AddArtifactToArtifactLicense(
	ctx context.Context,
	licenseID uuid.UUID,
	artifactId uuid.UUID,
	artifactVersionId *uuid.UUID,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(
		ctx,
		`INSERT INTO ArtifactLicense_Artifact (artifact_license_id, artifact_id, artifact_version_id)
		VALUES (@licenseId, @id, @versionId)
		ON CONFLICT (artifact_license_id, artifact_id, artifact_version_id) DO NOTHING`,
		pgx.NamedArgs{
			"licenseId": licenseID,
			"id":        artifactId,
			"versionId": artifactVersionId,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert relation: %w", err)
	}
	return nil
}

func GetArtifactLicenseByID(ctx context.Context, id uuid.UUID) (*types.ArtifactLicense, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		"SELECT "+artifactLicenseCompleteOutExpr+
			"FROM ArtifactLicense al "+
			"LEFT JOIN UserAccount u ON al.owner_useraccount_id = u.id "+
			"WHERE al.id = @id ",
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactLicense: %w", err)
	}

	if result, err :=
		pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactLicense]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not collect ArtifactLicense: %w", err)
	} else {
		return &result, nil
	}
}
