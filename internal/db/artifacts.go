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

func GetArtifactsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.ArtifactWithTaggedVersion, error) {
	db := internalctx.GetDb(ctx)
	if artifactRows, err := db.Query(ctx, `
			SELECT
				a.id,
				a.created_at,
				a.organization_id,
				a.name,
				o.slug AS organization_slug,
				ARRAY []::RECORD[] AS versions
			FROM Artifact a
			JOIN Organization o ON o.id = a.organization_id
			WHERE a.organization_id = @orgId
			ORDER BY a.name`,
		pgx.NamedArgs{
			"orgId": orgID,
		}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err :=
		pgx.CollectRows(artifactRows, pgx.RowToAddrOfStructByName[types.ArtifactWithTaggedVersion]); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		res := make([]types.ArtifactWithTaggedVersion, len(artifacts))
		for i, artifact := range artifacts {
			if versions, err := getVersionsForArtifact(ctx, artifact.ID, nil); err != nil {
				return nil, fmt.Errorf("failed to get artifact versions: %w", err)
			} else {
				artifact.Versions = versions
			}
			res[i] = *artifact
		}
		return res, nil
	}
}

func GetArtifactsByLicenseOwnerID(ctx context.Context, orgID uuid.UUID, ownerID uuid.UUID) (
	[]types.ArtifactWithTaggedVersion, error,
) {
	db := internalctx.GetDb(ctx)
	if artifactRows, err := db.Query(ctx, `
			SELECT
				a.id,
				a.created_at,
				a.organization_id,
				a.name,
				o.slug AS organization_slug,
				ARRAY []::RECORD[] AS versions
			FROM Artifact a
			JOIN Organization o ON o.id = a.organization_id
			WHERE a.organization_id = @orgId
			AND EXISTS(
				SELECT ala.id
				FROM ArtifactLicense_Artifact ala
				INNER JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
				WHERE al.owner_useraccount_id = @ownerId AND (al.expires_at IS NULL OR al.expires_at > now())
				AND ala.artifact_id = a.id
			)
			ORDER BY a.name`,
		pgx.NamedArgs{
			"orgId":   orgID,
			"ownerId": ownerID,
		}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err :=
		pgx.CollectRows(artifactRows, pgx.RowToAddrOfStructByName[types.ArtifactWithTaggedVersion]); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		res := make([]types.ArtifactWithTaggedVersion, len(artifacts))
		for i, artifact := range artifacts {
			if versions, err := getVersionsForArtifact(ctx, artifact.ID, &ownerID); err != nil {
				return nil, fmt.Errorf("failed to get artifact versions: %w", err)
			} else {
				artifact.Versions = versions
			}
			res[i] = *artifact
		}
		return res, nil
	}
}

func getVersionsForArtifact(ctx context.Context, artifactID uuid.UUID, ownerID *uuid.UUID) (
	[]types.TaggedArtifactVersion,
	error,
) {
	checkLicense := false
	if ownerID != nil {
		checkLicense = true
	}
	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT av.id, av.created_at, av.manifest_blob_digest,
				coalesce((
					SELECT array_agg(row (avt.id, avt.name) ORDER BY avt.name)
					FROM ArtifactVersion avt
					WHERE avt.manifest_blob_digest = av.manifest_blob_digest
					AND avt.artifact_id = av.artifact_id
					AND avt.name NOT LIKE '%:%'
				), ARRAY []::RECORD[]) as tags
			FROM ArtifactVersion av
			WHERE av.artifact_id = @artifactId
			AND av.name LIKE '%:%'
			AND (
				NOT @checkLicense
				-- license check
				OR EXISTS (
					-- license for all versions of the artifact
					SELECT *
					FROM ArtifactLicense_Artifact ala
					INNER JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
					WHERE ala.artifact_id = @artifactId AND ala.artifact_version_id IS NULL
					AND al.owner_useraccount_id = @ownerId AND (al.expires_at IS NULL OR al.expires_at > now())
				)
				OR EXISTS (
					-- or license only for specific versions or their parent versions
					WITH RECURSIVE ArtifactVersionAggregate (id, manifest_blob_digest) AS (
						SELECT avx.id, avx.manifest_blob_digest
						FROM ArtifactVersion avx
						WHERE avx.manifest_blob_digest = av.manifest_blob_digest AND avx.artifact_id = @artifactId

						UNION ALL

						SELECT DISTINCT avx.id, avx.manifest_blob_digest
						FROM ArtifactVersion avx
						JOIN ArtifactVersionPart avp ON avx.id = avp.artifact_version_id
						JOIN ArtifactVersionAggregate agg ON avp.artifact_blob_digest = agg.manifest_blob_digest
					)
					SELECT *
					FROM ArtifactVersionAggregate avagg
					INNER JOIN ArtifactLicense_Artifact ala ON ala.artifact_version_id = avagg.id
					INNER JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
					WHERE al.owner_useraccount_id = @ownerId AND (al.expires_at IS NULL OR al.expires_at > now())
					AND ala.artifact_id = @artifactId
				)
			)
			AND EXISTS (
				-- only versions that have a tag
				SELECT avt.id
				FROM ArtifactVersion avt
				WHERE avt.manifest_blob_digest = av.manifest_blob_digest
				AND avt.artifact_id = av.artifact_id
				AND avt.name NOT LIKE '%:%'
			)
			`,
		pgx.NamedArgs{
			"artifactId":   artifactID,
			"ownerId":      ownerID,
			"checkLicense": checkLicense,
		}); err != nil {
		return nil, err
	} else if versions, err := pgx.CollectRows(rows, pgx.RowToStructByName[types.TaggedArtifactVersion]); err != nil {
		return nil, err
	} else {
		return versions, nil
	}
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
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactVersionOutputExpr+`
		FROM Artifact a
		JOIN Organization o ON o.id = a.organization_id
		LEFT JOIN ArtifactVersion v ON a.id = v.artifact_id
		WHERE o.slug = @orgName
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

func CheckLicenseForArtifact(ctx context.Context, orgName, name, reference string, userID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`WITH RECURSIVE ArtifactVersionAggregate (id, artifact_id, manifest_blob_digest) AS (
			SELECT av.id, av.artifact_id, av.manifest_blob_digest
				FROM Artifact a
				JOIN ArtifactVersion av ON a.id = av.artifact_id
				JOIN ArtifactVersion avx ON a.id = avx.artifact_id AND avx.manifest_blob_digest = av.manifest_blob_digest
				JOIN Organization o ON o.id = a.organization_id
				WHERE o.slug = @orgName
				AND a.name = @name
				AND (avx.name = @reference OR avx.manifest_blob_digest = @reference)
			UNION ALL
			SELECT DISTINCT av.id, av.artifact_id, av.manifest_blob_digest
				FROM ArtifactVersion av
				JOIN ArtifactVersionPart avp ON av.id = avp.artifact_version_id
				JOIN ArtifactVersionAggregate agg ON avp.artifact_blob_digest = agg.manifest_blob_digest
		)
		SELECT exists(
			SELECT *
				FROM ArtifactVersionAggregate av
				JOIN ArtifactLicense_Artifact ala
					ON av.artifact_id = ala.artifact_id
						AND (ala.artifact_version_id IS NULL OR ala.artifact_version_id = av.id)
				JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
				WHERE al.owner_useraccount_id = @userId
					AND (al.expires_at IS NULL OR al.expires_at > now())
		)`,
		pgx.NamedArgs{"orgName": orgName, "name": name, "reference": reference, "userId": userID},
	)
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	} else if !result.Exists {
		return apierrors.ErrForbidden
	}
	return nil
}

func CheckOrganizationForArtifactBlob(ctx context.Context, digest string, orgID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT exists(
			SELECT *
				FROM Artifact a
				JOIN ArtifactVersion av ON a.id = av.artifact_id
				JOIN ArtifactVersionPart avp ON av.id = avp.artifact_version_id
				WHERE avp.artifact_blob_digest = @digest
					AND a.organization_id = @orgId
		)`,
		pgx.NamedArgs{"digest": digest, "orgId": orgID},
	)
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	} else if !result.Exists {
		return apierrors.ErrForbidden
	}
	return nil
}

func CheckLicenseForArtifactBlob(ctx context.Context, digest string, userID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`WITH RECURSIVE ArtifactVersionAggregate (id, artifact_id, manifest_blob_digest) AS (
			SELECT av.id, av.artifact_id, av.manifest_blob_digest
				FROM ArtifactVersion av
				JOIN ArtifactVersionPart avp ON av.id = avp.artifact_version_id
				WHERE avp.artifact_blob_digest = @digest
			UNION ALL
			SELECT DISTINCT av.id, av.artifact_id, av.manifest_blob_digest
				FROM ArtifactVersion av
				JOIN ArtifactVersionPart avp ON av.id = avp.artifact_version_id
				JOIN ArtifactVersionAggregate agg ON avp.artifact_blob_digest = agg.manifest_blob_digest
		)
		SELECT exists(
			SELECT *
				FROM ArtifactVersionAggregate av
				JOIN ArtifactLicense_Artifact ala
					ON av.artifact_id = ala.artifact_id
						AND (ala.artifact_version_id IS NULL OR ala.artifact_version_id = av.id)
				JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
				WHERE al.owner_useraccount_id = @userId
					AND (al.expires_at IS NULL OR al.expires_at > now())
		)`,
		pgx.NamedArgs{"digest": digest, "userId": userID},
	)
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return fmt.Errorf("could not query ArtifactVersion: %w", err)
	} else if !result.Exists {
		return apierrors.ErrForbidden
	}
	return nil
}

func GetArtifactVersion(ctx context.Context, orgName, name, reference string) (*types.ArtifactVersion, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactVersionOutputExpr+`
		FROM Artifact a
		JOIN Organization o ON o.id = a.organization_id
		LEFT JOIN ArtifactVersion v ON a.id = v.artifact_id
		WHERE o.slug = @orgName
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

func GetDigestArtifactVersion(ctx context.Context, orgName, name, reference string) (*types.ArtifactVersion, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactVersionOutputExpr+`
		FROM Artifact a
		JOIN Organization o ON o.id = a.organization_id
		JOIN ArtifactVersion v ON a.id = v.artifact_id
		JOIN ArtifactVersion avx ON a.id = avx.artifact_id AND v.manifest_blob_digest = avx.manifest_blob_digest
		WHERE o.slug = @orgName
			AND a.name = @name
			AND (avx.name = @reference OR avx.manifest_blob_digest = @reference)
			AND v.name LIKE '%:%'`,
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
			JOIN ArtifactVersion av ON av.manifest_blob_digest = avp.artifact_blob_digest
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
