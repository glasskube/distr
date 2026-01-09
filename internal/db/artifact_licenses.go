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
	artifactLicenseOutExpr = `al.id, al.created_at, al.name, al.expires_at, ` +
		`al.customer_organization_id, al.organization_id `
	artifactSelectionsOutExpor = `
		(
			SELECT array_agg(DISTINCT row(
				ala.artifact_id,
				coalesce((
					SELECT array_agg(alax.artifact_version_id) FILTER (WHERE alax.artifact_version_id IS NOT NULL)
					FROM ArtifactLicense_artifact alax
					WHERE alax.artifact_license_id = ala.artifact_license_id AND alax.artifact_id = ala.artifact_id
				 ), ARRAY[]::UUID[])
				))
			FROM ArtifactLicense_Artifact ala
			WHERE ala.artifact_license_id = al.id
		) as artifacts `
)

func GetArtifactLicenses(ctx context.Context, orgID uuid.UUID) ([]types.ArtifactLicense, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT `+artifactLicenseOutExpr+`, `+artifactSelectionsOutExpor+`
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

func CreateArtifactLicense(ctx context.Context, license *types.ArtifactLicenseBase) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		WITH inserted AS (
			INSERT INTO ArtifactLicense (
				name, expires_at, organization_id, customer_organization_id
			) VALUES (
				@name, @expiresAt, @organizationId, @customerOrganizationId
			) RETURNING *
		)
		SELECT `+artifactLicenseOutExpr+`
		FROM inserted al`,
		pgx.NamedArgs{
			"name":                   license.Name,
			"expiresAt":              license.ExpiresAt,
			"organizationId":         license.OrganizationID,
			"customerOrganizationId": license.CustomerOrganizationID,
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

func UpdateArtifactLicense(ctx context.Context, license *types.ArtifactLicenseBase) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		WITH updated AS (
			UPDATE ArtifactLicense SET
			name = @name,
            expires_at = @expiresAt,
            customer_organization_id = @customerOrganizationId
		 	WHERE id = @id RETURNING *
		)
		SELECT `+artifactLicenseOutExpr+`
		FROM updated al`,
		pgx.NamedArgs{
			"id":                     license.ID,
			"name":                   license.Name,
			"expiresAt":              license.ExpiresAt,
			"customerOrganizationId": license.CustomerOrganizationID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not update ArtifactLicense: %w", err)
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

func RemoveAllArtifactsFromLicense(
	ctx context.Context,
	id uuid.UUID,
) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(
		ctx,
		`DELETE FROM ArtifactLicense_Artifact
		WHERE artifact_license_id = @artifactLicenseId`,
		pgx.NamedArgs{
			"artifactLicenseId": id,
		},
	)
	if err != nil {
		return fmt.Errorf("could not delete relation: %w", err)
	} else {
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
	rows, err := db.Query(ctx, `
			SELECT `+artifactLicenseOutExpr+`, `+artifactSelectionsOutExpor+`
			FROM ArtifactLicense al
			WHERE al.id = @id `,
		pgx.NamedArgs{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactLicense: %w", err)
	}

	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactLicense]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not collect ArtifactLicense: %w", err)
	} else {
		return &result, nil
	}
}

func DeleteArtifactLicenseWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, `DELETE FROM ArtifactLicense WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return err
	} else if cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("could not delete ArtifactLicense: %w", err)
	}

	return nil
}

func DeleteArtifactLicensesWithOrganizationID(ctx context.Context, organizationID uuid.UUID) (int64, error) {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM ArtifactLicense WHERE organization_id = @organizationID`,
		pgx.NamedArgs{"organizationID": organizationID},
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return 0, fmt.Errorf("could not delete ArtifactLicense: %w", err)
	}

	return cmd.RowsAffected(), nil
}

func DeleteArtifactLicensesWithOrganizationSubscriptionType(
	ctx context.Context,
	subscriptionType []types.SubscriptionType,
) (int64, error) {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(
		ctx,
		`DELETE FROM ArtifactLicense WHERE organization_id IN (
			SELECT id FROM Organization WHERE subscription_type = ANY(@subscriptionType)
		)`,
		pgx.NamedArgs{"subscriptionType": subscriptionType},
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
		return 0, fmt.Errorf("could not delete ArtifactLicenses: %w", err)
	}

	return cmd.RowsAffected(), nil
}
