package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/containers/image/v5/manifest"
	"github.com/glasskube/distr/internal/apierrors"
	internalctx "github.com/glasskube/distr/internal/context"
	"github.com/glasskube/distr/internal/env"
	"github.com/glasskube/distr/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	artifactOutputExpr         = ` a.id, a.created_at, a.organization_id, a.name, a.image_id `
	artifactOutputWithSlugExpr = artifactOutputExpr + ", o.slug AS organization_slug"
	artifactVersionOutputExpr  = `
		v.id,
		v.created_at,
		v.created_by_useraccount_id,
		v.updated_at,
		v.updated_by_useraccount_id,
		v.name,
		v.manifest_blob_digest,
		v.manifest_blob_size,
		v.manifest_content_type,
		v.manifest_data,
		v.artifact_id
	`
	artifactDownloadsOutExpr = `
			count(DISTINCT avpl.id) as downloads_total,
			count(DISTINCT avpl.useraccount_id) as downloaded_by_count,
			coalesce(array_agg(DISTINCT avpl.useraccount_id)
				FILTER (WHERE avpl.useraccount_id IS NOT NULL), ARRAY[]::UUID[]) as downloaded_by_users
	`
)

func GetArtifactsByOrgID(ctx context.Context, orgID uuid.UUID) ([]types.ArtifactWithDownloads, error) {
	db := internalctx.GetDb(ctx)
	if artifactRows, err := db.Query(ctx, `
			SELECT `+artifactOutputWithSlugExpr+`,`+artifactDownloadsOutExpr+`
			FROM Artifact a
			JOIN Organization o ON o.id = a.organization_id
			LEFT JOIN ArtifactVersion av ON a.id = av.artifact_id
			LEFT JOIN ArtifactVersionPull avpl ON avpl.artifact_version_id = av.id
			WHERE a.organization_id = @orgId
			GROUP BY a.id, a.created_at, a.organization_id, a.name, o.slug
			ORDER BY a.name`,
		pgx.NamedArgs{
			"orgId": orgID,
		}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err := pgx.CollectRows(
		artifactRows, pgx.RowToStructByName[types.ArtifactWithDownloads],
	); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		return artifacts, nil
	}
}

func GetArtifactsByLicenseOwnerID(ctx context.Context, orgID uuid.UUID, ownerID uuid.UUID) (
	[]types.ArtifactWithDownloads, error,
) {
	db := internalctx.GetDb(ctx)
	if artifactRows, err := db.Query(ctx, `
			SELECT `+artifactOutputWithSlugExpr+`,`+artifactDownloadsOutExpr+`
			FROM Artifact a
			JOIN Organization o ON o.id = a.organization_id
			LEFT JOIN ArtifactVersion av ON a.id = av.artifact_id
			LEFT JOIN ArtifactVersionPull avpl ON avpl.artifact_version_id = av.id AND avpl.useraccount_id = @ownerId
			WHERE a.organization_id = @orgId
			AND EXISTS(
				SELECT ala.id
				FROM ArtifactLicense_Artifact ala
				INNER JOIN ArtifactLicense al ON ala.artifact_license_id = al.id
				WHERE al.owner_useraccount_id = @ownerId AND (al.expires_at IS NULL OR al.expires_at > now())
				AND ala.artifact_id = a.id
			)
			GROUP BY a.id, a.created_at, a.organization_id, a.name, o.slug
			ORDER BY a.name`,
		pgx.NamedArgs{
			"orgId":   orgID,
			"ownerId": ownerID,
		}); err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	} else if artifacts, err := pgx.CollectRows(
		artifactRows, pgx.RowToStructByName[types.ArtifactWithDownloads],
	); err != nil {
		return nil, fmt.Errorf("failed to collect artifacts: %w", err)
	} else {
		return artifacts, nil
	}
}

func GetArtifactByID(ctx context.Context, orgID uuid.UUID, artifactID uuid.UUID, ownerID *uuid.UUID) (
	*types.ArtifactWithTaggedVersion,
	error,
) {
	db := internalctx.GetDb(ctx)
	restrictDownloads := ownerID != nil

	if artifactRows, err := db.Query(
		ctx, `
			SELECT `+artifactOutputWithSlugExpr+`,
				ARRAY []::RECORD[] AS versions,`+artifactDownloadsOutExpr+`
			FROM Artifact a
			JOIN Organization o ON o.id = a.organization_id
			LEFT JOIN ArtifactVersion av ON a.id = av.artifact_id
			LEFT JOIN ArtifactVersionPull avpl ON avpl.artifact_version_id = av.id
				AND (NOT @restrict OR avpl.useraccount_id = @ownerId)
			WHERE a.id = @id AND a.organization_id = @orgId
			GROUP BY a.id, a.created_at, a.organization_id, a.name, o.slug`,
		pgx.NamedArgs{
			"id":       artifactID,
			"orgId":    orgID,
			"restrict": restrictDownloads,
			"ownerId":  ownerID,
		},
	); err != nil {
		return nil, fmt.Errorf("failed to query artifact by ID: %w", err)
	} else if artifact, err := pgx.CollectExactlyOneRow(
		artifactRows, pgx.RowToAddrOfStructByName[types.ArtifactWithTaggedVersion],
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to collect artifact by ID: %w", err)
	} else if versions, err := GetVersionsForArtifact(ctx, artifact.ID, ownerID); err != nil {
		return nil, fmt.Errorf("failed to get artifact versions: %w", err)
	} else if ownerID != nil && len(versions) == 0 {
		return nil, apierrors.ErrNotFound
	} else {
		artifact.Versions = versions
		return artifact, nil
	}
}

func GetArtifactByName(ctx context.Context, orgSlug, name string) (*types.Artifact, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT`+artifactOutputExpr+`
			FROM Artifact a
			JOIN Organization o on o.id = a.organization_id
			WHERE o.slug = @orgSlug AND a.name = @name
			ORDER BY a.name`,
		pgx.NamedArgs{
			"orgSlug": orgSlug,
			"name":    name,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query artifacts: %w", err)
	}
	if a, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.Artifact]); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = apierrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	} else {
		return a, nil
	}
}

func GetVersionsForArtifact(ctx context.Context, artifactID uuid.UUID, ownerID *uuid.UUID) (
	[]types.TaggedArtifactVersion,
	error,
) {
	checkLicense := ownerID != nil

	db := internalctx.GetDb(ctx)
	if rows, err := db.Query(ctx, `
			SELECT
				av.id,
				av.created_at,
				av.manifest_blob_digest,
				av.manifest_content_type,
				av.manifest_data,
				coalesce((
					SELECT array_agg(row (avt.id, avt.name, (
						SELECT ROW(
							count(distinct avplx.id),
							count(DISTINCT avplx.useraccount_id),
							coalesce(array_agg(DISTINCT avplx.useraccount_id)
									 FILTER (WHERE avplx.useraccount_id IS NOT NULL), ARRAY[]::UUID[])
						)
						FROM ArtifactVersionPull avplx WHERE avplx.artifact_version_id = avt.id
						)) ORDER BY avt.name
					)
					FROM ArtifactVersion avt
					WHERE avt.manifest_blob_digest = av.manifest_blob_digest
					AND avt.artifact_id = av.artifact_id
					AND avt.name NOT LIKE '%:%'
				), ARRAY []::RECORD[]) AS tags,
				av.manifest_blob_size + coalesce(sum(avp.artifact_blob_size), 0) AS size,
				`+artifactDownloadsOutExpr+`
			FROM ArtifactVersion av
			LEFT JOIN LATERAL (
				WITH RECURSIVE aggregate AS (
					SELECT avp.artifact_version_id as base_av_id,
						   avp.artifact_version_id as related_av_id,
						   avp.artifact_blob_digest,
						   avp.artifact_blob_size
					FROM ArtifactVersionPart avp
						WHERE avp.artifact_version_id = av.id
					UNION ALL
					SELECT aggregate.base_av_id, av1.id, avp.artifact_blob_digest, avp.artifact_blob_size
					FROM aggregate
					JOIN ArtifactVersion av1 ON av1.manifest_blob_digest = aggregate.artifact_blob_digest
					JOIN ArtifactVersionPart avp ON av1.id = avp.artifact_version_id
				)
				SELECT DISTINCT * FROM aggregate
			) avp ON av.id = avp.base_av_id
			LEFT JOIN ArtifactVersionPull avpl ON avpl.artifact_version_id = avp.related_av_id AND
				(NOT @checkLicense OR avpl.useraccount_id = @ownerId)
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
			GROUP BY av.id, av.created_at, av.manifest_blob_digest
			ORDER BY av.created_at DESC
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
		for i, version := range versions {
			version.InferredType = types.ManifestTypeGeneric
			if strings.HasPrefix(version.ManifestContentType, "application/vnd.docker") {
				version.InferredType = types.ManifestTypeContainerImage
			} else if !manifest.MIMETypeIsMultiImage(version.ManifestContentType) && len(version.ManifestData) > 0 {
				parsedManifest, err := manifest.FromBlob(version.ManifestData, version.ManifestContentType)
				if err != nil {
					return nil, err
				}

				if strings.HasPrefix(parsedManifest.ConfigInfo().MediaType, "application/vnd.cncf.helm") ||
					slices.ContainsFunc(parsedManifest.LayerInfos(), func(layer manifest.LayerInfo) bool {
						return strings.HasPrefix(layer.MediaType, "application/vnd.cncf.helm")
					}) {
					version.InferredType = types.ManifestTypeHelmChart
				}
			}
			versions[i] = version
		}
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

func CreateArtifactVersion(ctx context.Context, av *types.ArtifactVersion) error {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`INSERT INTO ArtifactVersion AS av (
            name,
			created_by_useraccount_id,
			manifest_blob_digest,
			manifest_blob_size,
			manifest_content_type,
			manifest_data,
			artifact_id
        ) VALUES (
        	@name, @createdById, @manifestBlobDigest, @manifestBlobSize, @manifestContentType, @manifestData,
			@artifactId
        ) RETURNING *`,
		pgx.NamedArgs{
			"name":                av.Name,
			"createdById":         av.CreatedByUserAccountID,
			"manifestBlobDigest":  av.ManifestBlobDigest,
			"manifestBlobSize":    av.ManifestBlobSize,
			"manifestContentType": av.ManifestContentType,
			"manifestData":        av.ManifestData,
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
        	artifact_version_id, artifact_blob_digest, artifact_blob_size
        ) VALUES (@versionId, @blobDigest, @blobSize)
		ON CONFLICT (artifact_version_id, artifact_blob_digest)
			DO UPDATE SET
				artifact_version_id = @versionId,
				artifact_blob_digest = @blobDigest,
				artifact_blob_size = @blobSize
		RETURNING *`,
		pgx.NamedArgs{
			"versionId":  avp.ArtifactVersionID,
			"blobDigest": avp.ArtifactBlobDigest,
			"blobSize":   avp.ArtifactBlobSize,
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

func CreateArtifactPullLogEntry(ctx context.Context, versionID, userID uuid.UUID, remoteAddress string) error {
	db := internalctx.GetDb(ctx)
	remoteAddressPtr := &remoteAddress
	if remoteAddress == "" {
		remoteAddressPtr = nil
	}
	_, err := db.Exec(
		ctx,
		`INSERT INTO ArtifactVersionPull (artifact_version_id, useraccount_id, remote_address)
		VALUES (@versionId, @userId, @remoteAddress)`,
		pgx.NamedArgs{"versionId": versionID, "userId": userID, "remoteAddress": remoteAddressPtr},
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

func EnsureArtifactTagLimitForInsert(ctx context.Context, orgID uuid.UUID) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT count(av.name) + 1 < coalesce(
			o.artifact_tag_limit,
			CASE WHEN @defaultLimit > 0 THEN @defaultLimit ELSE @maxLimit END
		)
		FROM ArtifactVersion av
		JOIN Artifact a on av.artifact_id = a.id
		JOIN Organization o ON a.organization_id = o.id
		WHERE o.id = @orgId AND av.name NOT LIKE '%:%'
		GROUP BY o.id;`,
		pgx.NamedArgs{
			"orgId":        orgID,
			"defaultLimit": env.ArtifactTagsDefaultLimitPerOrg(),
			"maxLimit":     math.MaxInt32,
		},
	)
	if err != nil {
		return false, fmt.Errorf("could not check quota: %w", err)
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Ok bool }])
	// If there are no rows, the organization has no tags yet, and the limit is not exceeded.
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("could not check quota: %w", err)
	} else {
		return result.Ok, nil
	}
}

func GetArtifactVersionPulls(
	ctx context.Context,
	orgID uuid.UUID,
	count int,
	before time.Time,
) ([]types.ArtifactVersionPull, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(
		ctx,
		`SELECT
			p.created_at,
			p.remote_address,
			CASE WHEN u.id IS NOT NULL THEN (`+userAccountOutputExpr+`) ELSE NULL END,
			(`+artifactOutputExpr+`),
			(`+artifactVersionOutputExpr+`)
		FROM ArtifactVersionPull p
			LEFT JOIN UserAccount u ON u.id = p.useraccount_id
			JOIN ArtifactVersion v ON v.id = p.artifact_version_id
			JOIN Artifact A on a.id = v.artifact_id
		WHERE a.organization_id = @orgId
			AND p.created_at < @before
		ORDER BY p.created_at DESC
		LIMIT @count`,
		pgx.NamedArgs{
			"orgId":  orgID,
			"count":  count,
			"before": before,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not query ArtifactVersionPulls: %w", err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowToStructByPos[types.ArtifactVersionPull])
	if err != nil {
		return nil, fmt.Errorf("could not scan ArtifactVersionPulls: %w", err)
	}
	return result, nil
}

func UpdateArtifactImage(ctx context.Context, artifact *types.ArtifactWithTaggedVersion, imageID uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	row := db.QueryRow(ctx,
		`UPDATE Artifact SET image_id = @imageId WHERE id = @id RETURNING image_id`,
		pgx.NamedArgs{"imageId": imageID, "id": artifact.ID},
	)
	if err := row.Scan(&artifact.ImageID); err != nil {
		return fmt.Errorf("could not save image id to artifact: %w", err)
	}
	return nil
}

func ArtifactIsReferencedInLicenses(ctx context.Context, artifactID uuid.UUID) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT count(ala.id) > 0
		FROM ArtifactLicense_Artifact ala
		WHERE ala.artifact_id = @artifactId`,
		pgx.NamedArgs{"artifactId": artifactID},
	)
	if err != nil {
		return false, err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return false, err
	}
	return result.Exists, nil
}

func ArtifactVersionIsReferencedInLicenses(ctx context.Context, versionID uuid.UUID) (bool, error) {
	db := internalctx.GetDb(ctx)
	rows, err := db.Query(ctx, `
		SELECT count(ala.id) > 0
		FROM ArtifactLicense_Artifact ala
		WHERE ala.artifact_version_id = @versionId`,
		pgx.NamedArgs{"versionId": versionID},
	)
	if err != nil {
		return false, err
	}
	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[struct{ Exists bool }])
	if err != nil {
		return false, err
	}
	return result.Exists, nil
}

func DeleteArtifactWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, `DELETE FROM Artifact WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
	} else if cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("could not delete Artifact: %w", err)
	}

	return nil
}

func DeleteArtifactVersionWithID(ctx context.Context, id uuid.UUID) error {
	db := internalctx.GetDb(ctx)
	cmd, err := db.Exec(ctx, `DELETE FROM ArtifactVersion WHERE id = @id`, pgx.NamedArgs{"id": id})
	if err != nil {
		if pgerr := (*pgconn.PgError)(nil); errors.As(err, &pgerr) && pgerr.Code == pgerrcode.ForeignKeyViolation {
			err = fmt.Errorf("%w: %w", apierrors.ErrConflict, err)
		}
	} else if cmd.RowsAffected() == 0 {
		err = apierrors.ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("could not delete ArtifactVersion: %w", err)
	}

	return nil
}
