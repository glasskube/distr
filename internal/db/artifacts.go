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
	artifactOutputExpr        = `a.id, a.created_at, a.organization_id, a.name `
	artifactVersionOutputExpr = `
		v.id,
		v.created_at,
		v.created_by_useraccount_id,
		v.updated_at,
		v.updated_by_useraccount_id,
		v.name,
		v.manifest_blob_digest,
		v.manifest_content_type,
		v.artifact_id
	`
)

func GetArtifactsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.Artifact, error) {
	// TODO impl, something like:
	/*
		SELECT
		    a.id, a.created_at, a.organization_id, a.name,
		    coalesce((
		            array_agg(row(av.id, av.created_at, av.name, av.artifact_id, avp.hash_sha256))
		    ), ARRAY[]::RECORD[]) as versions
		FROM artifact a
		INNER JOIN ArtifactVersion av ON a.id = av.artifact_id
		INNER JOIN ArtifactVersionPart avp ON av.id = avp.artifact_version_id
		INNER JOIN ArtifactBlob ab ON avp.artifact_blob_id = ab.id
		WHERE a.organization_id = 'b135b6b2-ebc9-4c13-a2c1-7eaa79455955' AND ab.is_lead = true
		GROUP BY a.id;

		db := internalctx.GetDb(ctx)
		if rows, err := db.Query(ctx, `
			SELECT `+artifactOutputExpr+`,
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
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT `+artifactOutputExpr+`
		FROM Artifact a
		WHERE a.organization_id = @orgId
		ORDER BY a.name ASC
		`,
		pgx.NamedArgs{"orgId": orgID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query Artifact: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.Artifact])
	if err != nil {
		return nil, fmt.Errorf("could not query Artifact: %w", err)
	}
	return result, nil
}

func GetArtifactsByLicenseOwnerID(ctx context.Context, ownerID uuid.UUID) ([]types.Artifact, error) {
	// TODO impl
	return nil, nil
}

func GetOrCreateArtifact(ctx context.Context, orgID uuid.UUID, artifactName string) (*types.Artifact, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT `+artifactOutputExpr+`
			FROM Artifact a
			WHERE a.name = @name AND a.organization_id = @orgId`,
		pgx.NamedArgs{
			"name":  artifactName,
			"orgId": orgID,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query artifact: %w", err)
	}
	if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.Artifact]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			artifact := &types.Artifact{Name: artifactName, OrganizationID: orgID}
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

func GetArtifactVersions(ctx context.Context, orgName, name string) ([]types.ArtifactVersion, error) {
	db := internalctx.GetDb(ctx)
	// TODO: Switch to org slug when implemented
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactVersionOutputExpr+`
		FROM Artifact a
		LEFT JOIN ArtifactVersion v ON a.id = v.artifact_id
		WHERE a.organization_id = @orgName
			AND a.name = @name
		ORDER BY v.name ASC`,
		pgx.NamedArgs{"orgName": orgName, "name": name},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.ArtifactVersion])
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	return result, nil
}

func GetLicensedArtifactVersions(
	ctx context.Context,
	orgName, name string,
	userID uuid.UUID,
) ([]types.ArtifactVersion, error) {
	panic("TODO: not implemented")
}

func GetArtifactVersion(ctx context.Context, orgName, name, reference string) (*types.ArtifactVersion, error) {
	db := internalctx.GetDb(ctx)
	// TODO: Switch to org slug when implemented
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactVersionOutputExpr+`
		FROM Artifact a
		LEFT JOIN ArtifactVersion v ON a.id = v.artifact_id
		WHERE a.organization_id = @orgName
			AND a.name = @name
			AND v.name = @reference`,
		pgx.NamedArgs{"orgName": orgName, "name": name, "reference": reference},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactVersion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	return &result, nil
}

func CreateArtifactVersion(ctx context.Context, av *types.ArtifactVersion) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO ArtifactVersion AS av (
            name, created_by_useraccount_id, manifest_blob_digest, manifest_content_type, artifact_id
        ) VALUES (
        	@name, @createdById, @manifestBlobDigest, @manifestContentType, @artifactId
        ) RETURNING *`,
		pgx.NamedArgs{
			"name":                av.Name,
			"createdById":         av.CreatedByUserAccountID,
			"manifestBlobDigest":  av.ManifestBlobDigest,
			"manifestContentType": av.ManifestContentType,
			"artifactId":          av.ArtifactID,
		},
	)
	if err != nil {
		return fmt.Errorf("could not insert ArtifactVersion: %w", err)
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

func CreateArtifactVersionPart(ctx context.Context, avp *types.ArtifactVersionPart) error {
	db := internalctx.GetDb(ctx)
	log := internalctx.GetLogger(ctx)
	log.Sugar().Infof("create %v", avp)
	if rows, err := db.Query(
		ctx,
		`INSERT INTO ArtifactVersionPart AS avp (
        	artifact_version_id, artifact_blob_digest
        ) VALUES (@versionId, @blobDigest)
		ON CONFLICT (artifact_version_id, artifact_blob_digest)
			DO UPDATE SET
				artifact_version_id = @versionId,
				artifact_blob_digest = @blobDigest
		RETURNING *`,
		pgx.NamedArgs{
			"versionId":  avp.ArtifactVersionID,
			"blobDigest": avp.ArtifactBlobDigest,
		},
	); err != nil {
		return fmt.Errorf("could not insert ArtifactVersionPart: %w", err)
	} else if result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[types.ArtifactVersionPart]); err != nil {
		return err
	} else {
		*avp = result
		return nil
	}
}

// CheckArtifactVersionHasChildren checks if one of the ArtifactVersionParts referencing this ArtifactVersion refer to
// another ArtifactVersion via its manifest digest.
func CheckArtifactVersionHasChildren(ctx context.Context, versionID uuid.UUID) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT exists(
			SELECT * FROM ArtifactVersionPart avp
			JOIN ArtifactVersion av ON av.manifest_blob_digest = avp.artifact_artifact_blob_digest
			WHERE avp.artifact_version_id = @versionId
		)`,
		pgx.NamedArgs{"versionId": versionID},
	)
	if err != nil {
		return false, fmt.Errorf("could not check artifact version parents: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return false, fmt.Errorf("could not check artifact version parents: %w", err)
	}
	return result.Exists, nil
}

func CreateArtifactPullLogEntry(ctx context.Context, versionID uuid.UUID, userID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	_, err := db.Exec(
		ctx,
		`INSERT INTO ArtifactVersionPull (artifact_version_id, useraccount_id) VALUES (@versionId, @userId)`,
		pgx.NamedArgs{"versionId": versionID, "userId": userID},
	)
	if err != nil {
		return fmt.Errorf("could not create artifact pull log entry: %w", err)
	}
	return nil
}

func GetArtifactVersionPullCount(ctx context.Context, versionID uuid.UUID) (int, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT count(SELECT * FROM ArtifactVersionPull WHERE artifact_version_id = @versionId)`,
		pgx.NamedArgs{"versionId": versionID},
	)
	if err != nil {
		return 0, fmt.Errorf("could not get pull count: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Count int }])
	if err != nil {
		return 0, fmt.Errorf("could not get pull count: %w", err)
	}
	return result.Count, nil
}

func GetArtifactVersionPullers(ctx context.Context, versionID uuid.UUID) ([]types.UserAccount, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`
		WITH LastPull AS (
			SELECT av.artifact_id, max(avp.created_at) AS latest_pull FROM ArtifactVersionPull avp
			JOIN ArtifactVersion av ON av.id = avp.artifact_version_id
			WHERE useraccount_id = @userId
			GROUP BY av.artifact_id
		)
		SELECT DISTINCT `+userAccountOutputExpr+`
			FROM UserAccount ua
			JOIN ArtifactVersionPull p ON ua.id = p.useraccount_id
			JOIN ArtifactVersion av ON av.id = p.artifact_version_id
			JOIN LastPull lp ON lp.artifact_id = av.artifact_id AND lp.latest_pull = avp.created_at
			WHERE p.artifact_version_id = @versionId
			ORDER BY p.created_at DESC
		`,
		pgx.NamedArgs{"versionId": versionID},
	)
	if err != nil {
		return nil, fmt.Errorf("could not get pullers: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.UserAccount])
	if err != nil {
		return nil, fmt.Errorf("could not get pullers: %w", err)
	}
	return result, nil
}
